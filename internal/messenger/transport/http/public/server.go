package public

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	api "quaztz/internal/api/messagepublic"
)

type server struct {
}

func NewHandler(jwtSecret string) http.Handler {
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

	return handler
}

// AddChatMember implements [messagepublic.StrictServerInterface].
func (s server) AddChatMember(ctx context.Context, request api.AddChatMemberRequestObject) (api.AddChatMemberResponseObject, error) {
	panic("unimplemented")
}

// CreateChat implements [messagepublic.StrictServerInterface].
func (s server) CreateChat(ctx context.Context, request api.CreateChatRequestObject) (api.CreateChatResponseObject, error) {
	panic("unimplemented")
}

// DeleteChat implements [messagepublic.StrictServerInterface].
func (s server) DeleteChat(ctx context.Context, request api.DeleteChatRequestObject) (api.DeleteChatResponseObject, error) {
	panic("unimplemented")
}

// DeleteChatMember implements [messagepublic.StrictServerInterface].
func (s server) DeleteChatMember(ctx context.Context, request api.DeleteChatMemberRequestObject) (api.DeleteChatMemberResponseObject, error) {
	panic("unimplemented")
}

// EditProfile implements [messagepublic.StrictServerInterface].
func (s server) EditProfile(ctx context.Context, request api.EditProfileRequestObject) (api.EditProfileResponseObject, error) {
	panic("unimplemented")
}

// GetChatInfo implements [messagepublic.StrictServerInterface].
func (s server) GetChatInfo(ctx context.Context, request api.GetChatInfoRequestObject) (api.GetChatInfoResponseObject, error) {
	panic("unimplemented")
}

// GetChatMembers implements [messagepublic.StrictServerInterface].
func (s server) GetChatMembers(ctx context.Context, request api.GetChatMembersRequestObject) (api.GetChatMembersResponseObject, error) {
	panic("unimplemented")
}

// GetChats implements [messagepublic.StrictServerInterface].
func (s server) GetChats(ctx context.Context, request api.GetChatsRequestObject) (api.GetChatsResponseObject, error) {
	panic("unimplemented")
}

// GetMessages implements [messagepublic.StrictServerInterface].
func (s server) GetMessages(ctx context.Context, request api.GetMessagesRequestObject) (api.GetMessagesResponseObject, error) {
	panic("unimplemented")
}

// GetUserInfo implements [messagepublic.StrictServerInterface].
func (s server) GetUserInfo(ctx context.Context, request api.GetUserInfoRequestObject) (api.GetUserInfoResponseObject, error) {
	panic("unimplemented")
}

// InvalidateSession implements [messagepublic.StrictServerInterface].
func (s server) InvalidateSession(ctx context.Context, request api.InvalidateSessionRequestObject) (api.InvalidateSessionResponseObject, error) {
	panic("unimplemented")
}

// LoginSession implements [messagepublic.StrictServerInterface].
func (s server) LoginSession(ctx context.Context, request api.LoginSessionRequestObject) (api.LoginSessionResponseObject, error) {
	panic("unimplemented")
}

// RefreshToken implements [messagepublic.StrictServerInterface].
func (s server) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	panic("unimplemented")
}

// RegisterUser implements [messagepublic.StrictServerInterface].
func (s server) RegisterUser(ctx context.Context, request api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	panic("unimplemented")
}

// SearchUsers implements [messagepublic.StrictServerInterface].
func (s server) SearchUsers(ctx context.Context, request api.SearchUsersRequestObject) (api.SearchUsersResponseObject, error) {
	panic("unimplemented")
}

// SendMessage implements [messagepublic.StrictServerInterface].
func (s server) SendMessage(ctx context.Context, request api.SendMessageRequestObject) (api.SendMessageResponseObject, error) {
	panic("unimplemented")
}

func badRequest(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, http.StatusBadRequest, api.InvalidRequest, err.Error())
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
