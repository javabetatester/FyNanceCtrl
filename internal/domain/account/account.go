package account

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Account struct {
	Id             ulid.ULID   `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId         ulid.ULID   `gorm:"type:varchar(26);index:idx_accounts_user_id;not null" json:"userId"`
	Name           string      `gorm:"type:varchar(100);not null" json:"name"`
	Type           AccountType `gorm:"type:varchar(20);not null;index:idx_accounts_type" json:"type"`
	Balance        float64     `gorm:"type:decimal(15,2);not null;default:0" json:"balance"`
	Color          string      `gorm:"type:varchar(7)" json:"color"`
	Icon           string      `gorm:"type:varchar(50)" json:"icon"`
	IncludeInTotal bool        `gorm:"not null;default:true" json:"includeInTotal"`
	IsActive       bool        `gorm:"not null;default:true;index:idx_accounts_active" json:"isActive"`
	CreditCardId   *ulid.ULID  `gorm:"type:varchar(26);index:idx_accounts_credit_card_id" json:"creditCardId,omitempty"`
	CreatedAt      time.Time   `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (Account) TableName() string {
	return "accounts"
}
