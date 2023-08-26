package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"

	"github.com/pkg/errors"
)

const (
	_specialWeb       = 0
	_specialGame      = 1
	_specialAv        = 2
	_specialPGC       = 3
	_specialLive      = 4
	_specialArticle   = 6
	_specialPgcSeason = 14
)

func BuildSmallCoverV2FromSpecialS(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	specialCard, ok := fanoutResult.SpecialCard[item.ID]
	if !ok {
		return nil, errors.Errorf("special card not exist")
	}
	// 0:url 1:游戏小卡 2:稿件 3:PGC 4:直播 6:专栏 7:每日精选 8:歌单 9:歌曲 10:相簿 11:小视频 12:特殊小卡 14:PGC-seasion-id
	switch specialCard.ReType {
	case _specialAv:
		return buildSmallCoverV2SpecialFromAv(ctx, index, item, fanoutResult)
	case _specialPGC:
		return buildSmallCoverV2SpecialFromPGC(ctx, index, item, fanoutResult)
	case _specialLive:
		return buildSmallCoverV2SpecialFromLive(ctx, index, item, fanoutResult)
	case _specialArticle:
		return buildSmallCoverV2SpecialFromArticle(ctx, index, item, fanoutResult)
	case _specialPgcSeason:
		return buildSmallCoverV2SpecialFromPgcSeason(ctx, index, item, fanoutResult)
	case _specialWeb, _specialGame:
		return buildSmallCoverV2SpecialFromWeb(ctx, index, item, fanoutResult)
	default:
		return nil, errors.Errorf("Unsupported special card: %d", specialCard.ReType)
	}
}

