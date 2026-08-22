package persistence

import (
	dAuth "agropal/internal/domain/auth"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthRepository_CreateAndGetRefreshToken(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewAuthRepository(testLogger)
	userID := insertTestUser(t)

	// ExpiresAt is a Unix timestamp (seconds); truncate to whole seconds
	// since that's the precision the round trip through TIMESTAMPTZ
	// preserves.
	token := dAuth.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Token:     "opaque-refresh-token",
		UserAgent: "go-test",
		IPAddress: "127.0.0.1",
	}

	// Regression test: refresh_tokens.expires_at is NOT NULL with no
	// default, but the model used to omit it from the INSERT entirely -
	// every CreateRefreshToken call (i.e. every login) would have failed.
	if err := repo.CreateRefreshToken(ctx, testDB, token); err != nil {
		t.Fatalf("CreateRefreshToken() error = %v, want nil", err)
	}

	got, err := repo.GetRefreshToken(ctx, testDB, token.ID)
	if err != nil {
		t.Fatalf("GetRefreshToken() error = %v, want nil", err)
	}
	if got != token {
		t.Errorf("GetRefreshToken() = %+v, want %+v", got, token)
	}
}

func TestAuthRepository_GetRefreshToken_NotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewAuthRepository(testLogger)

	_, err := repo.GetRefreshToken(ctx, testDB, uuid.NewString())
	if err == nil {
		t.Fatal("GetRefreshToken() error = nil, want a not-found error for a nonexistent token")
	}
}

func TestAuthRepository_DeleteRefreshToken(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewAuthRepository(testLogger)
	userID := insertTestUser(t)

	token := dAuth.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Token:     "opaque-refresh-token",
	}
	if err := repo.CreateRefreshToken(ctx, testDB, token); err != nil {
		t.Fatalf("CreateRefreshToken() error = %v, want nil", err)
	}

	if err := repo.DeleteRefreshToken(ctx, testDB, token.ID); err != nil {
		t.Fatalf("DeleteRefreshToken() error = %v, want nil", err)
	}

	if _, err := repo.GetRefreshToken(ctx, testDB, token.ID); err == nil {
		t.Error("GetRefreshToken() error = nil after DeleteRefreshToken, want a not-found error")
	}
}
