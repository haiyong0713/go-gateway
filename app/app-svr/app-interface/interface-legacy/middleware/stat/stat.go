package stat

import "go-common/library/stat/metric"

// 监控参数
var (
	MetricSearchAICardTotal  = metric.NewBusinessMetricCount("search_ai_card_total", "linktype", "type")
	MetricSearchAppCardTotal = metric.NewBusinessMetricCount("search_app_card_total", "linktype", "goto")
	MetricSearchAiMainFailed = metric.NewBusinessMetricCount("search_ai_main_failed", "path", "code")
)

// 降级参数
var (
	SearchDegreeArgs = []string{"mobi_app", "device", "rid", "keyword",
		"highlight", "lang", "duration", "order", "filtered", "platform", "zoneid", "from_source", "recommend", "parent_mode",
		"pn", "ps", "is_org_query", "teenagers_mode", "lessons_mode"}
	SearchTypeDegreeArgs = []string{"mobi_app", "device", "type", "keyword",
		"filtered", "zoneid", "order", "platform", "highlight", "category_id", "user_type", "order_sort", "pn", "ps"}
	HotSearchDegreeArgs    = []string{"build", "device", "mobi_app", "platform", "limit"}
	SquareSearchDegreeArgs = []string{"build", "device", "mobi_app", "platform", "limit", "from", "show", "s_locale", "c_locale"}
)
