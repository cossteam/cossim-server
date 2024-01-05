// usersorter.go
package usersorter

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// User is an interface for user data
type User interface{}

// SortAndGroupUsers sorts the user data based on a specified field and groups them by the first letter of the field values or special characters
func SortAndGroupUsers(data []User, fieldName string) map[string][]User {
	groupedUsers := make(map[string][]User)
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
	//
	//for _, char := range keys {
	//	fmt.Printf("%s: [\n", char)
	//	for _, user := range groupedUsers[char] {
	//		// Custom implementation for printing user data or further processing
	//		fmt.Printf("\t%+v\n", user)
	//	}
	//	fmt.Printf("],\n")
	//}

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
