package interfaces

import (
	"context"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	api "github.com/cossim/coss-server/services/msg/api/v1"
	"github.com/cossim/coss-server/services/msg/domain/service"
	"github.com/cossim/coss-server/services/msg/infrastructure/persistence"
)

type GrpcHandler struct {
	svc *service.MsgService
	api.UnimplementedMsgServiceServer
}

func NewGrpcHandler(c *config.AppConfig) *GrpcHandler {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}

	return &GrpcHandler{
		svc: service.NewMsgService(persistence.NewMsgRepo(dbConn)),
	}
}

// 发送用户消息
func (g *GrpcHandler) SendUserMessage(ctx context.Context, request *api.SendUserMsgRequest) (*api.SendUserMsgResponse, error) {
	_, err := g.svc.SendUserMessage(request.SenderId, request.ReceiverId, request.Content, uint(request.Type), uint(request.ReplayId))
	if err != nil {
		return nil, err
	}
	return &api.SendUserMsgResponse{}, nil
}

// 发送群聊消息
func (g *GrpcHandler) SendGroupMessage(ctx context.Context, request *api.SendGroupMsgRequest) (*api.SendGroupMsgResponse, error) {

	return &api.SendGroupMsgResponse{}, nil
}
