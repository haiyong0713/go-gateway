package steins

import (
	"time"
)

const RetrySpan = 50 * time.Millisecond

func Retry(callback func() error, retry int, sleep time.Duration) (err error) {
	for i := 0; i < retry; i++ {
		if err = callback(); err == nil {
			return
		}
		time.Sleep(sleep)
	}
	return

}
