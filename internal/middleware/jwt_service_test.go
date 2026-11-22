package middleware

import (
	"strings"
	"testing"
	"time"

	"Fynance/config"
	"Fynance/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestService(t *testing.T, issuer string) *JwtService {
	t.Helper()
	s, err := NewJwtService(config.JWTConfig{SecretKey: testSecret, Issuer: issuer, ExpiresIn: time.Hour}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestJwtService(t *testing.T) {
	s := newTestService(t, "fynance_api")

	t.Run("round trip keeps subject and plan", func(t *testing.T) {
		tok, err := s.sign("user-1", user.PlanPro, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		c, err := s.ParseToken(tok)
		if err != nil {
			t.Fatal(err)
		}
		if c.Sub != "user-1" || c.Plan != user.PlanPro {
			t.Fatalf("got sub=%q plan=%q", c.Sub, c.Plan)
		}
		if !s.ValidateToken(tok) {
			t.Fatal("valid token rejected")
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		tok, _ := s.sign("user-1", user.PlanFree, time.Now().Add(-2*time.Hour))
		if s.ValidateToken(tok) {
			t.Fatal("expired token accepted")
		}
	})

	t.Run("token from another issuer is rejected", func(t *testing.T) {
		tok, _ := newTestService(t, "someone_else").sign("user-1", user.PlanFree, time.Now())
		if s.ValidateToken(tok) {
			t.Fatal("foreign issuer accepted")
		}
	})

	t.Run("tampered signature is rejected", func(t *testing.T) {
		tok, _ := s.sign("user-1", user.PlanFree, time.Now())
		if s.ValidateToken(tok[:len(tok)-2] + "xx") {
			t.Fatal("tampered token accepted")
		}
	})

	t.Run("other signing methods are rejected", func(t *testing.T) {
		claim := &Claim{Sub: "user-1", RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "fynance_api",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}}
		tok, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claim).SignedString(jwt.UnsafeAllowNoneSignatureType)
		if s.ValidateToken(tok) {
			t.Fatal("alg=none accepted")
		}
		if strings.Count(tok, ".") != 2 {
			t.Fatal("unexpected token shape")
		}
	})

	t.Run("non-positive TTL fails construction", func(t *testing.T) {
		if _, err := NewJwtService(config.JWTConfig{SecretKey: testSecret, Issuer: "x"}, nil); err == nil {
			t.Fatal("expected error")
		}
	})
}
