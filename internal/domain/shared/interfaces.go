package shared

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type UserChecker interface {
	Exists(ctx context.Context, userID ulid.ULID) error
}

type UserGetter interface {
	UserChecker
	GetByID(ctx context.Context, userID ulid.ULID) (interface{}, error)
}

type BalanceUpdater interface {
	UpdateBalance(ctx context.Context, accountID, userID ulid.ULID, amount float64) error
}

type AccountGetter interface {
	GetAccountByID(ctx context.Context, accountID, userID ulid.ULID) (interface{}, error)
}

type AccountService interface {
	AccountGetter
	BalanceUpdater
}

type BudgetUpdater interface {
	UpdateSpent(ctx context.Context, categoryID, userID ulid.ULID, amount float64) error
	UpdateSpentWithDate(ctx context.Context, categoryID, userID ulid.ULID, amount float64, transactionDate interface{}) error
}
	
type TransactionCreator interface {
	CreateTransaction(ctx context.Context, transaction interface{}) error
}

type TransactionDeleter interface {
	DeleteTransaction(ctx context.Context, transactionID, userID ulid.ULID) error
}

type GoalContributionDeleter interface {
	DeleteContributionByTransactionId(ctx context.Context, transactionID, userID ulid.ULID) error
}

type InvestmentTransactionDeleter interface {
	DeleteInvestmentTransactionByTransactionId(ctx context.Context, transactionID, userID ulid.ULID) error
}
