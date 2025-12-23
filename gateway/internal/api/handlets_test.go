package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBudgetsHandler_MethodNotAllowed(t *testing.T) {
	h := &Handler{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/budgets", h.budgetsHandler)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/budgets",
		strings.NewReader(`{}`),
	)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "method not allowed") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestSetBudget_InvalidJSON(t *testing.T) {
	h := &Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/budgets", h.budgetsHandler)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/budgets",
		strings.NewReader(`{invalid-json}`),
	)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTransactionsHandler_MethodNotAllowed(t *testing.T) {
	h := &Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", h.transactionsHandler)

	req := httptest.NewRequest(http.MethodPut, "/api/transactions", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestCreateTransaction_InvalidJSON(t *testing.T) {
	h := &Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", h.transactionsHandler)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/transactions",
		strings.NewReader(`{`),
	)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestReportsSummary_MissingParams(t *testing.T) {
	h := &Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/reports/summary", h.reportsSummaryHandler)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/reports/summary",
		nil,
	)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBulkTransactions_EmptyList(t *testing.T) {
	h := &Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions/bulk", h.bulkTransactions)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/transactions/bulk",
		strings.NewReader(`{"transactions":[]}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
