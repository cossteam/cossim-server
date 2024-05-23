package grpc

import (
	"context"
	"errors"
	api "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
)

func (s *UserServiceServer) UserInfo(ctx context.Context, request *api.UserInfoRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}

	userInfo, err := s.ur.GetWithOptions(ctx, &entity.User{
		ID: request.UserId,
	})
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.UserErrNotExist)
		}
		return nil, code.WrapCodeToGRPC(code.UserErrGetUserInfoFailed.Reason(utils.FormatErrorStack(err)))
	}
	resp = &api.UserInfoResponse{
		UserId:    userInfo.ID,
		NickName:  userInfo.NickName,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    api.UserStatus(userInfo.Status),
		CossId:    userInfo.CossID,
	}

	return resp, nil
}

func (s *UserServiceServer) GetBatchUserInfo(ctx context.Context, request *api.GetBatchUserInfoRequest) (*api.GetBatchUserInfoResponse, error) {
	resp := &api.GetBatchUserInfoResponse{}
	if len(request.UserIds) == 0 {
		return resp, nil
	}

	users, err := s.ur.ListUser(ctx, &entity.ListUserOptions{
		UserID: request.UserIds,
	})
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.UserErrUnableToGetUserListInfo.Reason(utils.FormatErrorStack(err)))
	}
	for _, user := range users {
		resp.Users = append(resp.Users, &api.UserInfoResponse{
			UserId:    user.ID,
			NickName:  user.NickName,
			Email:     user.Email,
			Tel:       user.Tel,
			Avatar:    user.Avatar,
			Signature: user.Signature,
			Status:    api.UserStatus(user.Status),
			CossId:    user.CossID,
		})
	}

	return resp, nil
}

func (s *UserServiceServer) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	resp := &api.CreateUserResponse{}

	if err := s.ur.InsertAndUpdateUser(ctx, &entity.User{
		NickName:  request.NickName,
		Email:     request.Email,
		Password:  request.Password,
		Avatar:    request.Avatar,
		Status:    entity.UserStatus(request.Status),
		ID:        request.UserId,
		PublicKey: request.PublicKey,
		Bot:       uint(request.IsBot),
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.UserErrCreateUserFailed.Reason(utils.FormatErrorStack(err)))
	}
	resp.UserId = request.UserId
	return resp, nil
}

func (s *UserServiceServer) CreateUserRollback(ctx context.Context, request *api.CreateUserRollbackRequest) (*api.CreateUserRollbackResponse, error) {
	resp := &api.CreateUserRollbackResponse{}
	if err := s.ur.DeleteUser(ctx, request.UserId); err != nil {
		return nil, code.WrapCodeToGRPC(code.UserErrCreateUserRollbackFailed.Reason(utils.FormatErrorStack(err)))
	}
	return resp, nil
}
