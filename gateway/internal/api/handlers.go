package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/lyagu5h/finScope/gateway/internal/delivery/client"
	ledgerv1 "github.com/lyagu5h/finScope/gateway/internal/delivery/protos/ledger/v1"
	"github.com/lyagu5h/finScope/gateway/internal/middleware"
)
type Handler struct {
	ledger *client.Client
	logger *slog.Logger
	timeout time.Duration
}

func NewHandler(ledger *client.Client, logger *slog.Logger, timeout time.Duration) *Handler {
	return &Handler{
		ledger: ledger,
		timeout: timeout,
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/transactions", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.transactionsHandler), h.logger), 
			h.timeout,
		),
	)
	mux.Handle("/api/budgets", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.budgetsHandler), h.logger), 
			h.timeout,
		),
	)
	mux.Handle("/api/reports/summary", middleware.Timeout(
			middleware.Logging(http.HandlerFunc(h.reportsSummaryHandler), h.logger), 
			h.timeout,
		),
	)
	mux.Handle(
		"/api/transactions/bulk",
		middleware.Timeout(
			middleware.Logging(
				http.HandlerFunc(h.bulkTransactions),
				h.logger,
			),
			h.timeout,
		),
	)
	mux.HandleFunc("/ping", h.ping)


}

func (h *Handler) ping(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
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

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to are required")
		return
	}

	res, err := h.ledger.Ledger().GetReportSummary(
		r.Context(),
		&ledgerv1.ReportSummaryRequest{
			From: from,
			To:   to,
		},
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res.Totals)
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.ledger.Ledger().AddTransaction(
		r.Context(),
		toProtoCreateTransaction(req),
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionDTOFromProto(res))
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	res, err := h.ledger.Ledger().ListTransactions(
		r.Context(),
		&emptypb.Empty{},
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	out := make([]TransactionResponse, 0, len(res.Transactions))
	for _, tx := range res.Transactions {
		out = append(out, toTransactionDTOFromProto(tx))
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) setBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.ledger.Ledger().SetBudget(
		r.Context(),
		toProtoCreateBudget(req),
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBudgetDTOFromProto(res))
}

func (h *Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	res, err := h.ledger.Ledger().ListBudgets(
		r.Context(),
		&emptypb.Empty{},
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	out := make([]BudgetResponse, 0, len(res.Budgets))
	for _, b := range res.Budgets {
		out = append(out, toBudgetDTOFromProto(b))
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) bulkTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req BulkCreateTransactionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Transactions) == 0 {
		writeError(w, http.StatusBadRequest, "transactions list is empty")
		return
	}

	res, err := h.ledger.Ledger().BulkAddTransactions(
		r.Context(),
		toProtoBulkCreateTransactions(req),
	)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBulkResponseDTO(res))
}
