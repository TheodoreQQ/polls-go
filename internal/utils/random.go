package utils

import "math/rand"

func GenerateCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRTSUWXYZ123456789"
	code := make([]byte, 4)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)

}
