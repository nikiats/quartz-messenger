package private

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	api "quaztz/internal/api/messageprivate"
	"quaztz/internal/messenger/domain"
	"time"

	"github.com/oapi-codegen/runtime/types"
)

type server struct {
	m Messages
	a Authenticator
}

func NewHandler(token string, m Messages, a Authenticator) http.Handler {
	strict := api.NewStrictHandlerWithOptions(
		server{m, a},
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

		if subtle.ConstantTimeCompare(receivedHeader, desiredHeader) != 1 {
			unauthorizedRequest(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s server) SendMessage(ctx context.Context, req api.SendMessageRequestObject) (api.SendMessageResponseObject, error) {
	send := req.Body

	session := domain.Session{
		SessionID: send.SessionId,
		UserID:    send.SenderId,
	}
	message, err := s.m.Send(ctx, session, send.ChatId, send.ClientMessageId, send.Content)
	if err != nil {
		return api.SendMessage200JSONResponse{
			Success: false,
			Error: &struct {
				Code    api.SentMessageResponseErrorCode `json:"code"`
				Message string                           `json:"message"`
			}{
				Code:    api.SendFailed,
				Message: "message send failed",
			},
		}, nil
	}

	return api.SendMessage200JSONResponse{
		Success: true,
		Message: &struct {
			ChatId types.UUID "json:\"chatId\""
			SentAt time.Time  "json:\"sentAt\""
			Seqnum int64      "json:\"seqnum\""
		}{
			ChatId: message.ChatID,
			SentAt: message.SentAt,
			Seqnum: message.Seqnum,
		},
	}, nil
}

func (s server) ValidateSession(ctx context.Context, req api.ValidateSessionRequestObject) (api.ValidateSessionResponseObject, error) {
	token := req.Body.Token

	session, err := s.a.Authenticate(ctx, token)
	if err != nil {
		return api.ValidateSession200JSONResponse{
			Success: false,
			Error: &struct {
				Code    api.ValidateResponseErrorCode `json:"code"`
				Message string                        `json:"message"`
			}{
				api.InvalidToken,
				"incorrect token",
			},
		}, nil
	}

	return api.ValidateSession200JSONResponse{
		Session: &struct {
			ExpiresAt time.Time  `json:"expiresAt"`
			SessionId types.UUID `json:"sessionId"`
			UserId    types.UUID `json:"userId"`
		}{
			ExpiresAt: session.ExpiresAt,
			SessionId: session.SessionID,
			UserId:    session.UserID,
		},
	}, nil
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

	_ = json.NewEncoder(w).Encode(
		struct {
			Code    C      `json:"code"`
			Message string `json:"message"`
		}{code, message},
	)
}
