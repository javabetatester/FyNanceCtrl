package budget

import (
	"context"
	"errors"
	"time"

	"Fynance/internal/domain/category"
	"Fynance/internal/domain/shared"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/logger"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type Service struct {
	Repository      Repository
	CategoryService *category.Service
	shared.BaseService
}

var _ shared.BudgetUpdater = (*Service)(nil)

func NewService(repo Repository, categoryService *category.Service, userChecker *shared.UserCheckerService) *Service {
	return &Service{
		Repository:      repo,
		CategoryService: categoryService,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*Budget, error) {
	if err := s.EnsureUserExists(ctx, req.UserId); err != nil {
		return nil, err
	}

	categoryID, err := s.resolveCategoryID(ctx, req.CategoryId, req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	existing, _ := s.Repository.GetByCategoryID(ctx, categoryID, req.UserId, req.Month, req.Year)
	if existing != nil {
		return nil, appErrors.NewConflictError("orcamento para esta categoria neste periodo")
	}

	now := time.Now()
	budget := &Budget{
		Id:          pkg.GenerateULIDObject(),
		UserId:      req.UserId,
		CategoryId:  categoryID,
		Amount:      req.Amount,
		Spent:       0,
		Month:       req.Month,
		Year:        req.Year,
		AlertAt:     s.normalizeAlertAt(req.AlertAt),
		IsRecurring: req.IsRecurring,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.Repository.Create(ctx, budget); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	s.loadCategoryName(ctx, budget)

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
	result := s.Repository.Delete(ctx, budgetID, userID)
	if result != nil {
		if errors.Is(result, gorm.ErrRecordNotFound) {
			return appErrors.ErrNotFound
		}
		return appErrors.NewDatabaseError(result)
	}
	return nil
}

func (s *Service) GetBudgetByID(ctx context.Context, budgetID, userID ulid.ULID) (*Budget, error) {
	budget, err := s.Repository.GetByID(ctx, budgetID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if budget.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return budget, nil
}

func (s *Service) ListBudgets(ctx context.Context, userID ulid.ULID, month, year int, filters *BudgetFilters, pagination *pkg.PaginationParams) ([]*Budget, int64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetByUserID(ctx, userID, month, year, filters, pagination)
}

func (s *Service) GetBudgetSummary(ctx context.Context, userID ulid.ULID, month, year int) (*BudgetSummary, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	month, year = s.normalizePeriod(month, year)

	return s.Repository.GetSummary(ctx, userID, month, year)
}

func (s *Service) UpdateSpent(ctx context.Context, categoryID, userID ulid.ULID, amount float64) error {
	return s.UpdateSpentWithDate(ctx, categoryID, userID, amount, time.Now())
}

func (s *Service) UpdateSpentWithDate(ctx context.Context, categoryID, userID ulid.ULID, amount float64, transactionDate time.Time) error {
	budget, err := s.Repository.GetByCategoryID(ctx, categoryID, userID, int(transactionDate.Month()), transactionDate.Year())

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resolvedID, resolveErr := s.resolveCategoryID(ctx, categoryID, userID)
			if resolveErr == nil && resolvedID != categoryID {
				budget, err = s.Repository.GetByCategoryID(ctx, resolvedID, userID, int(transactionDate.Month()), transactionDate.Year())
				if err == nil && budget != nil {
					return s.Repository.UpdateSpent(ctx, budget.Id, amount)
				}
			}

			logger.Debug().
				Str("category_id", categoryID.String()).
				Str("user_id", userID.String()).
				Int("month", int(transactionDate.Month())).
				Int("year", transactionDate.Year()).
				Msg("budget not found for category, skipping update")
		} else {
			logger.Error().
				Err(err).
				Str("category_id", categoryID.String()).
				Str("user_id", userID.String()).
				Msg("error getting budget for category")
		}
		return nil
	}

	return s.Repository.UpdateSpent(ctx, budget.Id, amount)
}

func (s *Service) GetBudgetStatus(ctx context.Context, budgetID, userID ulid.ULID) (*BudgetStatusResponse, error) {
	budget, err := s.GetBudgetByID(ctx, budgetID, userID)
	if err != nil {
		return nil, err
	}

	return s.calculateBudgetStatus(budget), nil
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
		existing, _ := s.Repository.GetByCategoryID(ctx, budget.CategoryId, userID, currentMonth, currentYear)
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

func (s *Service) resolveCategoryID(ctx context.Context, categoryID, userID ulid.ULID) (ulid.ULID, error) {
	if s.CategoryService == nil {
		return categoryID, nil
	}

	if err := s.CategoryService.ValidateAndEnsureExists(ctx, categoryID, userID); err != nil {
		return ulid.ULID{}, err
	}

	return s.CategoryService.ResolveCategoryID(ctx, categoryID, userID)
}

func (s *Service) loadCategoryName(ctx context.Context, budget *Budget) {
	if s.CategoryService == nil {
		return
	}

	cat, err := s.CategoryService.GetByID(ctx, budget.CategoryId, budget.UserId)
	if err == nil && cat != nil {
		budget.CategoryName = cat.Name
	}
}

func (s *Service) validateCreateRequest(req *CreateBudgetRequest) error {
	if req.Amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	if req.Month < 1 || req.Month > 12 {
		return appErrors.NewValidationError("month", "deve estar entre 1 e 12")
	}

	if req.Year < 2000 || req.Year > 2100 {
		return appErrors.NewValidationError("year", "ano invalido")
	}

	return nil
}

func (s *Service) normalizeAlertAt(alertAt float64) float64 {
	if alertAt <= 0 {
		return 80
	}
	return alertAt
}

func (s *Service) normalizePeriod(month, year int) (int, int) {
	now := time.Now()
	if month <= 0 || month > 12 {
		month = int(now.Month())
	}
	if year <= 0 {
		year = now.Year()
	}
	return month, year
}

func (s *Service) calculateBudgetStatus(budget *Budget) *BudgetStatusResponse {
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
	}
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
