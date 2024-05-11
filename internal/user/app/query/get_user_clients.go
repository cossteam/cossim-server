package query

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetUserLoginClients struct {
	UserID string
}

type GetUserLoginClientsResponse struct {
	ClientIP   string `json:"client_ip"`
	DriverType string `json:"driver_type"`
	DriverID   string `json:"driver_id"`
	LoginAt    int64  `json:"login_at"`
}

type GetUserClientsHandler decorator.CommandHandler[*GetUserLoginClients, []*GetUserLoginClientsResponse]

func NewGetUserClientsHandler(logger *zap.Logger, userCache cache.UserCache, userDomain service.UserDomain) GetUserClientsHandler {
	return &getUserClientsHandler{
		logger:    logger,
		userCache: userCache,
		ud:        userDomain,
	}
}

type getUserClientsHandler struct {
	logger    *zap.Logger
	userCache cache.UserCache
	ud        service.UserDomain
}

func (h *getUserClientsHandler) Handle(ctx context.Context, cmd *GetUserLoginClients) ([]*GetUserLoginClientsResponse, error) {
	if cmd == nil {
		return nil, code.InvalidParameter
	}

	users, err := h.userCache.GetUserLoginInfos(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	var clients []*GetUserLoginClientsResponse
	for _, user := range users {
		clients = append(clients, &GetUserLoginClientsResponse{
			ClientIP:   user.ClientIP,
			DriverType: user.DriverType,
			DriverID:   user.DriverID,
			LoginAt:    user.CreatedAt,
		})
	}
	return clients, nil
}
