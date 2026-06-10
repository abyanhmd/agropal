package shared

import (
	dAuth "agropal/internal/domain/auth"
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const (
	KeyMethod    ContextKey = "method"
	KeyURLPath   ContextKey = "url_path"
	KeyUserAgent ContextKey = "user_agent"
	KeyIPAddress ContextKey = "ip_address"
	KeyRequestID ContextKey = "request_id"

	KeyUserID ContextKey = "user_id"
	KeyEmail  ContextKey = "email"
	KeyRole   ContextKey = "role"
)

func WithMetadata(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetMetadata(ctx context.Context, key ContextKey) string {
	val, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}

	return val
}

func SaveContextFromRequest(r *http.Request, claims *dAuth.ClaimInfo) context.Context {
	ctx := r.Context()

	ctx = WithMetadata(ctx, KeyMethod, r.Method)
	ctx = WithMetadata(ctx, KeyURLPath, r.URL.Path)
	ctx = WithMetadata(ctx, KeyUserAgent, r.UserAgent())
	ctx = WithMetadata(ctx, KeyIPAddress, r.RemoteAddr)

	if claims != nil {
		ctx = WithMetadata(ctx, KeyUserID, claims.UserID.String())
		ctx = WithMetadata(ctx, KeyEmail, claims.Email)
		ctx = WithMetadata(ctx, KeyRole, claims.Role)
	}

	return ctx
}

func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	userIDStr := GetMetadata(ctx, KeyUserID)
	if userIDStr == "" {
		return uuid.Nil, errors.New("invalid user id")
	}

	userID, errParse := uuid.Parse(userIDStr)
	if errParse != nil {
		return uuid.Nil, errors.New("invalid user id")
	}

	return userID, nil
}

func Ptr[T any](v T) *T {
	return &v
}

func Str(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
