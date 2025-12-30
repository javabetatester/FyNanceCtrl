package infrastructure

import (
	"Fynance/internal/domain/transaction"
	"Fynance/internal/pkg"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type TransactionCategoryRepository struct {
	DB *gorm.DB
}

var _ transaction.CategoryRepository = (*TransactionCategoryRepository)(nil)

type categoryDB struct {
	UserId    string    `gorm:"type:varchar(26);index;not null"`
	Id        string    `gorm:"type:varchar(26);primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	Icon      string    `gorm:"size:50"`
	CreatedAt time.Time `gorm:"type:timestamp;"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func toDomainCategory(cdb *categoryDB) (*transaction.Category, error) {
	uid, err := pkg.ParseULID(cdb.UserId)
	if err != nil {
		return nil, err
	}
	id, err := pkg.ParseULID(cdb.Id)
	if err != nil {
		return nil, err
	}
	return &transaction.Category{
		UserId:    uid,
		Id:        id,
		Name:      cdb.Name,
		Icon:      cdb.Icon,
		CreatedAt: cdb.CreatedAt,
		UpdatedAt: cdb.UpdatedAt,
	}, nil
}

func toDBCategory(c *transaction.Category) *categoryDB {
	return &categoryDB{
		UserId:    c.UserId.String(),
		Id:        c.Id.String(),
		Name:      c.Name,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (r *TransactionCategoryRepository) Create(ctx context.Context, category *transaction.Category) error {
	cdb := toDBCategory(category)
	return r.DB.WithContext(ctx).Table("categories").Create(&cdb).Error
}

func (r *TransactionCategoryRepository) Update(ctx context.Context, category *transaction.Category) error {
	cdb := toDBCategory(category)
	return r.DB.WithContext(ctx).Table("categories").Where("id = ?", cdb.Id).Updates(&cdb).Error
}

func (r *TransactionCategoryRepository) Delete(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error {
	return r.DB.WithContext(ctx).Table("categories").Where("id = ? AND user_id = ?", categoryID.String(), userID.String()).Delete(&categoryDB{}).Error
}

func (r *TransactionCategoryRepository) GetByID(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (*transaction.Category, error) {
	var row categoryDB
	err := r.DB.WithContext(ctx).Table("categories").Where("id = ? AND user_id = ?", categoryID.String(), userID.String()).First(&row).Error
	if err != nil {
		return nil, err
	}
	return toDomainCategory(&row)
}

func (r *TransactionCategoryRepository) GetAll(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Category, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("categories").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []categoryDB
	err := baseQuery.Order("name ASC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Category, 0, len(rows))
	for i := range rows {
		c, err := toDomainCategory(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (r *TransactionCategoryRepository) GetByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Category, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("categories").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []categoryDB
	err := baseQuery.Order("name ASC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]*transaction.Category, 0, len(rows))
	for i := range rows {
		c, err := toDomainCategory(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (r *TransactionCategoryRepository) GetByName(ctx context.Context, CategoryName string, userID ulid.ULID) (*transaction.Category, error) {
	var row categoryDB
	searchName := strings.TrimSpace(CategoryName)
	searchLower := strings.ToLower(searchName)

	err := r.DB.WithContext(ctx).Table("categories").
		Where("user_id = ? AND (LOWER(TRIM(name)) = ? OR name = ?)", userID.String(), searchLower, searchName).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var allRows []categoryDB
			err2 := r.DB.WithContext(ctx).Table("categories").
				Where("user_id = ?", userID.String()).
				Find(&allRows).Error
			if err2 != nil {
				return nil, err
			}

			for _, r := range allRows {
				if strings.ToLower(strings.TrimSpace(r.Name)) == searchLower {
					return toDomainCategory(&r)
				}
			}
		}
		return nil, err
	}
	return toDomainCategory(&row)
}

func (r *TransactionCategoryRepository) GetAllWithoutLimit(ctx context.Context, userID ulid.ULID) ([]*transaction.Category, error) {
	var rows []categoryDB
	err := r.DB.WithContext(ctx).Table("categories").
		Where("user_id = ?", userID.String()).
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*transaction.Category, 0, len(rows))
	for i := range rows {
		c, err := toDomainCategory(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *TransactionCategoryRepository) BelongsToUser(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Table("categories").Where("id = ? AND user_id = ?", categoryID.String(), userID.String()).Count(&count).Error
	return count > 0, err
}
