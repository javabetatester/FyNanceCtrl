package dashboard

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository Repository
}

func (s *Service) GetDashboard(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) (*DashboardResponse, error) {
	if month <= 0 || month > 12 {
		month = int(time.Now().Month())
	}
	if year <= 0 {
		year = time.Now().Year()
	}

	summary, err := s.Repository.GetFinancialSummary(ctx, userID, accountID, month, year)
	if err != nil {
		return nil, err
	}

	monthlyTrend, err := s.Repository.GetMonthlyTrend(ctx, userID, accountID, 6)
	if err != nil {
		return nil, err
	}

	categoryExpenses, err := s.Repository.GetExpensesByCategory(ctx, userID, accountID, month, year)
	if err != nil {
		return nil, err
	}

	recentTransactions, err := s.Repository.GetRecentTransactions(ctx, userID, accountID, 5)
	if err != nil {
		return nil, err
	}

	goals, err := s.Repository.GetActiveGoals(ctx, userID)
	if err != nil {
		return nil, err
	}

	budgetStatus, err := s.Repository.GetBudgetStatus(ctx, userID, accountID, month, year)
	if err != nil {
		return nil, err
	}

	accounts, err := s.Repository.GetAccountsSummary(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		Summary:            summary,
		MonthlyTrend:       monthlyTrend,
		CategoryExpenses:   categoryExpenses,
		RecentTransactions: recentTransactions,
		Goals:              goals,
		BudgetStatus:       budgetStatus,
		Accounts:           accounts,
	}, nil
}

type DashboardResponse struct {
	Summary            *FinancialSummary     `json:"summary"`
	MonthlyTrend       []*MonthlyTrendItem   `json:"monthlyTrend"`
	CategoryExpenses   []*CategoryExpense    `json:"categoryExpenses"`
	RecentTransactions []*TransactionSummary `json:"recentTransactions"`
	Goals              []*GoalSummary        `json:"goals"`
	BudgetStatus       []*BudgetStatusItem   `json:"budgetStatus"`
	Accounts           []*AccountSummary     `json:"accounts"`
}

type FinancialSummary struct {
	TotalBalance     float64 `json:"totalBalance"`
	MonthIncome      float64 `json:"monthIncome"`
	MonthExpenses    float64 `json:"monthExpenses"`
	MonthBalance     float64 `json:"monthBalance"`
	TotalInvestments float64 `json:"totalInvestments"`
	TotalGoals       float64 `json:"totalGoals"`
}

type MonthlyTrendItem struct {
	Month    string  `json:"month"`
	Year     int     `json:"year"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
}

type CategoryExpense struct {
	CategoryId   ulid.ULID `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Amount       float64   `json:"amount"`
	Percentage   float64   `json:"percentage"`
}

type TransactionSummary struct {
	Id          ulid.ULID `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CategoryId  ulid.ULID `json:"categoryId"`
	Date        time.Time `json:"date"`
}

type GoalSummary struct {
	Id            ulid.ULID `json:"id"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"targetAmount"`
	CurrentAmount float64   `json:"currentAmount"`
	Percentage    float64   `json:"percentage"`
	Status        string    `json:"status"`
}

type BudgetStatusItem struct {
	CategoryId   ulid.ULID `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	BudgetAmount float64   `json:"budgetAmount"`
	SpentAmount  float64   `json:"spentAmount"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"`
	Status       string    `json:"status"`
}

type AccountSummary struct {
	Id      ulid.ULID `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Balance float64   `json:"balance"`
	Color   string    `json:"color"`
}
