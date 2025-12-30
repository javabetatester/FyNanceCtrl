package auth

import (
	"context"

	"Fynance/config"
	appErrors "Fynance/internal/errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type GoogleOAuthProvider struct {
	config   *oauth2.Config
	clientID string
	enabled  bool
}

func NewGoogleOAuthProvider(cfg config.GoogleOAuthConfig) (OAuthProvider, error) {
	if !cfg.Enabled {
		return nil, appErrors.NewAuthError("OAUTH_DISABLED", "OAuth do Google está desabilitado")
	}
	if cfg.ClientID == "" {
		return nil, appErrors.NewAuthError("OAUTH_CONFIG_MISSING", "GOOGLE_OAUTH_CLIENT_ID não configurado")
	}

	var oauthConfig *oauth2.Config
	if cfg.ClientSecret != "" && cfg.RedirectURL != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
	}

	return &GoogleOAuthProvider{
		config:   oauthConfig,
		clientID: cfg.ClientID,
		enabled:  cfg.Enabled,
	}, nil
}

func (g *GoogleOAuthProvider) GetAuthURL(state string) string {
	if g.config == nil {
		return ""
	}
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (g *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	if g.config == nil {
		return "", appErrors.NewAuthError("OAUTH_CONFIG_INCOMPLETE", "Configuração OAuth incompleta para fluxo de código")
	}

	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return "", appErrors.NewAuthError("TOKEN_EXCHANGE_FAILED", "Falha ao trocar código por token").WithError(err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", appErrors.NewAuthError("ID_TOKEN_MISSING", "ID token não encontrado na resposta")
	}

	return idToken, nil
}

func (g *GoogleOAuthProvider) VerifyToken(ctx context.Context, credential string) (*OAuthUserInfo, error) {
	payload, err := idtoken.Validate(ctx, credential, g.clientID)
	if err != nil {
		return nil, appErrors.NewAuthError("TOKEN_INVALID", "Token do Google inválido").WithError(err)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, appErrors.NewAuthError("EMAIL_MISSING", "Email não encontrado no token")
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	return &OAuthUserInfo{
		Email:   email,
		Name:    name,
		Picture: picture,
	}, nil
}
