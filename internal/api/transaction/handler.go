package transaction

import (
	"agropal/internal/application/transaction"
	"agropal/internal/utils"
	"net/http"
)

type Handler struct {
	IoHandler          utils.IOHandler
	TransactionService transaction.Service
}

func NewHandler(
	ioHandler utils.IOHandler,
	transactionService transaction.Service,
) Handler {
	return Handler{
		IoHandler:          ioHandler,
		TransactionService: transactionService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {

}
