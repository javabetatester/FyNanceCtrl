package routes

import (
	"context"
	"net/http"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetUserPlan(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	plan, err := h.UserService.GetPlan(ctx, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	userEntity, err := h.UserService.GetByID(ctx, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.UserPlanResponse{
		Plan:      string(plan),
		PlanSince: userEntity.PlanSince,
	})
}

func (h *Handler) UpdateUserPlanInternal(userID string, newPlan user.Plan) error {
	ctx := context.Background()
	ulidUserID, err := pkg.ParseULID(userID)
	if err != nil {
		return appErrors.NewValidationError("user_id", "formato inválido")
	}
	return h.UserService.UpdatePlan(ctx, ulidUserID, newPlan)
}

func (h *Handler) UpdateUserName(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.UserUpdateNameRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	ctx := c.Request.Context()
	if err := h.UserService.UpdateName(ctx, userID, body.Name); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{
		Message: "Nome atualizado com sucesso",
	})
}

func (h *Handler) UpdateUserPassword(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.UserUpdatePasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	ctx := c.Request.Context()
	if err := h.UserService.UpdatePassword(ctx, userID, body.CurrentPassword, body.NewPassword); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{
		Message: "Senha atualizada com sucesso",
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.UserService.Delete(ctx, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.UserDeletionResponse{
		Message: "Conta excluída com sucesso",
	})
}
