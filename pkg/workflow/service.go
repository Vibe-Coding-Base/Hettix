package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
)

// Repository persists workflows.
type Repository interface {
	StoreWorkflow(ctx context.Context, wf Workflow) error
	FindWorkflows(ctx context.Context, projectID ulid.ULID) ([]Workflow, error)
	FindWorkflowByID(ctx context.Context, projectID, id ulid.ULID) (Workflow, error)
	DeleteWorkflow(ctx context.Context, projectID, id ulid.ULID) error
}

// Service manages and runs workflows for the active project.
type Service struct {
	activeProjectID ulid.ULID
	repo            Repository
	runner          *Runner
}

type Config struct {
	Repository Repository
	Runner     *Runner
}

func NewService(cfg Config) *Service {
	return &Service{repo: cfg.Repository, runner: cfg.Runner}
}

func (svc *Service) SetActiveProjectID(id ulid.ULID) {
	svc.activeProjectID = id
}

// Save creates or replaces a workflow. When id is nil a new workflow is created.
func (svc *Service) Save(ctx context.Context, id *ulid.ULID, name string, steps []Step) (Workflow, error) {
	projectID := svc.activeProjectID
	if projectID.Compare(ulid.ULID{}) == 0 {
		return Workflow{}, ErrProjectIDMustBeSet
	}

	if strings.TrimSpace(name) == "" {
		return Workflow{}, ErrNameRequired
	}

	for _, step := range steps {
		if !step.Type.valid() {
			return Workflow{}, fmt.Errorf("%w: %q", ErrUnknownStep, step.Type)
		}
	}

	wf := Workflow{
		ProjectID: projectID,
		Name:      name,
		Steps:     steps,
		CreatedAt: time.Now(),
	}

	if id != nil {
		existing, err := svc.repo.FindWorkflowByID(ctx, projectID, *id)
		if err != nil {
			return Workflow{}, err
		}

		wf.ID = existing.ID
		wf.CreatedAt = existing.CreatedAt
	} else {
		wf.ID = newID()
	}

	if err := svc.repo.StoreWorkflow(ctx, wf); err != nil {
		return Workflow{}, err
	}

	return wf, nil
}

func (svc *Service) Workflows(ctx context.Context) ([]Workflow, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return nil, ErrProjectIDMustBeSet
	}

	return svc.repo.FindWorkflows(ctx, svc.activeProjectID)
}

func (svc *Service) WorkflowByID(ctx context.Context, id ulid.ULID) (Workflow, error) {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return Workflow{}, ErrProjectIDMustBeSet
	}

	return svc.repo.FindWorkflowByID(ctx, svc.activeProjectID, id)
}

func (svc *Service) Delete(ctx context.Context, id ulid.ULID) error {
	if svc.activeProjectID.Compare(ulid.ULID{}) == 0 {
		return ErrProjectIDMustBeSet
	}

	return svc.repo.DeleteWorkflow(ctx, svc.activeProjectID, id)
}

// Run executes a stored workflow and returns a result per step.
func (svc *Service) Run(ctx context.Context, id ulid.ULID) ([]StepResult, error) {
	wf, err := svc.WorkflowByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return svc.runner.Run(ctx, wf.Steps), nil
}
