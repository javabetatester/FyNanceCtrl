package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
)

type OAuthUserInfo struct {
	Email string
	Name  string
	Picture string
}

type OAuthProvider interface {
	VerifyToken(ctx context.Context, credential string) (*OAuthUserInfo, error)
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
}

func generateSecurePassword() (string, error) {
	const passwordLength = 32
	bytes := make([]byte, passwordLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", appErrors.ErrInternalServer.WithError(err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func createUserFromOAuth(ctx context.Context, userService *user.Service, info *OAuthUserInfo) (*user.User, error) {
	password, err := generateSecurePassword()
	if err != nil {
		return nil, err
	}

	name := info.Name
	if name == "" {
		name = "Usuário OAuth"
	}

	newUser := user.User{
		Name:     name,
		Email:    info.Email,
		Password: password,
	}

	if err := userService.Create(ctx, &newUser); err != nil {
		return nil, err
	}

	return &newUser, nil
}

func GenerateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", appErrors.ErrInternalServer.WithError(err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
