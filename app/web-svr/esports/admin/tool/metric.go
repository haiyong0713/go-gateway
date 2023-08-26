package tool

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	StatusOfSucceed = "succeed"
	StatusOfFailed  = "failed"
)

var (
	metric4ClearCacheCount *prometheus.CounterVec
)

func init() {
	metric4ClearCacheCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports_admin",
			Name:      "clear_cache_count_stats",
			Help:      "esports clear cache count stats",
		},
		[]string{"biz_name", "status"})

	prometheus.MustRegister(
		metric4ClearCacheCount)
}

func AddClearCacheMetric(bizName, status string) {
	metric4ClearCacheCount.WithLabelValues([]string{bizName, status}...).Inc()
}
