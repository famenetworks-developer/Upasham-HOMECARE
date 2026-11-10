package auth_test

import (
	"testing"

	"upasham-backend/internal/auth"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	auth.SetJWTSecret("test-secret-key-1234567890")

	token, err := auth.GenerateToken("usr-001", "patient@test.com", "patient", "Test Patient", "pacc-001", "")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "usr-001" || claims.Role != "patient" || claims.PatientAccountID != "pacc-001" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Password hashing & comparison
	hashed, err := auth.HashPassword("MySecretPass@123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPassword(hashed, "MySecretPass@123") {
		t.Fatalf("password check should succeed for correct password")
	}

	if auth.CheckPassword(hashed, "WrongPassword") {
		t.Fatalf("password check should fail for wrong password")
	}
}
