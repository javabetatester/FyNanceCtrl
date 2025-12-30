package user

import (
	"context"
	"regexp"

	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repository: repo}
}

func (s *Service) Create(ctx context.Context, user *User) error {
	user.Id = pkg.GenerateULIDObject()

	now := pkg.SetTimestamps()
	user.CreatedAt = now
	user.UpdatedAt = now

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.Repository.Create(ctx, user)
}

func (s *Service) Update(ctx context.Context, user *User) error {
	return s.Repository.Update(ctx, user)
}

func (s *Service) Delete(ctx context.Context, id ulid.ULID) error {
	return s.Repository.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id ulid.ULID) (*User, error) {
	user, err := s.Repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.Repository.GetByEmail(ctx, email)
}

func (s *Service) GetPlan(ctx context.Context, id ulid.ULID) (Plan, error) {
	plan, err := s.Repository.GetPlan(ctx, id)
	if err != nil {
		return "", err
	}
	return plan, nil
}

func (s *Service) UpdatePlan(ctx context.Context, userID ulid.ULID, newPlan Plan) error {
	if !newPlan.IsValid() {
		return appErrors.NewValidationError("plan", "plano invalido")
	}

	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Plan == newPlan {
		return nil
	}

	user.Plan = newPlan
	user.PlanSince = pkg.SetTimestamps()
	user.UpdatedAt = pkg.SetTimestamps()

	return s.Repository.Update(ctx, user)
}

func (s *Service) Exists(ctx context.Context, userID ulid.ULID) error {
	_, err := s.GetByID(ctx, userID)
	return err
}

func (s *Service) UpdateName(ctx context.Context, userID ulid.ULID, name string) error {
	if name == "" {
		return appErrors.NewValidationError("name", "nome não pode estar vazio")
	}

	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Name = name
	user.UpdatedAt = pkg.SetTimestamps()

	return s.Repository.Update(ctx, user)
}

func (s *Service) UpdatePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return appErrors.ErrInvalidCredentials
	}

	if err := validatePasswordRequirements(newPassword); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return appErrors.ErrInternalServer.WithError(err)
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = pkg.SetTimestamps()

	return s.Repository.Update(ctx, user)
}

type UserServiceAdapter struct {
	service *Service
}

func NewUserServiceAdapter(service *Service) *UserServiceAdapter {
	return &UserServiceAdapter{service: service}
}

func (a *UserServiceAdapter) Exists(ctx context.Context, userID ulid.ULID) error {
	return a.service.Exists(ctx, userID)
}
func (a *UserServiceAdapter) GetByID(ctx context.Context, userID ulid.ULID) (interface{}, error) {
	return a.service.GetByID(ctx, userID)
}

func validatePasswordRequirements(password string) error {
	if len(password) < 8 {
		return appErrors.NewValidationError("new_password", "deve conter no mínimo 8 caracteres")
	}
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return appErrors.NewValidationError("new_password", "deve conter ao menos uma letra maiúscula")
	}
	hasSpecial, _ := regexp.MatchString(`[@$!%*?&]`, password)
	if !hasSpecial {
		return appErrors.NewValidationError("new_password", "deve conter ao menos um caractere especial (@$!%*?&)")
	}
	return nil
}
