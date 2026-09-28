package http

import (
	"context"
	"crypto/subtle"
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
	msg := req.Body

	if !validateReceivedMessage(msg) {
		return api.DeliverMessage400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.ValidationFailed,
				Message: "invalid request",
			},
		}, nil
	}

	s.registry.Deliver(msg.RecipientId, connection.Message{
		ChatID:   msg.ChatId,
		SenderID: msg.SenderId,
		Seqnum:   msg.Seqnum,
		Content:  msg.Content,
		SentAt:   msg.SentAt,
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

func validateReceivedMessage(message *api.ReceivedMessage) bool {
	if message.ChatId == uuid.Nil || message.SenderId == uuid.Nil || message.RecipientId == uuid.Nil {
		return false
	}
	if len(message.Content) == 0 {
		return false
	}
	if message.Seqnum <= 0 || message.SentAt.IsZero() {
		return false
	}
	return true
}

func requireToken(token string, next http.Handler) http.Handler {
	desiredHeader := []byte("Bearer " + token)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader := []byte(r.Header.Get("Authorization"))

		// защита от атак по времени
		if subtle.ConstantTimeCompare(receivedHeader, desiredHeader) != 1 {
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
	slog.Error("private api error", "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, api.Internal, "internal error")
}

func writeError(w http.ResponseWriter, status int, code api.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(api.Error{Code: code, Message: message})
}
