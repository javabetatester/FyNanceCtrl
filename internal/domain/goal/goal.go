package goal

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Goal struct {
	Id            ulid.ULID  `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId        ulid.ULID  `gorm:"type:varchar(26);index:idx_goals_user_id;not null" json:"userId"`
	Name          string     `gorm:"type:varchar(100);not null;index:idx_goals_user_name" json:"name"`
	TargetAmount  float64    `gorm:"type:decimal(15,2);not null" json:"targetAmount"`
	CurrentAmount float64    `gorm:"type:decimal(15,2);not null;default:0" json:"currentAmount"`
	StartedAt     time.Time  `gorm:"type:timestamp" json:"startedAt"`
	EndedAt       *time.Time `gorm:"type:timestamp" json:"endedAt"`
	Status        GoalStatus `gorm:"type:varchar(20);default:'ACTIVE';index:idx_goals_status" json:"status"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;not null" json:"updatedAt"`

	Icon            string   `gorm:"type:varchar(50);default:'target'" json:"icon"`
	Color           string   `gorm:"type:varchar(7);default:'#6366f1'" json:"color"`
	ImageUrl        *string  `gorm:"type:text" json:"imageUrl"`
	Priority        int      `gorm:"default:3" json:"priority"`
	LastMilestone   int      `gorm:"default:0" json:"lastMilestone"`
	TotalContribs   int      `gorm:"default:0" json:"totalContribs"`
	SuggestedAmount *float64 `gorm:"type:decimal(15,2)" json:"suggestedAmount"`
}

func (Goal) TableName() string {
	return "goals"
}

// GetProgress retorna a porcentagem de progresso da meta
func (g *Goal) GetProgress() float64 {
	if g.TargetAmount == 0 {
		return 0
	}
	return (g.CurrentAmount / g.TargetAmount) * 100
}

// GetCurrentMilestone retorna o milestone atual (25, 50, 75, 100)
func (g *Goal) GetCurrentMilestone() int {
	progress := g.GetProgress()
	if progress >= 100 {
		return 100
	} else if progress >= 75 {
		return 75
	} else if progress >= 50 {
		return 50
	} else if progress >= 25 {
		return 25
	}
	return 0
}

// HasNewMilestone verifica se há um novo milestone a ser alcançado
func (g *Goal) HasNewMilestone() bool {
	currentMilestone := g.GetCurrentMilestone()
	return currentMilestone > g.LastMilestone
}
