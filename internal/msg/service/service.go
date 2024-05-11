package service

// ServiceImpl struct
//type ServiceImpl struct {
//	ac                  *pkgconfig.AppConfig
//	dtmGrpcServer       string
//	relationServiceAddr string
//	userServiceAddr     string
//	groupServiceAddr    string
//	msgServiceAddr      string
//	redisClient         *cache.RedisClient
//	logger              *zap.Logger
//	sid                 string
//	cache               bool
//	//enc                 encryption.Encryptor
//
//	relationUserService   relationgrpcv1.UserRelationServiceClient
//	relationGroupService  relationgrpcv1.GroupRelationServiceClient
//	relationDialogService relationgrpcv1.DialogServiceClient
//	userService           usergrpcv1.UserServiceClient
//	userLoginService      usergrpcv1.UserLoginServiceClient
//	groupService          groupgrpcv1.GroupServiceClient
//	msgService            msggrpcv1.MsgServiceServer
//	msgGroupService       msggrpcv1.GroupMessageServiceServer
//	pushService           pushv1.PushServiceClient
//	//msgClient            *grpcHandler.Handler
//}
//
//func New(ac *pkgconfig.AppConfig, handler *grpcHandler.Handler) *ServiceImpl {
//	s := &ServiceImpl{
//		ac:            ac,
//		logger:        plog.NewDefaultLogger("msg_bff", int8(ac.Log.Level)),
//		sid:           xid.New().String(),
//		redisClient:   setupRedis(ac),
//		dtmGrpcServer: ac.Dtm.Addr(),
//		//rabbitMQClient: mqClient,
//		//pool:     make(map[string]map[string][]*client),
//		//mqClient: mqClient,
//		//pool:     make(map[string]map[string][]*client),
//	}
//	s.msgService = handler
//	s.msgGroupService = handler
//	s.cache = s.setCacheConfig()
//	//s.setupEncryption(ac)
//	return s
//}
//
//func (s *ServiceImpl) Stop(ctx context.Context) error {
//	return nil
//}
//
//func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
//	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
//}
//
////func (s *ServiceImpl) setupEncryption(ac *pkgconfig.AppConfig) {
////	enc2 := encryption.NewEncryptor([]byte(ac.Encryption.Passphrase), ac.Encryption.Name, ac.Encryption.Email, ac.Encryption.RsaBits, ac.Encryption.Enable)
////
////	err := enc2.ReadKeyPair()
////	if err != nil {
////		return
////	}
////
////	s.enc = enc2
////}
//
//func (s *ServiceImpl) setCacheConfig() bool {
//	if s.redisClient == nil && s.ac.Cache.Enable {
//		panic("redis is nil")
//	}
//	return s.ac.Cache.Enable
//}
