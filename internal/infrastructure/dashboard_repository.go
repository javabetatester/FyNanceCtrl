package infrastructure

import (
	"context"
	"strings"
	"time"

	"Fynance/internal/domain/dashboard"
	"Fynance/internal/domain/transaction"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	DB *gorm.DB
}

var _ dashboard.Repository = (*DashboardRepository)(nil)

func (r *DashboardRepository) GetFinancialSummary(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) (*dashboard.FinancialSummary, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	incomeQuery := r.DB.WithContext(ctx).Table("transactions").
		Where("user_id = ? AND type = ? AND date >= ? AND date < ?", userID.String(), "RECEIPT", startDate, endDate)
	if accountID != nil {
		incomeQuery = incomeQuery.Where("account_id = ?", accountID.String())
	}

	var monthIncome float64
	if err := incomeQuery.Select("COALESCE(SUM(amount), 0)").Scan(&monthIncome).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	expenseQuery := r.DB.WithContext(ctx).Table("transactions").
		Where("user_id = ? AND type = ? AND date >= ? AND date < ?", userID.String(), "EXPENSE", startDate, endDate)
	if accountID != nil {
		expenseQuery = expenseQuery.Where("account_id = ?", accountID.String())
	}

	var monthExpenses float64
	if err := expenseQuery.Select("COALESCE(SUM(amount), 0)").Scan(&monthExpenses).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	var totalBalance float64
	balanceQuery := r.DB.WithContext(ctx).Table("accounts").
		Where("user_id = ? AND include_in_total = ? AND is_active = ?", userID.String(), true, true)
	if accountID != nil {
		balanceQuery = balanceQuery.Where("id = ?", accountID.String())
	}
	if err := balanceQuery.Select("COALESCE(SUM(balance), 0)").Scan(&totalBalance).Error; err != nil {
		totalBalance = 0
	}

	var totalInvestments float64
	if err := r.DB.WithContext(ctx).Table("investments").
		Where("user_id = ?", userID.String()).
		Select("COALESCE(SUM(current_balance), 0)").
		Scan(&totalInvestments).Error; err != nil {
		totalInvestments = 0
	}

	var totalGoals float64
	if err := r.DB.WithContext(ctx).Table("goals").
		Where("user_id = ? AND status = ?", userID.String(), "ACTIVE").
		Select("COALESCE(SUM(current_amount), 0)").
		Scan(&totalGoals).Error; err != nil {
		totalGoals = 0
	}

	return &dashboard.FinancialSummary{
		TotalBalance:     totalBalance,
		MonthIncome:      monthIncome,
		MonthExpenses:    monthExpenses,
		MonthBalance:     monthIncome - monthExpenses,
		TotalInvestments: totalInvestments,
		TotalGoals:       totalGoals,
	}, nil
}

