package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}

func parseWorkers(r *http.Request, def int) int {
	v := r.URL.Query().Get("workers")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, st.Message())

	case codes.FailedPrecondition, codes.Aborted:
		writeError(w, http.StatusConflict, st.Message())

	case codes.DeadlineExceeded:
		writeError(w, http.StatusGatewayTimeout, "request timeout")

	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func timeToRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}