package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	largecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

func fakeBuilderContext(ctx context.Context, follow map[int64]int32) cardschema.FeedContext {
	authn, _ := auth.FromContext(ctx)
	attentionStore := make(map[int64]int8, len(follow))
	for fid, followed := range follow {
		attentionStore[fid] = int8(followed)
	}
	userSession := feedcard.NewUserSession(authn.Mid, attentionStore, &feedcard.IndexParam{})
	dev, _ := device.FromContext(ctx)
	fCtx := feedcard.NewFeedContext(userSession, feedcard.NewCtxDevice(&dev), time.Now())
	return fCtx
}

func buildTopicAvBasicCards(ctx context.Context, aids []int64, fanout *FanoutResult) ([]*model.VideoCard, error) {
	var res []*model.VideoCard
	for _, aid := range aids {
		fakeRcmd := &ai.Item{ID: aid}
		builderCtx := fakeBuilderContext(ctx, fanout.Account.IsAttention)
		archive, ok := fanout.Archive.Archive[aid]
		if !ok {
			return nil, errors.New(fmt.Sprintf("buildTopicAvInlineCard Invalid TopicInlineResRsp ResId=%d", aid))
		}
		// fake base
		base, err := jsonbuilder.NewBaseBuilder(builderCtx).
			SetParam(strconv.FormatInt(aid, 10)).
			SetCardType(appcardmodel.SmallCoverV2).
			SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoAv)). // 客户端通过cardGt区分卡片
			SetGoto(appcardmodel.GotoAv).
			SetMetricRcmd(fakeRcmd).
			Build()
		if err != nil {
			return nil, err
		}

		factory := jsonsmallcover.NewSmallCoverV2Builder(builderCtx)
		card, err := factory.DeriveArcPlayerBuilder().
			SetBase(base).
			SetRcmd(fakeRcmd).
			SetArcPlayer(archive).
			SetAuthorCard(fanout.Account.Card[archive.Arc.Author.Mid]).
			Build()
		if err != nil {
			return nil, err
		}
		res = append(res, &model.VideoCard{SmallCoverV2: card})
	}
	return res, nil
}

//nolint:unparam
func buildTopicAvInlineCards(ctx context.Context, general *topiccardmodel.GeneralParam, aids []int64, fanout *FanoutResult) ([]*model.VideoInlineCard, error) {
	var res []*model.VideoInlineCard
	for _, v := range aids {
		card, err := buildTopicAvInlineCard(ctx, general, &topicsvc.TopicInlineResRsp{ResId: v}, fanout, &model.CommonDetailsParams{Source: "/general/feed/list"})
		if err != nil {
			continue
		}
		res = append(res, &model.VideoInlineCard{LargeCoverInline: card})
	}
	return res, nil
}

