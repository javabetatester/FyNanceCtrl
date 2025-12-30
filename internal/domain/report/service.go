package report

import (
	"context"
	"time"

	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository  ReportRepository
	UserService *user.Service
}

func (s *Service) GetMonthlyReport(ctx context.Context, userID ulid.ULID, month, year int) (*MonthlyReport, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	if month < 1 || month > 12 {
		return nil, appErrors.NewValidationError("month", "deve estar entre 1 e 12")
	}

	if year < 2000 || year > 2100 {
		return nil, appErrors.NewValidationError("year", "ano invalido")
	}

	return s.Repository.GetMonthlyReport(userID, month, year)
}

func (s *Service) GetYearlyReport(ctx context.Context, userID ulid.ULID, year int) (*YearlyReport, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	if year < 2000 || year > 2100 {
		return nil, appErrors.NewValidationError("year", "ano invalido")
	}

	return s.Repository.GetYearlyReport(userID, year)
}

func (s *Service) GetCategoryReport(ctx context.Context, userID, categoryID ulid.ULID, startDate, endDate time.Time) (*CategoryReport, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	if endDate.Before(startDate) {
		return nil, appErrors.NewValidationError("end_date", "deve ser posterior a data inicial")
	}

	return s.Repository.GetCategoryReport(userID, categoryID, startDate, endDate)
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
