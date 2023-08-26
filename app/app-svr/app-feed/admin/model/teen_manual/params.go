package teen_manual

import (
	xtime "go-common/library/time"
)

type SearchTeenManualReq struct {
	Mid      int64  `form:"mid" validate:"min=0"`
	Operator string `form:"operator"`
	Pn       int64  `form:"pn" validate:"min=0" default:"1"`
	Ps       int64  `form:"ps" validate:"min=0" default:"20"`
}

type SearchTeenManualRly struct {
	Page *Page             `json:"page"`
	List []*TeenManualItem `json:"list"`
}

type TeenManualItem struct {
	Mid         int64      `json:"mid"`
	Name        string     `json:"name"`
	IsRealname  bool       `json:"is_realname"`
	AgeBand     string     `json:"age_band"`
	State       int64      `json:"state"`
	ManualForce int64      `json:"manual_force"`
	Operator    string     `json:"operator"`
	OperateTime xtime.Time `json:"operate_time"`
}

type Page struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

type OpenReq struct {
	Mid    int64  `form:"mid" validate:"required"`
	Remark string `form:"remark"`
}

type QuitReq struct {
	Mid    int64  `form:"mid" validate:"required"`
	Remark string `form:"remark"`
}

type TeenManualLogReq struct {
	Mid int64 `form:"mid" validate:"required"`
}

type TeenManualLogRly struct {
	List []*TeenagerManualLog `json:"list"`
}
