package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"net/http"
)

func (h *Handler) CreateProductPrice(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductPriceRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if req.Price == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidPriceAmount, nil)
		return
	}
	if req.DateAdded == nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.ProductService.CreateProductPrice(ctx, farmID, productID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dProduct.SuccessPriceAdded, nil)
}

func (h *Handler) DeleteProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	priceID, err := h.IoHandler.IdentifierFromPathValue(r, "price_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidPriceID, nil)
		return
	}

	if err = h.ProductService.DeleteProductPrice(ctx, priceID); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dProduct.SuccessPriceDeleted, nil)
}
