package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type ContextType string

const (
	ContextMethod    ContextType = "method"
	ContextURLPath   ContextType = "url_path"
	ContextUserAgent ContextType = "user_agent"
	ContextIPAddress ContextType = "ip_address"
	ContextRequestID ContextType = "request_id"

	ContextUserID ContextType = "user_id"
	ContextEmail  ContextType = "email"
	ContextRole   ContextType = "role"
)

func WithMetadata(ctx context.Context, key ContextType, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetMetadata(ctx context.Context, key ContextType) string {
	val, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}

	return val
}

func SaveContextFromRequest(r *http.Request, claims *Claims) context.Context {
	ctx := r.Context()

	ctx = WithMetadata(ctx, ContextMethod, r.Method)
	ctx = WithMetadata(ctx, ContextURLPath, r.URL.Path)
	ctx = WithMetadata(ctx, ContextUserAgent, r.UserAgent())
	ctx = WithMetadata(ctx, ContextIPAddress, r.RemoteAddr)

	if claims != nil {
		ctx = WithMetadata(ctx, ContextUserID, claims.UserID)
		ctx = WithMetadata(ctx, ContextEmail, claims.Email)
		ctx = WithMetadata(ctx, ContextRole, claims.Role)
	}

	return ctx
}

func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	userIDStr := GetMetadata(ctx, ContextUserID)
	if userIDStr == "" {
		return uuid.Nil, errors.New("invalid user id")
	}

	userID, errParse := uuid.Parse(userIDStr)
	if errParse != nil {
		return uuid.Nil, errors.New("invalid user id")
	}

	return userID, nil
}
