package card

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/app-card/interface/model/stat"

	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	season "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
)

func singleHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, tagm map[int64]*taggrpc.Tag, isAttenm, hasLike map[int64]int8, statm map[int64]*relationgrpc.StatReply, cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply) (hander Handler) {
	base := &Base{CardType: cardType, CardGoto: cardGoto, Rcmd: rcmd, Tagm: tagm, IsAttenm: isAttenm, HasLike: hasLike, Statm: statm, Cardm: cardm, Columnm: model.ColumnSvrSingle, AuthorRelations: authorRelations}
	if rcmd != nil {
		base.fillRcmdMeta(rcmd)
	}
	switch cardType {
	case model.LargeCoverV1:
		hander = &LargeCoverV1{Base: base}
	case model.OnePicV1:
		hander = &OnePicV1{Base: base}
	case model.ThreePicV1:
		hander = &ThreePicV1{Base: base}
	case model.SmallCoverV5:
		hander = &SmallCoverV5{Base: base}
	case model.SmallCoverH5:
		hander = &SmallCoverH5{Base: base}
	case model.SmallCoverH6:
		hander = &SmallCoverH6{Base: base}
	case model.SmallCoverH7:
		hander = &SmallCoverH7{Base: base}
	case model.OptionsV1:
		hander = &Option{Base: base}
	case model.Select:
		hander = &Select{Base: base}
	case model.BannerV4:
		hander = &Banner{Base: base}
	case model.BannerV4169:
		hander = &Banner{Base: base}
	case model.SmallCoverV8:
		hander = &SmallCoverV8{Base: base}
	case model.Introduction:
		hander = &Introduction{Base: base}
	case model.BannerSingleV8:
		hander = &BannerV8{Base: base}
	case model.LargeCoverSingleV9:
		hander = &LargeCoverInline{Base: base}
	case model.CmSingleV1:
		hander = &LargeCoverV1{Base: base}
	case model.LargeCoverSingleV8:
		hander = &LargeCoverInline{Base: base}
	case model.LargeCoverSingleV7:
		hander = &LargeCoverInline{Base: base}
	default:
		switch cardGoto {
		case model.CardGotoAv, model.CardGotoBangumi, model.CardGotoLive, model.CardGotoPlayer, model.CardGotoPlayerLive, model.CardGotoChannelRcmd, model.CardGotoUpRcmdAv,
			model.CardGotoPGC, model.CardGotoPlayerBangumi, model.CardGotoAvConverge, model.CardGotoSpecialB:
			base.CardType = model.LargeCoverV1
			hander = &LargeCoverV1{Base: base}
		case model.CardGotoAudio, model.CardGotoBangumiRcmd, model.CardGotoGameDownloadS, model.CardGotoShoppingS, model.CardGotoSpecialS, model.CardGotoMoe:
			base.CardType = model.SmallCoverV1
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoSpecial:
			base.CardType = model.MiddleCoverV1
			hander = &MiddleCover{Base: base}
		case model.CardGotoConverge, model.CardGotoRank, model.CardGotoConvergeAi:
			base.CardType = model.ThreeItemV1
			hander = &ThreeItemV1{Base: base}
		case model.CardGotoSubscribe, model.CardGotoSearchSubscribe:
			base.CardType = model.ThreeItemHV1
			hander = &ThreeItemH{Base: base}
		case model.CardGotoArticleS:
			base.CardType = model.ThreeItemHV3
			hander = &ThreeItemHV3{Base: base}
		case model.CardGotoLiveUpRcmd:
			base.CardType = model.TwoItemV1
			hander = &TwoItemV1{Base: base}
		case model.CardGotoBanner:
			base.CardType = model.BannerV1
			hander = &Banner{Base: base}
		case model.CardGotoAdAv, model.CardGotoAdPlayer, model.CardGotoAdInlineGesture, model.CardGotoAdInline360, model.CardGotoAdInlineLive, model.CardGotoAdWebGif, model.CardGotoAdInlineChoose, model.CardGotoAdInlineChooseTeam:
			base.CardType = model.CmV1
			hander = &LargeCoverV1{Base: base}
		case model.CardGotoAdWebGifReservation, model.CardGotoAdPlayerReservation:
			base.CardType = model.CmSingleV1
			hander = &LargeCoverV1{Base: base}
		case model.CardGotoAdWebS, model.CardGotoAdWeb:
			base.CardType = model.CmV1
			if cardGoto == model.CardGotoAdWebS && rcmd != nil && rcmd.SingleAdNew == 1 {
				base.CardType = model.CmSingleV1
			}
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoTopstick:
			base.CardType = model.TopStick
			hander = &Topstick{Base: base}
		case model.CardGotoChannelSquare:
			base.CardType = model.ChannelSquare
			hander = &ChannelSquare{Base: base}
		case model.CardGotoPgcsRcmd:
			base.CardType = model.ThreeItemHV4
			hander = &ThreeItemHV4{Base: base}
		case model.CardGotoUpRcmdS:
			base.CardType = model.UpRcmdCover
			hander = &UpRcmdCover{Base: base}
		case model.CardGotoSearchUpper:
			base.CardType = model.ThreeItemAll
			hander = &ThreeItemAll{Base: base}
		case model.CardGotoUpRcmdNew:
			base.CardType = model.TwoItemHV1
			hander = &TwoItemHV1{Base: base}
		case model.CardGotoEventTopic:
			base.CardType = model.MiddleCoverV3
			hander = &MiddleCoverV3{Base: base}
		case model.CardGotoVipRenew, model.CardGotoTunnel:
			base.CardType = model.SmallCoverV6
			hander = &SmallCoverV6{Base: base}
		case model.CardGotoNewTunnel:
			base.CardType = model.NotifyTunnelSingleV1
			hander = &UniversalNotifyTunnelV1{Base: base}
		case model.CardGotoBigTunnel:
			base.CardType = model.NotifyTunnelLargeSingleV1
			hander = &UniversalNotifyTunnelLargeV1{Base: base}
		case model.CardGotoMultilayerConverge:
			base.CardType = model.SmallCoverConvergeV1
			hander = &SmallCoverConvergeV1{Base: base}
		case model.CardGotoChannelNew:
			base.CardType = model.ChannelNew
			hander = &ChannelNew{Base: base}
		case model.CardGotoSpecialChannel:
			base.CardType = model.LargeCoverChannle
			hander = &LargeChannelSpecial{Base: base}
		case model.CardGotoChannelNewDetailCustom:
			base.CardType = model.ChannelThreeItemHV1
			hander = &ChannelThreeItemHV1{Base: base}
		case model.CardGotoChannelNewDetailRank:
			base.CardType = model.ChannelThreeItemHV2
			hander = &ChannelThreeItemHV2{Base: base}
		case model.CardGotoChannelScaned:
			base.CardType = model.ChannelScaned
			hander = &ChannelScaned{Base: base}
		case model.CardGotoChannelRcmdV2:
			base.CardType = model.ChannelRcmdV2
			hander = &ChannelRcmdV2{Base: base}
		case model.CardGotoChannelOGV, model.CardGotoChannelOGVLarge:
			base.CardType = cardType
			hander = &ChannelOGV{Base: base}
		case model.CardGotoAiStory:
			base.CardType = model.StorysV1
			hander = &Storys{Base: base}
		case model.CardGotoAdInlineAv:
			base.CardType = model.CmSingleV9
			hander = &LargeCoverInline{Base: base}
		case model.CardGotoAdInlinePgc:
			base.CardType = model.CmSingleV7
			hander = &LargeCoverInline{Base: base}
		default:
			log.Error("Fail to build handler, rowType=%s cardType=%s cardGoto=%s ai={%+v}",
				stat.RowTypeSingle, string(cardType), string(cardGoto), rcmd)
		}
	}
	stat.MetricAppCardTotal.Inc(stat.RowTypeSingle, string(base.CardType), string(cardGoto))
	return
}

type LargeCoverV1 struct {
	*Base
	CoverGif               string           `json:"cover_gif,omitempty"`
	Avatar                 *Avatar          `json:"avatar,omitempty"`
	CoverLeftText1         string           `json:"cover_left_text_1,omitempty"`
	CoverLeftText2         string           `json:"cover_left_text_2,omitempty"`
	CoverLeftText3         string           `json:"cover_left_text_3,omitempty"`
	CoverBadge             string           `json:"cover_badge,omitempty"`
	CoverBadgeStyle        *ReasonStyle     `json:"cover_badge_style,omitempty"`
	TopRcmdReason          string           `json:"top_rcmd_reason,omitempty"`
	BottomRcmdReason       string           `json:"bottom_rcmd_reason,omitempty"`
	Desc                   string           `json:"desc,omitempty"`
	OfficialIcon           model.Icon       `json:"official_icon,omitempty"`
	CanPlay                int32            `json:"can_play,omitempty"`
	CoverBadgeColor        model.CoverColor `json:"cover_badge_color,omitempty"`
	TopRcmdReasonStyle     *ReasonStyle     `json:"top_rcmd_reason_style,omitempty"`
	BottomRcmdReasonStyle  *ReasonStyle     `json:"bottom_rcmd_reason_style,omitempty"`
	RcmdReasonStyleV2      *ReasonStyle     `json:"rcmd_reason_style_v2,omitempty"`
	LeftCoverBadgeStyle    *ReasonStyle     `json:"left_cover_badge_style,omitempty"`
	RightCoverBadgeStyle   *ReasonStyle     `json:"right_cover_badge_style,omitempty"`
	CoverBadge2            string           `json:"cover_badge_2,omitempty"`
	CoverBadgeStyle2       *ReasonStyle     `json:"cover_badge_style_2,omitempty"`
	TitleSingleLine        int              `json:"title_single_line,omitempty"`
	CoverRightText         string           `json:"cover_right_text,omitempty"`
	LeftCoverBadgeNewStyle *ReasonStyle     `json:"left_cover_badge_new_style,omitempty"`
	// story详情页需要的封面图
	FfCover string `json:"ff_cover,omitempty"`
	// 跳转story图标
	GotoIcon       *model.GotoIcon `json:"goto_icon,omitempty"`
	CoverLeftIcon1 model.Icon      `json:"cover_left_icon_1,omitempty"`
	CoverLeftIcon2 model.Icon      `json:"cover_left_icon_2,omitempty"`
	OgvCreativeId  int64           `json:"ogv_creative_id,omitempty"`
}

