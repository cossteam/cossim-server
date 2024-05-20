package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
	"sync"
)

type ListGroupRequest struct {
	UserID   string
	GroupID  uint32
	PageNum  int
	PageSize int
}

func (cmd *ListGroupRequest) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type ListGroupRequestResponse struct {
	List []*GroupRequest
	//Total int64
	//Page  int32
}

type GroupRequest struct {
	ID            uint32
	GroupId       uint32
	GroupType     uint32
	CreatorId     string
	GroupName     string
	GroupAvatar   string
	SenderInfo    *RequestUserInfo
	RecipientInfo *RequestUserInfo
	Status        uint8
	Remark        string
	CreateAt      int64
	ExpiredAt     int64
}

type RequestUserInfo struct {
	ID       string
	Nickname string
	Avatar   string
}

type ListGroupRequestHandler decorator.CommandHandler[*ListGroupRequest, *ListGroupRequestResponse]

func NewListGroupRequestHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	userService rpc.UserService,
	groupService rpc.GroupService,
) ListGroupRequestHandler {
	return &listGroupRequestHandler{
		logger:       logger,
		grd:          grd,
		userService:  userService,
		groupService: groupService,
	}
}

type listGroupRequestHandler struct {
	logger *zap.Logger

	grd service.GroupRelationDomain

	userService  rpc.UserService
	groupService rpc.GroupService
}

func (h *listGroupRequestHandler) Handle(ctx context.Context, cmd *ListGroupRequest) (*ListGroupRequestResponse, error) {
	//if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
	//	h.logger.Error("IsUserInActiveGroup", zap.Error(err))
	//	return nil, err
	//}

	list, err := h.grd.ListGroupRequest(ctx, cmd.UserID, &service.ListGroupRequestOptions{})
	if err != nil {
		h.logger.Error("ListGroupRequest", zap.Error(err))
		return nil, err
	}

	resp := &ListGroupRequestResponse{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(list))
	defer close(errCh)

	for _, v := range list {
		wg.Add(1)
		go func(v *entity.GroupJoinRequest) {
			defer wg.Done()
			groupInfo, err := h.groupService.GetGroupInfo(ctx, v.GroupID)
			if err != nil {
				errCh <- err
				return
			}

			recipientInfo, err := h.getUserInfo(ctx, v.UserID)
			if err != nil {
				errCh <- err
				return
			}

			senderInfo, err := h.getSenderInfo(ctx, v)
			if err != nil {
				errCh <- err
				return
			}

			groupRequest := &GroupRequest{
				ID:          v.ID,
				GroupId:     v.GroupID,
				GroupType:   uint32(groupInfo.Type),
				CreatorId:   groupInfo.CreatorId,
				GroupName:   groupInfo.Name,
				GroupAvatar: groupInfo.Avatar,
				SenderInfo:  senderInfo,
				RecipientInfo: &RequestUserInfo{
					ID:       recipientInfo.ID,
					Nickname: recipientInfo.Nickname,
					Avatar:   recipientInfo.Avatar,
				},
				Status:    uint8(v.Status),
				Remark:    v.Remark,
				CreateAt:  v.CreatedAt,
				ExpiredAt: v.ExpiredAt,
			}

			mu.Lock()
			resp.List = append(resp.List, groupRequest)
			mu.Unlock()
		}(v)
	}

	wg.Wait()

	if len(errCh) > 0 {
		err := <-errCh
		return nil, err
	}

	return resp, nil
}

func (h *listGroupRequestHandler) getUserInfo(ctx context.Context, userID string) (*rpc.User, error) {
	userInfo, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		h.logger.Error("GetUserInfo", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	return userInfo, nil
}

func (h *listGroupRequestHandler) getSenderInfo(ctx context.Context, v *entity.GroupJoinRequest) (*RequestUserInfo, error) {
	userID := v.Inviter
	if userID == "" {
		userID = v.UserID
	}
	userInfo, err := h.getUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &RequestUserInfo{
		ID:       userInfo.ID,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
	}, nil
}
