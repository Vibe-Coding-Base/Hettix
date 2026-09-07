package httpql

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Record is the data an HTTPQL query is evaluated against: a request and its
// optional response, with fields already extracted into query-addressable form.
type Record struct {
	Request  RequestData
	Response *ResponseData
}

// RequestData holds the req.* fields.
type RequestData struct {
	ID        string
	Method    string
	Host      string
	Path      string
	Query     string
	Ext       string
	URL       string
	Proto     string
	Port      int
	Len       int
	TLS       bool
	CreatedAt time.Time
	Header    http.Header
	Body      string
}

// ResponseData holds the resp.* fields.
type ResponseData struct {
	Proto     string
	Reason    string
	Code      int
	Len       int
	RoundTrip int
	Header    http.Header
	Body      string
}

// Eval reports whether rec satisfies expr. A nil expr matches everything.
func Eval(expr Expression, rec Record) (bool, error) {
	if expr == nil {
		return true, nil
	}

	switch e := expr.(type) {
	case BinaryExpr:
		left, err := Eval(e.Left, rec)
		if err != nil {
			return false, err
		}

		if e.Op == TokenAnd {
			if !left {
				return false, nil
			}
			return Eval(e.Right, rec)
		}

		if left {
			return true, nil
		}
		return Eval(e.Right, rec)
	case NotExpr:
		v, err := Eval(e.Expr, rec)
		if err != nil {
			return false, err
		}
		return !v, nil
	case Clause:
		return evalClause(e, rec)
	case FreeText:
		return evalFreeText(e.Value, rec), nil
	default:
		return false, fmt.Errorf("httpql: unsupported expression type %T", expr)
	}
}

func evalClause(c Clause, rec Record) (bool, error) {
	if c.Field.Name == "header" {
		return matchHeader(c.Op, headerValues(c.Field, rec), c.Value)
	}

	switch resolveKind(c.Field) {
	case kindInt:
		return matchInt(c.Op, resolveInt(c.Field, rec), c.Value)
	case kindBool:
		return matchBool(c.Op, resolveBool(c.Field, rec), c.Value)
	case kindTime:
		return matchTime(c.Op, resolveTime(c.Field, rec), c.Value)
	case kindString:
		return matchString(c.Op, resolveString(c.Field, rec), c.Value)
	default:
		return false, fmt.Errorf("httpql: unknown field %s", c.Field)
	}
}

type fieldKind int

const (
	kindUnknown fieldKind = iota
	kindString
	kindInt
	kindBool
	kindTime
)

func resolveKind(f Field) fieldKind {
	switch f.Namespace {
	case "req":
		switch f.Name {
		case "port", "len":
			return kindInt
		case "tls":
			return kindBool
		case "created_at":
			return kindTime
		case "id", "method", "host", "path", "query", "ext", "url", "proto", "body":
			return kindString
		}
	case "resp":
		switch f.Name {
		case "code", "len", "roundtrip":
			return kindInt
		case "proto", "reason", "body":
			return kindString
		}
	}

	return kindUnknown
}

func resolveString(f Field, rec Record) string {
	if f.Namespace == "req" {
		switch f.Name {
		case "id":
			return rec.Request.ID
		case "method":
			return rec.Request.Method
		case "host":
			return rec.Request.Host
		case "path":
			return rec.Request.Path
		case "query":
			return rec.Request.Query
		case "ext":
			return rec.Request.Ext
		case "url":
			return rec.Request.URL
		case "proto":
			return rec.Request.Proto
		case "body":
			return rec.Request.Body
		}
	}

	if f.Namespace == "resp" && rec.Response != nil {
		switch f.Name {
		case "proto":
			return rec.Response.Proto
		case "reason":
			return rec.Response.Reason
		case "body":
			return rec.Response.Body
		}
	}

	return ""
}

func resolveInt(f Field, rec Record) int64 {
	if f.Namespace == "req" {
		switch f.Name {
		case "port":
			return int64(rec.Request.Port)
		case "len":
			return int64(rec.Request.Len)
		}
	}

	if f.Namespace == "resp" && rec.Response != nil {
		switch f.Name {
		case "code":
			return int64(rec.Response.Code)
		case "len":
			return int64(rec.Response.Len)
		case "roundtrip":
			return int64(rec.Response.RoundTrip)
		}
	}

	return 0
}

func resolveBool(f Field, rec Record) bool {
	if f.Namespace == "req" && f.Name == "tls" {
		return rec.Request.TLS
	}

	return false
}

func resolveTime(f Field, rec Record) time.Time {
	if f.Namespace == "req" && f.Name == "created_at" {
		return rec.Request.CreatedAt
	}

	return time.Time{}
}

func headerValues(f Field, rec Record) []string {
	var header http.Header
	if f.Namespace == "req" {
		header = rec.Request.Header
	} else if rec.Response != nil {
		header = rec.Response.Header
	}

	if header == nil {
		return nil
	}

	if f.HeaderKey != "" {
		return header.Values(f.HeaderKey)
	}

	var lines []string
	for key, values := range header {
		for _, v := range values {
			lines = append(lines, key+": "+v)
		}
	}

	return lines
}

