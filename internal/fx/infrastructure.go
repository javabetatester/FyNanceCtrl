package fx

import (
	"Fynance/config"
	"Fynance/internal/infrastructure"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var InfrastructureModule = fx.Module("infrastructure",
	fx.Provide(
		newDatabase,
		newUserRepository,
		newGoalRepository,
		newTransactionRepository,
		newCategoryRepository,
		newInvestmentRepository,
		newAccountRepository,
		newBudgetRepository,
		newDashboardRepository,
		newRecurringRepository,
		newReportRepository,
		newCreditCardRepository,
		newResourceCounter,
	),
)

func newDatabase(cfg *config.Config) (*gorm.DB, error) {
	return infrastructure.NewDb(cfg)
}

func newUserRepository(db *gorm.DB) *infrastructure.UserRepository {
	return &infrastructure.UserRepository{DB: db}
}

func newGoalRepository(db *gorm.DB) *infrastructure.GoalRepository {
	return &infrastructure.GoalRepository{DB: db}
}

func newTransactionRepository(db *gorm.DB) *infrastructure.TransactionRepository {
	return &infrastructure.TransactionRepository{DB: db}
}

func newCategoryRepository(db *gorm.DB) *infrastructure.TransactionCategoryRepository {
	return &infrastructure.TransactionCategoryRepository{DB: db}
}

func newInvestmentRepository(db *gorm.DB) *infrastructure.InvestmentRepository {
	return &infrastructure.InvestmentRepository{DB: db}
}

func newAccountRepository(db *gorm.DB) *infrastructure.AccountRepository {
	return &infrastructure.AccountRepository{DB: db}
}

func newBudgetRepository(db *gorm.DB) *infrastructure.BudgetRepository {
	return &infrastructure.BudgetRepository{DB: db}
}

func newDashboardRepository(db *gorm.DB) *infrastructure.DashboardRepository {
	return &infrastructure.DashboardRepository{DB: db}
}

func newRecurringRepository(db *gorm.DB) *infrastructure.RecurringRepository {
	return &infrastructure.RecurringRepository{DB: db}
}

func newReportRepository(db *gorm.DB) *infrastructure.ReportRepository {
	return &infrastructure.ReportRepository{DB: db}
}

func newCreditCardRepository(db *gorm.DB) *infrastructure.CreditCardRepository {
	return &infrastructure.CreditCardRepository{DB: db}
}

func newResourceCounter(db *gorm.DB) *infrastructure.ResourceCounter {
	return &infrastructure.ResourceCounter{DB: db}
}
