package product

import (
	"agropal/internal/request"
	"agropal/internal/response"
	"net/http"
)

func (h *Handler) CreateProductStock(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductStockRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(productID)

	if err = h.ProductService.CreateProductStock(ctx, farmID, productID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) UpdateProductStock(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductStockRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(productID)

	if err = h.ProductService.UpdateProductStock(ctx, farmID, productID, priceID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) DeleteProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stockID, err := h.IoHandler.IdentifierFromPathValue(r, "stock_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	if err = h.ProductService.DeleteProductStock(ctx, stockID); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}
