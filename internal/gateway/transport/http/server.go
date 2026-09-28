package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	api "quaztz/internal/api/gatewayprivate"
	"quaztz/internal/gateway/connection"
)

func NewHandler(registry *connection.Registry, token string) http.Handler {
	strict := api.NewStrictHandlerWithOptions(
		server{registry: registry},
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

type server struct {
	registry *connection.Registry
}

func (s server) DeliverMessage(
	_ context.Context,
	req api.DeliverMessageRequestObject,
) (api.DeliverMessageResponseObject, error) {
	m := req.Body

	if !messageValid(m) {
		return api.DeliverMessage400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.ValidationFailed,
				Message: "required fields are missing",
			},
		}, nil
	}

	s.registry.Deliver(m.RecipientId, connection.Message{
		ChatID:   m.ChatId,
		SenderID: m.SenderId,
		Seqnum:   m.Seqnum,
		Content:  m.Content,
		SentAt:   m.SentAt,
	})

	return api.DeliverMessage202Response{}, nil
}

func (s server) InvalidateSession(
	_ context.Context,
	req api.InvalidateSessionRequestObject,
) (api.InvalidateSessionResponseObject, error) {
	s.registry.CloseSession(req.SessionId)

	return api.InvalidateSession204Response{}, nil
}

func messageValid(m *api.ReceivedMessage) bool {
	return m.ChatId != uuid.Nil &&
		m.SenderId != uuid.Nil &&
		m.RecipientId != uuid.Nil &&
		m.Content != "" &&
		m.Seqnum > 0 &&
		!m.SentAt.IsZero()
}

func requireToken(token string, next http.Handler) http.Handler {
	desiredHeader := "Bearer " + token

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader := r.Header.Get("Authorization")

		if receivedHeader != desiredHeader {
			writeError(w, http.StatusUnauthorized, api.InvalidToken, "invalid service token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func badRequest(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, http.StatusBadRequest, api.ValidationFailed, err.Error())
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("private api", "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, api.Internal, "internal error")
}

func writeError(w http.ResponseWriter, status int, code api.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(api.Error{Code: code, Message: message})
}
