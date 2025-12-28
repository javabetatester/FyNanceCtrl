package main

import (
	"log"
	"time"

	"Fynance/config"
	"Fynance/internal/domain/account"
	"Fynance/internal/domain/auth"
	"Fynance/internal/domain/budget"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/dashboard"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/report"
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
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

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

	userService := user.Service{
		Repository: userRepo,
	}

	accountService := account.Service{
		Repository:  accountRepo,
		UserService: &userService,
	}

	transactionService := transaction.Service{
		Repository:         transactionRepo,
		CategoryRepository: categoryRepo,
		UserService:        &userService,
		AccountService:     &accountService,
	}

	authService := auth.Service{
		Repository:      userRepo,
		UserService:     &userService,
		CategoryService: &transactionService,
	}

	goalService := goal.Service{
		Repository:     goalRepo,
		UserService:    userService,
		AccountService: &accountService,
	}

	investmentService := investment.Service{
		Repository:      investmentRepo,
		TransactionRepo: transactionRepo,
		UserService:     &userService,
	}

	budgetService := budget.Service{
		Repository:         budgetRepo,
		CategoryRepository: categoryRepo,
		UserService:        &userService,
	}

	dashboardService := dashboard.Service{
		Repository: dashboardRepo,
	}

	recurringService := recurring.Service{
		Repository:         recurringRepo,
		TransactionRepo:    transactionRepo,
		CategoryRepository: categoryRepo,
		UserService:        &userService,
	}

	reportService := report.Service{
		Repository:  reportRepo,
		UserService: &userService,
	}

	creditCardService := creditcard.Service{
		Repository:     creditCardRepo,
		AccountService: &accountService,
		UserService:    &userService,
	}

	jwtService, err := middleware.NewJwtService(cfg.JWT, &userService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Falha ao inicializar servico JWT")
	}

	handler := routes.Handler{
		UserService:        userService,
		JwtService:         jwtService,
		AuthService:        authService,
		GoalService:        goalService,
		TransactionService: transactionService,
		InvestmentService:  investmentService,
		AccountService:     accountService,
		BudgetService:      budgetService,
		DashboardService:   dashboardService,
		RecurringService:   recurringService,
		ReportService:      reportService,
		CreditCardService:  creditCardService,
	}

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/api")
	public.Use(middleware.RateLimit(authRateLimiter))
	{
		public.POST("/auth/login", handler.Authenticate)
		public.POST("/auth/register", handler.Registration)
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
		}

		goals := private.Group("/goals")
		{
			goals.POST("", middleware.CheckResourceLimit("goals", resourceCounter), handler.CreateGoal)
			goals.PATCH("/:id", handler.UpdateGoal)
			goals.GET("", handler.ListGoals)
			goals.GET("/:id", handler.GetGoal)
			goals.DELETE("/:id", handler.DeleteGoal)
			goals.POST("/:id/contribution", handler.ContributeToGoal)
			goals.POST("/:id/withdraw", handler.WithdrawFromGoal)
			goals.GET("/:id/contributions", handler.GetGoalContributions)
			goals.GET("/:id/progress", handler.GetGoalProgress)
		}

		transactions := private.Group("/transactions")
		{
			transactions.POST("", middleware.CheckResourceLimit("transactions", resourceCounter), handler.CreateTransaction)
			transactions.GET("", handler.GetTransactions)
			transactions.GET("/:id", handler.GetTransaction)
			transactions.PATCH("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
		}

		categories := private.Group("/categories")
		{
			categories.POST("", middleware.CheckResourceLimit("categories", resourceCounter), handler.CreateCategory)
			categories.GET("", handler.ListCategories)
			categories.PATCH("/:id", handler.UpdateCategory)
			categories.DELETE("/:id", handler.DeleteCategory)
		}

		investments := private.Group("/investments")
		{
			investments.POST("", middleware.CheckResourceLimit("investments", resourceCounter), handler.CreateInvestment)
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
			accounts.POST("", middleware.CheckResourceLimit("accounts", resourceCounter), handler.CreateAccount)
			accounts.GET("", handler.ListAccounts)
			accounts.GET("/balance", handler.GetTotalBalance)
			accounts.GET("/:id", handler.GetAccount)
			accounts.PATCH("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
			accounts.POST("/transfer", handler.TransferBetweenAccounts)
		}

		budgets := private.Group("/budgets")
		{
			budgets.POST("", middleware.CheckResourceLimit("budgets", resourceCounter), handler.CreateBudget)
			budgets.GET("", handler.ListBudgets)
			budgets.GET("/summary", handler.GetBudgetSummary)
			budgets.GET("/:id", handler.GetBudget)
			budgets.GET("/:id/status", handler.GetBudgetStatus)
			budgets.PATCH("/:id", handler.UpdateBudget)
			budgets.DELETE("/:id", handler.DeleteBudget)
		}

		recurring := private.Group("/recurring")
		{
			recurring.POST("", middleware.CheckResourceLimit("recurring", resourceCounter), handler.CreateRecurring)
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
			creditCards.POST("", middleware.CheckResourceLimit("credit_cards", resourceCounter), handler.CreateCreditCard)
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
