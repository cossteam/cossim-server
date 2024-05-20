package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
)

type ListGroupMemberOptions struct {
}

type ListGroupRequestOptions struct {
}

type GroupRelationDomain interface {
	// SetGroupRemark 设置群组备注
	SetGroupRemark(ctx context.Context, id uint32, id2 string, remark string) error

	// AddGroupAdmin 添加群聊管理员
	AddGroupAdmin(ctx context.Context, groupID uint32, currentUserID string, targetUsers ...string) error

	// RemoveGroupMember 移除群聊成员
	RemoveGroupMember(ctx context.Context, groupID uint32, currentUserID string, TargetUser ...string) error

	// IsUserGroupAdmin 判断用户是否是群聊管理员
	IsUserGroupAdmin(ctx context.Context, groupID uint32, userID string) (bool, error)

	// SetGroupSilent 设置群聊静音
	SetGroupSilent(ctx context.Context, userID string, groupID uint32, silent bool) error

	// QuitGroup 退出群聊
	QuitGroup(ctx context.Context, groupID uint32, userID string) error

	// IsUserGroupOwner 判断用户是否是群主
	IsUserGroupOwner(ctx context.Context, groupID uint32, userID string) (bool, error)

	// ListUserGroupID 获取用户加入的群聊列表
	ListUserGroupID(ctx context.Context, userID string) ([]uint32, error)

	ListGroupAdminID(ctx context.Context, groupID uint32) ([]string, error)

	// IsUserIneGroup 判断用户是否在群聊中
	IsUserIneGroup(ctx context.Context, groupID uint32, userID string) (bool, error)

	// IsUserInActiveGroup 判断用户是否在群聊中并且群聊状态是否正常
	IsUserInActiveGroup(ctx context.Context, groupID uint32, userID string) error

	// ListGroupMember 获取群聊成员列表
	ListGroupMember(ctx context.Context, groupID uint32, opts *ListGroupMemberOptions) (*entity.GroupRelationList, error)

	// ListGroupRequest 获取群聊申请列表
	ListGroupRequest(ctx context.Context, userID string, opts *ListGroupRequestOptions) ([]*entity.GroupJoinRequest, error)
}

var _ GroupRelationDomain = &groupRelationDomain{}

type groupRelationDomain struct {
	repos *persistence.Repositories
	//groupRepo            repository.GroupRelationRepository
	//groupJoinRequestRepo repository.GroupRequestRepository
	groupService rpc.GroupService
}

func (d *groupRelationDomain) SetGroupRemark(ctx context.Context, groupID uint32, userID string, remark string) error {
	return d.repos.GroupRepo.SetUserGroupRemark(ctx, groupID, userID, remark)
}

func (d *groupRelationDomain) AddGroupAdmin(ctx context.Context, groupID uint32, currentUserID string, targetUsers ...string) error {
	if len(targetUsers) == 0 {
		return code.InvalidParameter
	}

	isOwner, err := d.IsUserGroupOwner(ctx, groupID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return code.Forbidden
	}

	uniqueUsers := make(map[string]struct{})
	for _, userID := range targetUsers {
		if userID == currentUserID {
			return code.RelationErrGroupAddAdmin.CustomMessage("不能将自己设置为管理员")
		}
		uniqueUsers[userID] = struct{}{}
	}
	if err = d.IsUserInActiveGroup(ctx, groupID, currentUserID); err != nil {
		return err
	}

	for userID := range uniqueUsers {
		inGroup, err := d.IsUserIneGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if !inGroup {
			return code.RelationGroupErrNotInGroup
		}

		isAdmin, err := d.IsUserGroupAdmin(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if isAdmin {
			return code.RelationErrGroupAddAdmin.CustomMessage("目标用户已经是管理员")
		}

		// 设置目标用户为管理员
		// TODO 群聊管理员人数限制
		if err = d.SetGroupAdmin(ctx, groupID, userID); err != nil {
			return err
		}
	}

	return nil
}

func (d *groupRelationDomain) SetGroupAdmin(ctx context.Context, groupID uint32, userID string) error {
	return d.repos.GroupRepo.UpdateIdentity(ctx, groupID, userID, entity.IdentityAdmin)
}

func (d *groupRelationDomain) RemoveGroupMember(ctx context.Context, groupID uint32, currentUserID string, targetUsers ...string) error {
	currentUserRel, err := d.Get(ctx, groupID, currentUserID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return code.RelationGroupErrNotInGroup
		}
		return err
	}

	if !isGroupAdmin(currentUserRel.Identity) {
		return code.Forbidden
	}

	for _, userID := range targetUsers {
		targetUserRel, err := d.Get(ctx, groupID, userID)
		if err != nil {
			if errors.Is(err, code.NotFound) {
				return code.RelationGroupErrNotInGroup
			}
			return err
		}

		// 如果目标用户是管理员，并且当前用户不是群主，则禁止操作
		if isGroupAdmin(targetUserRel.Identity) && currentUserRel.Identity != entity.IdentityOwner {
			return code.Forbidden
		}
	}

	return d.repos.TXRepositories(func(txr *persistence.Repositories) error {
		// 查询对话信息
		dialog, err := txr.DialogRepo.GetByGroupID(ctx, groupID)
		if err != nil {
			return err
		}

		// 删除对话关系
		if err := txr.DialogUserRepo.DeleteByDialogIDAndUserID(ctx, dialog.ID, targetUsers...); err != nil {
			return err
		}

		// 删除群聊关系
		if err := txr.GroupRepo.DeleteByGroupIDAndUserID(ctx, groupID, targetUsers...); err != nil {
			return err
		}

		return nil
	})
}

