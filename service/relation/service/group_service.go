package service

import (
	"context"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
)

func (s *Service) InsertUserGroup(ctx context.Context, request *v1.UserGroupRequest) (*v1.UserGroupResponse, error) {
	resp := &v1.UserGroupResponse{}

	// 创建领域模型
	userGroup := &entity.UserGroup{
		UID:     request.GetUserId(),
		GroupID: uint(request.GetGroupId()),
	}
	// 调用持久层方法插入用户群关系
	_, err := s.grr.InsertUserGroup(userGroup)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Service) GetUserGroupIDs(ctx context.Context, id *v1.GroupID) (*v1.UserIDs, error) {
	resp := &v1.UserIDs{}

	// 调用持久层方法获取用户群关系列表
	userGroupIDs, err := s.grr.GetUserGroupIDs(uint(id.GetGroupId()))
	if err != nil {
		return resp, err
	}

	resp.UserIds = userGroupIDs
	return resp, nil
}

func (s *Service) GetUserGroupShowSessionGroupIDs(ctx context.Context, in *v1.UserID) (*v1.GroupIDs, error) {
	ids, err := s.grr.GetUserGroupShowSessionGroupIds(in.UserId)
	if err != nil {
		return nil, err
	}
	gids := make([]uint32, len(ids))
	if len(ids) > 0 {
		for i, id := range ids {
			gids[i] = uint32(id)
		}
	}
	return &v1.GroupIDs{GroupIds: gids}, nil
}

func (s *Service) SetUserGroupShowSession(ctx context.Context, in *v1.SetUserGroupShowSessionRequest) (*v1.Empty, error) {
	err := s.grr.SetUserGroupShowSession(in.UserId, uint(in.GroupId), entity.ShowSession(in.ShowSession))
	if err != nil {
		return nil, err
	}
	return &v1.Empty{}, nil
}
