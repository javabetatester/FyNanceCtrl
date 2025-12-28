package routes

import (
	"Fynance/internal/contracts"
	"Fynance/internal/domain/transaction"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

func (h *Handler) CreateTransaction(c *gin.Context) {
	var body contracts.TransactionCreateRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	accountID, err := pkg.ParseULID(body.AccountID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("account_id", "formato inválido"))
		return
	}

	categoryID, err := pkg.ParseULID(body.CategoryID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
		return
	}

	transactionEntity := transaction.Transaction{
		Type:        transaction.Types(body.Type),
		UserId:      userID,
		AccountId:   accountID,
		CategoryId:  categoryID,
		Amount:      body.Amount,
		Description: body.Description,
		Date:        pkg.SetTimestamps(),
	}

	ctx := c.Request.Context()
	if err := h.TransactionService.CreateTransaction(ctx, &transactionEntity); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.TransactionCreateResponse{
		Message:     "Transação criada com sucesso",
		Transaction: transactionEntity,
	})
}

func (h *Handler) GetTransactions(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	accountIDStr := c.Query("account_id")
	var accountID *ulid.ULID
	if accountIDStr != "" {
		parsed, err := pkg.ParseULID(accountIDStr)
		if err != nil {
			h.respondError(c, appErrors.NewValidationError("account_id", "formato inválido"))
			return
		}
		accountID = &parsed
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	transactions, total, err := h.TransactionService.GetAllTransactions(ctx, userID, accountID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(transactions, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	transactionID, err := pkg.ParseULID(c.Param("id"))
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
	transactionEntity, err := h.TransactionService.GetTransactionByID(ctx, transactionID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.TransactionSingleResponse{Transaction: transactionEntity})
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	transactionID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.TransactionUpdateRequest
	if err = c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	accountID, err := pkg.ParseULID(body.AccountID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("account_id", "formato inválido"))
		return
	}

	categoryID, err := pkg.ParseULID(body.CategoryID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
		return
	}

	transactionEntity := transaction.Transaction{
		Id:          transactionID,
		UserId:      userID,
		AccountId:   accountID,
		CategoryId:  categoryID,
		Amount:      body.Amount,
		Description: body.Description,
		Type:        transaction.Types(body.Type),
		UpdatedAt:   pkg.SetTimestamps(),
	}

	if body.Date != nil {
		transactionEntity.Date = *body.Date
	}

	ctx := c.Request.Context()
	if err := h.TransactionService.UpdateTransaction(ctx, &transactionEntity); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transação atualizada com sucesso"})
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	transactionID, err := pkg.ParseULID(c.Param("id"))
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
	if err := h.TransactionService.DeleteTransaction(ctx, transactionID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Transação removida com sucesso"})
}
