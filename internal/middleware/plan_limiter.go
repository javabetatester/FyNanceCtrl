package middleware

import (
	"context"
	"net/http"

	"Fynance/internal/domain/plan"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
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

type UserService interface {
	GetPlan(ctx context.Context, id ulid.ULID) (user.Plan, error)
}

func CheckResourceLimit(resourceType string, counter ResourceCounter, userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userIDValue.(string)
		if !ok || userIDStr == "" {
			c.Next()
			return
		}

		userID, err := pkg.ParseULID(userIDStr)
		if err != nil {
			c.Next()
			return
		}

		userPlan, err := userService.GetPlan(c.Request.Context(), userID)
		if err != nil {
			c.Next()
			return
		}

		limits := plan.GetLimits(userPlan)

		var limit int
		var count int64

		switch resourceType {
		case "transactions":
			limit = limits.MaxTransactions
			count, err = counter.CountTransactions(userIDStr)
		case "categories":
			limit = limits.MaxCategories
			count, err = counter.CountCategories(userIDStr)
		case "accounts":
			limit = limits.MaxAccounts
			count, err = counter.CountAccounts(userIDStr)
		case "goals":
			limit = limits.MaxGoals
			count, err = counter.CountGoals(userIDStr)
		case "investments":
			limit = limits.MaxInvestments
			count, err = counter.CountInvestments(userIDStr)
		case "budgets":
			limit = limits.MaxBudgets
			count, err = counter.CountBudgets(userIDStr)
		case "recurring":
			limit = limits.MaxRecurring
			count, err = counter.CountRecurring(userIDStr)
		case "credit_cards":
			limit = limits.MaxCreditCards
			count, err = counter.CountCreditCards(userIDStr)
		default:
			c.Next()
			return
		}

		if err != nil {
			c.Next()
			return
		}

		if !plan.IsUnlimited(limit) && int(count) >= limit {
			resourceNames := map[string]string{
				"transactions": "transações",
				"categories":   "categorias",
				"accounts":     "contas",
				"goals":        "metas",
				"investments":  "investimentos",
				"budgets":      "orçamentos",
				"recurring":    "transações recorrentes",
				"credit_cards": "cartões de crédito",
			}
			resourceName := resourceNames[resourceType]
			if resourceName == "" {
				resourceName = resourceType
			}

			message := "Você atingiu o limite de " + resourceName + " do seu plano atual"
			if userPlan == user.PlanFree {
				message += " (FREE). Faça upgrade para um plano superior e crie mais " + resourceName + "."
			} else {
				message += ". Faça upgrade para criar mais " + resourceName + "."
			}

			appErr := appErrors.WrapError(nil, "PLAN_LIMIT_REACHED", message, http.StatusForbidden)
			appErr.Details = map[string]interface{}{
				"resource":     resourceType,
				"resourceName": resourceName,
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
