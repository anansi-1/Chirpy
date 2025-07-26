package auth_test

import (
    "testing"
    "time"
    "github.com/google/uuid"
	// "github/anansi-1/Chirpy/auth.go"
	"github/anansi-1/Chirpy/internal/auth"
)

func TestMakeAndValidateJWT_ValidToken(t *testing.T) {
    userID := uuid.New()
    secret := "test-secret"
    duration := 2 * time.Minute

    token, err := auth.MakeJWT(userID, secret, duration)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    if token == "" {
        t.Fatal("Token is empty")
    }

    validatedID, err := auth.ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("ValidateJWT failed: %v", err)
    }

    if validatedID != userID {
        t.Errorf("Expected user ID %v, got %v", userID, validatedID)
    }
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
    userID := uuid.New()
    secret := "test-secret"
    duration := -1 * time.Minute 

    token, err := auth.MakeJWT(userID, secret, duration)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    _, err = auth.ValidateJWT(token, secret)
    if err == nil {
        t.Fatal("Expected error for expired token, got nil")
    }
}

func TestValidateJWT_WrongSecret(t *testing.T) {
    userID := uuid.New()
    correctSecret := "correct-secret"
    wrongSecret := "wrong-secret"
    duration := 2 * time.Minute

    token, err := auth.MakeJWT(userID, correctSecret, duration)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    _, err = auth.ValidateJWT(token, wrongSecret)
    if err == nil {
        t.Fatal("Expected error for token signed with wrong secret, got nil")
    }
}


