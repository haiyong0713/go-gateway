package tool

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Metric4CommonQps *prometheus.CounterVec
	metric4BizStatus *prometheus.CounterVec

	metric4BizCountOfLow    *prometheus.CounterVec
	metric4BizCountOfHigh   *prometheus.CounterVec
	metric4BizLatencyOfLow  *prometheus.CounterVec
	metric4BizLatencyOfHigh *prometheus.CounterVec

	metric4CommonGauge *prometheus.GaugeVec
)

const (
	BizStatusOfSucceed = "succeed"
	BizStatusOfFailed  = "failed"
)

func init() {
	metric4BizCountOfLow = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "biz_count_stats_low",
			Help:      "activity job biz count(low)",
		},
		[]string{"srv_name", "biz_name"})
	metric4BizLatencyOfLow = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "biz_latency_stats_low",
			Help:      "activity job latency stats(low)",
		},
		[]string{"srv_name", "biz_name"})
	metric4BizCountOfHigh = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "biz_count_stats_high",
			Help:      "activity job biz count(high)",
		},
		[]string{"srv_name", "biz_name"})
	metric4BizLatencyOfHigh = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "biz_latency_stats_high",
			Help:      "activity job latency stats(high)",
		},
		[]string{"srv_name", "biz_name"})
	Metric4CommonQps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity_job",
			Name:      "common_biz_qps",
			Help:      "activity job common biz qps stats",
		},
		[]string{"biz_name"})
	metric4BizStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity_job",
			Name:      "common_biz_status",
			Help:      "activity job common biz status stats",
		},
		[]string{"biz_name", "status"})
	metric4CommonGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_activity_job",
			Name:      "common_gauge_status",
			Help:      "activity job common gauge status stats",
		},
		[]string{"biz_name"})

	prometheus.MustRegister(
		metric4BizCountOfLow,
		metric4BizLatencyOfLow,
		metric4BizCountOfHigh,
		metric4BizLatencyOfHigh,
		metric4BizStatus,
		metric4CommonGauge,
		Metric4CommonQps)
}

func IncrBizCountAndLatency(srvName, bizName string, startedTime time.Time) {
	startedAt := startedTime.UnixNano()
	now := time.Now().UnixNano()
	latency := (now - startedAt) / 1e6

	if latency > 300 {
		metric4BizCountOfHigh.WithLabelValues(
			[]string{
				srvName,
				bizName,
			}...).Inc()
		metric4BizLatencyOfHigh.WithLabelValues(
			[]string{
				srvName,
				bizName,
			}...).Add(float64(latency))
	} else {
		metric4BizCountOfLow.WithLabelValues(
			[]string{
				srvName,
				bizName,
			}...).Inc()
		metric4BizLatencyOfLow.WithLabelValues(
			[]string{
				srvName,
				bizName,
			}...).Add(float64(latency))
	}
}

func ResetCommonGauge(bizName string, num int64) {
	metric4CommonGauge.WithLabelValues([]string{bizName}...).Set(float64(num))
}

func IncrCommonGauge(bizName string) {
	metric4CommonGauge.WithLabelValues([]string{bizName}...).Inc()
}

func DecCommonGauge(bizName string) {
	metric4CommonGauge.WithLabelValues([]string{bizName}...).Dec()
}

func IncrCommonBizStatus(bizName, status string) {
	metric4BizStatus.WithLabelValues([]string{bizName, status}...).Inc()
}