//nolint:gocognit
func (c *LargeCoverV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button    interface{}
		avatar    *AvatarStatus
		upID      int64
		isBangumi bool
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		if isBangumi = a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && op.RedirectURL != ""; isBangumi {
			c.URI = model.FillURI("", 0, 0, op.RedirectURL, model.PGCTrackIDHandler(c.Rcmd))
		}
		c.CoverLeftText1 = model.DurationString(a.Arc.Duration)
		c.CoverLeftText2 = model.ArchiveViewString(a.Arc.Stat.View)
		c.CoverLeftText3 = model.DanmakuString(a.Arc.Stat.Danmaku)
		if op.SwitchLike == model.SwitchFeedIndexLike {
			c.CoverLeftText2 = model.LikeString(a.Arc.Stat.Like)
			c.CoverLeftText3 = model.ArchiveViewString(a.Arc.Stat.View)
		}
		switch op.CardGoto {
		case model.CardGotoAv, model.CardGotoUpRcmdAv, model.CardGotoPlayer, model.CardGotoAvConverge:
			authorface := a.Arc.Author.Face
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			if authorname != "" {
				if op.Switch != model.SwitchCooperationHide {
					authorname = unionAuthorGRPC(a, authorname)
				}
			}
			avatar = &AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil && ((c.Rcmd.RcmdReason.Style == 3 && c.IsAttenm[a.Arc.Author.Mid] == 1) ||
				(c.Rcmd.RcmdReason.Style == 6)) {
				c.Desc = authorname
			} else {
				c.Desc = authorname + " · " + model.PubDataByRequestAt(a.Arc.PubDate.Time(), c.Rcmd.RequestAt())
			}
			if op.CardGoto == model.CardGotoAv {
				c.Bvid, _ = GetBvIDStr(c.Param)
			}
			if op.CardGoto == model.CardGotoUpRcmdAv {
				button = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
			} else {
				if op.Channel != nil && op.Channel.ChannelID != 0 && op.Channel.ChannelName != "" {
					op.Channel.ChannelName = a.Arc.TypeName + " · " + op.Channel.ChannelName
					button = op.Channel
				} else if t, ok := c.Tagm[op.Tid]; ok {
					button = t
				} else {
					button = &ButtonStatus{Text: a.Arc.TypeName}
				}
				c.Base.MaskFromGRPC(a.Arc)
			}
			if !isBangumi {
				c.Base.PlayerArgs = playerArgsFrom(a)
			}
			if op.CardGoto == model.CardGotoPlayer && c.Base.PlayerArgs == nil {
				log.Warn("player card aid(%d) can't auto player", a.Arc.Aid)
				return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "PlayerArgs")
			}
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			upID = a.Arc.Author.Mid
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoAvConverge:
				var (
					urlf     func(uri string) string
					jumpGoto = op.Goto
				)
				switch op.Goto {
				case model.GotoAv:
					caid, _ := strconv.ParseInt(op.Param, 10, 64)
					ac, ok := am[caid]
					if !ok {
						return newResourceNotExistErr(ResourceArchive, caid)
					}
					if !model.AvIsNormalGRPC(ac) {
						return newInvalidResourceErr(ResourceArchive, caid, "AvIsNormalGRPC")
					}
					urlf = model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(ac, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
					c.Base.PlayerArgs = playerArgsFrom(ac)
					c.Bvid, _ = GetBvIDStr(c.Param)
				case model.GotoAvConverge:
					c.Base.PlayerArgs = nil
					c.Param = strconv.FormatInt(c.Rcmd.ID, 10)
					urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build)
				default:
					urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build)
				}
				c.Goto = jumpGoto
				c.Base.URI = model.FillURI(jumpGoto, op.Plat, op.Build, op.Param, urlf)
				c.Args.fromAvConverge(c.Base.Rcmd)
			}
			if c.Rcmd != nil {
				//nolint:exhaustive
				switch model.Gt(c.Rcmd.JumpGoto) {
				case model.GotoVerticalAv:
					c.FfCover = c.Rcmd.FfCover
					c.Goto = model.GotoVerticalAv
					c.URI = model.FillURI(model.GotoVerticalAv, op.Plat, op.Build, op.Param, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, c.Rcmd, op.Build, op.MobiApp, true))
					c.GotoIcon = model.FillGotoIcon(c.Rcmd.IconType, op.GotoIcon)
				}
			}
			c.CoverGif = op.GifCover
		case model.CardGotoChannelRcmd:
			t, ok := c.Tagm[op.Tid]
			if !ok {
				return newResourceNotExistErr(ResourceTag, op.Tid)
			}
			avatar = &AvatarStatus{Cover: t.Cover, Goto: model.GotoTag, Param: strconv.FormatInt(t.Id, 10), Type: model.AvatarSquare}
			c.Desc = model.SubscribeString(int32(t.Sub))
			button = &ButtonStatus{Goto: model.GotoTag, Param: strconv.FormatInt(t.Id, 10), IsAtten: int8(t.Attention)}
			c.Base.PlayerArgs = playerArgsFrom(a)
			c.Base.MaskFromGRPC(a.Arc)
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		case model.CardGotoAdAv:
			c.AdInfo = c.Rcmd.Ad
			avatar = &AvatarStatus{Cover: a.Arc.Author.Face, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
			c.Desc = a.Arc.Author.Name + " · " + model.PubDataByRequestAt(a.Arc.PubDate.Time(), c.Rcmd.RequestAt())
			button = c.Tagm[op.Tid]
			if (op.MobiApp == "iphone" && op.Build > 8430) || (op.MobiApp == "android" && op.Build > 5395000) {
				c.Base.PlayerArgs = playerArgsFrom(a)
				if c.Base.PlayerArgs == nil {
					log.Warn("player card ad aid(%d) can't auto player", a.Arc.Aid)
					return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "PlayerArgs")
				}
			}
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			upID = a.Arc.Author.Mid
		default:
			log.Warn("LargeCoverV1 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
		if op.CardGoto != model.CardGotoAvConverge || op.Goto != model.GotoConverge {
			c.CanPlay = a.Arc.Rights.Autoplay
		}
		if a.Arc.Rights.UGCPay == 1 && op.ShowUGCPay {
			c.CoverBadge2 = "付费"
			c.CoverBadgeStyle2 = reasonStyleFrom(model.BgColorYellow, "付费")
		}
	case map[int64]*bangumi.Season:
		sm := main.(map[int64]*bangumi.Season)
		s, ok := sm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceSeason, op.ID)
		}
		c.Base.from(op.Plat, op.Build, s.EpisodeID, s.Cover, s.Title, model.GotoBangumi, s.EpisodeID, nil)
		c.CoverLeftText2 = model.ArchiveViewString(s.PlayCount)
		c.CoverLeftText3 = model.BangumiFavString(s.Favorites, int32(s.SeasonType))
		avatar = &AvatarStatus{Cover: s.SeasonCover, Type: model.AvatarSquare}
		c.CoverBadge = s.TypeBadge
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, s.TypeBadge)
		c.Desc = s.UpdateDesc
		if t, ok := c.Tagm[op.Tid]; ok {
			button = t
		}
	case map[int32]*season.CardInfoProto:
		sm := main.(map[int32]*season.CardInfoProto)
		s, ok := sm[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceSeason, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, s.Cover, s.Title, model.GotoPGC, op.URI, nil)
		if s.Stat != nil {
			c.CoverLeftText2 = model.ArchiveViewString(int32(s.Stat.View))
			c.CoverLeftText3 = model.BangumiFavString(int32(s.Stat.Follow), s.SeasonType)
		}
		avatar = &AvatarStatus{Cover: s.Cover, Type: model.AvatarSquare}
		c.CoverBadge = s.SeasonTypeName
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, s.SeasonTypeName)
		if s.NewEp != nil {
			c.Desc = s.NewEp.IndexShow
		}
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main.(map[int32]*episodegrpc.EpisodeCardsProto)
		s, ok := sm[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		if s.Season == nil {
			return newInvalidResourceErr(ResourceEpisode, int64(s.EpisodeId), "empty `Season`")
		}
		title := s.Season.Title
		if s.ShowTitle != "" {
			title = title + "：" + s.ShowTitle
		}
		switch op.Goto {
		case model.GotoBangumi:
			cover := s.Cover
			if model.IsValidCover(c.Rcmd.CustomizedCover) {
				cover = c.Rcmd.CustomizedCover
			}
			if c.Rcmd.CustomizedTitle != "" {
				title = c.Rcmd.CustomizedTitle
			}
			c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.EpisodeId)), cover, title, model.GotoBangumi, strconv.Itoa(int(s.EpisodeId)), nil)
			if t, ok := c.Tagm[op.Tid]; ok {
				button = t
			}
		default:
			c.Base.from(op.Plat, op.Build, op.Param, s.Cover, title, model.GotoBangumi, op.URI, nil)
			c.Goto = model.GotoPGC
		}
		if s.Season.Stat != nil {
			c.CoverLeftText2 = model.ArchiveViewString(int32(s.Season.Stat.View))
			c.CoverLeftText3 = model.BangumiFavString(int32(s.Season.Stat.Follow), s.Season.SeasonType)
		}
		avatar = &AvatarStatus{Cover: s.Season.Cover, Type: model.AvatarSquare}
		c.CoverBadge = s.Season.SeasonTypeName
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, s.Season.SeasonTypeName)
		if s.Season != nil {
			c.Desc = s.Season.NewEpShow
		}
		if s.Url != "" {
			c.URI = s.Url // 直接用pgc接口返回的URL
		}
		c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		if c.Rcmd != nil {
			c.OgvCreativeId = c.Rcmd.CreativeId
		}
	case map[int32]*pgcAppGrpc.SeasonCardInfoProto:
		sm := main.(map[int32]*pgcAppGrpc.SeasonCardInfoProto)
		s, ok := sm[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, s.Cover, op.Title, op.Goto, op.URI, nil)
		if s.Stat != nil {
			c.CoverLeftText2 = model.ArchiveViewString(int32(s.Stat.View))
			c.CoverLeftText3 = model.BangumiFavString(int32(s.Stat.Follow), s.SeasonType)
		}
		avatar = &AvatarStatus{Cover: s.Cover, Type: model.AvatarSquare}
		c.CoverBadge = s.SeasonTypeName
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, s.SeasonTypeName)
		c.Desc = s.NewEp.IndexShow
		if s.Url != "" {
			c.URI = s.Url // 直接用pgc接口返回的URL
		}
		c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if r.LiveStatus != 1 {
			return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), r.Cover, r.Uname, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.CoverLeftText2 = model.LiveOnlineString(r.Online)
		avatar = &AvatarStatus{Cover: r.Cover, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound}
		c.Desc = r.Title
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		upID = r.UID
		button = r
		c.CanPlay = 1
		// new style
		if _, ok := op.SwitchStyle[model.SwitchFeedNewLive]; ok {
			c.TitleSingleLine = 1
			c.LeftCoverBadgeStyle = iconBadgeStyleFrom(r, 0)
			c.RightCoverBadgeStyle = iconBadgeStyleFrom(&ReasonStyle{
				Text: "直播中",
			}, model.BgColorRed)
			c.Avatar = &Avatar{
				Cover:   r.Face,
				URI:     c.URI,
				Event:   model.EventMainCard,
				EventV2: model.EventMainCard,
			}
		} else {
			c.CoverBadge = "直播"
			c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, "直播")
		}
		// new style
		// SmallCoverV1
	case map[int64]*show.Shopping:
		const _buttonText = "进入"
		sm := main.(map[int64]*show.Shopping)
		s, ok := sm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceShopping, op.ID)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), model.ShoppingCover(s.PerformanceImageP), s.Name, model.GotoWeb, s.URL, nil)
		if s.Type == 1 {
			c.CoverLeftText2 = s.Want
			c.CoverLeftText3 = s.CityName
			c.Desc = s.STime + " - " + s.ETime
		} else if s.Type == 2 {
			c.CoverLeftText2 = s.Want
			c.CoverLeftText3 = s.Subname
			c.Desc = s.Pricelt
		}
		button = &ButtonStatus{Text: _buttonText, Goto: model.GotoWeb, Param: s.URL, Type: model.ButtonTheme, Event: model.EventButtonClick, EventV2: model.EventV2ButtonClick}
		c.Args.fromShopping(s)
		c.CoverBadgeColor = model.PurpleCoverBadge
	case map[int64]*bangumi.EpPlayer:
		eps := main.(map[int64]*bangumi.EpPlayer)
		ep, ok := eps[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		if ep.Season == nil {
			return newInvalidResourceErr(ResourceEpisode, ep.EpID, "empty `Season`")
		}
		c.Base.from(op.Plat, op.Build, op.Param, ep.Cover, ep.Season.Title, model.GotoBangumi, "", nil)
		c.URI = model.FillURI("", 0, 0, ep.Uri, model.PGCTrackIDHandler(c.Rcmd))
		if ep.PlayerInfo != nil && ((op.MobiApp == "iphone" && op.Build > 8470) || (op.MobiApp == "android" && op.Build > 5405000)) {
			c.CanPlay = 1
		}
		avatar = &AvatarStatus{Cover: ep.Season.Cover, Type: model.AvatarSquare}
		c.CoverLeftText1 = model.DurationString(ep.Duration)
		if ep.Stat != nil {
			c.CoverLeftText2 = model.ArchiveViewString(int32(ep.Stat.Play))
			c.CoverLeftText3 = model.DanmakuString(int32(ep.Stat.Danmaku))
		}
		c.Desc = ep.NewDesc
		c.Base.PlayerArgs = playerArgsFrom(ep)
	case map[int32]*pgcinline.EpisodeCard:
		eps := main.(map[int32]*pgcinline.EpisodeCard)
		ep, ok := eps[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		if ep.Season == nil {
			return newInvalidResourceErr(ResourceEpisode, op.ID, "empty `Season`")
		}
		c.Base.from(op.Plat, op.Build, op.Param, ep.Cover, ep.Season.Title, model.GotoBangumi, "", nil)
		c.URI = model.FillURI("", 0, 0, ep.Url, model.PGCTrackIDHandler(c.Rcmd))
		if ep.PlayerInfo != nil && ((op.MobiApp == "iphone" && op.Build > 8470) || (op.MobiApp == "android" && op.Build > 5405000)) {
			c.CanPlay = 1
		}
		avatar = &AvatarStatus{Cover: ep.Season.Cover, Type: model.AvatarSquare}
		c.CoverLeftText1 = model.DurationString(ep.Duration)
		if ep.Stat != nil {
			c.CoverLeftText2 = model.ArchiveViewString(int32(ep.Stat.Play))
			c.CoverLeftText3 = model.DanmakuString(int32(ep.Stat.Danmaku))
		}
		c.Desc = ep.NewDesc
		c.Base.PlayerArgs = playerArgsFrom(ep)
	case *cm.AdInfo:
		ad := main.(*cm.AdInfo)
		c.AdInfo = ad
		c.CmInfo = &cm.CmInfo{HidePlayButton: true}
	case map[model.Gt]interface{}:
		const (
			_defalutCover = 1 // 展示默认up主头像
		)
		intfcm := main.(map[model.Gt]interface{})
		intfc, ok := intfcm[op.Goto]
		if !ok {
			return newUnexpectedGotoErr(string(op.Goto), "")
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[model.ColumnSvrDouble], op.Title, op.Goto, op.URI, nil)
		c.CoverGif = op.GifCover
		c.Desc = op.Desc
		c.CoverBadge = op.Badge
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
		//nolint:gosimple
		switch intfc.(type) {
		case map[int64]*arcgrpc.ArcPlayer:
			am := intfc.(map[int64]*arcgrpc.ArcPlayer)
			a, ok := am[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArchive, op.ID)
			}
			if !model.AvIsNormalGRPC(a) {
				return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
			}
			c.CoverLeftText1 = model.DurationString(a.Arc.Duration)
			c.CoverLeftText2 = model.ArchiveViewString(a.Arc.Stat.View)
			c.CoverLeftText3 = model.DanmakuString(a.Arc.Stat.Danmaku)
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoSpecialB:
				authorface := a.Arc.Author.Face
				authorname := a.Arc.Author.Name
				if c.Cardm != nil {
					if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
						authorface = au.Face
						authorname = au.Name
					}
				}
				avatar = &AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound, DefalutCover: _defalutCover}
				c.Desc = authorname + " · " + model.PubDataString(a.Arc.PubDate.Time())
				c.Cover = op.Cover
			}
		case map[int32]*episodegrpc.EpisodeCardsProto:
			sm := intfc.(map[int32]*episodegrpc.EpisodeCardsProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Season == nil {
				return newInvalidResourceErr(ResourceEpisode, int64(s.EpisodeId), "empty `Season`")
			}
			if s.Season.Stat != nil {
				c.CoverLeftText2 = model.ArchiveViewString(int32(s.Season.Stat.View))
				c.CoverLeftText3 = model.BangumiFavString(int32(s.Season.Stat.Follow), s.Season.SeasonType)
			}
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoSpecialB:
				avatar = &AvatarStatus{Cover: s.Season.Cover, Type: model.AvatarRound, DefalutCover: _defalutCover}
				c.Desc = s.Season.NewEpShow
				c.Cover = op.Cover
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int32]*pgccard.EpisodeCard:
			sm := intfc.(map[int32]*pgccard.EpisodeCard)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceSeason, op.ID)
			}
			if s.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(s.Stat.Play), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(s.Stat.Follow), "")
				c.CoverLeftIcon2 = model.IconFavorite
				c.CoverRightText = model.OgvCoverRightText(c.Rcmd, s, enablePgcScore(op))
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int64]*live.Room:
			rm := intfc.(map[int64]*live.Room)
			r, ok := rm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceRoom, op.ID)
			}
			if r.LiveStatus != 1 {
				return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
			}
			c.CoverLeftText2 = model.LiveOnlineString(r.Online)
			switch op.CardGoto {
			case model.CardGotoSpecialB:
				avatar = &AvatarStatus{Cover: r.Cover, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound, DefalutCover: _defalutCover}
				c.Desc = r.Uname
				c.Cover = op.Cover
				c.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.URI, model.LiveRoomHandler(r, op.Network))
				// 使用直播接口返回的url
				if r.Link != "" {
					c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
				}
			default:
				c.CoverRightText = r.Uname
			}
		case map[int64]*article.Meta:
			mm := intfc.(map[int64]*article.Meta)
			m, ok := mm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArticle, op.ID)
			}
			c.CoverLeftText2 = model.ArticleViewString(m.Stats.View)
			c.CoverLeftText3 = model.ArticleReplyString(m.Stats.Reply)
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoSpecialB:
				avatar = &AvatarStatus{Cover: m.Author.Face, Goto: model.GotoMid, Param: strconv.FormatInt(m.Author.Mid, 10), Type: model.AvatarRound, DefalutCover: _defalutCover}
				c.Desc = m.Author.Name + " · " + model.PubDataString(m.PublishTime.Time())
				c.Cover = op.Cover
			}
		default:
			log.Warn("LargeCoverV1 From: unexpected type %T", intfc)
			return newUnexpectedResourceTypeErr(intfc, "")
		}
		c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	case nil:
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[model.ColumnSvrDouble], op.Title, op.Goto, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoDownload:
			const _buttonText = "进入"
			c.Desc = op.Desc
			c.CoverLeftText2 = model.DownloadString(op.Download)
			if (op.Plat == model.PlatIPhone && op.Build > 8220) || (op.Plat == model.PlatAndroid && op.Build > 5335001) {
				button = &ButtonStatus{Text: _buttonText, Goto: op.Goto, Param: op.URI, Type: model.ButtonTheme, Event: model.EventGameClick, EventV2: model.EventV2ButtonClick}
			} else {
				button = &ButtonStatus{Text: _buttonText, Goto: op.Goto, Param: op.URI, Type: model.ButtonTheme, Event: model.EventButtonClick, EventV2: model.EventV2ButtonClick}
			}
			c.CoverBadgeColor = model.PurpleCoverBadge
		case model.CardGotoSpecial:
			c.CoverGif = op.GifCover
			c.Desc = op.Desc
			c.CoverBadge = op.Badge
			c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
			// v546 del
			if _, ok := op.SwitchStyle[model.SwitchSpecialInfo]; !ok {
				c.CoverBadgeColor = model.PurpleCoverBadge
			}
			c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
		default:
			log.Warn("LargeCoverV1 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("LargeCoverV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	if (op.CardGoto != model.CardGotoSpecialB && c.Rcmd != nil) || (op.CardGoto == model.CardGotoSpecialB && c.CoverBadge == "" && c.Rcmd != nil) {
		var isShowV2 bool
		c.TopRcmdReason, c.BottomRcmdReason = TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoAvConverge:
			if isShowV2 = IsShowRcmdReasonStyleV2(c.Rcmd); isShowV2 {
				c.Desc = ""
			}
		}
		if _, ok := op.SwitchStyle[model.SwitchFeedNewLive]; ok {
			if c.BottomRcmdReason != "" {
				c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.BottomRcmdReason, c.Base.Goto, op)
			} else {
				c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op)
			}
			c.TopRcmdReason = ""
			c.BottomRcmdReason = ""
		} else {
			if isShowV2 {
				if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV4(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV3(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else {
					c.RcmdReasonStyleV2 = reasonStyleFromV2(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op.Plat, op.Build)
				}
				c.TopRcmdReason = ""
				c.BottomRcmdReason = ""
			} else {
				c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op)
				c.BottomRcmdReasonStyle = bottomReasonStyleFrom(c.Rcmd, c.BottomRcmdReason, c.Base.Goto, op)
			}
		}
	}
	c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	if _, ok := op.SwitchStyle[model.SwitchFeedNewLive]; !ok {
		c.Avatar = avatarFrom(avatar)
	}
	if c.Rcmd == nil || !c.Rcmd.HideButton {
		c.DescButton = buttonFrom(button, op.Plat)
	}
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	if op.Cover != "" {
		c.Cover = op.Cover
	}
	if op.GifCover != "" {
		c.CoverGif = op.GifCover
	}
	if op.Desc != "" && op.CardGoto != model.CardGotoBangumi {
		c.Desc = op.Desc
	}
	c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	c.Right = true
	return nil
}

