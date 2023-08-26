package model

import (
	arcv1 "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/time"
)

type SimpleArc struct {
	ID      int
	AID     int64
	MID     int
	TypeID  int32
	Title   string
	Content string
	Cover   string
	Deleted int
	Result  int
	Valid   int
	Mtime   time.Time
	Pubtime time.Time
}

type UpperState int

type GrayType int

const (
	UpperNew          UpperState = 1 // 新增同步
	UpperFull         UpperState = 2 // 全量同步
	UpperManual       UpperState = 3 // 手动同步
	AttrBitSteinsGate            = uint(29)
	AttrBitIsPUGVPay             = uint(30)
	GrayAll           GrayType   = 1 // 全部可见
	GrayNotRecom      GrayType   = 2 // 非自动填充可见(不上推荐
)

type Upper struct {
	ID       int        `json:"id"`
	MID      int64      `json:"mid"`   // up主mid
	State    UpperState `json:"state"` // 同步类型
	Toinit   int        `json:"toinit"`
	Retry    int        `json:"retry"`
	Deleted  int        `json:"deleted"`
	Ctime    time.Time  `json:"ctime"`
	Mtime    time.Time  `json:"mtime"`     // 修改时间
	MtimeStr string     `json:"mtime_str"` // 修改时间
	CmsName  string     `json:"cms_name"`  // cms干预的昵称
	OriName  string     `json:"ori_name"`  // up主原名
	OriFace  string     `json:"ori_face"`
	CmsFace  string     `json:"cms_face"` // cms干预的头像
	Valid    int        `json:"valid"`    // 上下架状态,0=下架,1=上架
}

type ArcAllow struct {
	Aid         int64
	State       int32
	Ugcpay      int32
	Typeid      int32
	Copyright   int32
	AttrIsPgc   int32
	AttrIsPugv  int32
	AttrIsStein int32
}

func (a *ArcAllow) FromArcReply(reply *arcv1.Arc) {
	a.Aid = reply.Aid
	a.State = reply.State
	a.Ugcpay = reply.Rights.UGCPay
	a.Typeid = reply.TypeID
	a.Copyright = reply.Copyright
	a.AttrIsPgc = reply.AttrVal(arcv1.AttrBitIsPGC)
	a.AttrIsPugv = reply.AttrVal(AttrBitIsPUGVPay)
	a.AttrIsStein = reply.AttrVal(AttrBitSteinsGate)
}

// CanPlay distinguishes whether an archive can play or not
func (a *ArcAllow) CanPlay() bool {
	return a.State >= 0 || a.State == -6
}

// IsOrigin distinguishes whether an archive is original or not
func (a *ArcAllow) IsOrigin() bool {
	return a.Copyright == 1
}

// Archive archive def. corresponding to our table structure
type Archive struct {
	ID        int
	AID       int64
	MID       int64
	TypeID    int32
	Videos    int64
	Title     string
	Cover     string
	Content   string
	Duration  int64
	Copyright int32
	Pubtime   time.Time
	Ctime     time.Time
	Mtime     time.Time
	State     int32
	Manual    int
	Valid     int
	Submit    int
	Retry     int
	Result    int
	Deleted   int
	Priority  int8 //审核优先级
	GrayType  GrayType
}

// FromArcReply def
func (a *Archive) FromArcReply(arc *arcv1.Arc) {
	a.AID = arc.Aid
	a.MID = arc.Author.Mid
	a.Videos = arc.Videos
	a.TypeID = arc.TypeID
	a.Title = arc.Title
	a.Cover = arc.Pic
	a.Content = arc.Desc
	a.Duration = arc.Duration
	a.Copyright = arc.Copyright
	a.Pubtime = arc.PubDate
	a.State = arc.State
	a.GrayType = GrayAll
}
