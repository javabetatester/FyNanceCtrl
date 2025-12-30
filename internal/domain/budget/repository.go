package budget

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type BudgetFilters struct {
	Search *string
}

type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, budgetID, userID ulid.ULID) error
	GetByID(ctx context.Context, budgetID, userID ulid.ULID) (*Budget, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, month, year int, filters *BudgetFilters, pagination *pkg.PaginationParams) ([]*Budget, int64, error)
	GetByCategoryID(ctx context.Context, categoryID, userID ulid.ULID, month, year int) (*Budget, error)
	UpdateSpent(ctx context.Context, budgetID ulid.ULID, amount float64) error
	GetRecurring(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Budget, int64, error)
	GetSummary(ctx context.Context, userID ulid.ULID, month, year int) (*BudgetSummary, error)
}
