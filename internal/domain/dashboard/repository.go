package dashboard

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	GetFinancialSummary(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) (*FinancialSummary, error)
	GetMonthlyTrend(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, months int) ([]*MonthlyTrendItem, error)
	GetExpensesByCategory(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) ([]*CategoryExpense, error)
	GetRecentTransactions(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, limit int) ([]*TransactionSummary, error)
	GetActiveGoals(ctx context.Context, userID ulid.ULID) ([]*GoalSummary, error)
	GetBudgetStatus(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) ([]*BudgetStatusItem, error)
	GetAccountsSummary(ctx context.Context, userID ulid.ULID) ([]*AccountSummary, error)
}
