package cache

//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
//	"github.com/redis/go-redis/v9"
//	"time"
//)
//
//const (
//	UserExpireTime                      = 12 * time.Hour
//	UserEmailVerificationCodeExpireTime = 30 * time.Minute
//	UserVerificationCodeExpireTime      = 5 * time.Minute
//	UserLoginExpireTime                 = 24 * 7 * time.Hour
//	UserKeyPrefix                       = "user:"
//	UserInfoKey                         = UserKeyPrefix + "info:"
//	UserLoginKey                        = UserKeyPrefix + "login:"
//	UserVerificationCode                = UserKeyPrefix + "verification_code:"
//	UserEmailVerificationCode           = UserKeyPrefix + "email_verification_code:"
//)
//
//func GetUserInfoKey(userID string) string {
//	return UserInfoKey + userID
//}
//
//func GetUserLoginKey(userID string) string {
//	return UserLoginKey + userID
//}
//
//func GetUserLoginDriveKey(userID string, driverType string, index int) string {
//	return UserLoginKey + userID + ":" + driverType + ":" + fmt.Sprintf("%d", index)
//}
//
//func GetUserVerificationCodeKey(userID, code string) string {
//	return UserVerificationCode + userID + ":" + code
//}
//
//func GetUserEmailVerificationCodeKey(userID string) string {
//	return UserEmailVerificationCode + userID
//}
//
//type UserCache interface {
//	GetUserInfo(ctx context.Context, userID string) (*usergrpcv1.UserInfoResponse, error)
//	GetUsersInfo(ctx context.Context, userID []string) ([]*usergrpcv1.UserInfoResponse, error)
//	SetUserInfo(ctx context.Context, userID string, data *usergrpcv1.UserInfoResponse, expiration time.Duration) error
//	DeleteUsersInfo(userIDs []string) error
//	DeleteAllCache(ctx context.Context) error
//	GetUserLoginInfo(ctx context.Context, userID, driverType string, index int) (*UserInfo, error)
//	SetUserLoginInfo(ctx context.Context, userID, driverType string, index int, data *UserInfo, expiration time.Duration) error
//	GetUsersLoginInfo(ctx context.Context, userID []string) ([]*UserInfo, error)
//	DeleteUserLoginInfo(ctx context.Context, userID string, driverType string, index int) error
//	GetUserLoginInfos(ctx context.Context, userID string) ([]*UserInfo, error)
//	GetUserEmailVerificationCode(ctx context.Context, userID string) (string, error)
//	SetUserEmailVerificationCode(ctx context.Context, userID, code string, expiration time.Duration) error
//	DeleteUserEmailVerificationCode(ctx context.Context, userID string) error
//
//	GetUserVerificationCode(ctx context.Context, userID, code string) (string, error)
//	SetUserVerificationCode(ctx context.Context, userID, code string, expiration time.Duration) error
//	DeleteUserVerificationCode(ctx context.Context, userID, code string) error
//}
//
//var _ UserCache = &UserCacheRedis{}
//
//func NewUserCacheRedis(addr, password string, db int) (*UserCacheRedis, error) {
//	client := redis.NewClient(&redis.Options{
//		Addr:     addr,
//		Password: password,
//		DB:       db,
//	})
//
//	_, err := client.Ping(context.Background()).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	return &UserCacheRedis{
//		client: client,
//	}, nil
//}
//
//func NewUserCacheRedisWithClient(client *redis.Client) *UserCacheRedis {
//	return &UserCacheRedis{
//		client: client,
//	}
//}
//
//type UserCacheRedis struct {
//	client *redis.Client
//}
//
//func (u *UserCacheRedis) GetUserVerificationCode(ctx context.Context, userID, code string) (string, error) {
//	if userID == "" {
//		return "", ErrCacheKeyEmpty
//	}
//	key := GetUserVerificationCodeKey(userID, code)
//	code, err := u.client.Get(ctx, key).Result()
//	if err != nil {
//		return "", err
//	}
//	return code, nil
//}
//
//func (u *UserCacheRedis) SetUserVerificationCode(ctx context.Context, userID string, code string, expiration time.Duration) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	key := GetUserVerificationCodeKey(userID, code)
//	return u.client.Set(ctx, key, code, expiration).Err()
//}
//
//func (u *UserCacheRedis) DeleteUserVerificationCode(ctx context.Context, userID, code string) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	key := GetUserVerificationCodeKey(userID, code)
//	return u.client.Del(ctx, key).Err()
//}
//
//func (u *UserCacheRedis) GetUserInfos(ctx context.Context, userID string) ([]*UserInfo, error) {
//	key := UserInfoKey + userID
//	infoStrings, err := u.client.LRange(ctx, key, 0, -1).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	userInfos := make([]*UserInfo, 0)
//	for _, infoString := range infoStrings {
//		var userInfo UserInfo
//		if err := json.Unmarshal([]byte(infoString), &userInfo); err != nil {
//			return nil, err
//		}
//		userInfos = append(userInfos, &userInfo)
//	}
//
//	return userInfos, nil
//}
//
//func (u *UserCacheRedis) GetUserEmailVerificationCode(ctx context.Context, userID string) (string, error) {
//	if userID == "" {
//		return "", ErrCacheKeyEmpty
//	}
//	key := GetUserEmailVerificationCodeKey(userID)
//	code, err := u.client.Get(ctx, key).Result()
//	if err != nil {
//		return "", err
//	}
//	return code, nil
//}
//
//func (u *UserCacheRedis) SetUserEmailVerificationCode(ctx context.Context, userID string, code string, expiration time.Duration) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	key := GetUserEmailVerificationCodeKey(userID)
//	return u.client.Set(ctx, key, code, expiration).Err()
//}
//
//func (u *UserCacheRedis) DeleteUserEmailVerificationCode(ctx context.Context, userID string) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	key := GetUserEmailVerificationCodeKey(userID)
//	return u.client.Del(ctx, key).Err()
//}
//
//func (u *UserCacheRedis) SetUserLoginInfo(ctx context.Context, userID string, driverType string, index int, data *UserInfo, expiration time.Duration) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	if data == nil {
//		return ErrCacheContentEmpty
//	}
//
//	key := GetUserLoginDriveKey(userID, driverType, index)
//	userInfoJSON, err := json.Marshal(data)
//	if err != nil {
//		return fmt.Errorf("failed to marshal user info: %v", err)
//	}
//
//	return u.client.Set(ctx, key, userInfoJSON, expiration).Err()
//}
//
//func (u *UserCacheRedis) GetUserLoginInfos(ctx context.Context, userID string) ([]*UserInfo, error) {
//	if userID == "" {
//		return nil, ErrCacheKeyEmpty
//	}
//
//	iter := u.client.Scan(ctx, 0, UserLoginKey+userID+":*", 0).Iterator()
//
//	var userInfoList []*UserInfo
//	for iter.Next(ctx) {
//		key := iter.Val()
//		data, err := u.client.Get(ctx, key).Result()
//		if err != nil {
//			if err == redis.Nil {
//				continue // Key not found, skip to the next key
//			}
//			return nil, err
//		}
//
//		var userInfo UserInfo
//		if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
//			return nil, err
//		}
//
//		userInfoList = append(userInfoList, &userInfo)
//	}
//
//	if err := iter.Err(); err != nil {
//		return nil, err
//	}
//
//	return userInfoList, nil
//}
//
//func (u *UserCacheRedis) GetUserLoginInfo(ctx context.Context, userID string, driverType string, index int) (*UserInfo, error) {
//	if userID == "" {
//		return nil, ErrCacheKeyEmpty
//	}
//	key := GetUserLoginDriveKey(userID, driverType, index)
//	data, err := u.client.Get(ctx, key).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	// Unmarshal the data into the UserInfo struct
//	var userInfo UserInfo
//	if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
//		return nil, err
//	}
//
//	return &userInfo, nil
//}
//
//func (u *UserCacheRedis) GetUsersLoginInfo(ctx context.Context, userIDs []string) ([]*UserInfo, error) {
//	if len(userIDs) == 0 {
//		return nil, ErrCacheKeyEmpty
//	}
//
//	keys := make([]string, len(userIDs))
//	for i, userID := range userIDs {
//		keys[i] = GetUserLoginKey(userID)
//	}
//
//	data, err := u.client.MGet(ctx, keys...).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	userInfoList := make([]*UserInfo, len(data))
//	for i, d := range data {
//		if d == nil {
//			continue // Key not found
//		}
//
//		var userInfo UserInfo
//		if err := json.Unmarshal([]byte(d.(string)), &userInfo); err != nil {
//			return nil, err
//		}
//		userInfoList[i] = &userInfo
//	}
//	return userInfoList, nil
//}
//
//func (u *UserCacheRedis) DeleteUserLoginInfo(ctx context.Context, userID string, driverType string, index int) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	key := GetUserLoginDriveKey(userID, driverType, index)
//	return u.client.Del(ctx, key).Err()
//}
//
//func (u *UserCacheRedis) DeleteAllCache(ctx context.Context) error {
//	keys := make([]string, 0)
//	iter := u.client.Scan(ctx, 0, UserKeyPrefix+"*", 0).Iterator()
//	for iter.Next(ctx) {
//		keys = append(keys, iter.Val())
//	}
//	if err := iter.Err(); err != nil {
//		return err
//	}
//	if len(keys) == 0 {
//		return nil
//	}
//	return u.client.Del(ctx, keys...).Err()
//}
//
//func (u *UserCacheRedis) GetUserInfo(ctx context.Context, userID string) (*usergrpcv1.UserInfoResponse, error) {
//	if userID == "" {
//		return nil, ErrCacheKeyEmpty
//	}
//
//	key := GetUserInfoKey(userID)
//	val, err := u.client.Get(ctx, key).Result()
//	if err == redis.Nil {
//		return nil, nil
//	} else if err != nil {
//		return nil, fmt.Errorf("failed to get user info from cache: %v", err)
//	}
//
//	var userInfo usergrpcv1.UserInfoResponse
//	if err = json.Unmarshal([]byte(val), &userInfo); err != nil {
//		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
//	}
//
//	return &userInfo, nil
//}
//
//func (u *UserCacheRedis) GetUsersInfo(ctx context.Context, userIDs []string) ([]*usergrpcv1.UserInfoResponse, error) {
//	if len(userIDs) == 0 {
//		return nil, ErrCacheKeyEmpty
//	}
//
//	keys := make([]string, len(userIDs))
//	for i, userID := range userIDs {
//		if userID == "" {
//			return nil, ErrCacheKeyEmpty
//		}
//		keys[i] = GetUserInfoKey(userID)
//	}
//
//	vals, err := u.client.MGet(ctx, keys...).Result()
//	if err != nil {
//		return nil, fmt.Errorf("failed to get users info from cache: %v", err)
//	}
//
//	userInfos := make([]*usergrpcv1.UserInfoResponse, len(userIDs))
//	for i, val := range vals {
//		if val == nil {
//			continue
//		}
//
//		var userInfo usergrpcv1.UserInfoResponse
//		err = json.Unmarshal([]byte(val.(string)), &userInfo)
//		if err != nil {
//			return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
//		}
//
//		userInfos[i] = &userInfo
//	}
//
//	return userInfos, nil
//}
//
//func (u *UserCacheRedis) SetUserInfo(ctx context.Context, userID string, data *usergrpcv1.UserInfoResponse, expiration time.Duration) error {
//	if userID == "" {
//		return ErrCacheKeyEmpty
//	}
//	if data == nil {
//		return ErrCacheContentEmpty
//	}
//
//	key := GetUserInfoKey(userID)
//	userInfoJSON, err := json.Marshal(data)
//	if err != nil {
//		return fmt.Errorf("failed to marshal user info: %v", err)
//	}
//
//	return u.client.Set(ctx, key, userInfoJSON, expiration).Err()
//}
//
//func (u *UserCacheRedis) DeleteUsersInfo(userIDs []string) error {
//	if len(userIDs) == 0 {
//		return ErrCacheKeyEmpty
//	}
//	keys := make([]string, len(userIDs))
//	for i, userID := range userIDs {
//		if userID == "" {
//			return ErrCacheKeyEmpty
//		}
//		keys[i] = GetUserInfoKey(userID)
//	}
//	return u.client.Del(context.Background(), keys...).Err()
//}
