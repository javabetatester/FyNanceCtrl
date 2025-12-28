package goal

import (
	"context"
	"strings"
	"time"

	"Fynance/internal/domain/account"
	domaincontracts "Fynance/internal/domain/contracts"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository     Repository
	UserService    user.Service
	AccountService *account.Service
}

func (s *Service) CreateGoal(ctx context.Context, request *domaincontracts.GoalCreateRequest) error {
	if err := Validate(*request); err != nil {
		return err
	}

	if _, err := s.UserService.GetByID(ctx, request.UserId); err != nil {
		return appErrors.ErrUserNotFound.WithError(err)
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

	contribution := &Contribution{
		Id:          pkg.GenerateULIDObject(),
		GoalId:      goalID,
		UserId:      userID,
		Type:        ContributionDeposit,
		Amount:      amount,
		Description: strings.TrimSpace(description),
		CreatedAt:   time.Now(),
	}

	if err := s.Repository.CreateContribution(ctx, contribution); err != nil {
		_ = s.AccountService.UpdateBalance(ctx, accountID, userID, amount)
		return err
	}

	if err := s.Repository.UpdateCurrentAmountAtomic(ctx, goalID, amount); err != nil {
		_ = s.AccountService.UpdateBalance(ctx, accountID, userID, amount)
		return err
	}

	goal, err = s.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return err
	}
	newAmount := goal.CurrentAmount

	if newAmount >= goal.TargetAmount {
		now := time.Now()
		return s.Repository.UpdateFields(ctx, goalID, map[string]interface{}{
			"status":     Completed,
			"ended_at":   &now,
			"updated_at": now,
		})
	}

	return nil
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
	newAmount := goal.CurrentAmount

	if goal.Status == Completed && newAmount < goal.TargetAmount {
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

	return s.Repository.GetContributionsByGoalId(ctx, goalID, userID)
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

type GoalProgress struct {
	GoalId        ulid.ULID `json:"goalId"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"targetAmount"`
	CurrentAmount float64   `json:"currentAmount"`
	Remaining     float64   `json:"remaining"`
	Percentage    float64   `json:"percentage"`
	Status        string    `json:"status"`
}

func (s *Service) UpdateGoal(ctx context.Context, request *domaincontracts.GoalUpdateRequest) error {
	if err := ValidateUpdateGoal(*request); err != nil {
		return err
	}

	if err := s.CheckGoalBelongsToUser(ctx, request.Id, request.UserId); err != nil {
		return err
	}

	current, err := s.Repository.GetById(ctx, request.Id)
	if err != nil {
		return err
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
	return s.Repository.Delete(ctx, goalID)
}

func (s *Service) GetGoalByID(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (*Goal, error) {
	goal, err := s.Repository.GetByIdAndUser(ctx, goalID, userID)
	if err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *Service) GetGoalsByUserID(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Goal, int64, error) {
	return s.Repository.GetByUserId(ctx, userID, pagination)
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

func Validate(request domaincontracts.GoalCreateRequest) error {
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

func ValidateUpdateGoal(request domaincontracts.GoalUpdateRequest) error {
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
