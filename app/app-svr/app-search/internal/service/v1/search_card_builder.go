package v1

import (
	"context"

	"go-gateway/app/app-svr/app-search/internal/model"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

func buildTopGameInlineProcess(ctx context.Context, fanout *FanoutResult, data *search.TopGameData) (func(i *search.Item), bool) {
	if fanout == nil || (data.RoomId == 0 && data.Avid == 0) {
		return nil, false
	}
	if data.Avid > 0 {
		ugcInlineParams := &search.OptUGCInlineFnParams{
			Archive:  fanout.Archive.Archive[data.Avid],
			UserInfo: fanout.Account.Card[data.GameOfficialAccount],
			Follow:   fanout.Account.IsAttention,
			HasLike:  fanout.ThumbUp.HasLikeArchive,
			HasFav:   fanout.Favourite,
			HasCoin:  fanout.Coin,
		}
		if ugcInlineParams.Archive == nil {
			return nil, false
		}
		return search.OptUGCInlineFn(ctx, ugcInlineParams, search.TypeTopGame), true
	}
	return nil, false
}

func buildSportsInlineProcess(ctx context.Context, fanout *FanoutResult, data *esportsservice.QueryCardInfo) (func(i *search.Item), bool) {
	if fanout == nil {
		return nil, false
	}
	if data.UpMid > 0 {
		if liveInlineFn := search.OptLiveRoomInlineFn(ctx, fanout.Live.InlineRoom[data.UpMid], fanout.Account.Card[data.UpMid], fanout.Account.IsAttention, nil, search.TypeSports, model.SearchEsInlineCard, fanout.NftRegion); liveInlineFn != nil {
			return liveInlineFn, true
		}
	}
	if data.AvId > 0 {
		archive, ok := fanout.Archive.Archive[data.AvId]
		if !ok {
			return nil, false
		}
		ugcInlineParams := &search.OptUGCInlineFnParams{
			Archive:  archive,
			UserInfo: fanout.Account.Card[archive.Arc.Author.Mid],
			Follow:   fanout.Account.IsAttention,
			HasLike:  fanout.ThumbUp.HasLikeArchive,
			HasFav:   fanout.Favourite,
			HasCoin:  fanout.Coin,
		}
		if ugcInlineParams.Archive == nil {
			return nil, false
		}
		return search.OptUGCInlineFn(ctx, ugcInlineParams, model.GotoSports), true
	}
	return nil, false
}

func buildBrandAdAvInlineProcess(ctx context.Context, adContent *search.ADContent, apm map[int64]*arcgrpc.ArcPlayer, accCards map[int64]*account.Card, follows map[int64]bool, hasLike map[int64]thumbupgrpc.State, hasFav map[int64]int8, hasCoin map[int64]int64,
	nftRegion map[int64]*gallerygrpc.NFTRegion) (func(i *search.Item), bool) {
	if adContent == nil || len(adContent.Aids) == 0 {
		return nil, false
	}
	ugcInlineParams := &search.OptUGCInlineFnParams{
		Archive:   apm[adContent.Aids[0]],
		UserInfo:  accCards[adContent.UPMid],
		Follow:    follows,
		HasLike:   hasLike,
		HasFav:    hasFav,
		HasCoin:   hasCoin,
		NftRegion: nftRegion,
	}
	if avInlineFn := search.OptUGCInlineFn(ctx, ugcInlineParams, model.GotoBrandAdAv); avInlineFn != nil {
		return avInlineFn, true
	}
	return nil, false
}

func buildBrandAdLiveInlineProcess(ctx context.Context, adContent *search.ADContent, accCards map[int64]*account.Card, entryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, follows map[int64]bool,
	nftRegion map[int64]*gallerygrpc.NFTRegion) (func(i *search.Item), bool) {
	if adContent == nil {
		return nil, false
	}
	if liveInlineFn := search.OptLiveRoomInlineFn(ctx, entryRoom[adContent.UPMid], accCards[adContent.UPMid], follows, nil, model.GotoBrandAdLive, model.SearchLiveInlineCard, nftRegion); liveInlineFn != nil {
		return liveInlineFn, true
	}
	return nil, false
}
