package repo

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(max int) int {
	return rand.Intn(max)
}

func RandomString(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := rand.Intn(len(alphabet))
		sb.WriteByte(alphabet[c])
	}
	return sb.String()
}
