package httpql

import (
	"fmt"
	"strconv"
	"strings"
)

// Expression is a node in a parsed HTTPQL query.
type Expression interface {
	String() string
}

// BinaryExpr combines two sub-expressions with a boolean AND or OR operator.
type BinaryExpr struct {
	Op    TokenType // TokenAnd or TokenOr
	Left  Expression
	Right Expression
}

func (e BinaryExpr) String() string {
	op := "AND"
	if e.Op == TokenOr {
		op = "OR"
	}

	return fmt.Sprintf("(%s %s %s)", e.Left, op, e.Right)
}

// NotExpr negates a sub-expression.
type NotExpr struct {
	Expr Expression
}

func (e NotExpr) String() string {
	return fmt.Sprintf("(NOT %s)", e.Expr)
}

// Clause is a single `field operator value` comparison.
type Clause struct {
	Field Field
	Op    TokenType
	Value Value
}

func (e Clause) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Field, operatorString(e.Op), e.Value)
}

// FreeText matches a bare term against a request/response's textual fields.
type FreeText struct {
	Value string
}

func (e FreeText) String() string {
	return strconv.Quote(e.Value)
}

// Field is a namespaced field reference, optionally addressing a single header.
type Field struct {
	Namespace string // "req" or "resp"
	Name      string
	HeaderKey string // set when the field is header["Key"]
}

func (f Field) String() string {
	if f.Name == "header" && f.HeaderKey != "" {
		return fmt.Sprintf("%s.header[%q]", f.Namespace, f.HeaderKey)
	}

	return f.Namespace + "." + f.Name
}

// ValueKind distinguishes the concrete type of a parsed value.
type ValueKind int

const (
	ValueString ValueKind = iota
	ValueInt
	ValueBool
)

// Value is a typed literal on the right-hand side of a clause.
type Value struct {
	Kind ValueKind
	Str  string
	Int  int64
	Bool bool
}

func (v Value) String() string {
	switch v.Kind {
	case ValueInt:
		return strconv.FormatInt(v.Int, 10)
	case ValueBool:
		return strconv.FormatBool(v.Bool)
	default:
		return strconv.Quote(v.Str)
	}
}

func operatorString(t TokenType) string {
	for word, tt := range operatorWords {
		if tt == t {
			return word
		}
	}

	return "?"
}

func isNamespace(s string) bool {
	switch strings.ToLower(s) {
	case "req", "resp", "res":
		return true
	default:
		return false
	}
}
