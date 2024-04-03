package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type UserRelationRepository interface {
	CreateRelation(ur *entity.UserRelation) (*entity.UserRelation, error)
	UpdateRelation(ur *entity.UserRelation) (*entity.UserRelation, error)
	DeleteRelationByID(userId, friendId string) error
	UpdateRelationDeleteAtByID(userId, friendId string, deleteAt int64) error
	GetRelationByID(userId, friendId string) (*entity.UserRelation, error)
	GetRelationByIDs(userId string, friendIds []string) ([]*entity.UserRelation, error)
	GetRelationsByUserID(userId string) ([]*entity.UserRelation, error)
	GetBlacklistByUserID(userId string) ([]*entity.UserRelation, error)
	GetFriendRequestListByUserID(userId string) ([]*entity.UserRelation, error)
	UpdateRelationColumn(id uint, column string, value interface{}) error
	SetUserFriendSilentNotification(uid, friendId string, silentNotification entity.SilentNotification) error
	SetUserOpenBurnAfterReading(uid, friendId string, openBurnAfterReading entity.OpenBurnAfterReadingType, burnAfterReadingTimeOut int64) error
	SetFriendRemarkByUserIdAndFriendId(userId, friendId string, remark string) error
}
