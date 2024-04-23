package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	ErrCacheContentEmpty = errors.New("cache content cannot be empty")
	ErrCacheKeyEmpty     = errors.New("cache key cannot be empty")
)

const (
	GroupExpireTime = 12 * time.Hour
	GroupKeyPrefix  = "group:"
	GroupInfoKey    = GroupKeyPrefix + "info:"
)

func GetGroupInfoKey(groupID uint32) string {
	return GroupInfoKey + strconv.FormatUint(uint64(groupID), 10)
}

type GroupCache interface {
	GetGroup(ctx context.Context, groupID uint32) (*group.Group, error)
	GetGroups(ctx context.Context, groupID []uint32) ([]*group.Group, error)
	DeleteGroup(ctx context.Context, groupID ...uint32) error
	SetGroup(ctx context.Context, groups ...*group.Group) error
	DeleteAllCache(ctx context.Context) error
	Close() error
}

var _ GroupCache = &GroupCacheRedis{}

func NewGroupCacheRedis(addr, password string, db int) (*GroupCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &GroupCacheRedis{
		client: client,
	}, nil
}

type GroupCacheRedis struct {
	client *redis.Client
}

func (g *GroupCacheRedis) Close() error {
	return g.client.Close()
}

func (g *GroupCacheRedis) DeleteAllCache(ctx context.Context) error {
	keys := make([]string, 0)
	iter := g.client.Scan(ctx, 0, GroupKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return g.client.Del(ctx, keys...).Err()
}

func (g *GroupCacheRedis) SetGroup(ctx context.Context, groups ...*group.Group) error {
	if len(groups) == 0 {
		return ErrCacheKeyEmpty
	}

	args := make([]interface{}, 0, len(groups)*2+1)
	for _, group := range groups {
		key := GetGroupInfoKey(group.ID)
		groupJSON, err := json.Marshal(group)
		if err != nil {
			return err
		}

		args = append(args, key, groupJSON)
	}

	// 使用 MSET 命令设置键值对
	if err := g.client.MSet(ctx, args...).Err(); err != nil {
		return err
	}

	// 设置过期时间
	for _, group := range groups {
		key := GetGroupInfoKey(group.ID)
		if err := g.client.Expire(ctx, key, GroupExpireTime).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (g *GroupCacheRedis) GetGroup(ctx context.Context, groupID uint32) (*group.Group, error) {
	key := GetGroupInfoKey(groupID)
	groupJSON, err := g.client.Get(ctx, key).Result()
	if err != nil {
		//if err == redis.Nil {
		//	return nil, nil // 当缓存中不存在时，返回 nil 作为结果，以区分缓存不存在和其他错误
		//}
		return nil, err
	}

	group := &group.Group{}
	if err = json.Unmarshal([]byte(groupJSON), group); err != nil {
		return nil, err
	}

	return group, nil
}

func (g *GroupCacheRedis) GetGroups(ctx context.Context, groupIDs []uint32) ([]*group.Group, error) {
	if len(groupIDs) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		keys[i] = GroupInfoKey + strconv.FormatUint(uint64(id), 10)
	}

	groupJSONs, err := g.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	groups := make([]*group.Group, 0, len(groupIDs))
	for _, groupJSON := range groupJSONs {
		if groupJSON == nil {
			continue
		}

		group := &group.Group{}
		if err := json.Unmarshal([]byte(groupJSON.(string)), group); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (g *GroupCacheRedis) DeleteGroup(ctx context.Context, groupIDs ...uint32) error {
	if len(groupIDs) == 0 {
		return ErrCacheKeyEmpty
	}
	keys := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		keys[i] = GroupInfoKey + strconv.FormatUint(uint64(id), 10)
	}
	return g.client.Del(ctx, keys...).Err()
}
