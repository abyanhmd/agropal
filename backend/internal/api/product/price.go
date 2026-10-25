package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"net/http"
)

func (h *Handler) CreateProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	productID, err := h.ioHandler.IdentifierFromPathValue(r, "product_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidID, nil)
		return
	}

	var req request.ProductPriceRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.productService.CreateProductPrice(ctx, farmID, productID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessPriceAdded, nil)
}

func (h *Handler) DeleteProductPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	priceID, err := h.ioHandler.IdentifierFromPathValue(r, "price_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidPriceID, nil)
		return
	}

	if err = h.productService.DeleteProductPrice(ctx, priceID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessPriceDeleted, nil)
}
