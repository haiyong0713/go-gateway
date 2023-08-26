package like

import xtime "go-common/library/time"

type CourseInfo struct {
	IsJoin    bool           `json:"is_join"`
	BuySeason int32          `json:"buy_season"`
	List      []*CourseOrder `json:"list"`
}

type CourseOrder struct {
	SeasonID    int32  `json:"season_id"`
	SeasonTitle string `json:"season_title"`
	EpCount     int32  `json:"ep_count"`
	Cover       string `json:"cover"`
	Duration    int64  `json:"duration"`
	IsBuy       bool   `json:"is_buy"`
	OrderNo     string `json:"-"`
	RealPrice   int32  `json:"-"`
}

type WinterProgress struct {
	IsJoin         bool  `json:"is_join"`
	RealPrice      int32 `json:"real_price"`
	SeasonID       int32 `json:"season_id"`
	TotalProgress  int64 `json:"total_progress"`
	ClockIn        int64 `json:"clock_in"`
	WatchProgress  int64 `json:"watch_progress"`
	ShareProgress  int64 `json:"share_progress"`
	UploadProgress int64 `json:"upload_progress"`
	WatchDuration  int64 `json:"watch_duration"`
}

type ProgressHistory struct {
	Sid int64 `json:"sid"`
}

type WinterStudy struct {
	ID             int64      `json:"id"`
	Mid            int64      `json:"mid"`
	OrderNo        string     `json:"order_no"`
	SeasonID       int32      `json:"season_id"`
	RealPrice      int32      `json:"real_price"`
	Duration       int64      `json:"duration"`
	EpCount        int64      `json:"ep_count"`
	SeasonTitle    string     `json:"season_title"`
	Cover          string     `json:"cover"`
	IsEnd          int64      `json:"is_end"`
	TotalProgress  int64      `json:"total_progress"`
	ClockIn        int64      `json:"clock_in"`
	WatchProgress  int64      `json:"watch_progress"`
	ShareProgress  int64      `json:"share_progress"`
	UploadProgress int64      `json:"upload_progress"`
	WatchDuration  int64      `json:"watch_duration"`
	Ctime          xtime.Time `json:"ctime"`
}

type ParamWinterJoin struct {
	SeasonID int64  `form:"season_id" validate:"min=1"`
	IsNotice int64  `form:"is_notice"`
	IP       string `form:"-"`
	HTTPReserveReport
}
