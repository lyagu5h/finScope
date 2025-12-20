package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lyagu5h/finScope/gateway/internal/middleware"
	"github.com/lyagu5h/finScope/ledger/pkg/ledger"
)

type Handler struct {
	Store *ledger.Store
}

func NewHandler(store *ledger.Store) *Handler {
	return &Handler{
		Store: store, 
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/transactions", middleware.Logging(http.HandlerFunc(h.transactionsHandler)))
	mux.Handle("/api/budgets", middleware.Logging(http.HandlerFunc(h.budgetsHandler)))

}

func (h *Handler) transactionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTransactions(w, r)
	case http.MethodPost:
		h.createTransaction(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}


func (h *Handler) budgetsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBudgets(w, r)
	case http.MethodPost:
		h.setBudget(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx := toTransactionLedger(req)

	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Store.AddTransaction(tx); err != nil {
		if errors.Is(err, ledger.ErrBudgetExceeded) {
			writeError(w, http.StatusConflict, "budget exceeded")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionDTO(tx))
}

func (h *Handler) listTransactions(w http.ResponseWriter, _ *http.Request) {
	txs, err := h.Store.ListTransactions()

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	res := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		res = append(res, toTransactionDTO(tx))
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) setBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b := toBudgetLedger(req)

	if err := b.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Store.SetBudget(b); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, toBudgetDTO(b))
}

func (h *Handler) listBudgets(w http.ResponseWriter, _ *http.Request) {
	budgets, err := h.Store.ListBudgets()

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	res := make([]BudgetResponse, 0, len(budgets))

	for _, b := range budgets {
		res = append(res, toBudgetDTO(b))
	}

	writeJSON(w, http.StatusOK, res)
}