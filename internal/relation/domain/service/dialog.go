package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
)

type DialogRelationDomain interface {
	TopDialog(ctx context.Context, dialogID uint32, userID string, top bool) error
	OpenOrCloseDialog(ctx context.Context, dialogID uint32, userID string, open bool) error
	DeleteFriendDialog(ctx context.Context, userID, targetID string) error
	FindDialogsForGroups(ctx context.Context, groupIDs []uint32) ([]*entity.Dialog, error)
}

var _ DialogRelationDomain = &dialogRelationDomain{}

func NewDialogRelationDomain(dialogRepo repository.DialogRepository, dialogUserRepo repository.DialogUserRepository) DialogRelationDomain {
	return &dialogRelationDomain{
		dialogRepo:     dialogRepo,
		dialogUserRepo: dialogUserRepo,
	}
}

type dialogRelationDomain struct {
	dialogRepo     repository.DialogRepository
	dialogUserRepo repository.DialogUserRepository
}

func (d *dialogRelationDomain) TopDialog(ctx context.Context, dialogID uint32, userID string, top bool) error {
	return d.updateDialogStatus(ctx, dialogID, userID, func(param *repository.UpdateDialogStatusParam) {
		var t int64
		if top {
			t = ptime.Now()
		}
		param.TopAt = &t
	})
}

func (d *dialogRelationDomain) OpenOrCloseDialog(ctx context.Context, dialogID uint32, userID string, open bool) error {
	return d.updateDialogStatus(ctx, dialogID, userID, func(param *repository.UpdateDialogStatusParam) {
		param.IsShow = &open
	})
}

func (d *dialogRelationDomain) updateDialogStatus(ctx context.Context, dialogID uint32, userID string, updateFunc func(*repository.UpdateDialogStatusParam)) error {
	_, err := d.dialogRepo.Get(ctx, dialogID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return code.DialogErrGetDialogByIdFailed.CustomMessage("不存在的对话")
		}
		return err
	}

	_, err = d.dialogUserRepo.GetByDialogIDAndUserID(ctx, dialogID, userID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return code.DialogErrGetDialogByIdFailed.CustomMessage("不存在的对话关系")
		}
		return err
	}

	param := &repository.UpdateDialogStatusParam{
		DialogID: dialogID,
		UserID:   []string{userID},
	}
	updateFunc(param)

	return d.dialogUserRepo.UpdateDialogStatus(ctx, param)
}

func (d *dialogRelationDomain) FindDialogsForGroups(ctx context.Context, groupIDs []uint32) ([]*entity.Dialog, error) {
	r, err := d.dialogRepo.Find(ctx, &repository.DialogQuery{
		GroupID:  groupIDs,
		PageSize: 0,
		PageNum:  0,
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (d *dialogRelationDomain) DeleteFriendDialog(ctx context.Context, userID, targetID string) error {
	return d.dialogUserRepo.DeleteByDialogIDAndUserID(ctx, 0, userID, targetID)
}
