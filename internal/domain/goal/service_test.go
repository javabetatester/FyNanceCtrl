package goal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincontracts "Fynance/internal/domain/contracts"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type fakeGoalRepository struct {
	createFn                    func(ctx context.Context, goal *goal.Goal) error
	updateFn                    func(ctx context.Context, goal *goal.Goal) error
	deleteFn                    func(ctx context.Context, id ulid.ULID) error
	getByIDFn                   func(ctx context.Context, id ulid.ULID) (*goal.Goal, error)
	getByIdAndUserFn            func(ctx context.Context, id, userID ulid.ULID) (*goal.Goal, error)
	getByUserFn                 func(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error)
	listFn                      func(ctx context.Context, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error)
	checkGoalBelongsToUserFn    func(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error)
	updateFieldsFn              func(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error
	createContributionFn        func(ctx context.Context, contribution *goal.Contribution) error
	getContributionsFn          func(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*goal.Contribution, error)
	updateCurrentAmountFn       func(ctx context.Context, goalId ulid.ULID, amount float64) error
	updateCurrentAmountAtomicFn func(ctx context.Context, goalId ulid.ULID, delta float64) error
}

func (f *fakeGoalRepository) Create(ctx context.Context, g *goal.Goal) error {
	if f.createFn != nil {
		return f.createFn(ctx, g)
	}
	return nil
}

func (f *fakeGoalRepository) List(ctx context.Context, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	if f.listFn != nil {
		return f.listFn(ctx, pagination)
	}
	return nil, 0, nil
}

func (f *fakeGoalRepository) Update(ctx context.Context, g *goal.Goal) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, g)
	}
	return nil
}

func (f *fakeGoalRepository) UpdateFields(ctx context.Context, id ulid.ULID, fields map[string]interface{}) error {
	if f.updateFieldsFn != nil {
		return f.updateFieldsFn(ctx, id, fields)
	}
	return nil
}

