package contracts

import "Fynance/internal/domain/investment"

type InvestmentCreateRequest struct {
	AccountID     string  `json:"account_id" binding:"required"`
	Type          string  `json:"type" binding:"required,oneof=CDB LCI LCA TESOURO_DIRETO ACOES FUNDOS CRIPTOMOEDAS PREVIDENCIA"`
	Name          string  `json:"name" binding:"required"`
	InitialAmount float64 `json:"initial_amount" binding:"required,gt=0"`
	ReturnRate    float64 `json:"return_rate" binding:"omitempty"`
	CategoryID    string  `json:"category_id" binding:"omitempty"`
}

type InvestmentUpdateRequest struct {
	Name           *string  `json:"name" binding:"omitempty"`
	Type           *string  `json:"type" binding:"omitempty,oneof=CDB LCI LCA TESOURO_DIRETO ACOES FUNDOS CRIPTOMOEDAS PREVIDENCIA"`
	CurrentBalance *float64 `json:"current_balance" binding:"omitempty,gt=0"`
}

type InvestmentContributionRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	CategoryID  string  `json:"category_id" binding:"omitempty"`
	Description string  `json:"description" binding:"omitempty"`
}

type InvestmentWithdrawRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	CategoryID  string  `json:"category_id" binding:"omitempty"`
	Description string  `json:"description" binding:"omitempty"`
}

type InvestmentReturnResponse struct {
	Profit           float64 `json:"profit"`
	ReturnPercentage float64 `json:"returnPercentage"`
}

type InvestmentCreateResponse struct {
	Message    string                `json:"message"`
	Investment investment.Investment `json:"investment"`
}

type InvestmentListResponse struct {
	Investments []*investment.Investment `json:"investments"`
	Total       int                      `json:"total"`
}

type InvestmentSingleResponse struct {
	Investment *investment.Investment `json:"investment"`
}
