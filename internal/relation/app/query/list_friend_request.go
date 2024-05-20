package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"

	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ListFriendRequest struct {
	UserID   string
	PageNum  int
	PageSize int
}

type ListFriendRequestResponse struct {
	List  []*FriendRequest
	Total int64
}

type FriendRequest struct {
	ID            uint32
	CreateAt      int64
	ExpiredAt     int64
	RecipientID   string
	RecipientInfo *ShortUserInfo
	Remark        string
	SenderID      string
	SenderInfo    *ShortUserInfo
	Status        int
}

type ShortUserInfo struct {
	UserID   string
	CossID   string
	Avatar   string
	Nickname string
}

type ListFriendRequestHandler decorator.CommandHandler[*ListFriendRequest, *ListFriendRequestResponse]

func NewListFriendRequestHandler(
	logger *zap.Logger,
	urd service.UserRelationDomain,
	userService rpc.UserService,
	userFriendRequestService service.UserFriendRequestDomain,
) ListFriendRequestHandler {
	return &listFriendRequestHandler{
		logger:                   logger,
		urd:                      urd,
		userService:              userService,
		userFriendRequestService: userFriendRequestService,
	}
}

type listFriendRequestHandler struct {
	logger                   *zap.Logger
	urd                      service.UserRelationDomain
	userService              rpc.UserService
	userFriendRequestService service.UserFriendRequestDomain
}

func (h *listFriendRequestHandler) Handle(ctx context.Context, cmd *ListFriendRequest) (*ListFriendRequestResponse, error) {
	if cmd == nil || cmd.UserID == "" {
		return nil, code.InvalidParameter
	}

	if cmd.PageNum == 0 {
		cmd.PageNum = 1
	}
	if cmd.PageSize == 0 {
		cmd.PageSize = 10
	}

	requestList, err := h.userFriendRequestService.List(ctx, &service.ListFriendRequestOptions{
		UserID:   cmd.UserID,
		PageNum:  cmd.PageNum,
		PageSize: cmd.PageSize,
	})
	if err != nil {
		h.logger.Error("list friend request error", zap.Error(err))
		return nil, err
	}

	recipientInfoCh := make(chan *ShortUserInfo, len(requestList.List))
	senderInfoCh := make(chan *ShortUserInfo, len(requestList.List))

	for _, request := range requestList.List {
		go func(request *entity.UserFriendRequest) {
			recipientInfo, err := h.userService.GetUserInfo(ctx, request.RecipientID)
			if err != nil {
				h.logger.Error("get recipient info error", zap.Error(err))
				recipientInfoCh <- &ShortUserInfo{}
				return
			}
			recipientInfoCh <- &ShortUserInfo{
				UserID:   recipientInfo.ID,
				CossID:   recipientInfo.CossID,
				Avatar:   recipientInfo.Avatar,
				Nickname: recipientInfo.Nickname,
			}
		}(request)
	}

	for _, request := range requestList.List {
		go func(request *entity.UserFriendRequest) {
			senderInfo, err := h.userService.GetUserInfo(ctx, request.SenderID)
			if err != nil {
				h.logger.Error("get sender info error", zap.Error(err))
				senderInfoCh <- &ShortUserInfo{}
				return
			}
			senderInfoCh <- &ShortUserInfo{
				UserID:   senderInfo.ID,
				CossID:   senderInfo.CossID,
				Avatar:   senderInfo.Avatar,
				Nickname: senderInfo.Nickname,
			}
		}(request)
	}

	resp := &ListFriendRequestResponse{
		Total: requestList.Total,
		List:  make([]*FriendRequest, 0, len(requestList.List)),
	}
	for i := 0; i < len(requestList.List); i++ {
		recipientInfo := <-recipientInfoCh
		senderInfo := <-senderInfoCh

		request := requestList.List[i]
		resp.List = append(resp.List, &FriendRequest{
			ID:            request.ID,
			CreateAt:      request.CreatedAt,
			ExpiredAt:     request.ExpiredAt,
			RecipientID:   request.RecipientID,
			RecipientInfo: recipientInfo,
			Remark:        request.Remark,
			SenderID:      request.SenderID,
			SenderInfo:    senderInfo,
			Status:        int(request.Status),
		})
	}

	return resp, nil
}
