package transaction

import (
	aTransaction "agropal/internal/app/transaction"
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	IoHandler          *utils.IOHandler
	TransactionService *aTransaction.Service
}

func NewHandler(
	ioHandler *utils.IOHandler,
	transactionService *aTransaction.Service,
) Handler {
	return Handler{
		IoHandler:          ioHandler,
		TransactionService: transactionService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	var req request.CreateTransactionsRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.TransactionService.Create(ctx, farmID, productID, req.ToCommands()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusCreated, nil)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	transactionID, err := h.IoHandler.IdentifierFromPathValue(r, "transaction_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	result, err := h.TransactionService.GetOne(ctx, farmID, productID, transactionID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToTransactionResponse(result))
}

// GetMany lists transactions for a farm, or for a single product within
// that farm when product_id is present in the path.
func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	var result []dTransaction.Transaction
	if productIDStr := r.PathValue("product_id"); productIDStr != "" {
		productID, errParse := uuid.Parse(productIDStr)
		if errParse != nil {
			h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
			return
		}
		result, err = h.TransactionService.ListByProductID(ctx, farmID, productID)
	} else {
		result, err = h.TransactionService.ListByFarmID(ctx, farmID)
	}
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	transactionResponse := make([]response.TransactionResponse, len(result))
	for i, res := range result {
		transactionResponse[i] = DomainToTransactionResponse(res)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, transactionResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	transactionID, err := h.IoHandler.IdentifierFromPathValue(r, "transaction_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	var req request.TransactionRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.TransactionService.Update(ctx, farmID, productID, transactionID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

// Delete removes a set of transactions scoped to a farm/product, given a
// list of transaction IDs in the request body.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	var req request.DeleteTransactionsRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	transactionIDs, err := parseUUIDs(req.TransactionIDs)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	if err = h.TransactionService.Delete(ctx, farmID, productID, transactionIDs); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

// DeleteMany removes an arbitrary set of transactions the requester has
// access to, regardless of which farm/product they belong to.
func (h *Handler) DeleteMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.DeleteTransactionsRequest
	if err := h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	transactionIDs, err := parseUUIDs(req.TransactionIDs)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	if err = h.TransactionService.DeleteMany(ctx, transactionIDs); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

// DeleteAll removes every transaction across every farm the requester owns.
func (h *Handler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.TransactionService.DeleteAll(ctx); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	parsed := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		parsed[i] = u
	}
	return parsed, nil
}
