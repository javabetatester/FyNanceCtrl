package creditcard

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type CreditCard struct {
	Id             ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId         ulid.ULID `gorm:"type:varchar(26);index:idx_credit_cards_user_id;not null" json:"userId"`
	AccountId      ulid.ULID `gorm:"type:varchar(26);index:idx_credit_cards_account_id;not null" json:"accountId"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	CreditLimit    float64   `gorm:"type:decimal(15,2);not null" json:"creditLimit"`
	AvailableLimit float64   `gorm:"type:decimal(15,2);not null" json:"availableLimit"`
	ClosingDay     int       `gorm:"not null;check:closing_day >= 1 AND closing_day <= 31" json:"closingDay"`
	DueDay         int       `gorm:"not null;check:due_day >= 1 AND due_day <= 31" json:"dueDay"`
	Brand          CardBrand `gorm:"type:varchar(20);not null" json:"brand"`
	LastFourDigits string    `gorm:"type:varchar(4)" json:"lastFourDigits"`
	IsActive       bool      `gorm:"not null;default:true;index:idx_credit_cards_active" json:"isActive"`
	CreatedAt      time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (CreditCard) TableName() string {
	return "credit_cards"
}
