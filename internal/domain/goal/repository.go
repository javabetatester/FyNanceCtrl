package goal

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type GoalFilters struct {
	Status *GoalStatus
}

type Repository interface {
	Create(ctx context.Context, goal *Goal) error
	List(ctx context.Context, pagination *pkg.PaginationParams) ([]*Goal, int64, error)
	Update(ctx context.Context, goal *Goal) error
	UpdateFields(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID) (*Goal, error)
	GetByIDAndUser(ctx context.Context, id, userID ulid.ULID) (*Goal, error)
	GetByUserID(ctx context.Context, userId ulid.ULID, filters *GoalFilters, pagination *pkg.PaginationParams) ([]*Goal, int64, error)
	CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error)
	CreateContribution(ctx context.Context, contribution *Contribution) error
	GetContributionsByGoalID(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*Contribution, error)
	GetContributionByID(ctx context.Context, contributionId ulid.ULID, userId ulid.ULID) (*Contribution, error)
	GetContributionByTransactionID(ctx context.Context, transactionId ulid.ULID, userId ulid.ULID) (*Contribution, error)
	DeleteContribution(ctx context.Context, contributionId ulid.ULID) error
	UpdateCurrentAmount(ctx context.Context, goalId ulid.ULID, amount float64) error
	UpdateCurrentAmountAtomic(ctx context.Context, goalId ulid.ULID, delta float64) error
}
