package feed

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	"github.com/pkg/errors"
)

const (
	_ps           = 10
	_doubleColumn = 2
)

type VerticalTabReply struct {
	Items []card.Handler `json:"items"`
	Page  struct {
		Offset  int32 `json:"offset"`
		HasMore bool  `json:"has_more"`
	} `json:"page"`
	Config struct {
		Column       int8 `json:"column"`
		AutoplayCard int8 `json:"autoplay_card"`
	} `json:"config"`
}

func (s *Service) VerticalTab(ctx context.Context, param *feed.VerticalTabParam) (*VerticalTabReply, error) {
	vcp := &feed.VerticalChannelParam{
		ChannelID: param.ChannelID,
		Tag:       param.Tag,
		Mid:       param.Mid,
		Buvid:     param.Buvid,
		Offset:    param.Offset,
		Ps:        _ps,
		Ip:        metadata.String(ctx, metadata.RemoteIP),
	}
	verticalIndexReply, err := s.verticalIndex(ctx, vcp)
	if err != nil {
		return nil, err
	}
	out := &VerticalTabReply{
		Page: struct {
			Offset  int32 `json:"offset"`
			HasMore bool  `json:"has_more"`
		}{
			Offset:  verticalIndexReply.feed.GetOffset(),
			HasMore: verticalIndexReply.feed.GetHasMore(),
		},
		Config: struct {
			Column       int8 `json:"column"`
			AutoplayCard int8 `json:"autoplay_card"`
		}{
			Column:       _doubleColumn,
			AutoplayCard: 11,
		},
	}
	materials, err := s.doVerticalFanoutResult(ctx, param, verticalIndexReply.feed)
	if err != nil {
		return nil, err
	}
	out.Items = s.buildVerticalItems(ctx, materials, verticalIndexReply)
	return out, nil
}

func constructTag(tag *hmtchannelgrpc.TagReply) []*feed.VerticalTag {
	if len(tag.GetList()) == 0 {
		return []*feed.VerticalTag{}
	}
	out := []*feed.VerticalTag{
		{
			Key:        "",
			Title:      "全部",
			ServerInfo: "",
		},
	}
	for _, v := range tag.GetList() {
		out = append(out, &feed.VerticalTag{
			Key:        v.GetKey(),
			Title:      v.GetTitle(),
			ServerInfo: v.GetServerInfo(),
		})
	}
	return out
}

type verticalIndexReply struct {
	feed          *hmtchannelgrpc.ChannelFeedReply
	playlistWhite *hmtchannelgrpc.WhiteReply
}

