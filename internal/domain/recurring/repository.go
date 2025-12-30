package recurring

import (
	"context"
	"time"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type RecurringRepository interface {
	Create(ctx context.Context, recurring *RecurringTransaction) error
	Update(ctx context.Context, recurring *RecurringTransaction) error
	Delete(ctx context.Context, recurringID, userID ulid.ULID) error
	GetByID(ctx context.Context, recurringID, userID ulid.ULID) (*RecurringTransaction, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*RecurringTransaction, int64, error)
	GetActiveByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*RecurringTransaction, int64, error)
	GetDueTransactions(ctx context.Context, date time.Time, pagination *pkg.PaginationParams) ([]*RecurringTransaction, int64, error)
	UpdateLastProcessed(ctx context.Context, recurringID ulid.ULID, processedDate, nextDue time.Time) error
}
