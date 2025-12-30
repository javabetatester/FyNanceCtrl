package contracts

import (
	"time"

	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/transaction"
)

type RecurringCreateRequest struct {
	Type        string     `json:"type" binding:"required,oneof=RECEIPT EXPENSE"`
	CategoryId  string     `json:"category_id" binding:"required"`
	AccountId   string     `json:"account_id" binding:"omitempty"`
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description string     `json:"description" binding:"omitempty,max=255"`
	Frequency   string     `json:"frequency" binding:"required,oneof=DAILY WEEKLY MONTHLY YEARLY"`
	DayOfMonth  int        `json:"day_of_month" binding:"omitempty,min=1,max=31"`
	DayOfWeek   int        `json:"day_of_week" binding:"omitempty,min=0,max=6"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date" binding:"omitempty"`
}

type RecurringUpdateRequest struct {
	Amount      *float64   `json:"amount" binding:"omitempty,gt=0"`
	Description *string    `json:"description" binding:"omitempty,max=255"`
	IsActive    *bool      `json:"is_active" binding:"omitempty"`
	EndDate     *time.Time `json:"end_date" binding:"omitempty"`
	NextDue     *time.Time `json:"next_due" binding:"omitempty"`
}

type RecurringCreateResponse struct {
	Message   string                          `json:"message"`
	Recurring *recurring.RecurringTransaction `json:"recurring"`
}

type RecurringListResponse struct {
	Recurring []*recurring.RecurringTransaction `json:"recurring"`
	Total     int                               `json:"total"`
}

type RecurringSingleResponse struct {
	Recurring *recurring.RecurringTransaction `json:"recurring"`
}

type RecurringProcessRequest struct {
	ProcessDate *time.Time `json:"process_date" binding:"omitempty"`
}

type RecurringProcessResponse struct {
	Message     string                          `json:"message"`
	Transaction *transaction.Transaction        `json:"transaction"`
	Recurring   *recurring.RecurringTransaction `json:"recurring"`
}