func (f *fakeGoalRepository) Delete(ctx context.Context, id ulid.ULID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

func (f *fakeGoalRepository) GetById(ctx context.Context, id ulid.ULID) (*goal.Goal, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeGoalRepository) GetByUserId(ctx context.Context, userId ulid.ULID, pagination *pkg.PaginationParams) ([]*goal.Goal, int64, error) {
	if f.getByUserFn != nil {
		return f.getByUserFn(ctx, userId, pagination)
	}
	return nil, 0, nil
}

func (f *fakeGoalRepository) CheckGoalBelongsToUser(ctx context.Context, goalID ulid.ULID, userID ulid.ULID) (bool, error) {
	if f.checkGoalBelongsToUserFn != nil {
		return f.checkGoalBelongsToUserFn(ctx, goalID, userID)
	}
	return true, nil
}

func (f *fakeGoalRepository) CreateContribution(ctx context.Context, contribution *goal.Contribution) error {
	if f.createContributionFn != nil {
		return f.createContributionFn(ctx, contribution)
	}
	return nil
}

func (f *fakeGoalRepository) GetContributionsByGoalId(ctx context.Context, goalId ulid.ULID, userId ulid.ULID) ([]*goal.Contribution, error) {
	if f.getContributionsFn != nil {
		return f.getContributionsFn(ctx, goalId, userId)
	}
	return nil, nil
}

func (f *fakeGoalRepository) UpdateCurrentAmount(ctx context.Context, goalId ulid.ULID, amount float64) error {
	if f.updateCurrentAmountFn != nil {
		return f.updateCurrentAmountFn(ctx, goalId, amount)
	}
	return nil
}

func (f *fakeGoalRepository) GetByIdAndUser(ctx context.Context, id, userID ulid.ULID) (*goal.Goal, error) {
	if f.getByIdAndUserFn != nil {
		return f.getByIdAndUserFn(ctx, id, userID)
	}
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeGoalRepository) UpdateCurrentAmountAtomic(ctx context.Context, goalId ulid.ULID, delta float64) error {
	if f.updateCurrentAmountAtomicFn != nil {
		return f.updateCurrentAmountAtomicFn(ctx, goalId, delta)
	}
	if f.updateCurrentAmountFn != nil {
		g, err := f.GetById(ctx, goalId)
		if err != nil {
			return err
		}
		if g == nil {
			return nil
		}
		return f.updateCurrentAmountFn(ctx, goalId, g.CurrentAmount+delta)
	}
	return nil
}

type fakeUserRepository struct {
	getByIDFn func(ctx context.Context, id ulid.ULID) (*user.User, error)
}

func (f *fakeUserRepository) Create(ctx context.Context, _ *user.User) error { return nil }
func (f *fakeUserRepository) Update(ctx context.Context, _ *user.User) error { return nil }
func (f *fakeUserRepository) Delete(ctx context.Context, _ ulid.ULID) error { return nil }
func (f *fakeUserRepository) GetByEmail(ctx context.Context, _ string) (*user.User, error) {
	return nil, nil
}
func (f *fakeUserRepository) GetPlan(ctx context.Context, _ ulid.ULID) (user.Plan, error) {
	return "", nil
}
func (f *fakeUserRepository) GetById(ctx context.Context, id ulid.ULID) (*user.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return &user.User{Id: id}, nil
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   domaincontracts.GoalCreateRequest
		wantErr string
	}{
		{
			name: "missing name",
			input: domaincontracts.GoalCreateRequest{
				Target: 100,
			},
			wantErr: "VALIDATION_ERROR",
		},
		{
			name: "invalid target",
			input: domaincontracts.GoalCreateRequest{
				Name:   "Emergency fund",
				Target: 0,
			},
			wantErr: "VALIDATION_ERROR",
		},
		{
			name: "ended at in the past",
			input: domaincontracts.GoalCreateRequest{
				Name:   "Trip",
				Target: 500,
				EndedAt: func() *time.Time {
					past := time.Now().Add(-time.Hour)
					return &past
				}(),
			},
			wantErr: "VALIDATION_ERROR",
		},
		{
			name: "valid request",
			input: domaincontracts.GoalCreateRequest{
				Name:   "Retirement",
				Target: 1000,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := goal.Validate(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			appErr, ok := appErrors.AsAppError(err)
			if !ok {
				t.Fatalf("expected AppError, got %T", err)
			}
			if appErr.Code != tt.wantErr {
				t.Fatalf("expected code %s, got %s", tt.wantErr, appErr.Code)
			}
		})
	}
}

func TestServiceCreateGoal(t *testing.T) {
	t.Parallel()

	type fields struct {
		userRepo *fakeUserRepository
		goalRepo *fakeGoalRepository
	}

	successUserRepo := &fakeUserRepository{
		getByIDFn: func(ctx context.Context, id ulid.ULID) (*user.User, error) {
			return &user.User{Id: id}, nil
		},
	}

	makeService := func(f fields) goal.Service {
		return goal.Service{
			Repository: f.goalRepo,
			UserService: user.Service{
				Repository: f.userRepo,
			},
		}
	}

	ctx := context.Background()
	userID := ulid.Make()

	t.Run("fails when user not found", func(t *testing.T) {
		svc := makeService(fields{
			userRepo: &fakeUserRepository{
				getByIDFn: func(ctx context.Context, id ulid.ULID) (*user.User, error) {
					return nil, errors.New("not found")
				},
			},
			goalRepo: &fakeGoalRepository{},
		})

		err := svc.CreateGoal(ctx, &domaincontracts.GoalCreateRequest{
			UserId: userID,
			Name:   "New goal",
			Target: 100,
		})

		if err == nil {
			t.Fatalf("expected error")
		}
		appErr, ok := appErrors.AsAppError(err)
		if !ok {
			t.Fatalf("expected AppError, got %T", err)
		}
		if appErr.Code != appErrors.ErrUserNotFound.Code {
			t.Fatalf("expected code %s, got %s", appErrors.ErrUserNotFound.Code, appErr.Code)
		}
	})

	t.Run("creates goal successfully", func(t *testing.T) {
		var created goal.Goal
		svc := makeService(fields{
			userRepo: successUserRepo,
			goalRepo: &fakeGoalRepository{
				createFn: func(ctx context.Context, g *goal.Goal) error {
					created = *g
					return nil
				},
			},
		})

		err := svc.CreateGoal(ctx, &domaincontracts.GoalCreateRequest{
			UserId: userID,
			Name:   "Build emergency fund",
			Target: 2000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.Name != "Build emergency fund" {
			t.Fatalf("goal not saved as expected")
		}
		if created.UserId != userID {
			t.Fatalf("expected user id %s", userID)
		}
	})
}

func TestServiceUpdateGoalValidations(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	goalID := ulid.Make()

	svc := goal.Service{
		Repository: &fakeGoalRepository{
			checkGoalBelongsToUserFn: func(ctx context.Context, gID ulid.ULID, uID ulid.ULID) (bool, error) {
				return gID == goalID && uID == userID, nil
			},
			getByIDFn: func(ctx context.Context, id ulid.ULID) (*goal.Goal, error) {
				return &goal.Goal{
					Id:           id,
					UserId:       userID,
					Name:         "Original",
					TargetAmount: 100,
				}, nil
			},
			updateFn: func(ctx context.Context, g *goal.Goal) error {
				return nil
			},
		},
		UserService: user.Service{
			Repository: &fakeUserRepository{},
		},
	}

	tests := []struct {
		name    string
		request domaincontracts.GoalUpdateRequest
		wantErr string
	}{
		{
			name: "invalid target",
			request: domaincontracts.GoalUpdateRequest{
				Id:     goalID,
				UserId: userID,
				Name:   "Updated",
				Target: 0,
			},
			wantErr: "VALIDATION_ERROR",
		},
		{
			name: "belongs to another user",
			request: domaincontracts.GoalUpdateRequest{
				Id:     goalID,
				UserId: ulid.Make(),
				Name:   "Updated",
				Target: 100,
			},
			wantErr: appErrors.ErrResourceNotOwned.Code,
		},
		{
			name: "success",
			request: domaincontracts.GoalUpdateRequest{
				Id:     goalID,
				UserId: userID,
				Name:   "Updated",
				Target: 500,
			},
			wantErr: "",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateGoal(ctx, &tt.request)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
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
			if appErr.Code != tt.wantErr {
				t.Fatalf("expected code %s, got %s", tt.wantErr, appErr.Code)
			}
		})
	}
}
