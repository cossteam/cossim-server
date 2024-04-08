package cache

import (
	"context"
	"encoding/json"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	RelationExpireTime           = 12 * time.Hour
	RelationKeyPrefix            = "relation:"
	RelationFriendKey            = RelationKeyPrefix + "friend_info:"
	RelationFriendIDsKey         = RelationKeyPrefix + "friend_ids:"
	RelationFriendListKey        = RelationKeyPrefix + "friend_list:"
	RelationFriendRequestKey     = RelationKeyPrefix + "friend_request:"
	RelationFriendRequestListKey = RelationKeyPrefix + "friend_request_list:"
	RelationBlacklistKey         = RelationKeyPrefix + "blacklist:"
)

func GetFriendKey(ownerUserID string, targetUserID string) string {
	return RelationFriendKey + ownerUserID + ":" + targetUserID
}

func GetFriendIDsKey(ownerUserID string) string {
	return RelationFriendIDsKey + ownerUserID
}

func GetFriendRequestKey(ownerUserID string) string {
	return RelationFriendRequestKey + ownerUserID
}

func GetFriendListKey(ownerUserID string) string {
	return RelationFriendListKey + ownerUserID
}

func GetFriendRequestListKey(ownerUserID string) string {
	return RelationFriendRequestListKey + ownerUserID
}

func GetBlacklistKey(ownerUserID string) string {
	return RelationBlacklistKey + ownerUserID
}

type RelationUserCache interface {
	GetRelation(ctx context.Context, ownerUserID string, targetUserID string) (*relationgrpcv1.GetUserRelationResponse, error)
	GetRelations(ctx context.Context, ownerUserID string, targetUserID []string) ([]*relationgrpcv1.GetUserRelationResponse, error)
	SetRelation(ctx context.Context, ownerUserID string, targetUserID string, data *relationgrpcv1.GetUserRelationResponse, expiration time.Duration) error
	DeleteRelation(ctx context.Context, ownerUserID string, targetUserID string) error
	DeleteRelations(ctx context.Context, ownerUserID string, targetUserIDs []string) error
	DeleteFriendIDs(ctx context.Context, userIDs ...string) error
	GetFriendList(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetFriendListResponse, error)
	SetFriendList(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetFriendListResponse, expiration time.Duration) error
	DeleteFriendList(ctx context.Context, ownerUserID ...string) error
	GetFriendRequest(ctx context.Context, ownerUserID string) (*relationgrpcv1.FriendRequestList, error)
	SetFriendRequest(ctx context.Context, ownerUserID string, data *relationgrpcv1.FriendRequestList, expiration time.Duration) error
	GetFriendRequestList(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetFriendRequestListResponse, error)
	SetFriendRequestList(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetFriendRequestListResponse, expiration time.Duration) error
	DeleteFriendRequestList(ctx context.Context, ownerUserID ...string) error
	GetBlacklist(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetBlacklistResponse, error)
	SetBlacklist(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetBlacklistResponse, expiration time.Duration) error
	DeleteBlacklist(ctx context.Context, ownerUserID ...string) error
}

var _ RelationUserCache = &RelationUserCacheRedis{}

func NewRelationUserCacheRedis(addr, password string, db int) (*RelationUserCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RelationUserCacheRedis{
		client: client,
	}, nil
}

type RelationUserCacheRedis struct {
	client *redis.Client
}

