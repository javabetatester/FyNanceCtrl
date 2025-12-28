package routes

import (
	"Fynance/internal/contracts"
	domaincontracts "Fynance/internal/domain/contracts"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateGoal(c *gin.Context) {
	var body contracts.GoalCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	req := domaincontracts.GoalCreateRequest{
		UserId:  userID,
		Name:    body.Name,
		Target:  body.Target,
		EndedAt: body.EndAt,
	}

	ctx := c.Request.Context()
	if err := h.GoalService.CreateGoal(ctx, &req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.MessageResponse{Message: "Meta criada com sucesso"})
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	var body contracts.GoalUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "é obrigatório"))
		return
	}

	goalID, err := pkg.ParseULID(id)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	req := domaincontracts.GoalUpdateRequest{
		Id:      goalID,
		UserId:  userID,
		Name:    body.Name,
		Target:  body.Target,
		EndedAt: body.EndAt,
	}

	ctx := c.Request.Context()
	if err := h.GoalService.UpdateGoal(ctx, &req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Meta atualizada com sucesso"})
}

func (h *Handler) ListGoals(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	goals, total, err := h.GoalService.GetGoalsByUserID(ctx, userID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(goals, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetGoal(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "é obrigatório"))
		return
	}

	goalID, err := pkg.ParseULID(id)
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
	goalEntity, err := h.GoalService.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.GoalResponse{Goal: goalEntity})
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "e obrigatorio"))
		return
	}

	goalID, err := pkg.ParseULID(id)
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
	if err := h.GoalService.DeleteGoal(ctx, goalID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Meta removida com sucesso"})
}

func (h *Handler) ContributeToGoal(c *gin.Context) {
	var body contracts.GoalContributionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "e obrigatorio"))
		return
	}

	goalID, err := pkg.ParseULID(id)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	accountID, err := pkg.ParseULID(body.AccountID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("account_id", "formato invalido"))
		return
	}

	ctx := c.Request.Context()
	if err := h.GoalService.MakeContribution(ctx, goalID, accountID, userID, body.Amount, body.Description); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Aporte realizado com sucesso"})
}

func (h *Handler) WithdrawFromGoal(c *gin.Context) {
	var body contracts.GoalWithdrawRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "e obrigatorio"))
		return
	}

	goalID, err := pkg.ParseULID(id)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato invalido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	accountID, err := pkg.ParseULID(body.AccountID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("account_id", "formato invalido"))
		return
	}

	ctx := c.Request.Context()
	if err := h.GoalService.WithdrawFromGoal(ctx, goalID, accountID, userID, body.Amount, body.Description); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Resgate realizado com sucesso"})
}

func (h *Handler) GetGoalContributions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "e obrigatorio"))
		return
	}

	goalID, err := pkg.ParseULID(id)
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
	contributions, err := h.GoalService.GetContributions(ctx, goalID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.GoalContributionListResponse{Contributions: contributions, Total: len(contributions)})
}

func (h *Handler) GetGoalProgress(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, appErrors.NewValidationError("id", "e obrigatorio"))
		return
	}

	goalID, err := pkg.ParseULID(id)
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
	progress, err := h.GoalService.GetGoalProgress(ctx, goalID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.GoalProgressResponse{Progress: progress})
}
