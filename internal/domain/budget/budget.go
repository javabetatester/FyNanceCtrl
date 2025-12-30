package budget

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Budget struct {
	Id           ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId       ulid.ULID `gorm:"type:varchar(26);index:idx_budgets_user_id;not null" json:"userId"`
	CategoryId   ulid.ULID `gorm:"type:varchar(26);index:idx_budgets_category;not null" json:"categoryId"`
	CategoryName string    `gorm:"-" json:"categoryName,omitempty"`
	Amount       float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Spent        float64   `gorm:"type:decimal(15,2);not null;default:0" json:"spent"`
	Month        int       `gorm:"type:integer;not null;index:idx_budgets_period" json:"month"`
	Year         int       `gorm:"type:integer;not null;index:idx_budgets_period" json:"year"`
	AlertAt      float64   `gorm:"type:decimal(5,2);default:80" json:"alertAt"`
	IsRecurring  bool      `gorm:"not null;default:false" json:"isRecurring"`
	CreatedAt    time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`

	GroupId       *ulid.ULID `gorm:"type:varchar(26);index:idx_budgets_group" json:"groupId"`
	GroupName     string     `gorm:"-" json:"groupName,omitempty"`
	Color         string     `gorm:"type:varchar(7)" json:"color"`
	Icon          string     `gorm:"type:varchar(50)" json:"icon"`
	Priority      int        `gorm:"default:3" json:"priority"`
	HealthScore   int        `gorm:"default:100" json:"healthScore"`
	SavingsAmount float64    `gorm:"type:decimal(15,2);default:0" json:"savingsAmount"`
	TotalSaved    float64    `gorm:"type:decimal(15,2);default:0" json:"totalSaved"`
}

func (Budget) TableName() string {
	return "budgets"
}

// GetPercentage retorna a porcentagem gasta do orçamento
func (b *Budget) GetPercentage() float64 {
	if b.Amount == 0 {
		return 0
	}
	return (b.Spent / b.Amount) * 100
}

// GetRemaining retorna quanto ainda pode gastar
func (b *Budget) GetRemaining() float64 {
	remaining := b.Amount - b.Spent
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CalculateHealthScore calcula o score de saúde baseado no gasto
func (b *Budget) CalculateHealthScore() int {
	percentage := b.GetPercentage()
	if percentage <= 50 {
		return 100
	} else if percentage <= 70 {
		return 85
	} else if percentage <= 85 {
		return 70
	} else if percentage <= 95 {
		return 50
	} else if percentage <= 100 {
		return 30
	}
	return 10 // Acima do limite
}

// GetStatus retorna o status do orçamento
func (b *Budget) GetStatus() string {
	percentage := b.GetPercentage()
	if percentage >= 100 {
		return "exceeded"
	} else if percentage >= b.AlertAt {
		return "warning"
	}
	return "ok"
}

// IsWithinBudget verifica se está dentro do orçamento
func (b *Budget) IsWithinBudget() bool {
	return b.Spent <= b.Amount
}

type BudgetSummary struct {
	TotalBudget    float64 `json:"totalBudget"`
	TotalSpent     float64 `json:"totalSpent"`
	TotalRemaining float64 `json:"totalRemaining"`
	Percentage     float64 `json:"percentage"`
}
