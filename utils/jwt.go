package utils

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("")

type customClaims struct {
	UserId int `json:"userId"`
	jwt.StandardClaims
}

type expiryOptions func(expiry *time.Duration)

func WithExpiry(expiry time.Duration) func(time *time.Duration) {
	return func(time *time.Duration) {
		*time = expiry
	}
}

func GenerateToken(userId uint, opts ...expiryOptions) (string, error) {
	expiry := time.Hour * 24 * 7
	for _, opt := range opts {
		opt(&expiry)
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userId,
		"iat":    now.Unix(),
		"exp":    now.Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) int {
	tokenString := strings.TrimPrefix(tokenStr, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return 0
	}
	// 断言 Claims 类型并校验 token 有效性
	claims, ok := token.Claims.(*customClaims)
	if !ok {
		// token 错误
		return -2
	}
	if !token.Valid {
		// 过期
		return -1
	}
	return claims.UserId
}