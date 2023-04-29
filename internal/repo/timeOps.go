package repo

import (
	"math/rand"
	"strconv"
	"time"
)

// NewTS generates pseudo-random string based on current time
func NewTS() string {
	t := time.Now()

	rand.Seed(t.UnixNano())

	randSuffixInt := rand.Intn(1000)

	randSuffixString := strconv.Itoa(randSuffixInt)

	return t.Format("02.01.2006 15_04_05") + "." + randSuffixString
}
