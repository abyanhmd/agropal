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
	ioHandler          *utils.IOHandler
	transactionService *aTransaction.Service
}

func NewHandler(
	ioHandler *utils.IOHandler,
	transactionService *aTransaction.Service,
) Handler {
	return Handler{
		ioHandler:          ioHandler,
		transactionService: transactionService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	productID, err := h.ioHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	var req request.CreateTransactionsRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.transactionService.Create(ctx, farmID, productID, req.ToCommands()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dTransaction.SuccessCreated, nil)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.ioHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	transactionID, err := h.ioHandler.IdentifierFromPathValue(r, "transaction_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	result, err := h.transactionService.GetOne(ctx, farmID, productID, transactionID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToTransactionResponse(result))
}

// GetMany lists transactions for a farm, or for a single product within
// that farm when product_id is present in the path.
func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	var result []dTransaction.Transaction
	if productIDStr := r.PathValue("product_id"); productIDStr != "" {
		productID, errParse := uuid.Parse(productIDStr)
		if errParse != nil {
			h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
			return
		}
		result, err = h.transactionService.ListByProductID(ctx, farmID, productID)
	} else {
		result, err = h.transactionService.ListByFarmID(ctx, farmID)
	}
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	transactionResponse := make([]response.TransactionResponse, len(result))
	for i, res := range result {
		transactionResponse[i] = DomainToTransactionResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, transactionResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.ioHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	transactionID, err := h.ioHandler.IdentifierFromPathValue(r, "transaction_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	var req request.TransactionRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.transactionService.Update(ctx, farmID, productID, transactionID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dTransaction.SuccessUpdated, nil)
}

// Delete removes a set of transactions scoped to a farm/product, given a
// list of transaction IDs in the request body.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	productID, err := h.ioHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidProductID, nil)
		return
	}

	var req request.DeleteTransactionsRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	transactionIDs, err := parseUUIDs(req.TransactionIDs)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	if err = h.transactionService.Delete(ctx, farmID, productID, transactionIDs); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dTransaction.SuccessDeleted, nil)
}

// DeleteMany removes an arbitrary set of transactions the requester has
// access to, regardless of which farm/product they belong to.
func (h *Handler) DeleteMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.DeleteTransactionsRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	transactionIDs, err := parseUUIDs(req.TransactionIDs)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dTransaction.ErrInvalidID, nil)
		return
	}

	if err = h.transactionService.DeleteMany(ctx, transactionIDs); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dTransaction.SuccessDeleted, nil)
}

// DeleteAll removes every transaction across every farm the requester owns.
func (h *Handler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.transactionService.DeleteAll(ctx); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dTransaction.SuccessDeleted, nil)
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
