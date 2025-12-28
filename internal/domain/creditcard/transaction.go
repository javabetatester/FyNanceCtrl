package creditcard

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type CreditCardTransaction struct {
	Id                ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	CreditCardId      ulid.ULID `gorm:"type:varchar(26);index:idx_cc_transactions_card_id;not null" json:"creditCardId"`
	InvoiceId         ulid.ULID `gorm:"type:varchar(26);index:idx_cc_transactions_invoice_id;not null" json:"invoiceId"`
	UserId            ulid.ULID `gorm:"type:varchar(26);index:idx_cc_transactions_user_id;not null" json:"userId"`
	CategoryId        ulid.ULID `gorm:"type:varchar(26);index:idx_cc_transactions_category_id;not null" json:"categoryId"`
	Amount            float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description       string     `gorm:"type:varchar(255)" json:"description"`
	Date              time.Time  `gorm:"type:date;not null;index:idx_cc_transactions_date" json:"date"`
	Installments      int        `gorm:"not null;default:1;check:installments >= 1" json:"installments"`
	CurrentInstallment int       `gorm:"not null;default:1;check:current_installment >= 1" json:"currentInstallment"`
	IsRecurring       bool       `gorm:"not null;default:false" json:"isRecurring"`
	CreatedAt         time.Time  `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (CreditCardTransaction) TableName() string {
	return "credit_card_transactions"
}
