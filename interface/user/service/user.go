package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/interface/user/api/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/avatarbuilder"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strconv"
	"strings"
	ostime "time"
)

func (s *Service) Login(ctx context.Context, req *model.LoginRequest, driveType string, clientIp string) (*model.UserInfoResponse, string, error) {
	resp, err := s.userClient.UserLogin(ctx, &usergrpcv1.UserLoginRequest{
		Email:    req.Email,
		Password: utils.HashString(req.Password),
	})
	if err != nil {
		s.logger.Error("user login failed", zap.Error(err))
		return nil, "", err
	}

	token, err := utils.GenerateToken(resp.UserId, resp.Email, req.DriverId)
	if err != nil {
		s.logger.Error("failed to generate user token", zap.Error(err))
		return nil, "", code.UserErrLoginFailed
	}
	keys, err := cache.ScanKeys(s.redisClient, resp.UserId+":"+driveType+":*")
	if err != nil {
		s.logger.Error("redis scan err", zap.Error(err))
		return nil, "", code.UserErrErrLogoutFailed
	}

	//多端设备登录限制
	if s.ac.MultipleDeviceLimit.Enable {
		if len(keys) >= s.ac.MultipleDeviceLimit.Max {
			fmt.Println("登录设备达到限制")
			return nil, "", code.UserErrLoginFailed
		}

	}

	id := len(keys) + 1
	data := cache.UserInfo{
		ID:         uint(id),
		UserId:     resp.UserId,
		Token:      token,
		DriverType: driveType,
		CreateAt:   time.Now(),
		ClientIP:   clientIp,
	}

	err = cache.SetKey(s.redisClient, resp.UserId+":"+driveType+":"+strconv.Itoa(id), data, 60*60*24*7*ostime.Second)
	if err != nil {
		return nil, "", err
	}

	//推送登录提醒
	//查询是否在该设备第一次登录
	userLogin, err := s.userLoginClient.GetUserLoginByDriverIdAndUserId(ctx, &usergrpcv1.DriverIdAndUserId{
		UserId:   resp.UserId,
		DriverId: req.DriverId,
	})
	if err != nil {
		s.logger.Error("failed to get user login by driver id and user id", zap.Error(err))
	}

	_, err = s.userLoginClient.InsertUserLogin(ctx, &usergrpcv1.UserLogin{
		UserId:   resp.UserId,
		DriverId: req.DriverId,
		Token:    token,
	})
	if err != nil {
		return nil, "", err
	}

	fristLogin := false
	if userLogin != nil {
		if userLogin.DriverId == "" {
			fristLogin = true
		}
	}
	if fristLogin {
		if clientIp == "127.0.0.1" {
			clientIp = httputil.GetMyPublicIP()
		}
		info := httputil.OnlineIpInfo(clientIp)
		result := fmt.Sprintf("您在新设备登录，IP地址为：%s\n位置为：%s %s %s", clientIp, info.Country, info.RegionName, info.City)
		msg := constants.WsMsg{Uid: "10001", Event: constants.SystemNotificationEvent, SendAt: time.Now(), Data: map[string]interface{}{"user_ids": []string{resp.UserId}, "content": result, "type": 1}}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.UserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.Notice, msg)
	}

	return &model.UserInfoResponse{
		LoginNumber: data.ID,
		Email:       resp.Email,
		UserId:      resp.UserId,
		Nickname:    resp.NickName,
		Avatar:      resp.Avatar,
		Signature:   resp.Signature,
	}, token, nil
}