func isGroupAdmin(identity entity.GroupIdentity) bool {
	return identity == entity.IdentityOwner || identity == entity.IdentityAdmin
}

func (d *groupRelationDomain) IsUserGroupAdmin(ctx context.Context, groupID uint32, userID string) (bool, error) {
	rel, err := d.repos.GroupRepo.Get(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return false, code.RelationGroupErrNotInGroup
		}
		return false, err
	}

	return rel.Identity == entity.IdentityOwner || rel.Identity == entity.IdentityAdmin, nil
}

func (d *groupRelationDomain) SetGroupSilent(ctx context.Context, userID string, groupID uint32, silent bool) error {
	inGroup, err := d.IsUserIneGroup(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !inGroup {
		return code.RelationGroupErrNotInGroup
	}

	return d.repos.GroupRepo.UserGroupSilentNotification(ctx, groupID, userID, silent)
}

func (d *groupRelationDomain) QuitGroup(ctx context.Context, groupID uint32, userID string) error {
	isActive, err := d.groupService.IsActiveGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if !isActive {
		return code.GroupErrGroupStatusNotAvailable
	}

	owner, err := d.IsUserGroupOwner(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if owner {
		return code.RelationGroupErrGroupOwnerCantLeaveGroupFailed
	}

	dialog, err := d.repos.DialogRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return code.GroupErrGroupNotFound
		}
		return err
	}

	rel, err := d.repos.GroupRepo.GetUserGroupByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return code.RelationGroupErrNotInGroup
		}
		return err
	}

	return d.repos.TXRepositories(func(txr *persistence.Repositories) error {
		if err := txr.DialogUserRepo.DeleteByDialogIDAndUserID(ctx, dialog.ID, userID); err != nil {
			return err
		}
		if err := txr.GroupRepo.Delete(ctx, rel.ID); err != nil {
			return err
		}
		return nil
	})
}

func (d *groupRelationDomain) IsUserGroupOwner(ctx context.Context, groupID uint32, userID string) (bool, error) {
	rel, err := d.repos.GroupRepo.Get(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return false, code.RelationGroupErrNotInGroup
		}
		return false, err
	}

	return rel.Identity == entity.IdentityOwner, nil
}

func (d *groupRelationDomain) Get(ctx context.Context, groupID uint32, userID string) (*entity.GroupRelation, error) {
	return d.repos.GroupRepo.Get(ctx, groupID, userID)
}

func (d *groupRelationDomain) IsUserIneGroup(ctx context.Context, groupID uint32, userID string) (bool, error) {
	_, err := d.repos.GroupRepo.Get(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *groupRelationDomain) ListUserGroupID(ctx context.Context, userID string) ([]uint32, error) {
	return d.repos.GroupRepo.GetUserGroupIDs(ctx, userID)
}

func NewGroupRelationDomain(repos *persistence.Repositories, groupService rpc.GroupService) GroupRelationDomain {
	return &groupRelationDomain{repos: repos, groupService: groupService}
}

func (d *groupRelationDomain) ListGroupAdminID(ctx context.Context, groupID uint32) ([]string, error) {
	return d.repos.GroupRepo.ListGroupAdmin(ctx, groupID)
}

func (d *groupRelationDomain) ListGroupRequest(ctx context.Context, userID string, opts *ListGroupRequestOptions) ([]*entity.GroupJoinRequest, error) {
	r, err := d.repos.GroupJoinRequestRepo.Find(ctx, &repository.GroupJoinRequestQuery{
		UserID: []string{userID},
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (d *groupRelationDomain) ListGroupMember(ctx context.Context, groupID uint32, opts *ListGroupMemberOptions) (*entity.GroupRelationList, error) {
	list, err := d.repos.GroupRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &entity.GroupRelationList{List: list}, nil
}

func (d *groupRelationDomain) IsUserInActiveGroup(ctx context.Context, groupID uint32, userID string) error {
	isActive, err := d.groupService.IsActiveGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if !isActive {
		return code.GroupErrGroupStatusNotAvailable
	}

	_, err = d.repos.GroupRepo.Get(ctx, groupID, userID)
	if err != nil {
		return err
	}

	return nil
}
