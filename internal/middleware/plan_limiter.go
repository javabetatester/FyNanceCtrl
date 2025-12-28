package middleware

import (
	"net/http"

	"Fynance/internal/domain/plan"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"

	"github.com/gin-gonic/gin"
)

type ResourceCounter interface {
	CountTransactions(userID string) (int64, error)
	CountCategories(userID string) (int64, error)
	CountAccounts(userID string) (int64, error)
	CountGoals(userID string) (int64, error)
	CountInvestments(userID string) (int64, error)
	CountBudgets(userID string) (int64, error)
	CountRecurring(userID string) (int64, error)
	CountCreditCards(userID string) (int64, error)
}

func respondLimit(c *gin.Context, err *appErrors.AppError) {
	payload := gin.H{
		"error":   err.Code,
		"message": err.Message,
	}
	if len(err.Details) > 0 {
		payload["details"] = err.Details
	}
	c.JSON(err.StatusCode, payload)
	c.Abort()
}

func RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		planValue, exists := c.Get("plan")
		if !exists {
			err := appErrors.WrapError(nil, appErrors.ErrForbidden.Code, "Plano nao encontrado", http.StatusForbidden)
			respondLimit(c, err)
			return
		}

		userPlan, ok := planValue.(user.Plan)
		if !ok {
			err := appErrors.WrapError(nil, appErrors.ErrForbidden.Code, "Plano invalido", http.StatusForbidden)
			respondLimit(c, err)
			return
		}

		limits := plan.GetLimits(userPlan)

		allowed := false
		switch feature {
		case "dashboard":
			allowed = limits.HasDashboard
		case "reports":
			allowed = limits.HasReports
		case "export":
			allowed = limits.HasExport
		case "multiple_users":
			allowed = limits.HasMultipleUsers
		default:
			allowed = true
		}

		if !allowed {
			err := appErrors.WrapError(nil, "PLAN_LIMIT_REACHED",
				"Funcionalidade nao disponivel no seu plano. Faca upgrade para acessar.",
				http.StatusForbidden)
			err.Details = map[string]interface{}{
				"feature":      feature,
				"current_plan": string(userPlan),
			}
			respondLimit(c, err)
			return
		}

		c.Next()
	}
}

func CheckResourceLimit(resourceType string, counter ResourceCounter) gin.HandlerFunc {
	return func(c *gin.Context) {
		planValue, exists := c.Get("plan")
		if !exists {
			c.Next()
			return
		}

		userPlan, ok := planValue.(user.Plan)
		if !ok {
			c.Next()
			return
		}

		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userID, ok := userIDValue.(string)
		if !ok {
			c.Next()
			return
		}

		limits := plan.GetLimits(userPlan)

		var limit int
		var count int64
		var err error

		switch resourceType {
		case "transactions":
			limit = limits.MaxTransactions
			count, err = counter.CountTransactions(userID)
		case "categories":
			limit = limits.MaxCategories
			count, err = counter.CountCategories(userID)
		case "accounts":
			limit = limits.MaxAccounts
			count, err = counter.CountAccounts(userID)
		case "goals":
			limit = limits.MaxGoals
			count, err = counter.CountGoals(userID)
		case "investments":
			limit = limits.MaxInvestments
			count, err = counter.CountInvestments(userID)
		case "budgets":
			limit = limits.MaxBudgets
			count, err = counter.CountBudgets(userID)
		case "recurring":
			limit = limits.MaxRecurring
			count, err = counter.CountRecurring(userID)
		case "credit_cards":
			limit = limits.MaxCreditCards
			count, err = counter.CountCreditCards(userID)
		default:
			c.Next()
			return
		}

		if err != nil {
			c.Next()
			return
		}

		if !plan.IsUnlimited(limit) && int(count) >= limit {
			appErr := appErrors.WrapError(nil, "PLAN_LIMIT_REACHED",
				"Limite do plano atingido. Faca upgrade para criar mais recursos.",
				http.StatusForbidden)
			appErr.Details = map[string]interface{}{
				"resource":     resourceType,
				"current":      count,
				"limit":        limit,
				"current_plan": string(userPlan),
			}
			respondLimit(c, appErr)
			return
		}

		c.Next()
	}
}
