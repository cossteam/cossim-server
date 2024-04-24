package rpc

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var _ groupgrpcv1.GroupServiceServer = &GroupServiceServer{}

func (s *GroupServiceServer) GetGroupInfoByGid(ctx context.Context, request *groupgrpcv1.GetGroupInfoRequest) (*groupgrpcv1.Group, error) {
	resp := &groupgrpcv1.Group{}

	entity, err := s.repo.Get(ctx, request.Gid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.GroupErrGroupNotFound.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.GroupErrGetGroupInfoByGidFailed.Code()), err.Error())
	}

	resp = &groupgrpcv1.Group{
		Id:              entity.ID,
		Type:            groupgrpcv1.GroupType(entity.Type),
		Status:          groupgrpcv1.GroupStatus(entity.Status),
		MaxMembersLimit: int32(entity.MaxMembersLimit),
		CreatorId:       entity.CreatorID,
		Name:            entity.Name,
		Avatar:          entity.Avatar,
		SilenceTime:     entity.SilenceTime,
		JoinApprove:     entity.JoinApprove,
		Encrypt:         entity.Encrypt,
	}

	return resp, nil
}

func (s *GroupServiceServer) GetBatchGroupInfoByIDs(ctx context.Context, request *groupgrpcv1.GetBatchGroupInfoRequest) (*groupgrpcv1.GetBatchGroupInfoResponse, error) {
	resp := &groupgrpcv1.GetBatchGroupInfoResponse{}

	if len(request.GroupIds) == 0 {
		return resp, code.MyCustomErrorCode.CustomMessage("entity ids is empty")
	}

	//将uint32转成uint
	groupIds := make([]uint, len(request.GroupIds))
	for i, id := range request.GroupIds {
		groupIds[i] = uint(id)
	}

	groupIDsUint32 := make([]uint32, len(groupIds))
	for i, id := range groupIds {
		groupIDsUint32[i] = uint32(id)
	}

	entitys, err := s.repo.Find(ctx, group.Query{ID: groupIDsUint32})
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	var groupAPIs []*groupgrpcv1.Group
	for _, entity := range entitys {
		groupAPI := &groupgrpcv1.Group{
			Id:              entity.ID,
			Type:            groupgrpcv1.GroupType(entity.Type),
			Status:          groupgrpcv1.GroupStatus(entity.Status),
			MaxMembersLimit: int32(entity.MaxMembersLimit),
			CreatorId:       entity.CreatorID,
			Name:            entity.Name,
			Avatar:          entity.Avatar,
			SilenceTime:     entity.SilenceTime,
			JoinApprove:     entity.JoinApprove,
			Encrypt:         entity.Encrypt,
		}
		groupAPIs = append(groupAPIs, groupAPI)
	}

	resp.Groups = groupAPIs
	return resp, nil
}

func (s *GroupServiceServer) UpdateGroup(ctx context.Context, request *groupgrpcv1.UpdateGroupRequest) (*groupgrpcv1.Group, error) {
	resp := &groupgrpcv1.Group{}

	entity := &group.Group{
		ID:              request.Group.Id,
		Type:            group.Type(request.Group.Type),
		Status:          group.Status(request.Group.Status),
		MaxMembersLimit: int(request.Group.MaxMembersLimit),
		CreatorID:       request.Group.CreatorId,
		Name:            request.Group.Name,
		Avatar:          request.Group.Avatar,
		SilenceTime:     request.Group.SilenceTime,
		JoinApprove:     request.Group.JoinApprove,
		Encrypt:         request.Group.Encrypt,
	}

	if err := s.repo.Update(ctx, entity, func(h *group.Group) (*group.Group, error) {
		resp = &groupgrpcv1.Group{
			Id:              h.ID,
			Type:            groupgrpcv1.GroupType(h.Type),
			Status:          groupgrpcv1.GroupStatus(h.Status),
			MaxMembersLimit: int32(h.MaxMembersLimit),
			CreatorId:       h.CreatorID,
			Name:            h.Name,
			Avatar:          h.Avatar,
			SilenceTime:     h.SilenceTime,
			JoinApprove:     h.JoinApprove,
			Encrypt:         h.Encrypt,
		}
		return nil, nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *GroupServiceServer) CreateGroup(ctx context.Context, request *groupgrpcv1.CreateGroupRequest) (*groupgrpcv1.Group, error) {
	resp := &groupgrpcv1.Group{}

	entity := &group.Group{
		Type:            group.Type(request.Group.Type),
		Status:          group.Status(request.Group.Status),
		MaxMembersLimit: int(request.Group.MaxMembersLimit),
		CreatorID:       request.Group.CreatorId,
		Name:            request.Group.Name,
		Avatar:          request.Group.Avatar,
		SilenceTime:     request.Group.SilenceTime,
		JoinApprove:     request.Group.JoinApprove,
		Encrypt:         request.Group.Encrypt,
	}

	if err := s.repo.Create(ctx, entity, func(h *group.Group) (*group.Group, error) {
		resp = &groupgrpcv1.Group{
			Id:              h.ID,
			Type:            groupgrpcv1.GroupType(h.Type),
			Status:          groupgrpcv1.GroupStatus(h.Status),
			MaxMembersLimit: int32(h.MaxMembersLimit),
			CreatorId:       h.CreatorID,
			Name:            h.Name,
			Avatar:          h.Avatar,
			SilenceTime:     h.SilenceTime,
			JoinApprove:     h.JoinApprove,
			Encrypt:         h.Encrypt,
		}
		return nil, nil
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *GroupServiceServer) CreateGroupRevert(ctx context.Context, request *groupgrpcv1.CreateGroupRequest) (*groupgrpcv1.Group, error) {
	resp := &groupgrpcv1.Group{}
	if err := s.repo.Delete(ctx, request.Group.Id); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *GroupServiceServer) DeleteGroup(ctx context.Context, request *groupgrpcv1.DeleteGroupRequest) (*groupgrpcv1.EmptyResponse, error) {
	resp := &groupgrpcv1.EmptyResponse{}

	if err := s.repo.Delete(ctx, request.GetGid()); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *GroupServiceServer) DeleteGroupRevert(ctx context.Context, request *groupgrpcv1.DeleteGroupRequest) (*groupgrpcv1.EmptyResponse, error) {
	resp := &groupgrpcv1.EmptyResponse{}
	if err := s.repo.UpdateFields(ctx, request.Gid, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}
