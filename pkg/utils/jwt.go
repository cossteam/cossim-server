package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtKey = []byte("a_secret_create")

const ExpirationTime = 30

type Claims struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成token
func GenerateToken(userId, email string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			UserId:           userId,
			Email:            email,
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