func (c *LargeCoverV1) Get() *Base {
	return c.Base
}

type SmallCoverV1 struct {
	*Base
	CoverBadge             string       `json:"cover_badge,omitempty"`
	CoverBadgeStyle        *ReasonStyle `json:"cover_badge_style,omitempty"`
	CoverLeftText          string       `json:"cover_left_text,omitempty"`
	Desc1                  string       `json:"desc_1,omitempty"`
	Desc2                  string       `json:"desc_2,omitempty"`
	Desc3                  string       `json:"desc_3,omitempty"`
	TitleRightText         string       `json:"title_right_text,omitempty"`
	TitleRightPic          model.Icon   `json:"title_right_pic,omitempty"`
	TopRcmdReasonStyle     *ReasonStyle `json:"top_rcmd_reason_style,omitempty"`
	DestroyCard            int8         `json:"destroy_card,omitempty"`
	RcmdReasonStyle        *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	LeftCoverBadgeNewStyle *ReasonStyle `json:"left_cover_badge_new_style,omitempty"`
	SeasonId               []int64      `json:"season_id,omitempty"`
	Epid                   []int64      `json:"epid,omitempty"`
}

//nolint:gocognit
func (c *SmallCoverV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var button interface{}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*audio.Audio:
		var firstSong string
		am := main.(map[int64]*audio.Audio)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceAudio, op.ID)
		}
		if len(a.Songs) != 0 {
			firstSong = a.Songs[0].Title
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.CoverURL, a.Title, model.GotoAudio, op.URI, nil)
		c.Desc1, c.Desc2 = model.AudioDescString(firstSong, a.RecordNum)
		c.Desc3 = model.AudioPlayString(a.PlayNum) + "  " + model.AudioFavString(a.FavoriteNum)
		c.Args.fromAudio(a)
		button = a.Ctgs
	case *bangumi.Update:
		const (
			_title   = "你的追番更新啦"
			_updates = 99
		)
		u := main.(*bangumi.Update)
		if u == nil {
			return newResourceNotExistErr(ResourceBangumiUpdate, 0)
		}
		if u.Updates == 0 {
			return newInvalidResourceErr(ResourceBangumiUpdate, 0, "empty `Updates`")
		}
		c.Base.from(op.Plat, op.Build, "", u.SquareCover, _title, "", "", nil)
		updates := u.Updates
		if updates > _updates {
			updates = _updates
			c.TitleRightPic = model.IconBomb
		} else {
			c.TitleRightPic = model.IconTV
		}
		c.Desc1 = u.Title
		c.TitleRightText = strconv.Itoa(updates)
		fixtureForIOS617(c, op)
	case *bangumi.Remind:
		const _updates = 99
		u := main.(*bangumi.Remind)
		if u == nil {
			return newResourceNotExistErr(ResourceBangumiRemind, 0)
		}
		if u.Updates == 0 {
			return newInvalidResourceErr(ResourceBangumiRemind, 0, "Updates: %d", u.Updates)
		}
		if len(u.List) == 0 {
			return newInvalidResourceErr(ResourceBangumiRemind, 0, "empty `List`")
		}
		cover := u.List[0].SquareCover
		if cover == "" {
			cover = u.List[0].Cover
		}
		uriStr := u.List[0].Uri
		uri := strings.Split(uriStr, "?")
		if len(uri) == 1 {
			uriStr = uriStr + "?from=21"
		} else if len(uri) > 1 {
			uriStr = uriStr + "&from=21"
		}
		c.Base.from(op.Plat, op.Build, "", cover, u.List[0].UpdateTitle, "", uriStr, nil)
		updates := u.Updates
		if updates > _updates {
			updates = _updates
			c.TitleRightPic = model.IconBomb
		} else {
			c.TitleRightPic = model.IconTV
		}
		c.Desc1 = u.List[0].UpdateDesc
		if len(u.List) > 1 {
			c.Desc2 = u.List[1].UpdateDesc
		}
		for _, v := range u.List {
			c.SeasonId = append(c.SeasonId, v.SeasonId)
			c.Epid = append(c.Epid, v.Epid)
		}
		c.TitleRightText = strconv.Itoa(updates)
		fixtureForIOS617(c, op)
	case map[int64]*show.Shopping:
		const _buttonText = "进入"
		sm := main.(map[int64]*show.Shopping)
		s, ok := sm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceShopping, op.ID)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), model.ShoppingCover(s.PerformanceImageP), s.Name, model.GotoWeb, s.URL, nil)
		if s.Type == 1 {
			c.Desc1 = s.STime + " - " + s.ETime
			c.Desc2 = s.CityName
			c.Desc3 = "￥" + s.Pricelt
		} else if s.Type == 2 {
			c.Desc1 = s.Subname
			c.Desc2 = s.Want
			c.Desc3 = s.Pricelt
		}
		button = &ButtonStatus{Text: _buttonText, Goto: model.GotoWeb, Param: s.URL, Type: model.ButtonTheme, Event: model.EventButtonClick, EventV2: model.EventV2ButtonClick}
		c.Args.fromShopping(s)
	case *cm.AdInfo:
		ad := main.(*cm.AdInfo)
		c.AdInfo = ad
	case *bangumi.Moe:
		m := main.(*bangumi.Moe)
		if m == nil {
			return newResourceNotExistErr(ResourceMoe, 0)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(m.ID, 10), m.Square, m.Title, model.GotoWeb, m.Link, nil)
		c.Desc1 = m.Desc
		c.CoverBadge = m.Badge
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, m.Badge)
	case *tunnelgrpc.FeedCard:
		const (
			_destroyCard = 1 // 客户端点击完切回列表页销毁卡片
		)
		tunnelCard := main.(*tunnelgrpc.FeedCard)
		if tunnelCard == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		c.Base.from(op.Plat, op.Build, op.Param, tunnelCard.Cover, tunnelCard.Title, model.GotoWeb, tunnelCard.Link, nil)
		c.Goto = model.GotoGame
		c.Desc1 = tunnelCard.Intro
		c.DestroyCard = _destroyCard
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		intfc, ok := intfcm[op.Goto]
		if !ok {
			return newUnexpectedGotoErr(string(op.Goto), "")
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		c.Desc1 = op.Desc
		c.CoverBadge = op.Badge
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
		//nolint:gosimple
		switch intfc.(type) {
		case map[int64]*arcgrpc.ArcPlayer:
			am := intfc.(map[int64]*arcgrpc.ArcPlayer)
			a, ok := am[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArchive, op.ID)
			}
			if !model.AvIsNormalGRPC(a) {
				return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
			}
			c.CoverLeftText = model.DurationString(a.Arc.Duration)
			c.Desc2 = model.ArchiveViewString(a.Arc.Stat.View) + "  " + model.DanmakuString(a.Arc.Stat.Danmaku)
		case map[int32]*episodegrpc.EpisodeCardsProto:
			sm := intfc.(map[int32]*episodegrpc.EpisodeCardsProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Season == nil {
				return newInvalidResourceErr(ResourceEpisode, int64(s.EpisodeId), "empty `Season`")
			}
			if s.Season.Stat != nil {
				c.Desc2 = model.ArchiveViewString(int32(s.Season.Stat.View)) + "  " + model.BangumiFavString(int32(s.Season.Stat.Follow), s.Season.SeasonType)
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int32]*pgcAppGrpc.SeasonCardInfoProto:
			sm := intfc.(map[int32]*pgcAppGrpc.SeasonCardInfoProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Stat != nil {
				c.Desc2 = model.ArchiveViewString(int32(s.Stat.View)) + "  " + model.BangumiFavString(int32(s.Stat.Follow), s.SeasonType)
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int64]*live.Room:
			rm := intfc.(map[int64]*live.Room)
			r, ok := rm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceRoom, op.ID)
			}
			if r.LiveStatus != 1 {
				return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
			}
			c.Desc2 = model.LiveOnlineString(r.Online) + "  " + "UP主：" + r.Uname
			c.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.URI, model.LiveRoomHandler(r, op.Network))
			// 使用直播接口返回的url
			if r.Link != "" {
				c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
			}
		case map[int64]*article.Meta:
			mm := intfc.(map[int64]*article.Meta)
			m, ok := mm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArticle, op.ID)
			}
			c.Desc2 = model.ArticleViewString(m.Stats.View) + "  " + model.ArticleReplyString(m.Stats.Reply)
		default:
			log.Warn("SmallCoverV1 From: unexpected type %T", intfc)
			return newUnexpectedResourceTypeErr(intfc, "")
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		}
		c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	case nil:
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoDownload:
			const _buttonText = "进入"
			c.Desc1 = op.Desc
			c.Desc2 = model.DownloadString(op.Download)
			if (op.Plat == model.PlatIPhone && op.Build > 8220) || (op.Plat == model.PlatAndroid && op.Build > 5335001) {
				button = &ButtonStatus{Text: _buttonText, Goto: op.Goto, Param: op.URI, Type: model.ButtonTheme, Event: model.EventGameClick, EventV2: model.EventV2GameClick}
			} else {
				button = &ButtonStatus{Text: _buttonText, Goto: op.Goto, Param: op.URI, Type: model.ButtonTheme, Event: model.EventButtonClick, EventV2: model.EventV2GameClick}
			}
		case model.CardGotoSpecial:
			c.Desc1 = op.Desc
			c.CoverBadge = op.Badge
			c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			}
			c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
		default:
			log.Warn("SmallCoverV1 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("SmallCoverV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.DescButton = buttonFrom(button, op.Plat)
	c.Right = true
	return nil
}

// 仅针对 ios 617 的 bugfix
func fixtureForIOS617(in *SmallCoverV1, op *operate.Card) {
	if !(op.MobiApp == "iphone" && op.Build == 61700200) {
		return
	}
	if in.CoverBadgeStyle == nil {
		in.CoverBadgeStyle = &ReasonStyle{
			Text:    " ",
			BgStyle: 1,
		}
	}
	if in.LeftCoverBadgeNewStyle == nil {
		in.LeftCoverBadgeNewStyle = &ReasonStyle{
			IconURL:      "https://i0.hdslb.com/bfs/feed-admin/084f3275802a6bb2797a0a1ba106e676c04ce2e1.png",
			IconURLNight: "https://i0.hdslb.com/bfs/feed-admin/8d345f61fa8081ef2097d71740531857b59ed06e.png",
		}
	}
}

func (c *SmallCoverV1) Get() *Base {
	return c.Base
}

type MiddleCover struct {
	*Base
	Ratio                  int          `json:"ratio,omitempty"`
	Badge                  string       `json:"badge,omitempty"`
	BadgeStyle             *ReasonStyle `json:"badge_style,omitempty"`
	Desc                   string       `json:"desc,omitempty"`
	CoverLeftText1         string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1         model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2         string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2         model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3         string       `json:"cover_left_text_3,omitempty"`
	CoverRightText         string       `json:"cover_right_text,omitempty"`
	TopRcmdReasonStyle     *ReasonStyle `json:"top_rcmd_reason_style,omitempty"` // 客户端未使用
	RcmdReasonStyle        *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	LeftCoverBadgeNewStyle *ReasonStyle `json:"left_cover_badge_new_style,omitempty"`
}

func (c *MiddleCover) From(main interface{}, op *operate.Card) error {
	//nolint:gosimple
	switch main.(type) {
	case *cm.AdInfo:
		ad := main.(*cm.AdInfo)
		c.AdInfo = ad
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		intfc, ok := intfcm[op.Goto]
		if !ok {
			return newUnexpectedGotoErr(string(op.Goto), "")
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		c.Desc = op.Desc
		c.Badge = op.Badge
		c.Ratio = op.Ratio
		switch model.Columnm[c.Columnm] {
		case model.ColumnSvrSingle:
			c.BadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
		default:
			c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
		}
		//nolint:gosimple
		switch intfc.(type) {
		case map[int64]*arcgrpc.ArcPlayer:
			am := intfc.(map[int64]*arcgrpc.ArcPlayer)
			a, ok := am[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArchive, op.ID)
			}
			if !model.AvIsNormalGRPC(a) {
				return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
			}
			switch model.Columnm[c.Columnm] {
			case model.ColumnSvrSingle:
				c.CoverLeftText1 = model.DurationString(a.Arc.Duration)
				c.CoverLeftText2 = model.ArchiveViewString(a.Arc.Stat.View)
				c.CoverLeftText3 = model.DanmakuString(a.Arc.Stat.Danmaku)
			default:
				c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
				c.CoverLeftIcon2 = model.IconDanmaku
				c.CoverRightText = model.DurationString(a.Arc.Duration)
			}
		case map[int32]*episodegrpc.EpisodeCardsProto:
			sm := intfc.(map[int32]*episodegrpc.EpisodeCardsProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Season != nil && s.Season.Stat != nil {
				switch model.Columnm[c.Columnm] {
				case model.ColumnSvrSingle:
					c.CoverLeftText2 = model.ArchiveViewString(int32(s.Season.Stat.View))
					c.CoverLeftText3 = model.BangumiFavString(int32(s.Season.Stat.Follow), s.Season.SeasonType)
				default:
					c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
					c.CoverLeftIcon1 = model.IconPlay
					c.CoverLeftText2 = model.StatString(int32(s.Season.Stat.Follow), "")
					c.CoverLeftIcon2 = model.IconFavorite
				}
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int64]*live.Room:
			rm := intfc.(map[int64]*live.Room)
			r, ok := rm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceRoom, op.ID)
			}
			if r.LiveStatus != 1 {
				return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
			}
			switch model.Columnm[c.Columnm] {
			case model.ColumnSvrSingle:
				c.CoverLeftText2 = model.LiveOnlineString(r.Online)
			default:
				c.CoverLeftText1 = model.StatString(r.Online, "")
				c.CoverLeftIcon1 = model.IconOnline
			}
			c.CoverRightText = r.Uname
			c.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.URI, model.LiveRoomHandler(r, op.Network))
			// 使用直播接口返回的url
			if r.Link != "" {
				c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
			}
		case map[int64]*article.Meta:
			mm := intfc.(map[int64]*article.Meta)
			m, ok := mm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArticle, op.ID)
			}
			if m.Stats != nil {
				switch model.Columnm[c.Columnm] {
				case model.ColumnSvrSingle:
					c.CoverLeftText2 = model.ArticleViewString(m.Stats.View)
					c.CoverLeftText3 = model.ArticleReplyString(m.Stats.Reply)
				default:
					c.CoverLeftText1 = model.StatString(int32(m.Stats.View), "")
					c.CoverLeftIcon1 = model.IconRead
					c.CoverLeftText2 = model.StatString(int32(m.Stats.Reply), "")
					c.CoverLeftIcon2 = model.IconComment
				}
			}
		default:
			log.Warn("MiddleCover From: unexpected type %T", intfc)
			return newUnexpectedResourceTypeErr(intfc, "")
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		}
		c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
		//nolint:exhaustive
		switch model.Columnm[c.Columnm] {
		case model.ColumnSvrDouble:
			if c.RcmdReasonStyle != nil {
				c.Desc = ""
			}
		}
	case nil:
		if op == nil {
			return newEmptyOPErr()
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoSpecial:
			c.Desc = op.Desc
			c.Badge = op.Badge
			c.Ratio = op.Ratio
			if _, ok := op.SwitchStyle[model.SwitchSpecialInfo]; ok && c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			}
			c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
			switch model.Columnm[c.Columnm] {
			case model.ColumnSvrDouble:
				if c.RcmdReasonStyle != nil {
					c.Desc = ""
				}
				c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
			default:
				c.BadgeStyle = reasonStyleFrom(model.BgColorRed, op.Badge)
			}

		default:
			log.Warn("MiddleCover From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("MiddleCover From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *MiddleCover) Get() *Base {
	return c.Base
}

