package usersorter

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"sort"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// User is an interface for user data
type User interface{}

// CustomUserData Custom struct implementing the User interface
type CustomUserData struct {
	UserID    string `json:"user_id"`
	NickName  string `json:"nick_name"`
	Email     string `json:"email"`
	Tel       string `json:"tel"`
	Avatar    string `json:"avatar"`
	Signature string ` json:"signature"`
	Status    uint   `json:"status"`
}

func ConvertToGinH(data map[string][]interface{}) gin.H {
	result := make(gin.H)
	for key, value := range data {
		var users []CustomUserData
		for _, user := range value {
			if customUser, ok := user.(CustomUserData); ok {
				users = append(users, customUser)
			}
		}
		result[key] = users
	}
	return result
}

// SortAndGroupUsers sorts the user data based on a specified field and groups them by the first letter of the field values or special characters
func SortAndGroupUsers(data []User, fieldName string) map[string][]interface{} {
	groupedUsers := make(map[string][]interface{})
	keyMap := make(map[string]bool)

	for _, v := range data {
		_ = reflect.ValueOf(v).FieldByName(fieldName)
		fieldValue := fieldOf(v, fieldName)
		name := fmt.Sprintf("%v", fieldValue.Interface())

		if isChinese(name) {
			pinyinSlice := pinyin.Pinyin(name, pinyin.NewArgs())
			firstChar := getFirstChar(pinyinSlice)
			name = strings.ToUpper(firstChar)
		} else if isSpecialChar(name) {
			name = "#"
		}

		k := strings.ToUpper(name[:1])
		groupedUsers[k] = append(groupedUsers[k], v)
		keyMap[k] = true
	}

	var keys []string
	for k := range keyMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return groupedUsers
}

func fieldOf(i interface{}, name string) reflect.Value {
	val := reflect.ValueOf(i)
	field := reflect.Indirect(val).FieldByName(name)
	return field
}

// Rest of the functions (isSpecialChar, isChinese, getFirstChar) remain the same

func isSpecialChar(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return true
		}
	}
	return false
}

func isChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

func getFirstChar(pinyinSlice [][]string) string {
	var result strings.Builder

	for _, pinyinWord := range pinyinSlice {
		if len(pinyinWord) > 0 {
			result.WriteString(pinyinWord[0][:1])
		}
	}

	return result.String()
}
