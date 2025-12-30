package routes

import (
	"net/http"

	"Fynance/internal/contracts"
	"Fynance/internal/domain/creditcard"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateCreditCard(c *gin.Context) {
	var body contracts.CreditCardCreateRequest
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

	req := &creditcard.CreateCreditCardRequest{
		UserId:         userID,
		AccountId:      accountID,
		Name:           body.Name,
		CreditLimit:    body.CreditLimit,
		ClosingDay:     body.ClosingDay,
		DueDay:         body.DueDay,
		Brand:          creditcard.CardBrand(body.Brand),
		LastFourDigits: body.LastFourDigits,
	}

	ctx := c.Request.Context()
	card, err := h.CreditCardService.CreateCreditCard(ctx, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.CreditCardCreateResponse{
		Message:    "Cartao de credito criado com sucesso",
		CreditCard: card,
	})
}

func (h *Handler) ListCreditCards(c *gin.Context) {
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	cards, total, err := h.CreditCardService.ListCreditCards(ctx, userID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(cards, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetCreditCard(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
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
	card, err := h.CreditCardService.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.CreditCardSingleResponse{CreditCard: card})
}

func (h *Handler) UpdateCreditCard(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.CreditCardUpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	req := &creditcard.UpdateCreditCardRequest{
		Name:           body.Name,
		LastFourDigits: body.LastFourDigits,
		IsActive:       body.IsActive,
	}

	if body.CreditLimit != nil {
		req.CreditLimit = body.CreditLimit
	}
	if body.ClosingDay != nil {
		req.ClosingDay = body.ClosingDay
	}
	if body.DueDay != nil {
		req.DueDay = body.DueDay
	}
	if body.Brand != nil {
		brand := creditcard.CardBrand(*body.Brand)
		req.Brand = &brand
	}

	ctx := c.Request.Context()
	if err := h.CreditCardService.UpdateCreditCard(ctx, cardID, userID, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Cartao de credito atualizado com sucesso"})
}

func (h *Handler) DeleteCreditCard(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
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
	if err := h.CreditCardService.DeleteCreditCard(ctx, cardID, userID); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Cartao de credito removido com sucesso"})
}

func (h *Handler) ListInvoices(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	invoices, total, err := h.CreditCardService.ListInvoices(ctx, cardID, userID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(invoices, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetCurrentInvoice(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
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
	invoice, err := h.CreditCardService.GetCurrentInvoice(ctx, cardID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.InvoiceSingleResponse{Invoice: invoice})
}

func (h *Handler) GetInvoice(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	invoiceID, err := pkg.ParseULID(c.Param("invoiceId"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("invoice_id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	ctx := c.Request.Context()
	invoice, err := h.CreditCardService.GetInvoiceById(ctx, invoiceID, userID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	if invoice.CreditCardId != cardID {
		h.respondError(c, appErrors.NewValidationError("invoice_id", "fatura nao pertence a este cartao"))
		return
	}

	c.JSON(http.StatusOK, contracts.InvoiceSingleResponse{Invoice: invoice})
}

func (h *Handler) PayInvoice(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	invoiceID, err := pkg.ParseULID(c.Param("invoiceId"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("invoice_id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.InvoicePayRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	accountID, err := pkg.ParseULID(body.AccountID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("account_id", "formato inválido"))
		return
	}

	ctx := c.Request.Context()
	if err := h.CreditCardService.PayInvoice(ctx, cardID, invoiceID, accountID, userID, body.Amount); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, contracts.MessageResponse{Message: "Fatura paga com sucesso"})
}

func (h *Handler) CreateCreditCardTransaction(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	var body contracts.CreditCardTransactionCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, appErrors.ErrBadRequest.WithError(err))
		return
	}

	categoryID, err := pkg.ParseULID(body.CategoryID)
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("category_id", "formato inválido"))
		return
	}

	installments := body.Installments
	if installments < 1 {
		installments = 1
	}

	req := &creditcard.CreateTransactionRequest{
		CreditCardId: cardID,
		UserId:       userID,
		CategoryId:   categoryID,
		Amount:       body.Amount,
		Description:  body.Description,
		Date:         body.Date,
		Installments: installments,
		IsRecurring:  body.IsRecurring,
	}

	ctx := c.Request.Context()
	if err := h.CreditCardService.CreateTransaction(ctx, req); err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, contracts.MessageResponse{Message: "Gasto no cartao registrado com sucesso"})
}

func (h *Handler) ListCreditCardTransactions(c *gin.Context) {
	cardID, err := pkg.ParseULID(c.Param("id"))
	if err != nil {
		h.respondError(c, appErrors.NewValidationError("id", "formato inválido"))
		return
	}

	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		h.respondError(c, err)
		return
	}

	pagination := h.parsePagination(c)

	ctx := c.Request.Context()
	transactions, total, err := h.CreditCardService.ListTransactions(ctx, cardID, userID, pagination)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response := pkg.NewPaginatedResponse(transactions, pagination.Page, pagination.Limit, total)
	c.JSON(http.StatusOK, response)
}
