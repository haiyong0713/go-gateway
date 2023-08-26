package dao

import "github.com/prometheus/client_golang/prometheus"

var (
	Metric4SubscribeContestSend     *prometheus.GaugeVec
	Metric4SubscribeContestSendUser *prometheus.GaugeVec
)

func init() {
	Metric4SubscribeContestSend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_esports_service",
			Name:      "subscribe_contest_send",
			Help:      "esports subscribe contest send",
		},
		[]string{"biz_name"})
	Metric4SubscribeContestSendUser = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_esports_service",
			Name:      "subscribe_contest_send_user",
			Help:      "esports subscribe contest send user",
		},
		[]string{"biz_name"})
	prometheus.MustRegister(Metric4SubscribeContestSend, Metric4SubscribeContestSendUser)
}
