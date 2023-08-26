package tool

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	CacheOfLocal  = "local"
	CacheOfRemote = "remote"

	bizOpOfDBBackSource = "db_back_source"
	bizOpOfDBErr        = "db_err"
	bizOpOfDBNoResult   = "db_no_result"
)

var (
	metric4Limiter *prometheus.CounterVec
	metric4Breaker *prometheus.CounterVec

	Metric4RpcQps     *prometheus.CounterVec
	Metric4RpcCount   *prometheus.CounterVec
	Metric4RpcLatency *prometheus.CounterVec

	Metric4MemoryCache *prometheus.CounterVec

	Metric4CacheResetFailed *prometheus.CounterVec
	metric4DBRestoreBiz     *prometheus.CounterVec
)

func init() {
	metric4DBRestoreBiz = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "db_restore_count_stats",
			Help:      "esports db restore count stats",
		},
		[]string{"biz_op", "biz_name"})

	metric4Limiter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "biz_limiter_stats",
			Help:      "esports biz limiter calling stats",
		},
		[]string{"biz_name"})

	metric4Breaker = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "rpc_breaker_stats",
			Help:      "esports rpc breaker calling stats",
		},
		[]string{"biz_name"})

	Metric4RpcQps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "rpc_calling_stats",
			Help:      "esports rpc calling stats",
		},
		[]string{"biz_name", "path", "code"})
	Metric4RpcCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "rpc_req_count_stats",
			Help:      "esports rpc req count stats",
		},
		[]string{"biz_name", "path"})
	Metric4RpcLatency = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "rpc_req_latency_stats",
			Help:      "esports rpc req latency stats",
		},
		[]string{"biz_name", "path"})

	Metric4MemoryCache = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "memory_cache_stats",
			Help:      "esports memory cache stats",
		},
		[]string{"biz_name"})

	Metric4CacheResetFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports",
			Name:      "cache_reset_failed_stats",
			Help:      "cache reset failed stats",
		},
		[]string{"biz_name", "cache_type"})

	prometheus.MustRegister(
		metric4DBRestoreBiz,
		metric4Limiter,
		Metric4RpcQps,
		Metric4RpcCount,
		Metric4MemoryCache,
		Metric4CacheResetFailed,
		Metric4RpcLatency)
}

func AddDBBackSourceMetricsByKeyList(bizName string, list []int64) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBBackSource, bizName}...).Add(float64(len(list)))
}

func AddDBErrMetricsByKeyList(bizName string, list []int64) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBErr, bizName}...).Add(float64(len(list)))
}

func AddDBNoResultMetricsByKeyList(bizName string, list []int64) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBNoResult, bizName}...).Add(float64(len(list)))
}
