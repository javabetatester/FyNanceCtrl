package category

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error
	GetByID(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (*Category, error)
	GetAll(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Category, int64, error)
	GetAllWithoutLimit(ctx context.Context, userID ulid.ULID) ([]*Category, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Category, int64, error)
	BelongsToUser(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (bool, error)
	GetByName(ctx context.Context, categoryName string, userID ulid.ULID) (*Category, error)
}
