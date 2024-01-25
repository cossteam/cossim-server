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

func SetKey(client *redis.Client, key string, data map[string]string) error {
	err := client.HMSet(context.Background(), key, data).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetKey(client *redis.Client, key string) (map[string]string, error) {
	data, err := client.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DeleteKey(client *redis.Client, userID string) error {
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
func GetUserInfoList(data []string) ([]UserInfo, error) {
	list := make([]UserInfo, 0)
	for _, datum := range data {
		var user UserInfo
		err := json.Unmarshal([]byte(datum), &user)
		if err != nil {
			fmt.Println("GetUserInfoList JSON unmarshal error:", err)
			return nil, err
		}
		list = append(list, user)
	}
	return list, nil
}

// 用户信息转成[]interfaces{}
func GetUserInfoListToInterfaces(data []UserInfo) []interface{} {
	list := make([]interface{}, len(data))
	for i, datum := range data {
		list[i] = datum
	}
	return list
}

// 根据客户端类型分类用户登录信息列表
func CategorizeByDriveType(data []UserInfo) map[string][]UserInfo {
	result := make(map[string][]UserInfo)

	for _, userInfo := range data {
		result[userInfo.DriverType] = append(result[userInfo.DriverType], userInfo)
	}

	return result
}
