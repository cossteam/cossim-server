package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sony/sonyflake"
	"golang.org/x/net/html"
	"math/rand"
	"strings"
	"time"
)

// ExtractText 从HTML中提取文本内容
func ExtractText(htmlString string) (string, error) {
	// 使用 html.Parse 解析 HTML
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	// 递归遍历 HTML 结构并提取文本内容
	var textContent string
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			textContent += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(doc)

	return textContent, nil
}

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

func GenCossID() (string, error) {

	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("coss_id_%x", id), nil
}

// 将结构体转换为字节数组的方法
func StructToBytes(data interface{}) ([]byte, error) {
	// 使用json.Marshal函数将结构体编码为JSON格式
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// 将字节数组转换回结构体的方法
func BytesToStruct(data []byte, out interface{}) error {
	// 使用json.Unmarshal函数将JSON格式的字节数组解码为结构体
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}
