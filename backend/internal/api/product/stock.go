package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"net/http"
)

func (h *Handler) CreateProductStock(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductStockRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	stockID, err := h.productService.CreateProductStock(ctx, farmID, productID, req.ToCommand())
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessStockAdded, response.StockCreatedResponse{StockID: stockID})
}

func (h *Handler) DeleteProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stockID, err := h.ioHandler.IdentifierFromPathValue(r, "stock_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidStockID, nil)
		return
	}

	if err = h.productService.DeleteProductStock(ctx, stockID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessStockDeleted, nil)
}
