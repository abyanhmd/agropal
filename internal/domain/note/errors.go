package note

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "NOTE.NOT_FOUND"

	ErrMissingNoteID    response.ErrorKey = "NOTE.MISSING_NOTE_ID"
	ErrInvalidNoteID    response.ErrorKey = "NOTE.INVALID_NOTE_ID"
	ErrInvalidType      response.ErrorKey = "NOTE.INVALID_TYPE"
	ErrInvalidAssetType response.ErrorKey = "NOTE.INVALID_ASSET_TYPE"
	ErrInvalidAssetID   response.ErrorKey = "NOTE.INVALID_ASSET_ID"
	ErrInvalidAuthorID  response.ErrorKey = "NOTE.INVALID_AUTHOR_ID"

	ErrCreateFailed response.ErrorKey = "NOTE.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "NOTE.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "NOTE.DELETE_FAILED"
)
