package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

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
