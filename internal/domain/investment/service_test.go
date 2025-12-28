package investment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"Fynance/internal/domain/account"
	domaincontracts "Fynance/internal/domain/contracts"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type fakeInvestmentRepository struct {
	createFn              func(ctx context.Context, inv *investment.Investment) error
	updateFn              func(ctx context.Context, inv *investment.Investment) error
	deleteFn              func(ctx context.Context, id ulid.ULID, userId ulid.ULID) error
	getByIDFn             func(ctx context.Context, id ulid.ULID, userId ulid.ULID) (*investment.Investment, error)
	getByUserFn           func(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error)
	listFn                func(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error)
	getTotalBalanceFn     func(ctx context.Context, userId ulid.ULID) (float64, error)
	getByTypeFn           func(ctx context.Context, userId ulid.ULID, typ investment.Types, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error)
	updateBalanceAtomicFn func(ctx context.Context, investmentID ulid.ULID, delta float64) error
}

func (f *fakeInvestmentRepository) Create(ctx context.Context, inv *investment.Investment) error {
	if f.createFn != nil {
		return f.createFn(ctx, inv)
	}
	return nil
}

func (f *fakeInvestmentRepository) List(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if f.listFn != nil {
		return f.listFn(ctx, userId, pagination)
	}
	return nil, 0, nil
}

func (f *fakeInvestmentRepository) Update(ctx context.Context, inv *investment.Investment) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, inv)
	}
	return nil
}

func (f *fakeInvestmentRepository) Delete(ctx context.Context, id ulid.ULID, userId ulid.ULID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id, userId)
	}
	return nil
}

func (f *fakeInvestmentRepository) GetInvestmentById(ctx context.Context, id ulid.ULID, userId ulid.ULID) (*investment.Investment, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id, userId)
	}
	return nil, nil
}

func (f *fakeInvestmentRepository) GetByUserId(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if f.getByUserFn != nil {
		return f.getByUserFn(ctx, userId, pagination)
	}
	return nil, 0, nil
}

func (f *fakeInvestmentRepository) GetTotalBalance(ctx context.Context, userId ulid.ULID) (float64, error) {
	if f.getTotalBalanceFn != nil {
		return f.getTotalBalanceFn(ctx, userId)
	}
	return 0, nil
}

func (f *fakeInvestmentRepository) GetByType(ctx context.Context, userId ulid.ULID, typ investment.Types, pagination *pkg.PaginationParams) ([]*investment.Investment, int64, error) {
	if f.getByTypeFn != nil {
		return f.getByTypeFn(ctx, userId, typ, pagination)
	}
	return nil, 0, nil
}

func (f *fakeInvestmentRepository) UpdateBalanceAtomic(ctx context.Context, investmentID ulid.ULID, delta float64) error {
	if f.updateBalanceAtomicFn != nil {
		return f.updateBalanceAtomicFn(ctx, investmentID, delta)
	}
	if f.updateFn != nil && f.getByIDFn != nil {
		inv, err := f.getByIDFn(ctx, investmentID, ulid.ULID{})
		if err != nil || inv == nil {
			return nil
		}
		inv.CurrentBalance += delta
		return f.updateFn(ctx, inv)
	}
	return nil
}

type fakeTransactionRepository struct {
	createFn func(ctx context.Context, tx *transaction.Transaction) error
}

func (f *fakeTransactionRepository) Create(ctx context.Context, tx *transaction.Transaction) error {
	if f.createFn != nil {
		return f.createFn(ctx, tx)
	}
	return nil
}

func (f *fakeTransactionRepository) Update(ctx context.Context, tx *transaction.Transaction) error {
	return nil
}
func (f *fakeTransactionRepository) Delete(ctx context.Context, id ulid.ULID) error { return nil }
func (f *fakeTransactionRepository) GetByID(ctx context.Context, id ulid.ULID) (*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepository) GetByIDAndUser(ctx context.Context, transactionID, userID ulid.ULID) (*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeTransactionRepository) GetAll(ctx context.Context, userId ulid.ULID, accountID *ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	return nil, 0, nil
}
func (f *fakeTransactionRepository) GetByAmount(ctx context.Context, amount float64, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	return nil, 0, nil
}
func (f *fakeTransactionRepository) GetByName(ctx context.Context, name string, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	return nil, 0, nil
}
func (f *fakeTransactionRepository) GetByCategory(ctx context.Context, categoryID ulid.ULID, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	return nil, 0, nil
}
func (f *fakeTransactionRepository) GetByInvestmentId(ctx context.Context, investmentID ulid.ULID, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*transaction.Transaction, int64, error) {
	return nil, 0, nil
}
func (f *fakeTransactionRepository) GetNumberOfTransactions(ctx context.Context, userId ulid.ULID) (int64, error) {
	return 0, nil
}

