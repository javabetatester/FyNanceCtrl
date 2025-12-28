package contracts

import (
	"time"

	"Fynance/internal/domain/transaction"
)

type TransactionCreateRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=RECEIPT EXPENSE TRANSFER GOALS INVESTMENT WITHDRAW"`
	CategoryID  string  `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty,max=255"`
}

type TransactionUpdateRequest struct {
	AccountID   string     `json:"account_id" binding:"required"`
	Type        string     `json:"type" binding:"required,oneof=RECEIPT EXPENSE TRANSFER GOALS INVESTMENT WITHDRAW"`
	CategoryID  string     `json:"category_id" binding:"required"`
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description string     `json:"description" binding:"omitempty,max=255"`
	Date        *time.Time `json:"date"`
}

type CategoryCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Icon string `json:"icon" binding:"omitempty,max=50"`
}

type CategoryUpdateRequest struct {
	Name string `json:"name" binding:"required"`
	Icon string `json:"icon" binding:"omitempty,max=50"`
}

type TransactionCreateResponse struct {
	Message     string                  `json:"message"`
	Transaction transaction.Transaction `json:"transaction"`
}

type TransactionListResponse struct {
	Transactions []*transaction.Transaction `json:"transactions"`
	Total        int                        `json:"total"`
}

type TransactionSingleResponse struct {
	Transaction *transaction.Transaction `json:"transaction"`
}

type CategoryListResponse struct {
	Categories []*transaction.Category `json:"categories"`
	Total      int                     `json:"total"`
}

type CategoryResponse struct {
	Category *transaction.Category `json:"category"`
}
