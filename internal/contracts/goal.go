package contracts

import "time"

type GoalCreateRequest struct {
	Name   string     `json:"name" binding:"required"`
	Target float64    `json:"target" binding:"required,gt=0"`
	EndAt  *time.Time `json:"end_at"`
}

type GoalUpdateRequest struct {
	Name   string     `json:"name" binding:"required"`
	Target float64    `json:"target" binding:"required,gt=0"`
	EndAt  *time.Time `json:"end_at"`
}

type GoalContributionRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty,max=255"`
}

type GoalWithdrawRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty,max=255"`
}
