package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/constants"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"image/png"
)

func (s *Service) CreateAdmin(admin *entity.Admin) (interface{}, error) {
	err := s.repo.Ar.InsertAdmin(admin)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// 创建初始账号
func (s *Service) InitAdmin() error {
	UserId := "10000"
	Email := "admin@admin.com"
	Password := "123123a"
	Email2 := "tz@bot.com"

	img, err := govatar.GenerateForUsername(govatar.MALE, Email2)
	if err != nil {
		return err
	}

	// 将图像编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		panic(err)
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(buf.Bytes())
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.sp.UploadAvatar(context.Background(), key, reader, reader.Size(), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return err
		}
	}

	workflow.InitGrpc(s.dtmGrpcServer, s.userServiceAddr, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "init_admin_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建初始化数据
		resp1, err := s.userService.CreateUser(context.Background(), &usergrpcv1.CreateUserRequest{
			UserId:   UserId,
			Password: utils.HashString(Password),
			Email:    Email,
			NickName: Email,
			Avatar:   aUrl,
			Status:   usergrpcv1.UserStatus_USER_STATUS_NORMAL,
			//Status:   usergrpcv1.UserStatus_USER_STATUS_LOCK, //锁定账户

		})
		if err != nil {
			return err
		}
		// 创建初始化数据
		resp2, err := s.userService.CreateUser(context.Background(), &usergrpcv1.CreateUserRequest{
			UserId:   "10001",
			Password: utils.HashString(Password),
			Email:    Email2,
			NickName: "系统通知",
			Avatar:   aUrl,
			IsBot:    1,
			Status:   usergrpcv1.UserStatus_USER_STATUS_NORMAL,
			//Status:   usergrpcv1.UserStatus_USER_STATUS_LOCK, //锁定账户

		})
		if err != nil {
			return err
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userService.CreateUserRollback(context.Background(), &usergrpcv1.CreateUserRollbackRequest{UserId: resp1.UserId})
			if err != nil {
				return err
			}

			_, err = s.userService.CreateUserRollback(context.Background(), &usergrpcv1.CreateUserRollbackRequest{UserId: resp2.UserId})
			if err != nil {
				return err
			}
			return nil
		})

		err = s.repo.Ar.InsertAndUpdateAdmin(&entity.Admin{UserId: UserId, Role: entity.SuperAdminRole, Status: entity.NormalStatus})
		if err != nil {
			return err
		}
		return err
	}); err != nil {
		return err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return err
	}
	//TODO 是否需要激活管理员账号
	//if s.ac.Email.Enable {
	//	//生成uuid
	//	ekey := uuid.New().String()
	//
	//	//保存到redis
	//	err = cache.SetKey(s.redisClient, ekey, resp.UserId, 30*ostime.Minute)
	//	if err != nil {
	//		return "", err
	//	}
	//
	//	//注册成功发送邮件
	//	err = s.smtpClient.SendEmail(req.Email, "欢迎注册", s.smtpClient.GenerateEmailVerificationContent(s.gatewayAddress+s.gatewayPort, resp.UserId, ekey))
	//	if err != nil {
	//		s.logger.Error("failed to send email", zap.Error(err))
	//		return "", err
	//	}
	//}

	return nil
}

func (s *Service) SendAllNotification(ctx context.Context, content string) (interface{}, error) {
	UserId := constants.SystemNotification

	//TODO 系统账号统一管理
	//查询系统通知账号的所有好友
	list, err := s.relationUserService.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: UserId})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, v := range list.FriendList {
		ids = append(ids, v.UserId)
	}

	msgList := make([]*msggrpcv1.SendUserMsgRequest, 0)

	for _, v := range ids {
		relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: v, FriendId: UserId})
		if err != nil {
			return nil, err
		}
		msg2 := &msggrpcv1.SendUserMsgRequest{
			SenderId:   UserId,
			ReceiverId: v,
			Content:    content,
			DialogId:   relation.DialogId,
			Type:       1, //todo 消息类型枚举
		}
		msgList = append(msgList, msg2)
	}

	_, err = s.msgService.SendMultiUserMessage(context.Background(), &msggrpcv1.SendMultiUserMsgRequest{MsgList: msgList})
	if err != nil {
		s.logger.Error("发送系统通知失败", zap.Error(err))
		return nil, err
	}

	toBytes, err := utils.StructToBytes(&constants.SystemNotificationEventData{
		UserId:  UserId,
		Content: content,
		Type:    1,
	})
	msg := pushgrpcv1.PushWsBatchByUserIdsRequest{UserIds: ids, Event: pushgrpcv1.WSEventType_SystemNotificationEvent, PushOffline: true, Data: &any.Any{Value: toBytes}}
	if err != nil {
		return nil, err
	}

	data, err := utils.StructToBytes(&msg)

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws_Batch_User,
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
