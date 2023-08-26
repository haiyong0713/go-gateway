package fit

import go_common_library_time "go-common/library/time"

// UserSignDaysRes fitUserInfo接口的返回值
type UserSignDaysRes struct {
	IsJoin   int   `json:"is_join"`
	SignDays int64 `json:"sign_days"`
}

// PlanRecord act_fit_plan_config计划表字段
type PlanRecord struct {
	PlanTitle   string `json:"plan_title"`
	PlanTags    string `json:"plan_tags"`
	BodanId     string `json:"bodan_id"`
	PlanView    int64  `json:"plan_view"`
	PlanDanmaku int64  `json:"plan_danmaku"`
	PlanFav     int64  `json:"plan_fav"`
	Creator     string `json:"creator"`
}

// PlanRecordRes 接口返回计划字段
type PlanRecordRes struct {
	ID          int64  `json:"plan_id"`
	PlanTitle   string `json:"title"`
	PlanTags    string `json:"tags"`
	PlanView    int64  `json:"view"`
	PlanDanmaku int64  `json:"danmaku"`
	PlanFav     int64  `json:"fav"`
	PicCover    string `json:"pic_cover,omitempty"`
}

// PlanRecordListRes 接口返回计划列表
type PlanRecordListRes struct {
	PlanList []*PlanRecordRes `json:"list"`
	Page     int              `json:"page"`
	Size     int              `json:"size"`
	Total    int              `json:"total"`
}

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

// PlanWeekBodanList 系列计划播单列表返回结构
type PlanWeekBodanList struct {
	Count int            `json:"video_count"`
	List  []*BodanDetail `json:"bodan_list"`
}

// 热门视频tag返回结构
type HotTagsListRes struct {
	List []*TagInfo `json:"list"`
}

// 热门视频列表
type HotVideosRes struct {
	VideoList []*VideoDetail `json:"video_list"`
	Page      int            `json:"page"`
	Size      int            `json:"size"`
	Total     int            `json:"total"`
}

// TagInfo
type TagInfo struct {
	Title   string `json:"title"`
	BodanId int    `json:"bodan_id"`
}

// BodanDetail 每一个星期播单的字段
type BodanDetail struct {
	BodanId    int64          `json:"bodan_id"`
	BodanTitle string         `json:"bodan_title"`
	BodanDesc  string         `json:"bodan_desc"`
	List       []*VideoDetail `json:"list"`
}

// VideoDetail 视频详情字段
type VideoDetail struct {
	Aid       int64  `json:"video_id"`
	Title     string `json:"title"`
	Duration  int64  `json:"duration"`
	Pic       string `json:"pic"`
	View      int32  `json:"view"`
	Reply     int32  `json:"reply"`
	Danmaku   int32  `json:"danmaku"`
	ShortLink string `json:"short_link"`
	IsViewed  bool   `json:"is_viewed"`
}

// HistorySource 积分任务返回的history下的source字段
type HistorySource struct {
	Aid int64 `json:"aid"`
}
