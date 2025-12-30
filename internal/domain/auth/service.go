package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type Service struct {
	Repository     user.UserRepository
	UserService    *user.Service
	GoogleClientID string
}

func NewService(
	repo user.UserRepository,
	userSvc *user.Service,
	googleClientID string,
) *Service {
	return &Service{
		Repository:     repo,
		UserService:    userSvc,
		GoogleClientID: googleClientID,
	}
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

func (s *Service) GoogleLogin(ctx context.Context, credential string) (*user.User, error) {
	if s.GoogleClientID == "" {
		return nil, appErrors.NewAuthError("OAUTH_NOT_CONFIGURED", "Google OAuth não está configurado. Configure GOOGLE_OAUTH_CLIENT_ID e GOOGLE_OAUTH_ENABLED=true")
	}

	if credential == "" {
		return nil, appErrors.NewAuthError("CREDENTIAL_MISSING", "Credencial do Google não fornecida")
	}

	payload, err := idtoken.Validate(ctx, credential, s.GoogleClientID)
	if err != nil {
		errMsg := "Token do Google inválido"
		errStr := err.Error()

		if errStr != "" {
			errLower := strings.ToLower(errStr)
			if strings.Contains(errLower, "invalid_client") || strings.Contains(errLower, "not found") {
				errMsg = fmt.Sprintf("Client ID não encontrado ou não autorizado. Verifique: 1) O Client ID '%s' existe no Google Console, 2) O tipo é 'Web application', 3) O domínio está autorizado em 'Origens JavaScript autorizadas', 4) O mesmo Client ID está sendo usado no frontend", s.GoogleClientID)
			} else {
				errMsg = fmt.Sprintf("Token do Google inválido: %v. Verifique se o GOOGLE_OAUTH_CLIENT_ID está correto e corresponde ao Client ID usado no frontend", err)
			}
		}
		return nil, appErrors.NewAuthError("TOKEN_INVALID", errMsg).WithError(err)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, appErrors.NewAuthError("EMAIL_MISSING", "Email não encontrado no token")
	}

	name, _ := payload.Claims["name"].(string)
	if name == "" {
		name = "Usuário Google"
	}

	entity, err := s.Repository.GetByEmail(ctx, email)
	if err != nil {
		if appErr, ok := appErrors.AsAppError(err); ok && appErr.Code == appErrors.ErrUserNotFound.Code {
			password, err := generateSecurePassword()
			if err != nil {
				return nil, err
			}

			newUser := user.User{
				Name:     name,
				Email:    email,
				Password: password,
			}

			if err := s.UserService.Create(ctx, &newUser); err != nil {
				return nil, err
			}

			return &newUser, nil
		}
		return nil, err
	}

	return entity, nil
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
	hasSpecial, _ := regexp.MatchString(`[@$!%*?&]`, password)
	if !hasSpecial {
		return appErrors.NewValidationError("password", "deve conter ao menos um caractere especial (@$!%*?&)")
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
