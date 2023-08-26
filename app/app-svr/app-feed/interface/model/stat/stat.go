package stat

import "go-common/library/stat/metric"

var (
	MetricCMResource = metric.NewBusinessMetricCount("cm_resource", "id", "plat")
)
