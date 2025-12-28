package investment

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Investment struct {
	Id              ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	UserId          ulid.ULID `gorm:"type:varchar(26);index:idx_investments_user_id;not null" json:"userId"`
	Type            Types     `gorm:"type:varchar(20);not null;index:idx_investments_type" json:"type"`
	Name            string    `gorm:"type:varchar(100);not null;index:idx_investments_user_name,unique" json:"name"`
	CurrentBalance  float64   `gorm:"type:decimal(15,2);not null;default:0" json:"currentBalance"`
	ReturnBalance   float64   `gorm:"type:decimal(15,2);not null;default:0" json:"returnBalance"`
	ReturnRate      float64   `gorm:"type:decimal(5,2);default:0" json:"returnRate"`
	ApplicationDate time.Time `gorm:"type:date;not null;index:idx_investments_app_date" json:"applicationDate"`
	CreatedAt       time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`
}

func (Investment) TableName() string {
	return "investments"
}
