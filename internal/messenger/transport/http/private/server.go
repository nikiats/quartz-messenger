package private

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	api "quaztz/internal/api/messageprivate"
	"quaztz/internal/messenger/domain"
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
	content := domain.MessageContent{
		Text: send.Content.Text,
	}
	message, err := s.m.Send(ctx, session, send.ChatId, send.ClientMessageId, content)

	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		return api.SendMessage200JSONResponse{
			Success: false,
			Error: &api.SendMessageError{
				Code:    api.SendFailed,
				Message: "chat not found",
			},
		}, nil
	case errors.Is(err, domain.ErrSessionRevoked):
		return api.SendMessage200JSONResponse{
			Success: false,
			Error: &api.SendMessageError{
				Code:    api.SendFailed,
				Message: "session revoked",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.SendMessage200JSONResponse{
		Success: true,
		Message: &api.SentMessage{
			ChatId: message.ChatID,
			SentAt: message.SentAt,
			Seqnum: message.Seqnum,
		},
	}, nil
}

func (s server) ValidateSession(ctx context.Context, req api.ValidateSessionRequestObject) (api.ValidateSessionResponseObject, error) {
	token := req.Body.Token
	session, err := s.a.Authenticate(ctx, token)

	switch {
	case errors.Is(err, domain.ErrTokenInvalid):
		return api.ValidateSession200JSONResponse{
			Success: false,
			Error: &api.ValidateSessionError{
				Code:    api.AuthorizationFailed,
				Message: "token invalid",
			},
		}, nil
	case errors.Is(err, domain.ErrTokenExpired):
		return api.ValidateSession200JSONResponse{
			Success: false,
			Error: &api.ValidateSessionError{
				Code:    api.SessionExpired,
				Message: "token expired",
			},
		}, nil
	case errors.Is(err, domain.ErrSessionRevoked):
		return api.ValidateSession200JSONResponse{
			Success: false,
			Error: &api.ValidateSessionError{
				Code:    api.SessionRevoked,
				Message: "session revoked",
			},
		}, nil
	case errors.Is(err, domain.ErrSuspendedUser):
		return api.ValidateSession200JSONResponse{
			Success: false,
			Error: &api.ValidateSessionError{
				Code:    api.SuspendedUser,
				Message: "account is suspended",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.ValidateSession200JSONResponse{
		Success: true,
		Session: &api.SessionInfo{
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
