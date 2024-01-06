package utils

import (
	"github.com/google/uuid"
)

// 生成uuid
func GenUUid() string {
	newUUID := uuid.New()
	// 将 UUID 转换为字符串形式
	return newUUID.String()
}
