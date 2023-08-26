package relation

import (
	xtime "go-common/library/time"

	accv1 "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

const (
	AttrNoRelation = uint32(0)
)

type Vip struct {
	Type          int    `json:"vipType"`
	DueDate       int64  `json:"vipDueDate"`
	DueRemark     string `json:"dueRemark"`
	AccessStatus  int    `json:"accessStatus"`
	VipStatus     int    `json:"vipStatus"`
	VipStatusWarn string `json:"vipStatusWarn"`
}

// Following is user followinng info.
type Following struct {
	*MFollow
	Uname          string             `json:"uname"`
	Face           string             `json:"face"`
	Sign           string             `json:"sign"`
	OfficialVerify accv1.OfficialInfo `json:"official_verify"`
	Vip            Vip                `json:"vip"`
	Live           int                `json:"live"`
}

type MFollow struct {
	Mid       int64      `json:"mid"`
	Attribute uint32     `json:"attribute"`
	Source    uint32     `json:"-"`
	CTime     xtime.Time `json:"-"`
	MTime     xtime.Time `json:"mtime"`
	Tag       []int64    `json:"tag"`
	Special   int32      `json:"special"`
}

func (t *MFollow) ConverFollow(in *relationgrpc.FollowingReply) {
	if in == nil {
		return
	}
	t.Mid = in.Mid
	t.Attribute = in.Attribute
	t.Source = in.Source
	t.CTime = in.CTime
	t.MTime = in.MTime
	t.Tag = in.Tag
	t.Special = in.Special
}

type Tag struct {
	Mid            int64              `json:"mid"`
	Uname          string             `json:"uname"`
	Face           string             `json:"face"`
	Sign           string             `json:"sign"`
	OfficialVerify accv1.OfficialInfo `json:"official_verify"`
	Vip            Vip                `json:"vip"`
	Live           int                `json:"live"`
}

// ByMTime implements sort.Interface for []model.Following based on the MTime field.
type ByMTime []*relationgrpc.FollowingReply

func (mt ByMTime) Len() int           { return len(mt) }
func (mt ByMTime) Swap(i, j int)      { mt[i], mt[j] = mt[j], mt[i] }
func (mt ByMTime) Less(i, j int) bool { return mt[i].MTime < mt[j].MTime }
