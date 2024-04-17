package app

import (
	"github.com/cossim/coss-server/internal/group/app/command"
	"github.com/cossim/coss-server/internal/group/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateGroup command.CreateGroupHandler
	DeleteGroup command.DeleteGroupHandler
	UpdateGroup command.UpdateGroupHandler
}

type Queries struct {
	GetGroup query.GetGroupHandler
}

const (
	UserServiceName          = "user_service"
	RelationUserServiceName  = "relation_service"
	PushServiceName          = "push_service"
	GroupServiceName         = "group_service"
	RelationGroupServiceName = "relation_service"
)
