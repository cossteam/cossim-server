package repository

import "github.com/cossim/coss-server/service/relation/domain/entity"

type UserRelationRepository interface {
	CreateRelation(ur *entity.UserRelation) (*entity.UserRelation, error)
	UpdateRelation(ur *entity.UserRelation) (*entity.UserRelation, error)
	DeleteRelationByID(userId, friendId string) error
	GetRelationByID(userId, friendId string) (*entity.UserRelation, error)
	GetRelationsByUserID(userId string) ([]*entity.UserRelation, error)
	GetBlacklistByUserID(userId string) ([]*entity.UserRelation, error)
	GetFriendRequestListByUserID(userId string) ([]*entity.UserRelation, error)
}
