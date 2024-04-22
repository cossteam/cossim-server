package app

import (
	"github.com/cossim/coss-server/internal/live/app/command"
	"github.com/cossim/coss-server/internal/live/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	LiveHandler *command.LiveHandler
}

type Queries struct {
	LiveHandler *query.LiveHandler
}

const (
	UserServiceName          = "user_service"
	RelationUserServiceName  = "relation_service"
	PushServiceName          = "push_service"
	GroupServiceName         = "group_service"
	RelationGroupServiceName = "relation_service"
)
