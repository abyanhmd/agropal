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

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	var req request.ProductRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if req.Name == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidName, nil)
		return
	}
	if req.Description == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidDescription, nil)
		return
	}
	if req.Image == nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidImage, nil)
		return
	}

	if err = h.ProductService.CreateProduct(ctx, farmID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusCreated, nil)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
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

	var req request.ProductRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.ProductService.UpdateProduct(ctx, farmID, productID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

func (h *Handler) GetOneProduct(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.ProductService.GetOneProduct(ctx, farmID, productID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToProductDetailResponse(result))
}

func (h *Handler) GetManyProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dProduct.ErrInvalidFarmID, nil)
		return
	}

	result, err := h.ProductService.GetManyProducts(ctx, farmID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	productResponse := make([]response.ProductDetailResponse, len(result))
	for i, res := range result {
		productResponse[i] = DomainToProductDetailResponse(res)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, productResponse)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
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

	if err = h.ProductService.DeleteProduct(ctx, farmID, productID); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}
