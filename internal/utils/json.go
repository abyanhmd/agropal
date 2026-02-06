package utils

import (
	"agropal/internal/domain/shared"
	"agropal/internal/response"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type IOHandler struct {
	Logger   *Logger
	Registry *response.ResponseRegistry
}

func NewIOHandler(
	logger *Logger,
	registry *response.ResponseRegistry,
) *IOHandler {
	return &IOHandler{
		Logger:   logger,
		Registry: registry,
	}
}

type StandardEnvelope[T any] struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (io *IOHandler) ReadRequestBody(r *http.Request, result interface{}) error {
	// Max size is 1MB, written in Binary Bytes
	r.Body = http.MaxBytesReader(nil, r.Body, 1048576)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(result); err != nil {
		return err
	}

	return nil
}

func (io *IOHandler) WriteResponseBody(w http.ResponseWriter, statusCode int, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}

// WriteSuccess handles successful API responses with generics
func (io *IOHandler) WriteSuccess(ctx context.Context, w http.ResponseWriter, key string, data any) {
	detail, ok := io.Registry.GetSuccessRegistry(key)
	if !ok {
		//io.handleMissingKey(ctx, w, key, data)
		return
	}

	io.respond(ctx, w, detail.HTTPCode, StandardEnvelope[any]{
		Code:    detail.ID,
		Message: detail.Message,
		Data:    data,
	})
}

// WriteError ensures that errors are logged before being sent to the client
func (io *IOHandler) WriteError(ctx context.Context, w http.ResponseWriter, key string, internalErr error) {
	detail, ok := io.Registry.GetErrorRegistry(key)
	if !ok {
		detail = response.ResponseDetail{ID: "INTERNAL_ERROR", HTTPCode: 500, Message: "An unexpected error occurred"}
	}

	requestID := GetMetadata(ctx, ContextRequestID)

	io.Logger.Error(ctx, shared.System, internalErr.Error())
	io.respond(ctx, w, detail.HTTPCode, StandardEnvelope[any]{
		Code:      detail.ID,
		Message:   detail.Message,
		RequestID: requestID,
	})
}

func (io *IOHandler) respond(ctx context.Context, w http.ResponseWriter, status int, envelope any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		io.Logger.Error(ctx, shared.System, err.Error())
	}
}

func (io *IOHandler) IdentifierFromPathValue(r *http.Request, key string) (uuid.UUID, error) {
	idStr := r.PathValue(key)
	if idStr == "" {
		return uuid.Nil, nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
