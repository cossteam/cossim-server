package converter

import (
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/infra/persistence/po"
)

func UserEntityToPO(e *entity.User) *po.User {
	return &po.User{
		ID:           e.ID,
		CossID:       e.CossID,
		Email:        e.Email,
		Tel:          e.Tel,
		NickName:     e.NickName,
		Avatar:       e.Avatar,
		PublicKey:    e.PublicKey,
		Password:     e.Password,
		LastIp:       e.LastIp,
		LineIp:       e.LineIp,
		CreatedIp:    e.CreatedIp,
		Signature:    e.Signature,
		LineAt:       e.LineAt,
		LastAt:       e.LastAt,
		Status:       uint(e.Status),
		EmailVerity:  e.EmailVerity,
		Bot:          e.Bot,
		SecretBundle: e.SecretBundle,
		CreatedAt:    e.CreatedAt,
		//UpdatedAt:    e.UpdatedAt,
		//DeletedAt:    e.DeletedAt,
	}
}

func UserPOToEntity(po *po.User) *entity.User {
	return &entity.User{
		ID:           po.ID,
		CossID:       po.CossID,
		Email:        po.Email,
		Tel:          po.Tel,
		NickName:     po.NickName,
		Avatar:       po.Avatar,
		PublicKey:    po.PublicKey,
		Password:     po.Password,
		LastIp:       po.LastIp,
		LineIp:       po.LineIp,
		CreatedIp:    po.CreatedIp,
		Signature:    po.Signature,
		LineAt:       po.LineAt,
		LastAt:       po.LastAt,
		Status:       entity.UserStatus(po.Status),
		EmailVerity:  po.EmailVerity,
		Bot:          po.Bot,
		SecretBundle: po.SecretBundle,
		CreatedAt:    po.CreatedAt,
	}
}
