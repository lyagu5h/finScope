package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/lyagu5h/finScope/gateway/internal/middleware"
	app "github.com/lyagu5h/finScope/ledger/pkg/ledger"
)
type Handler struct {
	App app.LedgerService
	Logger *slog.Logger
	Timeout time.Duration
}

func NewHandler(app app.LedgerService, logger *slog.Logger, timeout time.Duration) *Handler {
	return &Handler{
		App: app,
		Logger: logger,
		Timeout: timeout,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.Handle("/api/transactions", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.transactionsHandler), logger), 
			h.Timeout,
		),
	)
	mux.Handle("/api/budgets", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.budgetsHandler), logger), 
			h.Timeout,
		),
	)
	mux.Handle("/api/reports/summary", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.reportsSummaryHandler), logger), 
			h.Timeout,
		),
	)
	mux.Handle(
		"/api/transactions/bulk",
		middleware.Timeout(
			middleware.Logging(
				http.HandlerFunc(h.bulkTransactions),
				logger,
			),
			h.Timeout,
		),
	)


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

func (h *Handler) reportsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "from and to parameters are required")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from date format")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to date format")
		return
	}

	if to.Before(from) {
		writeError(w, http.StatusBadRequest, "`to` must be after or equal to `from`")
		return
	}
	
	h.Logger.Info(
		"report summary request",
		slog.String("from", fromStr),
		slog.String("to", toStr),
	)

	summary, err := h.App.GetReportSummary(r.Context(), from, to)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, summary)
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

	createdTx, err := h.App.AddTransaction(r.Context(), tx); 
	if err != nil {
		if errors.Is(err, app.ErrBudgetExceeded) {
			writeError(w, http.StatusConflict, "budget exceeded")
			return
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionDTO(createdTx))
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	txs, err := h.App.ListTransactions(r.Context())

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
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

	if err := h.App.SetBudget(r.Context(), b); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toBudgetDTO(b))
}

func (h *Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	budgets, err := h.App.ListBudgets(r.Context())

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	res := make([]BudgetResponse, 0, len(budgets))

	for _, b := range budgets {
		res = append(res, toBudgetDTO(b))
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) bulkTransactions(w http.ResponseWriter, r *http.Request) {
	var req []CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workers := parseWorkers(r, 4)

	txs := make([]app.Transaction, 0, len(req))
	for _, item := range req {
		txs = append(txs, toTransactionLedger(item))
	}

	result, err := h.App.ImportTransactions(r.Context(), txs, workers)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
