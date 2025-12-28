package recurring

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type RecurringTransaction struct {
	Id            ulid.ULID     `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId        ulid.ULID     `gorm:"type:varchar(26);index:idx_recurring_user_id;not null" json:"userId"`
	Type          string        `gorm:"type:varchar(15);not null" json:"type"`
	CategoryId    ulid.ULID     `gorm:"type:varchar(26);index:idx_recurring_category_id" json:"categoryId"`
	AccountId     *ulid.ULID    `gorm:"type:varchar(26);index:idx_recurring_account_id" json:"accountId"`
	Amount        float64       `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description   string        `gorm:"type:varchar(255)" json:"description"`
	Frequency     FrequencyType `gorm:"type:varchar(20);not null" json:"frequency"`
	DayOfMonth    int           `gorm:"default:1" json:"dayOfMonth"`
	DayOfWeek     int           `gorm:"default:0" json:"dayOfWeek"`
	StartDate     time.Time     `gorm:"type:date;not null" json:"startDate"`
	EndDate       *time.Time    `gorm:"type:date" json:"endDate"`
	LastProcessed *time.Time    `gorm:"type:date" json:"lastProcessed"`
	NextDue       time.Time     `gorm:"type:date;not null;index:idx_recurring_next_due" json:"nextDue"`
	IsActive      bool          `gorm:"not null;default:true;index:idx_recurring_active" json:"isActive"`
	CreatedAt     time.Time     `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (RecurringTransaction) TableName() string {
	return "recurring_transactions"
}

type FrequencyType string

const (
	FrequencyDaily   FrequencyType = "DAILY"
	FrequencyWeekly  FrequencyType = "WEEKLY"
	FrequencyMonthly FrequencyType = "MONTHLY"
	FrequencyYearly  FrequencyType = "YEARLY"
)

func (f FrequencyType) IsValid() bool {
	switch f {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		return true
	}
	return false
}
