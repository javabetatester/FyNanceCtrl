package routes

import "Fynance/internal/domain/investment"

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
