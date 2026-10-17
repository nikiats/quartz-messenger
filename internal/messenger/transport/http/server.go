package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	api "quaztz/internal/api/messagepublic"
	"quaztz/internal/messenger/domain"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type server struct {
	c Chats
	u Users
	m Messages
	s Sessions
}

func NewHandler(c Chats, u Users, m Messages, a Authenticator, s Sessions) http.Handler {
	strict := api.NewStrictHandlerWithOptions(
		server{c, u, m, s},
		[]api.StrictMiddlewareFunc{
			authMiddleware(a),
		},
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

func (s server) AddChatMember(ctx context.Context, req api.AddChatMemberRequestObject) (api.AddChatMemberResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	memberID := req.Body.UserId
	chatID := int64(req.Id)
	err := s.c.AddMember(ctx, userID, chatID, memberID)

	switch {
	case errors.Is(err, domain.ErrRequiresAdmin):
		return api.AddChatMember403JSONResponse{
			ForbiddenJSONResponse: api.ForbiddenJSONResponse{
				Code:    api.NeedPrivileges,
				Message: "need admin privileges",
			},
		}, nil
	case errors.Is(err, domain.ErrAddToPrivateChat):
		return api.AddChatMember403JSONResponse{
			ForbiddenJSONResponse: api.ForbiddenJSONResponse{
				Code:    api.EditPrivateChatMember,
				Message: "can't add to private chat",
			},
		}, nil
	case errors.Is(err, domain.ErrChatNotFound):
		return api.AddChatMember404JSONResponse{
			Code:    api.ChatNotFound,
			Message: "chat not exists",
		}, nil
	case errors.Is(err, domain.ErrUserNotFound):
		return api.AddChatMember404JSONResponse{
			Code:    api.UserNotFound,
			Message: "user not found",
		}, nil
	case errors.Is(err, domain.ErrMemberExists):
		return api.AddChatMember409JSONResponse{
			Code:    api.AlreadyMember,
			Message: "member exists",
		}, nil
	case err != nil:
		return nil, err
	}

	return api.AddChatMember201Response{}, nil
}

func (s server) CreateChat(ctx context.Context, req api.CreateChatRequestObject) (api.CreateChatResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	create := req.Body

	if err := validateChatCreate(userID, create); err != nil {
		return api.CreateChat400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.InvalidRequest,
				Message: err.Error(),
			},
		}, nil
	}

	var title string
	if create.Title != nil {
		title = *create.Title
	}
	var clientRequestID uuid.UUID
	if create.ClientRequestId != nil {
		clientRequestID = *create.ClientRequestId
	}
	chat, created, err := s.c.Create(ctx, userID, domain.ChatType(create.Type), title, create.MemberIds, clientRequestID)

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return api.CreateChat404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.UserNotFound,
				Message: "user not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	response := chatResponse(chat)
	if created {
		return api.CreateChat201JSONResponse(response), nil
	}
	return api.CreateChat200JSONResponse(response), nil
}