func (r *DashboardRepository) GetMonthlyTrend(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, months int) ([]*dashboard.MonthlyTrendItem, error) {
	now := time.Now()
	items := make([]*dashboard.MonthlyTrendItem, 0, months)

	for i := months - 1; i >= 0; i-- {
		targetDate := now.AddDate(0, -i, 0)
		startDate := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)

		incomeQuery := r.DB.WithContext(ctx).Table("transactions").
			Where("user_id = ? AND type = ? AND date >= ? AND date < ?", userID.String(), "RECEIPT", startDate, endDate)
		if accountID != nil {
			incomeQuery = incomeQuery.Where("account_id = ?", accountID.String())
		}

		var income float64
		if err := incomeQuery.Select("COALESCE(SUM(amount), 0)").Scan(&income).Error; err != nil {
			return nil, appErrors.NewDatabaseError(err)
		}

		expenseQuery := r.DB.WithContext(ctx).Table("transactions").
			Where("user_id = ? AND type = ? AND date >= ? AND date < ?", userID.String(), "EXPENSE", startDate, endDate)
		if accountID != nil {
			expenseQuery = expenseQuery.Where("account_id = ?", accountID.String())
		}

		var expenses float64
		if err := expenseQuery.Select("COALESCE(SUM(amount), 0)").Scan(&expenses).Error; err != nil {
			return nil, appErrors.NewDatabaseError(err)
		}

		items = append(items, &dashboard.MonthlyTrendItem{
			Month:    startDate.Month().String(),
			Year:     startDate.Year(),
			Income:   income,
			Expenses: expenses,
			Balance:  income - expenses,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetExpensesByCategory(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) ([]*dashboard.CategoryExpense, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	type categoryResult struct {
		CategoryId *string `gorm:"column:category_id"`
		Name       *string `gorm:"column:name"`
		Amount     float64 `gorm:"column:amount"`
	}

	query := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.category_id, c.name, SUM(ABS(t.amount)) as amount").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ? AND t.type = ? AND t.date >= ? AND t.date < ?", userID.String(), "EXPENSE", startDate, endDate)
	if accountID != nil {
		query = query.Where("t.account_id = ?", accountID.String())
	}

	var results []categoryResult
	if err := query.Group("t.category_id, c.name").
		Having("SUM(ABS(t.amount)) > 0").
		Order("amount DESC").
		Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	if len(results) == 0 {
		return []*dashboard.CategoryExpense{}, nil
	}

	var total float64
	for _, r := range results {
		total += r.Amount
	}

	items := make([]*dashboard.CategoryExpense, 0, len(results))
	for _, r := range results {
		if r.Amount <= 0 {
			continue
		}
		var categoryID ulid.ULID
		if r.CategoryId != nil && *r.CategoryId != "" {
			categoryIdStr := strings.TrimSpace(*r.CategoryId)
			if categoryIdStr != "" && categoryIdStr != "NULL" {
				parsed, err := pkg.ParseULID(categoryIdStr)
				if err == nil {
					categoryID = parsed
				}
			}
		}
		categoryName := "Sem categoria"
		if r.Name != nil && *r.Name != "" {
			nameStr := strings.TrimSpace(*r.Name)
			if nameStr != "" && nameStr != "NULL" {
				categoryName = nameStr
			}
		}
		percentage := 0.0
		if total > 0 {
			percentage = (r.Amount / total) * 100
		}
		items = append(items, &dashboard.CategoryExpense{
			CategoryId:   categoryID,
			CategoryName: categoryName,
			Amount:       r.Amount,
			Percentage:   percentage,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetRecentTransactions(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, limit int) ([]*dashboard.TransactionSummary, error) {
	type transactionResult struct {
		Id           string    `gorm:"column:id"`
		Type         string    `gorm:"column:type"`
		Amount       float64   `gorm:"column:amount"`
		Description  string    `gorm:"column:description"`
		CategoryId   string    `gorm:"column:category_id"`
		CategoryName string    `gorm:"column:category_name"`
		Date         time.Time `gorm:"column:date"`
		AccountId    string    `gorm:"column:account_id"`
	}

	query := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.id, t.type, t.amount, t.description, t.category_id, c.name as category_name, t.date, t.account_id").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ?", userID.String())
	if accountID != nil {
		query = query.Where("t.account_id = ?", accountID.String())
	}

	var results []transactionResult
	if err := query.Order("t.date DESC, t.created_at DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*dashboard.TransactionSummary, 0, len(results))
	for _, r := range results {
		id, err := pkg.ParseULID(r.Id)
		if err != nil {
			continue
		}
		categoryID, err := pkg.ParseULID(r.CategoryId)
		if err != nil {
			categoryID = ulid.ULID{}
		}
		accountID, err := pkg.ParseULID(r.AccountId)
		if err != nil {
			continue
		}
		items = append(items, &dashboard.TransactionSummary{
			Id:           id,
			Type:         r.Type,
			Amount:       r.Amount,
			Description:  r.Description,
			CategoryId:   categoryID,
			CategoryName: r.CategoryName,
			Date:         r.Date,
			AccountId:    accountID,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetActiveGoals(ctx context.Context, userID ulid.ULID) ([]*dashboard.GoalSummary, error) {
	type goalResult struct {
		Id            string  `gorm:"column:id"`
		Name          string  `gorm:"column:name"`
		TargetAmount  float64 `gorm:"column:target_amount"`
		CurrentAmount float64 `gorm:"column:current_amount"`
		Status        string  `gorm:"column:status"`
	}

	var results []goalResult
	if err := r.DB.WithContext(ctx).Table("goals").
		Select("id, name, target_amount, current_amount, status").
		Where("user_id = ? AND status = ?", userID.String(), "ACTIVE").
		Order("created_at DESC").
		Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*dashboard.GoalSummary, 0, len(results))
	for _, r := range results {
		id, err := pkg.ParseULID(r.Id)
		if err != nil {
			continue
		}
		percentage := 0.0
		if r.TargetAmount > 0 {
			percentage = (r.CurrentAmount / r.TargetAmount) * 100
		}
		items = append(items, &dashboard.GoalSummary{
			Id:            id,
			Name:          r.Name,
			TargetAmount:  r.TargetAmount,
			CurrentAmount: r.CurrentAmount,
			Percentage:    percentage,
			Status:        r.Status,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetBudgetStatus(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) ([]*dashboard.BudgetStatusItem, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	type budgetResult struct {
		CategoryId string  `gorm:"column:category_id"`
		Name       string  `gorm:"column:name"`
		Amount     float64 `gorm:"column:amount"`
	}

	var budgets []budgetResult
	if err := r.DB.WithContext(ctx).Table("budgets b").
		Select("b.category_id, c.name, b.amount").
		Joins("LEFT JOIN categories c ON b.category_id = c.id").
		Where("b.user_id = ? AND b.month = ? AND b.year = ?", userID.String(), month, year).
		Scan(&budgets).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*dashboard.BudgetStatusItem, 0, len(budgets))
	for _, b := range budgets {
		categoryID, err := pkg.ParseULID(b.CategoryId)
		if err != nil {
			continue
		}
		spentQuery := r.DB.WithContext(ctx).Table("transactions").
			Where("user_id = ? AND category_id = ? AND type = ? AND date >= ? AND date < ?",
				userID.String(), b.CategoryId, "EXPENSE", startDate, endDate)
		if accountID != nil {
			spentQuery = spentQuery.Where("account_id = ?", accountID.String())
		}
		var spent float64
		if err := spentQuery.Select("COALESCE(SUM(amount), 0)").Scan(&spent).Error; err != nil {
			spent = 0
		}

		percentage := 0.0
		if b.Amount > 0 {
			percentage = (spent / b.Amount) * 100
		}

		status := "OK"
		if percentage >= 100 {
			status = "OVER"
		} else if percentage >= 80 {
			status = "WARNING"
		}

		items = append(items, &dashboard.BudgetStatusItem{
			CategoryId:   categoryID,
			CategoryName: b.Name,
			BudgetAmount: b.Amount,
			SpentAmount:  spent,
			Remaining:    b.Amount - spent,
			Percentage:   percentage,
			Status:       status,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetAccountsSummary(ctx context.Context, userID ulid.ULID) ([]*dashboard.AccountSummary, error) {
	type accountResult struct {
		Id      string  `gorm:"column:id"`
		Name    string  `gorm:"column:name"`
		Type    string  `gorm:"column:type"`
		Balance float64 `gorm:"column:balance"`
		Color   string  `gorm:"column:color"`
	}

	var results []accountResult
	if err := r.DB.WithContext(ctx).Table("accounts").
		Select("id, name, type, balance, color").
		Where("user_id = ? AND is_active = ? AND type != ?", userID.String(), true, "CREDIT_CARD").
		Order("name ASC").
		Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*dashboard.AccountSummary, 0, len(results))
	for _, r := range results {
		id, err := pkg.ParseULID(r.Id)
		if err != nil {
			continue
		}
		items = append(items, &dashboard.AccountSummary{
			Id:      id,
			Name:    r.Name,
			Type:    r.Type,
			Balance: r.Balance,
			Color:   r.Color,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetUserCategories(ctx context.Context, userID ulid.ULID) ([]*transaction.Category, error) {
	type categoryResult struct {
		Id        string    `gorm:"column:id"`
		UserId    string    `gorm:"column:user_id"`
		Name      string    `gorm:"column:name"`
		Icon      string    `gorm:"column:icon"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	var results []categoryResult
	if err := r.DB.WithContext(ctx).Table("categories").
		Where("user_id = ?", userID.String()).
		Order("name ASC").
		Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*transaction.Category, 0, len(results))
	for _, r := range results {
		id, err := pkg.ParseULID(r.Id)
		if err != nil {
			continue
		}
		uid, err := pkg.ParseULID(r.UserId)
		if err != nil {
			continue
		}
		items = append(items, &transaction.Category{
			Id:        id,
			UserId:    uid,
			Name:      r.Name,
			Icon:      r.Icon,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	return items, nil
}

func (r *DashboardRepository) GetMonthExpenses(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, month, year int) ([]*dashboard.TransactionSummary, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	type transactionResult struct {
		Id           string    `gorm:"column:id"`
		Type         string    `gorm:"column:type"`
		Amount       float64   `gorm:"column:amount"`
		Description  string    `gorm:"column:description"`
		CategoryId   *string   `gorm:"column:category_id"`
		CategoryName string    `gorm:"column:category_name"`
		Date         time.Time `gorm:"column:date"`
	}

	query := r.DB.WithContext(ctx).Table("transactions t").
		Select("t.id, t.type, t.amount, t.description, t.category_id, c.name as category_name, t.date").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ? AND t.type = ? AND t.date >= ? AND t.date < ?", userID.String(), "EXPENSE", startDate, endDate)
	if accountID != nil {
		query = query.Where("t.account_id = ?", accountID.String())
	}

	var results []transactionResult
	if err := query.Order("t.date DESC, t.created_at DESC").Scan(&results).Error; err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	items := make([]*dashboard.TransactionSummary, 0, len(results))
	for _, r := range results {
		id, err := pkg.ParseULID(r.Id)
		if err != nil {
			continue
		}
		var categoryID ulid.ULID
		if r.CategoryId != nil && *r.CategoryId != "" {
			parsed, err := pkg.ParseULID(*r.CategoryId)
			if err == nil {
				categoryID = parsed
			}
		}
		items = append(items, &dashboard.TransactionSummary{
			Id:           id,
			Type:         r.Type,
			Amount:       r.Amount,
			Description:  r.Description,
			CategoryId:   categoryID,
			CategoryName: r.CategoryName,
			Date:         r.Date,
		})
	}

	return items, nil
}
