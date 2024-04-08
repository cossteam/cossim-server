package cache

import (
	"context"
	"encoding/json"
	"fmt"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	UserExpireTime = 12 * time.Hour
	UserKeyPrefix  = "user:"
	UserInfoKey    = UserKeyPrefix + "user_info:"
)

func GetUserInfoKey(userID string) string {
	return UserInfoKey + userID
}

type UserCache interface {
	GetUserInfo(ctx context.Context, userID string) (*usergrpcv1.UserInfoResponse, error)
	GetUsersInfo(ctx context.Context, userID []string) ([]*usergrpcv1.UserInfoResponse, error)
	SetUserInfo(ctx context.Context, userID string, data *usergrpcv1.UserInfoResponse, expiration time.Duration) error
	DeleteUsersInfo(userIDs []string) error
}

var _ UserCache = &UserCacheRedis{}

func NewUserCacheRedis(addr, password string, db int) (*UserCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &UserCacheRedis{
		client: client,
	}, nil
}

type UserCacheRedis struct {
	client *redis.Client
}

func (u *UserCacheRedis) GetUserInfo(ctx context.Context, userID string) (*usergrpcv1.UserInfoResponse, error) {
	if userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	key := GetUserInfoKey(userID)
	val, err := u.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user info from cache: %v", err)
	}

	var userInfo usergrpcv1.UserInfoResponse
	if err = json.Unmarshal([]byte(val), &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	return &userInfo, nil
}

func (u *UserCacheRedis) GetUsersInfo(ctx context.Context, userIDs []string) ([]*usergrpcv1.UserInfoResponse, error) {
	if len(userIDs) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		if userID == "" {
			return nil, ErrCacheKeyEmpty
		}
		keys[i] = GetUserInfoKey(userID)
	}

	vals, err := u.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get users info from cache: %v", err)
	}

	userInfos := make([]*usergrpcv1.UserInfoResponse, len(userIDs))
	for i, val := range vals {
		if val == nil {
			continue
		}

		var userInfo usergrpcv1.UserInfoResponse
		err = json.Unmarshal([]byte(val.(string)), &userInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
		}

		userInfos[i] = &userInfo
	}

	return userInfos, nil
}

func (u *UserCacheRedis) SetUserInfo(ctx context.Context, userID string, data *usergrpcv1.UserInfoResponse, expiration time.Duration) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	if data == nil {
		return ErrCacheContentEmpty
	}

	key := GetUserInfoKey(userID)
	userInfoJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}

	return u.client.Set(ctx, key, userInfoJSON, expiration).Err()
}

func (u *UserCacheRedis) DeleteUsersInfo(userIDs []string) error {
	if len(userIDs) == 0 {
		return ErrCacheKeyEmpty
	}
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		if userID == "" {
			return ErrCacheKeyEmpty
		}
		keys[i] = GetUserInfoKey(userID)
	}
	return u.client.Del(context.Background(), keys...).Err()
}
