package routes

import (
	"Fynance/internal/contracts"
	"Fynance/internal/domain/account"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/transaction"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

func (h *Handler) CreateTransaction(c *gin.Context) {
	var body contracts.TransactionCreateRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ParseValidationErrors(err))
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

	ctx := c.Request.Context()
	accountEntity, err := h.AccountService.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	if accountEntity.Type == account.TypeCreditCard && accountEntity.CreditCardId != nil {
		categoryID, err := pkg.ParseULID(body.CategoryID)
		if err != nil {
			h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
			return
		}

		transactionAmount := body.Amount
		if transactionAmount < 0 {
			transactionAmount = -transactionAmount
		}
		if transactionAmount <= 0 {
			h.respondError(c, appErrors.NewValidationError("valor", "deve ser maior que zero"))
			return
		}

		if body.Date == nil {
			h.respondError(c, appErrors.NewValidationError("date", "é obrigatória"))
			return
		}

		transactionDate := *body.Date
		transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, transactionDate.Location())

		req := &creditcard.CreateTransactionRequest{
			CreditCardId: *accountEntity.CreditCardId,
			UserId:       userID,
			CategoryId:   categoryID,
			Amount:       transactionAmount,
			Description:  body.Description,
			Date:         transactionDate,
			Installments: 1,
			IsRecurring:  false,
		}

		if err := h.CreditCardService.CreateTransaction(ctx, req); err != nil {
			h.respondError(c, err)
			return
		}

		categoryIDPtr := &categoryID
		transactionEntity := transaction.Transaction{
			Type:        transaction.Expense,
			UserId:      userID,
			AccountId:   accountID,
			CategoryId:  categoryIDPtr,
			Amount:      -body.Amount,
			Description: body.Description,
			Date:        transactionDate,
		}

		if err := h.TransactionService.CreateTransaction(ctx, &transactionEntity); err != nil {
			h.respondError(c, err)
			return
		}

		c.JSON(http.StatusCreated, contracts.TransactionCreateResponse{
			Message:     "Gasto no cartão registrado com sucesso",
			Transaction: transactionEntity,
		})
		return
	}

	categoryID, err := pkg.ParseULID(body.CategoryID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
		return
	}

	categoryIDPtr := &categoryID

	if body.Date == nil {
		h.respondError(c, appErrors.NewValidationError("data", "é obrigatória"))
		return
	}

	transactionDate := *body.Date
	transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, transactionDate.Location())

	transactionAmount := body.Amount
	if transaction.Types(body.Type) == transaction.Expense && transactionAmount > 0 {
		transactionAmount = -transactionAmount
	}

	transactionEntity := transaction.Transaction{
		Type:        transaction.Types(body.Type),
		UserId:      userID,
		AccountId:   accountID,
		CategoryId:  categoryIDPtr,
		Amount:      transactionAmount,
		Description: body.Description,
		Date:        transactionDate,
	}

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

	var filters *transaction.TransactionFilters

	typeStr := c.Query("type")
	if typeStr != "" && typeStr != "ALL" {
		filters = &transaction.TransactionFilters{Type: &typeStr}
	}

	categoryIDStr := c.Query("category_id")
	if categoryIDStr != "" && categoryIDStr != "ALL" {
		parsed, err := pkg.ParseULID(categoryIDStr)
		if err == nil {
			if filters == nil {
				filters = &transaction.TransactionFilters{}
			}
			filters.CategoryID = &parsed
		}
	}

	searchStr := c.Query("search")
	if searchStr != "" {
		if filters == nil {
			filters = &transaction.TransactionFilters{}
		}
		filters.Search = &searchStr
	}

	dateFromStr := c.Query("date_from")
	if dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			if filters == nil {
				filters = &transaction.TransactionFilters{}
			}
			filters.DateFrom = &dateFrom
		}
	}

	dateToStr := c.Query("date_to")
	if dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			if filters == nil {
				filters = &transaction.TransactionFilters{}
			}
			dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.DateTo = &dateTo
		}
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	transactions, total, err := h.TransactionService.GetAllTransactions(ctx, userID, accountID, filters, pagination)
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
		h.respondError(c, appErrors.ParseValidationErrors(err))
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

	categoryIDPtr := &categoryID

	ctx := c.Request.Context()
	storedTransaction, err := h.TransactionService.GetTransactionByID(ctx, transactionID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	transactionDate := storedTransaction.Date
	if body.Date != nil {
		transactionDate = *body.Date
		transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, transactionDate.Location())
	}

	transactionAmount := body.Amount
	if transaction.Types(body.Type) == transaction.Expense && transactionAmount > 0 {
		transactionAmount = -transactionAmount
	}

	transactionEntity := transaction.Transaction{
		Id:          transactionID,
		UserId:      userID,
		AccountId:   accountID,
		CategoryId:  categoryIDPtr,
		Amount:      transactionAmount,
		Description: body.Description,
		Type:        transaction.Types(body.Type),
		Date:        transactionDate,
		UpdatedAt:   pkg.SetTimestamps(),
	}

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
