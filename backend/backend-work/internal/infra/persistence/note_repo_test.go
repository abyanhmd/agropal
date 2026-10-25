package persistence

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNoteRepository_CreateNoteAndListNotes(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)
	authorID := insertTestUser(t)
	assetID := uuid.New()

	note := dNote.Note{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
		Type:        dNote.TypeUpdate,
		AssetType:   shared.AssetProduct,
		AssetID:     assetID,
		UpdateField: shared.Ptr("name"),
		OldValue:    shared.Ptr("Old"),
		NewValue:    shared.Ptr("New"),
		AuthorID:    authorID,
		IsSystemLog: false,
	}
	if err := repo.CreateNote(ctx, testDB, note); err != nil {
		t.Fatalf("CreateNote() error = %v, want nil", err)
	}

	got, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, []uuid.UUID{assetID}, nil)
	if err != nil {
		t.Fatalf("ListNotes() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListNotes() returned %d notes, want 1", len(got))
	}
	if got[0].ID != note.ID {
		t.Errorf("ListNotes()[0].ID = %v, want %v", got[0].ID, note.ID)
	}
	// author_name is joined from users, not stored on notes itself.
	if got[0].AuthorName != "Test User" {
		t.Errorf("ListNotes()[0].AuthorName = %q, want %q (joined from users.name)", got[0].AuthorName, "Test User")
	}
	if shared.Str(got[0].OldValue) != "Old" || shared.Str(got[0].NewValue) != "New" {
		t.Errorf("ListNotes()[0] old/new = %q/%q, want Old/New", shared.Str(got[0].OldValue), shared.Str(got[0].NewValue))
	}
}

func TestNoteRepository_ListNotes_ReturnsOnlyTheMostRecentPerAsset(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)
	authorID := insertTestUser(t)
	assetID := uuid.New()

	older := dNote.Note{
		ID: uuid.New(), CreatedAt: time.Now().UTC().Add(-time.Hour), Type: dNote.TypeUpdate,
		AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID,
	}
	newer := dNote.Note{
		ID: uuid.New(), CreatedAt: time.Now().UTC(), Type: dNote.TypeUpdate,
		AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID,
	}
	for _, n := range []dNote.Note{older, newer} {
		if err := repo.CreateNote(ctx, testDB, n); err != nil {
			t.Fatalf("CreateNote() error = %v, want nil", err)
		}
	}

	// ListNotes uses SELECT DISTINCT ON (asset_id) ... ORDER BY asset_id,
	// created_at DESC - it returns only the latest note per asset, not
	// every matching note.
	got, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, []uuid.UUID{assetID}, nil)
	if err != nil {
		t.Fatalf("ListNotes() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].ID != newer.ID {
		t.Errorf("ListNotes() = %+v, want only the newer note %v", got, newer.ID)
	}
}

func TestNoteRepository_ListNotes_WithFieldFilter(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)
	authorID := insertTestUser(t)
	assetID := uuid.New()

	systemNote := dNote.Note{
		ID: uuid.New(), CreatedAt: time.Now().UTC(), Type: dNote.TypeUpdate,
		AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID, IsSystemLog: true,
	}
	if err := repo.CreateNote(ctx, testDB, systemNote); err != nil {
		t.Fatalf("CreateNote() error = %v, want nil", err)
	}

	// Regression test: the dynamic filter loop used to build
	// "AND n.col = 3" (a bare integer literal) instead of "AND n.col = $3"
	// (a bind placeholder), which either matched nothing or errored
	// outright depending on the column's type.
	got, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, []uuid.UUID{assetID}, map[string]string{"is_system_log": "true"})
	if err != nil {
		t.Fatalf("ListNotes() with field filter error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].ID != systemNote.ID {
		t.Errorf("ListNotes() with filter = %+v, want only %v", got, systemNote.ID)
	}
}

func TestNoteRepository_ListNotes_InvalidAssetID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)

	_, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, nil, nil)
	if !errors.Is(err, dNote.ErrInvalidAssetID) {
		t.Errorf("ListNotes() error = %v, want %v", err, dNote.ErrInvalidAssetID)
	}
}

func TestNoteRepository_CreateManyNotes(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)
	authorID := insertTestUser(t)
	assetID := uuid.New()

	notes := []dNote.Note{
		{ID: uuid.New(), CreatedAt: time.Now().UTC().Add(-time.Minute), Type: dNote.TypeUpdate, AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID, UpdateField: shared.Ptr("name")},
		{ID: uuid.New(), CreatedAt: time.Now().UTC(), Type: dNote.TypeUpdate, AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID, UpdateField: shared.Ptr("description")},
	}
	if err := repo.CreateManyNotes(ctx, testDB, notes); err != nil {
		t.Fatalf("CreateManyNotes() error = %v, want nil", err)
	}

	got, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, []uuid.UUID{assetID}, nil)
	if err != nil {
		t.Fatalf("ListNotes() error = %v, want nil", err)
	}
	// Only the most recent of the two survives DISTINCT ON.
	if len(got) != 1 || got[0].ID != notes[1].ID {
		t.Errorf("ListNotes() = %+v, want only the most recent note %v", got, notes[1].ID)
	}
}

// TestNoteRepository_CreateManyNotes_Empty matches a real call site:
// product.Service.UpdateProduct calls CreateManyNotes with an empty slice
// whenever a command has no changed fields.
func TestNoteRepository_CreateManyNotes_Empty(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)

	if err := repo.CreateManyNotes(ctx, testDB, []dNote.Note{}); err != nil {
		t.Errorf("CreateManyNotes(empty) error = %v, want nil", err)
	}
}

func TestNoteRepository_DeleteNotes(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewNoteRepository(testLogger)
	authorID := insertTestUser(t)
	assetID := uuid.New()

	note := dNote.Note{ID: uuid.New(), CreatedAt: time.Now().UTC(), Type: dNote.TypeUpdate, AssetType: shared.AssetProduct, AssetID: assetID, AuthorID: authorID}
	if err := repo.CreateNote(ctx, testDB, note); err != nil {
		t.Fatalf("CreateNote() error = %v, want nil", err)
	}

	if err := repo.DeleteNotes(ctx, testDB, []uuid.UUID{note.ID}); err != nil {
		t.Fatalf("DeleteNotes() error = %v, want nil", err)
	}

	got, err := repo.ListNotes(ctx, testDB, dNote.TypeUpdate, []uuid.UUID{assetID}, nil)
	if err != nil {
		t.Fatalf("ListNotes() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("ListNotes() after DeleteNotes = %+v, want empty", got)
	}
}
