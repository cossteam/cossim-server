package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrCacheContentEmpty = errors.New("cache content cannot be empty")
	ErrCacheKeyEmpty     = errors.New("cache key cannot be empty")
)

const (
	UserExpireTime                      = 12 * time.Hour
	UserEmailVerificationCodeExpireTime = 30 * time.Minute
	UserVerificationCodeExpireTime      = 5 * time.Minute
	UserLoginExpireTime                 = 24 * 7 * time.Hour
	UserKeyPrefix                       = "user:"
	UserInfoKey                         = UserKeyPrefix + "info:"
	UserLoginKey                        = UserKeyPrefix + "login:"
	UserVerificationCode                = UserKeyPrefix + "verification_code:"
	UserEmailVerificationCode           = UserKeyPrefix + "email_verification_code:"
)

func GetUserInfoKey(userID string) string {
	return UserInfoKey + userID
}

func GetUserLoginKey(userID string) string {
	return UserLoginKey + userID
}

func GetUserLoginDriveKey(userID string, driverID string) string {
	return UserLoginKey + userID + ":" + fmt.Sprintf("%s", driverID)
}

func GetUserVerificationCodeKey(userID, code string) string {
	return UserVerificationCode + userID + ":" + code
}

func GetUserEmailVerificationCodeKey(userID string) string {
	return UserEmailVerificationCode + userID
}

type UserCache interface {
	GetUserInfo(ctx context.Context, userID string) (*entity.User, error)
	GetUsersInfo(ctx context.Context, userID []string) ([]*entity.User, error)
	SetUserInfo(ctx context.Context, userID string, data *entity.User, expiration time.Duration) error
	DeleteUsersInfo(ctx context.Context, userIDs []string) error
	DeleteAllCache(ctx context.Context) error
	GetUserLoginInfo(ctx context.Context, userID string, driverID string) (*entity.UserLogin, error)
	SetUserLoginInfo(ctx context.Context, userID string, driverID string, data *entity.UserLogin, expiration time.Duration) error
	GetUsersLoginInfo(ctx context.Context, userID []string) ([]*entity.UserLogin, error)
	DeleteUserLoginInfo(ctx context.Context, userID string, driverID string) error
	DeleteUserAllLoginInfo(ctx context.Context, userID string) error
	GetUserLoginInfos(ctx context.Context, userID string) ([]*entity.UserLogin, error)
	GetUserEmailVerificationCode(ctx context.Context, userID string) (string, error)
	SetUserEmailVerificationCode(ctx context.Context, userID, code string, expiration time.Duration) error
	DeleteUserEmailVerificationCode(ctx context.Context, userID string) error
	GetUserVerificationCode(ctx context.Context, userID, code string) (string, error)
	SetUserVerificationCode(ctx context.Context, userID, code string, expiration time.Duration) error
	DeleteUserVerificationCode(ctx context.Context, userID, code string) error
	Close() error
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

func NewUserCacheRedisWithClient(client *redis.Client) *UserCacheRedis {
	return &UserCacheRedis{
		client: client,
	}
}

type UserCacheRedis struct {
	client *redis.Client
}

func (u *UserCacheRedis) DeleteUserAllLoginInfo(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}

	loginKeyPattern := UserLoginKey + userID + ":*"

	iter := u.client.Scan(ctx, 0, loginKeyPattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := u.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

func (u *UserCacheRedis) Close() error {
	return u.client.Close()
}

func (u *UserCacheRedis) GetUserVerificationCode(ctx context.Context, userID, code string) (string, error) {
	if userID == "" {
		return "", ErrCacheKeyEmpty
	}
	key := GetUserVerificationCodeKey(userID, code)
	code, err := u.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return code, nil
}

func (u *UserCacheRedis) SetUserVerificationCode(ctx context.Context, userID string, code string, expiration time.Duration) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetUserVerificationCodeKey(userID, code)
	return u.client.Set(ctx, key, code, expiration).Err()
}

func (u *UserCacheRedis) DeleteUserVerificationCode(ctx context.Context, userID, code string) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetUserVerificationCodeKey(userID, code)
	return u.client.Del(ctx, key).Err()
}

func (u *UserCacheRedis) GetUserInfos(ctx context.Context, userID string) ([]*entity.UserLogin, error) {
	key := UserInfoKey + userID
	infoStrings, err := u.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	userInfos := make([]*entity.UserLogin, 0)
	for _, infoString := range infoStrings {
		var userInfo entity.UserLogin
		if err := json.Unmarshal([]byte(infoString), &userInfo); err != nil {
			return nil, err
		}
		userInfos = append(userInfos, &userInfo)
	}

	return userInfos, nil
}

