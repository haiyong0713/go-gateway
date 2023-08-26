package adresource

import (
	"go-common/library/stat/metric"
)

var (
	StatSceneResourceID = metric.NewBusinessMetricCount("adresource", "scene", "id")
)