type Topstick struct {
	*Base
	Desc string `json:"desc,omitempty"`
}

func (c *Topstick) From(main interface{}, op *operate.Card) error {
	switch main.(type) {
	case nil:
		if op == nil {
			return newEmptyOPErr()
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoTopstick:
			c.Desc = op.Desc
		default:
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	}
	c.Right = true
	return nil
}

func (c *Topstick) Get() *Base {
	return c.Base
}

type ThreeItemV1 struct {
	*Base
	TitleIcon   model.Icon         `json:"title_icon,omitempty"`
	BannerCover string             `json:"banner_cover,omitempty"`
	BannerURI   string             `json:"banner_uri,omitempty"`
	MoreURI     string             `json:"more_uri,omitempty"`
	MoreText    string             `json:"more_text,omitempty"`
	Items       []*ThreeItemV1Item `json:"items,omitempty"`
}

type ThreeItemV1Item struct {
	Base
	CoverLeftText string       `json:"cover_left_text,omitempty"`
	CoverLeftIcon model.Icon   `json:"cover_left_icon,omitempty"`
	Desc1         string       `json:"desc_1,omitempty"`
	Desc2         string       `json:"desc_2,omitempty"`
	Badge         string       `json:"badge,omitempty"`
	BadgeStyle    *ReasonStyle `json:"badge_style,omitempty"`
}

//nolint:gocognit
func (c *ThreeItemV1) From(main interface{}, op *operate.Card) error {
	switch main.(type) {
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		if op == nil {
			return newEmptyOPErr()
		}
		switch op.CardGoto {
		case model.CardGotoRank:
			const (
				_title = "全站排行榜"
				_limit = 3
			)
			c.Base.from(op.Plat, op.Build, "0", "", _title, "", "", nil)
			// c.TitleIcon = model.IconRank
			c.MoreURI = model.FillURI(op.Goto, 0, 0, op.URI, nil)
			c.MoreText = "查看更多"
			c.Items = make([]*ThreeItemV1Item, 0, _limit)
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item := &ThreeItemV1Item{
						CoverLeftText: model.DurationString(a.Arc.Duration),
						Desc1:         model.ScoreString(v.Score),
					}
					item.Base.from(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
					item.Args.fromArchiveGRPC(a.Arc, nil)
					c.Items = append(c.Items, item)
					if len(c.Items) == _limit {
						break
					}
				}
			}
			if len(c.Items) < _limit {
				return newInvalidResourceErr(ResourceItems, 0, "lack of items")
			}
			c.Items[0].CoverLeftIcon = model.IconGoldMedal
			c.Items[1].CoverLeftIcon = model.IconSilverMedal
			c.Items[2].CoverLeftIcon = model.IconBronzeMedal
		case model.CardGotoConverge, model.CardGotoConvergeAi:
			limit := 3
			if op.Coverm[c.Columnm] != "" {
				limit = 2
			}
			c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
			c.MoreURI = model.FillURI(model.GotoConverge, 0, 0, op.Param, nil)
			c.MoreText = "查看更多"
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoConvergeAi:
				limit = 2
				c.Base.CardGoto = model.CardGotoConverge
				if len(op.Items) <= limit {
					c.MoreText = ""
					c.MoreURI = ""
				}
			}
			c.Items = make([]*ThreeItemV1Item, 0, len(op.Items))
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				var item *ThreeItemV1Item
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item = &ThreeItemV1Item{
						CoverLeftText: model.DurationString(a.Arc.Duration),
						Desc1:         model.ArchiveViewString(a.Arc.Stat.View),
						Desc2:         model.DanmakuString(a.Arc.Stat.Danmaku),
					}
					if op.SwitchLike == model.SwitchFeedIndexLike {
						item.Desc1 = model.LikeString(a.Arc.Stat.Like)
						item.Desc2 = model.ArchiveViewString(a.Arc.Stat.View)
					}
					item.Base.from(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
					item.Args.fromArchiveGRPC(a.Arc, nil)
				case map[int64]*live.Room:
					rm := intfc.(map[int64]*live.Room)
					r, ok := rm[v.ID]
					if !ok || r.LiveStatus != 1 {
						continue
					}
					item = &ThreeItemV1Item{
						Desc1:      model.LiveOnlineString(r.Online),
						Badge:      "直播",
						BadgeStyle: reasonStyleFrom(model.BgColorTransparentRed, "直播"),
					}
					item.Base.from(op.Plat, op.Build, v.Param, r.Cover, r.Title, model.GotoLive, v.URI, model.LiveRoomHandler(r, op.Network))
					// 使用直播接口返回的url
					if r.Link != "" {
						item.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
					}
					item.Args.fromLiveRoom(r)
				case map[int64]*article.Meta:
					mm := intfc.(map[int64]*article.Meta)
					m, ok := mm[v.ID]
					if !ok {
						continue
					}
					if len(m.ImageURLs) == 0 {
						continue
					}
					item = &ThreeItemV1Item{
						Badge:      "文章",
						BadgeStyle: reasonStyleFrom(model.BgColorTransparentRed, "文章"),
					}
					item.Base.from(op.Plat, op.Build, v.Param, m.ImageURLs[0], m.Title, model.GotoArticle, v.URI, nil)
					if m.Stats != nil {
						item.Desc1 = model.ArticleViewString(m.Stats.View)
						item.Desc2 = model.ArticleReplyString(m.Stats.Reply)
					}
					item.Args.fromArticle(m)
				default:
					log.Warn("ThreeItemV1 From: unexpected type %T", intfc)
					continue
				}
				c.Items = append(c.Items, item)
				if len(c.Items) == limit {
					break
				}
			}
			if len(c.Items) < limit {
				return newInvalidResourceErr(ResourceItems, 0, "lack of items")
			}
		default:
			log.Warn("ThreeItemV1 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("ThreeItemV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *ThreeItemV1) Get() *Base {
	return c.Base
}

type ThreeItemH struct {
	*Base
	Items []*ThreeItemHItem `json:"items,omitempty"`
}

type ThreeItemHItem struct {
	Base
	CoverType    model.Type `json:"cover_type,omitempty"`
	Desc         string     `json:"desc,omitempty"`
	OfficialIcon model.Icon `json:"official_icon,omitempty"`
}

func (c *ThreeItemH) From(main interface{}, op *operate.Card) error {
	switch main.(type) {
	case nil:
		if op == nil {
			return newEmptyOPErr()
		}
		switch op.CardGoto {
		case model.CardGotoSubscribe, model.CardGotoSearchSubscribe:
			const _limit = 3
			c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, "", "", nil)
			c.Items = make([]*ThreeItemHItem, 0, _limit)
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				var (
					item   *ThreeItemHItem
					button interface{}
				)
				switch v.Goto {
				case model.GotoTag:
					t, ok := c.Tagm[v.ID]
					if !ok || t.Attention == 1 {
						continue
					}
					item = &ThreeItemHItem{
						CoverType: model.AvatarSquare,
						Desc:      model.SubscribeString(int32(t.Sub)),
					}
					item.Base.from(op.Plat, op.Build, v.Param, t.Cover, t.Name, v.Goto, v.URI, nil)
					button = &ButtonStatus{Goto: model.GotoTag, Param: strconv.FormatInt(t.Id, 10)}
				case model.GotoMid:
					cd, ok := c.Cardm[v.ID]
					if !ok || c.IsAttenm[v.ID] == 1 {
						continue
					}
					item = &ThreeItemHItem{
						CoverType: model.AvatarRound,
					}
					item.Base.from(op.Plat, op.Build, v.Param, cd.Face, cd.Name, v.Goto, v.URI, nil)
					button = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(cd.Mid, 10)}
					if v.Desc != "" {
						item.Desc = v.Desc
					} else if stat, ok := c.Statm[cd.Mid]; ok {
						item.Desc = model.FanString(int32(stat.Follower))
					}
					item.OfficialIcon = model.OfficialIcon(cd)
				default:
					log.Warn("ThreeItemH From: unexpected type %T", v.Goto)
					continue
				}
				item.DescButton = buttonFrom(button, op.Plat)
				c.Items = append(c.Items, item)
				if len(c.Items) == _limit {
					break
				}
			}
			if len(c.Items) < _limit {
				return newInvalidResourceErr(ResourceItems, 0, "lack of items")
			}
		default:
			log.Warn("ThreeItemH From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("ThreeItemH From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *ThreeItemH) Get() *Base {
	return c.Base
}

type ThreeItemHV3 struct {
	*Base
	Covers          []string     `json:"covers,omitempty"`
	CoverTopText1   string       `json:"cover_top_text_1,omitempty"`
	CoverTopText2   string       `json:"cover_top_text_2,omitempty"`
	Desc            string       `json:"desc,omitempty"`
	Avatar          *Avatar      `json:"avatar,omitempty"`
	OfficialIcon    model.Icon   `json:"official_icon,omitempty"`
	RcmdReasonStyle *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *ThreeItemHV3) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		upID int64
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*article.Meta:
		mm := main.(map[int64]*article.Meta)
		m, ok := mm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArticle, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, "", m.Title, model.GotoArticle, op.URI, nil)
		c.Covers = m.ImageURLs
		c.CoverTopText1 = model.ArticleViewString(m.Stats.View)
		c.CoverTopText2 = model.ArticleReplyString(m.Stats.Reply)
		c.Desc = m.Summary
		if m.Author != nil {
			c.Avatar = avatarFrom(&AvatarStatus{Cover: m.Author.Face, Text: m.Author.Name + "·" + model.PubDataByRequestAt(m.PublishTime.Time(), c.Rcmd.RequestAt()), Goto: model.GotoMid, Param: strconv.FormatInt(m.Author.Mid, 10), Type: model.AvatarRound})
			upID = m.Author.Mid
		}
		c.Args.fromArticle(m)
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		}
	default:
		log.Warn("ThreeItemHV3 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	c.Right = true
	return nil
}

func (c *ThreeItemHV3) Get() *Base {
	return c.Base
}

type TwoItemV1 struct {
	*Base
	Items []*TwoItemV1Item `json:"items,omitempty"`
}

type TwoItemV1Item struct {
	Base
	CoverBadge      string       `json:"cover_badge,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_badge_style,omitempty"`
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
}

func (c *TwoItemV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64][]*live.Card:
		const _limit = 2
		csm := main.(map[int64][]*live.Card)
		cs, ok := csm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoomGroup, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, "", "", "", "", nil)
		c.Items = make([]*TwoItemV1Item, 0, _limit)
		for _, card := range cs {
			if card == nil || card.LiveStatus != 1 {
				continue
			}
			item := &TwoItemV1Item{
				CoverBadge:      "直播",
				CoverBadgeStyle: reasonStyleFrom(model.BgColorRed, "直播"),
				CoverLeftText1:  model.LiveOnlineString(card.Online),
			}
			item.DescButton = buttonFrom(card, op.Plat)
			item.Base.from(op.Plat, op.Build, strconv.FormatInt(card.RoomID, 10), card.ShowCover, card.Title, model.GotoLive, strconv.FormatInt(card.RoomID, 10), model.LiveUpHandler(card))
			item.Args.fromLiveUp(card)
			c.Items = append(c.Items, item)
			if len(c.Items) == _limit {
				break
			}
		}
	}
	c.Right = true
	return nil
}

func (c *TwoItemV1) Get() *Base {
	return c.Base
}

type CoverOnly struct {
	*Base
}

func (c *CoverOnly) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case nil:
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoVip:
			c.Base.from(op.Plat, op.Build, "", op.Cover, "", "", "", nil)
		}
	default:
		log.Warn("CoverOnly From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *CoverOnly) Get() *Base {
	return c.Base
}

type Banner struct {
	*Base
	Hash       string           `json:"hash,omitempty"`
	BannerItem []*banner.Banner `json:"banner_item,omitempty"`
}

func (c *Banner) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case nil:
		switch op.CardGoto {
		case model.CardGotoBanner:
			if len(op.Banner) == 0 {
				log.Warn("Banner len is null")
				return newInvalidResourceErr(ResourceOperateCard, 0, "empty `Banner`")
			}
			c.BannerItem = op.Banner
			c.Hash = op.Hash
		default:
			log.Warn("Banner From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("Banner From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *Banner) Get() *Base {
	return c.Base
}

type Text struct {
	*Base
	Content string `json:"content,omitempty"`
}

func (c *Text) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case nil:
		switch op.CardGoto {
		case model.CardGotoNews:
			c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, model.GotoWeb, op.URI, nil)
			c.Content = op.Desc
		default:
			log.Warn("Text From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("Text From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *Text) Get() *Base {
	return c.Base
}

type ThreeItemHV4 struct {
	*Base
	MoreURI  string              `json:"more_uri,omitempty"`
	MoreText string              `json:"more_text,omitempty"`
	Items    []*ThreeItemHV4Item `json:"items,omitempty"`
}

type ThreeItemHV4Item struct {
	Cover           string           `json:"cover,omitempty"`
	Title           string           `json:"title,omitempty"`
	Desc            string           `json:"desc,omitempty"`
	Goto            model.Gt         `json:"goto,omitempty"`
	Param           string           `json:"param,omitempty"`
	URI             string           `json:"uri,omitempty"`
	CoverBadge      string           `json:"cover_badge,omitempty"`
	CoverBadgeStyle *ReasonStyle     `json:"cover_badge_style,omitempty"`
	CoverBadgeColor model.CoverColor `json:"cover_badge_color,omitempty"`
}

func (c *ThreeItemHV4) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int32]*season.CardInfoProto:
		const _limit = 3
		c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, op.Goto, "", nil)
		c.Items = make([]*ThreeItemHV4Item, 0, _limit)
		for _, v := range op.Items {
			if v == nil {
				continue
			}
			var (
				item *ThreeItemHV4Item
			)
			sm := main.(map[int32]*season.CardInfoProto)
			s, ok := sm[int32(v.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceSeason, v.ID)
			}
			item = &ThreeItemHV4Item{
				Title:           s.Title,
				Cover:           s.Cover,
				Goto:            model.GotoPGC,
				URI:             model.FillURI(model.GotoPGC, 0, 0, strconv.FormatInt(int64(s.SeasonId), 10), nil),
				Param:           strconv.FormatInt(int64(s.SeasonId), 10),
				CoverBadge:      s.Badge,
				CoverBadgeStyle: reasonStyleFrom(model.BgColorRed, s.Badge),
				// CoverBadgeColor: model.PurpleCoverBadge,
				// Desc:SeasonTypeName + " · " +
			}
			if s.Rating != nil && s.Rating.Score > 0 {
				item.Desc = fmt.Sprintf("%s · %.1f分", s.SeasonTypeName, s.Rating.Score)
			}
			c.Items = append(c.Items, item)
			if len(c.Items) == _limit {
				break
			}
		}
		if len(c.Items) > _limit {
			// c.MoreText = "查看更多"
			// c.MoreURI = model.FillURI(op.Goto, 0, 0, op.URI, nil)
			c.Items = c.Items[:_limit]
		}
		if len(c.Items) < _limit {
			return newInvalidResourceErr(ResourceItems, 0, "lack of items")
		}
	default:
		log.Warn("ThreeItemHV4Item From: unexpected card_goto %s", op.CardGoto)
		return newUnexpectedCardGotoErr(string(op.CardGoto), "")
	}
	c.Right = true
	return nil
}

