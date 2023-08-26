package fit

import go_common_library_time "go-common/library/time"

const (
	// 小卡事件
	TunnelV2EventAlready  = 108009
	TunnelV2NotExists     = 108019
	TunnelV2CardStatusErr = 108014
)

// DBActFitPlanConfig db字段
type DBActFitPlanConfig struct {
	ID          int64                       `json:"id"`
	PlanTitle   string                      `json:"plan_title"`
	PlanTags    string                      `json:"plan_tags"`
	BodanId     string                      `json:"bodan_id"`
	PlanView    int64                       `json:"plan_view"`
	PlanDanmaku int64                       `json:"plan_danmaku"`
	PlanFav     int64                       `json:"plan_fav"`
	PicCover    string                      `json:"pic_cover"`
	Creator     string                      `json:"creator"`
	Status      int                         `json:"status"`
	Ctime       go_common_library_time.Time `json:"ctime"`
	Mtime       go_common_library_time.Time `json:"ctime"`
}

// PlanRecordRes 接口返回计划字段
type PlanRecordRes struct {
	ID      int64  `json:"plan_id"`
	BodanId string `json:"bodan_id"`
}

// ActPlatHistoryTopicMsg
type ActPlatHistoryTopicMsg struct {
	Activity  string `json:"activity"`
	Counter   string `json:"counter"`
	MID       int64  `json:"mid"`
	TimeStamp int64  `json:"timestamp"`
	Diff      int64  `json:"diff"`
	Total     int64  `json:"total"`
}

// UserSignDaysRes fitUserInfo接口的返回值
type UserSignDaysRes struct {
	SignDays int64 `json:"sign_days"`
	Time     int64 `json:"time"`
}

// ReservedUser 预约的用户
type ReservedUser struct {
	ID  int64
	MID int64
}
