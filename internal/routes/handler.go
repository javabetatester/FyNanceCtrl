package routes

import (
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
	appErrors "Fynance/internal/errors"
	"Fynance/internal/logger"
	"Fynance/internal/middleware"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type Handler struct {
	UserService        user.Service
	AuthService        auth.Service
	JwtService         *middleware.JwtService
	TransactionService transaction.Service
	GoalService        goal.Service
	InvestmentService  investment.Service
	AccountService     account.Service
	BudgetService      budget.Service
	DashboardService   dashboard.Service
	RecurringService   recurring.Service
	ReportService      report.Service
	CreditCardService  creditcard.Service
}

func (h *Handler) GetUserIDFromContext(c *gin.Context) (ulid.ULID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return ulid.ULID{}, appErrors.ErrUnauthorized
	}

	userID, err := pkg.ParseULID(userIDStr.(string))
	if err != nil {
		return ulid.ULID{}, appErrors.ErrUnauthorized.WithError(err)
	}

	return userID, nil
}

func (h *Handler) parsePagination(c *gin.Context) *pkg.PaginationParams {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	var pageNum, limitNum int
	if p, err := pkg.ParseInt(page); err == nil && p > 0 {
		pageNum = p
	} else {
		pageNum = 1
	}

	if l, err := pkg.ParseInt(limit); err == nil && l > 0 {
		limitNum = l
	} else {
		limitNum = 10
	}

	return &pkg.PaginationParams{
		Page:  pageNum,
		Limit: limitNum,
	}
}

func (h *Handler) respondError(c *gin.Context, err error) {
	appErr := appErrors.FromError(err)
	event := logger.Error().Str("code", appErr.Code).Str("path", c.FullPath())
	if appErr.Err != nil {
		event = event.Err(appErr.Err)
	}
	event.Msg("request_error")
	payload := gin.H{
		"error":   appErr.Code,
		"message": appErr.Message,
	}
	if len(appErr.Details) > 0 {
		payload["details"] = appErr.Details
	}
	c.JSON(appErr.StatusCode, payload)
}
