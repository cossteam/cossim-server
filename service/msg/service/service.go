package service

import (
	"github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/repository"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
)

func NewService(repo *persistence.Repositories) *Service {
	return &Service{
		mr: repo.Mr,
		dr: repo.Dr,
	}
}

type Service struct {
	mr repository.MsgRepository
	dr repository.DialogRepository
	v1.UnimplementedMsgServiceServer
	v1.UnimplementedDialogServiceServer
}
