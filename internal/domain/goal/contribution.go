package goal

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type ContributionType string

const (
	ContributionDeposit  ContributionType = "DEPOSIT"
	ContributionWithdraw ContributionType = "WITHDRAW"
)

type Contribution struct {
	Id            ulid.ULID        `gorm:"type:varchar(26);primaryKey" json:"id"`
	GoalId        ulid.ULID        `gorm:"type:varchar(26);index:idx_contributions_goal_id;not null" json:"goalId"`
	UserId        ulid.ULID        `gorm:"type:varchar(26);index:idx_contributions_user_id;not null" json:"userId"`
	AccountId     ulid.ULID        `gorm:"type:varchar(26);index:idx_contributions_account_id;not null" json:"accountId"`
	TransactionId *ulid.ULID       `gorm:"type:varchar(26);index:idx_contributions_transaction_id" json:"transactionId,omitempty"`
	Type          ContributionType `gorm:"type:varchar(20);not null" json:"type"`
	Amount        float64          `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description   string           `gorm:"type:varchar(255)" json:"description"`
	CreatedAt     time.Time        `gorm:"autoCreateTime;not null" json:"createdAt"`
}

func (Contribution) TableName() string {
	return "goal_contributions"
}
