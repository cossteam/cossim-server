package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type UserInfo struct {
	ID         uint   `json:"id"`
	UserId     string `json:"user_id"`
	Token      string `json:"token"`
	DriverType string `json:"driver_type"`
	CreateAt   int64  `json:"create_at"`
	ClientIP   string `json:"client_ip"`
	Rid        int64  `json:"rid"`
}

func (i UserInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type RedisClient struct {
	Client *redis.Client
	lock   *sync.Mutex
}

func NewRedisClient(address string, password string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s", address),
		Password: password,
		DB:       0,
	})
	return &RedisClient{Client: client, lock: &sync.Mutex{}}
}

func (r *RedisClient) SetHMapKey(key string, data map[string]string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.HMSet(context.Background(), key, data).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) GetHMapKey(key string) (map[string]string, error) {
	data, err := r.Client.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *RedisClient) DeleteHMapKey(userID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.HDel(context.Background(), userID).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) DeleteKeyField(key string, field string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.HDel(context.Background(), key, field).Err()
	if err != nil {
		return err
	}
	return nil
}

// 添加到 List右边
func (r *RedisClient) AddToList(key string, values []interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.RPush(context.Background(), key, values...).Err()
	if err != nil {
		return err
	}
	return nil
}

// 添加到List左边
func (r *RedisClient) AddToListLeft(key string, values []interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.LPush(context.Background(), key, values).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetList 获取 Redis List 中指定范围的元素
func (r *RedisClient) GetList(key string, start int64, stop int64) ([]string, error) {
	data, err := r.Client.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteList 删除 Redis 中的 List
func (r *RedisClient) DeleteList(key string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetAllListValues 获取指定键的所有值
func (r *RedisClient) GetAllListValues(key string) ([]string, error) {
	values, err := r.Client.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return values, nil
}

// RemoveFromList 从 Redis List 中删除指定元素
func (r *RedisClient) RemoveFromList(key string, count int64, value interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.LRem(context.Background(), key, count, value).Err()
	if err != nil {
		return err
	}
	return nil
}

// 设置 Redis 键的过期时间（以秒为单位）
func (r *RedisClient) SetKeyExpiration(key string, expiration time.Duration) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.Expire(context.Background(), key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// 设置 Redis 键在指定时间点过期
func (r *RedisClient) SetKeyExpirationAt(key string, expiration time.Time) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.ExpireAt(context.Background(), key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// 解析用户登录信息列表
func GetUserInfoList(data []string) ([]*UserInfo, error) {
	list := make([]*UserInfo, 0)
	for _, datum := range data {
		var user *UserInfo
		err := json.Unmarshal([]byte(datum), &user)
		if err != nil {
			fmt.Println("GetUserInfoList JSON unmarshal error:", err)
			return nil, err
		}
		list = append(list, user)
	}
	return list, nil
}

func GetUserInfo(data string) (*UserInfo, error) {
	var user *UserInfo
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		fmt.Println("GetUserInfoList JSON unmarshal error:", err)
		return nil, err
	}
	return user, nil
}

// 用户信息转成[]interfaces{}
func GetUserInfoListToInterfaces(data []*UserInfo) []interface{} {
	list := make([]interface{}, 0)
	for _, datum := range data {
		list = append(list, datum)
	}
	return list
}

func GetUserInfoToInterfaces(data *UserInfo) interface{} {
	return data
}

// 根据客户端类型分类用户登录信息列表
func CategorizeByDriveType(data []*UserInfo) map[string][]*UserInfo {
	result := make(map[string][]*UserInfo)

	for _, userInfo := range data {
		result[userInfo.DriverType] = append(result[userInfo.DriverType], userInfo)
	}

	return result
}

func (r *RedisClient) SetKey(key string, data interface{}, expiration time.Duration) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.Set(context.Background(), key, data, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) GetKey(key string) (interface{}, error) {
	data, err := r.Client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *RedisClient) DelKey(key string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	err := r.Client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

// UpdateKeyExpiration 更新键的过期时间
func (r *RedisClient) UpdateKeyExpiration(key string, expiration time.Duration) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	remaining, err := r.Client.TTL(context.Background(), key).Result()
	if err != nil {
		return err
	}
	// 如果键不存在或已过期，返回错误
	if remaining < 0 && remaining != -1 {
		return fmt.Errorf("key does not exist or has expired")
	}
	// 使用 EXPIRE 命令更新键的过期时间
	return r.Client.Expire(context.Background(), key, expiration).Err()
}

func (r *RedisClient) ScanKeys(pattern string) ([]string, error) {
	keys, err := r.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *RedisClient) ExistsKey(key string) (bool, error) {
	exists, err := r.Client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	found := false
	if exists == 1 {
		found = true
	}
	return found, nil
}