func buildTopicOgvInlineCard(ctx context.Context, rawResource *topicsvc.TopicInlineResRsp, fanout *FanoutResult) (*jsoncard.LargeCoverInline, error) {
	fakeRcmd := &ai.Item{ID: rawResource.ResId}
	builderCtx := fakeBuilderContext(ctx, nil)
	inlinePgc, ok := fanout.Bangumi.InlinePGC[int32(rawResource.ResId)]
	if !ok {
		return nil, errors.New(fmt.Sprintf("buildTopicOgvInlineCard Invalid TopicInlineResRsp ResId=%d", rawResource.ResId))
	}

	// fake base
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(fakeRcmd.ID, 10)).
		SetCardType(appcardmodel.LargeCoverV7).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoPGC)).
		SetGoto(appcardmodel.GotoPGC).
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return nil, err
	}

	factory := largecover.NewLargeCoverInlineBuilder(builderCtx)
	card, err := factory.DerivePgcBuilder().
		SetBase(base).
		SetRcmd(fakeRcmd).
		SetEpisode(inlinePgc).
		SetInline(&largecover.Inline{}).
		SetHasLike(fanout.ThumbUp.HasLikeArchive).
		WithAfter(updateOgvContent(inlinePgc)).
		WithAfter(doConfigOperationCard(rawResource)).
		WithAfter(hideUselessCardPart()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func buildTopicAvInlineCard(ctx context.Context, general *topiccardmodel.GeneralParam, rawResource *topicsvc.TopicInlineResRsp, fanout *FanoutResult, req *model.CommonDetailsParams) (*jsoncard.LargeCoverInline, error) {
	fakeRcmd := &ai.Item{ID: rawResource.ResId, JumpGoto: appfeedmodel.GotoAv}
	builderCtx := fakeBuilderContext(ctx, fanout.Account.IsAttention)
	archive, ok := fanout.Archive.Archive[rawResource.ResId]
	if !ok {
		return nil, errors.New(fmt.Sprintf("buildTopicAvInlineCard Invalid TopicInlineResRsp ResId=%d", rawResource.ResId))
	}

	// fake base
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(rawResource.ResId, 10)).
		SetCardType(appcardmodel.LargeCoverV9).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoAv)). // 客户端通过cardGt区分卡片
		SetGoto(appcardmodel.GotoAv).
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return nil, err
	}

	factory := largecover.NewLargeCoverInlineBuilder(builderCtx)
	card, err := factory.DeriveArcPlayerV2Builder().
		SetBase(base).
		SetRcmd(fakeRcmd).
		SetArcPlayer(archive).
		SetAuthorCard(fanout.Account.Card[archive.Arc.Author.Mid]).
		SetHasLike(fanout.ThumbUp.HasLikeArchive).
		SetHasCoin(fanout.Coin).
		SetHasFav(fanout.Favourite).
		SetInline(&largecover.Inline{}).
		WithAfter(updateStoryUriWithDynId(general, fanout.Archive.DynamicId[archive.Arc.Aid], archive, req)).
		WithAfter(doConfigOperationCard(rawResource)).
		WithAfter(hideUselessCardPart()).
		WithAfter(updateNativeWebDesc(archive.Arc.PubDate.Time(), req)).
		Build()

	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildTopicLiveInlineCard(ctx context.Context, rawResource *topicsvc.TopicInlineResRsp, fanout *FanoutResult) (*jsoncard.LargeCoverInline, error) {
	fakeRcmd := &ai.Item{ID: rawResource.ResId}
	builderCtx := fakeBuilderContext(ctx, fanout.Account.IsAttention)
	liveRoom, ok := fanout.Live.InlineRoom[rawResource.ResId]
	if !ok {
		return nil, errors.New(fmt.Sprintf("buildTopicLiveInlineCard Invalid TopicInlineResRsp ResId=%d", rawResource.ResId))
	}

	// fake base
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(rawResource.ResId, 10)).
		SetCardType(appcardmodel.LargeCoverV8).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoLive)). // 客户端通过cardGt区分卡片类型
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return nil, err
	}

	factory := largecover.NewLargeCoverInlineBuilder(builderCtx)
	card, err := factory.DeriveLiveEntryRoomBuilder().
		SetBase(base).
		SetRcmd(fakeRcmd).
		SetLiveRoom(liveRoom).
		SetAuthorCard(fanout.Account.Card[liveRoom.Uid]).
		SetEntryFrom(_newTopicLiveEntry).
		SetInline(&largecover.Inline{}).
		WithAfter(doConfigOperationCard(rawResource)).
		WithAfter(hideUselessCardPart()).
		WithAfter(addTopicCoverRightContentDescription(liveRoom)).
		Build()

	if err != nil {
		return nil, err
	}

	return card, nil
}

func hideUselessCardPart() func(*jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		// 隐藏三点结构
		in.ThreePoint = nil
		in.ThreePointV2 = nil
		in.ThreePointV3 = nil
		// 无天马分享面板
		in.SharePlane = nil
	}
}

func doConfigOperationCard(resource *topicsvc.TopicInlineResRsp) func(*jsoncard.LargeCoverInline) {
	// 话题侧配置影响卡片
	return func(in *jsoncard.LargeCoverInline) {
		if resource.TitleShowState == 0 {
			in.Base.Title = ""
		}
	}
}

func updateOgvContent(inlinePgc *pgcinline.EpisodeCard) func(*jsoncard.LargeCoverInline) {
	// 有一些话题详情页和天马卡不同的ogv卡片在这里解决
	return func(in *jsoncard.LargeCoverInline) {
		in.Base.Title = inlinePgc.ShowTitle
	}
}

func updateNativeWebDesc(pubDate time.Time, req *model.CommonDetailsParams) func(*jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		if req.Source != _sourceFromH5Details || pubDate.IsZero() {
			return
		}
		in.Desc = pubDate.Format("01-02")
	}
}

func updateStoryUriWithDynId(general *topiccardmodel.GeneralParam, dynId int64, archive *archivegrpc.ArcPlayer, req *model.CommonDetailsParams) func(*jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		if dynId == 0 || req.Source != "" || general.IsPad() || general.IsPadHD() || general.IsAndroidHD() {
			return
		}
		// 同时更新URI和ExtraURI,与dynHandler一致
		uri := topiccardmodel.FillURI(topiccardmodel.GotoStory, strconv.FormatInt(in.Args.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(archive, 0, true))
		cardUri := topiccardmodel.FillURI(topiccardmodel.GotoURL, uri, topiccardmodel.SuffixHandler(topiccardmodel.MakeStorySuffixUrl(in.Args.UpID, dynId, req.TopicId, req.SortBy, req.Offset, "")))
		in.Base.URI = cardUri
		in.ExtraURI = cardUri
	}
}

func addTopicCoverRightContentDescription(room *livexroomgate.EntryRoomInfoResp_EntryList) func(*jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		if room == nil || room.WatchedShow == nil {
			return
		}
		in.CoverRightContentDescription = room.WatchedShow.TextLarge
	}
}
