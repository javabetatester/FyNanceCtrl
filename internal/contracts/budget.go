package contracts

import "Fynance/internal/domain/budget"

type BudgetCreateRequest struct {
	CategoryId  string  `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Month       int     `json:"month" binding:"required,min=1,max=12"`
	Year        int     `json:"year" binding:"required,min=2000,max=2100"`
	AlertAt     float64 `json:"alert_at" binding:"omitempty,min=0,max=100"`
	IsRecurring bool    `json:"is_recurring"`
}

type BudgetUpdateRequest struct {
	Amount      *float64 `json:"amount" binding:"omitempty,gt=0"`
	AlertAt     *float64 `json:"alert_at" binding:"omitempty,min=0,max=100"`
	IsRecurring *bool    `json:"is_recurring"`
}

type BudgetCreateResponse struct {
	Message string         `json:"message"`
	Budget  *budget.Budget `json:"budget"`
}

type BudgetListResponse struct {
	Budgets []*budget.Budget `json:"budgets"`
	Total   int              `json:"total"`
}

type BudgetResponse struct {
	*budget.Budget
	Percentage   float64 `json:"percentage"`
	Remaining    float64 `json:"remaining"`
	SpentAmount  float64 `json:"spentAmount"`
	BudgetAmount float64 `json:"budgetAmount"`
	Status       string  `json:"status"`
}

type BudgetSingleResponse struct {
	Budget *budget.Budget `json:"budget"`
}

type BudgetSummaryResponse struct {
	Summary *budget.BudgetSummary `json:"summary"`
}

type BudgetStatusResponse struct {
	BudgetId   string  `json:"budgetId"`
	Amount     float64 `json:"amount"`
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
	AlertAt    float64 `json:"alertAt"`
}
