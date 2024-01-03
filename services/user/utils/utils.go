package utils

import (
	"im/services/user/config"
	"regexp"
)

// 判断是否为email
func IsEmail(email string) bool {
	re := regexp.MustCompile(config.EmailRegex)
	return re.MatchString(email)
}

// 判断密码是否为8-20位，由大小写字母，数字，特殊字符组成
func ValidatePassword(password string) bool {
	re := regexp.MustCompile(config.PasswordRegex)
	return re.MatchString(password)
}
