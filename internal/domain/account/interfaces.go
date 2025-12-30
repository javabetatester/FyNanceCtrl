package account

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type AccountServiceInterface interface {
	GetAccountByID(ctx context.Context, accountID, userID ulid.ULID) (*Account, error)
	UpdateBalance(ctx context.Context, accountID, userID ulid.ULID, amount float64) error
}
