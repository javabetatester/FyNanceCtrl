package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/budget"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type BudgetRepository struct {
	DB *gorm.DB
}

type budgetDB struct {
	Id          string    `gorm:"type:varchar(26);primaryKey"`
	UserId      string    `gorm:"type:varchar(26);index;not null"`
	CategoryId  string    `gorm:"type:varchar(26);index;not null"`
	Amount      float64   `gorm:"type:decimal(15,2);not null"`
	Spent       float64   `gorm:"type:decimal(15,2);not null;default:0"`
	Month       int       `gorm:"not null"`
	Year        int       `gorm:"not null"`
	AlertAt     float64   `gorm:"type:decimal(5,2);default:80"`
	IsRecurring bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (budgetDB) TableName() string {
	return "budgets"
}

func toDomainBudget(bdb *budgetDB) (*budget.Budget, error) {
	id, err := pkg.ParseULID(bdb.Id)
	if err != nil {
		return nil, err
	}

	userID, err := pkg.ParseULID(bdb.UserId)
	if err != nil {
		return nil, err
	}

	categoryID, err := pkg.ParseULID(bdb.CategoryId)
	if err != nil {
		return nil, err
	}

	return &budget.Budget{
		Id:          id,
		UserId:      userID,
		CategoryId:  categoryID,
		Amount:      bdb.Amount,
		Spent:       bdb.Spent,
		Month:       bdb.Month,
		Year:        bdb.Year,
		AlertAt:     bdb.AlertAt,
		IsRecurring: bdb.IsRecurring,
		CreatedAt:   bdb.CreatedAt,
		UpdatedAt:   bdb.UpdatedAt,
	}, nil
}

func toDBBudget(b *budget.Budget) *budgetDB {
	return &budgetDB{
		Id:          b.Id.String(),
		UserId:      b.UserId.String(),
		CategoryId:  b.CategoryId.String(),
		Amount:      b.Amount,
		Spent:       b.Spent,
		Month:       b.Month,
		Year:        b.Year,
		AlertAt:     b.AlertAt,
		IsRecurring: b.IsRecurring,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func (r *BudgetRepository) Create(ctx context.Context, b *budget.Budget) error {
	bdb := toDBBudget(b)
	return r.DB.WithContext(ctx).Table("budgets").Create(bdb).Error
}

func (r *BudgetRepository) Update(ctx context.Context, b *budget.Budget) error {
	bdb := toDBBudget(b)
	return r.DB.WithContext(ctx).Model(&budgetDB{}).Where("id = ? AND user_id = ?", bdb.Id, bdb.UserId).Updates(bdb).Error
}

func (r *BudgetRepository) Delete(ctx context.Context, budgetID, userID ulid.ULID) error {
	return r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", budgetID.String(), userID.String()).Delete(&budgetDB{}).Error
}

func (r *BudgetRepository) GetById(ctx context.Context, budgetID, userID ulid.ULID) (*budget.Budget, error) {
	var bdb budgetDB
	err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", budgetID.String(), userID.String()).First(&bdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainBudget(&bdb)
}

func (r *BudgetRepository) GetByUserId(ctx context.Context, userID ulid.ULID, month, year int, pagination *pkg.PaginationParams) ([]*budget.Budget, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("budgets").Where("user_id = ? AND month = ? AND year = ?", userID.String(), month, year)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []budgetDB
	err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	budgets := make([]*budget.Budget, 0, len(rows))
	for i := range rows {
		b, err := toDomainBudget(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		budgets = append(budgets, b)
	}
	return budgets, total, nil
}

func (r *BudgetRepository) GetByCategoryId(ctx context.Context, categoryID, userID ulid.ULID, month, year int) (*budget.Budget, error) {
	var bdb budgetDB
	err := r.DB.WithContext(ctx).Where("category_id = ? AND user_id = ? AND month = ? AND year = ?", categoryID.String(), userID.String(), month, year).First(&bdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainBudget(&bdb)
}

func (r *BudgetRepository) UpdateSpent(ctx context.Context, budgetID ulid.ULID, amount float64) error {
	return r.DB.WithContext(ctx).Model(&budgetDB{}).Where("id = ?", budgetID.String()).
		UpdateColumn("spent", gorm.Expr("spent + ?", amount)).
		UpdateColumn("updated_at", time.Now()).Error
}

func (r *BudgetRepository) GetRecurring(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*budget.Budget, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("budgets").Where("user_id = ? AND is_recurring = ?", userID.String(), true)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []budgetDB
	err := baseQuery.Group("category_id").
		Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	budgets := make([]*budget.Budget, 0, len(rows))
	for i := range rows {
		b, err := toDomainBudget(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		budgets = append(budgets, b)
	}
	return budgets, total, nil
}

func (r *BudgetRepository) GetSummary(ctx context.Context, userID ulid.ULID, month, year int) (*budget.BudgetSummary, error) {
	var result struct {
		TotalBudget float64
		TotalSpent  float64
	}

	err := r.DB.WithContext(ctx).Model(&budgetDB{}).
		Where("user_id = ? AND month = ? AND year = ?", userID.String(), month, year).
		Select("COALESCE(SUM(amount), 0) as total_budget, COALESCE(SUM(spent), 0) as total_spent").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	remaining := result.TotalBudget - result.TotalSpent
	percentage := 0.0
	if result.TotalBudget > 0 {
		percentage = (result.TotalSpent / result.TotalBudget) * 100
	}

	return &budget.BudgetSummary{
		TotalBudget:    result.TotalBudget,
		TotalSpent:     result.TotalSpent,
		TotalRemaining: remaining,
		Percentage:     percentage,
	}, nil
}
