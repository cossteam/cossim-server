package service

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	grpcHandler "github.com/cossim/coss-server/internal/msg/interface/grpc"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/encryption"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	pushv1 "github.com/cossim/hipush/api/grpc/v1"
	"github.com/goccy/go-json"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

var (
	rabbitMQClient *msg_queue.RabbitMQ

	wsRid   int64 = 0 //全局客户端id
	wsMutex       = sync.Mutex{}
	Enc     encryption.Encryptor
	pool    = make(map[string]map[string][]*client)
)

// Service struct
type Service struct {
	ac               *pkgconfig.AppConfig
	dtmGrpcServer    string
	dialogGrpcServer string
	redisClient      *cache.RedisClient
	pushClient       pushv1.PushServiceClient
	logger           *zap.Logger
	sid              string
	cache            bool

	relationUserService   relationgrpcv1.UserRelationServiceClient
	relationGroupService  relationgrpcv1.GroupRelationServiceClient
	relationDialogService relationgrpcv1.DialogServiceClient
	userService           usergrpcv1.UserServiceClient
	userLoginService      usergrpcv1.UserLoginServiceClient
	groupService          groupgrpcv1.GroupServiceClient
	msgService            msggrpcv1.MsgServiceServer
	//msgClient            *grpcHandler.Handler

}

func New(ac *pkgconfig.AppConfig, handler *grpcHandler.Handler) *Service {
	mqClient, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", ac.MessageQueue.Username, ac.MessageQueue.Password, ac.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	rabbitMQClient = mqClient

	conn, err := grpc.Dial(ac.Push.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	pushClient := pushv1.NewPushServiceClient(conn)

	setupEncryption(ac)

	s := &Service{
		ac:            ac,
		logger:        plog.NewDefaultLogger("msg_bff", int8(ac.Log.Level)),
		sid:           xid.New().String(),
		redisClient:   setupRedis(ac),
		pushClient:    pushClient,
		dtmGrpcServer: ac.Dtm.Addr(),
		//rabbitMQClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
		//mqClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
	}
	s.msgService = handler
	// 监听服务消息队列
	go s.ListenQueue()

	s.cache = s.setCacheConfig()
	return s
}

func (s *Service) ListenQueue() {
	if rabbitMQClient.GetConnection().IsClosed() {
		panic("mqClient Connection is closed")
	}
	msgs, err := rabbitMQClient.ConsumeServiceMessages(msg_queue.MsgService, msg_queue.Service_Exchange)
	if err != nil {
		panic(err)
	}
	go func() {
		//监听队列
		for msg := range msgs {
			var msg_query msg_queue.ServiceQueueMsg
			err := json.Unmarshal(msg.Body, &msg_query)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				return
			}

			mmap, ok := msg_query.Data.(map[string]interface{})
			if !ok {
				fmt.Println("解析消息失败：")
				return
			}
			//map解析成结构体
			jsonData, err := json.Marshal(mmap)
			if err != nil {
				fmt.Println("Failed to marshal map to JSON:", err)
				return
			}
			var wsm constants.WsMsg
			err = json.Unmarshal(jsonData, &wsm)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				return
			}

			switch msg_query.Action {
			//推送消息
			case msg_queue.SendMessage:
				s.SendMsg(wsm.Uid, wsm.DriverId, wsm.Event, wsm.Data, true)
			case msg_queue.LiveEvent:
				s.SendMsg(wsm.Uid, wsm.DriverId, wsm.Event, wsm.Data, false)
			case msg_queue.Notice:
				fmt.Println("发送系统通知", wsm.Data)
				datamap := wsm.Data.(map[string]interface{})
				//map解析成结构体
				jsonData, err := json.Marshal(datamap)
				if err != nil {
					fmt.Println("Failed to marshal map to JSON:", err)
					return
				}
				var data constants.SystemNotificationEventData
				err = json.Unmarshal(jsonData, &data)
				if err != nil {
					fmt.Println("解析消息失败：", err)
					return
				}

				UserId := constants.SystemNotification

				msgList := make([]*msggrpcv1.SendUserMsgRequest, 0)

				wg := sync.WaitGroup{}
				for _, v := range data.UserIds {
					//查询对话id
					relation, err := s.relationUserService.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{
						UserId:   v,
						FriendId: UserId,
					})
					if err != nil {
						return
					}
					msg2 := &msggrpcv1.SendUserMsgRequest{
						SenderId:   UserId,
						ReceiverId: v,
						Content:    data.Content,
						DialogId:   relation.DialogId,
						Type:       int32(data.Type), //TODO 消息类型枚举
					}
					msgList = append(msgList, msg2)
					wg.Add(1)
					v2 := v
					go func() {
						defer wg.Done()
						s.SendMsg(v2, wsm.DriverId, constants.SendUserMessageEvent, msg2, true)
					}()
				}

				_, err = s.msgService.SendMultiUserMessage(context.Background(), &msggrpcv1.SendMultiUserMsgRequest{MsgList: msgList})
				if err != nil {
					s.logger.Error("发送系统通知失败", zap.Error(err))
					return
				}

				fmt.Println("等待线程结束")
				wg.Wait()
				//sendIds := data.UserIds
				//data.UserIds = nil
				////delete(data, "user_ids")
				//s.SendMsgToUsers(sendIds, "", constants.SendUserMessageEvent, data, true)

			//强制断开ws
			case msg_queue.UserWebsocketClose:
				thismap, ok := wsm.Data.(map[string]interface{})
				if !ok {
					fmt.Println("解析消息失败：")
					return
				}
				t := thismap["driver_type"]
				id := thismap["rid"]
				//类型断言
				driType, ok := t.(string)
				if !ok {
					fmt.Println("解析消息失败：")
					return
				}
				rid := id.(float64)
				for _, c := range pool[wsm.Uid][driType] {

					if c.Rid == int64(rid) {
						fmt.Println("关闭连接", rid)
						c.Conn.Close()
					}
				}
			}
		}
	}()
}

func (s *Service) Stop(ctx context.Context) error {
	//关闭队列
	rabbitMQClient.Close()
	return nil
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	switch serviceName {
	case "user_service":
		s.userService = usergrpcv1.NewUserServiceClient(conn)
		s.userLoginService = usergrpcv1.NewUserLoginServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation_service":
		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupService = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogService = relationgrpcv1.NewDialogServiceClient(conn)
		s.dialogGrpcServer = conn.Target()
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupService = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
}

func setupEncryption(ac *pkgconfig.AppConfig) {
	enc := encryption.NewEncryptor([]byte(ac.Encryption.Passphrase), ac.Encryption.Name, ac.Encryption.Email, ac.Encryption.RsaBits, ac.Encryption.Enable)

	err := enc.ReadKeyPair()
	if err != nil {
		return
	}

	Enc = enc
}

func (s *Service) setCacheConfig() bool {
	if s.redisClient == nil && s.ac.Cache.Enable {
		panic("redis is nil")
	}
	return s.ac.Cache.Enable
}
