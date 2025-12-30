package recurring

import (
	"context"
	"strings"
	"time"

	"Fynance/internal/domain/category"
	"Fynance/internal/domain/shared"
	"Fynance/internal/domain/transaction"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository         RecurringRepository
	TransactionRepo    transaction.TransactionRepository
	CategoryService    *category.Service
	TransactionService transaction.TransactionHandler
	shared.BaseService
}

func NewService(
	repo RecurringRepository,
	transactionRepo transaction.TransactionRepository,
	categoryService *category.Service,
	transactionService transaction.TransactionHandler,
	userChecker *shared.UserCheckerService,
) *Service {
	return &Service{
		Repository:         repo,
		TransactionRepo:    transactionRepo,
		CategoryService:    categoryService,
		TransactionService: transactionService,
		BaseService: shared.BaseService{
			UserChecker: userChecker,
		},
	}
}

func (s *Service) CreateRecurring(ctx context.Context, req *CreateRecurringRequest) (*RecurringTransaction, error) {
	if err := s.EnsureUserExists(ctx, req.UserId); err != nil {
		return nil, err
	}

	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	now := time.Now()
	nextDue := s.calculateNextDue(req.StartDate, req.Frequency, req.DayOfMonth, req.DayOfWeek)

	recurring := &RecurringTransaction{
		Id:          pkg.GenerateULIDObject(),
		UserId:      req.UserId,
		Type:        req.Type,
		CategoryId:  req.CategoryId,
		AccountId:   req.AccountId,
		Amount:      req.Amount,
		Description: strings.TrimSpace(req.Description),
		Frequency:   req.Frequency,
		DayOfMonth:  req.DayOfMonth,
		DayOfWeek:   req.DayOfWeek,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		NextDue:     nextDue,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.Repository.Create(ctx, recurring); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return recurring, nil
}

func (s *Service) UpdateRecurring(ctx context.Context, recurringID, userID ulid.ULID, req *UpdateRecurringRequest) error {
	recurring, err := s.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		return err
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return appErrors.NewValidationError("amount", "deve ser maior que zero")
		}
		recurring.Amount = *req.Amount
	}

	if req.Description != nil {
		recurring.Description = strings.TrimSpace(*req.Description)
	}

	if req.IsActive != nil {
		recurring.IsActive = *req.IsActive
	}

	if req.EndDate != nil {
		recurring.EndDate = req.EndDate
	}

	if req.NextDue != nil {
		recurring.NextDue = *req.NextDue
	}

	recurring.UpdatedAt = time.Now()

	return s.Repository.Update(ctx, recurring)
}

func (s *Service) DeleteRecurring(ctx context.Context, recurringID, userID ulid.ULID) error {
	_, err := s.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		return err
	}

	return s.Repository.Delete(ctx, recurringID, userID)
}

func (s *Service) GetRecurringByID(ctx context.Context, recurringID, userID ulid.ULID) (*RecurringTransaction, error) {
	recurring, err := s.Repository.GetByID(ctx, recurringID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if recurring.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return recurring, nil
}

func (s *Service) ListRecurring(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*RecurringTransaction, int64, error) {
	if err := s.EnsureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetByUserID(ctx, userID, pagination)
}

func (s *Service) ProcessDueTransactions(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)

	dueTransactions, _, err := s.Repository.GetDueTransactions(ctx, today, nil)
	if err != nil {
		return err
	}

	for _, recurring := range dueTransactions {
		if err := s.processRecurring(ctx, recurring, today); err != nil {
			continue
		}
	}

	return nil
}

func (s *Service) ProcessRecurringManually(ctx context.Context, recurringID, userID ulid.ULID, processDate *time.Time) (*transaction.Transaction, error) {
	recurring, err := s.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.validateManualProcess(recurring, processDate); err != nil {
		return nil, err
	}

	date := time.Now().Truncate(24 * time.Hour)
	if processDate != nil {
		date = processDate.Truncate(24 * time.Hour)
	}

	tx, err := s.createTransactionFromRecurring(ctx, recurring, date)
	if err != nil {
		return nil, err
	}

	nextDue := s.calculateNextDue(date, recurring.Frequency, recurring.DayOfMonth, recurring.DayOfWeek)
	if err := s.Repository.UpdateLastProcessed(ctx, recurring.Id, date, nextDue); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return tx, nil
}

