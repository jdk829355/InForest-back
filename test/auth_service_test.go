package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/jdk829355/InForest_back/internal/service/auth"
)

func TestValidateTokenSuccess(t *testing.T) {
	t.Parallel()

	const (
		secret   = "test-secret"
		expected = "user-123"
	)

	svc, err := auth.NewAuthService(secret)
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": expected,
	})

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("unexpected error signing token: %v", err)
	}

	id, err := svc.ValidateToken(context.Background(), "Bearer "+signedToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != expected {
		t.Fatalf("expected id %q, got %q", expected, id)
	}
}

func TestValidateTokenInvalidSignature(t *testing.T) {
	t.Parallel()

	const secret = "correct-secret"

	svc, err := auth.NewAuthService(secret)
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
	})

	signedToken, err := token.SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatalf("unexpected error signing token: %v", err)
	}

	_, err = svc.ValidateToken(context.Background(), "Bearer "+signedToken)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
