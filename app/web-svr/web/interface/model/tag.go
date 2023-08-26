package model

import (
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	// TagStateOK means normal state
	TagStateOK = 0
	// TagStateDeleted means tag was deleted
	TagStateDeleted = 1
	// TagStateBlocked means tag was blocked
	TagStateBlocked = 2
	TagTypArc       = 3 //稿件
)

// TagAids .
type TagAids struct {
	Code  int     `json:"code"`
	Total int64   `json:"total"`
	Data  []int64 `json:"data"`
}

// TagDetail .
type TagDetail struct {
	Total int      `json:"total"`
	List  []*BvArc `json:"list"`
	*TagTop
}

type TagArcItem struct {
	Aid   int64  `json:"aid"`
	Title string `json:"title"`
	Pic   string `json:"pic"`
	Stat  struct {
		View int32 `json:"view"`
		Like int32 `json:"like"`
	} `json:"stat"`
	Duration int64 `json:"duration"`
	Author   struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
	} `json:"owner"`
	Bvid string `json:"bvid"`
}

func (a *TagArcItem) FormArc(arc *arcgrpc.Arc) {
	a.Aid = arc.GetAid()
	a.Title = arc.GetTitle()
	a.Pic = arc.GetPic()
	a.Stat.View = arc.GetStat().View
	a.Stat.Like = arc.GetStat().Like
	a.Duration = arc.GetDuration()
	a.Author.Mid = arc.GetAuthor().Mid
	a.Author.Name = arc.GetAuthor().Name
	a.Bvid, _ = bvid.AvToBv(arc.GetAid())
}

type TagArcsReq struct {
	TagID  int64  `json:"tid" form:"tid" validate:"required,min=1"`
	Source int32  `json:"source" form:"source"`
	Offset string `json:"offset" form:"offset"`
	PS     int32  `json:"ps" form:"ps" default:"20" validate:"min=1"`
}

type TagArcsReply struct {
	HasMore bool          `json:"has_more"`
	Offset  string        `json:"offset"`
	List    []*TagArcItem `json:"list"`
}
