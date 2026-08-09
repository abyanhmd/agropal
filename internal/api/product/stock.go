package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"net/http"
)

func (h *Handler) CreateProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	productID, err := h.IoHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidID, nil)
		return
	}

	var req request.ProductStockRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if req.Stock == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidStockAmount, nil)
		return
	}
	if req.Type == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidStockType, nil)
		return
	}
	if req.DateAdded == nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.ProductService.CreateProductStock(ctx, farmID, productID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dProduct.SuccessStockAdded, nil)
}

func (h *Handler) DeleteProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stockID, err := h.IoHandler.IdentifierFromPathValue(r, "stock_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidStockID, nil)
		return
	}

	if err = h.ProductService.DeleteProductStock(ctx, stockID); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dProduct.SuccessStockDeleted, nil)
}