func (s *Service) Logout(ctx context.Context, userID string, token string, request *model.LogoutRequest, driverType string) error {
	value, err := cache.GetKey(s.redisClient, userID+":"+driverType+":"+strconv.Itoa(int(request.LoginNumber)))
	if err != nil {
		return err
	}

	data := value.(string)

	info, err := cache.GetUserInfo(data)
	if err != nil {
		return err
	}
	//通知消息服务关闭ws
	rid := info.Rid
	t := info.DriverType
	if rid != 0 {
		msg := constants.WsMsg{
			Uid:   userID,
			Event: constants.OfflineEvent,
			Data: constants.OfflineEventData{
				Rid:        rid,
				DriverType: constants.DriverType(t),
			},
			SendAt: time.Now(),
		}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.UserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.UserWebsocketClose, msg)
		if err != nil {
			s.logger.Error("通知消息服务失败", zap.Error(err))
		}
	}

	//删除客户端信息
	err = cache.DelKey(s.redisClient, userID+":"+driverType+":"+strconv.Itoa(int(request.LoginNumber)))
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Register(ctx context.Context, req *model.RegisterRequest) (string, error) {
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		req.Nickname = req.Email
	}

	avatar, err := avatarbuilder.GenerateAvatar(req.Nickname, s.appPath)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Image))
	if err != nil {
		return "", err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(avatar)
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	headerUrl, err := s.sp.Upload(context.Background(), key, reader, reader.Size(), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}

	headerUrl.Host = s.gatewayAddress
	headerUrl.Path = s.downloadURL + headerUrl.Path

	var UserId string
	workflow.InitGrpc(s.dtmGrpcServer, s.userGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "register_user_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {

		resp, err := s.userClient.UserRegister(ctx, &usergrpcv1.UserRegisterRequest{
			Email:           req.Email,
			NickName:        req.Nickname,
			Password:        utils.HashString(req.Password),
			ConfirmPassword: req.ConfirmPass,
			PublicKey:       req.PublicKey,
			Avatar:          headerUrl.String(),
		})
		if err != nil {
			s.logger.Error("failed to register user", zap.Error(err))
			return err
		}
		UserId = resp.UserId
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userClient.CreateUserRollback(context.Background(), &usergrpcv1.CreateUserRollbackRequest{UserId: resp.UserId})
			if err != nil {
				return err
			}
			return nil
		})

		//TODO 系统账号统一管理
		_, err = s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: "10001"})
		if err != nil {
			return err
		}

		//添加系统好友
		_, err = s.relClient.AddFriend(ctx, &relationgrpcv1.AddFriendRequest{
			UserId:   resp.UserId,
			FriendId: "10001",
		})
		if err != nil {
			return err
		}

		return err
	}); err != nil {
		return "", err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return "", err
	}

	if s.ac.Email.Enable {
		//生成uuid
		ekey := uuid.New().String()

		//保存到redis
		err = cache.SetKey(s.redisClient, ekey, UserId, 30*ostime.Minute)
		if err != nil {
			return "", err
		}

		//注册成功发送邮件
		err = s.smtpClient.SendEmail(req.Email, "欢迎注册", s.smtpClient.GenerateEmailVerificationContent(s.gatewayAddress+s.gatewayPort, UserId, ekey))
		if err != nil {
			s.logger.Error("failed to send email", zap.Error(err))
			return "", err
		}
	}

	return UserId, nil
}

func (s *Service) Search(ctx context.Context, userID string, email string) (interface{}, error) {
	r, err := s.userClient.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	resp := &model.UserInfoResponse{
		UserId:    r.UserId,
		Nickname:  r.NickName,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Signature: r.Signature,
		Status:    model.UserStatus(r.Status),
	}

	relation, err := s.relClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: r.UserId,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return resp, nil
	}

	if relation.Status == relationgrpcv1.RelationStatus_RELATION_NORMAL {
		resp.RelationStatus = model.UserRelationStatusFriend
	} else if relation.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
		resp.RelationStatus = model.UserRelationStatusBlacked
	} else {
		resp.RelationStatus = model.UserRelationStatusUnknown
	}
	return resp, nil
}

func (s *Service) GetUserInfo(ctx context.Context, thisID string, userID string) (*model.UserInfoResponse, error) {
	resp := &model.UserInfoResponse{}
	r, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, err
	}

	resp = &model.UserInfoResponse{
		UserId:    r.UserId,
		Nickname:  r.NickName,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Signature: r.Signature,
		Status:    model.UserStatus(r.Status),
	}

	if thisID != userID {
		relation, err := s.relClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
			UserId:   thisID,
			FriendId: userID,
		})
		if err != nil {
			s.logger.Error("获取用户关系失败", zap.Error(err))
			return resp, nil
		}

		if relation.Status == relationgrpcv1.RelationStatus_RELATION_NORMAL {
			resp.RelationStatus = model.UserRelationStatusFriend
		} else if relation.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
			resp.RelationStatus = model.UserRelationStatusBlacked
		} else {
			resp.RelationStatus = model.UserRelationStatusUnknown
		}

		resp.Preferences = &model.Preferences{
			OpenBurnAfterReading: model.OpenBurnAfterReadingType(relation.OpenBurnAfterReading),
			SilentNotification:   model.SilentNotification(relation.IsSilent),
			Remark:               relation.Remark,
		}
	}

	return resp, nil
}