func matchString(op TokenType, actual string, v Value) (bool, error) {
	want := v.asString()

	switch op {
	case TokenEq:
		return actual == want, nil
	case TokenNe:
		return actual != want, nil
	case TokenCont:
		return strings.Contains(strings.ToLower(actual), strings.ToLower(want)), nil
	case TokenNCont:
		return !strings.Contains(strings.ToLower(actual), strings.ToLower(want)), nil
	case TokenLike:
		return matchLike(actual, want), nil
	case TokenNLike:
		return !matchLike(actual, want), nil
	case TokenRegex, TokenNRegex:
		re, err := compileRegexp(want)
		if err != nil {
			return false, err
		}
		matched := re.MatchString(actual)
		if op == TokenNRegex {
			return !matched, nil
		}
		return matched, nil
	case TokenGt:
		return actual > want, nil
	case TokenGte:
		return actual >= want, nil
	case TokenLt:
		return actual < want, nil
	case TokenLte:
		return actual <= want, nil
	default:
		return false, fmt.Errorf("httpql: operator %s not supported for string fields", operatorString(op))
	}
}

func matchInt(op TokenType, actual int64, v Value) (bool, error) {
	if v.Kind != ValueInt {
		return false, fmt.Errorf("httpql: expected an integer value, got %s", v)
	}

	switch op {
	case TokenEq:
		return actual == v.Int, nil
	case TokenNe:
		return actual != v.Int, nil
	case TokenGt:
		return actual > v.Int, nil
	case TokenGte:
		return actual >= v.Int, nil
	case TokenLt:
		return actual < v.Int, nil
	case TokenLte:
		return actual <= v.Int, nil
	default:
		return false, fmt.Errorf("httpql: operator %s not supported for integer fields", operatorString(op))
	}
}

func matchBool(op TokenType, actual bool, v Value) (bool, error) {
	if v.Kind != ValueBool {
		return false, fmt.Errorf("httpql: expected a boolean value, got %s", v)
	}

	switch op {
	case TokenEq:
		return actual == v.Bool, nil
	case TokenNe:
		return actual != v.Bool, nil
	default:
		return false, fmt.Errorf("httpql: operator %s not supported for boolean fields", operatorString(op))
	}
}

func matchTime(op TokenType, actual time.Time, v Value) (bool, error) {
	want, err := time.Parse(time.RFC3339, v.asString())
	if err != nil {
		return false, fmt.Errorf("httpql: invalid RFC3339 time %q: %w", v.asString(), err)
	}

	switch op {
	case TokenEq:
		return actual.Equal(want), nil
	case TokenNe:
		return !actual.Equal(want), nil
	case TokenGt:
		return actual.After(want), nil
	case TokenGte:
		return actual.After(want) || actual.Equal(want), nil
	case TokenLt:
		return actual.Before(want), nil
	case TokenLte:
		return actual.Before(want) || actual.Equal(want), nil
	default:
		return false, fmt.Errorf("httpql: operator %s not supported for time fields", operatorString(op))
	}
}

// matchHeader applies a string operator across a header's values. Positive
// operators match when any value matches; negated operators match only when no
// value matches.
func matchHeader(op TokenType, values []string, v Value) (bool, error) {
	negated := op == TokenNe || op == TokenNCont || op == TokenNRegex || op == TokenNLike

	positiveOp := op
	switch op {
	case TokenNe:
		positiveOp = TokenEq
	case TokenNCont:
		positiveOp = TokenCont
	case TokenNRegex:
		positiveOp = TokenRegex
	case TokenNLike:
		positiveOp = TokenLike
	}

	for _, val := range values {
		matched, err := matchString(positiveOp, val, v)
		if err != nil {
			return false, err
		}

		if matched {
			return !negated, nil
		}
	}

	return negated, nil
}

func evalFreeText(term string, rec Record) bool {
	needle := strings.ToLower(term)

	haystacks := []string{
		rec.Request.Method, rec.Request.URL, rec.Request.Proto, rec.Request.Body,
	}
	haystacks = append(haystacks, headerValues(Field{Namespace: "req", Name: "header"}, rec)...)

	if rec.Response != nil {
		haystacks = append(haystacks, rec.Response.Proto, rec.Response.Reason, rec.Response.Body)
		haystacks = append(haystacks, headerValues(Field{Namespace: "resp", Name: "header"}, rec)...)
	}

	for _, h := range haystacks {
		if strings.Contains(strings.ToLower(h), needle) {
			return true
		}
	}

	return false
}

func (v Value) asString() string {
	switch v.Kind {
	case ValueInt:
		return strconv.FormatInt(v.Int, 10)
	case ValueBool:
		if v.Bool {
			return "true"
		}
		return "false"
	default:
		return v.Str
	}
}

// matchLike implements SQL LIKE semantics: `%` matches any run of characters and
// `_` matches a single character. Matching is case-insensitive.
func matchLike(actual, pattern string) bool {
	re := likeToRegexp(pattern)
	return re.MatchString(strings.ToLower(actual))
}

var (
	regexCache sync.Map // pattern -> *regexp.Regexp
	likeCache  sync.Map // pattern -> *regexp.Regexp
)

func compileRegexp(pattern string) (*regexp.Regexp, error) {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("httpql: invalid regular expression %q: %w", pattern, err)
	}

	regexCache.Store(pattern, re)

	return re, nil
}

func likeToRegexp(pattern string) *regexp.Regexp {
	if cached, ok := likeCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}

	var b strings.Builder
	b.WriteString("^")

	for _, r := range strings.ToLower(pattern) {
		switch r {
		case '%':
			b.WriteString(".*")
		case '_':
			b.WriteString(".")
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}

	b.WriteString("$")

	re := regexp.MustCompile(b.String())
	likeCache.Store(pattern, re)

	return re
}
