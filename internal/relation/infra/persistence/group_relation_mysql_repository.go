package persistence

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"time"
)

var _ repository.GroupRelationRepository = &MySQLRelationGroupRepository{}

func NewMySQLRelationGroupRepository(db *gorm.DB, cache cache.RelationGroupCache) *MySQLRelationGroupRepository {
	return &MySQLRelationGroupRepository{
		db:    db,
		cache: cache,
	}
}

type MySQLRelationGroupRepository struct {
	db    *gorm.DB
	cache cache.RelationGroupCache
}

func (m *MySQLRelationGroupRepository) GetByGroupID(ctx context.Context, groupID uint32) ([]*entity.GroupRelation, error) {
	var models []*po.GroupRelation

	if err := m.db.WithContext(ctx).
		Where("group_id = ? and deleted_at = 0", groupID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	var es = make([]*entity.GroupRelation, 0)
	for _, joinRequest := range models {
		es = append(es, converter.GroupRelationPoToEntity(joinRequest))
	}

	return es, nil
}

func (m *MySQLRelationGroupRepository) UpdateFieldsByGroupID(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.WithContext(ctx).Model(&po.GroupRelation{}).
		Where("id = ?", id).
		Updates(fields).
		Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelationByGroupID(context.Background(), id); err != nil {
				zap.L().Error("cache.DeleteRelationByGroupID failed", zap.Error(err))
			}
		}
	}()

	return nil
}

func (m *MySQLRelationGroupRepository) Get(ctx context.Context, groupID uint32, userID string) (*entity.GroupRelation, error) {
	model := &po.GroupRelation{}

	if err := m.db.WithContext(ctx).
		Where(&po.GroupRelation{
			GroupID: groupID,
			UserID:  userID,
			BaseModel: po.BaseModel{
				DeletedAt: 0,
			},
		}).
		First(model).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.GroupRelationPoToEntity(model), nil
}

func (m *MySQLRelationGroupRepository) Create(ctx context.Context, createGroupRelation *entity.CreateGroupRelation) (*entity.GroupRelation, error) {
	model := &po.GroupRelation{
		GroupID:     createGroupRelation.GroupID,
		Identity:    uint8(createGroupRelation.Identity),
		EntryMethod: uint8(createGroupRelation.EntryMethod),
		JoinedAt:    createGroupRelation.JoinedAt,
		UserID:      createGroupRelation.UserID,
		Inviter:     createGroupRelation.Inviter,
	}

	if err := m.db.WithContext(ctx).
		Create(model).
		Error; err != nil {
		return nil, err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelationByGroupID(ctx, model.GroupID); err != nil {
				log.Printf("cache delete relation by group id error: %v", err)
			}
		}
	}()

	return converter.GroupRelationPoToEntity(model), nil
}

func (m *MySQLRelationGroupRepository) Update(ctx context.Context, ur *entity.GroupRelation) (*entity.GroupRelation, error) {
	model := converter.GroupRelationEntityToPo(ur)

	if err := m.db.WithContext(ctx).
		Where("id = ?", ur.ID).
		Updates(model).
		Error; err != nil {
		return ur, err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelation(context.Background(), ur.UserID, ur.GroupID); err != nil {
				zap.L().Error("delete relation cache failed", zap.Error(err))
			}
		}
	}()

	return converter.GroupRelationPoToEntity(model), nil
}

func (m *MySQLRelationGroupRepository) Delete(ctx context.Context, id uint32) error {
	model := &po.GroupRelation{}
	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).
		First(model).Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelationByGroupID(context.Background(), model.GroupID); err != nil {
				zap.L().Error("delete relation cache failed", zap.Error(err))
			}
		}
	}()

	return nil
}

func (m *MySQLRelationGroupRepository) GetGroupUserIDs(ctx context.Context, gid uint32) ([]string, error) {
	var userGroupIDs []string
	var model []*po.GroupRelation

	//if m.cache != nil {
	//	relations, err := m.cache.GetGroupRelations(ctx, gid)
	//	if err == nil && len(relations) > 0 {
	//		var userIDs []string
	//		for _, v := range relations {
	//			userIDs = append(userIDs, v.UserID)
	//		}
	//		return userIDs, nil
	//	}
	//}

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("group_id = ?  AND deleted_at = 0", gid).
		Pluck("user_id", &userGroupIDs).
		Find(&model).
		Error; err != nil {
		return nil, err
	}

	go func() {
		if m.cache != nil {
			var es []*entity.GroupRelation
			for _, v := range model {
				es = append(es, converter.GroupRelationPoToEntity(v))
			}
			if err := m.cache.SetGroupRelations(ctx, gid, es, cache.RelationExpireTime); err != nil {
				zap.L().Error("cache set group relations failed", zap.Error(err))
			}
		}
	}()

	return userGroupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32

	if m.cache != nil {
		r1, err := m.cache.GetUserJoinGroupIDs(ctx, uid)
		if err == nil && len(r1) > 0 {
			return r1, nil
		}
	}

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("user_id = ? AND deleted_at = 0", uid).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.SetUserJoinGroupIDs(context.Background(), uid, groupIDs); err != nil {
				zap.L().Error("set user group ids cache failed", zap.Error(err))
			}
		}
	}()

	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserGroupByGroupIDAndUserID(ctx context.Context, gid uint32, uid string) (*entity.GroupRelation, error) {
	model := &po.GroupRelation{}

	//if m.cache != nil {
	//	r1, err := m.cache.GetRelation(ctx, uid, gid)
	//	if err == nil && r1 != nil {
	//		return r1, nil
	//	}
	//}

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? and user_id = ? AND deleted_at = 0", gid, uid).
		First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	e := converter.GroupRelationPoToEntity(model)

	go func() {
		if m.cache != nil {
			if err := m.cache.SetRelation(ctx, uid, gid, e, cache.RelationExpireTime); err != nil {
				zap.L().Error("set relation cache failed", zap.Error(err))
			}
		}
	}()

	return e, nil
}

