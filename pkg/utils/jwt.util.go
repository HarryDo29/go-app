package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SecretKey struct {
	Key        string
	ExpireTime int
}

type UserInfo struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"roles"`
}

type MyCustomClaims struct {
	UserInfo `json:"user_info"`
	jwt.RegisteredClaims
}

func GenerateJWT(userInfo UserInfo, option SecretKey) (string, error) {
	claims := MyCustomClaims{
		UserInfo: userInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(option.ExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "my_go_backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(option.Key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string, option SecretKey) (*MyCustomClaims, error) {
	claims := MyCustomClaims{}

	token, err := jwt.ParseWithClaims( // ParseWithClaims là hàm để xác thực và parse token
		tokenString, // token từ client
		&claims,     // biến nhận payload từ token nếu như token hợp lệ
		func(token *jwt.Token) (interface{}, error) {
			// kiểm tra thuật toán ký của token có phải nhóm HMAC không, ví dụ HS256/HS384/HS512
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Hashing algorithm is not valid: %v", token.Header["alg"])
			}
			// dùng secret key của hệ thống để kiểm tra signature trong token
			return []byte(option.Key), nil // trả về secret key dưới dạng byte slice
		}) // keyFunc: hàm cung cấp secret key cho JWT library để verify token

	// nếu token bị lỗi (ví dụ: hết hạn, sai signature)
	if err != nil {
		return nil, err
	}
	// nếu signature của token bị lỗi
	if !token.Valid {
		return nil, fmt.Errorf("Token is not valid")
	}

	return &claims, nil
}
