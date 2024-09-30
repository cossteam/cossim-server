package query

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/mozillazg/go-pinyin"
	"go.uber.org/zap"
	"strings"
	"unicode"
)

type ListFriend struct {
	UserID   string
	PageNum  int
	PageSize int
}

type ListFriendResponse struct {
	List  []*UserInfo
	Total int64
	Page  int32
}

type UserInfo struct {
	UserID         string
	NickName       string
	Email          string
	Tel            string
	Avatar         string
	Signature      string
	CossId         string
	Letter         string
	Status         int
	DialogId       uint32
	RelationStatus int
	Preferences    *Preferences
}

type Preferences struct {
	Remark                      string
	SilentNotification          bool
	OpenBurnAfterReading        bool
	OpenBurnAfterReadingTimeOut int
}

type ListFriendHandler decorator.CommandHandler[*ListFriend, *ListFriendResponse]

func NewListFriendHandler(
	logger *zap.Logger,
	urd service.UserRelationDomain,
	userService rpc.UserService,
) ListFriendHandler {
	return &listFriendHandler{
		logger:      logger,
		urd:         urd,
		userService: userService,
	}
}

type listFriendHandler struct {
	logger      *zap.Logger
	urd         service.UserRelationDomain
	userService rpc.UserService
}

func (h *listFriendHandler) Handle(ctx context.Context, cmd *ListFriend) (*ListFriendResponse, error) {
	if cmd == nil {
		return nil, code.InvalidParameter
	}

	friendsList, err := h.urd.FriendsList(ctx, cmd.UserID, &service.FriendsListOptions{})
	if err != nil {
		h.logger.Error("list friend failed", zap.Error(err))
		return nil, err
	}

	userIDs := make([]string, len(friendsList))
	for i, v := range friendsList {
		userIDs[i] = v.UserID
	}

	usersInfo, err := h.userService.GetUsersInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	resp := &ListFriendResponse{}

	for _, friend := range friendsList {
		userInfo, ok := usersInfo[friend.UserID]
		if !ok {
			h.logger.Error("failed to get user info", zap.String("userID", friend.UserID))
			continue
		}

		fmt.Println("friend ", friend.UserID, friend.IsSilent)

		var rs int
		switch friend.Status {
		case entity.UserStatusNormal:
			rs = 1
		case entity.UserStatusBlocked:
			rs = 2
		case entity.UserStatusDeleted:
			rs = 3
		}

		letter := friend.Remark
		if letter == "" {
			letter = userInfo.Nickname
		}

		resp.List = append(resp.List, &UserInfo{
			UserID:         userInfo.ID,
			NickName:       userInfo.Nickname,
			Email:          userInfo.Email,
			Tel:            userInfo.Tel,
			Avatar:         userInfo.Avatar,
			Signature:      userInfo.Signature,
			Letter:         getGroupKey(letter),
			CossId:         userInfo.CossID,
			Status:         int(userInfo.Status),
			DialogId:       friend.DialogID,
			RelationStatus: rs,
			Preferences: &Preferences{
				Remark:                      friend.Remark,
				SilentNotification:          friend.IsSilent,
				OpenBurnAfterReading:        friend.OpenBurnAfterReading,
				OpenBurnAfterReadingTimeOut: int(friend.OpenBurnAfterReadingTimeOut),
			},
		})
	}

	return resp, nil
}

func getGroupKey(name string) string {
	if isChinese(name) {
		pinyinSlice := pinyin.Pinyin(name, pinyin.NewArgs())
		firstChar := getFirstChar(pinyinSlice)
		return strings.ToUpper(firstChar)
	} else if isSpecialChar(name) {
		return "#"
	}

	return strings.ToUpper(name[:1])
}

func isSpecialChar(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func isChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func getFirstChar(pinyinSlice [][]string) string {
	if len(pinyinSlice) > 0 && len(pinyinSlice[0]) > 0 {
		return pinyinSlice[0][0][:1]
	}
	return ""
}
