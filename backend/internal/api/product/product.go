package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"net/http"
)

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	var req request.ProductRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.productService.CreateProduct(ctx, farmID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessCreated, nil)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.productService.UpdateProduct(ctx, farmID, productID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessUpdated, nil)
}

func (h *Handler) GetOneProduct(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.productService.GetOneProduct(ctx, farmID, productID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToProductDetailResponse(result))
}

func (h *Handler) GetManyProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	result, err := h.productService.GetManyProducts(ctx, farmID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	productResponse := make([]response.ProductDetailResponse, len(result))
	for i, res := range result {
		productResponse[i] = DomainToProductDetailResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, productResponse)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
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

	if err = h.productService.DeleteProduct(ctx, farmID, productID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dProduct.SuccessDeleted, nil)
}
