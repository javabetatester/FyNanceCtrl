package main

import (
	"log"
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
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Aviso: não foi possível carregar .env do diretório atual: %v", err)
	}
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Aviso: não foi possível carregar ../../.env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger.Init(cfg)

	db, err := infrastructure.NewDb(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Falha ao inicializar banco de dados")
	}

	userRepo := &infrastructure.UserRepository{DB: db}
	goalRepo := &infrastructure.GoalRepository{DB: db}
	transactionRepo := &infrastructure.TransactionRepository{DB: db}
	categoryRepo := &infrastructure.TransactionCategoryRepository{DB: db}
	investmentRepo := &infrastructure.InvestmentRepository{DB: db}
	accountRepo := &infrastructure.AccountRepository{DB: db}
	budgetRepo := &infrastructure.BudgetRepository{DB: db}
	dashboardRepo := &infrastructure.DashboardRepository{DB: db}
	recurringRepo := &infrastructure.RecurringRepository{DB: db}
	reportRepo := &infrastructure.ReportRepository{DB: db}
	creditCardRepo := &infrastructure.CreditCardRepository{DB: db}
	resourceCounter := &infrastructure.ResourceCounter{DB: db}

	userService := user.NewService(userRepo)

	userAdapter := user.NewUserServiceAdapter(userService)
	userChecker := shared.NewUserCheckerService(userAdapter)

	categoryService := category.NewService(categoryRepo, userChecker)

	accountService := account.NewService(accountRepo, userChecker)

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

	authService := auth.NewService(
		userRepo,
		userService,
		googleClientID,
	)

	budgetService := budget.NewService(budgetRepo, categoryService, userChecker)

	investmentService := investment.NewService(investmentRepo, transactionRepo, accountService, userChecker)

	goalService := goal.NewService(goalRepo, accountService, nil, userChecker)

	var (
		_ category.CategoryServiceInterface   = categoryService
		_ account.AccountServiceInterface     = accountService
		_ shared.BudgetUpdater                = budgetService
		_ shared.GoalContributionDeleter      = goalService
		_ shared.InvestmentTransactionDeleter = investmentService
	)

	transactionService := transaction.NewService(
		transactionRepo,
		categoryService,
		accountService,
		budgetService,
		goalService,
		investmentService,
		userChecker,
	)

	var _ transaction.TransactionHandler = transactionService
	goalService.TransactionService = transactionService

	dashboardService := dashboard.Service{
		Repository: dashboardRepo,
	}

	var _ transaction.TransactionHandler = transactionService
	recurringService := recurring.NewService(recurringRepo, transactionRepo, categoryService, transactionService, userChecker)

	reportService := report.Service{
		Repository:  reportRepo,
		UserService: userService,
	}

	creditCardService := creditcard.Service{
		Repository:     creditCardRepo,
		AccountService: accountService,
		UserService:    userService,
	}

	jwtService, err := middleware.NewJwtService(cfg.JWT, userService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Falha ao inicializar servico JWT")
	}

	handler := routes.Handler{
		UserService:        *userService,
		JwtService:         jwtService,
		AuthService:        *authService,
		GoalService:        *goalService,
		TransactionService: *transactionService,
		InvestmentService:  *investmentService,
		AccountService:     *accountService,
		BudgetService:      *budgetService,
		DashboardService:   dashboardService,
		RecurringService:   *recurringService,
		ReportService:      reportService,
		CreditCardService:  creditCardService,

		AccountRepository:     accountRepo,
		TransactionRepository: transactionRepo,
		GoalRepository:        goalRepo,
		BudgetRepository:      budgetRepo,
		InvestmentRepository:  investmentRepo,
		RecurringRepository:   recurringRepo,
		CreditCardRepository:  creditCardRepo,
		CategoryRepository:    categoryRepo,
	}

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

	serverAddr := ":" + cfg.Server.Port
	logger.Info().
		Str("address", serverAddr).
		Str("environment", cfg.App.Environment).
		Msg("Servidor iniciando")

	if err := router.Run(serverAddr); err != nil {
		logger.Fatal().Err(err).Msg("Falha ao iniciar servidor")
	}
}
