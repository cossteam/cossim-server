package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func ParseDurationFromString(durationString string) (time.Duration, error) {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func HashString(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	hashedBytes := hasher.Sum(nil)
	hashedString := hex.EncodeToString(hashedBytes)
	return hashedString
}

// 生成6位随机数字
func RandomNum() string {
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	return code
}
