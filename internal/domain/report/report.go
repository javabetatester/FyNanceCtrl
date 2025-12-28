package report

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type MonthlyReport struct {
	UserId             ulid.ULID         `json:"userId"`
	Month              int               `json:"month"`
	Year               int               `json:"year"`
	TotalIncome        float64           `json:"totalIncome"`
	TotalExpenses      float64           `json:"totalExpenses"`
	NetBalance         float64           `json:"netBalance"`
	SavingsRate        float64           `json:"savingsRate"`
	IncomeByCategory   []CategoryAmount  `json:"incomeByCategory"`
	ExpensesByCategory []CategoryAmount  `json:"expensesByCategory"`
	DailyBalance       []DailyBalance    `json:"dailyBalance"`
	TopExpenses        []TransactionItem `json:"topExpenses"`
	Comparison         *MonthComparison  `json:"comparison"`
}

type CategoryAmount struct {
	CategoryId   ulid.ULID `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Amount       float64   `json:"amount"`
	Percentage   float64   `json:"percentage"`
	Count        int       `json:"count"`
}

type DailyBalance struct {
	Date     string  `json:"date"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
}

type TransactionItem struct {
	Id          ulid.ULID `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
}

type MonthComparison struct {
	PreviousIncome     float64 `json:"previousIncome"`
	PreviousExpenses   float64 `json:"previousExpenses"`
	IncomeChange       float64 `json:"incomeChange"`
	ExpensesChange     float64 `json:"expensesChange"`
	IncomeChangePerc   float64 `json:"incomeChangePerc"`
	ExpensesChangePerc float64 `json:"expensesChangePerc"`
}

type YearlyReport struct {
	UserId           ulid.ULID        `json:"userId"`
	Year             int              `json:"year"`
	TotalIncome      float64          `json:"totalIncome"`
	TotalExpenses    float64          `json:"totalExpenses"`
	NetBalance       float64          `json:"netBalance"`
	AverageSavings   float64          `json:"averageSavings"`
	MonthlyBreakdown []MonthSummary   `json:"monthlyBreakdown"`
	TopCategories    []CategoryAmount `json:"topCategories"`
}

type MonthSummary struct {
	Month    int     `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Balance  float64 `json:"balance"`
}

type CategoryReport struct {
	CategoryId   ulid.ULID         `json:"categoryId"`
	CategoryName string            `json:"categoryName"`
	TotalAmount  float64           `json:"totalAmount"`
	Count        int               `json:"count"`
	Average      float64           `json:"average"`
	Transactions []TransactionItem `json:"transactions"`
	MonthlyTrend []MonthSummary    `json:"monthlyTrend"`
}
