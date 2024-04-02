package cache

import (
	"context"
	"time"
)

// Cache 接口定义了一个通用的缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)       // 根据键获取缓存值
	Set(ctx context.Context, key string, value interface{})         // 设置缓存值
	Delete(ctx context.Context, key string)                         // 删除缓存值
	Exists(ctx context.Context, key string) bool                    // 检查缓存值是否存在
	Expire(ctx context.Context, key string, duration time.Duration) // 设置缓存过期时间
}
