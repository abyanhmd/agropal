package product

import (
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/request"

	"net/http"
)

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	var req request.ProductRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}
	cmd := req.ToCommand(farmID)

	if err = h.ProductService.CreateProduct(ctx, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	var req request.ProductRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}
	cmd := req.ToCommand(farmID)

	if err = h.ProductService.UpdateProduct(ctx, productID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) GetOneProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	result, err := h.ProductService.GetOneProduct(ctx, productID)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, result)
}

func (h *Handler) GetManyProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	result, err := h.ProductService.GetManyProducts(ctx, farmID)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, result)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	if err = h.ProductService.DeleteProduct(ctx, productID); err != nil {
		h.IoHandler.WriteError(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
}
