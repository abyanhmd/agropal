package note

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "NOTE.NOT_FOUND"

	ErrMissingNoteID    shared.ErrorKey = "NOTE.MISSING_NOTE_ID"
	ErrInvalidNoteID    shared.ErrorKey = "NOTE.INVALID_NOTE_ID"
	ErrInvalidType      shared.ErrorKey = "NOTE.INVALID_TYPE"
	ErrInvalidAssetType shared.ErrorKey = "NOTE.INVALID_ASSET_TYPE"
	ErrInvalidAssetID   shared.ErrorKey = "NOTE.INVALID_ASSET_ID"
	ErrInvalidAuthorID  shared.ErrorKey = "NOTE.INVALID_AUTHOR_ID"

	ErrCreateFailed shared.ErrorKey = "NOTE.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "NOTE.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "NOTE.DELETE_FAILED"
)
