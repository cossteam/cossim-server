package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"

	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
)

// GroupRequestDomain 定义了群组请求相关的业务逻辑接口
type GroupRequestDomain interface {
	ManageRequest(ctx context.Context, userID string, requestID uint32, action entity.RequestAction) error

	// Get 获取指定 ID 的群组加入请求
	Get(ctx context.Context, id uint32) (*entity.GroupJoinRequest, error)

	// Delete 删除指定 ID 的群组加入请求
	Delete(ctx context.Context, id uint32, userID string) error

	// AddGroupRequest 添加一个群组请求
	AddGroupRequest(ctx context.Context, groupID uint32, userID, remark string) error

	// InviteJoinGroup 邀请用户加入群组
	InviteJoinGroup(ctx context.Context, groupID uint32, inviter string, targetUser []string) error

	// VerifyGroupInviteCondition 验证用户是否符合邀请入群的条件
	VerifyGroupInviteCondition(ctx context.Context, groupID uint32, inviter string, targetUser []string) error

	// IsUserInGroup 检查指定用户是否在群组中
	IsUserInGroup(ctx context.Context, groupID uint32, userID string) (bool, error)
}

var _ GroupRequestDomain = &groupRequestDomain{}

func NewGroupRequestDomain(repo repository.GroupRequestRepository, groupRepo repository.GroupRelationRepository, groupService rpc.GroupService) GroupRequestDomain {
	return &groupRequestDomain{repo: repo, groupRepo: groupRepo, groupService: groupService}
}

type groupRequestDomain struct {
	repo         repository.GroupRequestRepository
	groupRepo    repository.GroupRelationRepository
	groupService rpc.GroupService
	repos        persistence.Repositories

	groupRelationDomain GroupRelationDomain
}

func (d *groupRequestDomain) ManageRequest(ctx context.Context, userID string, requestID uint32, action entity.RequestAction) error {
	r, err := d.Get(ctx, requestID)
	if err != nil {
		return err
	}
	if r.OwnerID != userID {
		return code.Forbidden
	}
	switch r.Status {
	case entity.Expired:
		return code.Expired
	case entity.Accepted, entity.Rejected:
		return code.DuplicateOperation
	default:
	}

	switch action {
	case entity.Accept:
		info, err := d.groupService.GetGroupInfo(ctx, r.GroupID)
		if err != nil {
			return err
		}
		// 检查群组成员是否已达上限
		ds, err := d.groupRepo.GetGroupUserIDs(ctx, r.GroupID)
		if err != nil {
			return err
		}
		if len(ds) >= int(info.MaxMembersLimit) {
			return code.RelationGroupErrGroupFull
		}
		err = d.repo.UpdateStatus(ctx, requestID, entity.Accepted)
	case entity.Reject:
		err = d.repo.UpdateStatus(ctx, requestID, entity.Rejected)
	case entity.Ignore:
		// 忽略请求
		fallthrough
	default:
		return code.InvalidParameter
	}

	if err != nil {
		return err
	}

	return nil
}

