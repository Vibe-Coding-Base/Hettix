package httpql

import (
	"errors"
	"fmt"
	"strings"
)

// maxDepth bounds recursion so that pathological input (deeply nested groups or
// repeated NOT) cannot exhaust the goroutine stack, which is a fatal,
// unrecoverable error in Go. HTTPQL is reachable from the unauthenticated API.
const maxDepth = 256

// ErrMaxDepthExceeded is returned when a query nests beyond maxDepth.
var ErrMaxDepthExceeded = fmt.Errorf("httpql: query nesting exceeds maximum depth of %d", maxDepth)

// Parse parses an HTTPQL query string into an expression tree. An empty query
// returns a nil expression, which callers treat as "match everything".
func Parse(input string) (Expression, error) {
	tokens, err := Lex(input)
	if err != nil {
		return nil, err
	}

	p := &parser{tokens: tokens}

	if p.peek().Type == TokenEOF {
		return nil, nil
	}

	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.peek().Type != TokenEOF {
		return nil, fmt.Errorf("httpql: unexpected trailing token %q at position %d", p.peek().Literal, p.peek().Pos)
	}

	return expr, nil
}

type parser struct {
	tokens []Token
	pos    int
	depth  int
}

func (p *parser) peek() Token {
	return p.tokens[p.pos]
}

func (p *parser) advance() Token {
	tok := p.tokens[p.pos]
	if tok.Type != TokenEOF {
		p.pos++
	}

	return tok
}

func (p *parser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == TokenOr {
		p.advance()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		left = BinaryExpr{Op: TokenOr, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseAnd() (Expression, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for {
		switch {
		case p.peek().Type == TokenAnd:
			p.advance()
		case startsPrimary(p.peek().Type):
			// Implicit AND between adjacent terms.
		default:
			return left, nil
		}

		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}

		left = BinaryExpr{Op: TokenAnd, Left: left, Right: right}
	}
}

func (p *parser) parseNot() (Expression, error) {
	if p.peek().Type == TokenNot {
		p.advance()

		expr, err := p.parseNot()
		if err != nil {
			return nil, err
		}

		return NotExpr{Expr: expr}, nil
	}

	return p.parsePrimary()
}

func (p *parser) parsePrimary() (Expression, error) {
	p.depth++
	if p.depth > maxDepth {
		return nil, ErrMaxDepthExceeded
	}
	defer func() { p.depth-- }()

	tok := p.peek()

	switch tok.Type {
	case TokenLParen:
		p.advance()

		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}

		if p.peek().Type != TokenRParen {
			return nil, fmt.Errorf("httpql: expected %q at position %d", ")", p.peek().Pos)
		}
		p.advance()

		return expr, nil
	case TokenString:
		p.advance()
		return FreeText{Value: tok.Literal}, nil
	case TokenIdent:
		return p.parseIdentTerm()
	default:
		return nil, fmt.Errorf("httpql: unexpected token %q at position %d", tok.Literal, tok.Pos)
	}
}

// parseIdentTerm parses either a `field operator value` clause (when the ident
// is a namespace followed by a dot) or a bare free-text term.
func (p *parser) parseIdentTerm() (Expression, error) {
	first := p.advance()

	if p.peek().Type != TokenDot || !isNamespace(first.Literal) {
		return FreeText{Value: first.Literal}, nil
	}

	field, err := p.parseField(first.Literal)
	if err != nil {
		return nil, err
	}

	opTok := p.advance()
	if !opTok.Type.isOperator() {
		return nil, fmt.Errorf("httpql: expected an operator after %q at position %d", field, opTok.Pos)
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return Clause{Field: field, Op: opTok.Type, Value: value}, nil
}

func (p *parser) parseField(namespace string) (Field, error) {
	p.advance() // consume the dot

	nameTok := p.peek()
	if nameTok.Type != TokenIdent {
		return Field{}, fmt.Errorf("httpql: expected a field name after %q. at position %d", namespace, nameTok.Pos)
	}
	p.advance()

	field := Field{Namespace: strings.ToLower(namespace), Name: canonicalFieldName(nameTok.Literal)}
	if field.Namespace == "res" {
		field.Namespace = "resp"
	}

	if p.peek().Type == TokenLBracket {
		p.advance()

		keyTok := p.peek()
		if keyTok.Type != TokenString {
			return Field{}, fmt.Errorf("httpql: expected a quoted header name at position %d", keyTok.Pos)
		}
		p.advance()

		if p.peek().Type != TokenRBracket {
			return Field{}, fmt.Errorf("httpql: expected %q at position %d", "]", p.peek().Pos)
		}
		p.advance()

		field.HeaderKey = keyTok.Literal
	}

	return field, nil
}

func (p *parser) parseValue() (Value, error) {
	tok := p.advance()

	switch tok.Type {
	case TokenString:
		return Value{Kind: ValueString, Str: tok.Literal}, nil
	case TokenInt:
		n, err := parseInt(tok.Literal)
		if err != nil {
			return Value{}, err
		}

		return Value{Kind: ValueInt, Int: n}, nil
	case TokenIdent:
		switch strings.ToLower(tok.Literal) {
		case "true":
			return Value{Kind: ValueBool, Bool: true}, nil
		case "false":
			return Value{Kind: ValueBool, Bool: false}, nil
		default:
			return Value{Kind: ValueString, Str: tok.Literal}, nil
		}
	default:
		return Value{}, fmt.Errorf("httpql: expected a value at position %d, got %q", tok.Pos, tok.Literal)
	}
}

func startsPrimary(t TokenType) bool {
	switch t {
	case TokenLParen, TokenNot, TokenIdent, TokenString, TokenInt:
		return true
	default:
		return false
	}
}

// canonicalFieldName maps legacy field names onto their current HTTPQL names so
// previously saved queries keep working.
func canonicalFieldName(name string) string {
	switch strings.ToLower(name) {
	case "statuscode":
		return "code"
	case "statusreason":
		return "reason"
	case "timestamp":
		return "created_at"
	default:
		return strings.ToLower(name)
	}
}

func parseInt(s string) (int64, error) {
	var n int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errors.New("httpql: invalid integer literal")
		}
		n = n*10 + int64(s[i]-'0')
	}

	return n, nil
}
