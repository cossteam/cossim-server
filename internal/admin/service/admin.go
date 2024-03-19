package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
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

	workflow.InitGrpc(s.dtmGrpcServer, s.userGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "init_admin_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建初始化数据
		resp1, err := s.userClient.CreateUser(context.Background(), &usergrpcv1.CreateUserRequest{
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
		resp2, err := s.userClient.CreateUser(context.Background(), &usergrpcv1.CreateUserRequest{
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
			_, err = s.userClient.CreateUserRollback(context.Background(), &usergrpcv1.CreateUserRollbackRequest{UserId: resp1.UserId})
			if err != nil {
				return err
			}

			_, err = s.userClient.CreateUserRollback(context.Background(), &usergrpcv1.CreateUserRollbackRequest{UserId: resp2.UserId})
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
	UserId := "10001"
	fmt.Println("relationClient", s.relationClient)
	fmt.Println("msgClient", s.msgClient)
	//TODO 系统账号统一管理
	//查询系统通知账号的所有好友
	list, err := s.relationClient.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: UserId})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, v := range list.FriendList {
		ids = append(ids, v.UserId)
	}

	err = s.publishServiceNoticeMessage(ctx, constants.WsMsg{Uid: UserId, Event: constants.SystemNotificationEvent, SendAt: time.Now(), Data: constants.SystemNotificationEventData{
		UserIds: ids,
		Content: content,
		Type:    1,
	}})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) publishServiceNoticeMessage(ctx context.Context, msg constants.WsMsg) error {
	return s.rabbitMQClient.PublishServiceMessage(msg_queue.AdminService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.Notice, msg)
}