func (c *ThreeItemHV4) Get() *Base {
	return c.Base
}

type UpRcmdCover struct {
	*Base
	CoverType    model.Type `json:"cover_type,omitempty"`
	Level        int32      `json:"level,omitempty"`
	OfficialIcon model.Icon `json:"official_icon,omitempty"`
	DescButton   *Button    `json:"desc_button,omitempty"`
	Desc1        string     `json:"desc_1,omitempty"`
	Desc2        string     `json:"desc_2,omitempty"`
	Desc3        string     `json:"desc_3,omitempty"`
}

func (c *UpRcmdCover) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case nil:
		switch op.CardGoto {
		case model.CardGotoUpRcmdS:
			c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), "", "", model.GotoMid, strconv.FormatInt(op.ID, 10), nil)
			var (
				button interface{}
			)
			cd, ok := c.Cardm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceAccount, op.ID)
			}
			c.Cover = cd.Face
			c.CoverType = model.AvatarRound
			c.Title = cd.Name
			c.Level = cd.Level
			c.OfficialIcon = model.OfficialIcon(cd)
			if stat, ok := c.Statm[cd.Mid]; ok {
				c.Desc1 = "粉丝: " + model.StatString(int32(stat.Follower), "")
			}
			c.Desc2 = "视频: " + strconv.Itoa(op.Limit)
			c.Desc3 = cd.Sign
			button = &ButtonStatus{
				Goto:    model.GotoMid,
				Param:   strconv.FormatInt(cd.Mid, 10),
				IsAtten: c.IsAttenm[op.ID],
				Event:   model.EventUpClick,
				EventV2: model.EventV2UpClick,
			}
			c.DescButton = buttonFrom(button, op.Plat)
		default:
			log.Warn("UpRcmdCover From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
		c.Right = true
		return nil
	}
	return nil
}

func (c *UpRcmdCover) Get() *Base {
	return c.Base
}

type ThreeItemAll struct {
	*Base
	Items []*ThreeItemAllItem `json:"items,omitempty"`
}

type ThreeItemAllItem struct {
	Base
	CoverType    model.Type `json:"cover_type,omitempty"`
	Desc         string     `json:"desc,omitempty"`
	DescButton   *Button    `json:"desc_button,omitempty"`
	OfficialIcon model.Icon `json:"official_icon,omitempty"`
	VipType      int32      `json:"vip_type,omitempty"`
	// 愚人节活动用
	Label accountgrpc.VipLabel `json:"label"`
}

func (c *ThreeItemAll) From(main interface{}, op *operate.Card) error {
	switch main.(type) {
	case nil:
		if op == nil {
			return newEmptyOPErr()
		}
		switch op.CardGoto {
		case model.CardGotoSearchUpper:
			const _limit = 3
			c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, "", "", nil)
			c.Items = []*ThreeItemAllItem{}
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				var (
					item   *ThreeItemAllItem
					button interface{}
				)
				switch v.Goto {
				case model.GotoMid:
					cd, ok := c.Cardm[v.ID]
					if !ok || c.IsAttenm[v.ID] == 1 || model.RelationOldChange(v.ID, c.AuthorRelations) == 1 {
						continue
					}
					item = &ThreeItemAllItem{
						CoverType: model.AvatarRound,
					}
					item.Base.from(op.Plat, op.Build, v.Param, cd.Face, cd.Name, v.Goto, v.URI, nil)
					button = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(cd.Mid, 10), Relation: model.RelationChange(v.ID, c.AuthorRelations)}
					if v.Desc != "" {
						item.Desc = v.Desc
					} else if stat, ok := c.Statm[cd.Mid]; ok {
						item.Desc = model.FanString(int32(stat.Follower))
					}
					item.OfficialIcon = model.OfficialIcon(cd)
					if item.OfficialIcon == 0 {
						switch cd.Vip.Type {
						case 1:
							item.OfficialIcon = model.IconRoleVipRed
						case 2:
							item.OfficialIcon = model.IconRoleYearVipRed
						}
					}
					item.VipType = cd.Vip.Type // add vip type to mark the nickname of Big Member in red
					item.Label = cd.Vip.Label
				default:
					log.Warn("ThreeItemAll From: unexpected type %T", v.Goto)
					continue
				}
				item.DescButton = buttonFrom(button, op.Plat)
				mid, _ := strconv.ParseInt(v.Param, 10, 64)
				item.DescButton.Relation = model.RelationChange(mid, c.Base.AuthorRelations)
				c.Items = append(c.Items, item)
			}
			if len(c.Items) < _limit {
				return newInvalidResourceErr(ResourceItems, 0, "lack of items")
			}
		default:
			log.Warn("ThreeItemAll From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
		c.Right = true
		return nil
	}
	return nil
}

func (c *ThreeItemAll) Get() *Base {
	return c.Base
}

type ChannelSquare struct {
	*Base
	Desc1 string               `json:"desc_1,omitempty"`
	Desc2 string               `json:"desc_2,omitempty"`
	Item  []*ChannelSquareItem `json:"item,omitempty"`
}

type ChannelSquareItem struct {
	Title          string     `json:"title,omitempty"`
	Cover          string     `json:"cover,omitempty"`
	URI            string     `json:"uri,omitempty"`
	Param          string     `json:"param,omitempty"`
	Goto           string     `json:"goto,omitempty"`
	CoverLeftText1 string     `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string     `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3 string     `json:"cover_left_text_3,omitempty"`
	FromType       string     `json:"from_type"`
}

// From ChannelSquare op:channel--av对应关系, main:av map, c.base.tagm:tag map
func (c *ChannelSquare) From(main interface{}, op *operate.Card) error {
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*api.Arc:
		t := c.Base.Tagm[op.ID]
		c.Base.from(op.Plat, op.Build, op.Param, t.Cover, t.Name, model.GotoTag, op.Param, nil)
		button := &ButtonStatus{Goto: model.GotoTag, IsAtten: int8(t.Attention)}
		c.DescButton = buttonFrom(button, op.Plat)
		c.Desc1 = t.Content
		c.Desc2 = model.SubscribeString(int32(t.Sub))
		for _, item := range op.Items {
			am := main.(map[int64]*api.Arc)
			av := am[item.ID]
			c.Item = append(c.Item, &ChannelSquareItem{
				Title:          av.Title,
				Cover:          av.Pic,
				URI:            model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10), model.ArcPlayHandler(av, nil, "", nil, op.Build, op.MobiApp, false)),
				Goto:           string(model.GotoAv),
				Param:          strconv.FormatInt(item.ID, 10),
				CoverLeftText1: model.StatString(av.Stat.View, ""),
				CoverLeftIcon1: model.IconPlay,
				CoverLeftText2: model.StatString(av.Stat.Danmaku, ""),
				CoverLeftIcon2: model.IconDanmaku,
				CoverLeftText3: model.DurationString(av.Duration),
				FromType:       item.FromType,
			})
		}
	}
	c.Right = true
	return nil
}

func (c *ChannelSquare) Get() *Base {
	return c.Base
}

type TwoItemHV1 struct {
	*Base
	Desc       string            `json:"desc,omitempty"`
	DescButton *Button           `json:"desc_button,omitempty"`
	Items      []*TwoItemHV1Item `json:"item,omitempty"`
}

type TwoItemHV1Item struct {
	Title          string     `json:"title,omitempty"`
	Cover          string     `json:"cover,omitempty"`
	URI            string     `json:"uri,omitempty"`
	Param          string     `json:"param,omitempty"`
	Bvid           string     `json:"bvid,omitempty"`
	Args           Args       `json:"args,omitempty"`
	Goto           string     `json:"goto,omitempty"`
	CoverLeftText1 string     `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon `json:"cover_left_icon_1,omitempty"`
	CoverRightText string     `json:"cover_right_text,omitempty"`
}

func (c *TwoItemHV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	au, ok := c.Cardm[op.ID]
	if !ok {
		return newResourceNotExistErr(ResourceAccount, op.ID)
	}
	c.Base.from(op.Plat, op.Build, op.Param, au.Face, au.Name, model.GotoMid, op.Param, nil)
	button := &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(op.ID, 10), IsAtten: c.IsAttenm[op.ID]}
	c.DescButton = buttonFrom(button, op.Plat)
	if op.Desc != "" {
		c.Desc = op.Desc
	} else {
		c.Desc = au.Sign
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		for _, item := range op.Items {
			am := main.(map[int64]*arcgrpc.ArcPlayer)
			var (
				a  *arcgrpc.ArcPlayer
				ok bool
			)
			if a, ok = am[item.ID]; !ok {
				continue
			}
			args := Args{}
			args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			c.Items = append(c.Items, &TwoItemHV1Item{
				Title:          a.Arc.Title,
				Cover:          a.Arc.Pic,
				URI:            model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, false)),
				Goto:           string(model.GotoAv),
				Param:          strconv.FormatInt(item.ID, 10),
				CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
				CoverLeftIcon1: model.IconPlay,
				CoverRightText: model.DurationString(a.Arc.Duration),
				Args:           args,
			})
			if len(c.Items) >= 2 {
				break
			}
		}
		if len(c.Items) < 2 {
			return newInvalidResourceErr(ResourceItems, 0, "lack of items")
		}
	}
	c.Right = true
	return nil
}

func (c *TwoItemHV1) Get() *Base {
	return c.Base
}

type OnePicV1 struct {
	*Base
	Desc1                     string
	Desc2                     string
	Avatar                    *Avatar          `json:"avatar,omitempty"`
	CoverLeftText1            string           `json:"cover_left_text_1,omitempty"`
	CoverLeftText2            string           `json:"cover_left_text_2,omitempty"`
	CoverRightText            string           `json:"cover_right_text,omitempty"`
	CoverRightBackgroundColor string           `json:"cover_right_background_color,omitempty"`
	CoverBadge                string           `json:"cover_badge,omitempty"`
	CoverBadgeStyle           *ReasonStyle     `json:"cover_badge_style,omitempty"`
	TopRcmdReason             string           `json:"top_rcmd_reason,omitempty"`
	BottomRcmdReason          string           `json:"bottom_rcmd_reason,omitempty"`
	Desc                      string           `json:"desc,omitempty"`
	OfficialIcon              model.Icon       `json:"official_icon,omitempty"`
	CanPlay                   int32            `json:"can_play,omitempty"`
	CoverBadgeColor           model.CoverColor `json:"cover_badge_color,omitempty"`
	TopRcmdReasonStyle        *ReasonStyle     `json:"top_rcmd_reason_style,omitempty"`
	BottomRcmdReasonStyle     *ReasonStyle     `json:"bottom_rcmd_reason_style,omitempty"`
}

func (c *OnePicV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button interface{}
		avatar *AvatarStatus
		upID   int64
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*bplus.Picture:
		pm := main.(map[int64]*bplus.Picture)
		p, ok := pm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourcePicture, op.ID)
		}
		if len(p.Imgs) == 0 {
			return newInvalidResourceErr(ResourcePicture, p.DynamicID, "empty `Imgs`")
		}
		if p.ViewCount == 0 {
			return newInvalidResourceErr(ResourcePicture, p.DynamicID, "ViewCount: %d", p.ViewCount)
		}
		c.Base.from(op.Plat, op.Build, op.Param, p.Imgs[0], p.DynamicText, model.GotoPicture, strconv.FormatInt(p.DynamicID, 10), nil)
		c.CoverBadge = "动态"
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
			c.CoverBadge = "活动"
		}
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.LikeString(p.LikeCount)
		} else {
			c.CoverLeftText1 = model.PictureViewString(p.ViewCount)
		}
		c.CoverLeftText2 = model.ArticleReplyString(p.CommentCount)
		if p.ImgCount > 1 {
			c.CoverRightText = model.PictureCountString(p.ImgCount)
			c.CoverRightBackgroundColor = "#66666666"
		}
		c.Desc1 = p.NickName
		c.Desc2 = model.PubDataByRequestAt(p.PublishTime.Time(), c.Rcmd.RequestAt())
		avatar = &AvatarStatus{Cover: p.FaceImg, Goto: model.GotoDynamicMid, Param: strconv.FormatInt(p.Mid, 10), Type: model.AvatarRound}
		button = p
		upID = p.Mid
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, c.CoverBadge)
		c.Args.fromPicture(p)
	default:
		log.Warn("OnePicV1 From: unexpected type %T", main)
	}
	if c.Rcmd != nil {
		c.TopRcmdReason, c.BottomRcmdReason = TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
		c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op)
		c.BottomRcmdReasonStyle = bottomReasonStyleFrom(c.Rcmd, c.BottomRcmdReason, c.Base.Goto, op)
	}
	c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	c.Avatar = avatarFrom(avatar)
	c.DescButton = buttonFrom(button, op.Plat)
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	c.Right = true
	return nil
}

func (c *OnePicV1) Get() *Base {
	return c.Base
}

type ThreePicV1 struct {
	*Base
	Covers                    []string     `json:"covers,omitempty"`
	Desc1                     string       `json:"desc_1,omitempty"`
	Desc2                     string       `json:"desc_2,omitempty"`
	Avatar                    *Avatar      `json:"avatar,omitempty"`
	TitleLeftText1            string       `json:"title_left_text_1,omitempty"`
	TitleLeftText2            string       `json:"title_left_text_2,omitempty"`
	CoverRightText            string       `json:"cover_right_text,omitempty"`
	CoverRightBackgroundColor string       `json:"cover_right_background_color,omitempty"`
	TopRcmdReason             string       `json:"top_rcmd_reason,omitempty"`
	BottomRcmdReason          string       `json:"bottom_rcmd_reason,omitempty"`
	TopRcmdReasonStyle        *ReasonStyle `json:"top_rcmd_reason_style,omitempty"`
	BottomRcmdReasonStyle     *ReasonStyle `json:"bottom_rcmd_reason_style,omitempty"`
	CoverBadge                string       `json:"cover_badge,omitempty"`
	CoverBadgeStyle           *ReasonStyle `json:"cover_badge_style,omitempty"`
}

func (c *ThreePicV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button interface{}
		avatar *AvatarStatus
		upID   int64
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*bplus.Picture:
		pm := main.(map[int64]*bplus.Picture)
		p, ok := pm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourcePicture, op.ID)
		}
		if len(p.Imgs) < 3 {
			return newInvalidResourceErr(ResourcePicture, p.DynamicID, "lack of `Imgs`")
		}
		if p.ViewCount == 0 {
			return newInvalidResourceErr(ResourcePicture, p.DynamicID, "ViewCount: %d", p.ViewCount)
		}
		c.Base.from(op.Plat, op.Build, op.Param, "", p.DynamicText, model.GotoPicture, strconv.FormatInt(p.DynamicID, 10), nil)
		c.CoverBadge = "动态"
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
			c.CoverBadge = "活动"
		}
		c.Covers = p.Imgs[:3]
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.TitleLeftText1 = model.LikeString(p.LikeCount)
		} else {
			c.TitleLeftText1 = model.PictureViewString(p.ViewCount)
		}
		c.TitleLeftText2 = model.ArticleReplyString(p.CommentCount)
		if p.ImgCount > 3 {
			c.CoverRightText = model.PictureCountString(p.ImgCount)
			c.CoverRightBackgroundColor = "#66666666"
		}
		c.Desc1 = p.NickName
		c.Desc2 = model.PubDataByRequestAt(p.PublishTime.Time(), c.Rcmd.RequestAt())
		avatar = &AvatarStatus{Cover: p.FaceImg, Goto: model.GotoDynamicMid, Param: strconv.FormatInt(p.Mid, 10), Type: model.AvatarRound}
		button = p
		upID = p.Mid
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, c.CoverBadge)
		c.Args.fromPicture(p)
	default:
		log.Warn("ThreePicV1 From: unexpected type %T", main)
	}
	if c.Rcmd != nil {
		c.TopRcmdReason, c.BottomRcmdReason = TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
		c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.TopRcmdReason, c.Base.Goto, op)
		c.BottomRcmdReasonStyle = bottomReasonStyleFrom(c.Rcmd, c.BottomRcmdReason, c.Base.Goto, op)
	}
	c.Avatar = avatarFrom(avatar)
	c.DescButton = buttonFrom(button, op.Plat)
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	c.Right = true
	return nil
}

func (c *ThreePicV1) Get() *Base {
	return c.Base
}

type SmallCoverV5 struct {
	*Base
	CoverGif        string           `json:"cover_gif,omitempty"`
	Up              *Up              `json:"up,omitempty"`
	CoverRightText1 string           `json:"cover_right_text_1,omitempty"`
	RightDesc1      string           `json:"right_desc_1,omitempty"`
	RightDesc2      string           `json:"right_desc_2,omitempty"`
	CanPlay         int32            `json:"can_play,omitempty"`
	RcmdReasonStyle *ReasonStyle     `json:"rcmd_reason_style,omitempty"`
	HotwordEntrance *HotwordEntrance `json:"hotword_entrance,omitempty"`
	IsPopular       bool             `json:"is_popular"`
}

type HotwordEntrance struct {
	HotwordID int64  `json:"hotword_id,omitempty"`
	HotText   string `json:"hot_text,omitempty"`
	H5URL     string `json:"h5_url,omitempty"`
	Icon      string `json:"icon,omitempty"`
}

type Up struct {
	ID           int64      `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Desc         string     `json:"desc,omitempty"`
	Avatar       *Avatar    `json:"avatar,omitempty"`
	OfficialIcon model.Icon `json:"official_icon,omitempty"`
	DescButton   *Button    `json:"desc_button,omitempty"`
	Cooperation  string     `json:"cooperation,omitempty"`
}