func (s server) DeleteChat(ctx context.Context, req api.DeleteChatRequestObject) (api.DeleteChatResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	chatID := int64(req.Id)
	err := s.c.Delete(ctx, userID, chatID)

	switch {
	case errors.Is(err, domain.ErrRequiresAdmin):
		return api.DeleteChat403JSONResponse{
			ForbiddenJSONResponse: api.ForbiddenJSONResponse{
				Code:    api.NeedPrivileges,
				Message: "need admin privileges",
			},
		}, nil
	case errors.Is(err, domain.ErrChatNotFound):
		return api.DeleteChat404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.DeleteChat204Response{}, nil
}

func (s server) DeleteChatMember(ctx context.Context, req api.DeleteChatMemberRequestObject) (api.DeleteChatMemberResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	memberID := req.UserId
	chatID := int64(req.Id)
	err := s.c.RemoveMember(ctx, userID, chatID, memberID)

	switch {
	case errors.Is(err, domain.ErrRequiresAdmin):
		return api.DeleteChatMember403JSONResponse{
			ForbiddenJSONResponse: api.ForbiddenJSONResponse{
				Code:    api.NeedPrivileges,
				Message: "need admin privileges",
			},
		}, nil
	case errors.Is(err, domain.ErrRemoveFromPrivateChat):
		return api.DeleteChatMember403JSONResponse{
			ForbiddenJSONResponse: api.ForbiddenJSONResponse{
				Code:    api.EditPrivateChatMember,
				Message: "can't remove from private chat",
			},
		}, nil
	case errors.Is(err, domain.ErrChatNotFound):
		return api.DeleteChatMember404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case errors.Is(err, domain.ErrMemberNotFound):
		return api.DeleteChatMember404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.MemberNotFound,
				Message: "member not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.DeleteChatMember204Response{}, nil
}

func (s server) EditProfile(ctx context.Context, req api.EditProfileRequestObject) (api.EditProfileResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	profile := req.Body

	if err := validateProfileEdit(profile); err != nil {
		return api.EditProfile400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.InvalidRequest,
				Message: err.Error(),
			},
		}, nil
	}

	err := s.u.EditProfile(ctx, userID, profile.VisibleName, profile.Username)

	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		return api.EditProfile409JSONResponse{
			ConflictJSONResponse: api.ConflictJSONResponse{
				Code:    api.UsernameTaken,
				Message: "username already taken",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.EditProfile200Response{}, nil
}

func (s server) GetChatInfo(ctx context.Context, req api.GetChatInfoRequestObject) (api.GetChatInfoResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	chatID := int64(req.Id)
	chat, err := s.c.Get(ctx, userID, chatID)

	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		return api.GetChatInfo404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.GetChatInfo200JSONResponse(chatResponse(chat)), nil
}

func (s server) GetChatMembers(ctx context.Context, req api.GetChatMembersRequestObject) (api.GetChatMembersResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	chatID := int64(req.Id)
	members, err := s.c.Members(ctx, userID, chatID)

	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		return api.GetChatMembers404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	response := make(api.GetChatMembers200JSONResponse, 0, len(members))
	for _, member := range members {
		response = append(response, api.Member{
			UserId:   member.UserID,
			Role:     api.MemberRole(member.Role),
			JoinedAt: member.JoinedAt,
		})
	}

	return response, nil
}

func (s server) GetChats(ctx context.Context, _ api.GetChatsRequestObject) (api.GetChatsResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	chats, err := s.c.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make(api.GetChats200JSONResponse, 0, len(chats))
	for _, chat := range chats {
		response = append(response, chatResponse(chat))
	}

	return response, nil
}

func (s server) GetMessages(ctx context.Context, req api.GetMessagesRequestObject) (api.GetMessagesResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	chatID := int64(req.Id)
	params := req.Params

	if err := validateMessagesParams(params); err != nil {
		return api.GetMessages400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.InvalidRequest,
				Message: err.Error(),
			},
		}, nil
	}

	query := domain.MessageQuery{
		Before: (*int64)(params.Before),
		After:  (*int64)(params.After),
		Limit:  defaultMessagesLimit,
	}
	if params.Limit != nil {
		query.Limit = *params.Limit
	}
	if params.Q != nil {
		query.Search = *params.Q
	}
	messages, hasMore, err := s.m.History(ctx, userID, chatID, query)

	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		return api.GetMessages404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	response := make([]api.Message, 0, len(messages))
	for _, message := range messages {
		response = append(response, api.Message{
			Id:       snowflake.ID(message.ID),
			ChatId:   snowflake.ID(message.ChatID),
			SenderId: message.SenderID,
			SentAt:   message.SentAt,
			Content: api.MessageContent{
				Text: message.Content.Text,
			},
		})
	}

	return api.GetMessages200JSONResponse{
		HasMore:  hasMore,
		Messages: response,
	}, nil
}

func (s server) GetUserInfo(ctx context.Context, req api.GetUserInfoRequestObject) (api.GetUserInfoResponseObject, error) {
	userID := req.Id
	user, err := s.u.Get(ctx, userID)

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return api.GetUserInfo404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.UserNotFound,
				Message: "user not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	response := api.GetUserInfo200JSONResponse{
		UserId:      user.ID,
		VisibleName: &user.VisibleName,
	}
	if user.Username != "" {
		response.Username = &user.Username
	}

	return response, nil
}

func (s server) InvalidateSession(ctx context.Context, req api.InvalidateSessionRequestObject) (api.InvalidateSessionResponseObject, error) {
	userID := sessionFrom(ctx).UserID
	sessionID := req.Id
	err := s.s.Revoke(ctx, userID, sessionID)

	switch {
	case errors.Is(err, domain.ErrSessionNotFound):
		return api.InvalidateSession404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.SessionNotFound,
				Message: "session not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.InvalidateSession204Response{}, nil
}

func (s server) LoginSession(ctx context.Context, req api.LoginSessionRequestObject) (api.LoginSessionResponseObject, error) {
	credentials := req.Body
	session, tokens, err := s.s.Login(ctx, credentials.Login, credentials.Password)

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return api.LoginSession401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.AuthorizationFailed,
				Message: "invalid login or password",
			},
		}, nil
	case errors.Is(err, domain.ErrSuspendedUser):
		return api.LoginSession401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.SuspendedUser,
				Message: "account is suspended",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.LoginSession201JSONResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		SessionId:    session.SessionID,
	}, nil
}

