package usersorter

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"sort"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// User is an interface for user data
type User interface{}

func (udlr CustomUserData) MarshalBinary() ([]byte, error) {
	// 将UserDialogListResponse对象转换为二进制数据
	data, err := json.Marshal(udlr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type Group interface{}

func (udlr CustomGroupData) MarshalBinary() ([]byte, error) {
	// 将UserDialogListResponse对象转换为二进制数据
	data, err := json.Marshal(udlr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CustomUserData Custom struct implementing the User interface
type CustomUserData struct {
	UserID         string       `json:"user_id"`
	NickName       string       `json:"nickname"`
	Email          string       `json:"email"`
	Tel            string       `json:"tel"`
	Avatar         string       `json:"avatar"`
	Signature      string       ` json:"signature"`
	Status         uint         `json:"status"`
	DialogId       uint32       `json:"dialog_id"`
	CossId         string       `json:"coss_id"`
	RelationStatus uint32       `json:"relation_status"`
	Preferences    *Preferences `json:"preferences"`
}

type Preferences struct {
	SilentNotification          uint32 `json:"silent_notification"`
	Remark                      string ` json:"remark"`
	OpenBurnAfterReading        uint32 `json:"open_burn_after_reading"`
	OpenBurnAfterReadingTimeOut int64  `json:"open_burn_after_reading_time_out"`
}

type CustomGroupData struct {
	GroupID  uint32 `json:"group_id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Status   uint   `json:"status"`
	DialogId uint32 `json:"dialog_id"`
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
func SortAndGroupUsers(data interface{}, fieldName string) map[string][]interface{} {
	groupedUsers := make(map[string][]interface{})
	keyMap := make(map[string]bool)

	if list, ok := data.([]User); ok {
		for _, v := range list {
			var name string
			remark := ""
			//_ = reflect.ValueOf(v).FieldByName("Preferences")
			preferencesField := reflect.ValueOf(v).FieldByName("Preferences")
			if preferencesField.IsValid() {
				preferencesValue := preferencesField.Interface().(*Preferences)
				remark = preferencesValue.Remark
				//// 这里可以访问 Preferences 结构体中的属性
				//fmt.Printf("SilentNotification: %d\n", preferencesValue.SilentNotification)
				//fmt.Printf("Remark: %s\n", preferencesValue.Remark)
				//fmt.Printf("OpenBurnAfterReading: %d\n", preferencesValue.OpenBurnAfterReading)
				//fmt.Printf("OpenBurnAfterReadingTimeOut: %d\n", preferencesValue.OpenBurnAfterReadingTimeOut)

			}

			if remark != "" {
				name = remark
			} else {
				_ = reflect.ValueOf(v).FieldByName(fieldName)
				fieldValue := fieldOf(v, fieldName)
				name = fmt.Sprintf("%v", fieldValue.Interface())
			}

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
	} else if list, ok := data.([]Group); ok {
		for _, v := range list {
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
	if !field.IsValid() {
		// 如果字段不存在，返回零值
		return reflect.Value{}
	}
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
