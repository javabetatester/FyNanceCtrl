package account

import (
	"context"
	"strings"
	"time"

	"Fynance/internal/domain/shared"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository Repository
	shared.BaseService
}

func NewService(repo Repository, userChecker *shared.UserCheckerService) *Service {
	return &Service{
		Repository: repo,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	if err := s.EnsureUserExists(ctx, req.UserId); err != nil {
		return nil, err
	}

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now()
	account := &Account{
		Id:             pkg.GenerateULIDObject(),
		UserId:         req.UserId,
		Name:           strings.TrimSpace(req.Name),
		Type:           req.Type,
		Balance:        req.InitialBalance,
		Color:          req.Color,
		Icon:           req.Icon,
		IncludeInTotal: req.IncludeInTotal,
		IsActive:       true,
		CreditCardId:   req.CreditCardId,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.Repository.Create(ctx, account); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return account, nil
}

func (s *Service) UpdateAccount(ctx context.Context, accountID, userID ulid.ULID, req *UpdateAccountRequest) error {
	account, err := s.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return appErrors.NewValidationError("name", "nao pode ser vazio")
		}
		account.Name = name
	}

	if req.Type != nil {
		if !req.Type.IsValid() {
			return appErrors.NewValidationError("type", "tipo de conta invalido")
		}
		account.Type = *req.Type
	}

	if req.Color != nil {
		account.Color = *req.Color
	}

	if req.Icon != nil {
		account.Icon = *req.Icon
	}

	if req.IncludeInTotal != nil {
		account.IncludeInTotal = *req.IncludeInTotal
	}

	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	account.UpdatedAt = time.Now()

	return s.Repository.Update(ctx, account)
}

func (s *Service) DeleteAccount(ctx context.Context, accountID, userID ulid.ULID) error {
	account, err := s.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if account.Balance != 0 {
		return appErrors.NewValidationError("account", "Conta possui saldo, não pode remover")
	}

	return s.Repository.Delete(ctx, accountID, userID)
}

func (s *Service) GetAccountByID(ctx context.Context, accountID, userID ulid.ULID) (*Account, error) {
	account, err := s.Repository.GetById(ctx, accountID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if account.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return account, nil
}

func (s *Service) ListAccounts(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Account, int64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetByUserId(ctx, userID, pagination)
}

func (s *Service) ListActiveAccounts(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Account, int64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetActiveByUserId(ctx, userID, pagination)
}

func (s *Service) UpdateBalance(ctx context.Context, accountID, userID ulid.ULID, amount float64) error {
	account, err := s.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	newBalance := account.Balance + amount
	if account.Type != TypeCreditCard && newBalance < 0 {
		return appErrors.NewValidationError("amount", "Saldo insuficiente")
	}

	return s.Repository.UpdateBalance(ctx, accountID, amount)
}

func (s *Service) GetTotalBalance(ctx context.Context, userID ulid.ULID) (float64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return 0, err
	}

	return s.Repository.GetTotalBalance(ctx, userID)
}

func (s *Service) Transfer(ctx context.Context, fromAccountID, toAccountID, userID ulid.ULID, amount float64) error {
	if amount <= 0 {
		return appErrors.NewValidationError("amount", "Valor deve ser maior que zero")
	}

	fromAccount, err := s.GetAccountByID(ctx, fromAccountID, userID)
	if err != nil {
		return err
	}

	if _, err := s.GetAccountByID(ctx, toAccountID, userID); err != nil {
		return err
	}

	if fromAccount.Type != TypeCreditCard && fromAccount.Balance < amount {
		return appErrors.NewValidationError("amount", "Saldo insuficiente na conta de origem")
	}

	tx, err := s.Repository.BeginTx(ctx)
	if err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if err := s.Repository.UpdateBalanceWithTx(ctx, tx, fromAccountID, -amount); err != nil {
		_ = s.Repository.RollbackTx(tx)
		return appErrors.NewDatabaseError(err)
	}

	if err := s.Repository.UpdateBalanceWithTx(ctx, tx, toAccountID, amount); err != nil {
		_ = s.Repository.RollbackTx(tx)
		return appErrors.NewDatabaseError(err)
	}

	if err := s.Repository.CommitTx(tx); err != nil {
		_ = s.Repository.RollbackTx(tx)
		return appErrors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) validateCreateRequest(req *CreateAccountRequest) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return appErrors.NewValidationError("name", "e obrigatorio")
	}

	if !req.Type.IsValid() {
		return appErrors.NewValidationError("type", "tipo de conta invalido")
	}

	return nil
}

type CreateAccountRequest struct {
	UserId         ulid.ULID
	Name           string
	Type           AccountType
	InitialBalance float64
	Color          string
	Icon           string
	IncludeInTotal bool
	CreditCardId   *ulid.ULID
}

type UpdateAccountRequest struct {
	Name           *string
	Type           *AccountType
	Color          *string
	Icon           *string
	IncludeInTotal *bool
	IsActive       *bool
}
