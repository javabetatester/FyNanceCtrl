package investment

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	Create(ctx context.Context, investment *Investment) error
	List(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*Investment, int64, error)
	Update(ctx context.Context, investment *Investment) error
	Delete(ctx context.Context, id ulid.ULID, userId ulid.ULID) error
	GetInvestmentById(ctx context.Context, id ulid.ULID, userId ulid.ULID) (*Investment, error)
	GetByUserId(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*Investment, int64, error)
	GetTotalBalance(ctx context.Context, userId ulid.ULID) (float64, error)
	GetByType(ctx context.Context, userId ulid.ULID, investmentType Types, pagination *pkg.PaginationParams) ([]*Investment, int64, error)
	UpdateBalanceAtomic(ctx context.Context, investmentID ulid.ULID, delta float64) error
}
