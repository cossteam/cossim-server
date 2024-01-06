package interfaces

import (
	"context"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	api "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/entity"
	"github.com/cossim/coss-server/service/msg/domain/service"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
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

// 获取私聊消息
func (g *GrpcHandler) GetUserMessageList(ctx context.Context, request *api.GetUserMsgListRequest) (*api.GetUserMsgListResponse, error) {
	res := g.svc.GetUserMessageList(request.UserId, request.FriendId, request.Content, entity.UserMessageType(request.Type), int(request.PageNum), int(request.PageSize))
	msgs := make([]*api.UserMessage, 0)
	if len(res.UserMessages) > 0 {
		for _, m := range res.UserMessages {
			msgs = append(msgs, &api.UserMessage{
				Id:         int64(m.ID),
				SenderId:   m.SendID,
				ReceiverId: m.ReceiveID,
				Content:    m.Content,
				Type:       int32(m.Type),
				ReplayId:   uint64(m.ReplyId),
				IsRead:     int32(m.IsRead),
				ReadAt:     m.ReadAt,
				CreatedAt:  m.CreatedAt.Unix(),
			})
		}
	}
	return &api.GetUserMsgListResponse{
		Total:        res.Total,
		CurrentPage:  res.CurrentPage,
		UserMessages: msgs,
	}, nil
}
