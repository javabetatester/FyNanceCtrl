package routes

import (
	"net/http"
	"time"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/recurring"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRecurring(c *gin.Context) {
	var body contracts.RecurringCreateRequest
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
		h.respondError(c, appErrors.NewValidationError("category_id", "formato invalido"))
		return
	}

	req := &recurring.CreateRecurringRequest{
		UserId:      userID,
		Type:        body.Type,
		CategoryId:  categoryID,
		Amount:      body.Amount,
		Description: body.Description,
		Frequency:   recurring.FrequencyType(body.Frequency),
		DayOfMonth:  body.DayOfMonth,
		DayOfWeek:   body.DayOfWeek,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
	}

	if body.AccountId != "" {
		accountID, err := pkg.ParseULID(body.AccountId)
		if err != nil {
			h.respondError(c, appErrors.NewValidationError("account_id", "formato invalido"))
			return
		}
		req.AccountId = &accountID
	}

	if req.DayOfMonth == 0 {
		req.DayOfMonth = time.Now().Day()
	}

	ctx := c.Request.Context()
	rec, err := h.RecurringService.CreateRecurring(ctx, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.RecurringCreateResponse{
		Message:   "Transacao recorrente criada com sucesso",
		Recurring: rec,
	})
}

func (h *Handler) ListRecurrings(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	recurring, total, err := h.RecurringService.ListRecurring(ctx, userID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(recurring, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	rec, err := h.RecurringService.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.RecurringSingleResponse{Recurring: rec})
}

func (h *Handler) UpdateRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.RecurringUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	req := &recurring.UpdateRecurringRequest{
		Amount:      body.Amount,
		Description: body.Description,
		IsActive:    body.IsActive,
		EndDate:     body.EndDate,
	}

	ctx := c.Request.Context()
	if err := h.RecurringService.UpdateRecurring(ctx, recurringID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transacao recorrente atualizada com sucesso"})
}

func (h *Handler) DeleteRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.RecurringService.DeleteRecurring(ctx, recurringID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transacao recorrente removida com sucesso"})
}

func (h *Handler) PauseRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	isActive := false
	req := &recurring.UpdateRecurringRequest{
		IsActive: &isActive,
	}

	ctx := c.Request.Context()
	if err := h.RecurringService.UpdateRecurring(ctx, recurringID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transacao recorrente pausada com sucesso"})
}

func (h *Handler) ResumeRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	isActive := true
	req := &recurring.UpdateRecurringRequest{
		IsActive: &isActive,
	}

	ctx := c.Request.Context()
	if err := h.RecurringService.UpdateRecurring(ctx, recurringID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transacao recorrente reativada com sucesso"})
}

func (h *Handler) ProcessRecurring(c *gin.Context) {
	recurringID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.RecurringProcessRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	ctx := c.Request.Context()
	tx, err := h.RecurringService.ProcessRecurringManually(ctx, recurringID, userID, body.ProcessDate)
	if err != nil {
		h.respondError(c, err)
		return
	}

	rec, err := h.RecurringService.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.RecurringProcessResponse{
		Message:     "Transacao recorrente processada com sucesso",
		Transaction: tx,
		Recurring:   rec,
	})
}