func (s *Service) ModifyUserInfo(ctx context.Context, userID string, req *model.UserInfoRequest) error {
	// 调用服务端设置用户公钥的方法
	_, err := s.userClient.ModifyUserInfo(ctx, &usergrpcv1.User{
		UserId:    userID,
		NickName:  req.NickName,
		Tel:       req.Tel,
		Avatar:    req.Avatar,
		Signature: req.Signature,
		//Action:    user.UserStatus(req.Action),
	})
	if err != nil {
		s.logger.Error("修改用户信息失败", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) ModifyUserPassword(ctx context.Context, userID string, req *model.PasswordRequest) error {
	//查询用户旧密码
	info, err := s.userClient.GetUserPasswordByUserId(ctx, &usergrpcv1.UserRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return err
	}
	if info.Password != utils.HashString(req.OldPasswprd) {
		return code.UserErrOldPassword
	}

	if req.Password != req.ConfirmPass {
		return code.UserErrPasswordNotMatch
	}

	if req.Password == req.OldPasswprd || req.ConfirmPass == req.OldPasswprd {
		return code.UserErrNewPasswordAndOldPasswordEqual
	}
	_, err = s.userClient.ModifyUserPassword(ctx, &usergrpcv1.ModifyUserPasswordRequest{
		UserId:   userID,
		Password: utils.HashString(req.Password),
	})
	if err != nil {
		s.logger.Error("修改用户密码失败", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) SetUserPublicKey(ctx context.Context, userID string, publicKey string) (interface{}, error) {
	// 调用服务端设置用户公钥的方法
	_, err := s.userClient.SetUserPublicKey(ctx, &usergrpcv1.SetPublicKeyRequest{
		UserId:    userID,
		PublicKey: publicKey,
	})
	if err != nil {
		s.logger.Error("设置用户公钥失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) ModifyUserSecretBundle(ctx context.Context, userID string, req *model.ModifyUserSecretBundleRequest) (interface{}, error) {
	_, err := s.userClient.SetUserSecretBundle(context.Background(), &usergrpcv1.SetUserSecretBundleRequest{
		UserId:       userID,
		SecretBundle: req.SecretBundle,
	})
	if err != nil {
		s.logger.Error("修改用户秘钥包失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) GetUserSecretBundle(ctx context.Context, userID string) (*model.UserSecretBundleResponse, error) {

	_, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, code.UserErrGetUserInfoFailed
	}

	bundle, err := s.userClient.GetUserSecretBundle(context.Background(), &usergrpcv1.GetUserSecretBundleRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户秘钥包失败", zap.Error(err))
		return nil, err
	}

	return &model.UserSecretBundleResponse{
		UserId:       bundle.UserId,
		SecretBundle: bundle.SecretBundle,
	}, nil
}

func (s *Service) GetUserLoginClients(ctx context.Context, userID string) ([]*model.GetUserLoginClientsResponse, error) {

	values, err := cache.GetAllListValues(s.redisClient, userID)
	if err != nil {
		s.logger.Error("获取用户登录客户端失败：", zap.Error(err))
		return nil, code.UserErrGetUserLoginClientsFailed
	}
	users, err := cache.GetUserInfoList(values)
	if err != nil {
		s.logger.Error("获取用户登录客户端失败：", zap.Error(err))
		return nil, code.UserErrGetUserLoginClientsFailed
	}
	var clients []*model.GetUserLoginClientsResponse
	for _, user := range users {
		clients = append(clients, &model.GetUserLoginClientsResponse{
			ClientIP:    user.ClientIP,
			DriverType:  user.DriverType,
			LoginNumber: user.ID,
			LoginAt:     user.CreateAt,
		})
	}
	return clients, nil
}

func (s *Service) UserActivate(ctx context.Context, userID string, key string) (interface{}, error) {
	value, err := cache.GetKey(s.redisClient, key)
	if err != nil {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	if value != userID {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	resp, err := s.userClient.ActivateUser(ctx, &usergrpcv1.UserRequest{
		UserId: userID,
	})

	if err != nil {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	//删除缓存
	err = cache.DelKey(s.redisClient, key)
	if err != nil {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	return resp, nil
}

func (s *Service) ResetUserPublicKey(ctx context.Context, req *model.ResetPublicKeyRequest) (interface{}, error) {
	//查询redis是否存在该验证码
	value, err := cache.GetKey(s.redisClient, req.Code)
	if err != nil {
		s.logger.Error("重置用户公钥失败", zap.Error(err))
		return nil, code.UserErrResetPublicKeyFailed
	}

	info, err := s.userClient.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		s.logger.Error("重置用户公钥失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	if value != info.UserId {
		return nil, code.UserErrResetPublicKeyFailed
	}

	_, err = s.userClient.SetUserPublicKey(ctx, &usergrpcv1.SetPublicKeyRequest{
		UserId:    info.UserId,
		PublicKey: req.PublicKey,
	})
	if err != nil {
		return nil, err
	}

	err = cache.DelKey(s.redisClient, req.Code)
	if err != nil {
		s.logger.Error("重置用户公钥失败", zap.Error(err))
		return nil, code.UserErrResetPublicKeyFailed
	}

	return value, nil
}

func (s *Service) SendEmailCode(ctx context.Context, email string) (interface{}, error) {
	//查询用户是否存在
	info, err := s.userClient.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
		Email: email,
	})
	if err != nil {
		s.logger.Error("发送邮箱验证码失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	//生成验证码
	code1 := utils.RandomNum()

	//设置验证码(30分钟超时)
	err = cache.SetKey(s.redisClient, code1, info.UserId, 30*ostime.Minute)
	if err != nil {
		s.logger.Error("发送邮箱验证码失败", zap.Error(err))
		return nil, code.UserErrSendEmailCodeFailed
	}

	if s.ac.Email.Enable {
		err := s.smtpClient.SendEmail(email, "重置pgp验证码(请妥善保管,有效时间30分钟)", code1)
		if err != nil {
			return nil, err
		}
	}
	return nil, err
}
