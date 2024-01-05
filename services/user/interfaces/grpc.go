package interfaces

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"time"

	api "github.com/cossim/coss-server/services/user/api/v1"
	"github.com/cossim/coss-server/services/user/domain/service"
	"github.com/cossim/coss-server/services/user/infrastructure/persistence"
)

type GrpcHandler struct {
	svc *service.UserService
	api.UnimplementedUserServiceServer
}

func NewGrpcHandler(c *config.AppConfig) *GrpcHandler {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}

	return &GrpcHandler{
		svc: service.NewUserService(persistence.NewUserRepo(dbConn)),
	}
}

// 用户登录
func (g *GrpcHandler) UserLogin(ctx context.Context, request *api.UserLoginRequest) (*api.UserLoginResponse, error) {
	userInfo, err := g.svc.Login(request)
	if err != nil {
		return nil, err
	}
	if userInfo.Password != request.Password {
		return nil, fmt.Errorf("密码错误")
	}
	//修改登录时间
	userInfo.LastAt = time.Now().Unix()
	_, err = g.svc.UpdateUserInfo(userInfo)
	if err != nil {
		return nil, err
	}
	//参数校验
	return &api.UserLoginResponse{
		UserId:   userInfo.ID,
		NickName: userInfo.NickName,
		Email:    userInfo.Email,
		Tel:      userInfo.Tel,
		Avatar:   userInfo.Avatar,
	}, nil
}

// 用户注册
func (g *GrpcHandler) UserRegister(ctx context.Context, request *api.UserRegisterRequest) (*api.UserRegisterResponse, error) {
	//添加用户
	userInfo, err := g.svc.Register(request)
	if err != nil {
		return nil, err
	}
	return &api.UserRegisterResponse{
		UserId: userInfo.ID,
	}, nil
}

func (g *GrpcHandler) UserInfo(ctx context.Context, request *api.UserInfoRequest) (*api.UserInfoResponse, error) {
	userInfo, err := g.svc.GetUserInfo(request.UserId)
	if err != nil {
		return nil, err
	}
	return &api.UserInfoResponse{
		UserId:   userInfo.ID,
		NickName: userInfo.NickName,
		Email:    userInfo.Email,
		Tel:      userInfo.Tel,
		Avatar:   userInfo.Avatar,
		Status:   api.UserStatus(userInfo.Status),
	}, nil
}

func (g *GrpcHandler) GetBatchUserInfo(ctx context.Context, request *api.GetBatchUserInfoRequest) (*api.GetBatchUserInfoResponse, error) {
	resp := &api.GetBatchUserInfoResponse{}
	users, err := g.svc.GetBatchGetUserInfo(request.UserIds)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		resp.Users = append(resp.Users, &api.UserInfoResponse{
			UserId:   user.ID,
			NickName: user.NickName,
			Email:    user.Email,
			Tel:      user.Tel,
			Avatar:   user.Avatar,
			Status:   api.UserStatus(user.Status),
		})
	}

	return resp, nil
}
