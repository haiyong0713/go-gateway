package feed

import (
	"strconv"

	"go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

func (i *Item) FromPlayerAv(a *arcgrpc.Arc) {
	if i.Title == "" {
		i.Title = a.Title
	}
	if i.Cover == "" {
		i.Cover = model.CoverURLHTTPS(a.Pic)
	} else {
		i.Cover = model.CoverURLHTTPS(i.Cover)
	}
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.Goto = model.GotoAv
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	i.Cid = a.FirstCid
	i.Rid = a.TypeID
	i.TName = a.TypeName
	i.Desc = strconv.Itoa(int(a.Stat.Danmaku)) + "弹幕"
	i.fillArcStat(a)
	i.Duration = a.Duration
	i.Mid = a.Author.Mid
	i.Name = a.Author.Name
	i.Face = a.Author.Face
	i.CTime = a.PubDate
	i.Cid = a.FirstCid
	i.Autoplay = a.Rights.Autoplay
}