func (c *SmallCoverV5) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button     interface{}
		avatar     *AvatarStatus
		rcmdReason string
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		c.CoverRightText1 = model.DurationString(a.Arc.Duration)
		if c.Rcmd != nil {
			rcmdReason, _ = TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
		}
		switch op.CardGoto {
		case model.CardGotoAv:
			authorface := a.Arc.Author.Face
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			c.Bvid, _ = GetBvIDStr(c.Param)
			if op.IsPopular {
				c.IsPopular = true
			}
			switch c.Rcmd.Style {
			case model.HotCardStyleShowUp:
				c.Up = &Up{
					ID:   a.Arc.Author.Mid,
					Name: authorname,
				}
				if stat, ok := c.Statm[a.Arc.Author.Mid]; ok {
					c.Up.Desc = model.AttentionString(int32(stat.Follower))
				}
				avatar = &AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
				c.Up.Avatar = avatarFrom(avatar)
				c.Up.OfficialIcon = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
				c.RightDesc1 = model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.PubDataString(a.Arc.PubDate.Time())
				button = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
				c.Up.DescButton = buttonFrom(button, op.Plat)
				if a.Arc.Rights.IsCooperation > 0 {
					c.Up.Cooperation = "等联合创作"
				}
			default:
				c.CoverGif = c.Rcmd.CoverGif
				if op.Switch != model.SwitchCooperationHide {
					c.RightDesc1 = unionAuthorGRPC(a, authorname)
				} else {
					c.RightDesc1 = authorname
				}
				c.RightDesc2 = model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.PubDataString(a.Arc.PubDate.Time())
			}
			// c.CanPlay = a.Rights.Autoplay
			if op.ShowHotword {
				c.HotwordEntrance = &HotwordEntrance{ // 热门热点
					HotwordID: op.Tid,
					Icon:      op.Cover,
					H5URL:     op.RedirectURL,
					HotText:   op.Subtitle,
				}
			}
		default:
			log.Warn("SmallCoverV5 From: unexpected type %T", main)
			return newUnexpectedResourceTypeErr(main, "")
		}
	default:
		log.Warn("SmallCoverV5 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverV5) Get() *Base {
	return c.Base
}

// SmallCoverH5 is the small card for h5 popular selected page
type SmallCoverH5 struct {
	*Base
	CoverRightText1 string `json:"cover_right_text_1,omitempty"`
	RightDesc1      string `json:"right_desc_1,omitempty"`
	RightDesc2      string `json:"right_desc_2,omitempty"`
	RcmdReason      string `json:"rcmd_reason,omitempty"`
	AuthorID        int64  `json:"author_id"`
	ShortLink       string `json:"short_link"`
}

func (c *SmallCoverH5) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main.(map[int64]*arcgrpc.Arc)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(&arcgrpc.ArcPlayer{Arc: a}) { // filter abnormal archive
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(0, 0, op.Param, a.Pic, a.Title, model.GotoAv, op.URI, nil)
		c.RcmdReason = op.Desc
		c.CoverRightText1 = model.DurationString(a.Duration)
		switch op.CardGoto {
		case model.CardGotoAv:
			authorname := a.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Author.Mid]; ok {
					authorname = au.Name
				}
			}
			c.ShortLink = a.ShortLinkV2
			c.RightDesc1 = authorname
			c.RightDesc2 = model.ArchiveViewString(a.Stat.View) + " · " + model.PubDataString(a.PubDate.Time())
			c.AuthorID = a.Author.Mid
			c.Bvid, _ = GetBvIDStr(c.Param)
		default:
			log.Warn("SmallCoverH5 From: unexpected type %T", main)
			return newUnexpectedResourceTypeErr(main, "")
		}
	}
	c.Right = true
	return nil
}

func (c *SmallCoverH5) Get() *Base {
	return c.Base
}

// SmallCoverH6 town station treasure page .
type SmallCoverH6 struct {
	*Base
	RightDesc1  string `json:"right_desc_1,omitempty"`
	RightDesc2  string `json:"right_desc_2,omitempty"`
	AuthorID    int64  `json:"author_id,omitempty"`
	RightDesc3  string `json:"right_desc_3,omitempty"`
	Achievement string `json:"achievement,omitempty"`
}

// From .
func (c *SmallCoverH6) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main.(map[int64]*arcgrpc.Arc)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(&arcgrpc.ArcPlayer{Arc: a}) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(0, 0, op.Param, a.Pic, a.Title, model.GotoAv, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoAv:
			c.RightDesc1 = a.Author.Name
			c.RightDesc2 = model.ArchiveViewString(a.Stat.View) + " · " + model.DanmakuString(a.Stat.Danmaku)
			c.AuthorID = a.Author.Mid
			c.RightDesc3 = model.PubDataString(a.PubDate.Time())
			c.Achievement = op.Desc
			c.Bvid, _ = GetBvIDStr(c.Param)
		default:
			log.Warn("SmallCoverH6 From: unexpected type %T", main)
			return newUnexpectedResourceTypeErr(main, "")
		}
	}
	c.Right = true
	return nil
}

// Get .
func (c *SmallCoverH6) Get() *Base {
	return c.Base
}

// SmallCoverH7 town station treasure page .
type SmallCoverH7 struct {
	*Base
	RightDesc1      string `json:"right_desc_1,omitempty"`
	RightDesc2      string `json:"right_desc_2,omitempty"`
	RightDesc3      string `json:"right_desc_3,omitempty"`
	AuthorID        int64  `json:"author_id,omitempty"`
	CoverRightText1 string `json:"cover_right_text_1,omitempty"`
	RcmdReason      string `json:"rcmd_reason,omitempty"`
}

// From .
func (c *SmallCoverH7) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main.(map[int64]*arcgrpc.Arc)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(&arcgrpc.ArcPlayer{Arc: a}) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(0, 0, op.Param, a.Pic, a.Title, model.GotoAv, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoAv:
			c.RightDesc1 = a.Author.Name
			c.RightDesc2 = model.ArchiveViewString(a.Stat.View) + " · " + model.PubDataString(a.PubDate.Time())
			c.AuthorID = a.Author.Mid
			c.RcmdReason = op.Desc
			c.CoverRightText1 = model.DurationString(a.Duration)
			c.Bvid, _ = GetBvIDStr(c.Param)
		default:
			log.Warn("SmallCoverH7 From: unexpected type %T", main)
			return newUnexpectedResourceTypeErr(main, "")
		}
	}
	c.Right = true
	return nil
}

// Get .
func (c *SmallCoverH7) Get() *Base {
	return c.Base
}

// Option struct.
type Option struct {
	*Base
	Option []string `json:"option,omitempty"`
}

// From is.
func (c *Option) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case []string:
		os := main.([]string)
		if len(os) == 0 {
			return newResourceNotExistErr(ResourceOption, 0)
		}
		c.Base.from(op.Plat, op.Build, op.Param, "", "选择感兴趣的内容", "", "", nil)
		c.Option = os
		c.DescButton = &Button{Text: "选好啦，刷新首页"}
	default:
		log.Warn("Option From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

// Get is.
func (c *Option) Get() *Base {
	return c.Base
}

type MiddleCoverV3 struct {
	*Base
	Desc1      string       `json:"desc1,omitempty"`
	Desc2      string       `json:"desc2,omitempty"`
	CoverBadge *ReasonStyle `json:"cover_badge_style,omitempty"`
}

func (c *MiddleCoverV3) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoWeb, op.URI, nil)
	c.Goto = op.Goto
	if op.Badge != "" {
		c.CoverBadge = reasonStyleFrom(model.BgColorPurple, op.Badge)
	}
	c.Desc1 = op.Desc
	c.Right = true
	return nil
}

func (c *MiddleCoverV3) Get() *Base {
	return c.Base
}

type Select struct {
	*Base
	Desc        string  `json:"desc,omitempty"`
	LeftButton  *Button `json:"left_button,omitempty"`
	RightButton *Button `json:"right_button,omitempty"`
}

func (c *Select) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case nil:
		switch op.CardGoto {
		case model.CardGotoFollowMode:
			if len(op.Buttons) < 2 {
				return newInvalidResourceErr(ResourceOperateCard, 0, "lack of `Buttons`")
			}
			c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, "", "", nil)
			c.Desc = op.Desc
			c.LeftButton = buttonFrom(&ButtonStatus{Text: op.Buttons[0].Text, Event: model.Event(op.Buttons[0].Event)}, op.Plat)
			c.RightButton = buttonFrom(&ButtonStatus{Text: op.Buttons[1].Text, Event: model.Event(op.Buttons[1].Event)}, op.Plat)
		default:
			log.Warn("Select From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("Select From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *Select) Get() *Base {
	return c.Base
}

type SmallCoverV6 struct {
	*Base
	Desc1 string `json:"desc_1,omitempty"`
}

func (c *SmallCoverV6) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case *vipgrpc.TipsRenewReply:
		vipInfo := main.(*vipgrpc.TipsRenewReply)
		if vipInfo == nil {
			return newResourceNotExistErr(ResourceVipInfo, 0)
		}
		if op.Title == "" {
			return newInvalidResourceErr(ResourceOperateCard, 0, "empty `Title`")
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoWeb, vipInfo.Link, nil)
		c.CardGoto = model.CardGotoVip
		c.Desc1 = vipInfo.Tip
	case *tunnelgrpc.FeedCard:
		tunnelCard := main.(*tunnelgrpc.FeedCard)
		if tunnelCard == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		c.Base.from(op.Plat, op.Build, op.Param, tunnelCard.Cover, tunnelCard.Title, model.GotoWeb, tunnelCard.Link, nil)
		c.Goto = model.GotoGame
		c.Desc1 = tunnelCard.Intro
	default:
		log.Warn("SmallCoverV6 From: unexpected card_goto %s", op.CardGoto)
		return newUnexpectedCardGotoErr(string(op.CardGoto), "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverV6) Get() *Base {
	return c.Base
}

type SmallCoverV8 struct {
	*Base
	CoverBadge            string       `json:"cover_badge,omitempty"`
	CoverBadgeStyle       *ReasonStyle `json:"cover_badge_style,omitempty"`
	RightDesc1            string       `json:"right_desc_1,omitempty"`
	RightDesc2            string       `json:"right_desc_2,omitempty"`
	CoverRightText        string       `json:"cover_right_text,omitempty"`
	BottomRcmdReasonStyle *ReasonStyle `json:"bottom_rcmd_reason_style,omitempty"`
	TopRcmdReasonStyle    *ReasonStyle `json:"top_rcmd_reason_style,omitempty"`
}

func (c *SmallCoverV8) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		upID int64
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		var (
			authorname = a.Arc.Author.Name
		)
		if c.Cardm != nil {
			if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
				authorname = au.Name
			}
		}
		if op.Switch != model.SwitchCooperationHide {
			c.RightDesc1 = unionAuthorGRPC(a, authorname)
		} else {
			c.RightDesc1 = authorname
		}
		c.RightDesc2 = model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.DanmakuString(a.Arc.Stat.Danmaku)
		c.CoverRightText = model.DurationString(a.Arc.Duration)
		c.Base.PlayerArgs = playerArgsFrom(a)
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		upID = a.Arc.Author.Mid
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if r.LiveStatus != 1 {
			return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.CoverBadge = "直播"
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, "直播")
		c.RightDesc1 = r.Uname
		c.RightDesc2 = model.LiveOnlineString(r.Online)
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		upID = r.UID
	case map[int64]*article.Meta:
		mm := main.(map[int64]*article.Meta)
		m, ok := mm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArticle, op.ID)
		}
		if len(m.ImageURLs) == 0 {
			return newInvalidResourceErr(ResourceArticle, m.ID, "empty `ImageURLs`")
		}
		c.Base.from(op.Plat, op.Build, op.Param, m.ImageURLs[0], m.Title, model.GotoArticle, strconv.FormatInt(m.ID, 10), nil)
		c.CoverBadge = "文章"
		c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, "文章")
		if m.Author != nil {
			c.RightDesc1 = m.Author.Name
			upID = m.Author.Mid
		}
		if m.Author != nil {
			c.RightDesc2 = model.ArticleViewString(m.Stats.View) + " · " + model.ArticleReplyString(m.Stats.Reply)
		}
	default:
		log.Warn("SmallCoverV8 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
		topRcmdReason, bttomRcmdReason := TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
		if (op.MobiApp == "iphone" && op.Build > 8740 || op.MobiApp == "android" && op.Build > 5455000) && topRcmdReason != "" {
			bttomRcmdReason = topRcmdReason
			topRcmdReason = ""
			c.RightDesc1 = ""
		}
		if c.Rcmd.RcmdReason != nil && c.Rcmd.RcmdReason.Style == 5 {
			c.RightDesc1 = ""
		}
		c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, topRcmdReason, c.Base.Goto, op)
		c.BottomRcmdReasonStyle = bottomReasonStyleFrom(c.Rcmd, bttomRcmdReason, c.Base.Goto, op)
	}
	c.Right = true
	return nil
}

