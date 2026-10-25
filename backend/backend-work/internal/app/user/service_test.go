package user

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier
// instead of opening a real transaction. Safe here because the fake
// repositories below never touch the Querier they're given - they stand in
// for the whole repository, not just the SQL layer underneath it.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier {
	return nil
}

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only the methods
// user.Service actually calls are wired to a *Func field; everything else
// panics if called - a test relying on an un-stubbed call fails loudly
// instead of silently zero-valuing.
type fakeUserRepository struct {
	findUserByIDFunc       func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
	updateUserFunc         func(ctx context.Context, q database.Querier, user dUser.User) error
	updateUserPasswordFunc func(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error
	softDeleteUserFunc     func(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error
	deleteUserFunc         func(ctx context.Context, q database.Querier, userID uuid.UUID) error
	listUsersFunc          func(ctx context.Context, q database.Querier) ([]dUser.User, error)
}

func (f *fakeUserRepository) CreateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: CreateUser not stubbed")
}

func (f *fakeUserRepository) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	if f.findUserByIDFunc == nil {
		panic("fakeUserRepository: findUserByIDFunc not stubbed")
	}
	return f.findUserByIDFunc(ctx, q, userID)
}

func (f *fakeUserRepository) FindUserByEmail(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByEmail not stubbed")
}

func (f *fakeUserRepository) FindUserByUsername(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByUsername not stubbed")
}

func (f *fakeUserRepository) ListUsers(ctx context.Context, q database.Querier) ([]dUser.User, error) {
	if f.listUsersFunc == nil {
		panic("fakeUserRepository: listUsersFunc not stubbed")
	}
	return f.listUsersFunc(ctx, q)
}

func (f *fakeUserRepository) UpdateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	if f.updateUserFunc == nil {
		panic("fakeUserRepository: updateUserFunc not stubbed")
	}
	return f.updateUserFunc(ctx, q, user)
}

func (f *fakeUserRepository) UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error {
	if f.updateUserPasswordFunc == nil {
		panic("fakeUserRepository: updateUserPasswordFunc not stubbed")
	}
	return f.updateUserPasswordFunc(ctx, q, userID, password)
}

func (f *fakeUserRepository) VerifyUserEmail(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: VerifyUserEmail not stubbed")
}

func (f *fakeUserRepository) SoftDeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error {
	if f.softDeleteUserFunc == nil {
		panic("fakeUserRepository: softDeleteUserFunc not stubbed")
	}
	return f.softDeleteUserFunc(ctx, q, userID, deletedAt)
}

func (f *fakeUserRepository) DeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	if f.deleteUserFunc == nil {
		panic("fakeUserRepository: deleteUserFunc not stubbed")
	}
	return f.deleteUserFunc(ctx, q, userID)
}

// fakeAuthRepository implements dAuth.Repository. Only InvalidateSessions is
// exercised by user.Service (on a password change).
type fakeAuthRepository struct {
	invalidateSessionsFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) error
}

func (f *fakeAuthRepository) CreateRefreshToken(context.Context, database.Querier, dAuth.RefreshToken) error {
	panic("fakeAuthRepository: CreateRefreshToken not stubbed")
}

func (f *fakeAuthRepository) GetRefreshToken(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
	panic("fakeAuthRepository: GetRefreshToken not stubbed")
}

func (f *fakeAuthRepository) DeleteRefreshToken(context.Context, database.Querier, string) error {
	panic("fakeAuthRepository: DeleteRefreshToken not stubbed")
}

func (f *fakeAuthRepository) DeleteExpiredRefreshTokens(context.Context, database.Querier, time.Time) error {
	panic("fakeAuthRepository: DeleteExpiredRefreshTokens not stubbed")
}

func (f *fakeAuthRepository) InvalidateSessions(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	if f.invalidateSessionsFunc == nil {
		return nil
	}
	return f.invalidateSessionsFunc(ctx, q, userID)
}

// fakeNoteRepository implements dNote.Repository. Only CreateNote is
// exercised by user.Service.
type fakeNoteRepository struct {
	createNoteFunc func(ctx context.Context, q database.Querier, note dNote.Note) error
}

func (f *fakeNoteRepository) CreateNote(ctx context.Context, q database.Querier, note dNote.Note) error {
	if f.createNoteFunc == nil {
		panic("fakeNoteRepository: createNoteFunc not stubbed")
	}
	return f.createNoteFunc(ctx, q, note)
}

func (f *fakeNoteRepository) CreateManyNotes(context.Context, database.Querier, []dNote.Note) error {
	panic("fakeNoteRepository: CreateManyNotes not stubbed")
}

func (f *fakeNoteRepository) ListNotes(context.Context, database.Querier, dNote.NoteType, []uuid.UUID, map[string]string) ([]dNote.Note, error) {
	panic("fakeNoteRepository: ListNotes not stubbed")
}

func (f *fakeNoteRepository) DeleteNotes(context.Context, database.Querier, []uuid.UUID) error {
	panic("fakeNoteRepository: DeleteNotes not stubbed")
}

