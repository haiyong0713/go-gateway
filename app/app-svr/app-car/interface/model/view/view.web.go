package view

import (
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/archive/service/api"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

type ViewWebParam struct {
	Otype string `form:"otype"`
	Oid   int64  `form:"oid"`
}

type ViewWeb struct {
	History      *History      `json:"history,omitempty"`
	Pages        []*PageWeb    `json:"pages,omitempty"`
	PagesStyle   string        `json:"pages_style,omitempty"`
	Otype        string        `json:"otype,omitempty"`
	Title        string        `json:"title,omitempty"`
	SeasonTitle  string        `json:"season_title,omitempty"`
	Desc         string        `json:"desc,omitempty"`
	Duration     int64         `json:"duration,omitempty"`
	Cover        string        `json:"cover,omitempty"`
	BadgeInfo    *BadgeInfo    `json:"badge_info,omitempty"`
	SeasonType   int           `json:"season_type,omitempty"`
	Bvid         string        `json:"bvid,omitempty"`
	Introduction *Introduction `json:"introduction,omitempty"`
	ReqUser      *ReqUser      `json:"req_user,omitempty"`
	Owner        *ViewOwner    `json:"owner,omitempty"`
	Button       *Button       `json:"button,omitempty"`
	Stat         *ViewStat     `json:"stat,omitempty"`
}

type PageWeb struct {
	Cid       int64      `json:"cid,omitempty"`
	Desc      string     `json:"desc,omitempty"`
	Dimension *Dimension `json:"dimension,omitempty"`
	Duration  int64      `json:"duration,omitempty"`
	Title     string     `json:"title,omitempty"`
	Part      string     `json:"part,omitempty"`
	BadgeInfo *BadgeInfo `json:"badge_info,omitempty"`
	ArcAid    int64      `json:"arc_aid,omitempty"`
	ArcCid    int64      `json:"arc_cid,omitempty"`
	ShareURL  string     `json:"share_url,omitempty"`
}

func (v *ViewWeb) FromViewUGC(a *api.ViewReply, his *hisApi.ModelHistory) {
	v.Title = a.Title
	v.Desc = a.Desc
	v.Duration = a.Duration
	v.Cover = a.Pic
	for _, vp := range a.Pages {
		p := &PageWeb{
			Cid:  vp.Cid,
			Desc: vp.Desc,
			Dimension: &Dimension{
				Width:  vp.Dimension.Width,
				Height: vp.Dimension.Height,
				Rotate: vp.Dimension.Rotate,
			},
			Duration: vp.Duration,
			Title:    vp.Part,
			ArcAid:   a.Aid,
			ArcCid:   vp.Cid,
		}
		// bvid
		if bvid, err := model.GetBvID(a.Aid); err == nil {
			p.ShareURL = model.FillURI(model.GotoWebBV, 0, 0, bvid, model.SuffixHandler(fmt.Sprintf("p=%d", vp.Page)))
		}
		v.Pages = append(v.Pages, p)
	}
	if his != nil {
		v.History = &History{
			Cid:      his.Cid,
			Progress: his.Pro,
		}
	}
	if bvid, err := model.GetBvID(a.Aid); err == nil {
		v.Bvid = bvid
	}
	v.Stat = &ViewStat{
		Reply: int64(a.Stat.Reply),
	}
}

func (v *ViewWeb) FromViewPGC(b *bangumi.View) {
	v.Title = b.Title
	v.SeasonType = b.Type
	v.SeasonTitle = b.SeasonTitle
	v.Cover = b.Cover
	v.Desc = b.Detail
	// 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
	if b.Type == 1 || b.Type == 4 {
		// 如果当前PGC第一个分P是番剧类型则全部展示宫格
		v.PagesStyle = GridStyle
	}
	if b.BadgeInfo != nil && b.BadgeInfo.Text != "" {
		v.BadgeInfo = reasonStyleFrom(model.PGCBageType[b.BadgeType], b.Badge)
	}
	if b.UserStatus != nil && b.UserStatus.Progress != nil {
		v.History = &History{
			Epid:     b.UserStatus.Progress.LastEpID,
			Progress: b.UserStatus.Progress.LastTime,
		}
	}
	positivepage := []*PageWeb{}
	sectionPage := []*PageWeb{}
	for _, m := range b.Modules {
		// positive正片、section其他
		switch m.Style {
		case "positive":
			for _, ep := range m.Data.Episodes {
				// 互动视频不展示
				if ep.Interaction != nil {
					continue
				}
				p := &PageWeb{}
				p.fromViewPagePGC(ep)
				positivepage = append(positivepage, p)
			}
		case "section":
			for _, ep := range m.Data.Episodes {
				// 互动视频不展示
				if ep.Interaction != nil {
					continue
				}
				p := &PageWeb{}
				p.fromViewPagePGC(ep)
				sectionPage = append(sectionPage, p)
			}
		default:
			continue
		}
	}
	// positive正片、section其他，先正片，后非正片
	v.Pages = positivepage
	v.Pages = append(v.Pages, sectionPage...)
	if b.Stat != nil {
		v.Stat = &ViewStat{
			Reply: b.Stat.Reply,
		}
	}
}

func (p *PageWeb) fromViewPagePGC(e *bangumi.Episodes) {
	p.Cid = e.ID
	p.ArcAid = e.Aid
	p.ArcCid = e.Cid
	p.Dimension = &Dimension{
		Width:  e.Dimension.Width,
		Height: e.Dimension.Height,
		Rotate: e.Dimension.Rotate,
	}
	p.Title = e.Title
	p.Part = e.LongTitle
	if e.LongTitle == "" {
		p.Part = e.Title
	}
	if e.BadgeInfo != nil && e.BadgeInfo.Text != "" {
		p.BadgeInfo = reasonStyleFrom(model.PGCBageType[e.BadgeType], e.Badge)
	}
	p.ShareURL = model.FillURI(model.GotoWebPGC, 0, 0, strconv.FormatInt(e.ID, 10), nil)
}

func (v *ViewWeb) FromViewOwnerWeb(stat int64, a *api.ViewReply, relations map[int64]*relationgrpc.InterrelationReply) {
	a.Access = 0
	a.Attribute = 0
	a.AttributeV2 = 0
	v.Owner = &ViewOwner{
		Mid:  a.Author.Mid,
		Name: a.Author.Name,
		Face: a.Author.Face,
		Fans: model.FanString(int32(stat)),
		RequestParam: &card.RequestParam{
			Vmid: a.Author.Mid,
		},
		SourceType: model.EntranceSpace,
		Relation:   model.RelationChange(a.Author.Mid, relations),
	}
}

func (v *ViewWeb) FromButton(gt string, selected int8) {
	switch gt {
	case model.GotoPGC:
		v.Button = &Button{
			Type:     gt,
			Selected: selected,
		}
	}
}
