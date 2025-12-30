package category

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type CategoryServiceInterface interface {
	ValidateAndEnsureExists(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error
	CreateDefaultCategories(ctx context.Context, userID ulid.ULID) error
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error
	GetByID(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (*Category, error)
	GetAll(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Category, int64, error)
}
