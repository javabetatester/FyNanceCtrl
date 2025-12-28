package goal

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	Create(ctx context.Context, goal *Goal) error
	List(ctx context.Context, pagination *pkg.PaginationParams) ([]*Goal, int64, error)
	Update(ctx context.Context, goal *Goal) error
	UpdateFields(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetById(ctx context.Context, id ulid.ULID) (*Goal, error)
	GetByIdAndUser(ctx context.Context, id, userID ulid.ULID) (*Goal, error)
	GetByUserId(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*Goal, int64, error)
	CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error)
	CreateContribution(ctx context.Context, contribution *Contribution) error
	GetContributionsByGoalId(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*Contribution, error)
	UpdateCurrentAmount(ctx context.Context, goalId ulid.ULID, amount float64) error
	UpdateCurrentAmountAtomic(ctx context.Context, goalId ulid.ULID, delta float64) error
}
