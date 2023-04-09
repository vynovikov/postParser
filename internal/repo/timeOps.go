package repo

import (
	"math/rand"
	"strconv"
	"time"
)

func NewTS() string {
	t := time.Now()

	rand.Seed(t.UnixNano())

	randSuffixInt := rand.Intn(1000)

	randSuffixString := strconv.Itoa(randSuffixInt)

	return t.Format("02.01.2006 15_16_17") + "." + randSuffixString
}
