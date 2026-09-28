package private

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	api "quaztz/internal/api/messageprivate"
)

type server struct{}

func NewHandler(token string) http.Handler {
	strict := api.NewStrictHandlerWithOptions(
		server{},
		nil,
		api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  badRequest,
			ResponseErrorHandlerFunc: internalError,
		},
	)

	handler := api.HandlerWithOptions(strict, api.StdHTTPServerOptions{
		ErrorHandlerFunc: badRequest,
	})

	return requireToken(token, handler)
}

func requireToken(token string, next http.Handler) http.Handler {
	desiredHeader := []byte("Bearer " + token)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader := []byte(r.Header.Get("Authorization"))

		// защита от атак по времени
		if subtle.ConstantTimeCompare(receivedHeader, desiredHeader) != 1 {
			unauthorizedRequest(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s server) SendMessage(ctx context.Context, request api.SendMessageRequestObject) (api.SendMessageResponseObject, error) {
	panic("unimplemented")
}

func (s server) ValidateSession(ctx context.Context, request api.ValidateSessionRequestObject) (api.ValidateSessionResponseObject, error) {
	panic("unimplemented")
}

func badRequest(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, http.StatusBadRequest, api.InvalidRequest, err.Error())
}

func unauthorizedRequest(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, api.UnauthorizedRequest, "unauthorized")
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("private api error", "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, api.Internal, "internal error")
}

func writeError[C ~string](w http.ResponseWriter, status int, code C, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(struct {
		Code    C      `json:"code"`
		Message string `json:"message"`
	}{code, message})
}
