package intercept

import "github.com/Vibe-Coding-Base/Hettix/pkg/filter"

type Settings struct {
	RequestsEnabled  bool
	ResponsesEnabled bool
	RequestFilter    filter.Expression
	ResponseFilter   filter.Expression
}
