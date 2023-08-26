package util

import (
	"time"

	"go-common/library/net/netutil"
)

func WithAttempts(attempts int, backOff netutil.BackoffConfig, f func() (interface{}, error)) (interface{}, error) {
	var retry int
	for {
		rly, err := f()
		if err == nil {
			return rly, nil
		}
		if retry++; retry < attempts {
			time.Sleep(backOff.Backoff(retry))
			continue
		}
		return nil, err
	}
}
