package httpql

import (
	"fmt"
	"time"
)

// ColumnKind describes how a database column backs an HTTPQL field.
type ColumnKind int

const (
	ColumnString ColumnKind = iota
	ColumnInt
	ColumnTime // stored as Unix milliseconds
)

// SQLColumn maps an HTTPQL field to a database column expression.
type SQLColumn struct {
	Expr string
	Kind ColumnKind
}

// CompileSQL translates an expression into a SQL WHERE fragment (with `?`
// placeholders and its arguments) that is guaranteed to be a *superset* of the
// rows the expression actually matches: every matching row satisfies the
// fragment, but the fragment may admit extra rows. Callers must still evaluate
// the expression in memory over the returned rows for an exact result.
//
// columns maps "namespace.name" (e.g. "req.method") to a backing column. Fields
// without a column, header lookups, regexes, free text and NOT are left to the
// in-memory pass and contribute a permissive "1=1".
func CompileSQL(expr Expression, columns map[string]SQLColumn) (string, []any) {
	if expr == nil {
		return "1=1", nil
	}

	switch e := expr.(type) {
	case BinaryExpr:
		leftSQL, leftArgs := CompileSQL(e.Left, columns)
		rightSQL, rightArgs := CompileSQL(e.Right, columns)

		op := "AND"
		if e.Op == TokenOr {
			op = "OR"
		}

		args := append(leftArgs, rightArgs...)

		return fmt.Sprintf("(%s %s %s)", leftSQL, op, rightSQL), args
	case Clause:
		return compileClause(e, columns)
	default:
		// NotExpr, FreeText and anything else: leave to the in-memory pass.
		return "1=1", nil
	}
}

func compileClause(c Clause, columns map[string]SQLColumn) (string, []any) {
	if c.Field.HeaderKey != "" || c.Field.Name == "header" {
		return "1=1", nil
	}

	col, ok := columns[c.Field.Namespace+"."+c.Field.Name]
	if !ok {
		return "1=1", nil
	}

	switch col.Kind {
	case ColumnInt:
		return compileIntClause(c, col.Expr)
	case ColumnTime:
		return compileTimeClause(c, col.Expr)
	default:
		return compileStringClause(c, col.Expr)
	}
}

func compileComparison(op TokenType) (string, bool) {
	switch op {
	case TokenEq:
		return "=", true
	case TokenNe:
		return "<>", true
	case TokenGt:
		return ">", true
	case TokenGte:
		return ">=", true
	case TokenLt:
		return "<", true
	case TokenLte:
		return "<=", true
	default:
		return "", false
	}
}

func compileStringClause(c Clause, col string) (string, []any) {
	// Only exact comparisons are pushed down: Go's == / < and SQLite's default
	// BINARY collation agree byte-for-byte. cont/like/regex rely on
	// case-insensitive or Unicode semantics that SQLite's LIKE does not
	// reproduce exactly, so pushing them could wrongly exclude a matching row;
	// they are left to the in-memory pass.
	if sqlOp, ok := compileComparison(c.Op); ok {
		return fmt.Sprintf("%s %s ?", col, sqlOp), []any{c.Value.asString()}
	}

	return "1=1", nil
}

func compileIntClause(c Clause, col string) (string, []any) {
	if c.Value.Kind != ValueInt {
		return "1=1", nil
	}

	if sqlOp, ok := compileComparison(c.Op); ok {
		return fmt.Sprintf("%s %s ?", col, sqlOp), []any{c.Value.Int}
	}

	return "1=1", nil
}

func compileTimeClause(c Clause, col string) (string, []any) {
	t, err := time.Parse(time.RFC3339, c.Value.asString())
	if err != nil {
		return "1=1", nil
	}

	if sqlOp, ok := compileComparison(c.Op); ok {
		return fmt.Sprintf("%s %s ?", col, sqlOp), []any{t.UnixMilli()}
	}

	return "1=1", nil
}
