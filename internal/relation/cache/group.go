package cache

import (
	"context"
	"encoding/json"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
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
	GetRelation(ctx context.Context, userID string, groupID uint32) (*relation.GroupRelation, error)
	GetRelations(ctx context.Context, userID string, groupID []uint32) (*relation.GroupRelationList, error)
	SetRelation(ctx context.Context, userID string, groupID uint32, data *relation.GroupRelation, expiration time.Duration) error
	DeleteRelation(ctx context.Context, userID string, groupID uint32) error

	DeleteRelationByGroupID(ctx context.Context, groupID uint32) error

	GetUsersGroupRelation(ctx context.Context, userID []string, groupID uint32) (*relation.GroupRelationList, error)

	GetGroupMember(ctx context.Context, groupID uint32)

	GetGroupJoinRequestListByUser(ctx context.Context, userID string) (*relation.GroupJoinRequestList, error)
	SetGroupJoinRequestListByUser(ctx context.Context, userID string, data *relation.GroupJoinRequestList, expiration time.Duration) error

	DeleteGroupJoinRequestListByUser(ctx context.Context, userID ...string) error

	// GetUserGroupIDs 获取用户加入的群聊id列表
	GetUserGroupIDs(ctx context.Context, userID string) ([]uint32, error)
	SetUserGroupIDs(ctx context.Context, userID string, groupIDs []uint32) error
	DeleteUserGroupIDs(ctx context.Context, userID ...string) error
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

func (r *RelationGroupCacheRedis) SetUserGroupIDs(ctx context.Context, userID string, groupIDs []uint32) error {
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

func (r *RelationGroupCacheRedis) DeleteUserGroupIDs(ctx context.Context, userID ...string) error {
	if len(userID) == 0 {
		return ErrCacheKeyEmpty
	}

	keys := make([]string, len(userID))
	for i, id := range userID {
		keys[i] = GetGroupUserJoinGroupListKey(id)
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationGroupCacheRedis) GetUserGroupIDs(ctx context.Context, userID string) ([]uint32, error) {
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

func (r *RelationGroupCacheRedis) SetGroupJoinRequestListByUser(ctx context.Context, userID string, data *relation.GroupJoinRequestList, expiration time.Duration) error {
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
	//TODO implement me
	panic("implement me")
}

func (r *RelationGroupCacheRedis) GetGroupJoinRequestListByUser(ctx context.Context, userID string) (*relation.GroupJoinRequestList, error) {
	key := GetGroupJoinRequestListByUserIDKey(userID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	friendRequestList := &relation.GroupJoinRequestList{}
	if err := json.Unmarshal([]byte(val), friendRequestList); err != nil {
		return nil, err
	}

	return friendRequestList, nil
}

func (r *RelationGroupCacheRedis) GetGroupMember(ctx context.Context, groupID uint32) {
	//TODO implement me
	panic("implement me")
}

func (r *RelationGroupCacheRedis) DeleteRelationByGroupId(ctx context.Context, groupID uint32) error {
	if groupID == 0 {
		return ErrCacheKeyEmpty
	}

	// 构建匹配模式，以便扫描符合条件的键
	pattern := GetGroupKey("*", groupID)

	// 使用 SCAN 命令进行键的扫描和删除
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

func (r *RelationGroupCacheRedis) GetUsersGroupRelation(ctx context.Context, userID []string, groupID uint32) (*relation.GroupRelationList, error) {
	if groupID == 0 || len(userID) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	pipeline := r.client.Pipeline()

	results := &relation.GroupRelationList{}

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
			var response relation.GroupRelation
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

func (r *RelationGroupCacheRedis) GetRelation(ctx context.Context, userID string, groupID uint32) (*relation.GroupRelation, error) {
	if groupID == 0 || userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	key := GetGroupKey(userID, groupID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	rel := &relation.GroupRelation{}
	if err := json.Unmarshal([]byte(val), rel); err != nil {
		return nil, err
	}

	return rel, nil
}

func (r *RelationGroupCacheRedis) GetRelations(ctx context.Context, userID string, groupID []uint32) (*relation.GroupRelationList, error) {
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

	relations := &relation.GroupRelationList{}
	for _, val := range vals {
		if val == nil {
			continue
		}

		rel := &relation.GroupRelation{}
		if err := json.Unmarshal([]byte(val.(string)), rel); err != nil {
			return nil, err
		}

		relations.List = append(relations.List, rel)
		relations.Total++
	}

	return relations, nil
}

func (r *RelationGroupCacheRedis) SetRelation(ctx context.Context, userID string, groupID uint32, data *relation.GroupRelation, expiration time.Duration) error {
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