func (m *MySQLRelationGroupRepository) GetUsersGroupByGroupIDAndUserIDs(ctx context.Context, gid uint32, uids []string) ([]*entity.GroupRelation, error) {
	var ugs []*po.GroupRelation
	var ugs2 = make([]*entity.GroupRelation, 0)

	if m.cache != nil {
		r1, err := m.cache.GetUsersGroupRelation(ctx, uids, gid)
		if err == nil && len(r1.List) > 0 {
			return r1.List, nil
		}
	}

	if err := m.db.WithContext(ctx).
		Where(" group_id = ? and user_id IN (?) AND deleted_at = 0", gid, uids).
		Find(&ugs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	for _, ug := range ugs {
		ugs2 = append(ugs2, converter.GroupRelationPoToEntity(ug))
	}

	go func() {
		if m.cache != nil {
			for _, ug := range ugs2 {
				if err := m.cache.SetRelation(context.Background(), ug.UserID, gid, ug, cache.RelationExpireTime); err != nil {
					zap.L().Error("set relation cache failed", zap.Error(err))
				}
			}
		}
	}()

	return ugs2, nil
}

func (m *MySQLRelationGroupRepository) GetUserJoinedGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32

	if m.cache != nil {
		groupIDs, err := m.cache.GetUserJoinGroupIDs(ctx, uid)
		if err == nil && len(groupIDs) > 0 {
			return groupIDs, nil
		}
	}

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("user_id = ? AND deleted_at = 0", uid).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.SetUserJoinGroupIDs(context.Background(), uid, groupIDs); err != nil {
				zap.L().Error("cache set user join group ids failed", zap.Error(err))
			}
		}
	}()

	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserManageGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Distinct("group_id").
		Where("user_id = ? AND identity IN ?", uid, []entity.GroupIdentity{entity.IdentityAdmin, entity.IdentityOwner}).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return nil, err
	}
	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) DeleteByGroupIDAndUserID(ctx context.Context, gid uint32, uid ...string) error {
	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? AND user_id IN (?) AND deleted_at = 0", gid, uid).
		Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			for _, v := range uid {
				if err := m.cache.DeleteRelation(context.Background(), v, gid); err != nil {
					log.Printf("delete group relation cache error: %v", err)
				}
			}
		}
	}()

	return nil
}

func (m *MySQLRelationGroupRepository) ListJoinRequest(ctx context.Context, gids []uint32) ([]*entity.GroupRelation, error) {
	var joinRequests []*po.GroupRelation

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("group_id IN (?)", gids).
		Find(&joinRequests).Error; err != nil {
		return nil, err
	}

	var joinRequests2 = make([]*entity.GroupRelation, 0)
	for _, joinRequest := range joinRequests {
		joinRequests2 = append(joinRequests2, converter.GroupRelationPoToEntity(joinRequest))
	}

	return joinRequests2, nil
}

func (m *MySQLRelationGroupRepository) ListGroupAdmin(ctx context.Context, gid uint32) ([]string, error) {
	var adminIds []string
	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("(identity = ? or identity = ?)AND group_id = ? AND deleted_at = 0", entity.IdentityAdmin, entity.IdentityOwner, gid).
		Pluck("user_id", &adminIds).Error; err != nil {
		return nil, err
	}
	return adminIds, nil
}

func (m *MySQLRelationGroupRepository) SetUserGroupRemark(ctx context.Context, gid uint32, uid string, remark string) error {

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("remark", remark).Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelation(ctx, uid, gid); err != nil {
				log.Printf("delete relation cache failed: %v", err)
			}
		}
	}()

	return nil
}

func (m *MySQLRelationGroupRepository) UpdateIdentity(ctx context.Context, gid uint32, uid string, identity entity.GroupIdentity) error {
	model := &po.GroupRelation{}
	model.Identity = uint8(identity)

	if err := m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("identity", model.Identity).
		Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelation(ctx, uid, gid); err != nil {
				log.Printf("delete relation cache error: %v", err)
			}
		}
	}()

	return nil
}

func (m *MySQLRelationGroupRepository) UserGroupSilentNotification(ctx context.Context, gid uint32, uid string, silentNotification bool) error {
	return m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("silent_notification", silentNotification).Error
}

func (m *MySQLRelationGroupRepository) UpdateFieldsByGroupAndUser(ctx context.Context, gid uint32, uid string, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		//Unscoped().
		Updates(fields).Error
}

func (m *MySQLRelationGroupRepository) DeleteByGroupID(ctx context.Context, gid uint32) error {
	return m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("group_id = ?", gid).
		Update("deleted_at", ptime.Now()).Error
}

func (m *MySQLRelationGroupRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&po.GroupRelation{}).
		Where("id = ?", id).
		//Unscoped().
		Updates(fields).Error
}
