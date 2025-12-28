package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/recurring"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type RecurringRepository struct {
	DB *gorm.DB
}

type recurringDB struct {
	Id            string     `gorm:"type:varchar(26);primaryKey"`
	UserId        string     `gorm:"type:varchar(26);index;not null"`
	Type          string     `gorm:"type:varchar(15);not null"`
	CategoryId    string     `gorm:"type:varchar(26);index"`
	AccountId     *string    `gorm:"type:varchar(26);index"`
	Amount        float64    `gorm:"type:decimal(15,2);not null"`
	Description   string     `gorm:"type:varchar(255)"`
	Frequency     string     `gorm:"type:varchar(20);not null"`
	DayOfMonth    int        `gorm:"default:1"`
	DayOfWeek     int        `gorm:"default:0"`
	StartDate     time.Time  `gorm:"type:date;not null"`
	EndDate       *time.Time `gorm:"type:date"`
	LastProcessed *time.Time `gorm:"type:date"`
	NextDue       time.Time  `gorm:"type:date;not null"`
	IsActive      bool       `gorm:"not null;default:true"`
	CreatedAt     time.Time  `gorm:"not null"`
	UpdatedAt     time.Time  `gorm:"not null"`
}

func (recurringDB) TableName() string {
	return "recurring_transactions"
}

func toDomainRecurring(rdb *recurringDB) (*recurring.RecurringTransaction, error) {
	id, err := pkg.ParseULID(rdb.Id)
	if err != nil {
		return nil, err
	}

	userID, err := pkg.ParseULID(rdb.UserId)
	if err != nil {
		return nil, err
	}

	categoryID, err := pkg.ParseULID(rdb.CategoryId)
	if err != nil {
		return nil, err
	}

	var accountID *ulid.ULID
	if rdb.AccountId != nil && *rdb.AccountId != "" {
		parsed, err := pkg.ParseULID(*rdb.AccountId)
		if err != nil {
			return nil, err
		}
		accountID = &parsed
	}

	return &recurring.RecurringTransaction{
		Id:            id,
		UserId:        userID,
		Type:          rdb.Type,
		CategoryId:    categoryID,
		AccountId:     accountID,
		Amount:        rdb.Amount,
		Description:   rdb.Description,
		Frequency:     recurring.FrequencyType(rdb.Frequency),
		DayOfMonth:    rdb.DayOfMonth,
		DayOfWeek:     rdb.DayOfWeek,
		StartDate:     rdb.StartDate,
		EndDate:       rdb.EndDate,
		LastProcessed: rdb.LastProcessed,
		NextDue:       rdb.NextDue,
		IsActive:      rdb.IsActive,
		CreatedAt:     rdb.CreatedAt,
		UpdatedAt:     rdb.UpdatedAt,
	}, nil
}

func toDBRecurring(r *recurring.RecurringTransaction) *recurringDB {
	var accountID *string
	if r.AccountId != nil {
		s := r.AccountId.String()
		accountID = &s
	}

	return &recurringDB{
		Id:            r.Id.String(),
		UserId:        r.UserId.String(),
		Type:          r.Type,
		CategoryId:    r.CategoryId.String(),
		AccountId:     accountID,
		Amount:        r.Amount,
		Description:   r.Description,
		Frequency:     string(r.Frequency),
		DayOfMonth:    r.DayOfMonth,
		DayOfWeek:     r.DayOfWeek,
		StartDate:     r.StartDate,
		EndDate:       r.EndDate,
		LastProcessed: r.LastProcessed,
		NextDue:       r.NextDue,
		IsActive:      r.IsActive,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (r *RecurringRepository) Create(ctx context.Context, rec *recurring.RecurringTransaction) error {
	rdb := toDBRecurring(rec)
	return r.DB.WithContext(ctx).Table("recurring_transactions").Create(rdb).Error
}

func (r *RecurringRepository) Update(ctx context.Context, rec *recurring.RecurringTransaction) error {
	rdb := toDBRecurring(rec)
	return r.DB.WithContext(ctx).Model(&recurringDB{}).Where("id = ? AND user_id = ?", rdb.Id, rdb.UserId).Updates(rdb).Error
}

func (r *RecurringRepository) Delete(ctx context.Context, recurringID, userID ulid.ULID) error {
	return r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", recurringID.String(), userID.String()).Delete(&recurringDB{}).Error
}

func (r *RecurringRepository) GetById(ctx context.Context, recurringID, userID ulid.ULID) (*recurring.RecurringTransaction, error) {
	var rdb recurringDB
	err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", recurringID.String(), userID.String()).First(&rdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainRecurring(&rdb)
}

func (r *RecurringRepository) GetByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*recurring.RecurringTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("recurring_transactions").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []recurringDB
	err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*recurring.RecurringTransaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainRecurring(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	return transactions, total, nil
}

func (r *RecurringRepository) GetActiveByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*recurring.RecurringTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("recurring_transactions").Where("user_id = ? AND is_active = ?", userID.String(), true)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []recurringDB
	err := baseQuery.Order("next_due ASC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*recurring.RecurringTransaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainRecurring(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	return transactions, total, nil
}

func (r *RecurringRepository) GetDueTransactions(ctx context.Context, date time.Time, pagination *pkg.PaginationParams) ([]*recurring.RecurringTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("recurring_transactions").Where("is_active = ? AND next_due <= ?", true, date)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []recurringDB
	err := baseQuery.Order("next_due ASC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*recurring.RecurringTransaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainRecurring(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	return transactions, total, nil
}

func (r *RecurringRepository) UpdateLastProcessed(ctx context.Context, recurringID ulid.ULID, processedDate, nextDue time.Time) error {
	return r.DB.WithContext(ctx).Model(&recurringDB{}).Where("id = ?", recurringID.String()).
		Updates(map[string]interface{}{
			"last_processed": processedDate,
			"next_due":       nextDue,
			"updated_at":     time.Now(),
		}).Error
}
