package reqlog

import (
	"context"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/scope"
)

type Repository interface {
	FindRequestLogs(ctx context.Context, filter FindRequestsFilter, scope *scope.Scope) ([]RequestLog, error)
	FindRequestLogByID(ctx context.Context, projectID, id ulid.ULID) (RequestLog, error)
	StoreRequestLog(ctx context.Context, reqLog RequestLog) error
	StoreResponseLog(ctx context.Context, projectID, reqLogID ulid.ULID, resLog ResponseLog) error
	ClearRequestLogs(ctx context.Context, projectID ulid.ULID) error
	Sitemap(ctx context.Context, projectID ulid.ULID) ([]SitemapEntry, error)
}

// SitemapEntry aggregates the traffic seen for one host+path endpoint.
type SitemapEntry struct {
	Host        string
	Path        string
	Methods     []string
	StatusCodes []int
	Count       int
}
