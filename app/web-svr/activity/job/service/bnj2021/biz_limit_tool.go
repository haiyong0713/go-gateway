package bnj2021

import (
	"sync"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	limitKeyOfAwardPay = "bnj_award_pay"
	limitKeyOfDraw     = "bnj_draw"
)

var (
	bizLimiterMap sync.Map

	metric4Limiter *prometheus.CounterVec
)

func init() {
	metric4Limiter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity_job",
			Name:      "biz_limiter_stats",
			Help:      "activity biz limiter calling stats",
		},
		[]string{"biz_name"})
	prometheus.MustRegister(metric4Limiter)
}

func resetLimiters(cfg map[string]int64) {
	for k, v := range cfg {
		if d, ok := BizLimitRule[k]; ok {
			if v != d {
				bizLimiterMap.Store(k, tollbooth.NewLimiter(float64(v), nil))
			}
		} else {
			bizLimiterMap.Store(k, tollbooth.NewLimiter(float64(v), nil))
		}
	}

	for k := range BizLimitRule {
		if _, ok := cfg[k]; !ok {
			bizLimiterMap.Delete(k)
		}
	}
}

func isBizLimitReachedByBiz(limitKey interface{}, bizName string) (reached bool) {
	if d, ok := bizLimiterMap.Load(limitKey); ok {
		if bizLimiter, ok := d.(*limiter.Limiter); ok {
			reached = bizLimiter.LimitReached(bizName)
			if reached {
				metric4Limiter.WithLabelValues([]string{bizName}...).Inc()
			}

			return
		}
	}

	return
}

func waitBizLimit(limitKey interface{}, bizName string) {
	for {
		if !isBizLimitReachedByBiz(limitKey, bizName) {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func BnjBizLimitRule() map[string]int64 {
	return BizLimitRule
}
