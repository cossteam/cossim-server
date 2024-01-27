package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/interface/user/api/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"go.uber.org/zap"
	"strings"
)

func (s *Service) Login(ctx context.Context, req *model.LoginRequest, driveType string, clientIp string) (*model.UserInfoResponse, string, error) {
	resp, err := s.userClient.UserLogin(ctx, &usergrpcv1.UserLoginRequest{
		Email:    req.Email,
		Password: utils.HashString(req.Password),
	})
	if err != nil {
		s.logger.Error("user login failed", zap.Error(err))
		return nil, "", code.UserErrNotExistOrPassword
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

func (s *Service) Logout(ctx context.Context, userID string, request *model.LogoutRequest) error {
	values, err := cache.GetAllListValues(s.redisClient, userID)
	if err != nil {
		return code.UserErrErrLogoutFailed
	}
	if len(values)-1 < int(request.LoginNumber) {
		s.logger.Error("logout : len(values) < int(request.LoginNumber)")
		return code.UserErrErrLogoutFailed
	}
	//todo 关闭websocket
	//list, err := cache.GetUserInfoList(values)
	//if err != nil {
	//	s.logger.Error("failed to get user info list", zap.Error(err))
	//	return code.UserErrErrLogoutFailed
	//}
	//rid := list[request.LoginNumber].Rid
	//s.rabbitMQClient.PublishServiceMessage()

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

	resp, err := s.userClient.UserRegister(ctx, &usergrpcv1.UserRegisterRequest{
		Email:           req.Email,
		NickName:        req.Nickname,
		Password:        utils.HashString(req.Password),
		ConfirmPassword: req.ConfirmPass,
		PublicKey:       req.PublicKey,
		Avatar:          "https://fastly.picsum.photos/id/1036/200/200.jpg?hmac=Yb5E0WTltIYlUDPDqT-d0Llaaq0mJnwiCUtxx8RrtVE",
	})
	if err != nil {
		s.logger.Error("failed to register user", zap.Error(err))
		return "", err
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
