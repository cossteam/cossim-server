package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/api/http/model"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/user"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/avatar"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"image/png"
	"io/ioutil"
	"mime/multipart"
	"regexp"
	"strings"
)

func (s *Service) Login(ctx context.Context, req *model.LoginRequest, clientIp string) (*model.UserInfoResponse, string, error) {
	userInfo, err := s.userService.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{Email: req.Email})
	if err != nil {
		s.logger.Error("failed to get user info by email", zap.Error(err))
		return nil, "", err
	}

	id, err := s.userLoginService.GetUserLoginByUserId(ctx, &usergrpcv1.GetUserLoginByUserIdRequest{UserId: userInfo.UserId})
	if err != nil {
		s.logger.Error("failed to get user login by user id", zap.Error(err))
	}
	var lastLoginTime int64 = 0
	if id != nil {
		lastLoginTime = id.LoginTime
	}

	token, err := utils.GenerateToken(userInfo.UserId, userInfo.Email, req.DriverId, userInfo.PublicKey)
	if err != nil {
		s.logger.Error("failed to generate user token", zap.Error(err))
		return nil, "", code.UserErrLoginFailed
	}

	infos, err := s.userCache.GetUserLoginInfos(ctx, userInfo.UserId)
	if err != nil {
		return nil, "", err
	}

	//多端设备登录限制
	if s.ac.MultipleDeviceLimit.Enable {
		if len(infos) >= s.ac.MultipleDeviceLimit.Max {
			s.logger.Error("user login failed", zap.Error(err))
			return nil, "", code.MyCustomErrorCode.CustomMessage("登录设备超出限制")
		}
	}

	// 推送登录提醒
	// 查询是否在该设备第一次登录
	userLogin, err := s.userLoginService.GetUserLoginByDriverIdAndUserId(ctx, &usergrpcv1.DriverIdAndUserId{
		UserId:   userInfo.UserId,
		DriverId: req.DriverId,
	})
	if err != nil {
		s.logger.Error("failed to get user login by driver driveID and user driveID", zap.Error(err))
	}

	fristLogin := false
	if userLogin != nil {
		if userLogin.DriverId == "" {
			fristLogin = true
		}
	}

	driveID := len(infos) + 1
	data2 := user.UserLogin{
		ID:        uint(driveID),
		UserId:    userInfo.UserId,
		Token:     token,
		CreatedAt: time.Now(),
		ClientIP:  clientIp,
	}

	workflow.InitGrpc(s.dtmGrpcServer, s.userGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "login_user_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		if err := s.userCache.SetUserLoginInfo(wf.Context, userInfo.UserId, driveID, &data2, cache.UserLoginExpireTime); err != nil {
			s.logger.Error("failed to set user login info", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			err = s.userCache.DeleteUserLoginInfo(wf.Context, userInfo.UserId, driveID)
			return err
		})

		ul, err := s.userLoginService.InsertUserLogin(ctx, &usergrpcv1.UserLogin{
			UserId:      userInfo.UserId,
			DriverId:    req.DriverId,
			Token:       token,
			DriverToken: req.DriverToken,
			Platform:    req.Platform,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userLoginService.DeleteUserLoginByID(wf.Context, &usergrpcv1.UserLoginIDRequest{ID: ul.ID})
			return err
		})

		if fristLogin {
			_, msgId, err := s.pushFirstLogin(ctx, userInfo, "", req.DriverId)
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}

			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
				_, err := s.msgService.DeleteUserMessageById(ctx, &msggrpcv1.DeleteUserMsgByIDRequest{
					ID: msgId,
				})
				return err
			})
		}

		_, err = s.userService.UserLogin(ctx, &usergrpcv1.UserLoginRequest{
			Email:    req.Email,
			Password: utils.HashString(req.Password),
		})
		if err != nil {
			s.logger.Error("user login failed", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		return nil
	}); err != nil {
		return nil, "", err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		if strings.Contains(err.Error(), "用户不存在或密码错误") {
			return nil, "", code.UserErrNotExistOrPassword
		}
		return nil, "", code.UserErrLoginFailed
	}

	return &model.UserInfoResponse{
		LoginNumber:    data2.ID,
		Status:         model.UserStatus(userInfo.Status),
		Email:          userInfo.Email,
		UserId:         userInfo.UserId,
		Nickname:       userInfo.NickName,
		Avatar:         userInfo.Avatar,
		Signature:      userInfo.Signature,
		CossId:         userInfo.CossId,
		NewDeviceLogin: fristLogin,
		LastLoginTime:  lastLoginTime,
	}, token, nil
}

