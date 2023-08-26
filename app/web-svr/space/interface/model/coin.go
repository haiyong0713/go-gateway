package model

import (
	v1 "git.bilibili.co/bapis/bapis-go/archive/service"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

const (
	ResourceTypeOGV   = "ogv"
	ResourceTypeUGC   = "ugc"
	ArcAccessVariable = 10000
)

// CoinArc coin archive.
type CoinArc struct {
	*v1.Arc
	Bvid         string `json:"bvid"`
	Coins        int64  `json:"coins"`
	Time         int64  `json:"time"`
	IP           string `json:"ip"`
	InterVideo   bool   `json:"inter_video"`
	ResourceType string `json:"resource_type"`
	Subtitle     string `json:"subtitle"`
}

func (ca *CoinArc) FormatAsEpCard(ep *pgccardgrpc.EpisodeCard) {
	ca.ResourceType = ResourceTypeOGV
	ca.Subtitle = ep.GetMeta().GetShortLongTitle()
	ca.Arc = &v1.Arc{
		Aid:         ep.GetAid(),
		Pic:         ep.GetCover(),
		Title:       ep.GetSeason().GetTitle(),
		Duration:    ep.GetDuration(),
		RedirectURL: ep.GetUrl(),
		Stat: v1.Stat{
			View:    int32(ep.GetStat().GetPlay()),
			Danmaku: int32(ep.GetStat().GetDanmaku()),
		},
	}
}
