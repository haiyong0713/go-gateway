package view

import (
	"context"

	appresource "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
)

type appResourceWrapper struct {
	appresource.AppResourceClient
}

func (apw *appResourceWrapper) CheckEntranceInfoc(ctx context.Context, in *appresource.CheckEntranceInfocRequest) (*appresource.CheckEntranceInfocReply, error) {
	return apw.AppResourceClient.CheckEntranceInfoc(ctx, in)
}

func (s *Service) onDemandViewDependency() dependency.ViewDependency {
	dep := dependency.ViewDependency{}
	dep.Archive = s.arcDao
	dep.ArchiveHornor = s.ahDao
	dep.Account = s.accDao
	dep.Relation = s.relDao
	dep.ThumbUP = s.thumbupDao
	dep.Fav = s.favDao
	dep.Coin = s.coinDao
	dep.Danmu = s.dmDao
	dep.History = s.arcDao
	dep.Reply = s.replyDao
	dep.Audio = s.audioDao
	dep.UGCPayRank = s.elcDao
	dep.UGCPay = s.ugcpayDao
	dep.Assist = s.assDao
	dep.Garb = s.garbDao
	dep.UpArc = s.arcDao
	dep.Live = s.liveDao
	dep.Steins = s.steinDao
	dep.PGC = s.banDao
	dep.Location = s.locDao
	dep.Resource = s.rscDao
	dep.AppResource = &appResourceWrapper{s.appResourceClient}
	dep.VideoUP = s.vuDao
	dep.UGCSeason = s.seasonDao
	dep.Channel = s.channelDao
	dep.Activity = s.actDao
	dep.NatPage = s.natDao
	dep.Manager = s.mngDao
	dep.AD = s.adDao
	dep.Game = s.gameDao
	dep.Music = s.musicDao
	dep.ArchiveExtra = s.aeDao
	return dep
}
