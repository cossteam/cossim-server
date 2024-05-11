package grpc

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *UserServiceServer) InsertUserLogin(ctx context.Context, request *v1.UserLogin) (*v1.InsertUserLoginResponse, error) {
	resp := &v1.InsertUserLoginResponse{}
	info, err := s.ulr.GetUserLoginByDriverIdAndUserId(ctx, request.DriverId, request.UserId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}
	var info2 *entity.UserLogin
	if info != nil {
		info2 = &entity.UserLogin{
			UserID:      request.UserId,
			Token:       request.Token,
			DriverID:    request.DriverId,
			LastAt:      time.Now(),
			DriverToken: request.DriverToken,
			Platform:    request.Platform,
			LoginCount:  info.LoginCount + 1,
		}
		err := s.ulr.InsertUserLogin(ctx, info2)
		if err != nil {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	} else {
		info2 = &entity.UserLogin{
			UserID:      request.UserId,
			Token:       request.Token,
			DriverID:    request.DriverId,
			LastAt:      time.Now(),
			DriverToken: request.DriverToken,
			Platform:    request.Platform,
			LoginCount:  1,
		}
		err := s.ulr.InsertUserLogin(ctx, info2)
		if err != nil {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}

	resp.ID = uint32(info2.ID)

	fmt.Println("resp.ID", resp.ID)
	return resp, nil
}

func (s *UserServiceServer) GetUserLoginByToken(ctx context.Context, request *v1.GetUserLoginByTokenRequest) (*v1.UserLogin, error) {
	resp := &v1.UserLogin{}
	info, err := s.ulr.GetUserLoginByToken(ctx, request.Token)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}
	if info != nil {
		resp.UserId = info.UserID
		resp.Token = info.Token
		resp.DriverId = info.DriverID
	}
	return resp, nil
}

func (s *UserServiceServer) GetUserLoginByDriverIdAndUserId(ctx context.Context, request *v1.DriverIdAndUserId) (*v1.UserLogin, error) {
	resp := &v1.UserLogin{}
	info, err := s.ulr.GetUserLoginByDriverIdAndUserId(ctx, request.DriverId, request.UserId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrGetUserLoginByDriverIdAndUserIdFailed.Code()), err.Error())
		}
	}

	if info != nil {
		resp.UserId = info.UserID
		resp.Token = info.Token
		resp.DriverId = info.DriverID
		resp.DriverToken = info.DriverToken
		resp.Platform = info.Platform
		resp.LoginTime = info.LastAt
	}
	return resp, nil
}

func (s *UserServiceServer) UpdateUserLoginTokenByDriverId(ctx context.Context, request *v1.TokenUpdate) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	err := s.ulr.UpdateUserLoginTokenByDriverId(ctx, request.DriverId, request.Token, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.UserErrUpdateUserLoginTokenFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *UserServiceServer) GetUserDriverTokenByUserId(ctx context.Context, request *v1.GetUserDriverTokenByUserIdRequest) (*v1.GetUserDriverTokenByUserIdResponse, error) {
	resp := &v1.GetUserDriverTokenByUserIdResponse{}
	tokenList, err := s.ulr.GetUserDriverTokenByUserId(ctx, request.UserId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrGetUserDriverTokenByUserIdFailed.Code()), err.Error())
		}
	}
	if len(tokenList) > 0 {
		for _, token := range tokenList {
			resp.Token = append(resp.Token, token)
		}
	}
	return resp, nil
}

func (s *UserServiceServer) GetUserLoginByUserId(ctx context.Context, request *v1.GetUserLoginByUserIdRequest) (*v1.UserLogin, error) {
	var resp = &v1.UserLogin{}
	info, err := s.ulr.GetUserByUserId(ctx, request.UserId)
	if err != nil {
		return nil, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
	}
	if info != nil {
		resp.UserId = info.UserID
		resp.Token = info.Token
		resp.DriverId = info.DriverID
		resp.DriverToken = info.DriverToken
		resp.Platform = info.Platform
		resp.LoginTime = info.LastAt
	}
	return resp, nil
}

func (s *UserServiceServer) DeleteUserLoginByID(ctx context.Context, request *v1.UserLoginIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	err := s.ulr.DeleteUserLoginByID(ctx, request.ID)
	if err != nil {
		return resp, status.Error(codes.Code(code.UserErrDeleteUserLoginByIDFailed.Code()), err.Error())
	}
	return resp, nil
}
