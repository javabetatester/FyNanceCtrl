package contracts

import "Fynance/internal/domain/account"

type AccountCreateRequest struct {
	Name           string  `json:"name" binding:"required,max=100"`
	Type           string  `json:"type" binding:"required,oneof=CHECKING SAVINGS CREDIT_CARD CASH INVESTMENT OTHER"`
	InitialBalance float64 `json:"initial_balance" binding:"omitempty"`
	Color          string  `json:"color" binding:"omitempty,max=7"`
	Icon           string  `json:"icon" binding:"omitempty,max=50"`
	IncludeInTotal *bool   `json:"include_in_total" binding:"omitempty"`
}

type AccountUpdateRequest struct {
	Name           *string `json:"name" binding:"omitempty,max=100"`
	Type           *string `json:"type" binding:"omitempty,oneof=CHECKING SAVINGS CREDIT_CARD CASH INVESTMENT OTHER"`
	Color          *string `json:"color" binding:"omitempty,max=7"`
	Icon           *string `json:"icon" binding:"omitempty,max=50"`
	IncludeInTotal *bool   `json:"include_in_total" binding:"omitempty"`
	IsActive       *bool   `json:"is_active" binding:"omitempty"`
}

type AccountTransferRequest struct {
	FromAccountId string  `json:"from_account_id" binding:"required"`
	ToAccountId   string  `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description" binding:"omitempty,max=255"`
}

type AccountCreateResponse struct {
	Message string           `json:"message"`
	Account *account.Account `json:"account"`
}

type AccountListResponse struct {
	Accounts []*account.Account `json:"accounts"`
	Total    int                `json:"total"`
}

type AccountSingleResponse struct {
	Account *account.Account `json:"account"`
}

type AccountBalanceResponse struct {
	TotalBalance float64 `json:"totalBalance"`
}
