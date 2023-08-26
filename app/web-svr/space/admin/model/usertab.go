package model

import (
	"fmt"
	xtime "go-common/library/time"
)

const (
	UserTabLog = 242
)

type UserTabReq struct {
	ID        int64      `json:"id" form:"id"`
	TabType   int        `json:"tab_type" form:"tab_type"`
	Mid       int64      `json:"mid" form:"mid"`
	TabName   string     `json:"tab_name" form:"tab_name"`
	TabOrder  int64      `json:"tab_order" form:"tab_order" default:"0"`
	TabCont   int64      `json:"tab_cont" form:"tab_cont"`
	Stime     xtime.Time `json:"stime" form:"stime"`
	Etime     xtime.Time `json:"etime" form:"etime"`
	Online    int        `json:"online" form:"online" default:"-1"`
	Deleted   int        `json:"deleted" form:"deleted" default:"0"`
	IsSync    int        `json:"is_sync" form:"is_sync"`
	IsDefault int8       `json:"is_default" form:"is_default"`
	LimitsStr string     `gorm:"column:limits"`
	Limits    []*Limit   `json:"limits" form:"limits" gorm:"-"`
	H5Link    string     `json:"h5_link" form:"h5_link"`
}

type CommercialTabReq struct {
	ID        int64      `json:"id"`
	TabType   int        `json:"tab_type"`
	Mid       int64      `json:"mid"`
	TabName   string     `json:"tab_name"`
	TabOrder  int64      `json:"tab_order" default:"0"`
	TabCont   int64      `json:"tab_cont" `
	Stime     xtime.Time `json:"stime"`
	Etime     xtime.Time `json:"etime"`
	Online    int        `json:"online"`
	Deleted   int        `json:"deleted" default:"0"`
	Username  string     `json:"username"`
	IsSync    int        `json:"is_sync" default:"1" form:"is_sync"`
	IsDefault int8       `json:"is_default" form:"is_default"`
}

type SpaceUserTab struct {
	UserTabReq
	Ctime xtime.Time `json:"ctime" form:"ctime"`
	Mtime xtime.Time `json:"mtime" form:"mtime"`
}

type UserTabListReq struct {
	TabType int   `json:"tab_type" form:"tab_type"`
	Mid     int64 `json:"mid" form:"mid"`
	Online  int   `json:"online" form:"online" default:"-1"`
	Ps      int   `json:"ps" form:"ps" default:"20"`
	Pn      int   `json:"pn" form:"pn" default:"1"`
}

type UserTabListReply struct {
	UserTabReq
	MidName  string `json:"mid_name" form:"mid_name"`
	Official int32  `json:"official" form:"official"`
}

func (u *UserTabListReply) String() string {
	return fmt.Sprintf("{ID:%d,MID:%+v,TabName:%+v,TabCont:%+v,TabType:%+v,IsSync:%+v,IsDefault:%+v} ",
		u.ID, u.Mid, u.TabName, u.TabCont, u.TabType, u.IsSync, u.IsDefault)
}

type UserTabList struct {
	List []*UserTabListReply `json:"list"`
	Page Page                `json:"page"`
}

type MidInfoReply struct {
	Mid      int64  `json:"mid" form:"mid"`
	MidName  string `json:"mid_name" form:"mid_name"`
	Official int32  `json:"official" form:"official"`
}

type Limit struct {
	Conditions string `json:"conditions"`
	Plat       int32  `json:"plat"`
	Build      int32  `json:"build"`
}
