package repository

import "github.com/cossim/coss-server/service/relation/domain/entity"

type GroupRelationRepository interface {
	CreateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error)
	UpdateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error)
	DeleteRelationByID(gid uint32, uid string) error
	GetUserGroupIDs(gid uint32) ([]string, error)
	GetUserGroupByID(gid uint32, uid string) (*entity.GroupRelation, error)
	GetJoinRequestListByID(gid uint32) ([]*entity.GroupRelation, error)
}
