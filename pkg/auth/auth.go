package auth

import (
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Claims struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthenticator(db *gorm.DB, rdb *cache.RedisClient) *Authenticator {
	return &Authenticator{db, rdb}
}

type Authenticator struct {
	DB  *gorm.DB
	RDB *cache.RedisClient
}

const _queryUser = "SELECT * FROM users WHERE id = ?"
const _queryAdmin = "SELECT * FROM admins WHERE user_id = ?"

func (a *Authenticator) ValidateToken(tokenString string, driverType string) (bool, error) {
	token, claims, err := utils.ParseToken(tokenString)
	if err != nil || !token.Valid {
		return false, fmt.Errorf("token validation failed: %s", err.Error())
	}
	type User struct {
		ID     string `json:"id"`
		Status int64  `json:"status"`
	}

	keys, err := a.RDB.ScanKeys(claims.UserId + ":" + driverType + ":*")
	if err != nil {
		fmt.Println("error => ", err)
		return false, err
	}
	if len(keys) <= 0 {
		return false, errors.New("token not found")
	}

	var found = false

	for _, key := range keys {
		v, err := a.RDB.GetKey(key)
		if err != nil {
			return false, err
		}
		data := v.(string)
		info, err := cache.GetUserInfo(data)
		if err != nil {
			return false, err
		}
		if info.Token == tokenString {
			found = true
		}
	}

	if !found {
		return false, errors.New("token not found")
	}
	var user User
	if err = a.DB.Raw(_queryUser, claims.UserId).Scan(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, fmt.Errorf("error retrieving user: %s", err.Error())
	}

	if user.Status != 1 {
		return false, errors.New("user status is abnormal")
	}

	return true, nil
}

func (a *Authenticator) ValidateAdminToken(tokenString string) (bool, error) {
	token, claims, err := utils.ParseToken(tokenString)
	if err != nil || !token.Valid {
		return false, fmt.Errorf("token validation failed: %s", err.Error())
	}
	type User struct {
		ID     string `json:"id"`
		Status int64  `json:"status"`
	}

	var admin User
	if err = a.DB.Raw(_queryAdmin, claims.UserId).Scan(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("not admin")
		}
		return false, fmt.Errorf("error retrieving admin: %s", err.Error())
	}
	if admin.Status != 1 {
		return false, errors.New("admin status is abnormal")
	}

	return true, nil
}
