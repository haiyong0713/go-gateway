package tool

import "github.com/prometheus/client_golang/prometheus"

var (
	Metric4AutoSub    *prometheus.CounterVec
	Metric4BizListLen *prometheus.GaugeVec
	Metric4Component  *prometheus.GaugeVec
)

func init() {
	Metric4AutoSub = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_esports_job",
			Name:      "auto_sub_stats",
			Help:      "esports job auto subscribe stats",
		},
		[]string{"season_id", "team_id", "status"})
	Metric4BizListLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_esports_job",
			Name:      "biz_list_len",
			Help:      "esports job biz list length",
		},
		[]string{"biz_name"})
	Metric4Component = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_esports_job",
			Name:      "going_seasons_contests",
			Help:      "esports component going seasons and contests",
		},
		[]string{"biz_name"})

	prometheus.MustRegister(Metric4AutoSub, Metric4BizListLen, Metric4Component)
}
