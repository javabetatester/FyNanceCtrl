package middleware

import (
	"context"
	"errors"
	"time"

	"Fynance/config"
	"Fynance/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type JwtService struct {
	secretKey   string
	issuer      string
	tokenTTL    time.Duration
	parser      *jwt.Parser
	userService *user.Service
}

func NewJwtService(settings config.JWTConfig, userService *user.Service) (*JwtService, error) {
	if settings.ExpiresIn <= 0 {
		return nil, errors.New("JWT_EXPIRES_IN inválido")
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuer(settings.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	return &JwtService{
		secretKey:   settings.SecretKey,
		issuer:      settings.Issuer,
		tokenTTL:    settings.ExpiresIn,
		parser:      parser,
		userService: userService,
	}, nil
}

type Claim struct {
	Sub  string    `json:"sub"`
	Plan user.Plan `json:"plan"`
	jwt.RegisteredClaims
}

func (s *JwtService) GenerateToken(ctx context.Context, id ulid.ULID) (string, error) {
	plan, err := s.userService.GetPlan(ctx, id)
	if err != nil {
		return "", err
	}
	return s.sign(id.String(), plan, time.Now())
}

func (s *JwtService) sign(sub string, plan user.Plan, now time.Time) (string, error) {
	claim := &Claim{
		Sub:  sub,
		Plan: plan,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.issuer,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claim).SignedString([]byte(s.secretKey))
}

func (s *JwtService) ValidateToken(tokenString string) bool {
	_, err := s.parse(tokenString)
	return err == nil
}

func (s *JwtService) ParseToken(tokenString string) (*Claim, error) {
	return s.parse(tokenString)
}

func (s *JwtService) parse(tokenString string) (*Claim, error) {
	token, err := s.parser.ParseWithClaims(tokenString, &Claim{}, func(*jwt.Token) (any, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claim)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
