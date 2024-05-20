package query

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ListGroupAnnouncement struct {
	UserID   string
	GroupID  uint32
	PageNum  int
	PageSize int
}

func (cmd *ListGroupAnnouncement) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type ListGroupAnnouncementResponse struct {
	List []*GroupAnnouncement
	//Total int64
	//Page  int32
}

type GroupAnnouncement struct {
	ID           uint32
	GroupID      uint32
	Title        string
	Content      string
	CreateAt     int64
	UpdateAt     int64
	OperatorInfo *ShortUserInfo
	ReadUserList []*GroupAnnouncementReadUser
}

type GroupAnnouncementReadUser struct {
	ID             uint32
	GroupID        uint32
	AnnouncementID uint32
	ReadAt         int64
	UserID         string
	ReaderInfo     *ShortUserInfo
}

type ListGroupAnnouncementHandler decorator.CommandHandler[*ListGroupAnnouncement, *ListGroupAnnouncementResponse]

func NewListGroupAnnouncementHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
	userService rpc.UserService,
) ListGroupAnnouncementHandler {
	return &listGroupAnnouncementHandler{
		logger:                  logger,
		grd:                     grd,
		groupAnnouncementDomain: groupAnnouncementDomain,
		userService:             userService,
	}
}

type listGroupAnnouncementHandler struct {
	logger *zap.Logger

	grd                     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain

	userService rpc.UserService
}

func (h *listGroupAnnouncementHandler) Handle(ctx context.Context, cmd *ListGroupAnnouncement) (*ListGroupAnnouncementResponse, error) {
	if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return nil, err
	}

	announcements, err := h.groupAnnouncementDomain.ListAnnouncement(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}

	resp := &ListGroupAnnouncementResponse{
		List: make([]*GroupAnnouncement, 0, len(announcements.List)),
	}

	operators := make([]string, 0, len(announcements.List))
	for _, a := range announcements.List {
		operators = append(operators, a.UserID)
	}

	operatorInfos, err := h.userService.GetUsersInfo(ctx, operators)
	if err != nil {
		return nil, err
	}

	for _, announcement := range announcements.List {
		operatorInfo, ok := operatorInfos[announcement.UserID]
		if !ok {
			return nil, fmt.Errorf("operator info not found for user ID: %s", announcement.UserID)
		}

		readers, err := h.groupAnnouncementDomain.ListAnnouncementRead(ctx, announcement.GroupID, announcement.ID)
		if err != nil {
			return nil, err
		}

		readUserList, err := buildReadUserList(ctx, h.userService, readers)
		if err != nil {
			return nil, err
		}

		resp.List = append(resp.List, &GroupAnnouncement{
			ID:       announcement.ID,
			GroupID:  announcement.GroupID,
			Title:    announcement.Title,
			Content:  announcement.Content,
			CreateAt: announcement.CreatedAt,
			UpdateAt: announcement.UpdatedAt,
			OperatorInfo: &ShortUserInfo{
				UserID:   operatorInfo.ID,
				CossID:   operatorInfo.CossID,
				Nickname: operatorInfo.Nickname,
				Avatar:   operatorInfo.Avatar,
			},
			ReadUserList: readUserList,
		})
	}

	return resp, nil
}

func buildReadUserList(ctx context.Context, userService rpc.UserService, readers []*entity.GroupAnnouncementRead) ([]*GroupAnnouncementReadUser, error) {
	readUserList := make([]*GroupAnnouncementReadUser, 0, len(readers))

	ids := make([]string, 0, len(readers))
	for _, r := range readers {
		ids = append(ids, r.UserId)
	}

	readerInfos, err := userService.GetUsersInfo(ctx, ids)
	if err != nil {
		return nil, err
	}

	for _, reader := range readers {
		readerInfo, ok := readerInfos[reader.UserId]
		if !ok {
			return nil, fmt.Errorf("reader info not found for user ID: %s", reader.UserId)
		}
		readUserList = append(readUserList, &GroupAnnouncementReadUser{
			ID:             reader.ID,
			GroupID:        reader.GroupID,
			AnnouncementID: reader.AnnouncementId,
			ReadAt:         reader.ReadAt,
			UserID:         reader.UserId,
			ReaderInfo: &ShortUserInfo{
				UserID:   readerInfo.ID,
				CossID:   readerInfo.CossID,
				Avatar:   readerInfo.Avatar,
				Nickname: readerInfo.Nickname,
			},
		})
	}

	return readUserList, nil
}
