package post

import (
	"agropal/internal/domain/shared"
)

var (
	ErrNotFound shared.ErrorKey = "POST.NOT_FOUND"

	ErrInvalidID         shared.ErrorKey = "POST.INVALID_ID"
	ErrInvalidTitle      shared.ErrorKey = "POST.INVALID_TITLE"
	ErrInvalidContent    shared.ErrorKey = "POST.INVALID_CONTENT"
	ErrInvalidSlug       shared.ErrorKey = "POST.INVALID_SLUG"
	ErrInvalidAuthorID   shared.ErrorKey = "POST.INVALID_AUTHOR_ID"
	ErrInvalidAssetType  shared.ErrorKey = "POST.INVALID_ASSET_TYPE"
	ErrInvalidActionType shared.ErrorKey = "POST.INVALID_ACTION_TYPE"

	ErrCreateFailed shared.ErrorKey = "POST.CREATE_FAILED"
	ErrUpdateFailed shared.ErrorKey = "POST.UPDATE_FAILED"
	ErrDeleteFailed shared.ErrorKey = "POST.DELETE_FAILED"

	ErrCommentNotFound  shared.ErrorKey = "POST.COMMENT_NOT_FOUND"
	ErrInvalidCommentID shared.ErrorKey = "POST.INVALID_COMMENT_ID"
	ErrInvalidParentID  shared.ErrorKey = "POST.INVALID_PARENT_ID"
)
