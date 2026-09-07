package httpql

import (
	"fmt"
	"strings"
	"unicode"
)

// Lex tokenizes input into a slice of tokens terminated by a TokenEOF. Unlike a
// channel-based lexer it holds no goroutine, so an abandoned parse cannot leak
// one. It returns an error on an unterminated string or comment.
func Lex(input string) ([]Token, error) {
	l := &lexer{input: input}
	return l.run()
}

type lexer struct {
	input string
	pos   int
}

func (l *lexer) run() ([]Token, error) {
	var tokens []Token

	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			return tokens, nil
		}
	}
}

func (l *lexer) next() (Token, error) {
	if err := l.skipTrivia(); err != nil {
		return Token{}, err
	}

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Pos: l.pos}, nil
	}

	start := l.pos
	c := l.input[l.pos]

	switch c {
	case '(':
		l.pos++
		return Token{TokenLParen, "(", start}, nil
	case ')':
		l.pos++
		return Token{TokenRParen, ")", start}, nil
	case '[':
		l.pos++
		return Token{TokenLBracket, "[", start}, nil
	case ']':
		l.pos++
		return Token{TokenRBracket, "]", start}, nil
	case '.':
		l.pos++
		return Token{TokenDot, ".", start}, nil
	case '"':
		return l.lexString()
	case '=':
		l.pos++
		if l.peek() == '~' {
			l.pos++
			return Token{TokenRegex, "=~", start}, nil
		}
		if l.peek() == '=' {
			l.pos++
		}
		return Token{TokenEq, "=", start}, nil
	case '!':
		l.pos++
		switch l.peek() {
		case '=':
			l.pos++
			return Token{TokenNe, "!=", start}, nil
		case '~':
			l.pos++
			return Token{TokenNRegex, "!~", start}, nil
		}
		return Token{}, fmt.Errorf("httpql: unexpected %q at position %d", "!", start)
	case '>':
		l.pos++
		if l.peek() == '=' {
			l.pos++
			return Token{TokenGte, ">=", start}, nil
		}
		return Token{TokenGt, ">", start}, nil
	case '<':
		l.pos++
		if l.peek() == '=' {
			l.pos++
			return Token{TokenLte, "<=", start}, nil
		}
		return Token{TokenLt, "<", start}, nil
	}

	if c >= '0' && c <= '9' {
		return l.lexNumber(), nil
	}

	if isWordStart(rune(c)) {
		return l.lexWord(), nil
	}

	return Token{}, fmt.Errorf("httpql: unexpected character %q at position %d", string(c), start)
}

func (l *lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}

	return l.input[l.pos]
}

func (l *lexer) skipTrivia() error {
	for l.pos < len(l.input) {
		c := l.input[l.pos]

		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			l.pos++
		case c == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/':
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
		case c == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '*':
			end := strings.Index(l.input[l.pos+2:], "*/")
			if end < 0 {
				return fmt.Errorf("httpql: unterminated block comment at position %d", l.pos)
			}
			l.pos += end + 4
		default:
			return nil
		}
	}

	return nil
}

func (l *lexer) lexString() (Token, error) {
	start := l.pos
	l.pos++ // opening quote

	var b strings.Builder

	for l.pos < len(l.input) {
		c := l.input[l.pos]
		switch c {
		case '"':
			l.pos++
			return Token{TokenString, b.String(), start}, nil
		case '\\':
			l.pos++
			if l.pos >= len(l.input) {
				return Token{}, fmt.Errorf("httpql: unterminated string at position %d", start)
			}
			switch l.input[l.pos] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(l.input[l.pos])
			}
			l.pos++
		default:
			b.WriteByte(c)
			l.pos++
		}
	}

	return Token{}, fmt.Errorf("httpql: unterminated string at position %d", start)
}

func (l *lexer) lexNumber() Token {
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
		l.pos++
	}

	return Token{TokenInt, l.input[start:l.pos], start}
}

func (l *lexer) lexWord() Token {
	start := l.pos
	for l.pos < len(l.input) && isWordPart(rune(l.input[l.pos])) {
		l.pos++
	}

	literal := l.input[start:l.pos]

	switch strings.ToLower(literal) {
	case "and":
		return Token{TokenAnd, literal, start}
	case "or":
		return Token{TokenOr, literal, start}
	case "not":
		return Token{TokenNot, literal, start}
	}

	if op, ok := operatorWords[strings.ToLower(literal)]; ok {
		return Token{op, literal, start}
	}

	return Token{TokenIdent, literal, start}
}

func isWordStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isWordPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
}
