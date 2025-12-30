package fx

import (
	"time"

	"Fynance/internal/domain/account"
	"Fynance/internal/domain/auth"
	"Fynance/internal/domain/budget"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/dashboard"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/healthscore"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/report"
	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	"Fynance/internal/infrastructure"
	"Fynance/internal/middleware"
	"Fynance/internal/routes"

	"go.uber.org/fx"
)

// RoutesModule fornece handlers e rate limiters
var RoutesModule = fx.Module("routes",
	fx.Provide(
		newHandler,
		newAuthRateLimiter,
		newHealthScoreService,
		newHealthScoreHandler,
	),
)

func newHandler(
	userSvc *user.Service,
	jwtSvc *middleware.JwtService,
	authSvc *auth.Service,
	goalSvc *goal.Service,
	transactionSvc *transaction.Service,
	investmentSvc *investment.Service,
	accountSvc *account.Service,
	budgetSvc *budget.Service,
	dashboardSvc dashboard.Service,
	recurringSvc *recurring.Service,
	reportSvc report.Service,
	creditCardSvc creditcard.Service,
	accountRepo *infrastructure.AccountRepository,
	transactionRepo *infrastructure.TransactionRepository,
	goalRepo *infrastructure.GoalRepository,
	budgetRepo *infrastructure.BudgetRepository,
	investmentRepo *infrastructure.InvestmentRepository,
	recurringRepo *infrastructure.RecurringRepository,
	creditCardRepo *infrastructure.CreditCardRepository,
	categoryRepo *infrastructure.TransactionCategoryRepository,
) *routes.Handler {
	return &routes.Handler{
		UserService:        *userSvc,
		JwtService:         jwtSvc,
		AuthService:        *authSvc,
		GoalService:        *goalSvc,
		TransactionService: *transactionSvc,
		InvestmentService:  *investmentSvc,
		AccountService:     *accountSvc,
		BudgetService:      *budgetSvc,
		DashboardService:   dashboardSvc,
		RecurringService:   *recurringSvc,
		ReportService:      reportSvc,
		CreditCardService:  creditCardSvc,

		AccountRepository:     accountRepo,
		TransactionRepository: transactionRepo,
		GoalRepository:        goalRepo,
		BudgetRepository:      budgetRepo,
		InvestmentRepository:  investmentRepo,
		RecurringRepository:   recurringRepo,
		CreditCardRepository:  creditCardRepo,
		CategoryRepository:    categoryRepo,
	}
}

func newAuthRateLimiter() *middleware.RateLimiter {
	return middleware.NewRateLimiter(100, time.Minute)
}

func newHealthScoreService() *healthscore.Service {
	return healthscore.NewService()
}

func newHealthScoreHandler(svc *healthscore.Service) *routes.HealthScoreHandler {
	return routes.NewHealthScoreHandler(svc)
}
