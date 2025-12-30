package fx

import (
	"Fynance/config"
	"Fynance/internal/domain/account"
	"Fynance/internal/domain/auth"
	"Fynance/internal/domain/budget"
	"Fynance/internal/domain/category"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/dashboard"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/report"
	"Fynance/internal/domain/shared"
	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	"Fynance/internal/infrastructure"
	"Fynance/internal/logger"

	"go.uber.org/fx"
)

// DomainModule fornece todos os services do domínio
var DomainModule = fx.Module("domain",
	fx.Provide(
		// User services
		newUserService,
		newUserServiceAdapter,
		newUserCheckerService,

		// Category service
		newCategoryService,

		// Account service
		newAccountService,

		// Auth service (requer GoogleClientID)
		newGoogleClientID,
		newAuthService,

		// Budget service
		newBudgetService,

		// Investment service
		newInvestmentService,

		// Goal service (inicialmente com nil TransactionService, será atualizado depois)
		newGoalService,

		// Transaction service (depende de vários serviços)
		newTransactionService,

		// Recurring service
		newRecurringService,

		// Dashboard service
		newDashboardService,

		// Report service
		newReportService,

		// CreditCard service
		newCreditCardService,
	),
	fx.Invoke(
		// Atualizar GoalService com TransactionService após ambos serem criados
		updateGoalServiceWithTransactionService,
	),
)

// updateGoalServiceWithTransactionService atualiza o GoalService com TransactionService
// Esta função é chamada após todos os serviços serem criados
func updateGoalServiceWithTransactionService(
	goalSvc *goal.Service,
	transactionSvc *transaction.Service,
) {
	goalSvc.TransactionService = transactionSvc
}

func newUserService(repo *infrastructure.UserRepository) *user.Service {
	return user.NewService(repo)
}

func newUserServiceAdapter(userSvc *user.Service) *user.UserServiceAdapter {
	return user.NewUserServiceAdapter(userSvc)
}

func newUserCheckerService(adapter *user.UserServiceAdapter) *shared.UserCheckerService {
	return shared.NewUserCheckerService(adapter)
}

func newCategoryService(
	repo *infrastructure.TransactionCategoryRepository,
	userChecker *shared.UserCheckerService,
) *category.Service {
	return category.NewService(repo, userChecker)
}

func newAccountService(
	repo *infrastructure.AccountRepository,
	userChecker *shared.UserCheckerService,
) *account.Service {
	return account.NewService(repo, userChecker)
}

func newGoogleClientID(cfg *config.Config) string {
	googleClientID := ""
	if cfg.GoogleOAuth.Enabled {
		if cfg.GoogleOAuth.ClientID == "" {
			logger.Warn().
				Msg("GOOGLE_OAUTH_ENABLED=true mas GOOGLE_OAUTH_CLIENT_ID está vazio. Verifique se a variável está definida no arquivo .env")
		} else {
			googleClientID = cfg.GoogleOAuth.ClientID
			clientIDPreview := googleClientID
			if len(clientIDPreview) > 20 {
				clientIDPreview = clientIDPreview[:20] + "..."
			}
			logger.Info().
				Str("client_id_preview", clientIDPreview).
				Int("client_id_length", len(googleClientID)).
				Msg("Google OAuth habilitado - Certifique-se de que este Client ID está autorizado no Google Console e corresponde ao usado no frontend")
		}
	} else {
		logger.Info().Msg("Google OAuth desabilitado (GOOGLE_OAUTH_ENABLED não está definido como 'true')")
	}
	return googleClientID
}

func newAuthService(
	repo *infrastructure.UserRepository,
	userSvc *user.Service,
	googleClientID string,
) *auth.Service {
	return auth.NewService(repo, userSvc, googleClientID)
}

func newBudgetService(
	repo *infrastructure.BudgetRepository,
	categorySvc *category.Service,
	userChecker *shared.UserCheckerService,
) *budget.Service {
	return budget.NewService(repo, categorySvc, userChecker)
}

func newInvestmentService(
	repo *infrastructure.InvestmentRepository,
	transactionRepo *infrastructure.TransactionRepository,
	accountSvc *account.Service,
	userChecker *shared.UserCheckerService,
) *investment.Service {
	return investment.NewService(repo, transactionRepo, accountSvc, userChecker)
}

func newGoalService(
	repo *infrastructure.GoalRepository,
	accountSvc *account.Service,
	userChecker *shared.UserCheckerService,
) *goal.Service {
	// Inicialmente com nil para TransactionService, será atualizado depois
	return goal.NewService(repo, accountSvc, nil, userChecker)
}

func newTransactionService(
	repo *infrastructure.TransactionRepository,
	categorySvc *category.Service,
	accountSvc *account.Service,
	budgetSvc *budget.Service,
	goalSvc *goal.Service,
	investmentSvc *investment.Service,
	userChecker *shared.UserCheckerService,
) *transaction.Service {
	return transaction.NewService(
		repo,
		categorySvc,
		accountSvc,
		budgetSvc,
		goalSvc,
		investmentSvc,
		userChecker,
	)
}

func newRecurringService(
	repo *infrastructure.RecurringRepository,
	transactionRepo *infrastructure.TransactionRepository,
	categorySvc *category.Service,
	transactionSvc *transaction.Service,
	userChecker *shared.UserCheckerService,
) *recurring.Service {
	return recurring.NewService(repo, transactionRepo, categorySvc, transactionSvc, userChecker)
}

func newDashboardService(repo *infrastructure.DashboardRepository) dashboard.Service {
	return dashboard.Service{
		Repository: repo,
	}
}

func newReportService(
	repo *infrastructure.ReportRepository,
	userSvc *user.Service,
) report.Service {
	return report.Service{
		Repository:  repo,
		UserService: userSvc,
	}
}

func newCreditCardService(
	repo *infrastructure.CreditCardRepository,
	accountSvc *account.Service,
	userSvc *user.Service,
) creditcard.Service {
	return creditcard.Service{
		Repository:     repo,
		AccountService: accountSvc,
		UserService:    userSvc,
	}
}
