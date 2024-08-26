package main

import (
	"testing"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

func BenchmarkRandomString(b *testing.B) {
	tests := []int{7, 14, 28, 56, 112}

	for _, n := range tests {
		b.Run("Length_"+string(rune(n)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				helpers.RandomString(n)
			}
		})
	}
}
