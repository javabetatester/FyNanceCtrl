package routes

import (
	"net/http"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/auth"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Authenticate(c *gin.Context) {
	var body contracts.AuthLoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	login := auth.Login{
		Email:    body.Email,
		Password: body.Password,
	}

	ctx := c.Request.Context()
	userEntity, err := h.AuthService.Login(ctx, login)
	if err != nil {
		h.respondError(c, err)
		return
	}

	token, err := h.JwtService.GenerateToken(ctx, userEntity.Id)
	if err != nil {
		h.respondError(c, appErrors.ErrInternalServer.WithError(err))
		return
	}

	c.JSON(http.StatusOK, contracts.AuthLoginResponse{
		Message: "Login realizado com sucesso",
		User:    userEntity.Name,
		Token:   token,
	})
}

func (h *Handler) Registration(c *gin.Context) {
	var body contracts.AuthRegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
		return
	}

	userEntity := user.User{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}

	ctx := c.Request.Context()
	if err := h.AuthService.Register(ctx, &userEntity); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.MessageResponse{Message: "Usuário registrado com sucesso"})
}