// AddGroupRequest 添加一个群组请求
func (d *groupRequestDomain) AddGroupRequest(ctx context.Context, groupID uint32, userID, remark string) error {
	// 检查群组是否为私密群组，私密群只能管理员邀请入群
	isPrivate, err := d.groupService.IsPrivateGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if isPrivate {
		return code.Forbidden
	}

	// 获取群组信息
	groupInfo, err := d.groupService.GetGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}

	// 获取对话信息
	dialog, err := d.repos.DialogRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return err
	}

	// 如果群组不需要审批，直接加入群聊
	if !groupInfo.JoinApprove {
		err := d.repos.TXRepositories(func(txr *persistence.Repositories) error {
			// 创建对话用户
			if _, err := txr.DialogUserRepo.Create(ctx, &repository.CreateDialogUser{DialogID: dialog.ID, UserID: userID}); err != nil {
				return err
			}
			// 创建群组关系
			gr := &entity.CreateGroupRelation{
				GroupID:     groupID,
				UserID:      userID,
				Identity:    entity.IdentityUser,
				JoinedAt:    ptime.Now(),
				EntryMethod: entity.EntrySearch,
			}
			if _, err := txr.GroupRepo.Create(ctx, gr); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// 检查是否已经存在邀请
	gr, err := d.repo.GetByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if gr != nil && gr.ID != 0 && (gr.Status == entity.Pending || gr.Status == entity.Invitation) {
		return code.RelationErrGroupRequestAlreadyProcessed
	}

	relations := make([]*entity.GroupJoinRequest, 0)
	// 添加管理员群聊申请记录
	ids, err := d.groupRelationDomain.ListGroupAdminID(ctx, groupID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		userGroup := &entity.GroupJoinRequest{
			UserID:  id,
			GroupID: groupID,
			Remark:  remark,
			OwnerID: id,
			Status:  entity.Pending,
		}
		relations = append(relations, userGroup)
	}

	// 添加用户群聊申请记录
	ur := &entity.GroupJoinRequest{
		GroupID: groupID,
		UserID:  userID,
		Remark:  remark,
		OwnerID: userID,
		Status:  entity.Pending,
	}
	relations = append(relations, ur)

	_, err = d.repo.Creates(ctx, relations)
	if err != nil {
		return err
	}

	return nil
}

// InviteJoinGroup 邀请用户加入群组
func (d *groupRequestDomain) InviteJoinGroup(ctx context.Context, groupID uint32, inviter string, targetUser []string) error {
	// 获取群组管理员 ID 列表
	adminIDs, err := d.groupRepo.ListGroupAdmin(ctx, groupID)
	if err != nil {
		return err
	}

	// 创建通知列表，包含邀请人和管理员
	notify := utils.RemoveDuplicate(append([]string{inviter}, adminIDs...))

	relations := []*entity.GroupJoinRequest{}
	// 为每个目标用户创建群组加入请求
	for _, userID := range targetUser {
		userGroup := &entity.GroupJoinRequest{
			UserID:    userID,
			GroupID:   groupID,
			Inviter:   inviter,
			OwnerID:   userID,
			InviterAt: ptime.Now(),
			Status:    entity.Invitation,
		}
		relations = append(relations, userGroup)
	}

	// 为每个管理员添加相同的请求
	for _, adminID := range notify {
		for _, userID := range targetUser {
			adminGroup := &entity.GroupJoinRequest{
				UserID:    userID,
				GroupID:   groupID,
				Inviter:   inviter,
				OwnerID:   adminID,
				InviterAt: ptime.Now(),
				Status:    entity.Invitation,
			}
			relations = append(relations, adminGroup)
		}
	}

	_, err = d.repo.Creates(ctx, relations)
	if err != nil {
		return err
	}

	return nil
}

// IsUserInGroup 检查用户是否在群组中
func (d *groupRequestDomain) IsUserInGroup(ctx context.Context, groupID uint32, userID string) (bool, error) {
	relation, err := d.groupRepo.Get(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return false, nil
		}
		return false, err
	}
	return relation != nil, nil
}

// VerifyGroupInviteCondition 验证用户是否可以邀请入群
func (d *groupRequestDomain) VerifyGroupInviteCondition(ctx context.Context, groupID uint32, inviter string, targetUser []string) error {
	if len(targetUser) == 0 {
		return code.InvalidParameter.CustomMessage("invite members is empty")
	}

	// 检查群组是否激活
	isActive, err := d.groupService.IsActiveGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if !isActive {
		return code.GroupErrGroupStatusNotAvailable
	}

	// 获取群组信息
	groupInfo, err := d.groupService.GetGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}

	// 私密群组只有管理员可以邀请
	if groupInfo.Type == rpc.PrivateGroup {
		relation, err := d.groupRepo.Get(ctx, groupID, inviter)
		if err != nil {
			return err
		}
		if relation.Identity == entity.IdentityUser {
			return code.Forbidden
		}
	}

	// 检查群组成员是否已达上限
	ds, err := d.groupRepo.GetGroupUserIDs(ctx, groupID)
	if err != nil {
		return err
	}
	if len(ds) >= int(groupInfo.MaxMembersLimit) {
		return code.RelationGroupErrGroupFull
	}

	for _, userID := range targetUser {
		// 检查是否已有邀请
		gr, err := d.repo.GetByGroupIDAndUserID(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if gr != nil && gr.ID != 0 && (gr.Status == entity.Pending || gr.Status == entity.Invitation) {
			return code.RelationErrGroupRequestAlreadyProcessed
		}

		// 检查用户是否已在群组中
		inGroup, err := d.IsUserInGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if inGroup {
			return code.RelationGroupErrAlreadyInGroup
		}
	}

	return nil
}

// Get 获取群组加入请求
func (d *groupRequestDomain) Get(ctx context.Context, id uint32) (*entity.GroupJoinRequest, error) {
	return d.repo.Get(ctx, id)
}

// Delete 删除群组加入请求
func (d *groupRequestDomain) Delete(ctx context.Context, id uint32, userID string) error {
	r, err := d.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if r.OwnerID != userID {
		return code.Forbidden
	}
	return d.repo.Delete(ctx, id)
}
