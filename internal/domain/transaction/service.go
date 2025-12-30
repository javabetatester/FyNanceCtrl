package transaction

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/account"
	"Fynance/internal/domain/category"
	"Fynance/internal/domain/shared"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/logger"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type BudgetUpdater interface {
	UpdateSpent(ctx context.Context, categoryID, userID ulid.ULID, amount float64) error
	UpdateSpentWithDate(ctx context.Context, categoryID, userID ulid.ULID, amount float64, transactionDate time.Time) error
}

type GoalContributionHandler interface {
	DeleteContributionByTransactionId(ctx context.Context, transactionID, userID ulid.ULID) error
}

type InvestmentTransactionHandler interface {
	DeleteInvestmentTransactionByTransactionId(ctx context.Context, transactionID, userID ulid.ULID) error
}

type Service struct {
	Repository        Repository
	CategoryService   *category.Service
	AccountService    *account.Service
	BudgetService     BudgetUpdater
	GoalService       GoalContributionHandler
	InvestmentService InvestmentTransactionHandler
	shared.BaseService
}

func NewService(
	repo Repository,
	categoryService *category.Service,
	accountService *account.Service,
	userChecker *shared.UserCheckerService,
) *Service {
	return &Service{
		Repository:      repo,
		CategoryService: categoryService,
		AccountService:  accountService,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) SetBudgetService(budgetService BudgetUpdater) {
	s.BudgetService = budgetService
}

func (s *Service) SetGoalService(goalService GoalContributionHandler) {
	s.GoalService = goalService
}

func (s *Service) SetInvestmentService(investmentService InvestmentTransactionHandler) {
	s.InvestmentService = investmentService
}

func (s *Service) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	if err := s.EnsureUserExists(ctx, transaction.UserId); err != nil {
		return err
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, transaction.AccountId, transaction.UserId)
	if err != nil {
		return err
	}

	if err := s.validateCreditCardRestrictions(accountEntity, transaction); err != nil {
		return err
	}

	if err := s.validateAndResolveCategory(ctx, transaction); err != nil {
		return err
	}

	if accountEntity.Type == account.TypeCreditCard {
		s.initTransaction(transaction)
		if err := s.Repository.Create(ctx, transaction); err != nil {
			return appErrors.NewDatabaseError(err)
		}
		return nil
	}

	if err := s.validateBalance(transaction, accountEntity); err != nil {
		return err
	}

	s.initTransaction(transaction)
	if err := s.Repository.Create(ctx, transaction); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if err := s.updateAccountBalance(ctx, transaction, accountEntity); err != nil {
		return err
	}

	s.updateBudgetIfExpense(ctx, transaction)

	return nil
}

