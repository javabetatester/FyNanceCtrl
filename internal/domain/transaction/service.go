package transaction

import (
	"context"
	"errors"

	"Fynance/internal/domain/account"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Service struct {
	Repository         Repository
	CategoryRepository CategoryRepository
	UserService        *user.Service
	AccountService     *account.Service
}

func (s *Service) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	if err := s.ensureUserExists(ctx, transaction.UserId); err != nil {
		return err
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, transaction.AccountId, transaction.UserId)
	if err != nil {
		return err
	}

	err = s.CategoryValidation(ctx, transaction.CategoryId, transaction.UserId)
	if err != nil {
		return err
	}

	if err := s.validateTransactionBalance(ctx, transaction, accountEntity); err != nil {
		return err
	}

	TransactionCreateStruct(transaction)
	if err := s.Repository.Create(ctx, transaction); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if err := s.updateAccountBalance(ctx, transaction, accountEntity); err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateTransaction(ctx context.Context, transaction *Transaction) error {
	if err := s.ensureUserExists(ctx, transaction.UserId); err != nil {
		return err
	}

	storedTransaction, err := s.GetTransactionByID(ctx, transaction.Id, transaction.UserId)
	if err != nil {
		return err
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, transaction.AccountId, transaction.UserId)
	if err != nil {
		return err
	}

	oldAccountEntity, err := s.AccountService.GetAccountByID(ctx, storedTransaction.AccountId, transaction.UserId)
	if err != nil {
		return err
	}

	transaction.UpdatedAt = time.Now()

	err = s.UpdateTransactionValidation(ctx, transaction)
	if err != nil {
		return err
	}

	if err := s.validateTransactionBalance(ctx, transaction, accountEntity); err != nil {
		return err
	}

	if storedTransaction.AccountId != transaction.AccountId {
		if err := s.revertAccountBalance(ctx, storedTransaction, oldAccountEntity); err != nil {
			return err
		}
	} else {
		amountDiff := transaction.Amount - storedTransaction.Amount
		if storedTransaction.Type == Expense {
			amountDiff = -amountDiff
		}

		if accountEntity.Type != account.TypeCreditCard {
			if accountEntity.Balance+amountDiff < 0 {
				return appErrors.NewValidationError("amount", "saldo insuficiente")
			}
		}

		if err := s.AccountService.UpdateBalance(ctx, transaction.AccountId, transaction.UserId, amountDiff); err != nil {
			return err
		}
	}

	storedTransaction.AccountId = transaction.AccountId
	storedTransaction.CategoryId = transaction.CategoryId
	storedTransaction.Amount = transaction.Amount
	storedTransaction.Description = transaction.Description
	storedTransaction.Type = transaction.Type
	if !transaction.Date.IsZero() {
		storedTransaction.Date = transaction.Date
	}
	storedTransaction.UpdatedAt = transaction.UpdatedAt

	if err := s.Repository.Update(ctx, storedTransaction); err != nil {
		return err
	}

	if storedTransaction.AccountId != transaction.AccountId {
		if err := s.updateAccountBalance(ctx, transaction, accountEntity); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) DeleteTransaction(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) error {
	transactionEntity, err := s.GetTransactionByID(ctx, transactionID, userID)
	if err != nil {
		return err
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, transactionEntity.AccountId, userID)
	if err != nil {
		return err
	}

	if err := s.revertAccountBalance(ctx, transactionEntity, accountEntity); err != nil {
		return err
	}

	return s.Repository.Delete(ctx, transactionID)
}

func (s *Service) GetTransactionByID(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) (*Transaction, error) {
	transaction, err := s.Repository.GetByIDAndUser(ctx, transactionID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrTransactionNotFound
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return transaction, nil
}

func (s *Service) GetAllTransactions(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, pagination *pkg.PaginationParams) ([]*Transaction, int64, error) {
	transactions, total, err := s.Repository.GetAll(ctx, userID, accountID, pagination)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	return transactions, total, nil
}

func (s *Service) GetTransactionsByAmount(ctx context.Context, amount float64, pagination *pkg.PaginationParams) ([]*Transaction, int64, error) {
	transactions, total, err := s.Repository.GetByAmount(ctx, amount, pagination)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	return transactions, total, nil
}

func (s *Service) GetTransactionsByName(ctx context.Context, name string, pagination *pkg.PaginationParams) ([]*Transaction, int64, error) {
	transactions, total, err := s.Repository.GetByName(ctx, name, pagination)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	return transactions, total, nil
}

func (s *Service) GetTransactionsByCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Transaction, int64, error) {
	transactions, total, err := s.Repository.GetByCategory(ctx, categoryID, userID, pagination)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	return transactions, total, nil
}

func (s *Service) CreateCategory(ctx context.Context, category *Category) error {
	if err := s.ensureUserExists(ctx, category.UserId); err != nil {
		return err
	}

	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	if err := s.CategoryExists(ctx, category.Name, category.UserId); err != nil {
		return err
	}

	CategoryCreateStruct(category)

	if err := s.CategoryRepository.Create(ctx, category); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) UpdateCategory(ctx context.Context, category *Category) error {
	if err := s.ensureUserExists(ctx, category.UserId); err != nil {
		return err
	}

	existingCategory, err := s.CategoryRepository.GetByID(ctx, category.Id, category.UserId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrCategoryNotFound
	}
	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	if !strings.EqualFold(existingCategory.Name, category.Name) {
		if err := s.CategoryExists(ctx, category.Name, category.UserId); err != nil {
			return err
		}
	}

	existingCategory.Name = category.Name
	existingCategory.Icon = category.Icon
	existingCategory.UpdatedAt = time.Now()

	return s.CategoryRepository.Update(ctx, existingCategory)
}

func (s *Service) DeleteCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return err
	}

	if _, err := s.CategoryRepository.GetByID(ctx, categoryID, userID); errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrCategoryNotFound
	} else if err != nil {
		return appErrors.NewDatabaseError(err)
	}
	return s.CategoryRepository.Delete(ctx, categoryID, userID)
}

