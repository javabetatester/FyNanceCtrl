package fx

import (
	"context"

	"Fynance/config"
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
)

// ServerModule fornece a configuração do servidor HTTP
var ServerModule = fx.Module("server",
	fx.Provide(
		newRouter,
	),
	fx.Invoke(
		setupRoutes,
	),
)

func newRouter() *gin.Engine {
	return gin.Default()
}

func setupRoutes(
	lc fx.Lifecycle,
	cfg *config.Config,
	router *gin.Engine,
	handler *routes.Handler,
	jwtSvc *middleware.JwtService,
	authRateLimiter *middleware.RateLimiter,
	userSvc *user.Service,
	resourceCounter *infrastructure.ResourceCounter,
	healthScoreHandler *routes.HealthScoreHandler,
) {
	router.Use(middleware.CORSMiddleware())

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
	private.Use(middleware.AuthMiddleware(jwtSvc))
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
			goals.POST("", middleware.CheckResourceLimit("goals", resourceCounter, userSvc), handler.CreateGoal)
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
			transactions.POST("", middleware.CheckResourceLimit("transactions", resourceCounter, userSvc), handler.CreateTransaction)
			transactions.GET("", handler.GetTransactions)
			transactions.GET("/:id", handler.GetTransaction)
			transactions.PATCH("/:id", handler.UpdateTransaction)
			transactions.DELETE("/:id", handler.DeleteTransaction)
		}

		categories := private.Group("/categories")
		{
			categories.POST("", middleware.CheckResourceLimit("categories", resourceCounter, userSvc), handler.CreateCategory)
			categories.GET("", handler.ListCategories)
			categories.PATCH("/:id", handler.UpdateCategory)
			categories.DELETE("/:id", handler.DeleteCategory)
		}

		investments := private.Group("/investments")
		{
			investments.POST("", middleware.CheckResourceLimit("investments", resourceCounter, userSvc), handler.CreateInvestment)
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
			accounts.POST("", middleware.CheckResourceLimit("accounts", resourceCounter, userSvc), handler.CreateAccount)
			accounts.GET("", handler.ListAccounts)
			accounts.GET("/balance", handler.GetTotalBalance)
			accounts.GET("/:id", handler.GetAccount)
			accounts.PATCH("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
			accounts.POST("/transfer", handler.TransferBetweenAccounts)
		}

		budgets := private.Group("/budgets")
		{
			budgets.POST("", middleware.CheckResourceLimit("budgets", resourceCounter, userSvc), handler.CreateBudget)
			budgets.GET("", handler.ListBudgets)
			budgets.GET("/summary", handler.GetBudgetSummary)
			budgets.GET("/:id", handler.GetBudget)
			budgets.GET("/:id/status", handler.GetBudgetStatus)
			budgets.PATCH("/:id", handler.UpdateBudget)
			budgets.DELETE("/:id", handler.DeleteBudget)
		}

		recurring := private.Group("/recurring")
		{
			recurring.POST("", middleware.CheckResourceLimit("recurring", resourceCounter, userSvc), handler.CreateRecurring)
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
			creditCards.POST("", middleware.CheckResourceLimit("credit_cards", resourceCounter, userSvc), handler.CreateCreditCard)
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

		private.GET("/health-score", healthScoreHandler.GetHealthScore)
	}

	serverAddr := ":" + cfg.Server.Port
	logger.Info().
		Str("address", serverAddr).
		Str("environment", cfg.App.Environment).
		Msg("Servidor iniciando")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := router.Run(serverAddr); err != nil {
					logger.Fatal().Err(err).Msg("Falha ao iniciar servidor")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info().Msg("Servidor parando...")
			return nil
		},
	})
}
