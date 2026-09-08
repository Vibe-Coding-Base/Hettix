package intruder

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"

	"github.com/Vibe-Coding-Base/Hettix/pkg/log"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
)

var (
	ErrProjectIDMustBeSet = errors.New("intruder: project ID must be set")
	ErrNoMarker           = errors.New("intruder: request template has no insertion marker (§)")
	ErrNoPayloads         = errors.New("intruder: at least one payload is required")
	ErrAttackNotFound     = errors.New("intruder: attack not found")
)

// Attack is a fuzzing run: a base request sent once per payload.
type Attack struct {
	ID        ulid.ULID
	ProjectID ulid.ULID
	Name      string
	Status    string
	Total     int
	Completed int
	CreatedAt time.Time
}

// Repository persists attacks and their results.
type Repository interface {
	StoreAttack(ctx context.Context, attack Attack) error
	UpdateAttackStatus(ctx context.Context, projectID, id ulid.ULID, status string) error
	StoreResult(ctx context.Context, attackID ulid.ULID, result Result) error
	FindAttacks(ctx context.Context, projectID ulid.ULID) ([]Attack, error)
	FindAttackByID(ctx context.Context, projectID, id ulid.ULID) (Attack, error)
	FindResults(ctx context.Context, projectID, attackID ulid.ULID) ([]Result, error)
	DeleteAttacks(ctx context.Context, projectID ulid.ULID) error
}

// Service coordinates fuzzing attacks for the active project.
type Service struct {
	activeProjectID ulid.ULID
	repo            Repository
	runner          *Runner
	logger          log.Logger

	entropy   *rand.Rand
	entropyMu sync.Mutex
}

type Config struct {
	Repository  Repository
	Runner      *Runner
	Logger      log.Logger
	Concurrency int
}

func NewService(cfg Config) *Service {
	s := &Service{
		repo:   cfg.Repository,
		runner: cfg.Runner,
		logger: cfg.Logger,
		//nolint:gosec // attack IDs don't need cryptographic randomness
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

// StartAttack persists a new attack and runs it in the background, storing each
// result as it completes. It returns the attack in its initial running state.
func (svc *Service) StartAttack(ctx context.Context, name string, tmpl Request, payloads []string) (Attack, error) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return Attack{}, ErrProjectIDMustBeSet
	}

	if !tmpl.HasMarker() {
		return Attack{}, ErrNoMarker
	}

	if len(payloads) == 0 {
		return Attack{}, ErrNoPayloads
	}

	attack := Attack{
		ID:        svc.newID(),
		ProjectID: projectID,
		Name:      name,
		Status:    StatusRunning,
		Total:     len(payloads),
		CreatedAt: time.Now(),
	}

	if err := svc.repo.StoreAttack(ctx, attack); err != nil {
		return Attack{}, err
	}

	go svc.run(attack, tmpl, payloads)

	return attack, nil
}

func (svc *Service) run(attack Attack, tmpl Request, payloads []string) {
	ctx := context.Background()

	svc.runner.Run(ctx, tmpl, payloads, func(res Result) {
		if err := svc.repo.StoreResult(ctx, attack.ID, res); err != nil {
			svc.logger.Errorw("Failed to store attack result.", "error", err, "attackID", attack.ID.String())
		}
	})

	if err := svc.repo.UpdateAttackStatus(ctx, attack.ProjectID, attack.ID, StatusCompleted); err != nil {
		svc.logger.Errorw("Failed to mark attack completed.", "error", err, "attackID", attack.ID.String())
	}
}

func (svc *Service) Attacks(ctx context.Context) ([]Attack, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindAttacks(ctx, svc.activeProjectID)
}

func (svc *Service) AttackByID(ctx context.Context, id ulid.ULID) (Attack, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return Attack{}, ErrProjectIDMustBeSet
	}

	return svc.repo.FindAttackByID(ctx, svc.activeProjectID, id)
}

func (svc *Service) Results(ctx context.Context, attackID ulid.ULID) ([]Result, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindResults(ctx, svc.activeProjectID, attackID)
}
