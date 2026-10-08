package utils

import (
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	const userID = uint(42)
	token, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	gotUserID, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("ParseToken() user ID = %d, want %d", gotUserID, userID)
	}
}

func TestGenerateTokenRequiresSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	if _, err := GenerateToken(42); err == nil {
		t.Fatal("GenerateToken() succeeded without JWT_SECRET")
	}
}
