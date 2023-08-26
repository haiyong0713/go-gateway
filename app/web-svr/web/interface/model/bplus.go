package model

import (
	"strconv"

	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	articlegMdl "git.bilibili.co/bapis/bapis-go/article/model"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
)

const (
	invalidTitle = "内容已失效"
)

type MaterialInfoReq struct {
	Aids       []int64  `form:"aids,split"`
	Bvids      []string `form:"bvids,split"`
	ArticleIDs []int64  `form:"article_ids,split"`
	EpIDs      []int32  `form:"ep_ids,split"`
}

type Dynamic struct {
	Archive []*Archive  `json:"archive,omitempty"`
	Article []*Article  `json:"article,omitempty"`
	PGC     []*PGCShare `json:"pgc,omitempty"`
}

type Article struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	TemplateID int32    `json:"template_id"`
	UpName     string   `json:"up_name"`
	ImgURLs    []string `json:"image_urls"`
	ViewNum    int64    `json:"view_num"`
	LikeNum    int64    `json:"like_num"`
	ReplyNum   int64    `json:"reply_num"`
}

type Archive struct {
	BVID     string `json:"bvid"`
	AID      int64  `json:"aid"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Pic      string `json:"pic"`
	Param    string `json:"param"`
	URI      string `json:"uri"`
	Goto     string `json:"goto"`
	Duration int64  `json:"duration"`
	UpName   string `json:"up_name"`
	View     int32  `json:"view"`
	Danmaku  int32  `json:"danmaku"`
}

type PGCShare struct {
	EpID     int32  `json:"ep_id"`
	Cover    string `json:"cover"`
	Title    string `json:"title"`
	Duration int32  `json:"duration"`
	View     int64  `json:"view"`
	Danmaku  int64  `json:"danmaku"`
	URL      string `json:"url"`
}

func (a *Article) FromArt(art *articlegMdl.Meta) {
	if art == nil {
		return
	}
	a.ID = art.ID
	a.Title = art.Title
	a.Summary = art.Summary
	a.TemplateID = art.TemplateID
	if art.Author != nil {
		a.UpName = art.Author.Name
	}
	a.ImgURLs = art.ImageURLs
	if art.Stats != nil {
		a.ViewNum = art.Stats.View
		a.LikeNum = art.Stats.Like
		a.ReplyNum = art.Stats.Reply
	}
}

func (a *Archive) FormArc(arc *archivegrpc.Arc) {
	if arc == nil {
		return
	}
	a.AID = arc.Aid
	a.Title = arc.Title
	a.Desc = arc.Desc
	a.Pic = arc.Pic
	a.Param = strconv.FormatInt(a.AID, 10)
	a.Goto = GotoAv
	a.URI = FillURI(a.Goto, a.Param, nil)
	a.Duration = arc.Duration
	a.UpName = arc.Author.Name
	a.View = arc.Stat.View
	a.Danmaku = arc.Stat.Danmaku
	if !arc.IsNormal() {
		a.Title = invalidTitle
		a.Pic = ""
		a.Duration = 0
	}
}

// nolint:gomnd
func (p *PGCShare) FromPgcShare(e *pgcShareGrpc.ShareMessageResBody) {
	if e == nil {
		return
	}
	p.EpID = e.EpId
	p.Cover = e.Cover
	p.Title = e.Title
	// PGC的播放时长是毫秒，需要和UGC统一转成秒
	p.Duration = e.Duration / 1000
	p.Danmaku = e.Dm
	p.View = e.View
	p.URL = e.Url
}
