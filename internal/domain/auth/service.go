package auth

import (
	"context"
	"regexp"

	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repository         user.Repository
	UserService        *user.Service
	CategoryService    *transaction.Service
}

func (s *Service) Login(ctx context.Context, login Login) (*user.User, error) {
	entity, err := s.Repository.GetByEmail(ctx, login.Email)
	if err != nil {
		if appErr, ok := appErrors.AsAppError(err); ok && appErr.Code == appErrors.ErrUserNotFound.Code {
			return nil, appErrors.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := PasswordValidate(login.Password, entity.Password); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Register(ctx context.Context, user *user.User) error {
	exists, err := s.emailExists(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return appErrors.ErrEmailAlreadyExists
	}
	if err := PasswordRequirements(user.Password); err != nil {
		return err
	}
	if err := s.UserService.Create(ctx, user); err != nil {
		return err
	}


	return nil
}

func (s *Service) emailExists(ctx context.Context, email string) (bool, error) {
	_, err := s.Repository.GetByEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	appErr, ok := appErrors.AsAppError(err)
	if !ok {
		return false, appErrors.ErrInternalServer.WithError(err)
	}
	if appErr.Code == appErrors.ErrUserNotFound.Code {
		return false, nil
	}
	return false, appErr
}

func PasswordRequirements(password string) error {
	if len(password) < 8 {
		return appErrors.NewValidationError("password", "deve conter no mínimo 8 caracteres")
	}
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return appErrors.NewValidationError("password", "deve conter ao menos uma letra maiúscula")
	}
	hasSpecial, _ := regexp.MatchString(`[@$!%*?&#]`, password)
	if !hasSpecial {
		return appErrors.NewValidationError("password", "deve conter ao menos um caractere especial (@$!%*?&#)")
	}
	return nil
}

func PasswordValidate(inputPassword string, storedPassword string) error {
	if inputPassword == "" {
		return appErrors.NewValidationError("password", "deve ser informado")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(inputPassword)); err != nil {
		return appErrors.ErrInvalidCredentials
	}
	return nil
}

func PasswordHashing(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", appErrors.ErrInternalServer.WithError(err)
	}
	return string(hash), nil
}