/* ==================== Helpers ====================== */

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

/* ==================== Tests: GetByID / verifyAccess ==================== */

func TestService_GetByID(t *testing.T) {
	selfID := uuid.New()
	otherID := uuid.New()
	adminID := uuid.New()
	target := dUser.User{ID: otherID, Account: dUser.AccountInfo{Name: "Target"}}

	t.Run("a user can fetch themselves", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: selfID}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		got, err := svc.GetByID(ctxWithUser(selfID), selfID)
		if err != nil {
			t.Fatalf("GetByID() error = %v, want nil", err)
		}
		if got.ID != selfID {
			t.Errorf("got.ID = %v, want %v", got.ID, selfID)
		}
	})

	t.Run("uuid.Nil resolves to the requester", func(t *testing.T) {
		var queriedID uuid.UUID
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) (dUser.User, error) {
				queriedID = id
				return dUser.User{ID: id}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		if _, err := svc.GetByID(ctxWithUser(selfID), uuid.Nil); err != nil {
			t.Fatalf("GetByID() error = %v, want nil", err)
		}
		if queriedID != selfID {
			t.Errorf("FindUserByID queried %v, want requester %v", queriedID, selfID)
		}
	})

	t.Run("a non-admin cannot fetch another user", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return target, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		_, err := svc.GetByID(ctxWithUser(selfID), otherID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("GetByID() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("an admin can fetch another user", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return target, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		got, err := svc.GetByID(ctxWithUserAndRole(adminID, config.RoleAdmin), otherID)
		if err != nil {
			t.Fatalf("GetByID() error = %v, want nil", err)
		}
		if got.ID != otherID {
			t.Errorf("got.ID = %v, want %v", got.ID, otherID)
		}
	})

	t.Run("missing requester ID in context is treated as an invalid user", func(t *testing.T) {
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeAuthRepository{}, &fakeNoteRepository{})

		_, err := svc.GetByID(context.Background(), uuid.New())
		if !errors.Is(err, dUser.ErrInvalidUserID) {
			t.Errorf("GetByID() error = %v, want %v", err, dUser.ErrInvalidUserID)
		}
	})
}

/* ==================== Tests: UpdateUser ==================== */

func TestService_UpdateUser(t *testing.T) {
	selfID := uuid.New()

	var saved dUser.User
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: selfID}, nil
		},
		updateUserFunc: func(_ context.Context, _ database.Querier, u dUser.User) error {
			saved = u
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

	// Mirrors what request.UserUpdateRequest.ToCommand produces: every
	// pointer field non-nil, since UserCommand.ToDomain dereferences them
	// unconditionally.
	cmd := UserCommand{
		Name:     shared.Ptr("New Name"),
		Email:    shared.Ptr(""),
		Username: shared.Ptr(""),
		Avatar:   shared.Ptr("avatar.png"),
		Password: shared.Ptr(""),
	}

	if err := svc.UpdateUser(ctxWithUser(selfID), cmd); err != nil {
		t.Fatalf("UpdateUser() error = %v, want nil", err)
	}
	if saved.ID != selfID {
		t.Errorf("saved.ID = %v, want %v (UpdateUser should overwrite cmd.UserID with the verified target)", saved.ID, selfID)
	}
	if saved.Account.Name != "New Name" {
		t.Errorf("saved.Account.Name = %q, want %q", saved.Account.Name, "New Name")
	}
}

/* ==================== Tests: UpdateUserPassword ==================== */

