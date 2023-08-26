package anticrawler

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random() int64 {
	return rand.Int63n(100) //0-99随机
}
