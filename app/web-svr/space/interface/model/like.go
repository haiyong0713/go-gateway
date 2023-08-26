package model

import (
	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

const (
	// thumbupgrpc.UserLikesReq.Business
	BusinessLike = "archive"
)

type LikeVideoReq struct {
	Vmid int64 `json:"vmid" form:"vmid" validate:"required"`
}

type LikeVideoRly struct {
	List []*LikeVideoItem `json:"list"`
}

type LikeVideoItem struct {
	*arcgrpc.Arc
	Bvid         string `json:"bvid"`
	InterVideo   bool   `json:"inter_video"`
	ResourceType string `json:"resource_type"`
	Subtitle     string `json:"subtitle"`
}

func (i *LikeVideoItem) FormatAsEpCard(ep *pgccardgrpc.EpisodeCard) {
	i.ResourceType = ResourceTypeOGV
	i.Subtitle = ep.GetMeta().GetShortLongTitle()
	i.Arc = &arcgrpc.Arc{
		Aid:         ep.GetAid(),
		Pic:         ep.GetCover(),
		Title:       ep.GetSeason().GetTitle(),
		Duration:    ep.GetDuration(),
		RedirectURL: ep.GetUrl(),
		Stat: arcgrpc.Stat{
			View:    int32(ep.GetStat().GetPlay()),
			Danmaku: int32(ep.GetStat().GetDanmaku()),
		},
	}
}
