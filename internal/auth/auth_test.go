package auth

import (
	"testing"

	"github.com/google/uuid"
)

const testSigningKey = "test-signing-key-must-have-at-least-thirty-two-bytes"

func TestJWTManagerIssuesAndVerifiesAccessToken(t *testing.T) {
	manager := NewJWTManager(testSigningKey)
	userID := uuid.New()
	token, err := manager.Issue(userID)
	if err != nil { t.Fatalf("Issue returned an error: %v", err) }
	actualUserID, err := manager.Verify(token)
	if err != nil { t.Fatalf("Verify returned an error: %v", err) }
	if actualUserID != userID { t.Fatalf("user ID = %s, want %s", actualUserID, userID) }
}

func TestJWTManagerRejectsTokenFromAnotherSigningKey(t *testing.T) {
	token, err := NewJWTManager(testSigningKey).Issue(uuid.New())
	if err != nil { t.Fatalf("Issue returned an error: %v", err) }
	if _, err = NewJWTManager("another-signing-key-with-at-least-thirty-two-bytes").Verify(token); err == nil {
		t.Fatal("expected token signed by another key to be rejected")
	}
}

func TestRefreshTokenIsRandomAndStoredAsHash(t *testing.T) {
	token, err := NewRefreshToken()
	if err != nil { t.Fatalf("NewRefreshToken returned an error: %v", err) }
	hash := HashRefreshToken(token)
	if len(hash) != 64 { t.Fatalf("hash length = %d, want 64", len(hash)) }
	if hash == token { t.Fatal("refresh token must not be stored as its raw value") }
}

func TestPasswordHashVerification(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil { t.Fatalf("HashPassword returned an error: %v", err) }
	if err = ComparePassword(hash, "correct-horse-battery-staple"); err != nil { t.Fatalf("valid password rejected: %v", err) }
	if err = ComparePassword(hash, "incorrect-password"); err == nil { t.Fatal("invalid password accepted") }
}
