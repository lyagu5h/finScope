package api

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeGRPCError(w http.ResponseWriter, err error) {
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
