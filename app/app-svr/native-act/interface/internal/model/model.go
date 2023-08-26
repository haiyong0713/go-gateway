package model

const (
	// 请求来源
	ReqFromIndex   = "index"
	ReqFromSubPage = "subpage"
	// ActivityFrom
	ActFromTm     = "tm_dynamic"
	ActFromAllAct = "all_activity"
	ActFromDt     = "dt_dynamic"
	// TabFrom
	TabFromUserSpace  = "user_space_activity_tab"
	TabFromTopicLayer = "topic_layer"
)

func IsFromIndex(from string) bool {
	return from == ReqFromIndex
}

func NeedLayerDynamic(activityFrom string) bool {
	return activityFrom == ActFromTm || activityFrom == ActFromDt
}
