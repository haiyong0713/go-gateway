package metric

import (
	"go-common/library/stat/metric"
)

var (
	// 卡片类型曝光分布
	DynamicCard = metric.NewBusinessMetricCount("dynamic_card", "from", "dyn_type")
	// 附加大卡分布
	DynamicAdditional = metric.NewBusinessMetricCount("dynamic_additional", "from", "dyn_type", "addition_type")
	// 卡片类型分布
	DynamicExt = metric.NewBusinessMetricCount("dynamic_ext", "from", "dyn_type", "ext_type")
	// 转发源卡类型分布
	DynamicForward = metric.NewBusinessMetricCount("dynamic_forward", "from", "dyn_type")
	// 关系链监控
	DyanmicRelationAPI = metric.NewBusinessMetricCount("dynamic_relation_api", "api_type", "reason")
	// 动态接口监控
	DynamicCoreAPI = metric.NewBusinessMetricCount("dynamic_core_api", "api_type", "reason")
	// 物料接口监控
	DyanmicItemAPI = metric.NewBusinessMetricCount("dynamic_item_api", "api_type", "reason")
	// 回填接口检测
	DynamicBackfillAPI = metric.NewBusinessMetricCount("dynamic_backfill_api", "api_type", "reason")
	// 销卡监控
	DynamicCardError = metric.NewBusinessMetricCount("dynamic_card_error", "from", "dyn_type", "reason")
	// 模块监控
	DynamicModuleError = metric.NewBusinessMetricCount("dynamic_module_error", "from", "dyn_type", "module_type", "reason")
)
