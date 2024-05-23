package grpc

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/pkg/errors"
)

var _ groupgrpcv1.GroupServiceServer = &GroupServiceServer{}

// todo
func (s *GroupServiceServer) GetGroupInfoByGid(ctx context.Context, request *groupgrpcv1.GetGroupInfoRequest) (*groupgrpcv1.Group, error) {
	if request == nil || request.Gid == 0 {
		return nil, code.WrapCodeToGRPC(code.InvalidParameter)
	}

	resp := &groupgrpcv1.Group{}

	e, err := s.repo.Get(ctx, request.Gid)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.GroupErrGroupNotFound.Reason(utils.FormatErrorStack(err)))
		}
		return nil, code.WrapCodeToGRPC(code.GroupErrGetGroupInfoByGidFailed.Reason(utils.FormatErrorStack(err)))
	}

	resp = &groupgrpcv1.Group{
		Id:              e.ID,
		Type:            groupgrpcv1.GroupType(e.Type),
		Status:          groupgrpcv1.GroupStatus(e.Status),
		MaxMembersLimit: int32(e.MaxMembersLimit),
		CreatorId:       e.CreatorID,
		Name:            e.Name,
		Avatar:          e.Avatar,
		SilenceTime:     e.SilenceTime,
		JoinApprove:     e.JoinApprove,
		Encrypt:         e.Encrypt,
	}

	return resp, nil
}

// todo
func (s *GroupServiceServer) GetBatchGroupInfoByIDs(ctx context.Context, request *groupgrpcv1.GetBatchGroupInfoRequest) (*groupgrpcv1.GetBatchGroupInfoResponse, error) {
	if request == nil || len(request.GroupIds) == 0 {
		return nil, code.WrapCodeToGRPC(code.InvalidParameter)
	}

	resp := &groupgrpcv1.GetBatchGroupInfoResponse{}

	groupIds := make([]uint, len(request.GroupIds))
	for i, id := range request.GroupIds {
		groupIds[i] = uint(id)
	}

	groupIDsUint32 := make([]uint32, len(groupIds))
	for i, id := range groupIds {
		groupIDsUint32[i] = uint32(id)
	}

	es, err := s.repo.Find(ctx, repository.Query{ID: groupIDsUint32})
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.GroupErrGetBatchGroupInfoByIDsFailed.Reason(utils.FormatErrorStack(err)))
	}

	var groupAPIs []*groupgrpcv1.Group
	for _, e := range es {
		groupAPI := &groupgrpcv1.Group{
			Id:              e.ID,
			Type:            groupgrpcv1.GroupType(e.Type),
			Status:          groupgrpcv1.GroupStatus(e.Status),
			MaxMembersLimit: int32(e.MaxMembersLimit),
			CreatorId:       e.CreatorID,
			Name:            e.Name,
			Avatar:          e.Avatar,
			SilenceTime:     e.SilenceTime,
			JoinApprove:     e.JoinApprove,
			Encrypt:         e.Encrypt,
		}
		groupAPIs = append(groupAPIs, groupAPI)
	}

	resp.Groups = groupAPIs
	return resp, nil
}
