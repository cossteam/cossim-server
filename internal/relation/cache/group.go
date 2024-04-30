package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const (
	RelationGroupKey                        = RelationKeyPrefix + "group_info:"
	RelationGroupJoinRequestListByUserIDKey = RelationKeyPrefix + "group_join_request_list:"
	RelationGroupUserJoinGroupListKey       = RelationKeyPrefix + "group_user_join_group_list:"
)

func GetGroupKey(userID string, groupID uint32) string {
	return RelationGroupKey + userID + ":" + strconv.FormatUint(uint64(groupID), 10)
}

func GetGroupJoinRequestListByUserIDKey(userID string) string {
	return RelationGroupJoinRequestListByUserIDKey + userID
}

func GetGroupUserJoinGroupListKey(userID string) string {
	return RelationGroupUserJoinGroupListKey + userID
}

type RelationGroupCache interface {
	GetRelation(ctx context.Context, userID string, groupID uint32) (*entity.GroupRelation, error)
	GetRelations(ctx context.Context, userID string, groupID []uint32) (*entity.GroupRelationList, error)
	SetRelation(ctx context.Context, userID string, groupID uint32, data *entity.GroupRelation, expiration time.Duration) error
	DeleteRelation(ctx context.Context, userID string, groupID uint32) error

	DeleteRelationByGroupID(ctx context.Context, groupID uint32) error

	GetUsersGroupRelation(ctx context.Context, userID []string, groupID uint32) (*entity.GroupRelationList, error)

	// GetGroupRelations 获取群聊下所有群聊用户的关系
	GetGroupRelations(ctx context.Context, groupID uint32) ([]*entity.GroupRelation, error)
	// SetGroupRelations 设置群聊下所有群聊用户的关系
	SetGroupRelations(ctx context.Context, groupID uint32, data []*entity.GroupRelation, expiration time.Duration) error

	GetGroupJoinRequestListByUser(ctx context.Context, userID string) (*entity.GroupJoinRequestList, error)
	SetGroupJoinRequestListByUser(ctx context.Context, userID string, data *entity.GroupJoinRequestList, expiration time.Duration) error

	DeleteGroupJoinRequestListByUser(ctx context.Context, userID ...string) error

	// GetUserJoinGroupIDs 获取用户加入的群聊id列表
	GetUserJoinGroupIDs(ctx context.Context, userID string) ([]uint32, error)
	SetUserJoinGroupIDs(ctx context.Context, userID string, groupIDs []uint32) error
	DeleteUserJoinGroupIDs(ctx context.Context, userID ...string) error

	DeleteAllCache(ctx context.Context) error
}

var _ RelationGroupCache = &RelationGroupCacheRedis{}

func NewRelationGroupCacheRedis(addr, password string, db int) (*RelationGroupCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RelationGroupCacheRedis{
		client: client,
	}, nil
}

type RelationGroupCacheRedis struct {
	client *redis.Client
}

func (r *RelationGroupCacheRedis) DeleteAllCache(ctx context.Context) error {
	keys := make([]string, 0)
	iter := r.client.Scan(ctx, 0, RelationKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationGroupCacheRedis) SetGroupRelations(ctx context.Context, groupID uint32, data []*entity.GroupRelation, expiration time.Duration) error {
	if groupID == 0 || data == nil {
		return ErrCacheKeyEmpty
	}

	pipeline := r.client.Pipeline()

	for _, rel := range data {
		key := GetGroupKey(rel.UserID, groupID)
		b, err := json.Marshal(rel)
		if err != nil {
			return err
		}
		pipeline.Set(ctx, key, b, expiration)
	}

	_, err := pipeline.Exec(ctx)
	return err
}

func (r *RelationGroupCacheRedis) GetGroupRelations(ctx context.Context, groupID uint32) ([]*entity.GroupRelation, error) {
	if groupID == 0 {
		return nil, ErrCacheKeyEmpty
	}

	pattern := GetGroupKey("*", groupID)

	var relations []*entity.GroupRelation

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		rel := &entity.GroupRelation{}
		if err := json.Unmarshal([]byte(val), rel); err != nil {
			return nil, err
		}

		relations = append(relations, rel)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return relations, nil
}

func (r *RelationGroupCacheRedis) SetUserJoinGroupIDs(ctx context.Context, userID string, groupIDs []uint32) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}

	groupIDsStr := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		groupIDsStr[i] = strconv.FormatUint(uint64(id), 10)
	}

	groupIDsJSON, err := json.Marshal(groupIDsStr)
	if err != nil {
		return err
	}

	key := GetGroupUserJoinGroupListKey(userID)
	return r.client.Set(ctx, key, groupIDsJSON, 0).Err()
}

