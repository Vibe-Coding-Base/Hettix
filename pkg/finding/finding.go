// Package finding records security findings for a project: a titled, severity-
// rated note optionally linked to the request that evidences it. Findings are
// created by the operator or by the AI assistant as it investigates traffic.
package finding

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
)

// Severity rates a finding's impact.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// ParseSeverity normalizes a severity string, defaulting to info.
func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low":
		return SeverityLow
	case "medium":
		return SeverityMedium
	case "high":
		return SeverityHigh
	case "critical":
		return SeverityCritical
	default:
		return SeverityInfo
	}
}

var (
	ErrProjectIDMustBeSet = errors.New("finding: project ID must be set")
	ErrFindingNotFound    = errors.New("finding: finding not found")
	ErrTitleRequired      = errors.New("finding: title is required")
)

// Finding is a recorded security observation.
type Finding struct {
	ID           ulid.ULID
	ProjectID    ulid.ULID
	Title        string
	Description  string
	Severity     Severity
	RequestLogID *ulid.ULID
	CreatedAt    time.Time
}

// Repository persists findings.
type Repository interface {
	StoreFinding(ctx context.Context, finding Finding) error
	FindFindings(ctx context.Context, projectID ulid.ULID) ([]Finding, error)
	FindFindingByID(ctx context.Context, projectID, id ulid.ULID) (Finding, error)
	DeleteFinding(ctx context.Context, projectID, id ulid.ULID) error
}

// Service manages findings for the active project.
type Service struct {
	activeProjectID ulid.ULID
	repo            Repository
	logger          log.Logger

	entropy   *rand.Rand
	entropyMu sync.Mutex
}

type Config struct {
	Repository Repository
	Logger     log.Logger
}

func NewService(cfg Config) *Service {
	s := &Service{
		repo:   cfg.Repository,
		logger: cfg.Logger,
		//nolint:gosec // finding IDs don't need cryptographic randomness
		entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if s.logger == nil {
		s.logger = log.NewNopLogger()
	}

	return s
}

func (svc *Service) SetActiveProjectID(id ulid.ULID) {
	svc.activeProjectID = id
}

func (svc *Service) newID() ulid.ULID {
	svc.entropyMu.Lock()
	defer svc.entropyMu.Unlock()

	return ulid.MustNew(ulid.Timestamp(time.Now()), svc.entropy)
}

// CreateFinding records a new finding for the active project.
func (svc *Service) CreateFinding(
	ctx context.Context,
	title, description string,
	severity Severity,
	requestLogID *ulid.ULID,
) (Finding, error) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return Finding{}, ErrProjectIDMustBeSet
	}

	if strings.TrimSpace(title) == "" {
		return Finding{}, ErrTitleRequired
	}

	finding := Finding{
		ID:           svc.newID(),
		ProjectID:    projectID,
		Title:        title,
		Description:  description,
		Severity:     severity,
		RequestLogID: requestLogID,
		CreatedAt:    time.Now(),
	}

	if err := svc.repo.StoreFinding(ctx, finding); err != nil {
		return Finding{}, err
	}

	return finding, nil
}

func (svc *Service) Findings(ctx context.Context) ([]Finding, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindFindings(ctx, svc.activeProjectID)
}

func (svc *Service) FindingByID(ctx context.Context, id ulid.ULID) (Finding, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return Finding{}, ErrProjectIDMustBeSet
	}

	return svc.repo.FindFindingByID(ctx, svc.activeProjectID, id)
}

func (svc *Service) DeleteFinding(ctx context.Context, id ulid.ULID) error {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return ErrProjectIDMustBeSet
	}

	return svc.repo.DeleteFinding(ctx, svc.activeProjectID, id)
}
