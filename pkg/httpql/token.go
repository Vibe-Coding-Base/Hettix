package httpql

// TokenType enumerates the lexical tokens of the HTTPQL query language.
type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenEOF

	TokenAnd
	TokenOr
	TokenNot

	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenDot

	// Operators.
	TokenEq
	TokenNe
	TokenCont
	TokenNCont
	TokenLike
	TokenNLike
	TokenRegex
	TokenNRegex
	TokenGt
	TokenGte
	TokenLt
	TokenLte

	TokenIdent  // bareword: field segment, boolean, or unquoted value
	TokenString // quoted string literal
	TokenInt    // integer literal
)

// Token is a single lexical unit with its literal text and source position.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

var operatorWords = map[string]TokenType{
	"eq":     TokenEq,
	"ne":     TokenNe,
	"cont":   TokenCont,
	"ncont":  TokenNCont,
	"like":   TokenLike,
	"nlike":  TokenNLike,
	"regex":  TokenRegex,
	"nregex": TokenNRegex,
	"gt":     TokenGt,
	"gte":    TokenGte,
	"lt":     TokenLt,
	"lte":    TokenLte,
}

func (t TokenType) isOperator() bool {
	return t >= TokenEq && t <= TokenLte
}
