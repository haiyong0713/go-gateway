package service

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedmodel "go-gateway/app/app-svr/app-feed/interface/model"

	"github.com/pkg/errors"
)

var (
	defaultFanoutPreProcesser = &fanoutPreProcesser{
		gotoProcesser: map[string]func(*ai.Item, *feedFanoutLoader){},
	}
)

var (
	archiveGotoTypeSet = sets.NewString(feedmodel.GotoAv, feedmodel.GotoPlayer, feedmodel.GotoUpRcmdAv, feedmodel.GotoInlineAv, feedmodel.GotoInlineAvV2)
	liveGotoTypeSet    = sets.NewString(feedmodel.GotoLive, feedmodel.GotoPlayerLive)
	// 	specialGotoTypeSet   = sets.NewString(feedmodel.GotoSpecial, feedmodel.GotoSpecialS)
	avCovergeGotoTypeSet = sets.NewString(feedmodel.GotoAvConverge, feedmodel.GotoMultilayerConverge)
	directAdGotoTypeSet  = sets.NewString(feedmodel.GotoAdWebS, feedmodel.GotoAdWeb, feedmodel.GotoAdPlayer, feedmodel.GotoAdInlineGesture, feedmodel.GotoAdInline360, feedmodel.GotoAdInlineLive, feedmodel.GotoAdWebGif)
	potentialAdGifSet    = sets.NewString(feedmodel.GotoAdWebS, feedmodel.GotoAdWeb, feedmodel.GotoAdPlayer, feedmodel.GotoAdInlineGesture, feedmodel.GotoAdInline360, feedmodel.GotoAdInlineLive, feedmodel.GotoAdWebGif, feedmodel.GotoAdAv)
	inlineGotoSet        = sets.NewString(feedmodel.GotoInlineAv, feedmodel.GotoInlinePGC, feedmodel.GotoInlineLive, feedmodel.GotoInlineAvV2)
)

const (
	//_dynamicCoverRcmdGif    = 1 // 运营GIF
	_dynamicCoverAiGif      = 2 // AI GIF
	_dynamicCoverAdGif      = 3 // 广告 GIF
	_dynamicCoverAdInline   = 4 // 广告inline
	_dynamicCoverInlineAv   = 5 // inlineAv
	_dynamicCoverRcmdInline = 6 // 运营Inline
)

func init() {
	for _, gotoType := range archiveGotoTypeSet.List() {
		defaultFanoutPreProcesser.register(gotoType, archiveProcesser)
	}
	for _, gotoType := range liveGotoTypeSet.List() {
		defaultFanoutPreProcesser.register(gotoType, liveProcesser)
	}
	for _, gotoType := range avCovergeGotoTypeSet.List() {
		defaultFanoutPreProcesser.register(gotoType, avConvergeProcesser)
	}
	// for _, gotoType := range specialGotoTypeSet.List() {
	// 	defaultFanoutPreProcesser.addProcesser(gotoType, specialProcesser)
	// }

	defaultFanoutPreProcesser.register(feedmodel.GotoAiStory, storyProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoTunnel, tunnelProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoSpecialChannel, specialChannelProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoPlayerBangumi, playerBangumiProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoPicture, pictureProcesser)
	// defaultFanoutPreProcesser.addProcesser(feedmodel.GotoSpecialB, specialBProcesser)
	// defaultFanoutPreProcesser.addProcesser(feedmodel.GotoChannelRcmd, channelRcmdProcesser)
	// defaultFanoutPreProcesser.addProcesser(feedmodel.GotoLiveUpRcmd, liveUpProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoAudio, audioProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoArticleS, articleSProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoConvergeAi, convergeAiProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoBangumi, bangumiProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoPGC, pgcProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoAdAv, adAvProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoInlinePGC, inlinePgcProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoInlineLive, inlineLiveProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoShoppingS, shopProcesser)
	defaultFanoutPreProcesser.register(feedmodel.GotoInlineBangumi, inlinePgcProcesser)
}

type fanoutPreProcesser struct {
	gotoProcesser map[string]func(*ai.Item, *feedFanoutLoader)
}

func (fpp *fanoutPreProcesser) register(gotoType string, fn func(*ai.Item, *feedFanoutLoader)) {
	if _, ok := fpp.gotoProcesser[gotoType]; ok {
		panic(errors.Errorf("conflicate processer: %q", gotoType))
	}
	fpp.gotoProcesser[gotoType] = fn
}

