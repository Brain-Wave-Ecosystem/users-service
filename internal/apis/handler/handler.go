package handler

import (
	"context"
	apperrors "github.com/Brain-Wave-Ecosystem/go-common/pkg/error"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/helpers"
	users "github.com/Brain-Wave-Ecosystem/users-service/gen/users"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/apis/service"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/models"
	"github.com/bufbuild/protovalidate-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"strconv"
)

var _ users.UsersServiceServer = (*Handler)(nil)

type Handler struct {
	service   *service.Service
	logger    *zap.Logger
	validator protovalidate.Validator

	users.UnimplementedUsersServiceServer
}

func NewHandler(service *service.Service, logger *zap.Logger) *Handler {
	v, _ := protovalidate.New(protovalidate.WithFailFast())

	return &Handler{
		service:   service,
		logger:    logger,
		validator: v,
	}
}

func (h *Handler) GetUserByIdentifier(ctx context.Context, request *users.GetUserByIdentifierRequest) (*users.GetUserByIdentifierResponse, error) {
	user, err := h.service.GetUserByIdentifier(ctx, request.GetIdentifier())
	if err != nil {
		return nil, err
	}

	return &users.GetUserByIdentifierResponse{User: user.ToGRPC()}, nil
}

func (h *Handler) GetUserProfile(ctx context.Context, _ *emptypb.Empty) (*users.GetUserProfileResponse, error) {
	userID, err := h.extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := h.service.GetUserByIdentifier(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, err
	}

	return &users.GetUserProfileResponse{User: user.ToGRPC()}, nil
}

func (h *Handler) LoginUserByEmail(ctx context.Context, request *users.LoginUserByEmailRequest) (*users.LoginUserResponse, error) {
	user, err := h.service.GetUserByEmail(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, err
	}

	return &users.LoginUserResponse{User: user.ToGRPC()}, nil
}

func (h *Handler) CreateUser(ctx context.Context, request *users.CreateUserRequest) (*users.CreateUserResponse, error) {
	user, err := h.service.CreateUser(ctx, models.ToUserWithPassword(request))
	if err != nil {
		return nil, err
	}

	return &users.CreateUserResponse{User: user.ToGRPC()}, nil
}

func (h *Handler) ConfirmUser(ctx context.Context, request *users.UserConfirmRequest) (*emptypb.Empty, error) {
	err := h.service.ConfirmUser(ctx, request.UserId)
	return nil, err
}

func (h *Handler) UpdateUser(ctx context.Context, request *users.UpdateUserRequest) (*emptypb.Empty, error) {
	userID, err := h.extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if userID != request.GetId() {
		return nil, apperrors.Forbidden("You do not have permission to modify this user's data")
	}

	err = h.service.UpdateUser(ctx, userID, &models.UpdateUser{
		AvatarURL: request.AvatarUrl,
		FullName:  request.FullName,
		Bio:       request.Bio,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) UpdateUserAdmin(ctx context.Context, request *users.UpdateUserAdminRequest) (*emptypb.Empty, error) {
	err := h.service.UpdateUser(ctx, request.GetId(), &models.UpdateUser{
		AvatarURL: request.AvatarUrl,
		FullName:  request.FullName,
		Bio:       request.Bio,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) UpdateUserPassword(ctx context.Context, request *users.UpdateUserPasswordRequest) (*emptypb.Empty, error) {
	userID, err := h.extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	h.logger.Debug("users-service | UpdateUserPassword invoked", zap.Int64("user_id", userID), zap.Int64("request_id", request.GetId()))

	if userID != request.GetId() {
		return nil, apperrors.Forbidden("You do not have permission to modify this user's data")
	}

	err = h.service.UpdateUserPassword(ctx, userID, request.GetPassword())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) UpdateUserPasswordAdmin(ctx context.Context, request *users.UpdateUserPasswordAdminRequest) (*emptypb.Empty, error) {
	err := h.service.UpdateUserPassword(ctx, request.GetId(), request.GetPassword())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) DeleteUser(ctx context.Context, request *users.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := h.extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	if userID != request.GetId() {
		return nil, apperrors.Forbidden("You do not have permission to modify this user's data")
	}

	err = h.service.DeleteUser(ctx, userID)

	return nil, err
}

func (h *Handler) DeleteUserAdmin(ctx context.Context, request *users.DeleteUserRequest) (*emptypb.Empty, error) {
	err := h.service.DeleteUser(ctx, request.Id)
	return nil, err
}

func (h *Handler) Health(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok"))
}

func (h *Handler) extractUserID(ctx context.Context) (int64, error) {
	userIDStr, err := helpers.ExtractFromMD(ctx, "user-id")
	if err != nil {
		return 0, apperrors.BadRequestHidden(err, "missing userID from token")
	}

	id, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, apperrors.BadRequestHidden(err, "invalid userID")
	}

	return id, nil
}
