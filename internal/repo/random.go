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

// RandomInt returns random int equal or less than max
func RandomInt(max int) int {
	return rand.Intn(max)
}

// RandomString returns random string witg length == n
func RandomString(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := rand.Intn(len(alphabet))
		sb.WriteByte(alphabet[c])
	}
	return sb.String()
}
