package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtKey = []byte("a_secret_create")

// ExpirationTime token过期时间 redis的目前是7天,不确定是否要一致
const ExpirationTime = 30

type Claims struct {
	UserId    string `json:"user_id"`
	Email     string `json:"email"`
	DriverId  string `json:"driver_id"`
	PublicKey string ` json:"public_key"`
	jwt.RegisteredClaims
}

// GenerateToken 生成token
func GenerateToken(userId, email, driverId, publicKey string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			UserId:           userId,
			Email:            email,
			DriverId:         driverId,
			PublicKey:        publicKey,
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Duration(ExpirationTime) * 24 * time.Hour)}},
		})
	token, err := t.SignedString(jwtKey)
	if err != nil {
		return "", nil
	}
	return token, nil
}

// ParseToken 解析token
func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	return token, claims, err
}
