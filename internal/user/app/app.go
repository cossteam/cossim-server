package app

import (
	"github.com/cossim/coss-server/internal/user/app/command"
	"github.com/cossim/coss-server/internal/user/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	UserLogin                 command.UserLoginHandler
	UserLogout                command.UserLogoutHandler
	UpdatePassword            command.UpdatePasswordHandler
	UserActivate              command.UserActivateHandler
	UserRegister              command.UserRegisterHandler
	UpdateUserBundle          command.UpdateUserBundleHandler
	SetUserPublicKey          command.SetUserPublicKeyHandler
	SendUserEmailVerification command.SendUserEmailVerificationHandler
	ResetUserPublicKey        command.ResetUserPublicKeyHandler
	UpdateUser                command.UpdateUserHandler
	UpdateUserAvatarHandler   command.UpdateUserAvatarHandler
	GenerateQRCode            command.GenerateQRCodeHandler
	UpdateQRCode              command.UpdateQRCodeHandler
	SSOLogin                  command.SSOLoginHandler
	//CreateGroup command.CreateGroupHandler
	//DeleteGroup command.DeleteGroupHandler
	//UpdateGroup command.UpdateGroupHandler
}

type Queries struct {
	GetUser             query.GetUserHandler
	GetUserBundle       query.GetUserBundleHandler
	GetUserLoginClients query.GetUserClientsHandler
	GetQRCode           query.GetQRCodeHandler
	//GetGroup    query.GetGroupHandler
	//SearchGroup query.SearchGroupHandler
}
