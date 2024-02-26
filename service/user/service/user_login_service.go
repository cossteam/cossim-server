package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils/time"
	v1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/cossim/coss-server/service/user/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Service) InsertUserLogin(ctx context.Context, in *v1.UserLogin) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	info, err := s.ulr.GetUserLoginByDriverIdAndUserId(in.DriverId, in.UserId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}
	if info != nil {
		err := s.ulr.InsertUserLogin(&entity.UserLogin{
			UserId:      in.UserId,
			Token:       in.Token,
			DriverId:    in.DriverId,
			LastAt:      time.Now(),
			DriverToken: in.DriverToken,
			LoginCount:  info.LoginCount + 1,
		})
		if err != nil {
			return nil, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	} else {
		err := s.ulr.InsertUserLogin(&entity.UserLogin{
			UserId:      in.UserId,
			Token:       in.Token,
			DriverId:    in.DriverId,
			LastAt:      time.Now(),
			DriverToken: in.DriverToken,
			LoginCount:  1,
		})
		if err != nil {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}

	return resp, nil
}

func (s *Service) GetUserLoginByToken(ctx context.Context, in *v1.GetUserLoginByTokenRequest) (*v1.UserLogin, error) {
	resp := &v1.UserLogin{}
	info, err := s.ulr.GetUserLoginByToken(in.Token)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
		}
	}
	if info != nil {
		resp.UserId = info.UserId
		resp.Token = info.Token
		resp.DriverId = info.DriverId
	}
	return resp, nil
}

func (s *Service) GetUserLoginByDriverIdAndUserId(ctx context.Context, in *v1.DriverIdAndUserId) (*v1.UserLogin, error) {
	resp := &v1.UserLogin{}
	info, err := s.ulr.GetUserLoginByDriverIdAndUserId(in.DriverId, in.UserId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.UserErrGetUserLoginByDriverIdAndUserIdFailed.Code()), err.Error())
		}
	}
	if info != nil {
		resp.UserId = info.UserId
		resp.Token = info.Token
		resp.DriverId = info.DriverId
	}
	return resp, nil
}

func (s *Service) UpdateUserLoginTokenByDriverId(ctx context.Context, in *v1.TokenUpdate) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	err := s.ulr.UpdateUserLoginTokenByDriverId(in.DriverId, in.Token, in.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.UserErrUpdateUserLoginTokenFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetUserDriverTokenByUserId(ctx context.Context, request *v1.GetUserDriverTokenByUserIdRequest) (*v1.GetUserDriverTokenByUserIdResponse, error) {
	resp := &v1.GetUserDriverTokenByUserIdResponse{}
	tokenList, err := s.ulr.GetUserDriverTokenByUserId(request.UserId)
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
