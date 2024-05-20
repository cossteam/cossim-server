package app

import (
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/internal/relation/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	AddBlacklist             command.AddBlacklistHandler
	DeleteBlacklist          command.DeleteBlacklistHandler
	DeleteFriendRequest      command.DeleteFriendRequestHandler
	ManageFriendRequest      command.ManageFriendRequestHandler
	AddFriend                command.AddFriendHandler
	SetUserBurn              command.SetUserBurnHandler
	SetUserRemark            command.SetUserRemarkHandler
	SetUserSilent            command.SetUserSilentHandler
	ExchangeE2EKey           command.ExchangeE2EKeyHandler
	DeleteFriend             command.DeleteFriendHandler
	DeleteGroupRequest       command.DeleteGroupRequestHandler
	InviteJoinGroup          command.InviteJoinGroupHandler
	AddGroupRequest          command.AddGroupRequestHandler
	ManageGroupRequest       command.ManageGroupRequestHandler
	QuitGroup                command.QuitGroupHandler
	SetGroupSilent           command.SetGroupSilentHandler
	SetGroupAnnouncementRead command.SetGroupAnnouncementReadHandler
	SetGroupRemark           command.SetGroupRemarkHandler
	RemoveGroupMember        command.RemoveGroupMemberHandler
	AddGroupAdmin            command.AddGroupAdminHandler
	AddGroupAnnouncement     command.AddGroupAnnouncementHandler
	DeleteGroupAnnouncement  command.DeleteGroupAnnouncementHandler
	UpdateGroupAnnouncement  command.UpdateGroupAnnouncementHandler
	ShowDialog               command.ShowDialogHandler
	TopDialog                command.TopDialogHandler
}

type Queries struct {
	UserBlacklist             query.UserBlacklistHandler
	ListFriendRequest         query.ListFriendRequestHandler
	ListFriend                query.ListFriendHandler
	ListGroupMember           query.ListGroupMemberHandler
	ListGroupRequest          query.ListGroupRequestHandler
	ListGroup                 query.ListGroupHandler
	ListGroupAnnouncement     query.ListGroupAnnouncementHandler
	GetGroupAnnouncement      query.GetGroupAnnouncementHandler
	ListGroupAnnouncementRead query.ListGroupAnnouncementReadHandler
}