func (s *Service) validateCreateRequest(ctx context.Context, req *CreateRecurringRequest) error {
	if req.Amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	if !req.Frequency.IsValid() {
		return appErrors.NewValidationError("frequency", "frequencia invalida")
	}

	if req.Type != "RECEIPT" && req.Type != "EXPENSE" {
		return appErrors.NewValidationError("type", "tipo invalido")
	}

	if s.CategoryService != nil {
		if err := s.CategoryService.ValidateAndEnsureExists(ctx, req.CategoryId, req.UserId); err != nil {
			return err
		}
	}

	if req.Frequency == FrequencyMonthly && (req.DayOfMonth < 1 || req.DayOfMonth > 31) {
		return appErrors.NewValidationError("day_of_month", "deve estar entre 1 e 31")
	}

	if req.Frequency == FrequencyWeekly && (req.DayOfWeek < 0 || req.DayOfWeek > 6) {
		return appErrors.NewValidationError("day_of_week", "deve estar entre 0 (domingo) e 6 (sabado)")
	}

	return nil
}

func (s *Service) validateManualProcess(recurring *RecurringTransaction, processDate *time.Time) error {
	if !recurring.IsActive {
		return appErrors.NewValidationError("recurring", "transacao recorrente esta pausada")
	}

	if recurring.AccountId == nil {
		return appErrors.NewValidationError("account_id", "transacao recorrente nao possui conta associada")
	}

	date := time.Now().Truncate(24 * time.Hour)
	if processDate != nil {
		date = processDate.Truncate(24 * time.Hour)
	}

	if recurring.EndDate != nil && date.After(*recurring.EndDate) {
		return appErrors.NewValidationError("process_date", "data de processamento e posterior a data de fim da recorrencia")
	}

	if date.Before(recurring.StartDate) {
		return appErrors.NewValidationError("process_date", "nao e possivel processar antes da data de inicio")
	}

	if recurring.LastProcessed != nil {
		lastProcessedDate := recurring.LastProcessed.Truncate(24 * time.Hour)
		if lastProcessedDate.Equal(date) {
			return appErrors.NewValidationError("recurring", "transacao recorrente ja foi processada hoje")
		}
	}

	return nil
}

func (s *Service) processRecurring(ctx context.Context, recurring *RecurringTransaction, today time.Time) error {
	if recurring.EndDate != nil && today.After(*recurring.EndDate) {
		return nil
	}

	if recurring.AccountId == nil {
		return nil
	}

	_, err := s.createTransactionFromRecurring(ctx, recurring, today)
	if err != nil {
		return err
	}

	nextDue := s.calculateNextDue(today, recurring.Frequency, recurring.DayOfMonth, recurring.DayOfWeek)
	_ = s.Repository.UpdateLastProcessed(ctx, recurring.Id, today, nextDue)

	return nil
}

func (s *Service) createTransactionFromRecurring(ctx context.Context, recurring *RecurringTransaction, date time.Time) (*transaction.Transaction, error) {
	categoryID := &recurring.CategoryId
	tx := &transaction.Transaction{
		Id:          pkg.GenerateULIDObject(),
		UserId:      recurring.UserId,
		AccountId:   *recurring.AccountId,
		Type:        transaction.Types(recurring.Type),
		CategoryId:  categoryID,
		Amount:      recurring.Amount,
		Description: recurring.Description + " (recorrente)",
		Date:        date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if s.TransactionService != nil {
		if err := s.TransactionService.CreateTransaction(ctx, tx); err != nil {
			return nil, err
		}
	} else {
		if err := s.TransactionRepo.Create(ctx, tx); err != nil {
			return nil, appErrors.NewDatabaseError(err)
		}
	}

	return tx, nil
}

func (s *Service) calculateNextDue(from time.Time, frequency FrequencyType, dayOfMonth, dayOfWeek int) time.Time {
	switch frequency {
	case FrequencyDaily:
		return from.AddDate(0, 0, 1)

	case FrequencyWeekly:
		daysUntil := (dayOfWeek - int(from.Weekday()) + 7) % 7
		if daysUntil == 0 {
			daysUntil = 7
		}
		return from.AddDate(0, 0, daysUntil)

	case FrequencyMonthly:
		nextMonth := from.AddDate(0, 1, 0)
		year, month, _ := nextMonth.Date()
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		day := dayOfMonth
		if day > lastDay {
			day = lastDay
		}
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	case FrequencyYearly:
		return from.AddDate(1, 0, 0)

	default:
		return from.AddDate(0, 1, 0)
	}
}

type CreateRecurringRequest struct {
	UserId      ulid.ULID
	Type        string
	CategoryId  ulid.ULID
	AccountId   *ulid.ULID
	Amount      float64
	Description string
	Frequency   FrequencyType
	DayOfMonth  int
	DayOfWeek   int
	StartDate   time.Time
	EndDate     *time.Time
}

type UpdateRecurringRequest struct {
	Amount      *float64
	Description *string
	IsActive    *bool
	EndDate     *time.Time
	NextDue     *time.Time
}
