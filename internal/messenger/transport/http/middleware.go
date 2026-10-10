package http

import (
	"context"
	"errors"
	"net/http"
	api "quaztz/internal/api/messagepublic"
	"quaztz/internal/messenger/domain"
	"strings"
)

var authNotRequired = map[string]bool{
	"RegisterUser": true,
	"LoginSession": true,
	"RefreshToken": true,
}

type sessionKey struct{}

func authMiddleware(auth Authenticator) api.StrictMiddlewareFunc {
	return func(next api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		if authNotRequired[operationID] {
			return next
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
			token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			session, err := auth.Authenticate(ctx, token)

			switch {
			case errors.Is(err, domain.ErrTokenInvalid):
				unauthorized(w, r, api.AuthorizationFailed)
				return nil, nil
			case errors.Is(err, domain.ErrTokenExpired):
				unauthorized(w, r, api.SessionExpired)
				return nil, nil
			case errors.Is(err, domain.ErrSessionRevoked):
				unauthorized(w, r, api.SessionRevoked)
				return nil, nil
			case errors.Is(err, domain.ErrSuspendedUser):
				unauthorized(w, r, api.SuspendedUser)
				return nil, nil
			case err != nil:
				return nil, err
			}

			ctx = context.WithValue(ctx, sessionKey{}, session)
			return next(ctx, w, r, req)
		}
	}
}

func sessionFrom(ctx context.Context) domain.Session {
	return ctx.Value(sessionKey{}).(domain.Session)
}