func buildSmallCoverV2SpecialFromWeb(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.OperateType[int(fanoutResult.SpecialCard[item.ID].ReType)]).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveWebBuilder().
		SetBase(base).
		SetRcmd(item).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildSmallCoverV2SpecialFromPgcSeason(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "season" {
		return nil, errors.Errorf("special_s season not exist")
	}
	seasonid, err := strconv.ParseInt(fanoutResult.SpecialCard[item.ID].ReValue, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	season, ok := fanoutResult.Bangumi.PgcSeason[int32(seasonid)]
	if !ok {
		return nil, errors.Errorf("pgc season not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.OperateType[int(fanoutResult.SpecialCard[item.ID].ReType)]).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveSpecialSeasonBuilder().
		SetBase(base).
		SetRcmd(item).
		SetSeason(season).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildSmallCoverV2SpecialFromArticle(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "article" {
		return nil, errors.Errorf("special_s article not exist")
	}
	metaid, err := strconv.ParseInt(fanoutResult.SpecialCard[item.ID].ReValue, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	article, ok := fanoutResult.Article[metaid]
	if !ok {
		return nil, errors.Errorf("article not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.OperateType[int(fanoutResult.SpecialCard[item.ID].ReType)]).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveArticleBuilder().
		SetBase(base).
		SetRcmd(item).
		SetArticle(article).
		WithAfter(func(v2 *jsoncard.SmallCoverV2) {
			v2.Badge = ""
			v2.BadgeStyle = nil
			v2.DescButton = nil
			v2.Args = jsoncard.Args{}
			v2.ThreePointV2 = constructSpecialThreePointV2(ctx)
		}).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		//WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func buildSmallCoverV2SpecialFromLive(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "live" {
		return nil, errors.Errorf("special_s live not exist")
	}
	roomid, err := strconv.ParseInt(fanoutResult.SpecialCard[item.ID].ReValue, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	room, ok := fanoutResult.Live.Room[roomid]
	if !ok {
		return nil, errors.Errorf("live room: %d not exist", item.ID)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.OperateType[int(fanoutResult.SpecialCard[item.ID].ReType)]).
		SetCardLen(1).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	factory := jsonsmallcover.NewSmallCoverV2Builder(ctx)
	card, err := factory.DeriveLiveRoomBuilder().
		SetBase(base).
		SetRcmd(item).
		SetLiveRoom(room).
		SetAuthorCard(fanoutResult.Account.Card[room.UID]).
		WithAfter(func(v2 *jsoncard.SmallCoverV2) {
			v2.Badge = ""
			v2.BadgeStyle = nil
			v2.DescButton = nil
			v2.CanPlay = 0
			v2.Args = jsoncard.Args{}
			v2.PlayerArgs = nil
			v2.ThreePointV2 = constructSpecialThreePointV2(ctx)
		}).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		//WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func buildSmallCoverV2SpecialFromPGC(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "pgc" {
		return nil, errors.Errorf("special_s pgc not exist")
	}
	epid, err := strconv.ParseInt(fanoutResult.SpecialCard[item.ID].ReValue, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	episode, ok := fanoutResult.Bangumi.PgcEpisodeByEpids[int32(epid)]
	if !ok {
		return nil, errors.Errorf("episode not exist: %d", epid)
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.OperateType[int(fanoutResult.SpecialCard[item.ID].ReType)]).
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
		//WithAfter(jsonsmallcover.V2FilledByMultiMaterials(fanoutResult.MultiMaterials[item.CreativeId], item, true)).
		WithAfter(jsonsmallcover.V2ReplacedByRcmd(item)).
		WithAfter(func(v2 *jsoncard.SmallCoverV2) {
			v2.Badge = ""
			v2.BadgeStyle = nil
			v2.DescButton = nil
			v2.CoverLeftText1 = appcardmodel.StatString(int32(episode.Stat.Play), "")
			v2.CoverLeftText2 = appcardmodel.StatString(int32(episode.Stat.Follow), "")
			v2.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(v2.CoverLeftIcon1, v2.CoverLeftText1)
			v2.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(v2.CoverLeftIcon2, v2.CoverLeftText2)
			if fanoutResult.SpecialCard[item.ID].GetUrl() != "" {
				v2.URI = appcardmodel.FillURI("", 0, 0, fanoutResult.SpecialCard[item.ID].Url, appcardmodel.PGCTrackIDHandler(item))
			}
		}).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		//WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}

	return card, nil
}

func buildSmallCoverV2SpecialFromAv(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if item.SingleSpecialInfo == nil || item.SingleSpecialInfo.SpType != "av" {
		return nil, errors.Errorf("special_s av not exist")
	}
	archiveStore := fanoutResult.Archive.Archive
	aid, err := strconv.ParseInt(fanoutResult.SpecialCard[item.ID].ReValue, 10, 64)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	archive, ok := archiveStore[aid]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}

	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV2).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoAv).
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
		WithAfter(func(v2 *jsoncard.SmallCoverV2) {
			v2.GotoIcon = nil
			v2.DescButton = nil
			v2.OfficialIcon = 0
			v2.CanPlay = 0
			v2.Args = jsoncard.Args{}
			v2.PlayerArgs = nil
			v2.ThreePointV2 = constructSpecialThreePointV2(ctx)
		}).
		WithAfter(jsonsmallcover.SmallCoverV2FromSpecial(ctx, fanoutResult.SpecialCard[item.ID], item)).
		//WithAfter(jsonsmallcover.SmallCoverV2TalkBack()).
		Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func constructSpecialThreePointV2(ctx cardschema.FeedContext) []*jsoncard.ThreePointV2 {
	if ctx.VersionControl().Can("feed.usingNewThreePointV2") {
		return constructDefaultThreePointV2()
	}
	return constructDefaultThreePointV2Legacy(ctx)
}

func constructDefaultThreePointV2() []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	out = append(out, &jsoncard.ThreePointV2{
		Title: "不感兴趣",
		Type:  appcardmodel.ThreePointDislike,
		ID:    1,
	})
	return out
}

func constructDefaultThreePointV2Legacy(ctx cardschema.FeedContext) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	_, dislikeReasonToast, _ := dislikeText(ctx)
	out = append(out, &jsoncard.ThreePointV2{
		Reasons: []*jsoncard.DislikeReason{
			{ID: 1, Name: "不感兴趣", Toast: dislikeReasonToast},
		},
		Type: appcardmodel.ThreePointDislike,
	})
	return out
}

func dislikeText(ctx cardschema.FeedContext) (string, string, string) {
	dislikeSubTitle := "(选择后将减少相似内容推荐)"
	dislikeReasonToast := "将减少相似内容推荐"
	dislikeTitle := "不感兴趣"
	if ctx.FeatureGates().FeatureEnabled(cardschema.FeatureCloseRcmd) {
		dislikeSubTitle = ""
		dislikeReasonToast = ""
	}
	if ctx.FeatureGates().FeatureEnabled(cardschema.FeatureDislikeText) {
		dislikeTitle = "我不想看"
	}
	return dislikeSubTitle, dislikeReasonToast, dislikeTitle
}
