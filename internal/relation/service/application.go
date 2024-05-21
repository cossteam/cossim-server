package service

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/app"
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/internal/relation/app/query"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"go.uber.org/zap"
)

// NewApplication
// TODO 考虑使用依赖注入
func NewApplication(ctx context.Context, ac *config.AppConfig, logger *zap.Logger) *app.Application {
	var userAddr string
	if ac.Discovers["user"].Direct {
		userAddr = ac.Discovers["user"].Addr()
	} else {
		userAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["user"].Name)
	}

	var groupAddr string
	if ac.Discovers["group"].Direct {
		groupAddr = ac.Discovers["group"].Addr()
	} else {
		groupAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["group"].Name)
	}

	var pushAddr string
	if ac.Discovers["push"].Direct {
		pushAddr = ac.Discovers["push"].Addr()
	} else {
		pushAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["push"].Name)
	}

	var msgAddr string
	if ac.Discovers["msg"].Direct {
		msgAddr = ac.Discovers["msg"].Addr()
	} else {
		msgAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["msg"].Name)
	}

	userService, err := rpc.NewUserGrpc(userAddr)
	if err != nil {
		panic(err)
	}

	pushService, err := rpc.NewPushGrpc(pushAddr)
	if err != nil {
		panic(err)
	}

	msgService, err := rpc.NewMsgServiceGrpc(msgAddr)
	if err != nil {
		panic(err)
	}

	groupService, err := rpc.NewGroupGrpc(groupAddr)
	if err != nil {
		panic(err)
	}

	repos := persistence.NewRepositories(ac)
	userRelationDomain := service.NewUserRelationDomain(repos)
	userFriendRequestDomain := service.NewUserFriendRequestDomain(repos)
	groupRequestDomain := service.NewGroupRequestDomain(repos.GroupJoinRequestRepo, repos.GroupRepo, groupService)
	groupRelationDomain := service.NewGroupRelationDomain(repos, groupService)
	groupAnnouncementDomain := service.NewGroupAnnouncementDomain(repos.GroupAnnouncementRepo)
	dialogRelationDomain := service.NewDialogRelationDomain(repos.DialogRepo, repos.DialogUserRepo)

	return &app.Application{
		Commands: app.Commands{
			AddBlacklist: command.NewAddBlacklistHandler(
				logger,
				userRelationDomain,
				userService,
				dialogRelationDomain,
			),
			DeleteBlacklist: command.NewDeleteBlacklistHandler(
				logger,
				userRelationDomain,
				userService,
			),
			ManageFriendRequest: command.NewManageFriendRequestHandler(
				logger,
				userFriendRequestDomain,
			),
			AddFriend: command.NewAddFriendHandler(
				logger,
				userRelationDomain,
				userFriendRequestDomain,
				pushService,
			),
			SetUserBurn: command.NewSetUserBurnHandler(
				logger,
				userRelationDomain,
			),
			SetUserRemark: command.NewSetUserRemarkHandler(
				logger,
				userRelationDomain,
			),
			SetUserSilent: command.NewSetUserSilentHandler(
				logger,
				userRelationDomain,
			),
			ExchangeE2EKey: command.NewExchangeE2EKeyHandler(
				logger,
				pushService,
			),
			DeleteFriend: command.NewDeleteFriendHandler(
				logger,
				ac.Dtm.Addr(),
				userRelationDomain,
				msgService,
			),
			DeleteGroupRequest: command.NewDeleteGroupRequestHandler(logger, groupRequestDomain),
			InviteJoinGroup: command.NewInviteJoinGroupHandler(
				logger,
				groupRequestDomain,
				pushService,
			),
			ManageGroupRequest: command.NewManageGroupRequestHandler(
				logger,
				groupRequestDomain,
				groupRelationDomain,
				pushService,
			),
			SetGroupAnnouncementRead: command.NewSetGroupAnnouncementReadHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
			),
			SetGroupRemark: command.NewSetGroupRemarkHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
			),
			RemoveGroupMember: command.NewRemoveGroupMemberHandler(
				logger,
				groupRelationDomain,
			),
			DeleteFriendRequest: command.NewDeleteFriendRequestHandler(
				logger,
				userFriendRequestDomain,
			),
			DeleteGroupAnnouncement: command.NewDeleteGroupAnnouncementHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
			),
			UpdateGroupAnnouncement: command.NewUpdateGroupAnnouncementHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
			),
			ShowDialog: command.NewShowDialogHandler(
				logger,
				dialogRelationDomain,
			),
			TopDialog: command.NewTopDialogHandler(
				logger,
				dialogRelationDomain,
			),
			AddGroupAnnouncement: command.NewAddGroupAnnouncementHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
				userService,
			),
			AddGroupRequest: command.NewAddGroupRequestHandler(
				logger,
				groupRequestDomain,
				groupRelationDomain,
				pushService,
			),
			AddGroupAdmin: command.NewAddGroupAdminHandler(
				logger,
				groupRelationDomain,
			),
			QuitGroup: command.NewQuitGroupHandler(
				logger,
				groupRelationDomain,
			),
		},
		Queries: app.Queries{
			UserBlacklist: query.NewUserBlacklistHandler(
				logger,
				userRelationDomain,
				userService,
			),
			ListFriend: query.NewListFriendHandler(
				logger,
				userRelationDomain,
				userService,
			),
			ListFriendRequest: query.NewListFriendRequestHandler(
				logger,
				userRelationDomain,
				userService,
				userFriendRequestDomain,
			),
			ListGroup: query.NewListGroupHandler(
				logger,
				groupRelationDomain,
				dialogRelationDomain,
				groupService,
			),
			ListGroupMember: query.NewListGroupMemberHandler(
				logger,
				groupRelationDomain,
				userService,
			),
			ListGroupRequest: query.NewListGroupRequestHandler(
				logger,
				groupRelationDomain,
				userService,
				groupService,
			),
			ListGroupAnnouncement: query.NewListGroupAnnouncementHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
				userService,
			),
			GetGroupAnnouncement: query.NewGetGroupAnnouncementHandler(
				logger,
				groupRelationDomain,
				groupAnnouncementDomain,
				userService,
			),
		},
	}
}
