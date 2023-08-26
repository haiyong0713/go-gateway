package tool

import (
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type BreakerSetting struct {
	Enabled                bool
	Codes                  map[string]bool
	Name                   string
	MaxRequests            uint32
	InternalSeconds        int64
	TimeoutSeconds         int64
	MaxConsecutiveFailures uint32
	MaxRateToOpen          float64
}

var (
	cbCodes       sync.Map
	cbs           sync.Map
	lastCbKeyList map[string]int
)

func init() {
	lastCbKeyList = make(map[string]int, 0)
}

func LoadCbByBizKey(key string) (*gobreaker.CircuitBreaker, bool) {
	if d, ok := cbs.Load(key); ok {
		if cb, ok := d.(*gobreaker.CircuitBreaker); ok {
			return cb, ok
		}
	}

	return nil, false
}

func RestoreCbsByCfg(m map[string]BreakerSetting) {
	for k, cfg := range m {
		if cfg.Enabled {
			cb := newCbByCfg(cfg)
			cbs.Store(k, cb)
			cbCodes.Store(k, cfg.Codes)
		} else {
			cbs.Delete(k)
			cbCodes.Delete(k)
		}
	}

	for k := range lastCbKeyList {
		if _, ok := m[k]; !ok {
			cbs.Delete(k)
			cbCodes.Delete(k)
		}
	}

	lastLimiterKeyList = make(map[string]int64, 0)
	for k := range m {
		lastCbKeyList[k] = 1
	}
}

func CanBreakerByCode(key string, code string) bool {
	if d, ok := cbCodes.Load(key); ok {
		if m, ok := d.(map[string]bool); ok {
			if can, ok := m[code]; ok && can {
				return true
			}
		}
	}

	return false
}

func newCbByCfg(cfg BreakerSetting) *gobreaker.CircuitBreaker {
	var st gobreaker.Settings
	{
		st.Name = cfg.Name
		st.MaxRequests = cfg.MaxRequests
		st.Interval = time.Duration(cfg.InternalSeconds) * time.Second
		st.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
		st.ReadyToTrip = func(counts gobreaker.Counts) bool {
			rate4Failed := float64(counts.TotalFailures) / float64(counts.Requests)

			return rate4Failed >= cfg.MaxRateToOpen || counts.ConsecutiveFailures > cfg.MaxConsecutiveFailures
		}
		st.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
			switch to {
			case gobreaker.StateOpen:
				metric4Breaker.WithLabelValues([]string{name}...).Inc()
			}
		}
	}

	return gobreaker.NewCircuitBreaker(st)
}