func (fpp *fanoutPreProcesser) processRcmd(item ...*ai.Item) *feedFanoutLoader {
	out := &feedFanoutLoader{}
	for _, i := range item {
		fn, ok := fpp.gotoProcesser[i.Goto]
		if !ok {
			log.Error("Unrecognized ai goto: %q", i.Goto)
			continue
		}
		posRecPorcessor(i, out)

		fn(i, out)
	}
	return out
}

func storyProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.StoryInfo == nil {
		return
	}
	for _, i := range item.StoryInfo.Items {
		if i.Goto != feedmodel.GotoVerticalAv {
			continue
		}
		dst.WithStoryArchive(i.ID)
		dst.WithTag(i.Tid)
	}
}

func tunnelProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithTunnelFeed(item.ID)
}

func specialChannelProcesser(item *ai.Item, dst *feedFanoutLoader) {
	// panic("unimpl")
}

func avConvergeProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ConvergeInfo != nil {
		for _, i := range item.ConvergeInfo.Items {
			if i.Goto != feedmodel.GotoAv {
				continue
			}
			dst.WithArchive(i.ID)
		}
	}
	if item.JumpGoto == feedmodel.GotoAv {
		if item.JumpID != 0 {
			dst.WithArchive(item.JumpID)
		}
	}
	if item.Tid > 0 {
		dst.WithTag(item.Tid)
	}
}

func playerBangumiProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithBangumiEP(item.ID)
}

func pictureProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID > 0 {
		dst.WithPicture(item.ID)
	}
	if item.RcmdReason != nil && item.RcmdReason.Style == 4 {
		dst.WithAccountProfile(item.RcmdReason.FollowedMid)
	}
}

func audioProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithAudio(item.ID)
}

func articleSProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithArticle(item.ID)
}

func convergeAiProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ConvergeInfo == nil {
		return
	}
	maxAidLimit := 10
	aids := make([]int64, 0, len(item.ConvergeInfo.Items))
	for _, i := range item.ConvergeInfo.Items {
		if i.Goto != feedmodel.GotoAv {
			continue
		}
		aids = append(aids, i.ID)
	}
	if len(aids) > maxAidLimit {
		aids = aids[:maxAidLimit]
	}
	dst.WithArchive(aids...)
}

func bangumiProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID != 0 {
		dst.WithArchive(item.ID)
		dst.WithBangumiSeasonAid(int32(item.ID))
	}
	if item.Tid != 0 {
		dst.WithTag(item.Tid)
	}
}

func pgcProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithBangumiSeason(int32(item.ID))
}

func inlinePgcProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithBangumiPlayerIDs(int32(item.ID))
}

func liveProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID != 0 {
		dst.WithLiveRoom(item.ID)
	}
	if item.SingleInline == model.SingleInlineV1 {
		dst.WithInlineLiveRoom(item.ID)
	}
}

func inlineLiveProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID != 0 {
		dst.WithInlineLiveRoom(item.ID)
	}
}

func adAvProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID != 0 {
		dst.WithArchive(item.ID)
	}
	if item.Tid != 0 {
		dst.WithTag(item.Tid)
	}
}

func archiveProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID != 0 {
		func() {
			if item.JumpGoto == feedmodel.GotoVerticalAv {
				dst.WithStoryArchive(item.ID)
				return
			}
			if item.Goto == feedmodel.GotoInlineAv || item.Goto == feedmodel.GotoInlineAvV2 || isSingleInline(item) {
				dst.WithThumbUpArchive(item.ID)
				item.SetDynamicCoverInfoc(_dynamicCoverInlineAv)
			}
			if item.Goto == feedmodel.GotoInlineAvV2 || isSingleInline(item) {
				dst.WithFavourite(item.ID)
			}
			dst.WithArchive(item.ID)
		}()
	}
	if item.Tid != 0 {
		dst.WithTag(item.Tid)
	}
}

func isSingleInline(item *ai.Item) bool {
	return item.SingleInline == model.SingleInlineV1
}

func posRecPorcessor(item *ai.Item, dst *feedFanoutLoader) {
	if item.PosRecID != 0 {
		dst.WithPosRec(item.PosRecID)
	}
	if inlineGotoSet.Has(item.Goto) && item.PosRecID > 0 {
		item.SetDynamicCoverInfoc(_dynamicCoverRcmdInline)
	}
}

func shopProcesser(item *ai.Item, dst *feedFanoutLoader) {
	if item.ID <= 0 {
		return
	}
	dst.WithShop(item.ID)
}