func (c *SmallCoverV8) Get() *Base {
	return c.Base
}

type Introduction struct {
	*Base
}

func (c *Introduction) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	c.Title = op.Title
	c.Right = true
	return nil
}

func (c *Introduction) Get() *Base {
	return c.Base
}

type SmallCoverConvergeV1 struct {
	*Base
	CoverLeftText1    string       `json:"cover_left_text_1,omitempty"`
	CoverRightTopText string       `json:"cover_right_top_text,omitempty"`
	RightDesc1        string       `json:"right_desc_1,omitempty"`
	RcmdReasonStyle   *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	RcmdReasonStyleV2 *ReasonStyle `json:"rcmd_reason_style_v2,omitempty"`
}

func (c *SmallCoverConvergeV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		isShowV2 bool
		title    string
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
		}
		c.RightDesc1 = model.ArchiveViewString(a.Arc.Stat.View) + "  " + model.DanmakuString(a.Arc.Stat.Danmaku)
		title = a.Arc.Title
		if c.Rcmd != nil {
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorname = au.Name
				}
			}
			rcmdReason, _ := rcmdReason(c.Rcmd.RcmdReason, authorname, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			if isShowV2 = IsShowRcmdReasonStyleV2(c.Rcmd); isShowV2 {
				if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV4(c.Rcmd, rcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV3(c.Rcmd, rcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else {
					c.RcmdReasonStyleV2 = reasonStyleFromV2(c.Rcmd, rcmdReason, c.Base.Goto, op.Plat, op.Build)
				}
			} else {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
			}
			if cinfo := c.Rcmd.ConvergeInfo; cinfo != nil {
				if cinfo.Quality.Play > 0 && cinfo.Quality.Danmu > 0 {
					c.RightDesc1 = model.ArchiveView64String(cinfo.Quality.Play) + "  " + model.Danmaku64String(cinfo.Quality.Danmu)
				} else if view := cinfo.Quality.Play; view > 0 {
					c.RightDesc1 = model.ArchiveView64String(cinfo.Quality.Play)
				} else if danmaku := cinfo.Quality.Danmu; danmaku > 0 {
					c.RightDesc1 = model.Danmaku64String(danmaku)
				}
				if cinfo.Count > 0 {
					c.CoverRightTopText = strconv.Itoa(cinfo.Count) + "个内容"
				}
				if cinfo.Title != "" {
					title = cinfo.Title
				}
			}
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		c.CoverLeftText1 = model.DurationString(a.Arc.Duration)
		switch op.Goto {
		case model.GotoAv:
			caid, _ := strconv.ParseInt(op.Param, 10, 64)
			ac, ok := am[caid]
			if !ok {
				return newResourceNotExistErr(ResourceArchive, caid)
			}
			if !model.AvIsNormalGRPC(ac) {
				return newInvalidResourceErr(ResourceArchive, caid, "AvIsNormalGRPC")
			}
			c.Base.PlayerArgs = playerArgsFrom(ac)
			switch op.Plat {
			case model.PlatIPhone, model.PlatIPad, model.PlatIPhoneB:
				c.Base.URI = url.QueryEscape(model.FillURI(op.Goto, op.Plat, op.Build, op.Param, model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(ac, 0), op.TrackID, nil, op.Build, op.MobiApp, true)))
			default:
				c.Base.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.Param, model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(ac, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
			}
		case model.GotoAvConverge:
			if c.Rcmd.ID != 0 {
				c.Param = strconv.FormatInt(c.Rcmd.ID, 10)
			}
			c.Base.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.Param, model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build))
		default:
			c.Base.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.Param, model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build))
		}
		c.Goto = op.Goto
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		c.Args.fromAvConverge(c.Base.Rcmd)
	default:
		log.Warn("SmallCoverConvergeV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverConvergeV1) Get() *Base {
	return c.Base
}

type ChannelNew struct {
	*Base
	DescButton2 *Button           `json:"desc_button_2,omitempty"`
	Desc1       string            `json:"desc_1,omitempty"`
	Items       []*ChannelNewItem `json:"items,omitempty"`
}

type ChannelNewItem struct {
	Title          string        `json:"title,omitempty"`
	Cover          string        `json:"cover,omitempty"`
	URI            string        `json:"uri,omitempty"`
	Param          string        `json:"param,omitempty"`
	Goto           string        `json:"goto,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon    `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon    `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3 string        `json:"cover_left_text_3,omitempty"`
	Badge          *ChannelBadge `json:"badge,omitempty"`
	LeftText1      string        `json:"left_text_1,omitempty"`
	Position       int64         `json:"position,omitempty"`
}

//nolint:gocognit
func (c *ChannelNew) From(main interface{}, op *operate.Card) error {
	var gt model.Gt
	switch op.Channel.CType {
	case model.OldChannel:
		gt = model.GotoTag
	case model.NewChannel:
		gt = model.GotoChannel
	default:
		return newUnexpectedResourceTypeErr(op.Channel.CType, "invalid Ctype=%d", op.Channel.CType)
	}
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, gt, op.Param, model.ChannelHandler("tab=all&sort=hot"))
	var upTime string
	if op.Channel.LastUpTime > 0 {
		upTime = model.PubDataString(time.Time(op.Channel.LastUpTime).Time())
	}
	if upTime != "" {
		c.Desc1 = fmt.Sprintf("%v更新", upTime)
	} else {
		c.Desc1 = "最近更新"
	}
	var buttonText string
	if op.Channel.UpCnt > 0 {
		buttonText = fmt.Sprintf("+%d 进入频道", op.Channel.UpCnt)
	} else {
		buttonText = "进入频道"
	}
	buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=all&sort=hot"))
	c.DescButton = &Button{Text: buttonText, URI: buttonURI}
	if op.Channel.FeatureCnt > 0 {
		buttonText := "精选视频" + model.StatString(op.Channel.FeatureCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=select"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	} else if op.Channel.TodayCnt > 0 {
		buttonText := "今日新投稿" + model.StatString(op.Channel.TodayCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=all&sort=new"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	}
	//nolint:gosimple
	var position int64
	position = op.Channel.Position
	for _, item := range op.Items {
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, item.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, item.ID, "AvIsNormalGRPC")
		}
		// 新频道广场页 我订阅的更新 视频角标\用户行为
		var (
			badge     *ChannelBadge
			letfText1 string
		)
		if op.Channel != nil {
			if op.Channel.Badges != nil {
				if bdg, ok := op.Channel.Badges[item.ID]; ok {
					if bdg != nil && bdg.Text != "" && bdg.Cover != "" {
						badge = &ChannelBadge{
							Text:    bdg.Text,
							BgCover: bdg.Cover,
						}
					}
				}
			}
			if op.Channel.IsFav != nil {
				if fav, ok := op.Channel.IsFav[item.ID]; ok && fav {
					letfText1 = fmt.Sprintf("已收藏·%s", model.DurationString(a.Arc.Duration))
				}
			}
			if letfText1 == "" {
				if op.Channel.Coins != nil {
					if coin, ok := op.Channel.Coins[item.ID]; ok && coin > 0 {
						letfText1 = fmt.Sprintf("已投币·%s", model.DurationString(a.Arc.Duration))
					}
				}
			}
		}
		position++
		c.Items = append(c.Items, &ChannelNewItem{
			Title: a.Arc.Title,
			Cover: a.Arc.Pic,
			URI: model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10),
				model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, true)),
			Goto:           string(model.GotoAv),
			Param:          strconv.FormatInt(item.ID, 10),
			CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
			CoverLeftIcon1: model.IconPlay,
			CoverLeftText2: model.StatString(a.Arc.Stat.Danmaku, ""),
			CoverLeftIcon2: model.IconDanmaku,
			CoverLeftText3: model.DurationString(a.Arc.Duration),
			Badge:          badge,
			LeftText1:      letfText1,
			Position:       position,
		})
		c.Base.Idx = position
		if len(c.Items) == 2 {
			break
		}
	}
	if len(c.Items) == 0 {
		log.Error("square guanzhu gengxin channel_id(%v) from items(%+v) is 0", op.Param, op.Items)
		return newInvalidResourceErr(ResourceItems, 0, "empty `Items`")
	}
	c.Right = true
	return nil
}

func (c *ChannelNew) Get() *Base {
	return c.Base
}

type LargeChannelSpecial struct {
	*Base
	BgCover         string       `json:"bg_cover,omitempty"`
	Desc1           string       `json:"desc_1,omitempty"`
	Desc2           string       `json:"desc_2,omitempty"`
	Badge           string       `json:"badge,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_badge_style,omitempty"`
	// RcmdReasonStyle  *ReasonStyle `json:"rcmd_reason_style_1,omitempty"`
	RcmdReasonStyle2 *ReasonStyle `json:"rcmd_reason_style_2,omitempty"`
}

func (c *LargeChannelSpecial) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	c.Desc1 = op.Desc
	if op.Channel != nil {
		c.BgCover = op.Channel.BgCover
		if op.Channel.Reason != "" {
			c.Desc2 = op.Channel.Reason
		}
	}
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoChannel, strconv.FormatInt(op.ID, 10), model.ChannelHandler("tab=all"))
	if op.Channel.TabURI != "" {
		c.URI = op.Channel.TabURI
	}
	// c.RcmdReasonStyle = reasonStyleFrom(model.BgColorRed, "频道", false)
	c.Badge = "频道"
	c.CoverBadgeStyle = reasonStyleFrom(model.BgColorRed, "频道")
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*channelgrpc.ChannelCard:
		chm := main.(map[int64]*channelgrpc.ChannelCard)
		if ch, ok := chm[op.ID]; ok && ch != nil {
			if ch.Subscribed {
				c.RcmdReasonStyle2 = reasonStyleFrom(model.BgColorOrange, "已订阅")
			}
			if c.Desc1 == "" {
				var descs []string
				if ch.RCnt != 0 {
					descs = append(descs, model.StatString(ch.RCnt, "投稿"))
				}
				if ch.FeaturedCnt != 0 {
					descs = append(descs, model.StatString(ch.FeaturedCnt, "个精选视频"))
				}
				if len(descs) > 0 {
					c.Desc1 = strings.Join(descs, " ")
				}
			}
		}
	}
	c.Right = true
	return nil
}

func (c *LargeChannelSpecial) Get() *Base {
	return c.Base
}

type ChannelThreeItemHV1 struct {
	*Base
	MoreText string                        `json:"more_text,omitempty"`
	MoreURI  string                        `json:"more_uri,omitempty"`
	Items    []*ChannelNewDetailCustomItem `json:"items,omitempty"`
}

type ChannelNewDetailCustomItem struct {
	Bvid           string        `json:"bvid,omitempty"`
	Title          string        `json:"title,omitempty"`
	Cover          string        `json:"cover,omitempty"`
	URI            string        `json:"uri,omitempty"`
	Param          string        `json:"param,omitempty"`
	Goto           string        `json:"goto,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon    `json:"cover_left_icon_1,omitempty"`
	Badge          *ChannelBadge `json:"badge,omitempty"`
	Position       int64         `json:"position,omitempty"`
	// 上报字段
	Sort string `json:"sort"`
	Filt int32  `json:"filt"`
}

func (c *ChannelThreeItemHV1) From(main interface{}, op *operate.Card) error {
	c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, model.GotoChannelDetailCustom, strconv.FormatInt(op.ID, 10), nil)
	c.URI = ""
	var position int64
	if op.Channel != nil {
		if op.Channel.CustomDesc != "" && op.Channel.CustomURI != "" {
			c.MoreText = op.Channel.CustomDesc
			c.MoreURI = op.Channel.CustomURI
		}
		position = op.Channel.Position
	}
	for _, item := range op.Items {
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, item.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, item.ID, "AvIsNormalGRPC")
		}
		// 频道详情 三联自定义视频卡角标
		var badge *ChannelBadge
		if op.Channel != nil {
			if bdg, ok := op.Channel.Badges[item.ID]; ok {
				if bdg != nil && bdg.Text != "" && bdg.Cover != "" {
					badge = &ChannelBadge{
						Text:    bdg.Text,
						BgCover: bdg.Cover,
					}
				}
			}
		}
		position++
		i := &ChannelNewDetailCustomItem{
			Title: a.Arc.Title,
			Cover: a.Arc.Pic,
			URI: model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10),
				model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, true)),
			Goto:           string(model.GotoAv),
			Param:          strconv.FormatInt(item.ID, 10),
			CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
			CoverLeftIcon1: model.IconPlay,
			Badge:          badge,
			Position:       position,
			Sort:           op.Channel.Sort,
			Filt:           op.Channel.Filt,
		}
		i.Bvid, _ = bvid.AvToBv(item.ID)
		c.Items = append(c.Items, i)
		c.Idx = position
	}
	c.Right = true
	return nil
}

func (c *ChannelThreeItemHV1) Get() *Base {
	return c.Base
}

type ChannelThreeItemHV2 struct {
	*Base
	MoreText string                         `json:"more_text,omitempty"`
	MoreURI  string                         `json:"more_uri,omitempty"`
	Items    []*ChannelNewDetailCustomItem2 `json:"items,omitempty"`
}

