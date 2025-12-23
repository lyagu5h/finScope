package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/lyagu5h/finScope/gateway/internal/delivery/client"
	ledgerv1 "github.com/lyagu5h/finScope/gateway/internal/delivery/protos/ledger/v1"
	"github.com/lyagu5h/finScope/gateway/internal/middleware"

	_ "github.com/lyagu5h/finScope/gateway/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	ledger  *client.Client
	logger  *slog.Logger
	timeout time.Duration
}

func NewHandler(ledger *client.Client, logger *slog.Logger, timeout time.Duration) *Handler {
	return &Handler{
		ledger:  ledger,
		timeout: timeout,
		logger:  logger,
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
	mux.Handle("/swagger/",
		httpSwagger.WrapHandler,
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
	mux.Handle(
		"/api/transactions/export.csv",
		middleware.Timeout(
			middleware.Logging(
				http.HandlerFunc(h.exportTransactionsCSV),
				h.logger,
			),
			h.timeout,
		),
	)
	mux.HandleFunc("/ping", h.ping)

}

// Ping godoc
// @Summary Health check
// @Tags system
// @Produce plain
// @Success 200 {string} string "pong"
// @Router /ping [get]
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

// ReportSummary godoc
// @Summary Get report summary
// @Tags reports
// @Produce json
// @Param from query string true "From date (YYYY-MM-DD)"
// @Param to query string true "To date (YYYY-MM-DD)"
// @Success 200 {object} map[string]float64
// @Failure 400 {object} ErrorResponse
// @Router /api/reports/summary [get]
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

// CreateTransaction godoc
// @Summary Create transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction payload"
// @Success 201 {object} TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/transactions [post]
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

// ListTransactions godoc
// @Summary List transactions
// @Tags transactions
// @Produce json
// @Success 200 {array} TransactionResponse
// @Router /api/transactions [get]
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

// SetBudget godoc
// @Summary Set budget
// @Tags budgets
// @Accept json
// @Produce json
// @Param budget body CreateBudgetRequest true "Budget payload"
// @Success 201 {object} BudgetResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/budgets [post]
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

// ListBudgets godoc
// @Summary List budgets
// @Tags budgets
// @Produce json
// @Success 200 {array} BudgetResponse
// @Router /api/budgets [get]
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

// BulkCreateTransactions godoc
// @Summary Bulk create transactions
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body BulkCreateTransactionsRequest true "Bulk transactions"
// @Success 200 {object} BulkCreateTransactionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/transactions/bulk [post]
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

// ExportTransactionsCSV godoc
// @Summary Export transactions to CSV
// @Tags transactions
// @Produce text/csv
// @Success 200 {string} string "CSV file"
// @Router /api/transactions/export.csv [get]
func (h *Handler) exportTransactionsCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()

	resp, err := h.ledger.Ledger().ListTransactions(ctx, &emptypb.Empty{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="transactions.csv"`)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{
		"id",
		"date",
		"category",
		"amount",
		"description",
	})

	for _, tx := range resp.Transactions {
		_ = writer.Write([]string{
			strconv.FormatInt(tx.Id, 10),
			tx.Date.AsTime().Format("2006-01-02"),
			tx.Category,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.Description,
		})
	}
}
