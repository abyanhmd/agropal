package utils

import (
	"agropal/internal/domain/shared"
	"agropal/internal/response"
	"agropal/pkg"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type IOHandler struct {
	Logger    *pkg.Logger
	Registry  *response.ResponseRegistry
	Validator *Validator
}

func NewIOHandler(
	logger *pkg.Logger,
	registry *response.ResponseRegistry,
	validator *Validator,
) *IOHandler {
	return &IOHandler{
		Logger:    logger,
		Registry:  registry,
		Validator: validator,
	}
}

type InfoEnvelope struct {
	Code    string `json:"code"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

type ValidationErrors map[string][]string

type StandardEnvelope struct {
	Info      InfoEnvelope     `json:"info"`
	Data      any              `json:"data,omitempty"`
	Errors    ValidationErrors `json:"errors,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
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
func (io *IOHandler) WriteSuccess(ctx context.Context, w http.ResponseWriter, key shared.SuccessKey, data any) {
	detail, ok := response.SuccessRegistry[key]
	if !ok {
		//io.handleMissingKey(ctx, w, key, data)
		return
	}

	io.respond(ctx, w, detail.HTTPCode, StandardEnvelope{
		Info: InfoEnvelope{
			Code:    detail.ResCode,
			Key:     key.String(),
			Message: detail.Message,
		},
		Data: data,
	})
}

// WriteError ensures that errors are logged before being sent to the client
func (io *IOHandler) WriteError(ctx context.Context, w http.ResponseWriter, key shared.ErrorKey, errorFields ValidationErrors) {
	detail, ok := response.ErrorRegistry[key]
	if !ok {
		detail = response.ResponseDetail{ResCode: "INTERNAL_ERROR", HTTPCode: 500, Message: "An unexpected error occurred"}
	}

	requestID := shared.GetMetadata(ctx, shared.KeyRequestID)

	io.respond(ctx, w, detail.HTTPCode, StandardEnvelope{
		Info: InfoEnvelope{
			Code:    detail.ResCode,
			Key:     key.Error(),
			Message: detail.Message,
		},
		Errors:    errorFields,
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

// ValidateStruct runs struct-tag validation on s. If validation fails it
// writes a 400 response with per-field messages and returns false; callers
// should stop handling the request in that case.
func (io *IOHandler) ValidateStruct(ctx context.Context, w http.ResponseWriter, s any) bool {
	if errs := io.Validator.ValidateStruct(s); errs != nil {
		io.WriteError(ctx, w, shared.ErrValidation, errs)
		return false
	}
	return true
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

func (io *IOHandler) HandleServiceError(ctx context.Context, w http.ResponseWriter, key error) {
	var envelope StandardEnvelope

	requestID := shared.GetMetadata(ctx, shared.KeyRequestID)

	detail, ok := response.ErrorRegistry[shared.ErrorKey(key.Error())]
	if !ok {
		envelope = StandardEnvelope{
			Info: InfoEnvelope{
				Code:    "1002",
				Key:     "INTERNAL_SERVER_ERROR",
				Message: "An unexpected error occurred. Please try again later.",
			},
			RequestID: requestID,
		}
	} else {
		envelope = StandardEnvelope{
			Info: InfoEnvelope{
				Code:    detail.ResCode,
				Key:     key.Error(),
				Message: detail.Message,
			},
			RequestID: requestID,
		}
	}

	io.respond(ctx, w, detail.HTTPCode, envelope)
}
