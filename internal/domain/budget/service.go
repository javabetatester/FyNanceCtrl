package budget

import (
	"context"
	"time"

	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository         Repository
	CategoryRepository transaction.CategoryRepository
	UserService        *user.Service
}

func (s *Service) CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*Budget, error) {
	if err := s.ensureUserExists(ctx, req.UserId); err != nil {
		return nil, err
	}

	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	existing, _ := s.Repository.GetByCategoryId(ctx, req.CategoryId, req.UserId, req.Month, req.Year)
	if existing != nil {
		return nil, appErrors.NewConflictError("orcamento para esta categoria neste periodo")
	}

	now := time.Now()
	budget := &Budget{
		Id:          pkg.GenerateULIDObject(),
		UserId:      req.UserId,
		CategoryId:  req.CategoryId,
		Amount:      req.Amount,
		Spent:       0,
		Month:       req.Month,
		Year:        req.Year,
		AlertAt:     req.AlertAt,
		IsRecurring: req.IsRecurring,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if budget.AlertAt <= 0 {
		budget.AlertAt = 80
	}

	if err := s.Repository.Create(ctx, budget); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return budget, nil
}

func (s *Service) UpdateBudget(ctx context.Context, budgetID, userID ulid.ULID, req *UpdateBudgetRequest) error {
	budget, err := s.GetBudgetByID(ctx, budgetID, userID)
	if err != nil {
		return err
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return appErrors.NewValidationError("amount", "deve ser maior que zero")
		}
		budget.Amount = *req.Amount
	}

	if req.AlertAt != nil {
		if *req.AlertAt < 0 || *req.AlertAt > 100 {
			return appErrors.NewValidationError("alert_at", "deve estar entre 0 e 100")
		}
		budget.AlertAt = *req.AlertAt
	}

	if req.IsRecurring != nil {
		budget.IsRecurring = *req.IsRecurring
	}

	budget.UpdatedAt = time.Now()

	return s.Repository.Update(ctx, budget)
}

func (s *Service) DeleteBudget(ctx context.Context, budgetID, userID ulid.ULID) error {
	_, err := s.GetBudgetByID(ctx, budgetID, userID)
	if err != nil {
		return err
	}

	return s.Repository.Delete(ctx, budgetID, userID)
}

func (s *Service) GetBudgetByID(ctx context.Context, budgetID, userID ulid.ULID) (*Budget, error) {
	budget, err := s.Repository.GetById(ctx, budgetID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if budget.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return budget, nil
}

func (s *Service) ListBudgets(ctx context.Context, userID ulid.ULID, month, year int, pagination *pkg.PaginationParams) ([]*Budget, int64, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	if month <= 0 || month > 12 {
		month = int(time.Now().Month())
	}
	if year <= 0 {
		year = time.Now().Year()
	}

	return s.Repository.GetByUserId(ctx, userID, month, year, pagination)
}

func (s *Service) GetBudgetSummary(ctx context.Context, userID ulid.ULID, month, year int) (*BudgetSummary, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	if month <= 0 || month > 12 {
		month = int(time.Now().Month())
	}
	if year <= 0 {
		year = time.Now().Year()
	}

	return s.Repository.GetSummary(ctx, userID, month, year)
}

func (s *Service) UpdateSpent(ctx context.Context, categoryID, userID ulid.ULID, amount float64) error {
	now := time.Now()
	budget, err := s.Repository.GetByCategoryId(ctx, categoryID, userID, int(now.Month()), now.Year())
	if err != nil {
		return nil
	}

	return s.Repository.UpdateSpent(ctx, budget.Id, amount)
}

func (s *Service) GetBudgetStatus(ctx context.Context, budgetID, userID ulid.ULID) (*BudgetStatusResponse, error) {
	budget, err := s.GetBudgetByID(ctx, budgetID, userID)
	if err != nil {
		return nil, err
	}

	remaining := budget.Amount - budget.Spent
	percentage := 0.0
	if budget.Amount > 0 {
		percentage = (budget.Spent / budget.Amount) * 100
	}

	status := "OK"
	if percentage >= 100 {
		status = "EXCEEDED"
	} else if percentage >= budget.AlertAt {
		status = "WARNING"
	}

	return &BudgetStatusResponse{
		BudgetId:   budget.Id,
		Amount:     budget.Amount,
		Spent:      budget.Spent,
		Remaining:  remaining,
		Percentage: percentage,
		Status:     status,
		AlertAt:    budget.AlertAt,
	}, nil
}

func (s *Service) CreateRecurringBudgets(ctx context.Context, userID ulid.ULID) error {
	recurring, _, err := s.Repository.GetRecurring(ctx, userID, nil)
	if err != nil {
		return err
	}

	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	for _, budget := range recurring {
		existing, _ := s.Repository.GetByCategoryId(ctx, budget.CategoryId, userID, currentMonth, currentYear)
		if existing != nil {
			continue
		}

		newBudget := &Budget{
			Id:          pkg.GenerateULIDObject(),
			UserId:      userID,
			CategoryId:  budget.CategoryId,
			Amount:      budget.Amount,
			Spent:       0,
			Month:       currentMonth,
			Year:        currentYear,
			AlertAt:     budget.AlertAt,
			IsRecurring: true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		_ = s.Repository.Create(ctx, newBudget)
	}

	return nil
}

func (s *Service) validateCreateRequest(ctx context.Context, req *CreateBudgetRequest) error {
	if req.Amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	if req.Month < 1 || req.Month > 12 {
		return appErrors.NewValidationError("month", "deve estar entre 1 e 12")
	}

	if req.Year < 2000 || req.Year > 2100 {
		return appErrors.NewValidationError("year", "ano invalido")
	}

	_, err := s.CategoryRepository.GetByID(ctx, req.CategoryId, req.UserId)
	if err != nil {
		return appErrors.ErrCategoryNotFound
	}

	return nil
}

func (s *Service) ensureUserExists(ctx context.Context, userID ulid.ULID) error {
	if s.UserService == nil {
		return appErrors.ErrInternalServer
	}

	_, err := s.UserService.GetByID(ctx, userID)
	if err != nil {
		return appErrors.ErrUserNotFound.WithError(err)
	}

	return nil
}

type CreateBudgetRequest struct {
	UserId      ulid.ULID
	CategoryId  ulid.ULID
	Amount      float64
	Month       int
	Year        int
	AlertAt     float64
	IsRecurring bool
}

type UpdateBudgetRequest struct {
	Amount      *float64
	AlertAt     *float64
	IsRecurring *bool
}

type BudgetStatusResponse struct {
	BudgetId   ulid.ULID `json:"budgetId"`
	Amount     float64   `json:"amount"`
	Spent      float64   `json:"spent"`
	Remaining  float64   `json:"remaining"`
	Percentage float64   `json:"percentage"`
	Status     string    `json:"status"`
	AlertAt    float64   `json:"alertAt"`
}