func (s server) RefreshToken(ctx context.Context, req api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	refreshToken := req.Body.RefreshToken
	tokens, err := s.s.Refresh(ctx, refreshToken)

	switch {
	case errors.Is(err, domain.ErrTokenInvalid):
		return api.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.AuthorizationFailed,
				Message: "refresh token is invalid",
			},
		}, nil
	case errors.Is(err, domain.ErrTokenExpired):
		return api.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.SessionExpired,
				Message: "session expired",
			},
		}, nil
	case errors.Is(err, domain.ErrSessionRevoked):
		return api.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.SessionRevoked,
				Message: "session revoked",
			},
		}, nil
	case errors.Is(err, domain.ErrSuspendedUser):
		return api.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{
				Code:    api.SuspendedUser,
				Message: "account is suspended",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.RefreshToken200JSONResponse{
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s server) RegisterUser(ctx context.Context, req api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	credentials := req.Body

	if err := validateUserCreate(credentials); err != nil {
		return api.RegisterUser400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.InvalidRequest,
				Message: err.Error(),
			},
		}, nil
	}

	err := s.u.Register(ctx, credentials.Login, credentials.Password)

	switch {
	case errors.Is(err, domain.ErrLoginTaken):
		return api.RegisterUser409JSONResponse{
			ConflictJSONResponse: api.ConflictJSONResponse{
				Code:    api.LoginTaken,
				Message: "login already taken",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.RegisterUser201Response{}, nil
}

func (s server) SearchUsers(ctx context.Context, req api.SearchUsersRequestObject) (api.SearchUsersResponseObject, error) {
	var query string
	if req.Params.Q != nil {
		query = *req.Params.Q
	}
	users, err := s.u.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	response := make(api.SearchUsers200JSONResponse, 0, len(users))
	for _, user := range users {
		found := api.User{
			UserId:      user.ID,
			VisibleName: &user.VisibleName,
		}
		if user.Username != "" {
			found.Username = &user.Username
		}
		response = append(response, found)
	}

	return response, nil
}

func (s server) SendMessage(ctx context.Context, req api.SendMessageRequestObject) (api.SendMessageResponseObject, error) {
	senderID := sessionFrom(ctx).UserID
	chatID := int64(req.Id)
	send := req.Body

	if err := validateSendMessage(send); err != nil {
		return api.SendMessage400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Code:    api.InvalidRequest,
				Message: err.Error(),
			},
		}, nil
	}

	content := domain.MessageContent{
		Text: send.Content.Text,
	}
	message, err := s.m.Send(ctx, senderID, chatID, send.ClientMessageId, content)

	switch {
	case errors.Is(err, domain.ErrChatNotFound):
		return api.SendMessage404JSONResponse{
			NotFoundJSONResponse: api.NotFoundJSONResponse{
				Code:    api.ChatNotFound,
				Message: "chat not found",
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return api.SendMessage201JSONResponse{
		Id:       snowflake.ID(message.ID),
		ChatId:   snowflake.ID(message.ChatID),
		SenderId: message.SenderID,
		SentAt:   message.SentAt,
		Content: api.MessageContent{
			Text: message.Content.Text,
		},
	}, nil
}

func chatResponse(chat domain.Chat) api.Chat {
	response := api.Chat{
		Id:    snowflake.ID(chat.ID),
		Type:  api.ChatType(chat.Type),
		Title: &chat.Title,
	}
	if chat.Activity.ID != 0 {
		response.LastMessage = &api.LastMessage{
			Id:       snowflake.ID(chat.Activity.ID),
			SenderId: chat.Activity.SenderID,
			Preview:  chat.Activity.Preview,
			SentAt:   chat.Activity.SentAt,
		}
	}
	return response
}

func unauthorized(w http.ResponseWriter, _ *http.Request, code api.Error401Code) {
	writeError(w, http.StatusUnauthorized, code, "authorization failed")
}

func badRequest(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, http.StatusBadRequest, api.InvalidRequest, err.Error())
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("public API error", "path", r.URL.Path, "error", err)
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
