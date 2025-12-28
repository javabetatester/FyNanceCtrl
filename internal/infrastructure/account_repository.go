package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/account"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type AccountRepository struct {
	DB *gorm.DB
}

type accountDB struct {
	Id             string    `gorm:"type:varchar(26);primaryKey"`
	UserId         string    `gorm:"type:varchar(26);index;not null"`
	Name           string    `gorm:"type:varchar(100);not null"`
	Type           string    `gorm:"type:varchar(20);not null"`
	Balance        float64   `gorm:"type:decimal(15,2);not null;default:0"`
	Color          string    `gorm:"type:varchar(7)"`
	Icon           string    `gorm:"type:varchar(50)"`
	IncludeInTotal bool      `gorm:"not null;default:true"`
	IsActive       bool      `gorm:"not null;default:true"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (accountDB) TableName() string {
	return "accounts"
}

func toDomainAccount(adb *accountDB) (*account.Account, error) {
	id, err := pkg.ParseULID(adb.Id)
	if err != nil {
		return nil, err
	}

	userID, err := pkg.ParseULID(adb.UserId)
	if err != nil {
		return nil, err
	}

	return &account.Account{
		Id:             id,
		UserId:         userID,
		Name:           adb.Name,
		Type:           account.AccountType(adb.Type),
		Balance:        adb.Balance,
		Color:          adb.Color,
		Icon:           adb.Icon,
		IncludeInTotal: adb.IncludeInTotal,
		IsActive:       adb.IsActive,
		CreatedAt:      adb.CreatedAt,
		UpdatedAt:      adb.UpdatedAt,
	}, nil
}

func toDBAccount(a *account.Account) *accountDB {
	return &accountDB{
		Id:             a.Id.String(),
		UserId:         a.UserId.String(),
		Name:           a.Name,
		Type:           string(a.Type),
		Balance:        a.Balance,
		Color:          a.Color,
		Icon:           a.Icon,
		IncludeInTotal: a.IncludeInTotal,
		IsActive:       a.IsActive,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	adb := toDBAccount(a)
	return r.DB.WithContext(ctx).Table("accounts").Create(adb).Error
}

func (r *AccountRepository) Update(ctx context.Context, a *account.Account) error {
	adb := toDBAccount(a)
	return r.DB.WithContext(ctx).Model(&accountDB{}).Where("id = ? AND user_id = ?", adb.Id, adb.UserId).Updates(adb).Error
}

func (r *AccountRepository) Delete(ctx context.Context, accountID, userID ulid.ULID) error {
	return r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", accountID.String(), userID.String()).Delete(&accountDB{}).Error
}

func (r *AccountRepository) GetById(ctx context.Context, accountID, userID ulid.ULID) (*account.Account, error) {
	var adb accountDB
	err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", accountID.String(), userID.String()).First(&adb).Error
	if err != nil {
		return nil, err
	}
	return toDomainAccount(&adb)
}

func (r *AccountRepository) GetByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*account.Account, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("accounts").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []accountDB
	err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	accounts := make([]*account.Account, 0, len(rows))
	for i := range rows {
		a, err := toDomainAccount(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, a)
	}
	return accounts, total, nil
}

func (r *AccountRepository) GetActiveByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*account.Account, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("accounts").Where("user_id = ? AND is_active = ?", userID.String(), true)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []accountDB
	err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	accounts := make([]*account.Account, 0, len(rows))
	for i := range rows {
		a, err := toDomainAccount(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, a)
	}
	return accounts, total, nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID ulid.ULID, amount float64) error {
	return r.DB.WithContext(ctx).Model(&accountDB{}).Where("id = ?", accountID.String()).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).
		UpdateColumn("updated_at", time.Now()).Error
}

func (r *AccountRepository) UpdateBalanceWithTx(ctx context.Context, tx interface{}, accountID ulid.ULID, amount float64) error {
	dbTx := tx.(*gorm.DB)
	return dbTx.WithContext(ctx).Model(&accountDB{}).Where("id = ?", accountID.String()).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).
		UpdateColumn("updated_at", time.Now()).Error
}

func (r *AccountRepository) BeginTx(ctx context.Context) (interface{}, error) {
	return r.DB.WithContext(ctx).Begin(), nil
}

func (r *AccountRepository) CommitTx(tx interface{}) error {
	return tx.(*gorm.DB).Commit().Error
}

func (r *AccountRepository) RollbackTx(tx interface{}) error {
	return tx.(*gorm.DB).Rollback().Error
}

func (r *AccountRepository) GetTotalBalance(ctx context.Context, userID ulid.ULID) (float64, error) {
	var total float64
	err := r.DB.WithContext(ctx).Model(&accountDB{}).
		Where("user_id = ? AND is_active = ? AND include_in_total = ?", userID.String(), true, true).
		Select("COALESCE(SUM(balance), 0)").Scan(&total).Error
	return total, err
}
