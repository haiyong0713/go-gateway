package model

import (
	epCardSvc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	ogvEpisodeSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	ogvSeasonSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

type EpisodeDetail struct {
	Ep *ogvEpisodeSvc.EpisodeInfoProto
}

func (ed EpisodeDetail) IsSteinsGate() bool {
	return ed.Ep.GetAttr()&int32(1) == 1
}

func (ed EpisodeDetail) IsValid() bool {
	if ed.Ep == nil {
		return false
	}
	if ed.Ep.IsDelete != 0 || ed.Ep.Published != 1 {
		return false
	}
	return true
}

type SeasonDetail struct {
	Ss *ogvSeasonSvc.ProfileInfoProto
}

func (sd SeasonDetail) ComposeTitleCover(ep EpisodeDetail, arc ArchiveInfo) (title, cover string) {
	if sd.Ss == nil || ep.Ep == nil {
		return arc.Arc.GetTitle(), arc.Arc.GetPic()
	}
	title = sd.Ss.Title + " " + ep.Ep.ShowTitle
	if len(ep.Ep.ShowTitle) == 0 {
		title += ep.Ep.IndexTitle
	}
	cover = ep.Ep.Cover
	return
}

type EpCard struct {
	Ec *epCardSvc.EpisodeCard
}

func (ec EpCard) IsSteinsGate() bool {
	return ec.Ec.GetType().GetIsInteractive()
}

func (ec EpCard) IsValid() bool {
	if ec.Ec == nil {
		return false
	}
	if ec.Ec.GetIsDelete() != 0 {
		return false
	}
	return true
}
