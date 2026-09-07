package httpql

import "encoding/gob"

// The concrete expression types are registered so that a parsed query stored in
// a gob-encoded project (as an Expression interface value) round-trips. The AST
// holds only exported, serializable fields — no compiled regexes — so gob can
// encode it directly.
func init() {
	gob.Register(BinaryExpr{})
	gob.Register(NotExpr{})
	gob.Register(Clause{})
	gob.Register(FreeText{})
}
