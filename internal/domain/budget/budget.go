package budget

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Budget struct {
	Id          ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId      ulid.ULID `gorm:"type:varchar(26);index:idx_budgets_user_id;not null" json:"userId"`
	CategoryId  ulid.ULID `gorm:"type:varchar(26);index:idx_budgets_category;not null" json:"categoryId"`
	Amount      float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Spent       float64   `gorm:"type:decimal(15,2);not null;default:0" json:"spent"`
	Month       int       `gorm:"not null;index:idx_budgets_period" json:"month"`
	Year        int       `gorm:"not null;index:idx_budgets_period" json:"year"`
	AlertAt     float64   `gorm:"type:decimal(5,2);default:80" json:"alertAt"`
	IsRecurring bool      `gorm:"not null;default:false" json:"isRecurring"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (Budget) TableName() string {
	return "budgets"
}

type BudgetSummary struct {
	TotalBudget    float64 `json:"totalBudget"`
	TotalSpent     float64 `json:"totalSpent"`
	TotalRemaining float64 `json:"totalRemaining"`
	Percentage     float64 `json:"percentage"`
}
