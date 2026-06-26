package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword nhận vào mật khẩu plain-text và trả về chuỗi đã được hash mã hóa
func HashPassword(password string) (string, error) {
	// Chuyển password từ string sang []byte
	// bcrypt.DefaultCost hiện tại có giá trị là 10 (mức độ băm cân bằng giữa hiệu năng và bảo mật)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Chuyển kết quả []byte ngược lại thành string để lưu vào Database
	return string(hashedBytes), nil
}

// VerifyPassword so sánh mật khẩu plain-text do user nhập vào với chuỗi hash lấy từ DB lên
func VerifyPassword(hashedPassword, password string) bool {
	// Hàm này trả về nil nếu hai mật khẩu khớp nhau
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
