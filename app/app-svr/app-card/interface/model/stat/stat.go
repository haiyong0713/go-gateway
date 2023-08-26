package stat

import (
	"go-common/library/stat/metric"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
)

const (
	// RowType is to define the arrangement of cards, etc single、double...
	RowTypeSingle = "single" //单列
	RowTypeDouble = "double" //双列
	RowTypeIPad   = "ipad"   //ipad
)

var (
	MetricResponseCardTotal = metric.NewBusinessMetricCount("response_card_total", "row_type", "goto", "card_type", "desc")
	MetricAICardTotal       = metric.NewBusinessMetricCount("ai_card_total", "row_type", "goto", "jumpgoto")
	MetricDiscardCardTotal  = metric.NewBusinessMetricCount("discard_card_total", "row_type", "goto", "jumpgoto", "card_type", "reason")
	MetricAppCardTotal      = metric.NewBusinessMetricCount("app_card_total", "row_type", "card_type", "card_goto")
	MetricFfCoverTotal      = metric.NewBusinessMetricCount("ff_cover_total", "from")
	MetricFeedGuidanceTotal = metric.NewBusinessMetricCount("feed_guidance_discard_total", "mobi_app", "reason")
	MetricStoryAICardTotal  = metric.NewBusinessMetricCount("story_ai_card_total", "goto", "plat")
	MetricStoryCardTotal    = metric.NewBusinessMetricCount("story_card_total", "goto", "plat")
)

func BuildRowType(cs cdm.ColumnStatus, plat int8) string {
	if cdm.IsPad(plat) {
		return RowTypeIPad
	}
	switch cs {
	case cdm.ColumnSvrSingle, cdm.ColumnUserSingle:
		return RowTypeSingle
	case cdm.ColumnSvrDouble, cdm.ColumnUserDouble:
		return RowTypeDouble
	default:
		return ""
	}
}
