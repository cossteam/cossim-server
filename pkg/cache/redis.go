package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
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

func SetHMapKey(client *redis.Client, key string, data map[string]string) error {
	err := client.HMSet(context.Background(), key, data).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetHMapKey(client *redis.Client, key string) (map[string]string, error) {
	data, err := client.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DeleteHMapKey(client *redis.Client, userID string) error {
	err := client.HDel(context.Background(), userID).Err()
	if err != nil {
		return err
	}
	return nil
}

func DeleteKeyField(client *redis.Client, key string, field string) error {
	err := client.HDel(context.Background(), key, field).Err()
	if err != nil {
		return err
	}
	return nil
}

func AddToList(client *redis.Client, key string, values []interface{}) error {
	err := client.RPush(context.Background(), key, values).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetList 获取 Redis List 中指定范围的元素
func GetList(client *redis.Client, key string, start int64, stop int64) ([]string, error) {
	data, err := client.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteList 删除 Redis 中的 List
func DeleteList(client *redis.Client, key string) error {
	err := client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetAllListValues 获取指定键的所有值
func GetAllListValues(client *redis.Client, key string) ([]string, error) {
	values, err := client.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return values, nil
}

// RemoveFromList 从 Redis List 中删除指定元素
func RemoveFromList(client *redis.Client, key string, count int64, value interface{}) error {
	err := client.LRem(context.Background(), key, count, value).Err()
	if err != nil {
		return err
	}
	return nil
}

// 设置 Redis 键的过期时间（以秒为单位）
func SetKeyExpiration(client *redis.Client, key string, expiration time.Duration) error {
	err := client.Expire(context.Background(), key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// 设置 Redis 键在指定时间点过期
func SetKeyExpirationAt(client *redis.Client, key string, expiration time.Time) error {
	err := client.ExpireAt(context.Background(), key, expiration).Err()
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

func SetKey(client *redis.Client, key string, data interface{}, expiration time.Duration) error {
	err := client.Set(context.Background(), key, data, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetKey(client *redis.Client, key string) (interface{}, error) {
	data, err := client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DelKey(client *redis.Client, key string) error {
	err := client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

// UpdateKeyExpiration 更新键的过期时间
func UpdateKeyExpiration(client *redis.Client, key string, expiration time.Duration) error {
	remaining, err := client.TTL(context.Background(), key).Result()
	if err != nil {
		return err
	}
	fmt.Println("remaining => ", remaining)
	// 如果键不存在或已过期，返回错误
	if remaining < 0 && remaining != -1 {
		return fmt.Errorf("key does not exist or has expired")
	}
	// 使用 EXPIRE 命令更新键的过期时间
	return client.Expire(context.Background(), key, expiration).Err()
}

func ScanKeys(client *redis.Client, pattern string) ([]string, error) {
	keys, err := client.Keys(context.Background(), pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func ExistsKey(client *redis.Client, key string) (bool, error) {
	exists, err := client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	found := false
	if exists == 1 {
		found = true
	}
	return found, nil
}