func (s *Service) UpdateTransaction(ctx context.Context, transaction *Transaction) error {
	if err := s.EnsureUserExists(ctx, transaction.UserId); err != nil {
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

	if err := s.validateCreditCardRestrictions(accountEntity, transaction); err != nil {
		return err
	}

	oldAccountEntity, err := s.AccountService.GetAccountByID(ctx, storedTransaction.AccountId, transaction.UserId)
	if err != nil {
		return err
	}

	transaction.UpdatedAt = time.Now()

	if err := s.validateUpdate(ctx, transaction); err != nil {
		return err
	}

	if err := s.validateBalance(transaction, accountEntity); err != nil {
		return err
	}

	if err := s.processBalanceUpdate(ctx, storedTransaction, transaction, oldAccountEntity, accountEntity); err != nil {
		return err
	}

	oldCategoryId := storedTransaction.CategoryId
	oldDate := storedTransaction.Date
	oldType := storedTransaction.Type
	oldAmount := storedTransaction.Amount

	s.applyTransactionUpdate(storedTransaction, transaction)

	if err := s.Repository.Update(ctx, storedTransaction); err != nil {
		return err
	}

	if storedTransaction.AccountId != transaction.AccountId {
		if err := s.updateAccountBalance(ctx, transaction, accountEntity); err != nil {
			return err
		}
	}

	s.updateBudgetOnChange(ctx, transaction.UserId, oldCategoryId, oldDate, oldType, oldAmount, transaction)

	return nil
}

func (s *Service) DeleteTransaction(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) error {
	transactionEntity, err := s.GetTransactionByID(ctx, transactionID, userID)
	if err != nil {
		return err
	}
	accountEntity, err := s.AccountService.GetAccountByID(ctx, transactionEntity.AccountId, userID)
	if err != nil {
		appErr, isAppErr := appErrors.AsAppError(err)
		if !isAppErr || appErr.Code != "NOT_FOUND" {
			return err
		}
	} else {
		if err := s.revertAccountBalance(ctx, transactionEntity, accountEntity); err != nil {
			return err
		}
	}

	s.revertBudgetIfExpense(ctx, transactionEntity)

	if transactionEntity.Type == Goals && s.GoalService != nil {
		_ = s.GoalService.DeleteContributionByTransactionId(ctx, transactionID, userID)
	}

	if (transactionEntity.Type == Investment || transactionEntity.Type == Withdraw) &&
		transactionEntity.InvestmentId != nil && s.InvestmentService != nil {
		_ = s.InvestmentService.DeleteInvestmentTransactionByTransactionId(ctx, transactionID, userID)
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

func (s *Service) GetAllTransactions(ctx context.Context, userID ulid.ULID, accountID *ulid.ULID, filters *TransactionFilters, pagination *pkg.PaginationParams) ([]*Transaction, int64, error) {
	return s.Repository.GetAll(ctx, userID, accountID, filters, pagination)
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

func (s *Service) GetNumberOfTransactions(ctx context.Context, userID ulid.ULID) (int64, error) {
	count, err := s.Repository.GetNumberOfTransactions(ctx, userID)
	if err != nil {
		return 0, appErrors.NewDatabaseError(err)
	}
	return count, nil
}

func (s *Service) CategoryValidation(ctx context.Context, categoryId ulid.ULID, userID ulid.ULID) error {
	if s.CategoryService == nil {
		return appErrors.ErrInternalServer
	}
	return s.CategoryService.ValidateAndEnsureExists(ctx, categoryId, userID)
}

func (s *Service) CreateDefaultCategories(ctx context.Context, userID ulid.ULID) error {
	if s.CategoryService == nil {
		return appErrors.ErrInternalServer
	}
	return s.CategoryService.CreateDefaultCategories(ctx, userID)
}

func (s *Service) CreateCategory(ctx context.Context, cat *category.Category) error {
	if s.CategoryService == nil {
		return appErrors.ErrInternalServer
	}
	return s.CategoryService.Create(ctx, cat)
}

func (s *Service) UpdateCategory(ctx context.Context, cat *category.Category) error {
	if s.CategoryService == nil {
		return appErrors.ErrInternalServer
	}
	return s.CategoryService.Update(ctx, cat)
}

func (s *Service) DeleteCategory(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) error {
	if s.CategoryService == nil {
		return appErrors.ErrInternalServer
	}
	return s.CategoryService.Delete(ctx, categoryID, userID)
}

func (s *Service) GetCategoryByID(ctx context.Context, categoryID ulid.ULID, userID ulid.ULID) (*category.Category, error) {
	if s.CategoryService == nil {
		return nil, appErrors.ErrInternalServer
	}
	return s.CategoryService.GetByID(ctx, categoryID, userID)
}

func (s *Service) GetAllCategories(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*category.Category, int64, error) {
	if s.CategoryService == nil {
		return nil, 0, appErrors.ErrInternalServer
	}
	return s.CategoryService.GetAll(ctx, userID, pagination)
}

func (s *Service) validateCreditCardRestrictions(accountEntity *account.Account, transaction *Transaction) error {
	if accountEntity.Type == account.TypeCreditCard && transaction.Type != Expense {
		return appErrors.NewValidationError("type", "cartao de credito so permite despesas")
	}
	return nil
}

func (s *Service) validateAndResolveCategory(ctx context.Context, transaction *Transaction) error {
	if transaction.Type == Investment || transaction.Type == Withdraw {
		return nil
	}

	if transaction.CategoryId == nil {
		return appErrors.NewValidationError("category_id", "é obrigatório")
	}

	if s.CategoryService == nil {
		return nil
	}

	return s.CategoryService.ValidateAndEnsureExists(ctx, *transaction.CategoryId, transaction.UserId)
}

func (s *Service) validateBalance(transaction *Transaction, accountEntity *account.Account) error {
	if transaction.Type != Expense {
		return nil
	}

	if accountEntity.Type == account.TypeCreditCard {
		return nil
	}

	amount := transaction.Amount
	if amount < 0 {
		amount = -amount
	}

	if accountEntity.Balance < amount {
		return appErrors.NewValidationError("valor", "saldo insuficiente")
	}

	return nil
}

func (s *Service) validateUpdate(ctx context.Context, transaction *Transaction) error {
	if transaction.Amount == 0 {
		return appErrors.NewValidationError("valor", "deve ser diferente de zero")
	}

	if transaction.Type != Investment && transaction.Type != Withdraw {
		if transaction.CategoryId == nil {
			return appErrors.NewValidationError("category_id", "é obrigatório")
		}

		if s.CategoryService != nil {
			if _, err := s.CategoryService.GetByID(ctx, *transaction.CategoryId, transaction.UserId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) processBalanceUpdate(ctx context.Context, stored, updated *Transaction, oldAccount, newAccount *account.Account) error {
	if stored.AccountId != updated.AccountId {
		return s.revertAccountBalance(ctx, stored, oldAccount)
	}

	amountDiff := updated.Amount - stored.Amount
	if stored.Type == Expense {
		amountDiff = -amountDiff
	}

	if newAccount.Type != account.TypeCreditCard {
		if newAccount.Balance+amountDiff < 0 {
			return appErrors.NewValidationError("valor", "saldo insuficiente")
		}
	}

	return s.AccountService.UpdateBalance(ctx, updated.AccountId, updated.UserId, amountDiff)
}

func (s *Service) applyTransactionUpdate(stored, updated *Transaction) {
	stored.AccountId = updated.AccountId
	stored.CategoryId = updated.CategoryId
	stored.Amount = updated.Amount
	stored.Description = updated.Description
	stored.Type = updated.Type
	if !updated.Date.IsZero() {
		stored.Date = updated.Date
	}
	stored.UpdatedAt = updated.UpdatedAt
}

func (s *Service) updateAccountBalance(ctx context.Context, transaction *Transaction, accountEntity *account.Account) error {
	if accountEntity.Type == account.TypeCreditCard {
		return nil
	}

	var amount float64
	switch transaction.Type {
	case Receipt:
		amount = transaction.Amount
	case Expense:
		if transaction.Amount < 0 {
			amount = transaction.Amount
		} else {
			amount = -transaction.Amount
		}
	default:
		return nil
	}

	return s.AccountService.UpdateBalance(ctx, transaction.AccountId, transaction.UserId, amount)
}

func (s *Service) revertAccountBalance(ctx context.Context, transaction *Transaction, accountEntity *account.Account) error {
	if accountEntity.Type == account.TypeCreditCard {
		return nil
	}

	var amount float64
	switch transaction.Type {
	case Receipt:
		amount = -transaction.Amount
	case Expense:
		if transaction.Amount < 0 {
			amount = -transaction.Amount
		} else {
			amount = transaction.Amount
		}
	default:
		return nil
	}

	return s.AccountService.UpdateBalance(ctx, transaction.AccountId, transaction.UserId, amount)
}

func (s *Service) updateBudgetIfExpense(ctx context.Context, transaction *Transaction) {
	if transaction.Type != Expense || s.BudgetService == nil || transaction.CategoryId == nil {
		return
	}

	spentAmount := transaction.Amount
	if spentAmount < 0 {
		spentAmount = -spentAmount
	}

	if err := s.BudgetService.UpdateSpentWithDate(ctx, *transaction.CategoryId, transaction.UserId, spentAmount, transaction.Date); err != nil {
		logger.Error().
			Err(err).
			Str("category_id", transaction.CategoryId.String()).
			Str("user_id", transaction.UserId.String()).
			Float64("amount", spentAmount).
			Msg("error updating budget spent")
	}
}

func (s *Service) revertBudgetIfExpense(ctx context.Context, transaction *Transaction) {
	if transaction.Type != Expense || s.BudgetService == nil || transaction.CategoryId == nil {
		return
	}

	spentAmount := transaction.Amount
	if spentAmount < 0 {
		spentAmount = -spentAmount
	}

	_ = s.BudgetService.UpdateSpentWithDate(ctx, *transaction.CategoryId, transaction.UserId, -spentAmount, transaction.Date)
}

func (s *Service) updateBudgetOnChange(ctx context.Context, userID ulid.ULID, oldCategoryId *ulid.ULID, oldDate time.Time, oldType Types, oldAmount float64, newTx *Transaction) {
	if s.BudgetService == nil {
		return
	}

	if oldType == Expense && oldCategoryId != nil {
		oldSpentAmount := oldAmount
		if oldSpentAmount < 0 {
			oldSpentAmount = -oldSpentAmount
		}
		_ = s.BudgetService.UpdateSpentWithDate(ctx, *oldCategoryId, userID, -oldSpentAmount, oldDate)
	}

	if newTx.Type == Expense && newTx.CategoryId != nil {
		newSpentAmount := newTx.Amount
		if newSpentAmount < 0 {
			newSpentAmount = -newSpentAmount
		}
		transactionDate := newTx.Date
		if transactionDate.IsZero() {
			transactionDate = time.Now()
		}
		_ = s.BudgetService.UpdateSpentWithDate(ctx, *newTx.CategoryId, userID, newSpentAmount, transactionDate)
	}
}

func (s *Service) initTransaction(transaction *Transaction) {
	transaction.Id = pkg.GenerateULIDObject()
	now := pkg.SetTimestamps()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now
}

func TransactionCreateStruct(transaction *Transaction) {
	transaction.Id = pkg.GenerateULIDObject()
	now := pkg.SetTimestamps()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now
}

func NormalizeCategoryName(name string) string {
	return shared.NormalizeName(name)
}
