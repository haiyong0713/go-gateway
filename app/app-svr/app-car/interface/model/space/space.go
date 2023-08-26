package space

import (
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/mine"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type SpaceParam struct {
	model.DeviceInfo
	Vmid     int64  `form:"vmid"`
	Pn       int    `form:"pn" default:"1" validate:"min=1"`
	Ps       int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FromType string `form:"from_type"`
	ParamStr string `form:"param"`
}

type Info struct {
	Account  *mine.Mine      `json:"account,omitempty"`
	Top      *Top            `json:"top,omitempty"`
	Items    []cardm.Handler `json:"items,omitempty"`
	Page     *cardm.Page     `json:"page,omitempty"`
	Relation *model.Relation `json:"relation,omitempty"`
}

type Top struct {
	Title      string     `json:"title,omitempty"`
	Buttom     *TopButtom `json:"buttom,omitempty"`
	ShowBanner bool       `json:"show_banner,omitempty"`
}

type TopButtom struct {
	Text string `json:"text,omitempty"`
	Icon string `json:"icon,omitempty"`
	URI  string `json:"uri,omitempty"`
}

func (t *Info) FromSpace(mine *mine.Mine, count, stat, mid int64, authorRelations map[int64]*relationgrpc.InterrelationReply) {
	const (
		_showBanner = 10
	)
	if mine == nil {
		return
	}
	t.Account = mine
	t.Top = &Top{
		Title: fmt.Sprintf("TA的视频（%d）", count),
	}
	if count > 0 {
		t.Top.Buttom = &TopButtom{
			Text: "播放全部",
			Icon: "play",
			URI:  fmt.Sprintf("bilithings://player?sourceType=%s&vmid=%d", model.EntranceSpace, mine.Mid),
		}
	}
	if count > _showBanner {
		t.Top.ShowBanner = true
	}
	if t.Account != nil {
		t.Account.Fans = model.FanString(int32(stat))
	}
	if mid > 0 && mid != mine.Mid {
		t.Relation = model.RelationChange(mine.Mid, authorRelations)
	}
}
