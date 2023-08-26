package tool

import (
	"math/rand"
	"time"
)

const (
	seconds4TenHours = 36000
	seconds4Day      = 86400
)

func CalculateExpiredSeconds(delayDay int64) int64 {
	now := time.Now()
	year, month, day := now.Date()
	nextDay := time.Date(year, month, day, 24, 0, 0, 0, now.Location()).Unix()

	rand.Seed(time.Now().UnixNano())
	randSeconds := rand.Int63n(seconds4TenHours)

	return nextDay + randSeconds - now.Unix() + delayDay*seconds4Day
}
