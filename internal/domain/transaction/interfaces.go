package transaction

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type TransactionHandler interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) error
}
