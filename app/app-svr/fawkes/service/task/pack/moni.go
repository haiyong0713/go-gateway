package pack

import "go-common/library/stat/metric"

var _metricCleanNasSize = metric.NewGaugeVec(&metric.GaugeVecOpts{
	Namespace: "fawkes",
	Subsystem: "nas",
	Name:      "delete_bytes",
	Help:      "delete nas bytes",
	Labels:    []string{"app_key", "build_type"},
})

var _metricCleanNasCount = metric.NewCounterVec(&metric.CounterVecOpts{
	Namespace: "fawkes",
	Subsystem: "nas",
	Name:      "delete_count",
	Help:      "delete nas count",
	Labels:    []string{"app_key", "build_type"},
})
