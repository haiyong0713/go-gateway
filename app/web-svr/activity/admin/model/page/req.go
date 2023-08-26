package page

import xtime "go-common/library/time"

type ReqPageList struct {
	Page     int        `json:"page" form:"page" default:"1" validate:"min=1"`
	PageSize int        `json:"pagesize" form:"pagesize" default:"15" validate:"min=1"`
	Keyword  string     `json:"keyword" form:"keyword"`
	States   []int      `json:"state" form:"state,split"`
	Mold     []int      `json:"mold" form:"mold,split"`
	ReplyID  int        `json:"reply_id" form:"reply_id"`
	SCTime   xtime.Time `json:"sctime" form:"sctime"`
	ECTime   xtime.Time `json:"ectime" form:"ectime"`
	Creator  string     `json:"creator" form:"creator"`
	Plat     []int      `json:"plat" form:"plat,split"`
	Dept     []int      `json:"dept" form:"dept,split"`
	Order    string     `json:"order" form:"order" default:"ctime"`
}
