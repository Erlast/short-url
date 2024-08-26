package helpers

import (
	"math/rand"
)

// LenString длина генерируемой случайной строки.
const LenString = 7

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString функция генерации случайно строки длиной n.
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
