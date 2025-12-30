package routes

import (
	"net/http"
	"strconv"
	"time"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/budget"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBudget(c *gin.Context) {
	var body contracts.BudgetCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	categoryID, err := pkg.ParseULID(body.CategoryId)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
		return
	}

	req := &budget.CreateBudgetRequest{
		UserId:      userID,
		CategoryId:  categoryID,
		Amount:      body.Amount,
		Month:       body.Month,
		Year:        body.Year,
		AlertAt:     body.AlertAt,
		IsRecurring: body.IsRecurring,
	}

	ctx := c.Request.Context()
	b, err := h.BudgetService.CreateBudget(ctx, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.BudgetCreateResponse{
		Message: "Orcamento criado com sucesso",
		Budget:  b,
	})
}

func (h *Handler) ListBudgets(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var month, year int
	if m := c.Query("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 12 {
			month = parsed
		}
	}

	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed > 0 {
			year = parsed
		}
	}

	var filters *budget.BudgetFilters
	searchStr := c.Query("search")
	if searchStr != "" {
		filters = &budget.BudgetFilters{Search: &searchStr}
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	budgets, total, err := h.BudgetService.ListBudgets(ctx, userID, month, year, filters, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	budgetResponses := make([]*contracts.BudgetResponse, 0, len(budgets))
	for _, b := range budgets {
		remaining := b.Amount - b.Spent
		percentage := 0.0
		if b.Amount > 0 {
			percentage = (b.Spent / b.Amount) * 100
		}

		status := "OK"
		if percentage >= 100 {
			status = "EXCEEDED"
		} else if percentage >= b.AlertAt {
			status = "WARNING"
		}

		budgetResponses = append(budgetResponses, &contracts.BudgetResponse{
			Budget:       b,
			Percentage:   percentage,
			Remaining:    remaining,
			SpentAmount:  b.Spent,
			BudgetAmount: b.Amount,
			Status:       status,
		})
	}

	response := pkg.NewPaginatedResponse(budgetResponses, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetBudget(c *gin.Context) {
	budgetID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	b, err := h.BudgetService.GetBudgetByID(ctx, budgetID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.BudgetSingleResponse{Budget: b})
}

func (h *Handler) UpdateBudget(c *gin.Context) {
	budgetID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.BudgetUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	req := &budget.UpdateBudgetRequest{
		Amount:      body.Amount,
		AlertAt:     body.AlertAt,
		IsRecurring: body.IsRecurring,
	}

	ctx := c.Request.Context()
	if err := h.BudgetService.UpdateBudget(ctx, budgetID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Orcamento atualizado com sucesso"})
}

func (h *Handler) DeleteBudget(c *gin.Context) {
	budgetID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.BudgetService.DeleteBudget(ctx, budgetID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Orcamento removido com sucesso"})
}

func (h *Handler) GetBudgetSummary(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if m := c.Query("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	if y := c.Query("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}

	ctx := c.Request.Context()
	summary, err := h.BudgetService.GetBudgetSummary(ctx, userID, month, year)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.BudgetSummaryResponse{Summary: summary})
}

func (h *Handler) GetBudgetStatus(c *gin.Context) {
	budgetID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	status, err := h.BudgetService.GetBudgetStatus(ctx, budgetID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}
