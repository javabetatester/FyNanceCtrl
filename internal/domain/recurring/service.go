package recurring

import (
	"context"
	"strings"
	"time"

	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository         Repository
	TransactionRepo    transaction.Repository
	CategoryRepository transaction.CategoryRepository
	UserService        *user.Service
}

func (s *Service) CreateRecurring(ctx context.Context, req *CreateRecurringRequest) (*RecurringTransaction, error) {
	if err := s.ensureUserExists(ctx, req.UserId); err != nil {
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
	recurring, err := s.Repository.GetById(ctx, recurringID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if recurring.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return recurring, nil
}

func (s *Service) ListRecurring(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*RecurringTransaction, int64, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetByUserId(ctx, userID, pagination)
}

func (s *Service) ProcessDueTransactions(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)

	dueTransactions, _, err := s.Repository.GetDueTransactions(ctx, today, nil)
	if err != nil {
		return err
	}

	for _, recurring := range dueTransactions {
		if recurring.EndDate != nil && today.After(*recurring.EndDate) {
			continue
		}

		tx := &transaction.Transaction{
			Id:          pkg.GenerateULIDObject(),
			UserId:      recurring.UserId,
			Type:        transaction.Types(recurring.Type),
			CategoryId:  recurring.CategoryId,
			Amount:      recurring.Amount,
			Description: recurring.Description + " (recorrente)",
			Date:        today,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.TransactionRepo.Create(ctx, tx); err != nil {
			continue
		}

		nextDue := s.calculateNextDue(today, recurring.Frequency, recurring.DayOfMonth, recurring.DayOfWeek)
		_ = s.Repository.UpdateLastProcessed(ctx, recurring.Id, today, nextDue)
	}

	return nil
}

func (s *Service) ProcessRecurringManually(ctx context.Context, recurringID, userID ulid.ULID, processDate *time.Time) (*transaction.Transaction, error) {
	recurring, err := s.GetRecurringByID(ctx, recurringID, userID)
	if err != nil {
		return nil, err
	}

	if !recurring.IsActive {
		return nil, appErrors.NewValidationError("recurring", "transacao recorrente esta pausada")
	}

	if recurring.EndDate != nil && processDate != nil && processDate.After(*recurring.EndDate) {
		return nil, appErrors.NewValidationError("process_date", "data de processamento e posterior a data de fim da recorrencia")
	}

	date := time.Now().Truncate(24 * time.Hour)
	if processDate != nil {
		date = processDate.Truncate(24 * time.Hour)
	}

	if date.Before(recurring.StartDate) {
		return nil, appErrors.NewValidationError("process_date", "nao e possivel processar antes da data de inicio")
	}

	tx := &transaction.Transaction{
		Id:          pkg.GenerateULIDObject(),
		UserId:      recurring.UserId,
		Type:        transaction.Types(recurring.Type),
		CategoryId:  recurring.CategoryId,
		Amount:      recurring.Amount,
		Description: recurring.Description + " (recorrente)",
		Date:        date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.TransactionRepo.Create(ctx, tx); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	nextDue := s.calculateNextDue(date, recurring.Frequency, recurring.DayOfMonth, recurring.DayOfWeek)
	if err := s.Repository.UpdateLastProcessed(ctx, recurring.Id, date, nextDue); err != nil {
		return nil, appErrors.NewDatabaseError(err)
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

	_, err := s.CategoryRepository.GetByID(ctx, req.CategoryId, req.UserId)
	if err != nil {
		return appErrors.ErrCategoryNotFound
	}

	if req.Frequency == FrequencyMonthly && (req.DayOfMonth < 1 || req.DayOfMonth > 31) {
		return appErrors.NewValidationError("day_of_month", "deve estar entre 1 e 31")
	}

	if req.Frequency == FrequencyWeekly && (req.DayOfWeek < 0 || req.DayOfWeek > 6) {
		return appErrors.NewValidationError("day_of_week", "deve estar entre 0 (domingo) e 6 (sabado)")
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
}
