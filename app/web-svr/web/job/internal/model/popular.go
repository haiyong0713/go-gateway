package model

import xtime "go-common/library/time"

type Series struct {
	ID      int64      `json:"id"`
	Type    string     `json:"type"`
	Number  int64      `json:"number"`
	Subject string     `json:"subject"`
	Stime   xtime.Time `json:"stime"`
	Etime   xtime.Time `json:"etime"`
	Status  int        `json:"status"`
	Name    string     `json:"name"`
}

type SeriesConfig struct {
	Series
	Label         string `json:"label"`
	Hint          string `json:"hint"`
	Color         int    `json:"color"`
	Cover         string `json:"cover"`
	ShareTitle    string `json:"share_title"`
	ShareSubtitle string `json:"share_subtitle"`
	MediaID       int64  `json:"media_id"` // 播单ID
}

type SeriesResource struct {
	RID        int64  `json:"rid"`
	Rtype      string `json:"rtype"`
	SerieID    int64  `json:"serie_id"`
	Position   int    `json:"position"`
	RcmdReason string `json:"rcmd_reason"`
}

type ActPlatHistory struct {
	Activity  string `json:"activity"`
	Counter   string `json:"counter"`
	CounterID int64  `json:"counter_id"`
	Mid       int64  `json:"mid"`
	TimeStamp int64  `json:"timestamp"`
	Diff      int64  `json:"diff"`
	Total     int64  `json:"total"`
}

type MgrSeriesConfig struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Number        int64      `json:"number"`
	Hint          string     `json:"hint"`
	Subject       string     `json:"subject"`
	Color         int64      `json:"color"`
	Cover         string     `json:"cover"`
	ShareTitle    string     `json:"share_title"`
	ShareSubtitle string     `json:"share_subtitle"`
	PushTitle     string     `json:"push_title"`
	PushSubtitle  string     `json:"push_subtitle"`
	Status        int64      `json:"status"`
	TaskStatus    int64      `json:"task_status"`
	MediaID       int64      `json:"media_id"`
	Stime         xtime.Time `json:"stime"`
	Etime         xtime.Time `json:"etime"`
}

type MgrSeriesList struct {
	Goto            string `json:"goto"`
	Param           int64  `json:"param"`
	Cover           string `json:"cover"`
	Title           string `json:"title"`
	CoverRightText1 string `json:"cover_right_text_1"`
	RightDesc1      string `json:"right_desc_1"`
	RightDesc2      string `json:"right_desc_2"`
	RcmdReason      string `json:"rcmd_reason"`
}

type MgrSeriesData struct {
	Config *MgrSeriesConfig `json:"config"`
	List   []*MgrSeriesList `json:"list"`
}