func (u *UserCacheRedis) GetUserEmailVerificationCode(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", ErrCacheKeyEmpty
	}
	key := GetUserEmailVerificationCodeKey(userID)
	code, err := u.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return code, nil
}

func (u *UserCacheRedis) SetUserEmailVerificationCode(ctx context.Context, userID string, code string, expiration time.Duration) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetUserEmailVerificationCodeKey(userID)
	return u.client.Set(ctx, key, code, expiration).Err()
}

func (u *UserCacheRedis) DeleteUserEmailVerificationCode(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetUserEmailVerificationCodeKey(userID)
	return u.client.Del(ctx, key).Err()
}

func (u *UserCacheRedis) SetUserLoginInfo(ctx context.Context, userID string, driverID string, data *entity.UserLogin, expiration time.Duration) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	if data == nil {
		return ErrCacheContentEmpty
	}

	key := GetUserLoginDriveKey(userID, driverID)

	userInfoJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}

	return u.client.Set(ctx, key, userInfoJSON, expiration).Err()
}

func (u *UserCacheRedis) GetUserLoginInfos(ctx context.Context, userID string) ([]*entity.UserLogin, error) {
	if userID == "" {
		return nil, ErrCacheKeyEmpty
	}

	iter := u.client.Scan(ctx, 0, UserLoginKey+userID+":*", 0).Iterator()

	var userInfoList []*entity.UserLogin
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := u.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Key not found, skip to the next key
			}
			return nil, err
		}

		var userInfo entity.UserLogin
		if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
			return nil, err
		}

		userInfoList = append(userInfoList, &userInfo)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return userInfoList, nil
}

func (u *UserCacheRedis) GetUserLoginInfo(ctx context.Context, userID string, driverID string) (*entity.UserLogin, error) {
	if userID == "" {
		return nil, ErrCacheKeyEmpty
	}
	key := GetUserLoginDriveKey(userID, driverID)
	data, err := u.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// Unmarshal the data into the UserInfo struct
	var userInfo entity.UserLogin
	if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (u *UserCacheRedis) GetUsersLoginInfo(ctx context.Context, userIDs []string) ([]*entity.UserLogin, error) {
	if len(userIDs) == 0 {
		return nil, ErrCacheKeyEmpty
	}

	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = GetUserLoginKey(userID)
	}

	data, err := u.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	userInfoList := make([]*entity.UserLogin, len(data))
	for i, d := range data {
		if d == nil {
			continue // Key not found
		}

		var userInfo entity.UserLogin
		if err := json.Unmarshal([]byte(d.(string)), &userInfo); err != nil {
			return nil, err
		}
		userInfoList[i] = &userInfo
	}
	return userInfoList, nil
}

func (u *UserCacheRedis) DeleteUserLoginInfo(ctx context.Context, userID string, driverID string) error {
	if userID == "" {
		return ErrCacheKeyEmpty
	}
	key := GetUserLoginDriveKey(userID, driverID)
	return u.client.Del(ctx, key).Err()
}

func (u *UserCacheRedis) DeleteAllCache(ctx context.Context) error {
	keys := make([]string, 0)
	iter := u.client.Scan(ctx, 0, UserInfoKey+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return u.client.Del(ctx, keys...).Err()
}

func (u *UserCacheRedis) GetUserInfo(ctx context.Context, userID string) (*entity.User, error) {
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

	var userInfo entity.User
	if err = json.Unmarshal([]byte(val), &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	return &userInfo, nil
}

func (u *UserCacheRedis) GetUsersInfo(ctx context.Context, userIDs []string) ([]*entity.User, error) {
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

	userInfos := make([]*entity.User, 0)
	for _, val := range vals {
		if val == nil {
			continue
		}

		var userInfo entity.User
		err = json.Unmarshal([]byte(val.(string)), &userInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
		}

		userInfos = append(userInfos, &userInfo)
	}

	return userInfos, nil
}

func (u *UserCacheRedis) SetUserInfo(ctx context.Context, userID string, data *entity.User, expiration time.Duration) error {
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

func (u *UserCacheRedis) DeleteUsersInfo(ctx context.Context, userIDs []string) error {
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
	return u.client.Del(ctx, keys...).Err()
}
