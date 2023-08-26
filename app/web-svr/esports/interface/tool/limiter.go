package tool

import (
	"strconv"
	"sync"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
)

const (
	BizLimitKey4DBRestoreOfLow = "db_restore_low"
)

var (
	lastLimiterKeyList map[string]int64
	limiters           sync.Map
)

func init() {
	lastLimiterKeyList = make(map[string]int64, 0)
}

func RestoreLimiters(m map[string]int64) {
	for k, v := range m {
		if d, ok := lastLimiterKeyList[k]; !ok || v != d {
			f, _ := strconv.ParseFloat(strconv.FormatInt(v, 10), 64)
			lmt := tollbooth.NewLimiter(f, nil)
			limiters.Store(k, lmt)
		}
	}

	for k := range lastLimiterKeyList {
		if _, ok := m[k]; !ok {
			limiters.Delete(k)
		}
	}

	lastLimiterKeyList = make(map[string]int64, 0)
	for k, v := range m {
		lastLimiterKeyList[k] = v
	}
}

func IsLimiterAllowedByUniqBizKey(lmtKey, bizKey string) bool {
	if lmt, ok := limiters.Load(lmtKey); ok {
		if d, ok := lmt.(*limiter.Limiter); ok {
			allow := !d.LimitReached(bizKey)
			if !allow {
				metric4Limiter.WithLabelValues([]string{bizKey}...).Inc()
			}

			return allow
		}
	}

	return true
}
