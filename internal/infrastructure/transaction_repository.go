package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/transaction"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	DB *gorm.DB
}

var _ transaction.TransactionRepository = (*TransactionRepository)(nil)

type transactionDB struct {
	Id           string    `gorm:"type:varchar(26);primaryKey;column:id"`
	UserId       string    `gorm:"type:varchar(26);index;not null;column:user_id"`
	AccountId    string    `gorm:"type:varchar(26);index;not null;column:account_id"`
	Type         string    `gorm:"type:varchar(15);not null;column:type"`
	CategoryId   *string   `gorm:"type:varchar(26);index;column:category_id"`
	CategoryName string    `gorm:"->;column:category_name"`
	InvestmentId *string   `gorm:"type:varchar(26);index;column:investment_id"`
	Amount       float64   `gorm:"not null;column:amount"`
	Description  string    `gorm:"size:255;column:description"`
	Date         time.Time `gorm:"not null;column:date"`
	CreatedAt    time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    time.Time `gorm:"not null;column:updated_at"`
}

func toDomainTransaction(tdb *transactionDB) (*transaction.Transaction, error) {
	return toDomainTransactionWithCategory(tdb)
}

func toDomainTransactionWithCategory(tdb *transactionDB) (*transaction.Transaction, error) {
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
	var cid *ulid.ULID
	if tdb.CategoryId != nil && *tdb.CategoryId != "" {
		parsed, err := pkg.ParseULID(*tdb.CategoryId)
		if err != nil {
			return nil, err
		}
		cid = &parsed
	}

	var invID *ulid.ULID
	if tdb.InvestmentId != nil && *tdb.InvestmentId != "" {
		parsed, err := pkg.ParseULID(*tdb.InvestmentId)
		if err != nil {
			return nil, err
		}
		invID = &parsed
	}

	tx := &transaction.Transaction{
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
	}

	if tdb.CategoryName != "" {
		tx.CategoryName = tdb.CategoryName
	}

	return tx, nil
}

func toDBTransaction(t *transaction.Transaction) *transactionDB {
	var invID *string
	if t.InvestmentId != nil {
		s := t.InvestmentId.String()
		invID = &s
	}
	var categoryID *string
	if t.CategoryId != nil {
		s := t.CategoryId.String()
		categoryID = &s
	}
	return &transactionDB{
		Id:           t.Id.String(),
		UserId:       t.UserId.String(),
		AccountId:    t.AccountId.String(),
		Type:         string(t.Type),
		CategoryId:   categoryID,
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
	err := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.id = ?", transactionID.String()).
		First(&tdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainTransaction(&tdb)
}

func (r *TransactionRepository) GetByIDAndUser(ctx context.Context, transactionID, userID ulid.ULID) (*transaction.Transaction, error) {
	var tdb transactionDB
	err := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.id = ? AND t.user_id = ?", transactionID.String(), userID.String()).
		First(&tdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainTransaction(&tdb)
}

func (r *TransactionRepository) GetAll(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, filters *transaction.TransactionFilters, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	countQuery := r.DB.WithContext(ctx).Table("transactions t").Where("t.user_id = ?", userID.String())
	dataQuery := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ?", userID.String())

	if accountID != nil {
		countQuery = countQuery.Where("t.account_id = ?", accountID.String())
		dataQuery = dataQuery.Where("t.account_id = ?", accountID.String())
	}

	if filters != nil {
		if filters.Type != nil && *filters.Type != "" && *filters.Type != "ALL" {
			countQuery = countQuery.Where("t.type = ?", *filters.Type)
			dataQuery = dataQuery.Where("t.type = ?", *filters.Type)
		}

		if filters.CategoryID != nil {
			countQuery = countQuery.Where("t.category_id = ?", filters.CategoryID.String())
			dataQuery = dataQuery.Where("t.category_id = ?", filters.CategoryID.String())
		}

		if filters.Search != nil && *filters.Search != "" {
			searchPattern := "%" + *filters.Search + "%"
			countQuery = countQuery.Where("t.description ILIKE ?", searchPattern)
			dataQuery = dataQuery.Where("t.description ILIKE ?", searchPattern)
		}

		if filters.DateFrom != nil {
			countQuery = countQuery.Where("t.date >= ?", *filters.DateFrom)
			dataQuery = dataQuery.Where("t.date >= ?", *filters.DateFrom)
		}

		if filters.DateTo != nil {
			countQuery = countQuery.Where("t.date <= ?", *filters.DateTo)
			dataQuery = dataQuery.Where("t.date <= ?", *filters.DateTo)
		}
	}

	pagination = pkg.NormalizePagination(pagination)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := dataQuery.Order("t.date DESC, t.created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		item, err := toDomainTransaction(&rows[i])
		if err != nil {
			continue
		}
		out = append(out, item)
	}

	return out, total, nil
}

func (r *TransactionRepository) GetByAmount(ctx context.Context, amount float64, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("amount = ?", amount)
	return pkg.Paginate(baseQuery, pagination, "date DESC, created_at DESC", toDomainTransaction)
}

func (r *TransactionRepository) GetByName(ctx context.Context, name string, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Table("transactions").Where("description LIKE ?", "%"+name+"%")
	return pkg.Paginate(baseQuery, pagination, "date DESC, created_at DESC", toDomainTransaction)
}

func (r *TransactionRepository) GetByCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	countQuery := r.DB.WithContext(ctx).Table("transactions t").Where("t.user_id = ? AND t.category_id = ?", userID.String(), categoryID.String())
	dataQuery := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ? AND t.category_id = ?", userID.String(), categoryID.String())

	pagination = pkg.NormalizePagination(pagination)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := dataQuery.Order("t.date DESC, t.created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		item, err := toDomainTransaction(&rows[i])
		if err != nil {
			continue
		}
		out = append(out, item)
	}

	return out, total, nil
}

func (r *TransactionRepository) GetByInvestmentID(ctx context.Context, investmentID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	countQuery := r.DB.WithContext(ctx).Table("transactions t").Where("t.investment_id = ? AND t.user_id = ?", investmentID.String(), userID.String())
	dataQuery := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.investment_id = ? AND t.user_id = ?", investmentID.String(), userID.String())

	pagination = pkg.NormalizePagination(pagination)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []transactionDB
	err := dataQuery.Order("t.date DESC, t.created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]*transaction.Transaction, 0, len(rows))
	for i := range rows {
		item, err := toDomainTransaction(&rows[i])
		if err != nil {
			continue
		}
		out = append(out, item)
	}

	return out, total, nil
}

func (r *TransactionRepository) GetNumberOfTransactions(ctx context.Context, userID ulid.ULID) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&transaction.Transaction{}).Where("user_id = ?", userID.String()).Count(&count).Error
	return count, err
}
