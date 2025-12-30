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

var _ budget.BudgetRepository = (*BudgetRepository)(nil)

type budgetDB struct {
	Id          string    `gorm:"type:varchar(26);primaryKey;column:id"`
	UserId      string    `gorm:"type:varchar(26);index;not null;column:user_id"`
	CategoryId  string    `gorm:"type:varchar(26);index;not null;column:category_id"`
	Amount      float64   `gorm:"type:decimal(15,2);not null;column:amount"`
	Spent       float64   `gorm:"type:decimal(15,2);not null;default:0;column:spent"`
	Month       int       `gorm:"type:integer;not null;column:month"`
	Year        int       `gorm:"type:integer;not null;column:year"`
	AlertAt     float64   `gorm:"type:decimal(5,2);default:80;column:alert_at"`
	IsRecurring bool      `gorm:"not null;default:false;column:is_recurring"`
	CreatedAt   time.Time `gorm:"not null;column:created_at"`
	UpdatedAt   time.Time `gorm:"not null;column:updated_at"`
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
	result := r.DB.WithContext(ctx).Table("budgets").Create(&bdb)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *BudgetRepository) Update(ctx context.Context, b *budget.Budget) error {
	bdb := toDBBudget(b)
	return r.DB.WithContext(ctx).Model(&budgetDB{}).Where("id = ? AND user_id = ?", bdb.Id, bdb.UserId).Updates(bdb).Error
}

func (r *BudgetRepository) Delete(ctx context.Context, budgetID, userID ulid.ULID) error {
	result := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", budgetID.String(), userID.String()).Delete(&budgetDB{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, budgetID, userID ulid.ULID) (*budget.Budget, error) {
	type budgetDBWithCategory struct {
		budgetDB
		CategoryName string `gorm:"->;column:category_name"`
	}
	var bdb budgetDBWithCategory
	query := r.DB.WithContext(ctx).
		Table("budgets b").
		Select("b.id, b.user_id, b.category_id, b.amount, b.spent, b.month, b.year, b.alert_at, b.is_recurring, b.created_at, b.updated_at, c.name as category_name").
		Joins("LEFT JOIN categories c ON b.category_id = c.id").
		Where("b.id = ? AND b.user_id = ?", budgetID.String(), userID.String()).
		Limit(1)
	err := query.Take(&bdb).Error
	if err != nil {
		return nil, err
	}
	if bdb.budgetDB.Id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	b, err := toDomainBudget(&bdb.budgetDB)
	if err != nil {
		return nil, err
	}
	if bdb.CategoryName != "" {
		b.CategoryName = bdb.CategoryName
	}
	return b, nil
}

func (r *BudgetRepository) GetByUserID(ctx context.Context, userID ulid.ULID, month, year int, filters *budget.BudgetFilters, pagination *pkg.PaginationParams) ([]*budget.Budget, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	type budgetDBWithCategory struct {
		Id           string    `gorm:"column:id"`
		UserId       string    `gorm:"column:user_id"`
		CategoryId   string    `gorm:"column:category_id"`
		Amount       float64   `gorm:"column:amount"`
		Spent        float64   `gorm:"column:spent"`
		Month        int       `gorm:"column:month"`
		Year         int       `gorm:"column:year"`
		AlertAt      float64   `gorm:"column:alert_at"`
		IsRecurring  bool      `gorm:"column:is_recurring"`
		CreatedAt    time.Time `gorm:"column:created_at"`
		UpdatedAt    time.Time `gorm:"column:updated_at"`
		CategoryName string    `gorm:"column:category_name"`
	}

	baseQuery := r.DB.WithContext(ctx).
		Table("budgets b").
		Select("b.id, b.user_id, b.category_id, b.amount, b.spent, b.month, b.year, b.alert_at, b.is_recurring, b.created_at, b.updated_at, c.name as category_name").
		Joins("LEFT JOIN categories c ON b.category_id = c.id AND c.user_id = b.user_id").
		Where("b.user_id = ?", userID.String())

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		baseQuery = baseQuery.Where("c.name ILIKE ?", searchPattern)
	}

	countQuery := r.DB.WithContext(ctx).
		Table("budgets").
		Where("user_id = ?", userID.String())

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN categories c ON budgets.category_id = c.id").
			Where("c.name ILIKE ?", searchPattern)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []budgetDBWithCategory
	err := baseQuery.Order("b.created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	budgets := make([]*budget.Budget, 0, len(rows))
	for i := range rows {
		if rows[i].Id == "" || rows[i].UserId == "" {
			continue
		}

		bdb := &budgetDB{
			Id:          rows[i].Id,
			UserId:      rows[i].UserId,
			CategoryId:  rows[i].CategoryId,
			Amount:      rows[i].Amount,
			Spent:       rows[i].Spent,
			Month:       rows[i].Month,
			Year:        rows[i].Year,
			AlertAt:     rows[i].AlertAt,
			IsRecurring: rows[i].IsRecurring,
			CreatedAt:   rows[i].CreatedAt,
			UpdatedAt:   rows[i].UpdatedAt,
		}

		b, err := toDomainBudget(bdb)
		if err != nil {
			continue
		}

		if rows[i].CategoryName != "" {
			b.CategoryName = rows[i].CategoryName
		}

		budgets = append(budgets, b)
	}
	return budgets, total, nil
}

func (r *BudgetRepository) GetByCategoryID(ctx context.Context, categoryID, userID ulid.ULID, month, year int) (*budget.Budget, error) {
	var bdb budgetDB
	err := r.DB.WithContext(ctx).
		Table("budgets").
		Where("category_id = ? AND user_id = ? AND month = ? AND year = ?", categoryID.String(), userID.String(), month, year).
		First(&bdb).Error
	if err != nil {
		return nil, err
	}
	b, err := toDomainBudget(&bdb)
	if err != nil {
		return nil, err
	}
	return b, nil
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

	countQuery := r.DB.WithContext(ctx).Table("budgets").Where("user_id = ? AND is_recurring = ?", userID.String(), true)
	dataQuery := r.DB.WithContext(ctx).
		Table("budgets b").
		Select("b.id, b.user_id, b.category_id, b.amount, b.spent, b.month, b.year, b.alert_at, b.is_recurring, b.created_at, b.updated_at, c.name as category_name").
		Joins("LEFT JOIN categories c ON b.category_id = c.id").
		Where("b.user_id = ? AND b.is_recurring = ?", userID.String(), true)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type budgetDBWithCategory struct {
		budgetDB
		CategoryName string `gorm:"->;column:category_name"`
	}
	var rows []budgetDBWithCategory
	err := dataQuery.Group("b.category_id, b.id, c.name").
		Order("b.created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	budgets := make([]*budget.Budget, 0, len(rows))
	for i := range rows {
		if rows[i].budgetDB.Id == "" || rows[i].budgetDB.UserId == "" || rows[i].budgetDB.CategoryId == "" {
			continue
		}
		b, err := toDomainBudget(&rows[i].budgetDB)
		if err != nil {
			continue
		}
		if rows[i].CategoryName != "" {
			b.CategoryName = rows[i].CategoryName
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
