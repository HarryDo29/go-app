package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateUniqueOTP() (string, error) {
	digits := []byte("0123456789")

	for i := len(digits) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}

		digits[i], digits[j.Int64()] =
			digits[j.Int64()], digits[i]
	}

	return string(digits[:6]), nil
}