type fakeUserRepo struct {
	getByIDFn func(ctx context.Context, id ulid.ULID) (*user.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, _ *user.User) error               { return nil }
func (f *fakeUserRepo) Update(ctx context.Context, _ *user.User) error               { return nil }
func (f *fakeUserRepo) Delete(ctx context.Context, _ ulid.ULID) error { return nil }
func (f *fakeUserRepo) GetByEmail(ctx context.Context, _ string) (*user.User, error) { return nil, nil }
func (f *fakeUserRepo) GetPlan(ctx context.Context, _ ulid.ULID) (user.Plan, error)  { return "", nil }
func (f *fakeUserRepo) GetById(ctx context.Context, id ulid.ULID) (*user.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return &user.User{Id: id}, nil
}

type fakeAccountService struct {
	getByIDFn       func(ctx context.Context, accountID, userID ulid.ULID) (*account.Account, error)
	updateBalanceFn func(ctx context.Context, accountID, userID ulid.ULID, amount float64) error
}

func (f *fakeAccountService) GetAccountByID(ctx context.Context, accountID, userID ulid.ULID) (*account.Account, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, accountID, userID)
	}
	return &account.Account{
		Id:      accountID,
		UserId:  userID,
		Type:    account.TypeChecking,
		Balance: 1000,
	}, nil
}

func (f *fakeAccountService) UpdateBalance(ctx context.Context, accountID, userID ulid.ULID, amount float64) error {
	if f.updateBalanceFn != nil {
		return f.updateBalanceFn(ctx, accountID, userID, amount)
	}
	return nil
}

func TestServiceMakeContributionValidations(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	investmentID := ulid.Make()

	baseInvestment := &investment.Investment{
		Id:             investmentID,
		UserId:         userID,
		Name:           "Stocks",
		CurrentBalance: 100,
	}

	tests := []struct {
		name        string
		amount      float64
		getByIDErr  error
		wantErrCode string
	}{
		{
			name:        "invalid amount",
			amount:      0,
			wantErrCode: "VALIDATION_ERROR",
		},
		{
			name:        "investment not found",
			amount:      50,
			getByIDErr:  appErrors.ErrInvestmentNotFound,
			wantErrCode: appErrors.ErrInvestmentNotFound.Code,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeInvestmentRepository{
				getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
					if tt.getByIDErr != nil {
						return nil, tt.getByIDErr
					}
					return baseInvestment, nil
				},
			}

			accountID := ulid.Make()
			svc := investment.Service{
				Repository:      repo,
				TransactionRepo: &fakeTransactionRepository{},
				UserService: &user.Service{
					Repository: &fakeUserRepo{},
				},
				AccountService: &fakeAccountService{},
			}

			err := svc.MakeContribution(ctx, investmentID, accountID, userID, tt.amount, "aporte")
			if tt.wantErrCode == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error")
			}
			appErr, ok := appErrors.AsAppError(err)
			if !ok {
				t.Fatalf("expected AppError, got %T", err)
			}
			if appErr.Code != tt.wantErrCode {
				t.Fatalf("expected code %s, got %s", tt.wantErrCode, appErr.Code)
			}
		})
	}

	t.Run("success updates balance", func(t *testing.T) {
		var updated *investment.Investment
		repo := &fakeInvestmentRepository{
			getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
				copy := *baseInvestment
				return &copy, nil
			},
			updateFn: func(ctx context.Context, inv *investment.Investment) error {
				updated = inv
				return nil
			},
		}

		svc := investment.Service{
			Repository:      repo,
			TransactionRepo: &fakeTransactionRepository{},
			UserService: &user.Service{
				Repository: &fakeUserRepo{},
			},
		}

		accountID := ulid.Make()
		err := svc.MakeContribution(ctx, investmentID, accountID, userID, 50, "aporte")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated == nil || updated.CurrentBalance != 150 {
			t.Fatalf("expected balance 150, got %+v", updated)
		}
	})
}

