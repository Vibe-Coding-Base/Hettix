package intercept

import "github.com/Vibe-Coding-Base/Hettix/pkg/httpql"

type Settings struct {
	RequestsEnabled  bool
	ResponsesEnabled bool
	RequestFilter    httpql.Expression
	ResponseFilter   httpql.Expression
}
