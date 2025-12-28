package infrastructure

import (
	"Fynance/internal/domain/transaction"
	"Fynance/internal/pkg"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	DB *gorm.DB
}

type transactionDB struct {
	Id           string    `gorm:"type:varchar(26);primaryKey"`
	UserId       string    `gorm:"type:varchar(26);index;not null"`
	AccountId    string    `gorm:"type:varchar(26);index;not null"`
	Type         string    `gorm:"type:varchar(15);not null"`
	CategoryId   string    `gorm:"type:varchar(26);index"`
	InvestmentId *string   `gorm:"type:varchar(26);index"`
	Amount       float64   `gorm:"not null"`
	Description  string    `gorm:"size:255"`
	Date         time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func toDomainTransaction(tdb *transactionDB) (*transaction.Transaction, error) {
	id, err := pkg.ParseULID(tdb.Id)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(tdb.UserId)
	if err != nil {
		return nil, err
	}
	aid, err := pkg.ParseULID(tdb.AccountId)
	if err != nil {
		return nil, err
	}
	cid, err := pkg.ParseULID(tdb.CategoryId)
	if err != nil {
		return nil, err
	}

	var invID *ulid.ULID
	if tdb.InvestmentId != nil && *tdb.InvestmentId != "" {
		parsed, err := pkg.ParseULID(*tdb.InvestmentId)
		if err != nil {
			return nil, err
		}
		invID = &parsed
	}

	return &transaction.Transaction{
		Id:           id,
		UserId:       uid,
		AccountId:    aid,
		Type:         transaction.Types(tdb.Type),
		CategoryId:   cid,
		InvestmentId: invID,
		Amount:       tdb.Amount,
		Description:  tdb.Description,
		Date:         tdb.Date,
		CreatedAt:    tdb.CreatedAt,
		UpdatedAt:    tdb.UpdatedAt,
	}, nil
}

func toDBTransaction(t *transaction.Transaction) *transactionDB {
	var invID *string
	if t.InvestmentId != nil {
		s := t.InvestmentId.String()
		invID = &s
	}
	return &transactionDB{
		Id:           t.Id.String(),
		UserId:       t.UserId.String(),
		AccountId:    t.AccountId.String(),
		Type:         string(t.Type),
		CategoryId:   t.CategoryId.String(),
		InvestmentId: invID,
		Amount:       t.Amount,
		Description:  t.Description,
		Date:         t.Date,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, t *transaction.Transaction) error {
	tdb := toDBTransaction(t)
	return r.DB.WithContext(ctx).Table("transactions").Create(tdb).Error
}

func (r *TransactionRepository) Update(ctx context.Context, t *transaction.Transaction) error {
	tdb := toDBTransaction(t)
	return r.DB.WithContext(ctx).Table("transactions").Where("id = ?", tdb.Id).Updates(tdb).Error
}

func (r *TransactionRepository) Delete(ctx context.Context, transactionID ulid.ULID) error {
	return r.DB.WithContext(ctx).Table("transactions").Where("id = ?", transactionID.String()).Delete(&transactionDB{}).Error
}

func (r *TransactionRepository) GetByID(ctx context.Context, transactionID ulid.ULID) (*transaction.Transaction, error) {
	var tdb transactionDB
	err := r.DB.WithContext(ctx).Table("transactions").Where("id = ?", transactionID.String()).First(&tdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainTransaction(&tdb)
}

func (r *TransactionRepository) GetByIDAndUser(ctx context.Context, transactionID, userID ulid.ULID) (*transaction.Transaction, error) {
	var tdb transactionDB
	err := r.DB.WithContext(ctx).Table("transactions").Where("id = ? AND user_id = ?", transactionID.String(), userID.String()).First(&tdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainTransaction(&tdb)
}

func (r *TransactionRepository) GetAll(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("user_id = ?", userID.String())
	if accountID != nil {
		baseQuery = baseQuery.Where("account_id = ?", accountID.String())
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	query := baseQuery.Order("date DESC, created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit)
	
	err := query.Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainTransaction(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, nil
}

func (r *TransactionRepository) GetByAmount(ctx context.Context, amount float64, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("amount = ?", amount)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := baseQuery.Order("date DESC, created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainTransaction(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, nil
}

func (r *TransactionRepository) GetByName(ctx context.Context, name string, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("description LIKE ?", "%"+name+"%")

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := baseQuery.Order("date DESC, created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainTransaction(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, nil
}

func (r *TransactionRepository) GetByCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("user_id = ? AND category_id = ?", userID.String(), categoryID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := baseQuery.Order("date DESC, created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainTransaction(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, nil
}

func (r *TransactionRepository) GetNumberOfTransactions(ctx context.Context, userID ulid.ULID) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&transaction.Transaction{}).Where("user_id = ?", userID.String()).Count(&count).Error
	return count, err
}

func (r *TransactionRepository) GetByInvestmentId(ctx context.Context, investmentID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("investment_id = ? AND user_id = ?", investmentID.String(), userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := baseQuery.Order("date DESC, created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		t, err := toDomainTransaction(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, nil
}