func (s *Service) verticalIndex(ctx context.Context, param *feed.VerticalChannelParam) (*verticalIndexReply, error) {
	const (
		_playlistWhite = 2
	)
	out := &verticalIndexReply{}
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) (err error) {
		if out.feed, err = s.channelDao.ChannelFeed(ctx, param); err != nil {
			log.Error("Failed to request ChannelFeed: %+v", err)
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if out.playlistWhite, err = s.channelDao.ChannelWhite(ctx, param, _playlistWhite); err != nil {
			log.Error("Failed to request ChannelWhite: %+v", err)
		}
		return
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return out, nil
}

// nolint: gocognit
func (s *Service) doVerticalFanoutResult(ctx context.Context, param *feed.VerticalTabParam,
	feed *hmtchannelgrpc.ChannelFeedReply) (*Materials, error) {
	aids := make([]int64, 0)
	epids := make([]int32, 0)
	thumbupAids := make([]int64, 0)
	for _, v := range feed.GetList() {
		switch v.GetType() {
		case hmtchannelgrpc.ResourceType_UGC_RESOURCE:
			aids = append(aids, v.GetId())
			thumbupAids = append(thumbupAids, v.GetId())
		case hmtchannelgrpc.ResourceType_OGV_RESOURCE:
			epids = append(epids, int32(v.GetId()))
		default:
			log.Error("Failed to match type: %d", v.GetType())
			continue
		}
	}
	threePointMeta := &threePointMeta.ThreePointMetaText{
		WatchLater: "稍后再看",
	}
	// 三点处理
	if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
		i18n.TranslateAsTCV2(&threePointMeta.WatchLater)
	}
	out := &Materials{
		HotAidSet:          convertHotAid(s.hotAids),
		ThreePointMetaText: threePointMeta,
		IconList:           feed.GetIconList(),
	}
	uids := make([]int64, 0)
	g := errgroup.WithContext(ctx)
	if len(aids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if out.Archive, err = s.ArcsPlayer(ctx, aids); err != nil {
				return
			}
			for _, a := range out.Archive {
				uids = append(uids, a.Arc.Author.Mid)
			}
			if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
				for _, a := range out.Archive {
					i18n.TranslateAsTCV2(&a.Arc.Title, &a.Arc.Desc, &a.Arc.TypeName)
				}
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if param.Mid < 0 {
				return nil
			}
			if out.HasCoin, err = s.coin.ArchiveUserCoins(ctx, aids, param.Mid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if param.Mid < 0 {
				return nil
			}
			out.HasFavourite, err = s.fav.IsFavVideos(ctx, param.Mid, aids)
			if err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if out.InlinePGC, err = s.bgm.InlineCards(ctx, epids, param.MobiApp, param.Platform, param.Device,
				param.Build, param.Mid, false, false, false, param.Buvid, nil); err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, v := range out.InlinePGC {
				thumbupAids = append(thumbupAids, v.Aid)
			}
			if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
				for _, ep := range out.InlinePGC {
					i18n.TranslateAsTCV2(&ep.Season.Title, &ep.NewDesc, &ep.Season.TypeName, &ep.Season.NewEpShow)
				}
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if out.PgcEpisodeByEpids, err = s.bgm.EpCardsFromPgcByEpids(ctx, epids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	g2 := errgroup.WithContext(ctx)
	if len(uids) > 0 {
		g2.Go(func(ctx context.Context) (err error) {
			if out.AccountCard, err = s.acc.Cards3GRPC(ctx, uids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g2.Go(func(ctx context.Context) (err error) {
			if out.RelationStatMid, err = s.rel.StatsGRPC(ctx, uids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g2.Go(func(ctx context.Context) (err error) {
			out.IsAttention = s.acc.IsAttentionGRPC(ctx, uids, param.Mid)
			return nil
		})
	}
	g2.Go(func(ctx context.Context) (err error) {
		if out.HasLike, err = s.thumbupDao.HasLike(ctx, param.Buvid, param.Mid, thumbupAids); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := g2.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return out, nil
}

func (s *Service) buildVerticalItems(ctx context.Context, materials *Materials, vir *verticalIndexReply) []card.Handler {
	list := constructFeedList(vir.feed)
	out := make([]card.Handler, 0, len(list))
	isPlaylist := vir.playlistWhite.Val == "playlist"
	for k, v := range list {
		fakeItem := &ai.Item{ID: v.GetId(), Goto: fakeGoto(v.GetType(), v.GetShowType())}
		if v.GetReason() != "" {
			fakeItem.RcmdReason = &ai.RcmdReason{
				Content:    v.Reason,
				Style:      2,
				CornerMark: 2,
			}
		}
		fakeItem.SetIsPlaylist(isPlaylist)
		fn, ok := VerticalCardMap[fakeItem.Goto]
		if !ok {
			log.Error("Failed to match VerticalCardMap: %+v", v)
			continue
		}
		item, err := fn(fakeBuilderContext(ctx, materials.IsAttention), int64(k), fakeItem, materials, s.c.Feed, v.GetServerInfo())
		if err != nil {
			log.Error("Failed to build vertical card: %+v", err)
			continue
		}
		out = append(out, item)
	}
	return out
}

func constructFeedList(feed *hmtchannelgrpc.ChannelFeedReply) []*hmtchannelgrpc.Resource {
	out := []*hmtchannelgrpc.Resource{}
	if len(feed.GetIconList()) > 1 {
		out = append(out, &hmtchannelgrpc.Resource{
			Id:         0,
			Type:       99,
			ServerInfo: "",
		})
	}
	out = append(out, feed.GetList()...)
	return out
}

func fakeBuilderContext(ctx context.Context, follow map[int64]int8) cardschema.FeedContext {
	authn, _ := auth.FromContext(ctx)
	userSession := feedcard.NewUserSession(authn.Mid, follow, &feedcard.IndexParam{})
	dev, _ := device.FromContext(ctx)
	fCtx := feedcard.NewFeedContext(userSession, feedcard.NewCtxDevice(&dev), time.Now())
	return fCtx
}

func fakeGoto(type_ hmtchannelgrpc.ResourceType, showType hmtchannelgrpc.ShowType) string {
	var resourceTypeNav = 99
	if showType == hmtchannelgrpc.ShowType_LARGE_SHOW_TYPE {
		switch type_ {
		case hmtchannelgrpc.ResourceType_UGC_RESOURCE:
			return "inline_av"
		case hmtchannelgrpc.ResourceType_OGV_RESOURCE:
			return "inline_pgc"
		case hmtchannelgrpc.ResourceType(resourceTypeNav):
			return "navigation"
		default:
			return ""
		}
	}
	switch type_ {
	case hmtchannelgrpc.ResourceType_UGC_RESOURCE:
		return "av"
	case hmtchannelgrpc.ResourceType_OGV_RESOURCE:
		return "pgc"
	case hmtchannelgrpc.ResourceType(resourceTypeNav):
		return "navigation"
	default:
		return ""
	}
}

var VerticalCardMap = map[string]func(cardschema.FeedContext, int64, *ai.Item, *Materials, *conf.Feed, string) (cardschema.FeedCard, error){
	"av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
		return buildVerticalSmallCoverUGC(feedContext, i, item, materials, feed, serverInfo)
	},
	"pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
		return buildVerticalSmallCoverOGV(feedContext, i, item, materials, feed, serverInfo)
	},
	"navigation": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
		return buildVerticalNav(feedContext, i, item, materials, feed, serverInfo)
	},
	"inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
		return buildVerticalLargeCoverV9(feedContext, i, item, materials, feed, serverInfo)
	},
	"inline_pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
		return buildVerticalLargeCoverV7(feedContext, i, item, materials, feed, serverInfo)
	},
}

func buildVerticalNav(ctx cardschema.FeedContext, _ int64, item *ai.Item, materials *Materials, _ *conf.Feed, _ string) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(model.VerticalLargeCoverV11).
		SetCardGoto("navigation").
		SetGoto(model.GotoNavigation).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		Build()
	if err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverV11{
		Base: base,
		Item: make([]*jsoncard.NavItem, 0, len(materials.IconList)),
	}
	for _, icon := range materials.IconList {
		out.Item = append(out.Item, &jsoncard.NavItem{
			Name: icon.GetName(),
			Pic:  icon.GetPic(),
			URI:  icon.GetUrl(),
		})
	}
	return out, nil
}

func buildVerticalLargeCoverV9(ctx cardschema.FeedContext, _ int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
	fanoutResult := setFanoutResult(materials, feed)
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(model.VerticalLargeCoverV9).
		SetCardGoto(model.CardGt(item.Goto)).
		SetGoto(model.GotoAv).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveSingleArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		SetStoryIcon(fanoutResult.StoryIcon).
		SetHasFav(fanoutResult.Favourite).
		SetHotAidSet(fanoutResult.HotAidSet).
		SetHasCoin(fanoutResult.Coin).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			in.ThreePoint = nil
			in.ThreePointV2 = nil
			in.ThreePointMeta = threePointMeta.VerticalUGCThreePoint(fanoutResult.ThreePointMetaText)
			in.ServerInfo = serverInfo
		}).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildVerticalLargeCoverV7(ctx cardschema.FeedContext, _ int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
	fanoutResult := setFanoutResult(materials, feed)
	inlinePgc, ok := fanoutResult.Bangumi.InlinePGC[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("inline pgc: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(model.VerticalLargeCoverV7).
		SetCardGoto(model.CardGt(item.Goto)).
		SetGoto(model.GotoPGC).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(ctx)
	card, err := factory.DeriveSingleBangumiBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(inlinePgc).
		SetHasLike(fanoutResult.ThumbUp.HasLikeArchive).
		SetInline(fanoutResult.Inline).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			in.ThreePoint = nil
			in.ThreePointV2 = nil
			in.ThreePointMeta = threePointMeta.VerticalOGVThreePoint(fanoutResult.ThreePointMetaText)
			in.ServerInfo = serverInfo
			if in.SharePlane != nil {
				in.SharePlane.ShareFrom = "ogv_global_composite_tab_inline_normal_share"
			}
		}).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildVerticalSmallCoverUGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
	fanoutResult := setFanoutResult(materials, feed)
	archive, ok := fanoutResult.Archive.Archive[item.ID]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	// fake icon type
	item.IconType = model.AIUpIconType
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(model.VerticalSmallCoverV2).
		SetCardGoto(model.CardGt(item.Goto)).
		SetGoto(model.GotoAv).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveArcPlayerBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArcPlayer(archive).
		SetChannelCard(fanoutResult.Channel[item.Tid]).
		SetTag(fanoutResult.Tag[item.Tid]).
		SetAuthorCard(fanoutResult.Account.Card[archive.Arc.Author.Mid]).
		SetCoverGif(item.CoverGif).
		SetStoryIcon(fanoutResult.StoryIcon).
		SetOpenCourseMark(fanoutResult.OpenCourseMark).
		WithAfter(func(in *jsoncard.SmallCoverV2) {
			in.ThreePoint = nil
			in.ThreePointV2 = nil
			in.ThreePointMeta = threePointMeta.VerticalUGCThreePoint(fanoutResult.ThreePointMetaText)
			in.ServerInfo = serverInfo
			in.SharePlane = constructUGCSharePlane(archive)
			if item.IsPlaylist() && archive.Arc.AttrVal(api.AttrBitSteinsGate) != api.AttrYes { // 互动视频没有播单
				in.URI = constructMediaURI(archive)
			}
			if in.RcmdReason != "" && in.RcmdReasonStyle != nil {
				return
			}
			in.DescButton = &jsoncard.Button{
				Type:    model.ButtonGrey,
				Text:    archive.Arc.Author.Name,
				URI:     model.FillURI(model.GotoMid, 0, 0, strconv.FormatInt(archive.Arc.Author.Mid, 10), nil),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func constructMediaURI(archive *api.ArcPlayer) string {
	return fmt.Sprintf("bilibili://music/playlist/playpage/%d?oid=%d&from_spmid=%s&from_h5=%s&page_type=7&title=%s",
		archive.GetArc().GetAid(), archive.GetArc().GetAid(), "main.composite-tab.0.0", "main.composite-tab.0.0", "系统生成列表")
}

func constructUGCSharePlane(arcPlayer *api.ArcPlayer) *model.SharePlane {
	shareSubtitle, playNumber := card.GetShareSubtitle(arcPlayer.Arc.Stat.View)
	bvid_, _ := card.GetBvID(arcPlayer.Arc.Aid)
	return &model.SharePlane{
		Title:         arcPlayer.Arc.Title,
		ShareSubtitle: shareSubtitle,
		Desc:          arcPlayer.Arc.Desc,
		Cover:         arcPlayer.Arc.Pic,
		Aid:           arcPlayer.Arc.Aid,
		Bvid:          bvid_,
		ShareTo:       model.ShareTo,
		Author:        arcPlayer.Arc.Author.Name,
		AuthorId:      arcPlayer.Arc.Author.Mid,
		ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%d", arcPlayer.Arc.Aid),
		PlayNumber:    playNumber,
	}
}

func buildVerticalSmallCoverOGV(ctx cardschema.FeedContext, _ int64, item *ai.Item, materials *Materials, feed *conf.Feed, serverInfo string) (cardschema.FeedCard, error) {
	fanoutResult := setFanoutResult(materials, feed)
	episode, ok := fanoutResult.Bangumi.PgcEpisodeByEpids[int32(item.ID)]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(model.VerticalSmallCoverV2).
		SetCardGoto(model.CardGt(item.Goto)).
		SetGoto(model.Gt(item.Goto)).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveEpPGCBuilder().
		SetBase(base).
		SetRcmd(item).
		SetEpisode(episode).
		WithAfter(func(in *jsoncard.SmallCoverV2) {
			in.ThreePoint = nil
			in.ThreePointV2 = nil
			in.ThreePointMeta = threePointMeta.VerticalOGVThreePoint(fanoutResult.ThreePointMetaText)
			in.ServerInfo = serverInfo
			in.SharePlane = constructOGVSharePlane(episode)
		}).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func constructOGVSharePlane(episode *pgccard.EpisodeCard) *model.SharePlane {
	bvid, _ := bvid.AvToBv(episode.Aid)
	return &model.SharePlane{
		Desc:        episode.TianmaSmallCardMeta.RcmdReason,
		Cover:       episode.Cover,
		Aid:         episode.Aid,
		Bvid:        bvid,
		EpId:        episode.EpisodeId,
		SeasonId:    episode.GetSeason().GetSeasonId(),
		ShareTo:     model.ShareTo,
		ShareFrom:   "ogv_global_composite_tab_inline_normal_share",
		SeasonTitle: episode.Season.Title,
	}
}

type VerticalTagReply struct {
	Items []*feed.VerticalTag `json:"items"`
}

func (s *Service) VerticalTag(ctx context.Context, param *feed.VerticalTagParam) (*VerticalTagReply, error) {
	tag, err := s.channelDao.ChannelTag(ctx, param)
	if err != nil {
		log.Error("Failed to request ChannelTag: %+v", err)
		return nil, err
	}
	items := constructTag(tag)
	for _, item := range items {
		if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
			i18n.TranslateAsTCV2(&item.Title)
		}
	}
	out := &VerticalTagReply{
		Items: items,
	}
	return out, nil
}
