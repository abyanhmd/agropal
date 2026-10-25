package persistence

import (
	dUser "agropal/internal/domain/user"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestUser() dUser.User {
	id := uuid.New()
	return dUser.User{
		ID: id,
		Account: dUser.AccountInfo{
			Name:     "Test User",
			Email:    id.String() + "@example.com",
			Username: id.String(),
			Avatar:   "avatar.png",
		},
		Security: dUser.SecurityInfo{
			Password: "hashed-password",
			Role:     "USER",
		},
	}
}

func TestUserRepository_CreateAndFindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()

	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	got, err := repo.FindUserByID(ctx, testDB, user.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v, want nil", err)
	}
	if got.Account.Email != user.Account.Email || got.Account.Username != user.Account.Username {
		t.Errorf("FindUserByID() = %+v, want Email=%q Username=%q", got, user.Account.Email, user.Account.Username)
	}
}

func TestUserRepository_FindUserByEmailAndUsername(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	byEmail, err := repo.FindUserByEmail(ctx, testDB, user.Account.Email)
	if err != nil {
		t.Fatalf("FindUserByEmail() error = %v, want nil", err)
	}
	if byEmail.ID != user.ID {
		t.Errorf("FindUserByEmail().ID = %v, want %v", byEmail.ID, user.ID)
	}

	byUsername, err := repo.FindUserByUsername(ctx, testDB, user.Account.Username)
	if err != nil {
		t.Fatalf("FindUserByUsername() error = %v, want nil", err)
	}
	if byUsername.ID != user.ID {
		t.Errorf("FindUserByUsername().ID = %v, want %v", byUsername.ID, user.ID)
	}
}

func TestUserRepository_FindUserByID_NotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)

	_, err := repo.FindUserByID(ctx, testDB, uuid.New())
	if !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("FindUserByID() error = %v, want %v", err, dUser.ErrNotFound)
	}
}

func TestUserRepository_UpdateUser(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	t.Run("updates name and avatar only", func(t *testing.T) {
		updated := user
		updated.Account.Name = "New Name"
		updated.Account.Avatar = "new-avatar.png"
		if err := repo.UpdateUser(ctx, testDB, updated); err != nil {
			t.Fatalf("UpdateUser() error = %v, want nil", err)
		}

		got, err := repo.FindUserByID(ctx, testDB, user.ID)
		if err != nil {
			t.Fatalf("FindUserByID() error = %v, want nil", err)
		}
		if got.Account.Name != "New Name" || got.Account.Avatar != "new-avatar.png" {
			t.Errorf("got.Account = %+v, want Name=New Name Avatar=new-avatar.png", got.Account)
		}
		// UpdateUser's query only touches name/avatar - email/username must
		// survive untouched.
		if got.Account.Email != user.Account.Email || got.Account.Username != user.Account.Username {
			t.Errorf("Email/Username changed unexpectedly: got %+v, want Email=%q Username=%q", got.Account, user.Account.Email, user.Account.Username)
		}
	})

	t.Run("returns ErrNotFound for a nonexistent user", func(t *testing.T) {
		ghost := newTestUser()
		err := repo.UpdateUser(ctx, testDB, ghost)
		if !errors.Is(err, dUser.ErrNotFound) {
			t.Errorf("UpdateUser() error = %v, want %v", err, dUser.ErrNotFound)
		}
	})
}

func TestUserRepository_UpdateUserPassword(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	if err := repo.UpdateUserPassword(ctx, testDB, user.ID, "new-hashed-password"); err != nil {
		t.Fatalf("UpdateUserPassword() error = %v, want nil", err)
	}
	got, err := repo.FindUserByID(ctx, testDB, user.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v, want nil", err)
	}
	if got.Security.Password != "new-hashed-password" {
		t.Errorf("Security.Password = %q, want %q", got.Security.Password, "new-hashed-password")
	}

	err = repo.UpdateUserPassword(ctx, testDB, uuid.New(), "x")
	if !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("UpdateUserPassword() for a nonexistent user error = %v, want %v", err, dUser.ErrNotFound)
	}
}

func TestUserRepository_VerifyUserEmail(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	verifiedAt := time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.VerifyUserEmail(ctx, testDB, user.ID, verifiedAt); err != nil {
		t.Fatalf("VerifyUserEmail() error = %v, want nil", err)
	}

	got, err := repo.FindUserByID(ctx, testDB, user.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v, want nil", err)
	}
	if got.Metadata.EmailVerifiedAt == nil {
		t.Fatal("Metadata.EmailVerifiedAt is nil, want a timestamp")
	}
	if !got.Metadata.EmailVerifiedAt.Equal(verifiedAt) {
		t.Errorf("Metadata.EmailVerifiedAt = %v, want %v", got.Metadata.EmailVerifiedAt, verifiedAt)
	}

	err = repo.VerifyUserEmail(ctx, testDB, uuid.New(), verifiedAt)
	if !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("VerifyUserEmail() for a nonexistent user error = %v, want %v", err, dUser.ErrNotFound)
	}
}

func TestUserRepository_SoftDeleteUser(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	if err := repo.SoftDeleteUser(ctx, testDB, user.ID, time.Now().UTC()); err != nil {
		t.Fatalf("SoftDeleteUser() error = %v, want nil", err)
	}

	// FindUserByID filters out soft-deleted rows (WHERE deleted_at IS NULL).
	if _, err := repo.FindUserByID(ctx, testDB, user.ID); !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("FindUserByID() after soft delete error = %v, want %v", err, dUser.ErrNotFound)
	}

	err := repo.SoftDeleteUser(ctx, testDB, uuid.New(), time.Now().UTC())
	if !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("SoftDeleteUser() for a nonexistent user error = %v, want %v", err, dUser.ErrNotFound)
	}
}

func TestUserRepository_DeleteUser(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewUserRepository(testLogger)
	user := newTestUser()
	if err := repo.CreateUser(ctx, testDB, user); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	if err := repo.DeleteUser(ctx, testDB, user.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v, want nil", err)
	}

	var count int
	if err := testDB.Get(&count, `SELECT COUNT(*) FROM users WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("row count after DeleteUser = %d, want 0 (hard delete)", count)
	}

	err := repo.DeleteUser(ctx, testDB, uuid.New())
	if !errors.Is(err, dUser.ErrNotFound) {
		t.Errorf("DeleteUser() for a nonexistent user error = %v, want %v", err, dUser.ErrNotFound)
	}
}
