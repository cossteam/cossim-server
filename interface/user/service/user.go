package service

import (
	"bytes"
	"context"
	"fmt"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/user/api/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/msg_queue"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/avatarbuilder"
	"github.com/cossim/coss-server/pkg/utils/time"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
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

	token, err := utils.GenerateToken(resp.UserId, resp.Email)
	if err != nil {
		s.logger.Error("failed to generate user token", zap.Error(err))
		return nil, "", code.UserErrLoginFailed
	}

	values, err := cache.GetAllListValues(s.redisClient, resp.UserId)
	if err != nil {
		s.logger.Error("login :redis get key err =>", zap.Error(err))

		return nil, "", code.UserErrLoginFailed
	}

	//多端设备登录限制
	if s.conf.MultipleDeviceLimit.Enable {
		list, err := cache.GetUserInfoList(values)
		if err != nil {
			return nil, "", err
		}

		typeMap := cache.CategorizeByDriveType(list)
		if _, ok := typeMap[driveType]; ok {
			if len(typeMap[driveType]) >= s.conf.MultipleDeviceLimit.Max {
				fmt.Println("登录设备达到限制")
				return nil, "", code.UserErrLoginFailed
			}
		}
	}

	data := cache.UserInfo{
		ID:         uint(len(values)),
		UserId:     resp.UserId,
		Token:      token,
		DriverType: driveType,
		CreateAt:   time.Now(),
		ClientIP:   clientIp,
	}

	list := []interface{}{data}
	err = cache.AddToList(s.redisClient, resp.UserId, list)
	if err != nil {
		s.logger.Error("Redis error:", zap.Error(err))
		return nil, "", code.UserErrLoginFailed
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

func (s *Service) Logout(ctx context.Context, userID string, token string, request *model.LogoutRequest) error {
	values, err := cache.GetAllListValues(s.redisClient, userID)
	if err != nil {
		s.logger.Error("logout :redis get key err", zap.Error(err))
		return code.UserErrErrLogoutFailed
	}
	if len(values)-1 < int(request.LoginNumber) {
		s.logger.Error("logout : len(values) < int(request.LoginNumber)")
		return code.UserErrErrLogoutFailed
	}

	list, err := cache.GetUserInfoList(values)
	if err != nil {
		s.logger.Error("failed to get user info list", zap.Error(err))
		return code.UserErrErrLogoutFailed
	}

	if list[request.LoginNumber].Token != token {
		s.logger.Error("logout : list[request.LoginNumber].Token != token")
		return code.UserErrErrLogoutFailed
	}

	//通知消息服务关闭ws
	rid := list[request.LoginNumber].Rid
	t := list[request.LoginNumber].DriverType
	if rid != 0 {
		msg := msgconfig.WsMsg{
			Uid:    userID,
			Event:  msgconfig.OfflineEvent,
			Data:   map[string]interface{}{"rid": rid, "driver_type": t},
			SendAt: time.Now(),
		}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.UserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.UserWebsocketClose, msg)
		if err != nil {
			s.logger.Error("通知消息服务失败", zap.Error(err))
		}
	}

	//删除客户端信息
	err = cache.RemoveFromList(s.redisClient, userID, 0, values[request.LoginNumber])
	if err != nil {
		s.logger.Error("failed to logout user", zap.Error(err))
		return code.UserErrErrLogoutFailed
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
		return "", err
	}

	if s.conf.Email.Enable {
		//生成uuid
		ekey := uuid.New().String()

		//保存到redis
		err = cache.SetKey(s.redisClient, ekey, resp.UserId, 30*ostime.Minute)
		if err != nil {
			return "", err
		}

		//注册成功发送邮件
		err = s.smtpClient.SendEmail(req.Email, "欢迎注册", s.smtpClient.GenerateEmailVerificationContent("192.168.100.142:8080", resp.UserId, ekey))
		if err != nil {
			s.logger.Error("failed to send email", zap.Error(err))
			return "", err
		}
	}

	return resp.UserId, nil
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
