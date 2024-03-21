package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type GroupJoinRequestRepository interface {
	// 添加入群申请
	AddJoinRequest(entity *entity.GroupJoinRequest) (*entity.GroupJoinRequest, error)
	//获取用户入群申请列表
	GetJoinRequestListByID(userId string) ([]*entity.GroupJoinRequest, error)
	//批量添加入群申请
	AddJoinRequestBatch(entity []*entity.GroupJoinRequest) ([]*entity.GroupJoinRequest, error)
	//获取用户入群申请列表
	GetGroupJoinRequestListByUserId(userID string) ([]*entity.GroupJoinRequest, error)
	//根据群ID和用户ID获取用户入群申请列表
	GetGroupJoinRequestByGroupIdAndUserId(groupID uint, userID string) (*entity.GroupJoinRequest, error)
	// 管理用户入群申请状态
	ManageGroupJoinRequestByID(id uint, status entity.RequestStatus) error
	//根据ID查询入群申请
	GetGroupJoinRequestByRequestID(id uint) (*entity.GroupJoinRequest, error)
	//根据多个groupId获取入群申请
	GetJoinRequestBatchListByGroupIDs(gids []uint) ([]*entity.GroupJoinRequest, error)
	//根据多个请求ID获取入群申请
	GetJoinRequestListByRequestIDs(ids []uint) ([]*entity.GroupJoinRequest, error)

	DeleteJoinRequestByID(id uint) error
}
