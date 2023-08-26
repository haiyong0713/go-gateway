package anti_crawler

import "time"

type UserLogParam struct {
	Buvid     string  `form:"buvid"`
	Mid       int64   `form:"mid"`
	ReqHost   string  `form:"req_host"`
	Path      string  `form:"path"`
	TimeRange []int64 `form:"time_range,split"`
	Page      int     `form:"page" validate:"min=1"`
	PerPage   int     `form:"perPage" validate:"min=1,max=20"`
	Stime     int64   `form:"-"`
	Etime     int64   `form:"-"`
}

type InfocMsg struct {
	Mid            int64     `json:"mid"`
	Buvid          string    `json:"buvid"`
	Host           string    `json:"host"`
	Path           string    `json:"path"`
	Method         string    `json:"method"`
	Header         string    `json:"header"`
	Query          string    `json:"query"`
	Body           string    `json:"body"`
	Referer        string    `json:"referer"`
	IP             string    `json:"ip"`
	Ctime          int64     `json:"ctime"`
	CtimeHuman     time.Time `json:"ctime_human"`
	ResponseHeader string    `json:"response_header"`
	ResponseBody   string    `json:"response_body"`
}

type BusinessConfigListReq struct {
	Value string `form:"value"`
}

type BusinessConfigUpdateReq struct {
	Value    string `form:"value" validate:"required"`
	Forever  int    `form:"forever" validate:"min=0,max=1"` // 是否永久，1：是，0：否
	Datetime int64  `form:"datetime"`
}

type BusinessConfigDeleteReq struct {
	Value string `form:"value" validate:"required"`
}

type WList struct {
	Value    string `json:"value"`
	Forever  int    `json:"forever"`
	Deadline int64  `json:"deadline"`
}