func TestService_UpdateUserPassword(t *testing.T) {
	selfID := uuid.New()
	const oldPassword = "OldPassw0rd!"
	oldHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	newUserRepo := func(savedUserID *uuid.UUID, savedPassword *string) *fakeUserRepository {
		return &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: selfID, Security: dUser.SecurityInfo{Password: string(oldHash)}}, nil
			},
			updateUserPasswordFunc: func(_ context.Context, _ database.Querier, userID uuid.UUID, password string) error {
				*savedUserID = userID
				*savedPassword = password
				return nil
			},
		}
	}

	t.Run("valid current password re-hashes and revokes sessions", func(t *testing.T) {
		var savedUserID uuid.UUID
		var savedPassword string
		var invalidatedID uuid.UUID
		authRepo := &fakeAuthRepository{
			invalidateSessionsFunc: func(_ context.Context, _ database.Querier, userID uuid.UUID) error {
				invalidatedID = userID
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, newUserRepo(&savedUserID, &savedPassword), authRepo, &fakeNoteRepository{})

		if err := svc.UpdateUserPassword(ctxWithUser(selfID), selfID, oldPassword, "NewPassw0rd!"); err != nil {
			t.Fatalf("UpdateUserPassword() error = %v, want nil", err)
		}
		if savedUserID != selfID {
			t.Errorf("savedUserID = %v, want %v", savedUserID, selfID)
		}
		if savedPassword == "NewPassw0rd!" || savedPassword == "" {
			t.Errorf("savedPassword = %q, want a bcrypt hash (not the plaintext or empty)", savedPassword)
		}
		if bcrypt.CompareHashAndPassword([]byte(savedPassword), []byte("NewPassw0rd!")) != nil {
			t.Errorf("savedPassword is not a hash of the new password")
		}
		if invalidatedID != selfID {
			t.Errorf("InvalidateSessions called with %v, want %v", invalidatedID, selfID)
		}
	})

	t.Run("wrong current password is rejected and nothing is written", func(t *testing.T) {
		var savedUserID uuid.UUID
		var savedPassword string
		userRepo := newUserRepo(&savedUserID, &savedPassword)
		userRepo.updateUserPasswordFunc = func(context.Context, database.Querier, uuid.UUID, string) error {
			t.Fatal("UpdateUserPassword should not be called when the current password is wrong")
			return nil
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		err := svc.UpdateUserPassword(ctxWithUser(selfID), selfID, "WrongPassw0rd!", "NewPassw0rd!")
		if !errors.Is(err, shared.ErrUnauthorized) {
			t.Errorf("UpdateUserPassword() error = %v, want %v", err, shared.ErrUnauthorized)
		}
	})
}

/* ==================== Tests: Delete ==================== */

func TestService_Delete(t *testing.T) {
	requesterID := uuid.New()

	t.Run("soft delete calls SoftDeleteUser and records an archive note", func(t *testing.T) {
		var softDeletedID uuid.UUID
		var note dNote.Note
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: requesterID}, nil
			},
			softDeleteUserFunc: func(_ context.Context, _ database.Querier, userID uuid.UUID, _ time.Time) error {
				softDeletedID = userID
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(_ context.Context, _ database.Querier, n dNote.Note) error {
				note = n
				return nil
			},
		}
		var signedOut uuid.UUID
		authRepo := &fakeAuthRepository{
			invalidateSessionsFunc: func(_ context.Context, _ database.Querier, userID uuid.UUID) error {
				signedOut = userID
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, authRepo, noteRepo)

		if err := svc.Delete(ctxWithUser(requesterID), shared.DeleteSoft, requesterID); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if softDeletedID != requesterID {
			t.Errorf("SoftDeleteUser called with %v, want %v", softDeletedID, requesterID)
		}
		if signedOut != requesterID {
			t.Errorf("sessions invalidated for %v, want %v (a deleted account must be signed out)", signedOut, requesterID)
		}
		if note.Type != dNote.TypeArchive {
			t.Errorf("note.Type = %q, want %q", note.Type, dNote.TypeArchive)
		}
	})

	t.Run("hard delete calls DeleteUser and records a delete note", func(t *testing.T) {
		var hardDeletedID uuid.UUID
		var note dNote.Note
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: requesterID}, nil
			},
			deleteUserFunc: func(_ context.Context, _ database.Querier, userID uuid.UUID) error {
				hardDeletedID = userID
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(_ context.Context, _ database.Querier, n dNote.Note) error {
				note = n
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, noteRepo)

		if err := svc.Delete(ctxWithUser(requesterID), shared.DeleteHard, requesterID); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if hardDeletedID != requesterID {
			t.Errorf("DeleteUser called with %v, want %v", hardDeletedID, requesterID)
		}
		if note.Type != dNote.TypeDelete {
			t.Errorf("note.Type = %q, want %q", note.Type, dNote.TypeDelete)
		}
	})

	t.Run("forbids deleting another user without admin", func(t *testing.T) {
		otherID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherID}, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		err := svc.Delete(ctxWithUser(requesterID), shared.DeleteSoft, otherID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Delete() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

/* ==================== Tests: GetUsers ==================== */

func TestService_GetUsers(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()
	want := []dUser.User{{ID: uuid.New()}, {ID: uuid.New()}}

	t.Run("an admin gets the full user list", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			listUsersFunc: func(context.Context, database.Querier) ([]dUser.User, error) {
				return want, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		got, err := svc.GetUsers(ctxWithUserAndRole(adminID, config.RoleAdmin))
		if err != nil {
			t.Fatalf("GetUsers() error = %v, want nil", err)
		}
		if len(got) != len(want) {
			t.Fatalf("GetUsers() returned %d users, want %d", len(got), len(want))
		}
	})

	t.Run("a non-admin is forbidden and the repo is never touched", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			listUsersFunc: func(context.Context, database.Querier) ([]dUser.User, error) {
				t.Fatal("ListUsers should not be called for a non-admin caller")
				return nil, nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, &fakeNoteRepository{})

		_, err := svc.GetUsers(ctxWithUserAndRole(userID, config.RoleUser))
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("GetUsers() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("an unauthenticated caller is rejected", func(t *testing.T) {
		svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakeAuthRepository{}, &fakeNoteRepository{})

		if _, err := svc.GetUsers(context.Background()); err == nil {
			t.Fatal("GetUsers() error = nil, want an unauthenticated error")
		}
	})
}
