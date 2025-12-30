package user

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type User struct {
	Id        ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex:idx_users_email;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`
	Plan          Plan      `gorm:"type:varchar(10);default:'FREE';index:idx_users_plan" json:"plan"`
	PlanSince     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"planSince"`
	OnboardingStep int      `gorm:"default:0" json:"onboardingStep"`
}

func (User) TableName() string {
	return "users"
}

type Plan string

const (
	PlanFree  Plan = "FREE"
	PlanBasic Plan = "BASIC"
	PlanPro   Plan = "PRO"
)

func (p Plan) IsValid() bool {
	switch p {
	case PlanFree, PlanBasic, PlanPro:
		return true
	}
	return false
}
