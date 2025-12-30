package account

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, accountID, userID ulid.ULID) error
	GetByID(ctx context.Context, accountID, userID ulid.ULID) (*Account, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Account, int64, error)
	GetActiveByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Account, int64, error)
	GetByCreditCardID(ctx context.Context, creditCardID, userID ulid.ULID) (*Account, error)
	UpdateBalance(ctx context.Context, accountID ulid.ULID, amount float64) error
	UpdateBalanceWithTx(ctx context.Context, tx interface{}, accountID ulid.ULID, amount float64) error
	GetTotalBalance(ctx context.Context, userID ulid.ULID) (float64, error)
	BeginTx(ctx context.Context) (interface{}, error)
	CommitTx(tx interface{}) error
	RollbackTx(tx interface{}) error
}
