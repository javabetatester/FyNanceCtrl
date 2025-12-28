package contracts

import (
	"time"

	domainGoal "Fynance/internal/domain/goal"
)

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

type GoalResponse struct {
	Goal *domainGoal.Goal `json:"goal"`
}

type GoalListResponse struct {
	Goals []*domainGoal.Goal `json:"goals"`
	Total int                `json:"total"`
}

type GoalContributionListResponse struct {
	Contributions []*domainGoal.Contribution `json:"contributions"`
	Total         int                        `json:"total"`
}

type GoalProgressResponse struct {
	Progress *domainGoal.GoalProgress `json:"progress"`
}
