package infrastructure

import (
	"gorm.io/gorm"
)

type ResourceCounter struct {
	DB *gorm.DB
}

func (r *ResourceCounter) CountTransactions(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("transactions").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountCategories(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("categories").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountAccounts(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("accounts").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountGoals(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("goals").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountInvestments(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("investments").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountBudgets(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("budgets").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountRecurring(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("recurring_transactions").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ResourceCounter) CountCreditCards(userID string) (int64, error) {
	var count int64
	err := r.DB.Table("credit_cards").Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