func (r *RelationUserCacheRedis) GetFriendRequest(ctx context.Context, ownerUserID string) (*relationgrpcv1.FriendRequestList, error) {
	if ownerUserID == "" {
		return nil, ErrCacheKeyEmpty
	}
	key := GetFriendRequestKey(ownerUserID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var friendRequestList relationgrpcv1.FriendRequestList
	if err := json.Unmarshal([]byte(val), &friendRequestList); err != nil {
		return nil, err
	}

	return &friendRequestList, nil
}

func (r *RelationUserCacheRedis) SetFriendRequest(ctx context.Context, ownerUserID string, data *relationgrpcv1.FriendRequestList, expiration time.Duration) error {
	if ownerUserID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetFriendRequestKey(ownerUserID)
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationUserCacheRedis) DeleteFriendList(ctx context.Context, ownerUserID ...string) error {
	if len(ownerUserID) == 0 {
		return ErrCacheKeyEmpty
	}
	keys := make([]string, 0, len(ownerUserID))
	for _, userID := range ownerUserID {
		keys = append(keys, GetFriendListKey(userID))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationUserCacheRedis) DeleteFriendRequestList(ctx context.Context, ownerUserID ...string) error {
	if len(ownerUserID) == 0 {
		return ErrCacheKeyEmpty
	}
	keys := make([]string, 0, len(ownerUserID))
	for _, userID := range ownerUserID {
		keys = append(keys, GetFriendRequestListKey(userID))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationUserCacheRedis) DeleteBlacklist(ctx context.Context, ownerUserID ...string) error {
	if len(ownerUserID) == 0 {
		return ErrCacheKeyEmpty
	}
	keys := make([]string, 0, len(ownerUserID))
	for _, userID := range ownerUserID {
		keys = append(keys, GetBlacklistKey(userID))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationUserCacheRedis) GetBlacklist(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetBlacklistResponse, error) {
	key := GetBlacklistKey(ownerUserID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var blacklist relationgrpcv1.GetBlacklistResponse
	if err = json.Unmarshal([]byte(val), &blacklist); err != nil {
		return nil, err
	}
	return &blacklist, nil
}

func (r *RelationUserCacheRedis) SetBlacklist(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetBlacklistResponse, expiration time.Duration) error {
	key := GetBlacklistKey(ownerUserID)
	if data == nil {
		return r.client.Del(ctx, key).Err()
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationUserCacheRedis) GetFriendRequestList(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetFriendRequestListResponse, error) {
	key := GetFriendRequestListKey(ownerUserID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	friendRequestList := &relationgrpcv1.GetFriendRequestListResponse{}
	if err := json.Unmarshal([]byte(val), friendRequestList); err != nil {
		return nil, err
	}

	return friendRequestList, nil
}

func (r *RelationUserCacheRedis) SetFriendRequestList(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetFriendRequestListResponse, expiration time.Duration) error {
	key := GetFriendRequestListKey(ownerUserID)
	if data == nil {
		return ErrCacheContentEmpty
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationUserCacheRedis) SetFriendList(ctx context.Context, ownerUserID string, data *relationgrpcv1.GetFriendListResponse, expiration time.Duration) error {
	key := GetFriendListKey(ownerUserID)
	if data == nil {
		return ErrCacheContentEmpty
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RelationUserCacheRedis) GetFriendList(ctx context.Context, ownerUserID string) (*relationgrpcv1.GetFriendListResponse, error) {
	key := GetFriendListKey(ownerUserID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	friendList := &relationgrpcv1.GetFriendListResponse{}
	if err := json.Unmarshal([]byte(val), friendList); err != nil {
		return nil, err
	}

	return friendList, nil
}

func (r *RelationUserCacheRedis) GetRelation(ctx context.Context, ownerUserID string, targetUserID string) (*relationgrpcv1.GetUserRelationResponse, error) {
	key := GetFriendKey(ownerUserID, targetUserID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	relation := &relationgrpcv1.GetUserRelationResponse{}
	if err := json.Unmarshal([]byte(val), relation); err != nil {
		return nil, err
	}

	return relation, nil
}

func (r *RelationUserCacheRedis) GetRelations(ctx context.Context, ownerUserID string, targetUserIDs []string) ([]*relationgrpcv1.GetUserRelationResponse, error) {
	keys := make([]string, len(targetUserIDs))
	for i, targetUserID := range targetUserIDs {
		keys[i] = GetFriendKey(ownerUserID, targetUserID)
	}

	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	relations := make([]*relationgrpcv1.GetUserRelationResponse, len(vals))
	for i, val := range vals {
		if val == nil {
			continue
		}

		relation := &relationgrpcv1.GetUserRelationResponse{}
		if err := json.Unmarshal([]byte(val.(string)), relation); err != nil {
			return nil, err
		}

		relations[i] = relation
	}

	return relations, nil
}

func (r *RelationUserCacheRedis) DeleteRelation(ctx context.Context, ownerUserID string, targetUserID string) error {
	key := GetFriendKey(ownerUserID, targetUserID)
	return r.client.Del(ctx, key).Err()
}

func (r *RelationUserCacheRedis) DeleteRelations(ctx context.Context, ownerUserID string, targetUserIDs []string) error {
	keys := make([]string, len(targetUserIDs))
	for i, targetUserID := range targetUserIDs {
		keys[i] = GetFriendKey(ownerUserID, targetUserID)
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationUserCacheRedis) DeleteFriendIDs(ctx context.Context, userIDs ...string) error {
	// 构建 Redis 中所有关系对象的键
	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, GetFriendIDsKey(userID))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RelationUserCacheRedis) SetRelation(ctx context.Context, ownerUserID string, targetUserID string, data *relationgrpcv1.GetUserRelationResponse, expiration time.Duration) error {
	key := GetFriendKey(ownerUserID, targetUserID)
	if data == nil {
		return ErrCacheContentEmpty
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, expiration).Err()
}
