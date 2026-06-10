package persistence

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type noteRepo struct {
	logger *pkg.Logger
}

func NewNoteRepository(l *pkg.Logger) dNote.Repository {
	return &noteRepo{logger: l}
}

func (r *noteRepo) CreateNote(ctx context.Context, q database.Querier, note dNote.Note) error {
	query := `
		INSERT INTO notes (
			note_id, created_at, type,
			asset_type, asset_id, field,
			author_id, is_system_log
		) VALUES (
			:note_id, :created_at, :type,
			:asset_type, :asset_id, :field,
			:author_id, :is_system_log
		);`

	noteModel := ToNoteModel(note)
	if _, err := q.NamedExecContext(ctx, query, noteModel); err != nil {
		r.logger.Error(ctx, shared.System, err.Error())
		return err
	}

	return nil
}

func (r *noteRepo) CreateManyNotes(ctx context.Context, q database.Querier, notes []dNote.Note) error {
	query := `
		INSERT INTO notes (
			note_id, created_at, type,
			asset_type, asset_id, field,
			author_id, is_system_log
		) VALUES (
			:note_id, :created_at, :type,
			:asset_type, :asset_id, :field,
			:author_id, :is_system_log
		);`

	noteModels := make([]NoteModel, len(notes))
	for i, n := range notes {
		noteModels[i] = ToNoteModel(n)
	}
	if _, err := q.NamedExecContext(ctx, query, noteModels); err != nil {
		r.logger.Error(ctx, shared.System, err.Error())
		return err
	}

	return nil
}

func (r *noteRepo) ListNotes(ctx context.Context, q database.Querier, noteType dNote.NoteType, assetIDs []uuid.UUID, fields map[string]string) ([]dNote.Note, error) {
	if len(assetIDs) == 0 {
		return nil, dNote.ErrInvalidAssetID
	}

	query := `
		SELECT DISTINCT ON (n.asset_id) 
			n.note_id, n.created_at, n.type,
			n.asset_type, n.asset_id, n.field,
			n.author_id, u.name AS author_name, 
			n.is_system_log
		FROM notes n
		JOIN users u ON n.author_id = u.user_id
		WHERE n.asset_id = $1
			AND n.type = $2
		`

	args := []interface{}{pq.Array(assetIDs), string(noteType)}

	for colName, colValue := range fields {
		if len(colValue) == 0 {
			continue
		}
		query += fmt.Sprintf(" AND n.%s = %d", pq.QuoteIdentifier(colName), len(args))
		args = append(args, colValue)
	}

	query += ` ORDER BY n.asset_id, n.created_at DESC;`

	var modelNotes []NoteModel
	if err := q.SelectContext(ctx, &modelNotes, query, args...); err != nil {
		if len(modelNotes) == 0 {
			return nil, dNote.ErrNotFound
		}
		r.logger.Error(ctx, shared.System, err.Error())
		return nil, err
	}

	notes := make([]dNote.Note, len(modelNotes))
	for i, m := range modelNotes {
		notes[i] = m.ToDomain()
	}
	return notes, nil
}

func (r *noteRepo) DeleteNotes(ctx context.Context, q database.Querier, noteIDs []uuid.UUID) error {
	query := `DELETE FROM notes WHERE note_id = ANY($1);`

	if _, err := q.ExecContext(ctx, query, pq.Array(noteIDs)); err != nil {
		r.logger.Error(ctx, shared.System, err.Error())
		return err
	}

	return nil
}
