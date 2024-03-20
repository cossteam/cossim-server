package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/api/http/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/avatar"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"image/png"
	"io/ioutil"
	"mime/multipart"
	"regexp"
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
	keys, err := s.redisClient.ScanKeys(resp.UserId + ":" + driveType + ":*")
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

	err = s.redisClient.SetKey(resp.UserId+":"+driveType+":"+strconv.Itoa(id), data, 60*60*24*7*ostime.Second)
	if err != nil {
		return nil, "", err
	}

	// 推送登录提醒
	// 查询是否在该设备第一次登录
	userLogin, err := s.userClient.GetUserLoginByDriverIdAndUserId(ctx, &usergrpcv1.DriverIdAndUserId{
		UserId:   resp.UserId,
		DriverId: req.DriverId,
	})
	if err != nil {
		s.logger.Error("failed to get user login by driver id and user id", zap.Error(err))
	}
	_, err = s.userClient.InsertUserLogin(ctx, &usergrpcv1.UserLogin{
		UserId:      resp.UserId,
		DriverId:    req.DriverId,
		Token:       token,
		DriverToken: req.DriverToken,
		ClientType:  driveType,
		Platform:    req.Platform,
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
		msg := constants.WsMsg{Uid: constants.SystemNotification, Event: constants.SystemNotificationEvent, SendAt: time.Now(), Data: constants.SystemNotificationEventData{
			UserIds: []string{resp.UserId},
			Content: result,
			Type:    1,
		}}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.UserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.Notice, msg)
	}

	if s.cache {
		err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", resp.UserId))
		if err != nil {
			return nil, "", err
		}

	}

	return &model.UserInfoResponse{
		LoginNumber: data.ID,
		Email:       resp.Email,
		UserId:      resp.UserId,
		Nickname:    resp.NickName,
		Avatar:      resp.Avatar,
		Signature:   resp.Signature,
		CossId:      resp.CossId,
	}, token, nil
}

func (s *Service) Logout(ctx context.Context, userID string, token string, request *model.LogoutRequest, driverType string) error {
	value, err := s.redisClient.GetKey(userID + ":" + driverType + ":" + strconv.Itoa(int(request.LoginNumber)))
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
	err = s.redisClient.DelKey(userID + ":" + driverType + ":" + strconv.Itoa(int(request.LoginNumber)))
	if err != nil {
		return err
	}

	if s.cache {
		err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
		if err != nil {
			return err
		}

		err = s.redisClient.DelKey(fmt.Sprintf("user:%s", userID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Register(ctx context.Context, req *model.RegisterRequest) (string, error) {
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		req.Nickname = req.Email
	}

	img, err := govatar.Generate(avatar.RandomGender())
	if err != nil {
		return "", err
	}

	// 将图像编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return "", err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(buf.Bytes())
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.sp.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return "", err
		}
	}

	var UserId string
	workflow.InitGrpc(s.dtmGrpcServer, s.userGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "register_user_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		resp, err := s.userClient.UserRegister(wf.Context, &usergrpcv1.UserRegisterRequest{
			Email:           req.Email,
			NickName:        req.Nickname,
			Password:        utils.HashString(req.Password),
			ConfirmPassword: req.ConfirmPass,
			PublicKey:       req.PublicKey,
			Avatar:          aUrl,
		})
		if err != nil {
			s.logger.Error("failed to register user", zap.Error(err))
			if strings.Contains(err.Error(), "邮箱已被注册") {
				return code.UserErrEmailAlreadyRegistered
			}
			return err
		}
		UserId = resp.UserId

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userClient.CreateUserRollback(wf.Context, &usergrpcv1.CreateUserRollbackRequest{UserId: resp.UserId})
			return err
		})

		//TODO 系统账号统一管理
		_, err = s.userClient.UserInfo(wf.Context, &usergrpcv1.UserInfoRequest{UserId: constants.SystemNotification})
		if err != nil {
			s.logger.Error("failed to register user", zap.Error(err))
			return code.UserErrRegistrationFailed
		}

		//添加系统好友
		_, err = s.relClient.AddFriend(wf.Context, &relationgrpcv1.AddFriendRequest{
			UserId:   resp.UserId,
			FriendId: constants.SystemNotification,
		})
		if err != nil {
			s.logger.Error("failed to register user", zap.Error(err))
			return code.UserErrRegistrationFailed
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
		err = s.redisClient.SetKey(ekey, UserId, 30*ostime.Minute)
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
		CossId:    r.CossId,
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
		CossId:    r.CossId,
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
			OpenBurnAfterReading:        model.OpenBurnAfterReadingType(relation.OpenBurnAfterReading),
			SilentNotification:          model.SilentNotification(relation.IsSilent),
			Remark:                      relation.Remark,
			OpenBurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
		}
	}

	return resp, nil
}

func (s *Service) ModifyUserInfo(ctx context.Context, userID string, req *model.UserInfoRequest) error {

	//判断coosid是否存在
	if req.CossId != "" {
		pattern := "^[a-zA-Z0-9_]{10,20}$"
		if !regexp.MustCompile(pattern).MatchString(req.CossId) {
			return code.UserErrCossIdFormat
		}
		info, err := s.userClient.GetUserInfoByCossId(ctx, &usergrpcv1.GetUserInfoByCossIdlRequest{
			CossId: req.CossId,
		})
		if err != nil {
			c := code.Cause(err)
			fmt.Println("ModifyUserInfo c => ", c.Message())
			if !errors.Is(c, code.UserErrNotExist) {
				return err
			}
		}
		if info.UserId != "" {
			return code.UserErrCossIdAlreadyRegistered
		}
	}
	// 调用服务端设置用户公钥的方法
	_, err := s.userClient.ModifyUserInfo(ctx, &usergrpcv1.User{
		UserId:    userID,
		NickName:  req.NickName,
		Tel:       req.Tel,
		Avatar:    req.Avatar,
		Signature: req.Signature,
		CossId:    req.CossId,
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

	values, err := s.redisClient.GetAllListValues(userID)
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
	value, err := s.redisClient.GetKey(key)
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
	err = s.redisClient.DelKey(key)
	if err != nil {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	return resp, nil
}

func (s *Service) ResetUserPublicKey(ctx context.Context, req *model.ResetPublicKeyRequest) (interface{}, error) {
	//查询redis是否存在该验证码
	value, err := s.redisClient.GetKey(req.Code)
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

	err = s.redisClient.DelKey(req.Code)
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
	err = s.redisClient.SetKey(code1, info.UserId, 30*ostime.Minute)
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

// 修改用户头像
func (s *Service) ModifyUserAvatar(ctx context.Context, userID string, avatar multipart.File) (string, error) {
	_, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(avatar)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return "", err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(data)
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.sp.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return "", err
		}
	}

	_, err = s.userClient.ModifyUserInfo(ctx, &usergrpcv1.User{
		UserId: userID,
		Avatar: aUrl,
	})
	if err != nil {
		return "", err
	}

	if s.cache {
		//查询所有好友
		friends, err := s.relClient.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{
			UserId: userID,
		})

		for _, friend := range friends.FriendList {
			err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", friend.UserId))
			if err != nil {
				return "", err
			}

			err = s.redisClient.DelKey(fmt.Sprintf("friend:%s", friend.UserId))
			if err != nil {
				return "", err
			}
		}

	}

	return aUrl, nil
}