func TestServiceMakeWithdrawValidations(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	investmentID := ulid.Make()

	ctx := context.Background()

	svc := investment.Service{
		Repository: &fakeInvestmentRepository{
			getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
				return &investment.Investment{
					Id:             id,
					UserId:         uid,
					Name:           "Stocks",
					CurrentBalance: 100,
				}, nil
			},
			updateFn: func(ctx context.Context, inv *investment.Investment) error { return nil },
		},
		TransactionRepo: &fakeTransactionRepository{},
		UserService: &user.Service{
			Repository: &fakeUserRepo{},
		},
		AccountService: &fakeAccountService{},
	}

	accountID := ulid.Make()

	t.Run("amount must be positive", func(t *testing.T) {
		err := svc.MakeWithdraw(ctx, investmentID, accountID, userID, 0, "resgate")
		if err == nil {
			t.Fatalf("expected error")
		}
		appErr, _ := appErrors.AsAppError(err)
		if appErr.Code != "VALIDATION_ERROR" {
			t.Fatalf("expected validation error, got %s", appErr.Code)
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		err := svc.MakeWithdraw(ctx, investmentID, accountID, userID, 200, "resgate")
		if err == nil {
			t.Fatalf("expected error")
		}
		appErr, _ := appErrors.AsAppError(err)
		if appErr.Code != "VALIDATION_ERROR" {
			t.Fatalf("expected validation error, got %s", appErr.Code)
		}
	})
}

func TestServiceDeleteInvestment(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	investmentID := ulid.Make()

	ctx := context.Background()

	svc := investment.Service{
		Repository: &fakeInvestmentRepository{
			getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
				return &investment.Investment{
					Id:             id,
					UserId:         uid,
					Name:           "Stocks",
					CurrentBalance: 10,
				}, nil
			},
			deleteFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) error {
				return nil
			},
		},
		TransactionRepo: &fakeTransactionRepository{},
		UserService: &user.Service{
			Repository: &fakeUserRepo{},
		},
	}

	err := svc.DeleteInvestment(ctx, investmentID, userID)
	if err == nil {
		t.Fatalf("expected error when balance > 0")
	}
	appErr, _ := appErrors.AsAppError(err)
	if appErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error, got %s", appErr.Code)
	}

	// adjust repo to zero balance and ensure deletion succeeds
	svc.Repository = &fakeInvestmentRepository{
		getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
			return &investment.Investment{
				Id:             id,
				UserId:         uid,
				CurrentBalance: 0,
			}, nil
		},
		deleteFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) error {
			return nil
		},
	}

	if err := svc.DeleteInvestment(ctx, investmentID, userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpdateInvestment(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	investmentID := ulid.Make()

	updateCalled := false
	repo := &fakeInvestmentRepository{
		getByIDFn: func(ctx context.Context, id ulid.ULID, uid ulid.ULID) (*investment.Investment, error) {
			return &investment.Investment{
				Id:             id,
				UserId:         uid,
				Name:           "Original",
				Type:           investment.Types("stocks"),
				CurrentBalance: 100,
				UpdatedAt:      time.Now(),
			}, nil
		},
		updateFn: func(ctx context.Context, inv *investment.Investment) error {
			updateCalled = true
			if inv.Name != "Updated" {
				return errors.New("name not updated")
			}
			return nil
		},
	}

	svc := investment.Service{
		Repository:      repo,
		TransactionRepo: &fakeTransactionRepository{},
		UserService: &user.Service{
			Repository: &fakeUserRepo{},
		},
	}

	req := domaincontracts.UpdateInvestmentRequest{
		UserId: userID,
		Id:     investmentID,
	}

	newName := " Updated "
	req.Name = &newName

	if err := svc.UpdateInvestment(context.Background(), investmentID, userID, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Fatalf("expected update to be called")
	}
}
