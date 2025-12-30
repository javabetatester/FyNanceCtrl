package infrastructure

import (
	"time"

	"Fynance/internal/domain/report"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type ReportRepository struct {
	DB *gorm.DB
}

func (r *ReportRepository) GetMonthlyReport(userID ulid.ULID, month, year int) (*report.MonthlyReport, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var totalIncome float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "RECEIPT", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)

	var totalExpenses float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "EXPENSE", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	netBalance := totalIncome - totalExpenses
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (netBalance / totalIncome) * 100
	}

	incomeByCategory := r.getCategoryBreakdown(userID, "RECEIPT", startDate, endDate, totalIncome)
	expensesByCategory := r.getCategoryBreakdown(userID, "EXPENSE", startDate, endDate, totalExpenses)
	dailyBalance := r.getDailyBalance(userID, startDate, endDate)
	topExpenses := r.getTopExpenses(userID, startDate, endDate, 10)
	comparison := r.getMonthComparison(userID, month, year)

	return &report.MonthlyReport{
		UserId:             userID,
		Month:              month,
		Year:               year,
		TotalIncome:        totalIncome,
		TotalExpenses:      totalExpenses,
		NetBalance:         netBalance,
		SavingsRate:        savingsRate,
		IncomeByCategory:   incomeByCategory,
		ExpensesByCategory: expensesByCategory,
		DailyBalance:       dailyBalance,
		TopExpenses:        topExpenses,
		Comparison:         comparison,
	}, nil
}

func (r *ReportRepository) getCategoryBreakdown(userID ulid.ULID, txType string, startDate, endDate time.Time, total float64) []report.CategoryAmount {
	type result struct {
		CategoryId   string
		CategoryName string
		Amount       float64
		Count        int
	}

	var results []result
	r.DB.Table("transactions t").
		Select("t.category_id, c.name as category_name, COALESCE(SUM(t.amount), 0) as amount, COUNT(*) as count").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ? AND t.type = ? AND t.date BETWEEN ? AND ?", userID.String(), txType, startDate, endDate).
		Group("t.category_id, c.name").
		Order("amount DESC").
		Scan(&results)

	categories := make([]report.CategoryAmount, 0, len(results))
	for _, res := range results {
		categoryID, err := pkg.ParseULID(res.CategoryId)
		if err != nil {
			continue
		}
		percentage := 0.0
		if total > 0 {
			percentage = (res.Amount / total) * 100
		}
		categories = append(categories, report.CategoryAmount{
			CategoryId:   categoryID,
			CategoryName: res.CategoryName,
			Amount:       res.Amount,
			Percentage:   percentage,
			Count:        res.Count,
		})
	}
	return categories
}

func (r *ReportRepository) getDailyBalance(userID ulid.ULID, startDate, endDate time.Time) []report.DailyBalance {
	type result struct {
		Date   time.Time
		Type   string
		Amount float64
	}

	var results []result
	r.DB.Table("transactions").
		Select("DATE(date) as date, type, COALESCE(SUM(amount), 0) as amount").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID.String(), startDate, endDate).
		Group("DATE(date), type").
		Order("date ASC").
		Scan(&results)

	dailyMap := make(map[string]*report.DailyBalance)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dailyMap[dateStr] = &report.DailyBalance{
			Date:     dateStr,
			Income:   0,
			Expenses: 0,
			Balance:  0,
		}
	}

	for _, res := range results {
		dateStr := res.Date.Format("2006-01-02")
		if daily, ok := dailyMap[dateStr]; ok {
			switch res.Type {
			case "RECEIPT":
				daily.Income = res.Amount
			case "EXPENSE":
				daily.Expenses = res.Amount
			}
		}
	}

	dailyBalance := make([]report.DailyBalance, 0, len(dailyMap))
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if daily, ok := dailyMap[dateStr]; ok {
			daily.Balance = daily.Income - daily.Expenses
			dailyBalance = append(dailyBalance, *daily)
		}
	}

	return dailyBalance
}

func (r *ReportRepository) getTopExpenses(userID ulid.ULID, startDate, endDate time.Time, limit int) []report.TransactionItem {
	type result struct {
		Id           string
		Description  string
		Amount       float64
		CategoryName string `gorm:"column:category_name"`
		Date         time.Time
	}

	var results []result
	r.DB.Table("transactions t").
		Select("t.id, t.description, t.amount, c.name as category_name, t.date").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.user_id = ? AND t.type = ? AND t.date BETWEEN ? AND ?", userID.String(), "EXPENSE", startDate, endDate).
		Order("t.amount DESC").
		Limit(limit).
		Scan(&results)

	items := make([]report.TransactionItem, 0, len(results))
	for _, res := range results {
		id, err := pkg.ParseULID(res.Id)
		if err != nil {
			continue
		}
		items = append(items, report.TransactionItem{
			Id:          id,
			Description: res.Description,
			Amount:      res.Amount,
			Category:    res.CategoryName,
			Date:        res.Date,
		})
	}
	return items
}

