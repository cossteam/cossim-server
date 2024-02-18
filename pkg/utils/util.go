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

// 求差集
func SliceDifference(slice1, slice2 []uint32) []uint32 {
	var diff []uint32
	set := make(map[uint32]struct{})

	// 将slice2中的元素存入一个集合
	for _, num := range slice2 {
		set[num] = struct{}{}
	}

	// 遍历slice1，如果元素不在集合中，则加入差集
	for _, num := range slice1 {
		if _, ok := set[num]; !ok {
			diff = append(diff, num)
		}
	}

	return diff
}
