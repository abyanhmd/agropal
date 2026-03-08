package post

import "agropal/internal/response"

var (
	ErrNotFound response.ErrorKey = "POST.NOT_FOUND"

	ErrInvalidID         response.ErrorKey = "POST.INVALID_ID"
	ErrInvalidTitle      response.ErrorKey = "POST.INVALID_TITLE"
	ErrInvalidContent    response.ErrorKey = "POST.INVALID_CONTENT"
	ErrInvalidSlug       response.ErrorKey = "POST.INVALID_SLUG"
	ErrInvalidAuthorID   response.ErrorKey = "POST.INVALID_AUTHOR_ID"
	ErrInvalidAssetType  response.ErrorKey = "POST.INVALID_ASSET_TYPE"
	ErrInvalidActionType response.ErrorKey = "POST.INVALID_ACTION_TYPE"

	ErrCreateFailed response.ErrorKey = "POST.CREATE_FAILED"
	ErrUpdateFailed response.ErrorKey = "POST.UPDATE_FAILED"
	ErrDeleteFailed response.ErrorKey = "POST.DELETE_FAILED"

	ErrCommentNotFound  response.ErrorKey = "POST.COMMENT_NOT_FOUND"
	ErrInvalidCommentID response.ErrorKey = "POST.INVALID_COMMENT_ID"
	ErrInvalidParentID  response.ErrorKey = "POST.INVALID_PARENT_ID"
)