func (r *ReportRepository) getMonthComparison(userID ulid.ULID, month, year int) *report.MonthComparison {
	prevDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	prevStartDate := time.Date(prevDate.Year(), prevDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	prevEndDate := prevStartDate.AddDate(0, 1, 0).Add(-time.Second)

	var prevIncome float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "RECEIPT", prevStartDate, prevEndDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&prevIncome)

	var prevExpenses float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "EXPENSE", prevStartDate, prevEndDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&prevExpenses)

	currStartDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	currEndDate := currStartDate.AddDate(0, 1, 0).Add(-time.Second)

	var currIncome float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "RECEIPT", currStartDate, currEndDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&currIncome)

	var currExpenses float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "EXPENSE", currStartDate, currEndDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&currExpenses)

	incomeChange := currIncome - prevIncome
	expensesChange := currExpenses - prevExpenses
	incomeChangePerc := 0.0
	expensesChangePerc := 0.0

	if prevIncome > 0 {
		incomeChangePerc = (incomeChange / prevIncome) * 100
	}
	if prevExpenses > 0 {
		expensesChangePerc = (expensesChange / prevExpenses) * 100
	}

	return &report.MonthComparison{
		PreviousIncome:     prevIncome,
		PreviousExpenses:   prevExpenses,
		IncomeChange:       incomeChange,
		ExpensesChange:     expensesChange,
		IncomeChangePerc:   incomeChangePerc,
		ExpensesChangePerc: expensesChangePerc,
	}
}

func (r *ReportRepository) GetYearlyReport(userID ulid.ULID, year int) (*report.YearlyReport, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	var totalIncome float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "RECEIPT", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)

	var totalExpenses float64
	r.DB.Table("transactions").
		Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "EXPENSE", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	netBalance := totalIncome - totalExpenses
	averageSavings := netBalance / 12

	monthlyBreakdown := make([]report.MonthSummary, 0, 12)
	for m := 1; m <= 12; m++ {
		mStart := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		mEnd := mStart.AddDate(0, 1, 0).Add(-time.Second)

		var mIncome float64
		r.DB.Table("transactions").
			Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "RECEIPT", mStart, mEnd).
			Select("COALESCE(SUM(amount), 0)").Scan(&mIncome)

		var mExpenses float64
		r.DB.Table("transactions").
			Where("user_id = ? AND type = ? AND date BETWEEN ? AND ?", userID.String(), "EXPENSE", mStart, mEnd).
			Select("COALESCE(SUM(amount), 0)").Scan(&mExpenses)

		monthlyBreakdown = append(monthlyBreakdown, report.MonthSummary{
			Month:    m,
			Income:   mIncome,
			Expenses: mExpenses,
			Balance:  mIncome - mExpenses,
		})
	}

	topCategories := r.getCategoryBreakdown(userID, "EXPENSE", startDate, endDate, totalExpenses)

	return &report.YearlyReport{
		UserId:           userID,
		Year:             year,
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
		NetBalance:       netBalance,
		AverageSavings:   averageSavings,
		MonthlyBreakdown: monthlyBreakdown,
		TopCategories:    topCategories,
	}, nil
}

func (r *ReportRepository) GetCategoryReport(userID, categoryID ulid.ULID, startDate, endDate time.Time) (*report.CategoryReport, error) {
	var categoryName string
	r.DB.Table("categories").Where("id = ?", categoryID.String()).Select("name").Scan(&categoryName)

	var totalAmount float64
	var count int64
	r.DB.Table("transactions").
		Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID.String(), categoryID.String(), startDate, endDate).
		Select("COALESCE(SUM(amount), 0) as total_amount").Scan(&totalAmount)

	r.DB.Table("transactions").
		Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID.String(), categoryID.String(), startDate, endDate).
		Count(&count)

	average := 0.0
	if count > 0 {
		average = totalAmount / float64(count)
	}

	type txResult struct {
		Id          string
		Description string
		Amount      float64
		Date        time.Time
	}

	var txResults []txResult
	r.DB.Table("transactions").
		Select("id, description, amount, date").
		Where("user_id = ? AND category_id = ? AND date BETWEEN ? AND ?", userID.String(), categoryID.String(), startDate, endDate).
		Order("date DESC").
		Limit(50).
		Scan(&txResults)

	transactions := make([]report.TransactionItem, 0, len(txResults))
	for _, tx := range txResults {
		id, err := pkg.ParseULID(tx.Id)
		if err != nil {
			continue
		}
		transactions = append(transactions, report.TransactionItem{
			Id:          id,
			Description: tx.Description,
			Amount:      tx.Amount,
			Category:    categoryName,
			Date:        tx.Date,
		})
	}

	return &report.CategoryReport{
		CategoryId:   categoryID,
		CategoryName: categoryName,
		TotalAmount:  totalAmount,
		Count:        int(count),
		Average:      average,
		Transactions: transactions,
	}, nil
}
