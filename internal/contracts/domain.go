package contracts

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type GoalCreateRequestDomain struct {
	UserId  ulid.ULID  `json:"user_id"`
	Name    string     `json:"name"`
	Target  float64    `json:"target"`
	EndedAt *time.Time `json:"end_at"`
}

type GoalUpdateRequestDomain struct {
	Id      ulid.ULID  `json:"id"`
	UserId  ulid.ULID  `json:"user_id"`
	Name    string     `json:"name"`
	Target  float64    `json:"target"`
	EndedAt *time.Time `json:"end_at"`
}

type CreateInvestmentRequestDomain struct {
	UserId        ulid.ULID `json:"user_id"`
	AccountId     ulid.ULID `json:"account_id"`
	CategoryId    ulid.ULID `json:"category_id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	InitialAmount float64   `json:"initial_amount"`
	ReturnRate    float64   `json:"return_rate"`
}

type ContributionRequestDomain struct {
	UserId      ulid.ULID `json:"user_id"`
	AccountId   ulid.ULID `json:"account_id"`
	Id          ulid.ULID `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

type WithdrawRequestDomain struct {
	UserId      ulid.ULID `json:"user_id"`
	Id          ulid.ULID `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

type UpdateInvestmentRequestDomain struct {
	UserId         ulid.ULID `json:"user_id"`
	Id             ulid.ULID `json:"id"`
	Name           *string   `json:"name,omitempty"`
	Type           *string   `json:"type,omitempty"`
	CurrentBalance *float64  `json:"current_balance,omitempty"`
}
