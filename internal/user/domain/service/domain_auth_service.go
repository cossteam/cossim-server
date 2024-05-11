package service

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
)

type AuthDomain interface {
	GenerateUserToken(ctx context.Context, ac *entity.AuthClaims) (string, error)
	ParseToken(ctx context.Context, token string) (*entity.AuthClaims, error)
	Access(ctx context.Context, token string) error
}

var _ AuthDomain = &authDomain{}

type authDomain struct {
	secret    string
	ur        repository.UserRepository
	userCache cache.UserCache
}

func NewAuthDomain(secret string, ur repository.UserRepository, userCache cache.UserCache) AuthDomain {
	return &authDomain{secret: secret, ur: ur, userCache: userCache}
}

func (d *authDomain) GenerateUserToken(ctx context.Context, ac *entity.AuthClaims) (string, error) {
	token, err := utils.GenerateToken(ac.UserID, ac.Email, ac.DriverID, ac.PublicKey, d.secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *authDomain) ParseToken(ctx context.Context, token string) (*entity.AuthClaims, error) {
	_, claims, err := utils.ParseToken(token, d.secret)
	if err != nil {
		return nil, err
	}

	return &entity.AuthClaims{
		UserID:    claims.UserId,
		Email:     claims.Email,
		DriverID:  claims.DriverId,
		PublicKey: claims.PublicKey,
	}, nil
}

func (d *authDomain) Access(ctx context.Context, token string) error {
	parseToken, err := d.ParseToken(ctx, token)
	if err != nil {
		return err
	}

	infos, err := d.userCache.GetUserLoginInfos(ctx, parseToken.UserID)
	if err == nil {
		var found bool
		for _, v := range infos {
			if v.Token == token {
				found = true
				break
			}
		}
		if !found {
			return code.Unauthorized
		}
		return nil
	}

	info, err := d.ur.GetUserInfoByUid(ctx, parseToken.UserID)
	if err == nil && info.Status != entity.UserStatusNormal {
		return nil
	}

	return code.Unauthorized
}
