package transaction

import (
	"context"
	"time"

	"Fynance/internal/domain/category"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type TransactionFilters struct {
	Type       *string
	CategoryID *ulid.ULID
	Search     *string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, transactionID ulid.ULID) error
	GetByID(ctx context.Context, transactionID ulid.ULID) (*Transaction, error)
	GetByIDAndUser(ctx context.Context, transactionID, userID ulid.ULID) (*Transaction, error)
	GetAll(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, filters *TransactionFilters, pagination *pkg.PaginationParams) ([]*Transaction, int64, error)
	GetByAmount(ctx context.Context, amount float64, pagination *pkg.PaginationParams) ([]*Transaction, int64, error)
	GetByName(ctx context.Context, name string, pagination *pkg.PaginationParams) ([]*Transaction, int64, error)
	GetByCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Transaction, int64, error)
	GetByInvestmentID(ctx context.Context, investmentID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Transaction, int64, error)
	GetNumberOfTransactions(ctx context.Context, userID ulid.ULID) (int64, error)
}

type CategoryRepository = category.CategoryRepository
