package service

import (
	"fmt"
	"github.com/cossim/coss-server/interface/msg/config"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	msggrpcv1 "github.com/cossim/coss-server/service/msg/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
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
	relationUserClient   relationgrpcv1.UserRelationServiceClient
	relationGroupClient  relationgrpcv1.GroupRelationServiceClient
	relationDialogClient relationgrpcv1.DialogServiceClient
	userClient           usergrpcv1.UserServiceClient
	groupClient          groupgrpcv1.GroupServiceClient
	msgClient            msggrpcv1.MsgServiceClient
	redisClient          *redis.Client

	//mqClient *msg_queue.RabbitMQ

	logger    *zap.Logger
	sid       string
	discovery discovery.Discovery
	conf      *pkgconfig.AppConfig
	//Enc       encryption.Encryptor

	//pool  map[string]map[string][]*client
}

func New() (s *Service) {
	mqClient, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", config.Conf.MessageQueue.Username, config.Conf.MessageQueue.Password, config.Conf.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	rabbitMQClient = mqClient
	return &Service{
		conf:        config.Conf,
		logger:      setupLogger(),
		sid:         xid.New().String(),
		redisClient: setupRedis(),
		//Enc:       setupEncryption(),
		//rabbitMQClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
		//mqClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
	}
}

func setupLogger() *zap.Logger {
	return plog.NewDevLogger("msg_bff")
}

func (s *Service) Start(discover bool) {
	// 监听服务消息队列
	go s.ListenQueue()

	if discover {
		d, err := discovery.NewConsulRegistry(s.conf.Register.Addr())
		if err != nil {
			panic(err)
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid); err != nil {
			panic(err)
		}
		s.logger.Info("Service registration successful", zap.String("service", s.conf.Register.Name), zap.String("addr", s.conf.HTTP.Addr()), zap.String("sid", s.sid))
		go s.discover()
	} else {
		s.direct()
	}
}

func Restart(discover bool) *Service {
	s := New()
	s.logger.Info("Service restart")
	s.Start(discover)
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
			var wsm config.WsMsg
			err = json.Unmarshal(jsonData, &wsm)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				return
			}

			switch msg_query.Action {
			//推送消息
			case msg_queue.SendMessage:
				s.SendMsg(wsm.Uid, wsm.Event, wsm.Data)
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

func (s *Service) Stop(discover bool) {
	//关闭队列
	rabbitMQClient.Close()
	if discover {
		if err := s.discovery.Cancel(s.sid); err != nil {
			s.logger.Error("Failed to cancel service registration: %v", zap.Error(err))
			return
		}
		s.logger.Info("Service registration canceled", zap.String("service", s.conf.Register.Name), zap.String("addr", s.conf.GRPC.Addr()), zap.String("sid", s.sid))
	}
}

func (s *Service) discover() {
	var wg sync.WaitGroup
	type serviceInfo struct {
		ServiceName string
		Addr        string
	}
	ch := make(chan serviceInfo)

	for serviceName, c := range s.conf.Discovers {
		wg.Add(1)
		go func(serviceName string, c pkgconfig.ServiceConfig) {
			defer wg.Done()
			for {
				addr, err := s.discovery.Discover(c.Name)
				if err != nil {
					s.logger.Info("Service discovery failed", zap.String("service", c.Name))
					time.Sleep(15 * time.Second)
					continue
				}
				s.logger.Info("Service discovery successful", zap.String("service", s.conf.Register.Name), zap.String("addr", addr))
				ch <- serviceInfo{ServiceName: serviceName, Addr: addr}
				break
			}
		}(serviceName, c)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for info := range ch {
		if err := s.handlerGrpcClient(info.ServiceName, info.Addr); err != nil {
			panic(err)
		}
	}
}

func (s *Service) direct() {
	for serviceName, _ := range s.conf.Discovers {
		if err := s.handlerGrpcClient(serviceName, s.conf.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	switch serviceName {
	case "user":
		s.userClient = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation":
		s.relationUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	case "msg":
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "msg"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.Addr(),
		Password: config.Conf.Redis.Password, // no password set
		DB:       0,                          // use default DB
		//Protocol: cfg,
	})
}