func (s *Service) GetCategoryByID(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (*Category, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	category, err := s.CategoryRepository.GetByID(ctx, categoryID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if s.isDefaultCategory(categoryID, userID) {
			defaultCategories := GetDefaultCategoriesAsDomain(userID)
			for _, defaultCat := range defaultCategories {
				if defaultCat.Id == categoryID {
					return defaultCat, nil
				}
			}
		}
		return nil, appErrors.ErrCategoryNotFound
	}
	if err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return category, nil
}

func (s *Service) GetAllCategories(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Category, int64, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}
	
	customCategories, _, err := s.CategoryRepository.GetAll(ctx, userID, nil)
	if err != nil {
		return nil, 0, appErrors.NewDatabaseError(err)
	}
	
	defaultCategories := GetDefaultCategoriesAsDomain(userID)
	
	customMap := make(map[string]*Category)
	for _, cat := range customCategories {
		customMap[cat.Name] = cat
	}
	
	allCategories := make([]*Category, 0, len(defaultCategories)+len(customCategories))
	for _, defaultCat := range defaultCategories {
		if customCat, exists := customMap[defaultCat.Name]; exists {
			allCategories = append(allCategories, customCat)
		} else {
			allCategories = append(allCategories, defaultCat)
		}
	}
	
	for _, customCat := range customCategories {
		isDefault := false
		for _, defaultCat := range DefaultCategories {
			if customCat.Name == defaultCat.Name {
				isDefault = true
				break
			}
		}
		if !isDefault {
			allCategories = append(allCategories, customCat)
		}
	}
	
	total := int64(len(allCategories))
	
	if pagination != nil {
		pagination.Normalize()
		start := pagination.Offset()
		end := start + pagination.Limit
		
		if start >= len(allCategories) {
			return []*Category{}, total, nil
		}
		if end > len(allCategories) {
			end = len(allCategories)
		}
		
		allCategories = allCategories[start:end]
	}
	
	return allCategories, total, nil
}

func (s *Service) CategoryExists(ctx context.Context, categoryName string, userID ulid.ULID) error {
	trimmedName := strings.TrimSpace(categoryName)
	if trimmedName == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}

	_, err := s.CategoryRepository.GetByName(ctx, trimmedName, userID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return appErrors.NewConflictError("categoria")
}

func (s *Service) CategoryValidation(ctx context.Context, categoryId ulid.ULID, userID ulid.ULID) error {
	category, err := s.CategoryRepository.GetByID(ctx, categoryId, userID)
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if s.isDefaultCategory(categoryId, userID) {
			if err := s.createDefaultCategoryIfNeeded(ctx, categoryId, userID); err != nil {
				return appErrors.ErrCategoryNotFound
			}
			return nil
		}
		return appErrors.ErrCategoryNotFound
	}

	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if category != nil {
		return nil
	}

	return appErrors.ErrCategoryNotFound
}

