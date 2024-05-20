package admin

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
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"image/png"
)

type AdminService interface {
	InitAdmin() error
	CreateAdmin(ctx context.Context, admin *entity.Admin) (interface{}, error)
	SendAllNotification(ctx context.Context, content string) (interface{}, error)
	GetAdminByUserID(ctx context.Context, userId string) (*entity.Admin, error)
}

func (s *ServiceImpl) CreateAdmin(ctx context.Context, admin *entity.Admin) (interface{}, error) {
	err := s.ad.InsertAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// 创建初始账号
func (s *ServiceImpl) InitAdmin() error {
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
		resp1, err := s.userService.CreateUser(wf.Context, &usergrpcv1.CreateUserRequest{
			UserId:   constants.SystemAdmin,
			Password: utils.HashString(Password),
			Email:    Email,
			NickName: Email,
			Avatar:   aUrl,
			IsBot:    1,
			Status:   usergrpcv1.UserStatus_USER_STATUS_NORMAL,
			//Status:   usergrpcv1.UserStatus_USER_STATUS_LOCK, //锁定账户

		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
		// 创建初始化数据
		resp2, err := s.userService.CreateUser(wf.Context, &usergrpcv1.CreateUserRequest{
			UserId:   constants.SystemNotification,
			Password: utils.HashString(Password),
			Email:    Email2,
			NickName: "系统通知",
			Avatar:   aUrl,
			IsBot:    1,
			Status:   usergrpcv1.UserStatus_USER_STATUS_NORMAL,
			//Status:   usergrpcv1.UserStatus_USER_STATUS_LOCK, //锁定账户

		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
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

		err = s.ad.InsertAndUpdateAdmin(wf.Context, &entity.Admin{UserId: constants.SystemAdmin, Role: entity.SuperAdminRole, Status: entity.NormalStatus})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
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
	//	err = cache.SetKey(s.redisClient, ekey, resp.UserID, 30*ostime.Minute)
	//	if err != nil {
	//		return "", err
	//	}
	//
	//	//注册成功发送邮件
	//	err = s.smtpClient.SendEmail(req.Email, "欢迎注册", s.smtpClient.GenerateEmailVerificationContent(s.gatewayAddress+s.gatewayPort, resp.UserID, ekey))
	//	if err != nil {
	//		s.logger.Error("failed to send email", zap.Error(err))
	//		return "", err
	//	}
	//}

	return nil
}

func (s *ServiceImpl) SendAllNotification(ctx context.Context, content string) (interface{}, error) {
	UserId := constants.SystemNotification

	//查询系统通知账号的所有好友
	list, err := s.relationUserService.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: UserId})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, v := range list.FriendList {
		ids = append(ids, v.UserId)
	}
	//获取系统通知账号信息
	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: UserId})
	if err != nil {
		return nil, err
	}

	msgList := make([]*msggrpcv1.SendUserMsgRequest, 0)
	//sendList := make([]*pushgrpcv1.SendWsUserMsg, 0)
	pushList := make([]*pushgrpcv1.WsMsg, 0)
	for _, v := range ids {
		relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: v, FriendId: UserId})
		if err != nil {
			return nil, err
		}

		//获取用户对话信息
		dialogUser, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
			DialogId: relation.DialogId,
			UserId:   v,
		})
		if err != nil {
			s.logger.Error("获取用户会话失败", zap.Error(err))
			return nil, err
		}

		//打开对话
		if dialogUser.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
			_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
				DialogId: relation.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
				UserId:   v,
			})
			if err != nil {
				return nil, err
			}
		}

		msg2 := &msggrpcv1.SendUserMsgRequest{
			SenderId:   UserId,
			ReceiverId: v,
			Content:    content,
			DialogId:   relation.DialogId,
			Type:       int32(msggrpcv1.MessageType_Text),
		}

		send := &pushgrpcv1.SendWsUserMsg{
			SenderId:                constants.SystemNotification,
			Content:                 content,
			MsgType:                 uint32(msggrpcv1.MessageType_Text),
			ReceiverId:              v,
			SendAt:                  pkgtime.Now(),
			DialogId:                relation.DialogId,
			BurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
			SenderInfo: &pushgrpcv1.SenderInfo{
				Avatar: info.Avatar,
				Name:   info.NickName,
				UserId: info.UserId,
			},
		}

		toBytes, err := utils.StructToBytes(send)
		if err != nil {
			s.logger.Error("failed to send push", zap.Error(err))
			continue
		}

		push := &pushgrpcv1.WsMsg{
			Uid:   v,
			Event: pushgrpcv1.WSEventType_SendUserMessageEvent,
			Data:  &any.Any{Value: toBytes},
		}

		pushList = append(pushList, push)
		//sendList = append(sendList, send)
		msgList = append(msgList, msg2)
	}

	_, err = s.msgService.SendMultiUserMessage(context.Background(), &msggrpcv1.SendMultiUserMsgRequest{MsgList: msgList})
	if err != nil {
		s.logger.Error("发送系统通知失败", zap.Error(err))
		return nil, err
	}

	msg := pushgrpcv1.PushWsBatchRequest{Msgs: pushList}
	if err != nil {
		return nil, err
	}

	data, err := utils.StructToBytes(&msg)
	if err != nil {
		return nil, err
	}
	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws_Batch,
		Data: data,
	})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	return nil, nil
}

func (s *ServiceImpl) GetAdminByUserID(ctx context.Context, userId string) (*entity.Admin, error) {
	find, err := s.ad.Find(ctx, &entity.Query{UserId: &userId})
	if err != nil {
		return nil, err
	}
	if len(find) == 0 {
		return nil, code.MyCustomErrorCode.CustomMessage("非管理员")
	}
	return find[0], nil
}
