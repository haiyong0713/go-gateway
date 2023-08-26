package metric

import (
	"go-common/library/stat/metric"
)

var (
	FeatureSdkError = metric.NewBusinessMetricCount("feature_sdk_error", "bus_type", "err_type", "bus_key_name")
)
