package product

import (
	"agropal/internal/request"
	"agropal/internal/response"
	"net/http"
)

func (h *Handler) CreateProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.ProductPriceRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(productID)

	if err = h.ProductService.CreateProductPrice(ctx, farmID, productID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) UpdateProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	priceID, err := h.IoHandler.IdentifierFromPathValue(r, "price_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.ProductPriceRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(productID)

	if err = h.ProductService.UpdateProductPrice(ctx, farmID, productID, priceID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) DeleteProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	priceID, err := h.IoHandler.IdentifierFromPathValue(r, "price_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	if err = h.ProductService.DeleteProductPrice(ctx, priceID); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}