func (s *Service) Logout(ctx context.Context, userID string, token string, request *model.LogoutRequest) error {
	loginInfo, err := s.userCache.GetUserLoginInfo(ctx, userID, int(request.LoginNumber))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return code.MyCustomErrorCode.CustomMessage("登录信息不存在")
		}
		s.logger.Error("failed to get user login info", zap.Error(err))
		return err
	}

	//通知消息服务关闭ws
	rid := loginInfo.Rid

	data2 := &constants.OfflineEventData{
		Rid: rid,
	}

	toBytes, err := utils.StructToBytes(data2)
	if err != nil {
		s.logger.Error("failed to struct to bytes", zap.Error(err))
		return err
	}

	if rid != "" {
		msg := &pushgrpcv1.WsMsg{
			Uid:    userID,
			Event:  pushgrpcv1.WSEventType_OfflineEvent,
			Data:   &any.Any{Value: toBytes},
			SendAt: time.Now(),
			Rid:    loginInfo.Rid,
		}
		toBytes2, err := utils.StructToBytes(msg)
		if err != nil {
			return err
		}

		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Type: pushgrpcv1.Type_Ws, Data: toBytes2})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}

	}

	// 删除客户端信息
	if err := s.userCache.DeleteUserLoginInfo(ctx, userID, int(request.LoginNumber)); err != nil {
		s.logger.Error("failed to delete user login info", zap.Error(err))
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
	err = s.storageService.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
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
		resp, err := s.userService.UserRegister(wf.Context, &usergrpcv1.UserRegisterRequest{
			Email:           req.Email,
			NickName:        req.Nickname,
			Password:        utils.HashString(req.Password),
			ConfirmPassword: req.ConfirmPass,
			PublicKey:       req.PublicKey,
			Avatar:          aUrl,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
		UserId = resp.UserId

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userService.CreateUserRollback(wf.Context, &usergrpcv1.CreateUserRollbackRequest{UserId: resp.UserId})
			return err
		})

		_, err = s.userService.UserInfo(wf.Context, &usergrpcv1.UserInfoRequest{UserId: constants.SystemNotification})
		if err != nil {
			s.logger.Error("failed to register user", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		//添加系统好友
		_, err = s.relationService.AddFriend(wf.Context, &relationgrpcv1.AddFriendRequest{
			UserId:   resp.UserId,
			FriendId: constants.SystemNotification,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		return err
	}); err != nil {
		return "", err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		if strings.Contains(err.Error(), "邮箱已被注册") {
			return "", code.UserErrEmailAlreadyRegistered
		}
		return "", code.UserErrRegistrationFailed
	}

	if s.ac.Email.Enable {
		ekey := uuid.New().String()

		if err = s.userCache.SetUserEmailVerificationCode(ctx, UserId, ekey, cache.UserEmailVerificationCodeExpireTime); err != nil {
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
	r, err := s.userService.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
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

	relation, err := s.relationService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
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
	r, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
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
		Tel:       r.Tel,
		Signature: r.Signature,
		CossId:    r.CossId,
		Status:    model.UserStatus(r.Status),
	}

	relation, err := s.relationService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   thisID,
		FriendId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
	}

	if relation != nil {
		if thisID != userID && relation.UserId != "" {
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
		info, err := s.userService.GetUserInfoByCossId(ctx, &usergrpcv1.GetUserInfoByCossIdlRequest{
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

	_, err := s.userService.ModifyUserInfo(ctx, &usergrpcv1.User{
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
	info, err := s.userService.GetUserPasswordByUserId(ctx, &usergrpcv1.UserRequest{
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
	_, err = s.userService.ModifyUserPassword(ctx, &usergrpcv1.ModifyUserPasswordRequest{
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
	_, err := s.userService.SetUserPublicKey(ctx, &usergrpcv1.SetPublicKeyRequest{
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
	_, err := s.userService.SetUserSecretBundle(context.Background(), &usergrpcv1.SetUserSecretBundleRequest{
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
	_, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, code.UserErrGetUserInfoFailed
	}

	bundle, err := s.userService.GetUserSecretBundle(context.Background(), &usergrpcv1.GetUserSecretBundleRequest{
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
	users, err := s.userCache.GetUserLoginInfos(ctx, userID)
	if err != nil {
		return nil, err
	}

	var clients []*model.GetUserLoginClientsResponse
	for _, user := range users {
		clients = append(clients, &model.GetUserLoginClientsResponse{
			ClientIP:    user.ClientIP,
			LoginNumber: user.ID,
			LoginAt:     user.CreatedAt,
		})
	}
	return clients, nil
}

func (s *Service) UserActivate(ctx context.Context, userID string, key string) (interface{}, error) {
	verificationCode, err2 := s.userCache.GetUserEmailVerificationCode(ctx, userID)
	if err2 != nil {
		return nil, err2
	}

	if verificationCode != key {
		return nil, code.MyCustomErrorCode.CustomMessage("验证码不存在或已过期")
	}

	resp, err := s.userService.ActivateUser(ctx, &usergrpcv1.UserRequest{
		UserId: userID,
	})

	if err != nil {
		s.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	if err := s.userCache.DeleteUserEmailVerificationCode(ctx, userID); err != nil {
		s.logger.Error("删除用户邮箱激活验证码失败", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) ResetUserPublicKey(ctx context.Context, userID string, req *model.ResetPublicKeyRequest) (interface{}, error) {
	info, err := s.userService.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		s.logger.Error("重置用户公钥失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	verificationCode, err := s.userCache.GetUserVerificationCode(ctx, userID, req.Code)
	if err != nil {
		return nil, err
	}

	if verificationCode != req.Code {
		return nil, code.MyCustomErrorCode.CustomMessage("验证码不存在或已过期")
	}

	_, err = s.userService.SetUserPublicKey(ctx, &usergrpcv1.SetPublicKeyRequest{
		UserId:    info.UserId,
		PublicKey: req.PublicKey,
	})
	if err != nil {
		return nil, err
	}

	if err := s.userCache.DeleteUserVerificationCode(ctx, userID, req.Code); err != nil {
		s.logger.Error("删除用户邮箱验证码失败", zap.Error(err))
	}

	return verificationCode, nil
}

func (s *Service) SendEmailCode(ctx context.Context, email string) (interface{}, error) {
	info, err := s.userService.GetUserInfoByEmail(ctx, &usergrpcv1.GetUserInfoByEmailRequest{
		Email: email,
	})
	if err != nil {
		s.logger.Error("发送邮箱验证码失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	code1 := utils.RandomNum()
	if err := s.userCache.SetUserVerificationCode(ctx, info.UserId, code1, cache.UserVerificationCodeExpireTime); err != nil {
		s.logger.Error("发送邮箱验证码失败", zap.Error(err))
		return nil, err
	}

	if s.ac.Email.Enable {
		if err := s.smtpClient.SendEmail(email, "重置pgp验证码(请妥善保管,有效时间5分钟)", code1); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// 修改用户头像
func (s *Service) ModifyUserAvatar(ctx context.Context, userID string, avatar multipart.File) (string, error) {
	_, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
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
	err = s.storageService.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
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

	_, err = s.userService.ModifyUserInfo(ctx, &usergrpcv1.User{
		UserId: userID,
		Avatar: aUrl,
	})
	if err != nil {
		return "", err
	}

	return aUrl, nil
}

func (s *Service) pushFirstLogin(ctx context.Context, userInfo *usergrpcv1.UserInfoResponse, clientIp, driverId string) (uint32, uint32, error) {
	if clientIp == "127.0.0.1" {
		clientIp = httputil.GetMyPublicIP()
	}
	info := httputil.OnlineIpInfo(clientIp)

	result := fmt.Sprintf("您在新设备登录，IP地址为：%s\n位置为：%s %s %s", clientIp, info.Country, info.RegionName, info.City)
	if info.RegionName == info.City {
		result = fmt.Sprintf("您在新设备登录，IP地址为：%s\n位置为：%s %s", clientIp, info.Country, info.City)
	}

	//查询系统通知信息
	systemInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: constants.SystemNotification})
	if err != nil {
		return 0, 0, err
	}
	//查询与系统通知的对话id
	relation, err := s.relationService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userInfo.UserId,
		FriendId: constants.SystemNotification,
	})
	if err != nil {
		return 0, 0, err
	}

	//查询与系统的对话是否关闭
	dialogUser, err := s.dialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: relation.DialogId,
		UserId:   userInfo.UserId,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return 0, 0, err
	}

	if dialogUser.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
		_, err := s.dialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
			DialogId: relation.DialogId,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserId:   userInfo.UserId,
		})
		if err != nil {
			return 0, 0, err
		}
	}

	//插入消息
	msg2 := &msggrpcv1.SendUserMsgRequest{
		SenderId:   constants.SystemNotification,
		ReceiverId: userInfo.UserId,
		Content:    result,
		DialogId:   relation.DialogId,
		Type:       int32(msggrpcv1.MessageType_Text),
	}

	id, err := s.msgService.SendUserMessage(ctx, msg2)
	if err != nil {
		return 0, 0, err
	}

	rmsg := &pushgrpcv1.SendWsUserMsg{
		MsgType:    uint32(msggrpcv1.MessageType_Text),
		DialogId:   relation.DialogId,
		Content:    result,
		SenderId:   constants.SystemNotification,
		ReceiverId: userInfo.UserId,
		SendAt:     time.Now(),
		MsgId:      id.MsgId,
		SenderInfo: &pushgrpcv1.SenderInfo{
			UserId: systemInfo.UserId,
			Avatar: systemInfo.Avatar,
			Name:   systemInfo.NickName,
		},
	}

	toBytes, err := utils.StructToBytes(rmsg)
	if err != nil {
		return 0, 0, err
	}

	msg := &pushgrpcv1.WsMsg{Uid: userInfo.UserId, DriverId: driverId, Event: pushgrpcv1.WSEventType_SendUserMessageEvent, PushOffline: true, SendAt: time.Now(), Data: &any.Any{Value: toBytes}}

	toBytes2, err := utils.StructToBytes(msg)
	if err != nil {
		return 0, 0, err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Type: pushgrpcv1.Type_Ws, Data: toBytes2})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}
	return relation.DialogId, id.MsgId, nil
}
