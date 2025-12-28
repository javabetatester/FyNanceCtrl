package creditcard

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Invoice struct {
	Id             ulid.ULID     `gorm:"type:varchar(26);primaryKey" json:"id"`
	CreditCardId   ulid.ULID      `gorm:"type:varchar(26);index:idx_invoices_credit_card_id;not null" json:"creditCardId"`
	UserId         ulid.ULID      `gorm:"type:varchar(26);index:idx_invoices_user_id;not null" json:"userId"`
	ReferenceMonth int            `gorm:"not null;check:reference_month >= 1 AND reference_month <= 12" json:"referenceMonth"`
	ReferenceYear  int            `gorm:"not null" json:"referenceYear"`
	OpeningDate    time.Time      `gorm:"type:date;not null" json:"openingDate"`
	ClosingDate    time.Time      `gorm:"type:date;not null" json:"closingDate"`
	DueDate        time.Time      `gorm:"type:date;not null;index:idx_invoices_due_date" json:"dueDate"`
	TotalAmount    float64        `gorm:"type:decimal(15,2);not null;default:0" json:"totalAmount"`
	PaidAmount     float64        `gorm:"type:decimal(15,2);not null;default:0" json:"paidAmount"`
	Status         InvoiceStatus  `gorm:"type:varchar(20);not null;default:'OPEN';index:idx_invoices_status" json:"status"`
	PaidAt         *time.Time     `gorm:"type:timestamp" json:"paidAt"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (Invoice) TableName() string {
	return "invoices"
}

type InvoiceStatus string

const (
	InvoiceOpen    InvoiceStatus = "OPEN"
	InvoiceClosed  InvoiceStatus = "CLOSED"
	InvoicePaid    InvoiceStatus = "PAID"
	InvoicePartial InvoiceStatus = "PARTIAL"
	InvoiceOverdue InvoiceStatus = "OVERDUE"
)

func (s InvoiceStatus) IsValid() bool {
	switch s {
	case InvoiceOpen, InvoiceClosed, InvoicePaid, InvoicePartial, InvoiceOverdue:
		return true
	}
	return false
}
