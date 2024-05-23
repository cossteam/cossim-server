package grpc

import (
	"context"
	"errors"
	api "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
)

var _ api.UserAuthServiceServer = &UserServiceServer{}

func (s *UserServiceServer) ParseToken(ctx context.Context, request *api.ParseTokenRequest) (*api.AuthClaims, error) {
	_, claims, err := utils.ParseToken(request.Token, s.secret)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.InvalidParameter.CustomMessage("token解析失败").Reason(utils.FormatErrorStack(err)))
	}

	return &api.AuthClaims{
		UserID:    claims.UserId,
		Email:     claims.Email,
		DriverID:  claims.DriverId,
		PublicKey: claims.PublicKey,
		ExpireAt:  claims.ExpiresAt.UnixNano() / 1e6,
	}, nil
}

func (s *UserServiceServer) GenerateUserToken(ctx context.Context, request *api.GenerateUserTokenRequest) (*api.GenerateUserTokenResponse, error) {
	token, err := utils.GenerateToken(request.UserID, request.Email, request.DriverID, s.secret)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.MyCustomErrorCode.CustomMessage("生成token失败").Reason(utils.FormatErrorStack(err)))
	}

	return &api.GenerateUserTokenResponse{Token: token}, nil
}

func (s *UserServiceServer) Access(ctx context.Context, request *api.AccessRequest) (*api.AuthClaims, error) {
	resp := &api.AuthClaims{}

	_, claims, err := utils.ParseToken(request.Token, s.secret)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.InvalidParameter.CustomMessage("token解析失败").Reason(utils.FormatErrorStack(err)))
	}

	resp = &api.AuthClaims{
		UserID:    claims.UserId,
		Email:     claims.Email,
		DriverID:  claims.DriverId,
		PublicKey: claims.PublicKey,
		ExpireAt:  claims.ExpiresAt.UnixNano() / 1e6,
	}

	infos, err := s.userCache.GetUserLoginInfos(ctx, claims.UserId)
	if err == nil {
		var found bool
		for _, v := range infos {
			if v.Token == request.Token {
				found = true
				break
			}
		}

		if !found {
			return nil, code.WrapCodeToGRPC(code.Unauthorized)
		}

		return resp, nil
	}

	info, err := s.ur.GetUser(ctx, claims.UserId)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.UserErrNotExistOrPassword)
		}
		return nil, code.WrapCodeToGRPC(code.UserErrGetUserInfoFailed.Reason(utils.FormatErrorStack(err)))
	}

	if info.Status != entity.UserStatusNormal {
		return nil, code.WrapCodeToGRPC(code.UserErrStatusException)
	}

	return resp, nil
}