type ChannelNewDetailCustomItem2 struct {
	Bvid           string        `json:"bvid,omitempty"`
	Title          string        `json:"title,omitempty"`
	Cover          string        `json:"cover,omitempty"`
	URI            string        `json:"uri,omitempty"`
	Param          string        `json:"param,omitempty"`
	Goto           string        `json:"goto,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon    `json:"cover_left_icon_1,omitempty"`
	Badge          *ChannelBadge `json:"badge,omitempty"`
	Position       int64         `json:"position,omitempty"`
	// 上报字段
	Sort string `json:"sort"`
	Filt int32  `json:"filt"`
}

func (c *ChannelThreeItemHV2) From(main interface{}, op *operate.Card) error {
	c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, model.GotoChannelDetailRank, strconv.FormatInt(op.ID, 10), nil)
	c.URI = ""
	var (
		rankType int32
		position int64
	)
	if op.Channel != nil {
		if op.Channel.CustomDesc != "" && op.Channel.CustomURI != "" {
			c.MoreText = op.Channel.CustomDesc
			c.MoreURI = op.Channel.CustomURI
			rankType = op.Channel.RankType
		}
		position = op.Channel.Position
	}
	for _, item := range op.Items {
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, item.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, item.ID, "AvIsNormalGRPC")
		}
		var (
			coverLeftText  string
			coverLeftIcon1 model.Icon
		)
		switch rankType {
		case 1:
			coverLeftText = model.StatString(a.Arc.Stat.View, "")
			coverLeftIcon1 = model.IconPlay
		case 4:
			coverLeftText = model.StatString(a.Arc.Stat.Fav, "")
			coverLeftIcon1 = model.IconStar
		case 5:
			coverLeftText = model.StatString(a.Arc.Stat.Coin, "")
			coverLeftIcon1 = model.IconRoleCoin
		}
		position++
		i := &ChannelNewDetailCustomItem2{
			Title: a.Arc.Title,
			Cover: a.Arc.Pic,
			URI: model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10),
				model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, true)),
			Goto:           string(model.GotoAv),
			Param:          strconv.FormatInt(item.ID, 10),
			CoverLeftText1: coverLeftText,
			CoverLeftIcon1: coverLeftIcon1,
			Position:       position,
			Sort:           op.Channel.Sort,
			Filt:           op.Channel.Filt,
		}
		i.Bvid, _ = bvid.AvToBv(item.ID)
		c.Items = append(c.Items, i)
		c.Base.Idx = position
	}
	c.Right = true
	return nil
}

func (c *ChannelThreeItemHV2) Get() *Base {
	return c.Base
}

type ChannelScaned struct {
	*Base
	ID          int64                `json:"id,omitempty"`
	IsButton    bool                 `json:"is_button"`
	IsAtten     bool                 `json:"is_atten"`
	DescButton2 *Button              `json:"desc_button_2,omitempty"`
	Desc1       string               `json:"desc_1,omitempty"`
	Items       []*ChannelScanedItem `json:"items,omitempty"`
}

type ChannelScanedItem struct {
	Title          string        `json:"title,omitempty"`
	Cover          string        `json:"cover,omitempty"`
	URI            string        `json:"uri,omitempty"`
	Param          string        `json:"param,omitempty"`
	Goto           string        `json:"goto,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon    `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon    `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3 string        `json:"cover_left_text_3,omitempty"`
	Badge          *ChannelBadge `json:"badge,omitempty"`
	LeftText1      string        `json:"left_text_1,omitempty"`
	Position       int64         `json:"position,omitempty"`
}

func (c *ChannelScaned) From(main interface{}, op *operate.Card) error {
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoChannel, op.Param, model.ChannelHandler("tab=select"))
	c.ID = op.ID
	c.IsButton = true
	c.IsAtten = op.Channel.IsAtten
	var upTime string
	if op.Channel.LastUpTime > 0 {
		upTime = model.PubDataString(time.Time(op.Channel.LastUpTime).Time())
	}
	if upTime != "" {
		c.Desc1 = fmt.Sprintf("%v更新", upTime)
	} else {
		c.Desc1 = "最近更新"
	}
	var buttonText string
	if op.Channel.AttenCnt > 0 {
		buttonText = model.StatString(op.Channel.AttenCnt, " 订阅")
	}
	c.DescButton = &Button{Text: buttonText}
	if op.Channel.FeatureCnt > 0 {
		buttonText := "精选视频" + model.StatString(op.Channel.FeatureCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=select"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	} else if op.Channel.TodayCnt > 0 {
		buttonText := "今日新投稿" + model.StatString(op.Channel.TodayCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=all&sort=new"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	}
	//nolint:gosimple
	var position int64
	position = op.Channel.Position
	for _, item := range op.Items {
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, item.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, item.ID, "AdAvIsNormalGRPC")
		}
		var badge *ChannelBadge
		if op.Channel != nil {
			if op.Channel.Badges != nil {
				if bdg, ok := op.Channel.Badges[item.ID]; ok {
					if bdg != nil && bdg.Text != "" && bdg.Cover != "" {
						badge = &ChannelBadge{
							Text:    bdg.Text,
							BgCover: bdg.Cover,
						}
					}
				}
			}
		}
		position++
		c.Items = append(c.Items, &ChannelScanedItem{
			Title: a.Arc.Title,
			Cover: a.Arc.Pic,
			URI: model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10),
				model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, true)),
			Goto:           string(model.GotoAv),
			Param:          strconv.FormatInt(item.ID, 10),
			CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
			CoverLeftIcon1: model.IconPlay,
			CoverLeftText2: model.StatString(a.Arc.Stat.Danmaku, ""),
			CoverLeftIcon2: model.IconDanmaku,
			CoverLeftText3: model.DurationString(a.Arc.Duration),
			Badge:          badge,
			Position:       position,
		})
		c.Base.Idx = position
		if len(c.Items) == 2 {
			break
		}
	}
	if len(c.Items) == 0 {
		log.Error("square scaned channel_id(%v) from items(%+v) is 0", op.Param, op.Items)
		return newInvalidResourceErr(ResourceItems, 0, "empty `Items`")
	}
	c.Right = true
	return nil
}

func (c *ChannelScaned) Get() *Base {
	return c.Base
}

type ChannelRcmdV2 struct {
	*Base
	ID          int64                `json:"id,omitempty"`
	IsButton    bool                 `json:"is_button"`
	IsAtten     bool                 `json:"is_atten"`
	DescButton2 *Button              `json:"desc_button_2,omitempty"`
	Desc1       string               `json:"desc_1,omitempty"`
	Items       []*ChannelRcmdV2Item `json:"items,omitempty"`
}

type ChannelRcmdV2Item struct {
	Title          string        `json:"title,omitempty"`
	Cover          string        `json:"cover,omitempty"`
	URI            string        `json:"uri,omitempty"`
	Param          string        `json:"param,omitempty"`
	Goto           string        `json:"goto,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon    `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon    `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3 string        `json:"cover_left_text_3,omitempty"`
	Badge          *ChannelBadge `json:"badge,omitempty"`
	LeftText1      string        `json:"left_text_1,omitempty"`
	Position       int64         `json:"position,omitempty"`
}

func (c *ChannelRcmdV2) From(main interface{}, op *operate.Card) error {
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoChannel, op.Param, model.ChannelHandler("tab=select"))
	c.ID = op.ID
	c.IsButton = true
	c.IsAtten = op.Channel.IsAtten
	var upTime string
	if op.Channel.LastUpTime > 0 {
		upTime = model.PubDataString(time.Time(op.Channel.LastUpTime).Time())
	}
	if upTime != "" {
		c.Desc1 = fmt.Sprintf("%v更新", upTime)
	} else {
		c.Desc1 = "最近更新"
	}
	var buttonText string
	if op.Channel.AttenCnt > 0 {
		buttonText = model.StatString(op.Channel.AttenCnt, " 订阅")
		if FavTextReplace(op.MobiApp, int64(op.Build)) {
			buttonText = model.StatString(op.Channel.AttenCnt, " 收藏")
		}
	}
	c.DescButton = &Button{Text: buttonText}
	if op.Channel.FeatureCnt > 0 {
		buttonText := "精选视频" + model.StatString(op.Channel.FeatureCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=select"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	} else if op.Channel.TodayCnt > 0 {
		buttonText := "今日新投稿" + model.StatString(op.Channel.TodayCnt, "个 >")
		buttonURI := model.FillURI(model.GotoChannel, op.Plat, op.Build, op.Param, model.ChannelHandler("tab=all&sort=new"))
		c.DescButton2 = &Button{Text: buttonText, URI: buttonURI}
	}
	//nolint:gosimple
	var position int64
	position = op.Channel.Position
	for _, item := range op.Items {
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, item.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, item.ID, "AdAvIsNormalGRPC")
		}
		var (
			badge     *ChannelBadge
			letfText1 string
		)
		if op.Channel != nil {
			if op.Channel.Badges != nil {
				if bdg, ok := op.Channel.Badges[item.ID]; ok {
					if bdg != nil && bdg.Text != "" && bdg.Cover != "" {
						badge = &ChannelBadge{
							Text:    bdg.Text,
							BgCover: bdg.Cover,
						}
					}
				}
			}
			if fav, ok := op.Channel.IsFav[item.ID]; ok && fav {
				letfText1 = fmt.Sprintf("已收藏·%s", model.DurationString(a.Arc.Duration))
			}
			if letfText1 == "" {
				if coin, ok := op.Channel.Coins[item.ID]; ok && coin > 0 {
					letfText1 = fmt.Sprintf("已投币·%s", model.DurationString(a.Arc.Duration))
				}
			}
		}
		position++
		c.Items = append(c.Items, &ChannelRcmdV2Item{
			Title: a.Arc.Title,
			Cover: a.Arc.Pic,
			URI: model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10),
				model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, true)),
			Goto:           string(model.GotoAv),
			Param:          strconv.FormatInt(item.ID, 10),
			CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
			CoverLeftIcon1: model.IconPlay,
			CoverLeftText2: model.StatString(a.Arc.Stat.Danmaku, ""),
			CoverLeftIcon2: model.IconDanmaku,
			CoverLeftText3: model.DurationString(a.Arc.Duration),
			Badge:          badge,
			LeftText1:      letfText1,
			Position:       position,
		})
		c.Base.Idx = position
		if len(c.Items) == 2 {
			break
		}
	}
	if len(c.Items) == 0 {
		log.Error("rcmd channel_id(%v) from items(%+v) is 0", op.Param, op.Items)
		return newInvalidResourceErr(ResourceItems, 0, "empty `Items`")
	}
	c.Right = true
	return nil
}

func FavTextReplace(mobiApp string, build int64) bool {
	return (mobiApp == "android" && build >= 6470000) ||
		(mobiApp == "iphone" && build >= 64700000) ||
		(mobiApp == "ipad" && build >= 32900000)
}

func (c *ChannelRcmdV2) Get() *Base {
	return c.Base
}

type ChannelOGV struct {
	*Base
	MoreText  string            `json:"more_text,omitempty"`
	MoreURI   string            `json:"more_uri,omitempty"`
	HasFold   bool              `json:"has_fold,omitempty"`
	FoldOpen  string            `json:"fold_open,omitempty"`
	FoldClose string            `json:"fold_close,omitempty"`
	Items     []*ChannelOGVItem `json:"items,omitempty"`
}

type ChannelOGVItem struct {
	ID         int32     `json:"id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Cover      string    `json:"cover,omitempty"`
	URI        string    `json:"uri,omitempty"`
	Param      string    `json:"param,omitempty"`
	Goto       string    `json:"goto,omitempty"`
	LabelText1 string    `json:"label_text_1,omitempty"`
	LabelText2 string    `json:"label_text_2,omitempty"`
	LabelText3 string    `json:"label_text_3,omitempty"`
	Badge      *OGVBadge `json:"badge,omitempty"`
	Position   int64     `json:"position,omitempty"`
	// 上报字段
	Sort string `json:"sort"`
	Filt int32  `json:"filt"`
}

type OGVBadge struct {
	Text      string `json:"text,omitempty"`
	BadgeType int32  `json:"badge_type"`
}

func (c *ChannelOGV) From(main interface{}, op *operate.Card) error {
	c.Base.from(op.Plat, op.Build, op.Param, "", op.Title, "", strconv.FormatInt(op.ID, 10), nil)
	c.URI = ""
	season := main.(*appCardgrpc.SeasonCards)
	c.Title = fmt.Sprintf("相关剧集 %d", len(season.GetCards()))
	c.HasFold = op.Channel.HasFold
	c.FoldOpen = "展开更多"
	c.FoldClose = "收起"
	if op.Channel.HasMore && season.Link != "" {
		c.MoreText = season.LinkTitle
		c.MoreURI = season.Link
	}
	//nolint:gosimple
	var position int64
	position = op.Channel.Position
	for _, card := range season.GetCards() {
		position++
		i := &ChannelOGVItem{
			Position: position,
		}
		i.ID = card.GetSeasonId()
		i.Title = card.GetTitle()
		i.Cover = card.GetCover()
		i.URI = card.GetUri()
		i.LabelText1 = card.GetStyles()
		i.LabelText2 = card.GetActors()
		i.Sort = op.Channel.Sort
		i.Filt = op.Channel.Filt
		if card.GetBadge() != "" {
			i.Badge = &OGVBadge{
				Text:      card.GetBadge(),
				BadgeType: card.GetBadgeType(),
			}
		}
		if card.GetStats() != nil {
			i.LabelText3 = model.Stat64String(card.GetStats().GetView(), "播放")
		}
		c.Base.Idx = position
		c.Items = append(c.Items, i)
	}
	if len(c.Items) == 0 {
		log.Error("select OGV channel_id(%v) from items(%+v) is 0", op.Param, op.Items)
		return newInvalidResourceErr(ResourceItems, 0, "empty `Items`")
	}
	c.Right = true
	return nil
}

func (c *ChannelOGV) Get() *Base {
	return c.Base
}

type Storys struct {
	*Base
	Items []*StoryItems `json:"items,omitempty"`
}

type StoryItems struct {
	Base
	Avatar         *Avatar    `json:"avatar,omitempty"`
	FfCover        string     `json:"ff_cover,omitempty"`
	CoverLeftText1 string     `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon `json:"cover_left_icon_1,omitempty"`
	OfficialIcon   model.Icon `json:"official_icon,omitempty"`
	OfficialIconV2 model.Icon `json:"official_icon_v2,omitempty"`
	IsAtten        bool       `json:"is_atten,omitempty"`
}

func (c *Storys) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	if c.Rcmd == nil {
		return newResourceNotExistErr(ResourceAI, 0)
	}
	if c.Rcmd.StoryInfo == nil {
		return newInvalidResourceErr(ResourceAI, 0, "empty `StoryInfo`")
	}
	if len(c.Rcmd.StoryInfo.Items) == 0 {
		return newInvalidResourceErr(ResourceAI, 0, "empty `StoryInfo.Items`")
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		// 卡片
		c.Base.from(op.Plat, op.Build, "", "", c.Rcmd.StoryInfo.Title, "", "", nil)
		for _, v := range c.Rcmd.StoryInfo.Items {
			am := main.(map[int64]*arcgrpc.ArcPlayer)
			a, ok := am[v.ID]
			if !ok || !model.AvIsNormalGRPC(a) {
				continue
			}
			item := &StoryItems{
				FfCover: v.FfCover,
			}
			item.Base.from(op.Plat, op.Build, strconv.FormatInt(v.ID, 10), v.Cover, a.Arc.Title, model.GotoVerticalAv, strconv.FormatInt(v.ID, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), c.Rcmd.TrackID, &ai.Item{StoryParam: v.StoryParam}, op.Build, op.MobiApp, true))
			item.Base.CardGoto = model.CardGotoVerticalAv
			item.Args.fromArchiveGRPC(a.Arc, c.Tagm[v.Tid])
			item.PlayerArgs = playerArgsFrom(a)
			item.Avatar = avatarFrom(&AvatarStatus{Cover: a.Arc.Author.Face, Text: a.Arc.Author.Name, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound})
			item.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
			item.CoverLeftIcon1 = model.IconPlay
			item.OfficialIcon = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
			item.OfficialIconV2 = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
			if c.IsAttenm[a.Arc.Author.Mid] == 1 {
				item.OfficialIcon = model.IconIsAttenm
				item.IsAtten = true
			}
			// 小卡里面的三点不感兴趣
			item.ThreePointFrom(op.MobiApp, op.Build, 0, nil, 0, 0)
			// 老结构体的不感兴趣不下发
			item.ThreePoint = nil
			c.Items = append(c.Items, item)
		}
	default:
		log.Warn("Storys From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *Storys) Get() *Base {
	return c.Base
}
