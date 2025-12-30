package main

import (
	"os"
	"time"

	"Fynance/config"
	"Fynance/internal/domain/account"
	"Fynance/internal/domain/auth"
	"Fynance/internal/domain/budget"
	"Fynance/internal/domain/category"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/dashboard"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/healthscore"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/report"
	"Fynance/internal/domain/shared"
	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	"Fynance/internal/infrastructure"
	"Fynance/internal/logger"
	"Fynance/internal/middleware"
	"Fynance/internal/routes"

	docs "Fynance/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func LoadConfigProvider() (*config.Config, error) {
	return config.Load()
}

func NewDBProvider(cfg *config.Config) (*gorm.DB, error) {
	return infrastructure.NewDb(cfg)
}

func RepositoryProviders() fx.Option {
	return fx.Provide(
		func(db *gorm.DB) *infrastructure.UserRepository {
			return &infrastructure.UserRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.GoalRepository {
			return &infrastructure.GoalRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.TransactionRepository {
			return &infrastructure.TransactionRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.TransactionCategoryRepository {
			return &infrastructure.TransactionCategoryRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.InvestmentRepository {
			return &infrastructure.InvestmentRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.AccountRepository {
			return &infrastructure.AccountRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.BudgetRepository {
			return &infrastructure.BudgetRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.DashboardRepository {
			return &infrastructure.DashboardRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.RecurringRepository {
			return &infrastructure.RecurringRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.ReportRepository {
			return &infrastructure.ReportRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.CreditCardRepository {
			return &infrastructure.CreditCardRepository{DB: db}
		},
		func(db *gorm.DB) *infrastructure.ResourceCounter {
			return &infrastructure.ResourceCounter{DB: db}
		},
	)
}

func ServiceProviders() fx.Option {
	return fx.Provide(
		// UserService
		func(userRepo *infrastructure.UserRepository) *user.Service {
			return user.NewService(userRepo)
		},
		// UserChecker
		func(userService *user.Service) *shared.UserCheckerService {
			userAdapter := user.NewUserServiceAdapter(userService)
			return shared.NewUserCheckerService(userAdapter)
		},
		// CategoryService
		func(
			categoryRepo *infrastructure.TransactionCategoryRepository,
			userChecker *shared.UserCheckerService,
		) *category.Service {
			return category.NewService(categoryRepo, userChecker)
		},
		// AccountService
		func(
			accountRepo *infrastructure.AccountRepository,
			userChecker *shared.UserCheckerService,
		) *account.Service {
			return account.NewService(accountRepo, userChecker)
		},
		// AuthService
		func(
			userRepo *infrastructure.UserRepository,
			userService *user.Service,
			cfg *config.Config,
		) *auth.Service {
			googleClientID := ""
			envClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
			envEnabled := os.Getenv("GOOGLE_OAUTH_ENABLED")

			logger.Info().
				Str("env_google_oauth_enabled", envEnabled).
				Str("env_google_oauth_client_id_preview", func() string {
					if len(envClientID) > 20 {
						return envClientID[:20] + "..."
					}
					return envClientID
				}()).
				Int("env_client_id_length", len(envClientID)).
				Bool("config_google_oauth_enabled", cfg.GoogleOAuth.Enabled).
				Str("config_client_id_preview", func() string {
					if len(cfg.GoogleOAuth.ClientID) > 20 {
						return cfg.GoogleOAuth.ClientID[:20] + "..."
					}
					return cfg.GoogleOAuth.ClientID
				}()).
				Int("config_client_id_length", len(cfg.GoogleOAuth.ClientID)).
				Msg("Debug: Configuração Google OAuth")

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

			return auth.NewService(userRepo, userService, googleClientID)
		},
		// BudgetService
		func(
			budgetRepo *infrastructure.BudgetRepository,
			categoryService *category.Service,
			userChecker *shared.UserCheckerService,
		) *budget.Service {
			return budget.NewService(budgetRepo, categoryService, userChecker)
		},
		// InvestmentService
		func(
			investmentRepo *infrastructure.InvestmentRepository,
			transactionRepo *infrastructure.TransactionRepository,
			accountService *account.Service,
			userChecker *shared.UserCheckerService,
		) *investment.Service {
			return investment.NewService(investmentRepo, transactionRepo, accountService, userChecker)
		},
		// GoalService (sem TransactionService inicialmente, será atualizado depois)
		func(
			goalRepo *infrastructure.GoalRepository,
			accountService *account.Service,
			userChecker *shared.UserCheckerService,
		) *goal.Service {
			return goal.NewService(goalRepo, accountService, nil, userChecker)
		},
		// TransactionService
		func(
			transactionRepo *infrastructure.TransactionRepository,
			categoryService *category.Service,
			accountService *account.Service,
			budgetService *budget.Service,
			goalService *goal.Service,
			investmentService *investment.Service,
			userChecker *shared.UserCheckerService,
		) *transaction.Service {
			service := transaction.NewService(
				transactionRepo,
				categoryService,
				accountService,
				budgetService,
				goalService,
				investmentService,
				userChecker,
			)
			// Atualizar GoalService com TransactionService
			goalService.TransactionService = service
			return service
		},
		// DashboardService
		func(dashboardRepo *infrastructure.DashboardRepository) *dashboard.Service {
			return &dashboard.Service{
				Repository: dashboardRepo,
			}
		},
		// RecurringService
		func(
			recurringRepo *infrastructure.RecurringRepository,
			transactionRepo *infrastructure.TransactionRepository,
			categoryService *category.Service,
			transactionService *transaction.Service,
			userChecker *shared.UserCheckerService,
		) *recurring.Service {
			return recurring.NewService(recurringRepo, transactionRepo, categoryService, transactionService, userChecker)
		},
		// ReportService
		func(
			reportRepo *infrastructure.ReportRepository,
			userService *user.Service,
		) *report.Service {
			return &report.Service{
				Repository:  reportRepo,
				UserService: userService,
			}
		},
		// CreditCardService
		func(
			creditCardRepo *infrastructure.CreditCardRepository,
			accountService *account.Service,
			userService *user.Service,
		) *creditcard.Service {
			return &creditcard.Service{
				Repository:     creditCardRepo,
				AccountService: accountService,
				UserService:    userService,
			}
		},
		// JwtService
		func(userService *user.Service, cfg *config.Config) (*middleware.JwtService, error) {
			return middleware.NewJwtService(cfg.JWT, userService)
		},
	)
}

// HandlerProvider fornece o handler HTTP
func HandlerProvider(
	userService *user.Service,
	jwtService *middleware.JwtService,
	authService *auth.Service,
	goalService *goal.Service,
	transactionService *transaction.Service,
	investmentService *investment.Service,
	accountService *account.Service,
	budgetService *budget.Service,
	dashboardService *dashboard.Service,
	recurringService *recurring.Service,
	reportService *report.Service,
	creditCardService *creditcard.Service,
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
		UserService:        *userService,
		JwtService:         jwtService,
		AuthService:        *authService,
		GoalService:        *goalService,
		TransactionService: *transactionService,
		InvestmentService:  *investmentService,
		AccountService:     *accountService,
		BudgetService:      *budgetService,
		DashboardService:   *dashboardService,
		RecurringService:   *recurringService,
		ReportService:      *reportService,
		CreditCardService:  *creditCardService,

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

// RouterProvider fornece o router do Gin
func RouterProvider(
	handler *routes.Handler,
	jwtService *middleware.JwtService,
	userService *user.Service,
	resourceCounter *infrastructure.ResourceCounter,
	cfg *config.Config,
) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	authRateLimiter := middleware.NewRateLimiter(100, time.Minute)

	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/api")
	public.Use(middleware.RateLimit(authRateLimiter))
	{
		public.POST("/auth/login", handler.Authenticate)
		public.POST("/auth/register", handler.Registration)
		public.POST("/auth/google", handler.GoogleAuth)
	}

	private := router.Group("/api")
	private.Use(middleware.AuthMiddleware(jwtService))
	private.Use(middleware.RequireOwnership())
	private.Use(middleware.RateLimitByUser())
	{
		private.GET("/dashboard", handler.GetDashboard)

		users := private.Group("/users")
		{
			users.GET("/plan", handler.GetUserPlan)
			users.PATCH("/me", handler.UpdateUserName)
			users.PATCH("/me/password", handler.UpdateUserPassword)
			users.DELETE("/me", handler.DeleteUser)
		}

		goals := private.Group("/goals")
		{
			goals.POST("", middleware.CheckResourceLimit("goals", resourceCounter, userService), handler.CreateGoal)
			goals.PATCH("/:id", handler.UpdateGoal)
			goals.GET("", handler.ListGoals)
			goals.GET("/:id", handler.GetGoal)
			goals.DELETE("/:id", handler.DeleteGoal)
			goals.POST("/:id/contribution", handler.ContributeToGoal)
			goals.POST("/:id/withdraw", handler.WithdrawFromGoal)
			goals.GET("/:id/contributions", handler.GetGoalContributions)
			goals.GET("/:id/progress", handler.GetGoalProgress)
			goals.DELETE("/contributions/:contribution_id", handler.DeleteContribution)
		}

		transactions := private.Group("/transactions")
		{
			transactions.POST("", middleware.CheckResourceLimit("transactions", resourceCounter, userService), handler.CreateTransaction)
			transactions.GET("", handler.GetTransactions)
			transactions.GET("/:id", handler.GetTransaction)
			transactions.PATCH("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
		}

		categories := private.Group("/categories")
		{
			categories.POST("", middleware.CheckResourceLimit("categories", resourceCounter, userService), handler.CreateCategory)
			categories.GET("", handler.ListCategories)
			categories.PATCH("/:id", handler.UpdateCategory)
			categories.DELETE("/:id", handler.DeleteCategory)
		}

		investments := private.Group("/investments")
		{
			investments.POST("", middleware.CheckResourceLimit("investments", resourceCounter, userService), handler.CreateInvestment)
			investments.GET("", handler.ListInvestments)
			investments.GET("/:id", handler.GetInvestment)
			investments.POST("/:id/contribution", handler.MakeContribution)
			investments.POST("/:id/withdraw", handler.MakeWithdraw)
			investments.GET("/:id/return", handler.GetInvestmentReturn)
			investments.DELETE("/:id", handler.DeleteInvestment)
			investments.PATCH("/:id", handler.UpdateInvestment)
		}

		accounts := private.Group("/accounts")
		{
			accounts.POST("", middleware.CheckResourceLimit("accounts", resourceCounter, userService), handler.CreateAccount)
			accounts.GET("", handler.ListAccounts)
			accounts.GET("/balance", handler.GetTotalBalance)
			accounts.GET("/:id", handler.GetAccount)
			accounts.PATCH("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
			accounts.POST("/transfer", handler.TransferBetweenAccounts)
		}

		budgets := private.Group("/budgets")
		{
			budgets.POST("", middleware.CheckResourceLimit("budgets", resourceCounter, userService), handler.CreateBudget)
			budgets.GET("", handler.ListBudgets)
			budgets.GET("/summary", handler.GetBudgetSummary)
			budgets.GET("/:id", handler.GetBudget)
			budgets.GET("/:id/status", handler.GetBudgetStatus)
			budgets.PATCH("/:id", handler.UpdateBudget)
			budgets.DELETE("/:id", handler.DeleteBudget)
		}

		recurring := private.Group("/recurring")
		{
			recurring.POST("", middleware.CheckResourceLimit("recurring", resourceCounter, userService), handler.CreateRecurring)
			recurring.GET("", handler.ListRecurrings)
			recurring.GET("/:id", handler.GetRecurring)
			recurring.PATCH("/:id", handler.UpdateRecurring)
			recurring.DELETE("/:id", handler.DeleteRecurring)
			recurring.POST("/:id/pause", handler.PauseRecurring)
			recurring.POST("/:id/resume", handler.ResumeRecurring)
			recurring.POST("/:id/process", handler.ProcessRecurring)
		}

		reports := private.Group("/reports")
		reports.Use(middleware.RequireFeature("reports"))
		{
			reports.GET("/current-month", handler.GetCurrentMonthReport)
			reports.GET("/period", handler.GetPeriodReport)
			reports.GET("/yearly", handler.GetYearlyReport)
			reports.GET("/category/:category_id", handler.GetCategoryReport)
		}

		creditCards := private.Group("/credit-cards")
		{
			creditCards.POST("", middleware.CheckResourceLimit("credit_cards", resourceCounter, userService), handler.CreateCreditCard)
			creditCards.GET("", handler.ListCreditCards)
			creditCards.GET("/:id", handler.GetCreditCard)
			creditCards.PATCH("/:id", handler.UpdateCreditCard)
			creditCards.DELETE("/:id", handler.DeleteCreditCard)
			creditCards.GET("/:id/invoices", handler.ListInvoices)
			creditCards.GET("/:id/invoices/current", handler.GetCurrentInvoice)
			creditCards.GET("/:id/invoices/:invoiceId", handler.GetInvoice)
			creditCards.POST("/:id/invoices/:invoiceId/pay", handler.PayInvoice)
			creditCards.POST("/:id/transactions", handler.CreateCreditCardTransaction)
			creditCards.GET("/:id/transactions", handler.ListCreditCardTransactions)
		}

		healthScoreService := healthscore.NewService()
		healthScoreHandler := routes.NewHealthScoreHandler(healthScoreService)
		private.GET("/health-score", healthScoreHandler.GetHealthScore)
	}

	return router
}
