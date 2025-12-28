package contracts

import (
	"time"

	"Fynance/internal/domain/creditcard"
)

type CreditCardCreateRequest struct {
	AccountID      string  `json:"account_id" binding:"required"`
	Name           string  `json:"name" binding:"required,max=100"`
	CreditLimit    float64 `json:"credit_limit" binding:"required,gt=0"`
	ClosingDay     int     `json:"closing_day" binding:"required,min=1,max=31"`
	DueDay         int     `json:"due_day" binding:"required,min=1,max=31"`
	Brand          string  `json:"brand" binding:"required,oneof=VISA MASTERCARD ELO AMEX HIPERCARD OTHER"`
	LastFourDigits string  `json:"last_four_digits" binding:"omitempty,max=4"`
}

type CreditCardUpdateRequest struct {
	Name           *string  `json:"name" binding:"omitempty,max=100"`
	CreditLimit    *float64 `json:"credit_limit" binding:"omitempty,gt=0"`
	ClosingDay     *int     `json:"closing_day" binding:"omitempty,min=1,max=31"`
	DueDay         *int     `json:"due_day" binding:"omitempty,min=1,max=31"`
	Brand          *string  `json:"brand" binding:"omitempty,oneof=VISA MASTERCARD ELO AMEX HIPERCARD OTHER"`
	LastFourDigits *string  `json:"last_four_digits" binding:"omitempty,max=4"`
	IsActive       *bool    `json:"is_active" binding:"omitempty"`
}

type CreditCardTransactionCreateRequest struct {
	CategoryID  string    `json:"category_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Description string    `json:"description" binding:"omitempty,max=255"`
	Date        time.Time `json:"date" binding:"required"`
	Installments int      `json:"installments" binding:"omitempty,min=1"`
	IsRecurring bool      `json:"is_recurring" binding:"omitempty"`
}

type InvoicePayRequest struct {
	AccountID string  `json:"account_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

type CreditCardCreateResponse struct {
	Message    string                `json:"message"`
	CreditCard *creditcard.CreditCard `json:"creditCard"`
}

type CreditCardListResponse struct {
	CreditCards []*creditcard.CreditCard `json:"creditCards"`
	Total       int                       `json:"total"`
}

type CreditCardSingleResponse struct {
	CreditCard *creditcard.CreditCard `json:"creditCard"`
}

type InvoiceListResponse struct {
	Invoices []*creditcard.Invoice `json:"invoices"`
	Total    int                   `json:"total"`
}

type InvoiceSingleResponse struct {
	Invoice *creditcard.Invoice `json:"invoice"`
}

type CreditCardTransactionListResponse struct {
	Transactions []*creditcard.CreditCardTransaction `json:"transactions"`
	Total        int                                  `json:"total"`
}
