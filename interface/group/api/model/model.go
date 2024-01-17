package model

import api "github.com/cossim/coss-server/service/group/api/v1"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type UpdateGroupRequest struct {
	Type            uint32 `json:"type"`
	Status          uint32 `json:"status"`
	MaxMembersLimit uint32 `json:"max_members_limit"`
	CreatorID       string `json:"creator_id"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	GroupId         uint32 `json:"group_id"`
}

type CreateGroupRequest struct {
	Type            uint32 `json:"type"`
	MaxMembersLimit uint32 `json:"max_members_limit"`
	Name            string `json:"name" binding:"required"`
	Avatar          string `json:"avatar"`
}

type CreateGroupResponse struct {
	Id              uint32 `json:"id"`
	Avatar          string `json:"avatar"`
	Name            string `json:"name"`
	Type            uint32 `json:"type"`
	Status          int32  `json:"status"`
	MaxMembersLimit int32  `json:"max_members_limit"`
	CreatorId       string `json:"creator_id"`
	DialogId        uint32 `json:"dialog_id"`
}

func IsValidGroupType(value api.GroupType) bool {
	return value == api.GroupType_TypePublic || value == api.GroupType_TypePrivate
}
