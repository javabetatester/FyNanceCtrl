package routes

import (
	"net/http"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/account"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateAccount(c *gin.Context) {
	var body contracts.AccountCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	includeInTotal := true
	if body.IncludeInTotal != nil {
		includeInTotal = *body.IncludeInTotal
	}

	req := &account.CreateAccountRequest{
		UserId:         userID,
		Name:           body.Name,
		Type:           account.AccountType(body.Type),
		InitialBalance: body.InitialBalance,
		Color:          body.Color,
		Icon:           body.Icon,
		IncludeInTotal: includeInTotal,
	}

	ctx := c.Request.Context()
	acc, err := h.AccountService.CreateAccount(ctx, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.AccountCreateResponse{
		Message: "Conta criada com sucesso",
		Account: acc,
	})
}

func (h *Handler) ListAccounts(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()

	var accountType *string
	typeStr := c.Query("type")
	if typeStr != "" && typeStr != "ALL" {
		accountType = &typeStr
	}

	var search *string
	searchStr := c.Query("search")
	if searchStr != "" {
		search = &searchStr
	}

	pagination := h.parsePagination(c)

	accounts, total, err := h.AccountRepository.GetByUserIDWithFilters(ctx, userID, accountType, search, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(accounts, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetAccount(c *gin.Context) {
	accountID, err := pkg.ParseULID(c.Param("id"))
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
	acc, err := h.AccountService.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.AccountSingleResponse{Account: acc})
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	accountID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.AccountUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	req := &account.UpdateAccountRequest{
		Name:           body.Name,
		Color:          body.Color,
		Icon:           body.Icon,
		IncludeInTotal: body.IncludeInTotal,
		IsActive:       body.IsActive,
	}

	if body.Type != nil {
		t := account.AccountType(*body.Type)
		req.Type = &t
	}

	ctx := c.Request.Context()
	if err := h.AccountService.UpdateAccount(ctx, accountID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Conta atualizada com sucesso"})
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	accountID, err := pkg.ParseULID(c.Param("id"))
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
	if err := h.AccountService.DeleteAccount(ctx, accountID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Conta removida com sucesso"})
}

func (h *Handler) TransferBetweenAccounts(c *gin.Context) {
	var body contracts.AccountTransferRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	fromAccountID, err := pkg.ParseULID(body.FromAccountId)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("from_account_id", "formato inválido"))
		return
	}

	toAccountID, err := pkg.ParseULID(body.ToAccountId)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("to_account_id", "formato inválido"))
		return
	}

	ctx := c.Request.Context()
	if err := h.AccountService.Transfer(ctx, fromAccountID, toAccountID, userID, body.Amount); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transferencia realizada com sucesso"})
}

func (h *Handler) GetTotalBalance(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	total, err := h.AccountService.GetTotalBalance(ctx, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.AccountBalanceResponse{TotalBalance: total})
}
