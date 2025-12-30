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

var _ recurring.RecurringRepository = (*RecurringRepository)(nil)

type recurringDB struct {
	Id            string     `gorm:"type:varchar(26);primaryKey;column:id"`
	UserId        string     `gorm:"type:varchar(26);index;not null;column:user_id"`
	Type          string     `gorm:"type:varchar(15);not null;column:type"`
	CategoryId    string     `gorm:"type:varchar(26);index;column:category_id"`
	CategoryName  string     `gorm:"->;column:category_name"`
	AccountId     *string    `gorm:"type:varchar(26);index;column:account_id"`
	Amount        float64    `gorm:"type:decimal(15,2);not null;column:amount"`
	Description   string     `gorm:"type:varchar(255);column:description"`
	Frequency     string     `gorm:"type:varchar(20);not null;column:frequency"`
	DayOfMonth    int        `gorm:"default:1;column:day_of_month"`
	DayOfWeek     int        `gorm:"default:0;column:day_of_week"`
	StartDate     time.Time  `gorm:"type:date;not null;column:start_date"`
	EndDate       *time.Time `gorm:"type:date;column:end_date"`
	LastProcessed *time.Time `gorm:"type:date;column:last_processed"`
	NextDue       time.Time  `gorm:"type:date;not null;column:next_due"`
	IsActive      bool       `gorm:"not null;default:true;column:is_active"`
	CreatedAt     time.Time  `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time  `gorm:"not null;column:updated_at"`
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

	rec := &recurring.RecurringTransaction{
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
	}
	if rdb.CategoryName != "" {
		rec.CategoryName = rdb.CategoryName
	}
	return rec, nil
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

func (r *RecurringRepository) GetByID(ctx context.Context, recurringID, userID ulid.ULID) (*recurring.RecurringTransaction, error) {
	var rdb recurringDB
	err := r.DB.WithContext(ctx).
		Table("recurring_transactions r").
		Select("r.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON r.category_id = c.id").
		Where("r.id = ? AND r.user_id = ?", recurringID.String(), userID.String()).
		Order("r.id").
		First(&rdb).Error
	if err != nil {
		return nil, err
	}
	rec, err := toDomainRecurring(&rdb)
	if err != nil {
		return nil, err
	}
	if rdb.CategoryName != "" {
		rec.CategoryName = rdb.CategoryName
	}
	return rec, nil
}

func (r *RecurringRepository) GetByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*recurring.RecurringTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	countQuery := r.DB.WithContext(ctx).Table("recurring_transactions r").Where("r.user_id = ?", userID.String())
	dataQuery := r.DB.WithContext(ctx).
		Table("recurring_transactions r").
		Select("r.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON r.category_id = c.id").
		Where("r.user_id = ?", userID.String())

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []recurringDB
	err := dataQuery.Order("r.created_at DESC").
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
			continue
		}
		transactions = append(transactions, t)
	}
	return transactions, total, nil
}

func (r *RecurringRepository) GetActiveByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*recurring.RecurringTransaction, int64, error) {
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