func (r *RelationGroupCacheRedis) DeleteUserJoinGroupIDs(ctx context.Context, userID ...string) error {
	if len(userID) == 0 {
		return ErrCacheKeyEmpty
	}

	keys := make([]string, len(userID))
	for i, id := range userID {
		keys[i] = GetGroupUserJoinGroupListKey(id)
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationGroupCacheRedis) GetUserJoinGroupIDs(ctx context.Context, userID string) ([]uint32, error) {
	if userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	key := GetGroupUserJoinGroupListKey(userID)
	var resp []uint32
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *RelationGroupCacheRedis) DeleteGroupJoinRequestListByUser(ctx context.Context, userIDs ...string) error {
	if len(userIDs) == 0 {
		return ErrCacheKeyEmpty
	}

	keys := make([]string, len(userIDs))
	for i, targetUserID := range userIDs {
		keys[i] = GetGroupJoinRequestListByUserIDKey(targetUserID)
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationGroupCacheRedis) SetGroupJoinRequestListByUser(ctx context.Context, userID string, data *entity.GroupJoinRequestList, expiration time.Duration) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}

	key := GetGroupJoinRequestListByUserIDKey(userID)
	if data == nil {
		return ErrCacheContentEmpty
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationGroupCacheRedis) DeleteRelationByGroupID(ctx context.Context, groupID uint32) error {
	if groupID == 0 {
		return ErrCacheKeyEmpty
	}

	pattern := GetGroupKey("*", groupID)

	fmt.Println("pattern => ", pattern)

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		fmt.Println("iter.Val() => ", iter.Val())
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

func (r *RelationGroupCacheRedis) GetGroupJoinRequestListByUser(ctx context.Context, userID string) (*entity.GroupJoinRequestList, error) {
	key := GetGroupJoinRequestListByUserIDKey(userID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	friendRequestList := &entity.GroupJoinRequestList{}
	if err := json.Unmarshal([]byte(val), friendRequestList); err != nil {
		return nil, err
	}

	return friendRequestList, nil
}

func (r *RelationGroupCacheRedis) GetUsersGroupRelation(ctx context.Context, userID []string, groupID uint32) (*entity.GroupRelationList, error) {
	if groupID == 0 || len(userID) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	pipeline := r.client.Pipeline()

	results := &entity.GroupRelationList{}

	for _, uID := range userID {
		key := GetGroupKey(uID, groupID)
		pipeline.Get(ctx, key)
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}

	for _, cmd := range cmds {
		result, err := cmd.(*redis.StringCmd).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			} else {
				return nil, err
			}
		} else {
			var response entity.GroupRelation
			err = json.Unmarshal([]byte(result), &response)
			if err != nil {
				return nil, err
			}

			results.List = append(results.List, &response)
			results.Total++
		}
	}

	return results, nil
}

func (r *RelationGroupCacheRedis) GetRelation(ctx context.Context, userID string, groupID uint32) (*entity.GroupRelation, error) {
	if groupID == 0 || userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	key := GetGroupKey(userID, groupID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	rel := &entity.GroupRelation{}
	if err := json.Unmarshal([]byte(val), rel); err != nil {
		return nil, err
	}

	return rel, nil
}

func (r *RelationGroupCacheRedis) GetRelations(ctx context.Context, userID string, groupID []uint32) (*entity.GroupRelationList, error) {
	if len(groupID) == 0 || userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(groupID))
	for i, targetUserID := range groupID {
		keys[i] = GetGroupKey(userID, targetUserID)
	}

	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	relations := &entity.GroupRelationList{}
	for _, val := range vals {
		if val == nil {
			continue
		}

		rel := &entity.GroupRelation{}
		if err := json.Unmarshal([]byte(val.(string)), rel); err != nil {
			return nil, err
		}

		relations.List = append(relations.List, rel)
		relations.Total++
	}

	return relations, nil
}

func (r *RelationGroupCacheRedis) SetRelation(ctx context.Context, userID string, groupID uint32, data *entity.GroupRelation, expiration time.Duration) error {
	if groupID == 0 || userID == "" {
		return ErrCacheKeyEmpty
	}

	key := GetGroupKey(userID, groupID)
	if data == nil {
		return ErrCacheContentEmpty
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationGroupCacheRedis) DeleteRelation(ctx context.Context, userID string, groupID uint32) error {
	if groupID == 0 || userID == "" {
		return ErrCacheKeyEmpty
	}

	key := GetGroupKey(userID, groupID)
	return r.client.Del(ctx, key).Err()
}
