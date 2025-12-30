package goal

import (
	"context"
	"errors"
	"strings"
	"time"

	"Fynance/internal/domain/account"
	"Fynance/internal/domain/category"
	domaincontracts "Fynance/internal/domain/contracts"
	"Fynance/internal/domain/shared"
	"Fynance/internal/domain/transaction"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type TransactionHandler interface {
	CreateTransaction(ctx context.Context, transaction *transaction.Transaction) error
	DeleteTransaction(ctx context.Context, transactionID ulid.ULID, userID ulid.ULID) error
}

type Service struct {
	Repository         Repository
	AccountService     *account.Service
	TransactionService TransactionHandler
	shared.BaseService
}

func NewService(repo Repository, accountService *account.Service, transactionService TransactionHandler, userChecker *shared.UserCheckerService) *Service {
	return &Service{
		Repository:         repo,
		AccountService:     accountService,
		TransactionService: transactionService,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) CreateGoal(ctx context.Context, request *domaincontracts.GoalCreateRequest) error {
	if err := s.validateGoalCreate(request); err != nil {
		return err
	}

	if err := s.EnsureUserExists(ctx, request.UserId); err != nil {
		return err
	}

	now := time.Now()
	entity := &Goal{
		Id:            pkg.GenerateULIDObject(),
		UserId:        request.UserId,
		Name:          request.Name,
		TargetAmount:  request.Target,
		CurrentAmount: 0,
		StartedAt:     now,
		EndedAt:       request.EndedAt,
		Status:        Active,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return s.Repository.Create(ctx, entity)
}

func (s *Service) MakeContribution(ctx context.Context, goalID, accountID, userID ulid.ULID, amount float64, description string) error {
	if amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	goal, err := s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return err
	}

	if goal.Status != Active {
		return appErrors.NewValidationError("goal", "meta nao esta ativa")
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if accountEntity.Type == account.TypeCreditCard {
		return appErrors.NewValidationError("account_id", "conta nao pode ser cartao de credito")
	}

	if accountEntity.Balance < amount {
		return appErrors.NewValidationError("amount", "saldo insuficiente na conta")
	}

	if err := s.AccountService.UpdateBalance(ctx, accountID, userID, -amount); err != nil {
		return err
	}

	var transactionID *ulid.ULID
	if s.TransactionService != nil {
		tx, err := s.createGoalTransaction(ctx, goal, accountID, userID, amount, description)
		if err != nil {
			_ = s.AccountService.UpdateBalance(ctx, accountID, userID, amount)
			return err
		}
		transactionID = &tx.Id
	}

	contribution := &Contribution{
		Id:            pkg.GenerateULIDObject(),
		GoalId:        goalID,
		UserId:        userID,
		AccountId:     accountID,
		TransactionId: transactionID,
		Type:          ContributionDeposit,
		Amount:        amount,
		Description:   strings.TrimSpace(description),
		CreatedAt:     time.Now(),
	}

	if err := s.Repository.CreateContribution(ctx, contribution); err != nil {
		s.rollbackContribution(ctx, transactionID, accountID, userID, amount)
		return err
	}

	if err := s.Repository.UpdateCurrentAmountAtomic(ctx, goalID, amount); err != nil {
		s.rollbackContribution(ctx, transactionID, accountID, userID, amount)
		_ = s.Repository.DeleteContribution(ctx, contribution.Id)
		return err
	}

	return s.checkAndUpdateGoalStatus(ctx, goalID, userID)
}

func (s *Service) WithdrawFromGoal(ctx context.Context, goalID, accountID, userID ulid.ULID, amount float64, description string) error {
	if amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	goal, err := s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return err
	}

	if goal.CurrentAmount < amount {
		return appErrors.NewValidationError("amount", "saldo insuficiente na meta")
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if accountEntity.Type == account.TypeCreditCard {
		return appErrors.NewValidationError("account_id", "conta nao pode ser cartao de credito")
	}

	contribution := &Contribution{
		Id:          pkg.GenerateULIDObject(),
		GoalId:      goalID,
		UserId:      userID,
		Type:        ContributionWithdraw,
		Amount:      amount,
		Description: strings.TrimSpace(description),
		CreatedAt:   time.Now(),
	}

	if err := s.Repository.CreateContribution(ctx, contribution); err != nil {
		return err
	}

	if err := s.Repository.UpdateCurrentAmountAtomic(ctx, goalID, -amount); err != nil {
		return err
	}

	if err := s.AccountService.UpdateBalance(ctx, accountID, userID, amount); err != nil {
		_ = s.Repository.UpdateCurrentAmountAtomic(ctx, goalID, amount)
		return err
	}

	goal, err = s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return err
	}

	if goal.Status == Completed && goal.CurrentAmount < goal.TargetAmount {
		return s.Repository.UpdateFields(ctx, goalID, map[string]interface{}{
			"status":     Active,
			"ended_at":   nil,
			"updated_at": time.Now(),
		})
	}

	return nil
}

func (s *Service) GetContributions(ctx context.Context, goalID, userID ulid.ULID) ([]*Contribution, error) {
	if err := s.CheckGoalBelongsToUser(ctx, goalID, userID); err != nil {
		return nil, err
	}

	return s.Repository.GetContributionsByGoalID(ctx, goalID, userID)
}

func (s *Service) DeleteContribution(ctx context.Context, contributionID, userID ulid.ULID) error {
	contribution, err := s.Repository.GetContributionByID(ctx, contributionID, userID)
	if err != nil {
		return err
	}

	if contribution.Type != ContributionDeposit {
		return appErrors.NewValidationError("contribution", "apenas contribuições podem ser removidas")
	}

	goal, err := s.GetGoalByID(ctx, contribution.GoalId, userID)
	if err != nil {
		return err
	}

	if goal.Status == Completed && goal.CurrentAmount-contribution.Amount < goal.TargetAmount {
		now := time.Now()
		if err := s.Repository.UpdateFields(ctx, contribution.GoalId, map[string]interface{}{
			"status":     Active,
			"ended_at":   nil,
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	if err := s.Repository.UpdateCurrentAmountAtomic(ctx, contribution.GoalId, -contribution.Amount); err != nil {
		return err
	}

	if err := s.AccountService.UpdateBalance(ctx, contribution.AccountId, userID, contribution.Amount); err != nil {
		_ = s.Repository.UpdateCurrentAmountAtomic(ctx, contribution.GoalId, contribution.Amount)
		return err
	}

	if contribution.TransactionId != nil && s.TransactionService != nil {
		if err := s.TransactionService.DeleteTransaction(ctx, *contribution.TransactionId, userID); err != nil {
			_ = s.Repository.UpdateCurrentAmountAtomic(ctx, contribution.GoalId, contribution.Amount)
			_ = s.AccountService.UpdateBalance(ctx, contribution.AccountId, userID, -contribution.Amount)
			return err
		}
	}

	return s.Repository.DeleteContribution(ctx, contributionID)
}

func (s *Service) DeleteContributionByTransactionId(ctx context.Context, transactionID, userID ulid.ULID) error {
	contribution, err := s.Repository.GetContributionByTransactionID(ctx, transactionID, userID)
	if err != nil {
		return err
	}
	if contribution == nil {
		return nil
	}

	if contribution.Type != ContributionDeposit {
		return nil
	}

	goal, err := s.GetGoalByID(ctx, contribution.GoalId, userID)
	if err != nil {
		return err
	}

	if goal.Status == Completed && goal.CurrentAmount-contribution.Amount < goal.TargetAmount {
		now := time.Now()
		if err := s.Repository.UpdateFields(ctx, contribution.GoalId, map[string]interface{}{
			"status":     Active,
			"ended_at":   nil,
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	if err := s.Repository.UpdateCurrentAmountAtomic(ctx, contribution.GoalId, -contribution.Amount); err != nil {
		return err
	}

	if err := s.AccountService.UpdateBalance(ctx, contribution.AccountId, userID, contribution.Amount); err != nil {
		_ = s.Repository.UpdateCurrentAmountAtomic(ctx, contribution.GoalId, contribution.Amount)
		return err
	}

	return s.Repository.DeleteContribution(ctx, contribution.Id)
}
func (s *Service) GetGoalProgress(ctx context.Context, goalID, userID ulid.ULID) (*GoalProgress, error) {
	goal, err := s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return nil, err
	}

	percentage := 0.0
	if goal.TargetAmount > 0 {
		percentage = (goal.CurrentAmount / goal.TargetAmount) * 100
	}

	remaining := goal.TargetAmount - goal.CurrentAmount
	if remaining < 0 {
		remaining = 0
	}

	return &GoalProgress{
		GoalId:        goalID,
		Name:          goal.Name,
		TargetAmount:  goal.TargetAmount,
		CurrentAmount: goal.CurrentAmount,
		Remaining:     remaining,
		Percentage:    percentage,
		Status:        string(goal.Status),
	}, nil
}

func (s *Service) UpdateGoal(ctx context.Context, request *domaincontracts.GoalUpdateRequest) error {
	if err := s.validateGoalUpdate(request); err != nil {
		return err
	}

	if err := s.CheckGoalBelongsToUser(ctx, request.Id, request.UserId); err != nil {
		return err
	}

	current, err := s.Repository.GetByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrGoalNotFound.WithError(err)
		}
		return appErrors.NewDatabaseError(err)
	}

	current.Name = request.Name
	current.TargetAmount = request.Target
	current.EndedAt = request.EndedAt
	current.UpdatedAt = time.Now()

	return s.Repository.Update(ctx, current)
}

func (s *Service) DeleteGoal(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) error {
	if err := s.CheckGoalBelongsToUser(ctx, goalID, userID); err != nil {
		return err
	}
	if err := s.Repository.Delete(ctx, goalID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrGoalNotFound.WithError(err)
		}
		return appErrors.NewDatabaseError(err)
	}
	return nil
}

func (s *Service) GetGoalByID(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (*Goal, error) {
	goal, err := s.Repository.GetByIDAndUser(ctx, goalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrGoalNotFound.WithError(err)
		}
		return nil, appErrors.NewDatabaseError(err)
	}
	return goal, nil
}

func (s *Service) GetGoalsByUserID(ctx context.Context, userID ulid.ULID, filters *GoalFilters, pagination *pkg.PaginationParams) ([]*Goal, int64, error) {
	return s.Repository.GetByUserID(ctx, userID, filters, pagination)
}

func (s *Service) ListGoals(ctx context.Context, pagination *pkg.PaginationParams) ([]*Goal, int64, error) {
	return s.Repository.List(ctx, pagination)
}

func (s *Service) CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) error {
	userBelongs, err := s.Repository.CheckGoalBelongsToUser(ctx, goalID, userID)
	if err != nil {
		return err
	}
	if !userBelongs {
		return appErrors.ErrResourceNotOwned
	}
	return nil
}

func (s *Service) validateGoalCreate(request *domaincontracts.GoalCreateRequest) error {
	if request.Name == "" {
		return appErrors.NewValidationError("name", "é obrigatório")
	}
	if request.Target <= 0 {
		return appErrors.NewValidationError("target", "deve ser maior que zero")
	}
	if request.EndedAt != nil && request.EndedAt.Before(time.Now()) {
		return appErrors.NewValidationError("ended_at", "deve ser uma data futura")
	}
	return nil
}

func (s *Service) validateGoalUpdate(request *domaincontracts.GoalUpdateRequest) error {
	if request.Name == "" {
		return appErrors.NewValidationError("name", "e obrigatorio")
	}
	if request.Target <= 0 {
		return appErrors.NewValidationError("target", "deve ser maior que zero")
	}
	if request.EndedAt != nil && request.EndedAt.Before(time.Now()) {
		return appErrors.NewValidationError("ended_at", "deve ser uma data futura")
	}
	return nil
}

func (s *Service) createGoalTransaction(ctx context.Context, goal *Goal, accountID, userID ulid.ULID, amount float64, description string) (*transaction.Transaction, error) {
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "Contribuição para meta: " + goal.Name
	}

	// Busca categoria de metas
	defaultCategories := category.GetDefaultCategoriesForUser(userID)
	var categoryID ulid.ULID
	for _, cat := range defaultCategories {
		if cat.Name == "Metas" || cat.Name == "Investimentos" {
			categoryID = cat.Id
			break
		}
	}
	if categoryID == (ulid.ULID{}) {
		categoryID = defaultCategories[0].Id
	}

	categoryIDPtr := &categoryID
	tx := &transaction.Transaction{
		Type:        transaction.Goals,
		UserId:      userID,
		AccountId:   accountID,
		CategoryId:  categoryIDPtr,
		Amount:      -amount,
		Description: desc,
		Date:        time.Now(),
	}
	transaction.TransactionCreateStruct(tx)

	if err := s.TransactionService.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *Service) rollbackContribution(ctx context.Context, transactionID *ulid.ULID, accountID, userID ulid.ULID, amount float64) {
	if transactionID != nil && s.TransactionService != nil {
		_ = s.TransactionService.DeleteTransaction(ctx, *transactionID, userID)
	}
	_ = s.AccountService.UpdateBalance(ctx, accountID, userID, amount)
}

func (s *Service) checkAndUpdateGoalStatus(ctx context.Context, goalID, userID ulid.ULID) error {
	goal, err := s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return err
	}

	if goal.CurrentAmount >= goal.TargetAmount {
		now := time.Now()
		return s.Repository.UpdateFields(ctx, goalID, map[string]interface{}{
			"status":     Completed,
			"ended_at":   &now,
			"updated_at": now,
		})
	}

	return nil
}

type GoalProgress struct {
	GoalId        ulid.ULID `json:"goalId"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"targetAmount"`
	CurrentAmount float64   `json:"currentAmount"`
	Remaining     float64   `json:"remaining"`
	Percentage    float64   `json:"percentage"`
	Status        string    `json:"status"`
}