func (s *Service) isDefaultCategory(categoryID ulid.ULID, userID ulid.ULID) bool {
	defaultCategories := GetDefaultCategoriesAsDomain(userID)
	for _, defaultCat := range defaultCategories {
		if defaultCat.Id == categoryID {
			return true
		}
	}
	return false
}

func (s *Service) createDefaultCategoryIfNeeded(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error {
	defaultCategories := GetDefaultCategoriesAsDomain(userID)
	for _, defaultCat := range defaultCategories {
		if defaultCat.Id == categoryID {
			existing, err := s.CategoryRepository.GetByName(ctx, defaultCat.Name, userID)
			if err == nil && existing != nil {
				return nil
			}
			
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				if isUniqueConstraintError(err) {
					return nil
				}
				return err
			}
			
			category := &Category{
				Id:        defaultCat.Id,
				UserId:    userID,
				Name:      defaultCat.Name,
				Icon:      defaultCat.Icon,
				CreatedAt: defaultCat.CreatedAt,
				UpdatedAt: defaultCat.UpdatedAt,
			}
			
			if err := s.CategoryRepository.Create(ctx, category); err != nil {
				if isUniqueConstraintError(err) {
					return nil
				}
				return err
			}
			return nil
		}
	}
	return appErrors.ErrCategoryNotFound
}

func (s *Service) GetNumberOfTransactions(ctx context.Context, userID ulid.ULID) (int64, error) {
	count, err := s.Repository.GetNumberOfTransactions(ctx, userID)
	if err != nil {
		return 0, appErrors.NewDatabaseError(err)
	}
	return count, nil
}

func TransactionCreateStruct(transaction *Transaction) {
	transaction.Id = pkg.GenerateULIDObject()
	now := pkg.SetTimestamps()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now
}

func CategoryCreateStruct(category *Category) {
	category.Id = pkg.GenerateULIDObject()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
}

func (s *Service) UpdateTransactionValidation(ctx context.Context, transaction *Transaction) error {
	if transaction.Amount < 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	if _, err := s.GetCategoryByID(ctx, transaction.CategoryId, transaction.UserId); err != nil {
		return err
	}

	return nil
}

func (s *Service) TransactionExists(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) error {
	_, err := s.GetTransactionByID(ctx, transactionID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrTransactionNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateDefaultCategories(ctx context.Context, userID ulid.ULID) error {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return err
	}

	for _, defaultCat := range DefaultCategories {
		categoryName := strings.TrimSpace(defaultCat.Name)
		if categoryName == "" {
			continue
		}

		existing, err := s.CategoryRepository.GetByName(ctx, categoryName, userID)
		if err == nil && existing != nil {
			continue
		}

		category := &Category{
			UserId: userID,
			Name:   categoryName,
			Icon:   defaultCat.Icon,
		}

		CategoryCreateStruct(category)
		if err := s.CategoryRepository.Create(ctx, category); err != nil {
			if isUniqueConstraintError(err) {
				continue
			}
		}
	}

	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "23505") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "violates unique constraint") ||
		strings.Contains(errStr, "idx_categories_user_name")
}

func (s *Service) ensureUserExists(ctx context.Context, userID ulid.ULID) error {
	if s.UserService == nil {
		return appErrors.ErrInternalServer.WithError(errors.New("user service not configured"))
	}
	_, err := s.UserService.GetByID(ctx, userID)
	if err != nil {
		return appErrors.ErrUserNotFound
	}
	return nil
}

func (s *Service) validateTransactionBalance(_ context.Context, transaction *Transaction, accountEntity *account.Account) error {
	if transaction.Type == Expense {
		if accountEntity.Type != account.TypeCreditCard {
			if accountEntity.Balance < transaction.Amount {
				return appErrors.NewValidationError("amount", "saldo insuficiente")
			}
		}
	}
	return nil
}

func (s *Service) updateAccountBalance(ctx context.Context, transaction *Transaction, _ *account.Account) error {
	var amount float64
	switch transaction.Type {
	case Receipt:
		amount = transaction.Amount
	case Expense:
		amount = -transaction.Amount
	default:
		return nil
	}

	return s.AccountService.UpdateBalance(ctx, transaction.AccountId, transaction.UserId, amount)
}

func (s *Service) revertAccountBalance(ctx context.Context, transaction *Transaction, _ *account.Account) error {
	var amount float64
	switch transaction.Type {
	case Receipt:
		amount = -transaction.Amount
	case Expense:
		amount = transaction.Amount
	default:
		return nil
	}

	return s.AccountService.UpdateBalance(ctx, transaction.AccountId, transaction.UserId, amount)
}
