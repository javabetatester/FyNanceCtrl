package shared

import (
	"context"

	appErrors "Fynance/internal/errors"

	"github.com/oklog/ulid/v2"
)

type UserCheckerService struct {
	userService UserGetter
}

func NewUserCheckerService(userService UserGetter) *UserCheckerService {
	return &UserCheckerService{userService: userService}
}

func (s *UserCheckerService) EnsureUserExists(ctx context.Context, userID ulid.ULID) error {
	if s.userService == nil {
		return appErrors.ErrInternalServer
	}

	if err := s.userService.Exists(ctx, userID); err != nil {
		return appErrors.ErrUserNotFound.WithError(err)
	}

	return nil
}

type BaseService struct {
	UserChecker *UserCheckerService
}

func (b *BaseService) EnsureUserExists(ctx context.Context, userID ulid.ULID) error {
	if b.UserChecker == nil {
		return appErrors.ErrInternalServer
	}
	return b.UserChecker.EnsureUserExists(ctx, userID)
}
