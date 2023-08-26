package card

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"

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
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	"go-gateway/app/app-svr/app-card/interface/model/stat"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	season "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	tunnelcommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/pkg/errors"
)

func doubleHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, tagm map[int64]*taggrpc.Tag, isAttenm, hasLike map[int64]int8, statm map[int64]*relationgrpc.StatReply, cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply) (hander Handler) {
	base := &Base{CardType: cardType, CardGoto: cardGoto, Rcmd: rcmd, Tagm: tagm, IsAttenm: isAttenm, HasLike: hasLike, Statm: statm, Cardm: cardm, Columnm: model.ColumnSvrDouble, AuthorRelations: authorRelations}
	if rcmd != nil {
		base.fillRcmdMeta(rcmd)
	}
	switch cardType {
	case model.OnePicV3:
		base.CardLen = 1
		hander = &OnePicV3{Base: base}
	case model.ThreePicV3:
		base.CardLen = 1
		hander = &ThreePicV3{Base: base}
	case model.ThreePicV2:
		base.CardLen = 1
		hander = &ThreePicV2{Base: base}
	case model.SmallCoverV2:
		base.CardLen = 1
		hander = &SmallCoverV2{Base: base}
	case model.OptionsV2:
		hander = &Option{Base: base}
	case model.OnePicV2:
		base.CardLen = 1
		hander = &OnePicV2{Base: base}
	case model.Select:
		hander = &Select{Base: base}
	case model.BannerV5:
		hander = &Banner{Base: base}
	case model.BannerV5169:
		hander = &Banner{Base: base}
	case model.SmallCoverV9:
		base.CardLen = 1
		hander = &SmallCoverV9{Base: base}
	case model.LargeCoverV6:
		hander = &LargeCoverInline{Base: base}
	case model.SmallCoverV4:
		hander = &SmallCoverV4{Base: base}
	case model.LargeCoverV9:
		hander = &LargeCoverInline{Base: base}
	case model.BannerV8:
		hander = &BannerV8{Base: base}
	case model.OgvSmallCover:
		base.CardLen = 1
		hander = &OgvSmallCover{Base: base}
	case model.SmallCoverV10:
		base.CardLen = 1
		hander = &SmallCoverV10{Base: base}
	case model.SmallCoverV11:
		base.CardLen = 1
		hander = &SmallCoverV11{Base: base}
	default:
		switch cardGoto {
		case model.CardGotoAv, model.CardGotoLive, model.CardGotoArticleS, model.CardGotoSpecialS, model.CardGotoShoppingS, model.CardGotoAudio, model.CardGotoGameDownloadS,
			model.CardGotoBangumi, model.CardGotoMoe, model.CardGotoPGC, model.CardGotoAvConverge:
			base.CardType = model.SmallCoverV2
			base.CardLen = 1
			hander = &SmallCoverV2{Base: base}
		case model.CardGotoAdAv, model.CardGotoAdPgc:
			base.CardType = model.CmV2
			base.CardLen = 1
			hander = &SmallCoverV2{Base: base}
		case model.CardGotoAdPlayer, model.CardGotoAdInlineGesture, model.CardGotoAdInline360, model.CardGotoAdWebGif,
			model.CardGotoAdInlineChoose, model.CardGotoAdInlineChooseTeam, model.CardGotoAdWebGifReservation,
			model.CardGotoAdPlayerReservation, model.CardGotoAdInline3D, model.CardGotoAdInlineEggs, model.CardGotoAdInline3DV2:
			base.CardType = model.CmV2
			hander = &LargeCoverV2{Base: base}
		case model.CardGotoAdInlineLive:
			base.CardType = model.CmV2
			hander = &LargeCoverInline{Base: base}
		case model.CardGotoAdInlinePgc:
			base.CardType = model.CmDoubleV7
			hander = &LargeCoverInline{Base: base}
		case model.CardGotoChannelRcmd, model.CardGotoUpRcmdAv:
			base.CardType = model.SmallCoverV3
			base.CardLen = 1
			hander = &SmallCoverV3{Base: base}
		case model.CardGotoSpecial:
			base.CardType = model.MiddleCoverV2
			hander = &MiddleCover{Base: base}
		case model.CardGotoPlayer, model.CardGotoPlayerLive, model.CardGotoPlayerOGV, model.CardGotoPlayerBangumi:
			base.CardType = model.LargeCoverV2
			hander = &LargeCoverV2{Base: base}
		case model.CardGotoSubscribe, model.CardGotoSearchSubscribe:
			base.CardType = model.ThreeItemHV2
			hander = &ThreeItemH{Base: base}
		case model.CardGotoLiveUpRcmd:
			base.CardType = model.TwoItemV2
			return &TwoItemV2{Base: base}
		case model.CardGotoConverge, model.CardGotoRank, model.CardGotoConvergeAi:
			base.CardType = model.ThreeItemV2
			hander = &ThreeItemV2{Base: base}
		case model.CardGotoBangumiRcmd:
			base.CardType = model.SmallCoverV4
			hander = &SmallCoverV4{Base: base}
		case model.CardGotoBanner:
			base.CardType = model.BannerV2
			return &Banner{Base: base}
		case model.CardGotoAdWebS, model.CardGotoAdDynamic:
			base.CardType = model.CmV2
			base.CardLen = 1
			hander = &SmallCoverV2{Base: base}
		case model.CardGotoAdWeb:
			base.CardType = model.CmV2
			hander = &MiddleCover{Base: base}
		case model.CardGotoNews:
			base.CardType = model.News
			hander = &Text{Base: base}
		case model.CardGotoEntrance:
			base.CardType = model.MultiItemH
			hander = &MultiItem{Base: base}
		case model.CardGotoTagRcmd, model.CardGotoContentRcmd:
			base.CardType = model.MultiItem
			hander = &MultiItem{Base: base}
		case model.CardGotoVip:
			base.CardType = model.VipV1
			return &CoverOnly{Base: base}
		case model.CardGotoSpecialB:
			base.CardType = model.LargeCoverV3
			return &LargeCoverV3{Base: base}
		case model.CardGotoVipRenew, model.CardGotoTunnel:
			base.CardType = model.SmallCoverV7
			hander = &SmallCoverV7{Base: base}
		case model.CardGotoNewTunnel:
			base.CardType = model.NotifyTunnelV1
			hander = &UniversalNotifyTunnelV1{Base: base}
		case model.CardGotoBigTunnel:
			base.CardType = model.NotifyTunnelLargeV1
			hander = &UniversalNotifyTunnelLargeV1{Base: base}
		case model.CardGotoMultilayerConverge:
			base.CardType = model.SmallCoverConvergeV2
			base.CardLen = 1
			hander = &SmallCoverConvergeV2{Base: base}
		case model.CardGotoSpecialChannel:
			base.CardType = model.SmallCoverChannle
			base.CardLen = 1
			hander = &SmallChannelSpecial{Base: base}
		case model.CardGotoInlineAv:
			base.CardType = model.LargeCoverV5
			hander = &LargeCoverV5{Base: base}
		case model.CardGotoAiStory:
			base.CardType = model.StorysV2
			hander = &Storys{Base: base}
		case model.CardGotoInlinePGC:
			base.CardType = model.LargeCoverV7
			hander = &LargeCoverInline{Base: base}
		case model.CardGotoInlineLive:
			base.CardType = model.LargeCoverV8
			hander = &LargeCoverInline{Base: base}
		case model.CardGotoChannelNewDetail:
			base.CardLen = 1
			base.CardType = model.ChannelSmallCoverV1
			hander = &ChannelSmallCoverV1{Base: base}
		case model.CardGotoAdLive:
			base.CardType = model.CmV2
			base.CardLen = 1
			hander = &SmallCoverV9{Base: base}
		case model.CardGotoAdInlineAv:
			base.CardType = model.CmDoubleV9
			hander = &LargeCoverInline{Base: base}
		default:
			log.Error("Fail to build handler, rowType=%s cardType=%s cardGoto=%s ai={%+v}",
				stat.RowTypeDouble, string(cardType), string(cardGoto), rcmd)
		}
	}
	stat.MetricAppCardTotal.Inc(stat.RowTypeDouble, string(base.CardType), string(cardGoto))
	return
}

type SmallCoverV2 struct {
	*Base
	CoverGif                     string           `json:"cover_gif,omitempty"`
	CoverBlur                    model.BlurStatus `json:"cover_blur,omitempty"`
	CoverLeftText1               string           `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1               model.Icon       `json:"cover_left_icon_1,omitempty"`
	CoverLeft1ContentDescription string           `json:"cover_left_1_content_description,omitempty"`
	CoverLeftText2               string           `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2               model.Icon       `json:"cover_left_icon_2,omitempty"`
	CoverLeft2ContentDescription string           `json:"cover_left_2_content_description,omitempty"`
	CoverRightText               string           `json:"cover_right_text,omitempty"`
	CoverRightIcon               model.Icon       `json:"cover_right_icon,omitempty"`
	CoverRightContentDescription string           `json:"cover_right_content_description,omitempty"`
	CoverRightBackgroundColor    string           `json:"cover_right_background_color,omitempty"`
	Subtitle                     string           `json:"subtitle,omitempty"`
	Badge                        string           `json:"badge,omitempty"`
	BadgeStyle                   *ReasonStyle     `json:"badge_style,omitempty"`
	RcmdReason                   string           `json:"rcmd_reason,omitempty"`
	DescButton                   *Button          `json:"desc_button,omitempty"`
	Desc                         string           `json:"desc,omitempty"`
	Avatar                       *Avatar          `json:"avatar,omitempty"`
	OfficialIcon                 model.Icon       `json:"official_icon,omitempty"`
	CanPlay                      int32            `json:"can_play,omitempty"`
	RcmdReasonStyle              *ReasonStyle     `json:"rcmd_reason_style,omitempty"`
	RcmdReasonStyleV2            *ReasonStyle     `json:"rcmd_reason_style_v2,omitempty"`
	LeftCoverBadgeNewStyle       *ReasonStyle     `json:"left_cover_badge_new_style,omitempty"`
	// story详情页需要的封面图
	FfCover string `json:"ff_cover,omitempty"`
	// 跳转story图标
	GotoIcon   *model.GotoIcon   `json:"goto_icon,omitempty"`
	SharePlane *model.SharePlane `json:"share_plane,omitempty"`
	ServerInfo string            `json:"server_info,omitempty"` // 港澳台tab
}

//nolint:gocognit
func (c *SmallCoverV2) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		upID      int64
		button    interface{}
		avatar    *AvatarStatus
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
		switch op.CardGoto {
		case model.CardGotoAdAv:
			if !model.AdAvIsNormalGRPC(a) {
				return newInvalidResourceErr(ResourceArchive, op.ID, "AdAvIsNormalGRPC")
			}
			c.AdInfo = op.AdInfo
		default:
			if !model.AvIsNormalGRPC(a) {
				return newInvalidResourceErr(ResourceArchive, op.ID, "AvIsNormalGRPC")
			}
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, strconv.FormatInt(a.Arc.Aid, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		if isBangumi = a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && op.RedirectURL != ""; isBangumi {
			c.URI = model.FillURI("", 0, 0, op.RedirectURL, model.PGCTrackIDHandler(c.Rcmd))
		}
		// c.CoverLeftText1 = model.RecommendString(a.Stat.Like, a.Stat.DisLike)
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		if op.SwitchLike == model.SwitchFeedIndexLike {
			c.CoverLeftText1 = model.StatString(a.Arc.Stat.Like, "")
			c.CoverLeftIcon1 = model.IconLike
			c.CoverLeftText2 = model.StatString(a.Arc.Stat.View, "")
			c.CoverLeftIcon2 = model.IconPlay
		}
		c.CoverRightText = model.DurationString(a.Arc.Duration)
		c.CoverRightContentDescription = model.DurationContentDescription(a.Arc.Duration)
		var isShowV2 bool
		if c.Rcmd != nil {
			if c.CardGoto == model.CardGotoAv {
				cq := CastArchiveCustomizedQuality(c.Rcmd, a)
				cqCount := len(cq)
				switch cqCount {
				case 1:
					c.CoverLeftText1 = cq[0].Text
					c.CoverLeftIcon1 = cq[0].Icon
					c.CoverLeftText2 = ""
					c.CoverLeftIcon2 = 0
				case 2:
					c.CoverLeftText1 = cq[0].Text
					c.CoverLeftIcon1 = cq[0].Icon
					c.CoverLeftText2 = cq[1].Text
					c.CoverLeftIcon2 = cq[1].Icon
				}
				if c.Rcmd.HideDuration == 1 {
					c.CoverRightText = ""
					c.CoverRightContentDescription = ""
				}
			}
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorname = au.Name
				}
			}
			c.RcmdReason, c.Desc = rcmdReason(c.Rcmd.RcmdReason, authorname, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoAvConverge:
				if isShowV2 = IsShowRcmdReasonStyleV2(c.Rcmd); isShowV2 {
					c.Desc = ""
				}
			}
			if isShowV2 {
				if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV4(c.Rcmd, c.RcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
					c.RcmdReasonStyleV2 = reasonStyleFromV3(c.Rcmd, c.RcmdReason, c.Base.Goto, op.Plat, op.Build)
				} else {
					c.RcmdReasonStyleV2 = reasonStyleFromV2(c.Rcmd, c.RcmdReason, c.Base.Goto, op.Plat, op.Build)
				}
				c.RcmdReason = ""
			} else {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
			}
		}
		if !isShowV2 {
			if c.RcmdReason == "" {
				if op.Channel != nil && op.Channel.ChannelID != 0 && op.Channel.ChannelName != "" {
					op.Channel.ChannelName = a.Arc.TypeName + " · " + op.Channel.ChannelName
					button = op.Channel
				} else if t, ok := c.Tagm[op.Tid]; ok {
					tag := &taggrpc.Tag{}
					*tag = *t
					tag.Name = a.Arc.TypeName + " · " + tag.Name
					button = tag
				} else {
					button = &ButtonStatus{Text: a.Arc.TypeName}
				}
				// customized by recommendation service
				if c.Rcmd != nil && c.Rcmd.CustomizedDesc != nil && c.Tagm[op.Tid] != nil && c.CardGoto == model.CardGotoAv {
					customizedBtn, ok := CastArchiveCustomizedDesc(c.Rcmd, a, c.Tagm[op.Tid])
					if ok {
						button = customizedBtn
					}
				}
			}
		}
		if !isBangumi {
			c.Base.PlayerArgs = playerArgsFrom(a)
			c.CanPlay = a.Arc.Rights.Autoplay
		}
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		upID = a.Arc.Author.Mid
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoAdAv:
			c.AdInfo = op.AdInfo
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
					return newInvalidResourceErr(ResourceArchive, caid, "AdAvIsNormalGRPC")
				}
				urlf = model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(ac, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
				c.Base.PlayerArgs = playerArgsFrom(ac)
			case model.GotoAvConverge:
				c.Base.PlayerArgs = nil
				c.CanPlay = 0
				if c.Rcmd.ID != 0 {
					c.Param = strconv.FormatInt(c.Rcmd.ID, 10)
				}
				urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, 0, 0)
			default:
				urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, 0, 0)
			}
			c.Goto = jumpGoto
			c.Base.URI = model.FillURI(jumpGoto, op.Plat, op.Build, op.Param, urlf)
			c.Args.fromAvConverge(c.Base.Rcmd)
		}
		if c.Rcmd != nil {
			//nolint:exhaustive
			switch model.Gt(c.Rcmd.JumpGoto) {
			case model.GotoVerticalAv:
				c.FfCover = a.Arc.FirstFrame
				c.Goto = model.GotoVerticalAv
				c.URI = model.FillURI(model.GotoVerticalAv, op.Plat, op.Build, op.Param, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, c.Rcmd, op.Build, op.MobiApp, true))
				c.GotoIcon = model.FillGotoIcon(c.Rcmd.IconType, op.GotoIcon)
			}
			if c.Rcmd.IconType == model.AIUpIconType && c.RcmdReason == "" && c.CardGoto == model.CardGotoAv {
				button = &CustomizedButtonMeta{
					Text: a.Arc.Author.Name,
					URI:  model.FillURI(model.GotoMid, 0, 0, strconv.FormatInt(a.Arc.Author.Mid, 10), nil),
				}
				if c.Rcmd.CustomizedDesc != nil && c.Tagm[op.Tid] != nil {
					customizedBtn, ok := CastArchiveCustomizedDesc(c.Rcmd, a, c.Tagm[op.Tid])
					if ok {
						button = customizedBtn
					}
				}
				c.GotoIcon = model.FillGotoIcon(c.Rcmd.IconType, op.GotoIcon)
			}
		}
		c.CoverGif = op.GifCover
	case map[int64]*bangumi.Season:
		sm := main.(map[int64]*bangumi.Season)
		s, ok := sm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceSeason, op.ID)
		}
		c.Base.from(op.Plat, op.Build, s.EpisodeID, s.Cover, s.Title, model.GotoBangumi, s.EpisodeID, nil)
		c.CoverLeftText1 = model.StatString(s.PlayCount, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(s.Favorites, "")
		c.CoverLeftIcon2 = model.IconFavorite
		c.Badge = s.TypeBadge
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, s.TypeBadge)
		c.Subtitle = s.UpdateDesc
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		} else {
			if t, ok := c.Tagm[op.Tid]; ok && op.Desc != "" && t.Name != "" {
				tag := &taggrpc.Tag{}
				*tag = *t
				tag.Name = op.Desc + " · " + tag.Name
				button = tag
			}
		}
	case map[int32]*season.CardInfoProto:
		sm := main.(map[int32]*season.CardInfoProto)
		s, ok := sm[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceSeason, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, s.Cover, s.Title, model.GotoPGC, op.URI, nil)
		c.CoverLeftText1 = model.StatString(int32(s.Stat.View), "")
		c.CoverLeftIcon1 = model.IconPlay
		if s.Stat != nil {
			c.CoverLeftText2 = model.StatString(int32(s.Stat.Follow), "")
		}
		c.CoverLeftIcon2 = model.IconFavorite
		c.Badge = s.SeasonTypeName
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, s.SeasonTypeName)
		if s.NewEp != nil {
			c.Subtitle = s.NewEp.IndexShow
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
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			} else {
				if t, ok := c.Tagm[op.Tid]; ok && op.Desc != "" && t.Name != "" {
					tag := &taggrpc.Tag{}
					*tag = *t
					tag.Name = op.Desc + " · " + tag.Name
					button = tag
				}
			}
		default:
			c.Base.from(op.Plat, op.Build, op.Param, s.Cover, title, model.GotoBangumi, op.URI, nil)
			c.Goto = model.GotoPGC
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			}
		}
		if s.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText2 = model.StatString(int32(s.Season.Stat.Follow), "")
			c.CoverLeftIcon2 = model.IconFavorite
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
		}
		if s.Url != "" {
			c.URI = s.Url // 直接用pgc接口返回的URL
		}
		c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		c.Badge = s.Season.SeasonTypeName
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, s.Season.SeasonTypeName)
		if _, ok := op.SwitchStyle[model.SwitchPGCHideSubtitle]; !ok && s.Season != nil {
			c.Subtitle = s.Season.NewEpShow
		} else if button == nil && c.RcmdReasonStyle == nil && s.Season != nil {
			button = &ButtonStatus{Text: s.Season.NewEpShow}
		}
		if c.Rcmd != nil {
			c.OgvCreativeId = c.Rcmd.CreativeId
		}
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if r.LiveStatus != 1 {
			return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
		}
		c.Base.from(op.Plat, op.Build, op.Param, r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.CoverLeftText1 = model.StatString(r.Online, "")
		c.CoverLeftIcon1 = model.IconOnline
		c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
		c.CoverRightText = r.Uname
		c.CoverRightContentDescription = r.Uname
		c.Badge = "直播"
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "直播")
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		if c.Rcmd != nil && (c.Rcmd.RcmdReason != nil || c.IsAttenm[r.UID] == 1) {
			c.RcmdReason, c.Desc = rcmdReason(c.Rcmd.RcmdReason, "", c.IsAttenm[r.UID], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else {
			button = r
		}
		upID = r.UID
		c.CanPlay = 1
	case map[int64]*show.Shopping:
		sm := main.(map[int64]*show.Shopping)
		s, ok := sm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceShopping, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, model.ShoppingCover(s.PerformanceImage), s.Name, model.GotoWeb, s.URL, nil)
		if s.Type == 1 {
			c.CoverLeftText1 = model.ShoppingDuration(s.STime, s.ETime)
			c.CoverRightText = s.CityName
			c.CoverRightIcon = model.IconLocation
			if len(s.Tags) != 0 {
				c.Desc = s.Tags[0].TagName
			}
		} else if s.Type == 2 {
			c.CoverLeftText1 = s.Want
			c.Desc = s.Subname
		}
		c.Badge = "会员购"
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "会员购")
		c.Args.fromShopping(s)
	case map[int64]*audio.Audio:
		am := main.(map[int64]*audio.Audio)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceAudio, op.ID)
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.CoverURL, a.Title, model.GotoAudio, strconv.FormatInt(a.MenuID, 10), nil)
		c.CoverBlur = model.BlurYes
		c.CoverLeftText1 = model.StatString(a.PlayNum, "")
		c.CoverLeftIcon1 = model.IconHeadphone
		c.CoverRightText = model.AudioTotalStirng(a.RecordNum)
		c.Badge = model.AudioBadgeString(a.Type)
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, model.AudioBadgeString(a.Type))
		button = a.Ctgs
		c.Args.fromAudio(a)
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
		if m.Stats != nil {
			c.CoverLeftText1 = model.StatString(int32(m.Stats.View), "")
			c.CoverLeftIcon1 = model.IconRead
			c.CoverLeftText2 = model.StatString(int32(m.Stats.Reply), "")
			c.CoverLeftIcon2 = model.IconComment
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
		}
		c.Badge = "文章"
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "文章")
		c.Args.fromArticle(m)
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		} else {
			button = m.Categories
		}
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
			return newInvalidResourceErr(ResourceRoom, p.DynamicID, "ViewCount: %d", p.ViewCount)
		}
		c.Base.from(op.Plat, op.Build, op.Param, p.Imgs[0], p.DynamicText, model.GotoPicture, strconv.FormatInt(p.DynamicID, 10), nil)
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.StatString(p.LikeCount, "")
			c.CoverLeftIcon1 = model.IconLike
		} else {
			c.CoverLeftText1 = model.StatString(int32(p.ViewCount), "")
			c.CoverLeftIcon1 = model.IconRead
		}
		c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
		if p.ImgCount > 1 {
			c.CoverRightText = model.PictureCountString(p.ImgCount)
			c.CoverRightContentDescription = model.PictureCountString(p.ImgCount)
			c.CoverRightBackgroundColor = "#66666666"
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReason, _ = rcmdReason(c.Rcmd.RcmdReason, p.NickName, c.IsAttenm[p.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else {
			button = p
			c.Badge = "动态"
			c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "动态")
		}
		c.Args.fromPicture(p)
	case *cm.AdInfo:
		ad := main.(*cm.AdInfo)
		c.AdInfo = ad
	case *bangumi.Moe:
		m := main.(*bangumi.Moe)
		if m == nil {
			return newResourceNotExistErr(ResourceMoe, 0)
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(m.ID, 10), m.Square, m.Title, model.GotoWeb, m.Link, nil)
		c.Desc = m.Desc
		c.Badge = m.Badge
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, m.Badge)
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		intfc, ok := intfcm[op.Goto]
		if !ok {
			return newUnexpectedGotoErr(string(op.Goto), "")
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		c.CoverGif = op.GifCover
		c.Badge = op.Badge
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
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
			c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
			c.CoverLeftIcon2 = model.IconDanmaku
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
			c.CoverRightText = model.DurationString(a.Arc.Duration)
			c.CoverRightContentDescription = model.DurationContentDescription(a.Arc.Duration)
			if isBangumi = a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.Arc.RedirectURL != ""; isBangumi {
				c.URI = model.FillURI("", 0, 0, a.Arc.RedirectURL, model.PGCTrackIDHandler(c.Rcmd))
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
				c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(s.Season.Stat.Follow), "")
				c.CoverLeftIcon2 = model.IconFavorite
				c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
				c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int32]*pgcAppGrpc.SeasonCardInfoProto:
			sm := intfc.(map[int32]*pgcAppGrpc.SeasonCardInfoProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceSeason, op.ID)
			}
			if s.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(s.Stat.View), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(s.Stat.Follow), "")
				c.CoverLeftIcon2 = model.IconFavorite
				c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
				c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
			}
			if s.Url != "" {
				c.URI = s.Url // 直接用pgc接口返回的URL
			}
			c.URI = model.FillURI("", 0, 0, c.URI, model.PGCTrackIDHandler(c.Rcmd))
		case map[int32]*pgccard.EpisodeCard:
			sm := intfc.(map[int32]*pgccard.EpisodeCard)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(s.Stat.Play), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(s.Stat.Follow), "")
				c.CoverLeftIcon2 = model.IconFavorite
				c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
				c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
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
			c.CoverLeftText1 = model.StatString(r.Online, "")
			c.CoverLeftIcon1 = model.IconOnline
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverRightText = r.Uname
			c.CoverRightContentDescription = r.Uname
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
				c.CoverLeftText1 = model.StatString(int32(m.Stats.View), "")
				c.CoverLeftIcon1 = model.IconRead
				c.CoverLeftText2 = model.StatString(int32(m.Stats.Reply), "")
				c.CoverLeftIcon2 = model.IconComment
				c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
				c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
			}
		default:
			log.Warn("SmallCoverV1 From: unexpected type %T", intfc)
			return newUnexpectedResourceTypeErr(intfc, "")
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
		} else {
			c.Desc = op.Desc
		}
		c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	case nil:
		if op == nil {
			return newEmptyOPErr()
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
		switch op.CardGoto {
		case model.CardGotoDownload:
			c.CoverLeftText1 = model.DownloadString(op.Download)
			avatar = &AvatarStatus{Cover: op.Avatar, Goto: op.Goto, Param: op.URI, Type: model.AvatarSquare}
			c.Desc = op.Desc
		case model.CardGotoSpecial:
			c.CoverGif = op.GifCover
			c.Badge = op.Badge
			c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
			if _, ok := op.SwitchStyle[model.SwitchSpecialInfo]; ok && c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			} else {
				c.Desc = op.Desc
			}
			c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
		default:
			log.Warn("SmallCoverV2 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("SmallCoverV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	c.Avatar = avatarFrom(avatar)
	// 物料描述不为空，且推荐理由为空
	if op.Desc != "" {
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoSpecialS, model.CardGotoSpecial:
			if c.RcmdReasonStyle == nil && c.RcmdReasonStyleV2 == nil {
				c.Desc = op.Desc
			}
		case model.CardGotoAv, model.CardGotoLive:
			if c.RcmdReasonStyle == nil && c.RcmdReasonStyleV2 == nil {
				button = &ButtonStatus{Text: op.Desc}
			} else {
				c.Desc = op.Desc
			}
			if c.Rcmd != nil && c.Rcmd.IconType == model.AIUpIconType {
				c.GotoIcon = nil
			}
		case model.CardGotoPGC, model.CardGotoAudio, model.CardGotoArticleS, model.CardGotoArticle:
			if c.RcmdReasonStyle == nil && c.RcmdReasonStyleV2 == nil {
				button = &ButtonStatus{Text: op.Desc}
			}
		}
	}
	c.DescButton = buttonFrom(button, op.Plat)
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
	c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	c.Right = true
	return nil
}

func (c *SmallCoverV2) Get() *Base {
	return c.Base
}

type SmallCoverV3 struct {
	*Base
	Avatar           *Avatar      `json:"avatar,omitempty"`
	CoverLeftText    string       `json:"cover_left_text,omitempty"`
	CoverRightButton *Button      `json:"cover_right_button,omitempty"`
	RcmdReason       string       `json:"rcmd_reason,omitempty"`
	Desc             string       `json:"desc,omitempty"`
	DescButton       *Button      `json:"desc_button,omitempty"`
	OfficialIcon     model.Icon   `json:"official_icon,omitempty"`
	CanPlay          int32        `json:"can_play,omitempty"`
	RcmdReasonStyle  *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *SmallCoverV3) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button     interface{}
		descButton interface{}
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
		switch op.CardGoto {
		case model.CardGotoUpRcmdAv:
			authorface := a.Arc.Author.Face
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			c.Avatar = avatarFrom(&AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound})
			c.CoverLeftText = authorname
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
				c.RcmdReason, _ = rcmdReason(c.Rcmd.RcmdReason, "", c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
			} else if op.Channel != nil && op.Channel.ChannelID != 0 && op.Channel.ChannelName != "" {
				op.Channel.ChannelName = a.Arc.TypeName + " · " + op.Channel.ChannelName
				//nolint:ineffassign
				button = op.Channel
			} else {
				descButton = c.Tagm[op.Tid]
			}
			button = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
			c.Base.PlayerArgs = playerArgsFrom(a)
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		case model.CardGotoChannelRcmd:
			t, ok := c.Tagm[op.Tid]
			if !ok {
				return newResourceNotExistErr(ResourceTag, op.Tid)
			}
			c.Avatar = avatarFrom(&AvatarStatus{Cover: t.Cover, Goto: model.GotoTag, Param: strconv.FormatInt(t.Id, 10), Type: model.AvatarSquare})
			c.CoverLeftText = t.Name
			c.Desc = model.SubscribeString(int32(t.Sub))
			button = &ButtonStatus{Goto: model.GotoTag, Param: strconv.FormatInt(t.Id, 10), IsAtten: int8(t.Attention)}
			c.Base.PlayerArgs = playerArgsFrom(a)
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		default:
			log.Warn("SmallCoverV3 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
		c.CanPlay = a.Arc.Rights.Autoplay
	default:
		log.Warn("SmallCoverV3 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.CoverRightButton = buttonFrom(button, op.Plat)
	c.DescButton = buttonFrom(descButton, op.Plat)
	c.Right = true
	return nil
}

func (c *SmallCoverV3) Get() *Base {
	return c.Base
}

type MiddleCoverV2 struct {
	*Base
	Ratio int    `json:"ratio,omitempty"`
	Desc  string `json:"desc,omitempty"`
	Badge string `json:"badge,omitempty"`
}

func (c *MiddleCoverV2) Get() *Base {
	return c.Base
}

type LargeCoverV2 struct {
	*Base
	Avatar           *Avatar      `json:"avatar,omitempty"`
	Badge            string       `json:"badge,omitempty"`
	CoverRightButton *Button      `json:"cover_right_button,omitempty"`
	CoverLeftText1   string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1   model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2   string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2   model.Icon   `json:"cover_left_icon_2,omitempty"`
	RcmdReason       string       `json:"rcmd_reason,omitempty"`
	DescButton       *Button      `json:"desc_button,omitempty"`
	OfficialIcon     model.Icon   `json:"official_icon,omitempty"`
	CanPlay          int32        `json:"can_play,omitempty"`
	RcmdReasonStyle  *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	ShowTop          int8         `json:"show_top,omitempty"`
	ShowBottom       int8         `json:"show_bottom,omitempty"`
	BadgeStyle       *ReasonStyle `json:"badge_style,omitempty"`
}

//nolint:gocognit
func (c *LargeCoverV2) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button      interface{}
		coverButton interface{}
		upID        int64
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
		authorface := a.Arc.Author.Face
		authorname := a.Arc.Author.Name
		if c.Cardm != nil {
			if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
				authorface = au.Face
				authorname = au.Name
			}
		}
		c.Avatar = avatarFrom(&AvatarStatus{Cover: authorface, Text: authorname, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound})
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		if op.SwitchLike == model.SwitchFeedIndexLike {
			c.CoverLeftText1 = model.StatString(a.Arc.Stat.Like, "")
			c.CoverLeftIcon1 = model.IconLike
			c.CoverLeftText2 = model.StatString(a.Arc.Stat.View, "")
			c.CoverLeftIcon2 = model.IconPlay
		}
		coverButton = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReason, _ = rcmdReason(c.Rcmd.RcmdReason, "", c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else if t, ok := c.Tagm[op.Tid]; ok {
			tag := &taggrpc.Tag{}
			*tag = *t
			tag.Name = a.Arc.TypeName + " · " + tag.Name
			button = tag
		} else {
			button = &ButtonStatus{Text: a.Arc.TypeName}
		}
		c.CanPlay = a.Arc.Rights.Autoplay
		c.Base.PlayerArgs = playerArgsFrom(a)
		if op.CardGoto == model.CardGotoPlayer && c.Base.PlayerArgs == nil {
			log.Warn("player card aid(%d) can't auto player", a.Arc.Aid)
			return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "PlayerArgs")
		}
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		upID = a.Arc.Author.Mid
		c.DescButton = buttonFrom(button, op.Plat)
		c.CoverRightButton = buttonFrom(coverButton, op.Plat)
		c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if r.LiveStatus != 1 {
			return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
		}
		c.Base.from(op.Plat, op.Build, op.Param, r.Cover, r.Title, model.GotoLive, op.URI, model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.Avatar = avatarFrom(&AvatarStatus{Cover: r.Face, Text: r.Uname, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound})
		c.CoverLeftText1 = model.StatString(r.Online, "")
		c.CoverLeftIcon1 = model.IconOnline
		coverButton = &ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), IsAtten: c.IsAttenm[r.UID]}
		if c.Rcmd != nil && (c.Rcmd.RcmdReason != nil || c.IsAttenm[r.UID] == 1) {
			c.RcmdReason, _ = rcmdReason(c.Rcmd.RcmdReason, r.Uname, c.IsAttenm[r.UID], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else {
			button = r
		}
		c.Badge = "直播"
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "直播")
		c.CanPlay = 1
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		upID = r.UID
		c.DescButton = buttonFrom(button, op.Plat)
		c.CoverRightButton = buttonFrom(coverButton, op.Plat)
		c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	case map[int64]*bangumi.EpPlayer:
		eps := main.(map[int64]*bangumi.EpPlayer)
		ep, ok := eps[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		title := ep.NewDesc
		if ep.Season != nil {
			title = fmt.Sprintf("%s %s", ep.Season.Title, title)
		}
		c.Base.from(op.Plat, op.Build, op.Param, ep.Cover, title, model.GotoBangumi, "", nil)
		c.URI = model.FillURI("", 0, 0, ep.Uri, model.PGCTrackIDHandler(c.Rcmd))
		if ep.PlayerInfo != nil && ((op.MobiApp == "iphone" && op.Build > 8470) || (op.MobiApp == "android" && op.Build > 5405000)) {
			c.CanPlay = 1
		}
		c.Base.PlayerArgs = playerArgsFrom(ep)
		//nolint:exhaustive
		switch c.Base.CardGoto {
		case model.CardGotoPlayerBangumi: //天马首页的PGC播放卡
			if ep.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(ep.Stat.Play), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(ep.Stat.Danmaku), "")
				c.CoverLeftIcon2 = model.IconDanmaku
			}
			button = ep
			c.DescButton = buttonFrom(button, op.Plat)
		case model.CardGotoPlayerOGV:
			if (op.MobiApp == "iphone" && op.Build <= 8470) || (op.MobiApp == "android" && op.Build <= 5405000) {
				return newInsufficientClientVersionErr("CardGotoPlayerOGV")
			}
		}
	case map[int32]*pgcinline.EpisodeCard:
		eps := main.(map[int32]*pgcinline.EpisodeCard)
		ep, ok := eps[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		title := ep.NewDesc
		if ep.Season != nil {
			title = fmt.Sprintf("%s %s", ep.Season.Title, title)
		}
		c.Base.from(op.Plat, op.Build, op.Param, ep.Cover, title, model.GotoBangumi, "", nil)
		c.URI = model.FillURI("", 0, 0, ep.Url, model.PGCTrackIDHandler(c.Rcmd))
		if ep.PlayerInfo != nil && ((op.MobiApp == "iphone" && op.Build > 8470) || (op.MobiApp == "android" && op.Build > 5405000)) {
			c.CanPlay = 1
		}
		c.Base.PlayerArgs = playerArgsFrom(ep)
		//nolint:exhaustive
		switch c.Base.CardGoto {
		case model.CardGotoPlayerBangumi: //天马首页的PGC播放卡
			if ep.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(ep.Stat.Play), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(ep.Stat.Danmaku), "")
				c.CoverLeftIcon2 = model.IconDanmaku
			}
			button = ep
			c.DescButton = buttonFrom(button, op.Plat)
		case model.CardGotoPlayerOGV:
			if (op.MobiApp == "iphone" && op.Build <= 8470) || (op.MobiApp == "android" && op.Build <= 5405000) {
				return newInsufficientClientVersionErr("CardGotoPlayerOGV")
			}
		}
	case *cm.AdInfo:
		ad := main.(*cm.AdInfo)
		c.AdInfo = ad
	default:
		log.Warn("MiddleCoverV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	//nolint:exhaustive
	switch op.SwitchLargeCoverShow {
	case model.SwitchLargeCoverShowAll:
		c.ShowTop = 1
		c.ShowBottom = 1
	case model.SwitchLargeCoverShowBottom:
		c.ShowBottom = 1
	}
	c.Right = true
	return nil
}

func (c *LargeCoverV2) Get() *Base {
	return c.Base
}

type ThreeItemV2 struct {
	*Base
	TitleIcon model.Icon         `json:"title_icon,omitempty"`
	MoreURI   string             `json:"more_uri,omitempty"`
	MoreText  string             `json:"more_text,omitempty"`
	Items     []*ThreeItemV2Item `json:"items,omitempty"`
}

type ThreeItemV2Item struct {
	Base
	CoverLeftIcon model.Icon   `json:"cover_left_icon,omitempty"`
	DescText1     string       `json:"desc_text_1,omitempty"`
	DescIcon1     model.Icon   `json:"desc_icon_1,omitempty"`
	DescText2     string       `json:"desc_text_2,omitempty"`
	DescIcon2     model.Icon   `json:"desc_icon_2,omitempty"`
	Badge         string       `json:"badge,omitempty"`
	BadgeStyle    *ReasonStyle `json:"badge_style,omitempty"`
}

//nolint:gocognit
func (c *ThreeItemV2) From(main interface{}, op *operate.Card) error {
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
			c.TitleIcon = model.IconRank
			c.MoreURI = model.FillURI(op.Goto, 0, 0, op.URI, nil)
			c.MoreText = "查看更多"
			c.Items = make([]*ThreeItemV2Item, 0, _limit)
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				var item *ThreeItemV2Item
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item = &ThreeItemV2Item{
						DescText1: model.ScoreString(v.Score),
					}
					item.Base.from(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
					item.Args.fromArchiveGRPC(a.Arc, nil)
				default:
					log.Warn("ThreeItemV2 From: unexpected type %T", intfc)
					continue
				}
				c.Items = append(c.Items, item)
				if len(c.Items) == _limit {
					break
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
			c.Items = make([]*ThreeItemV2Item, 0, len(op.Items))
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				var item *ThreeItemV2Item
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item = &ThreeItemV2Item{
						DescText1: model.StatString(a.Arc.Stat.View, ""),
						DescIcon1: model.IconPlay,
						DescText2: model.StatString(a.Arc.Stat.Danmaku, ""),
						DescIcon2: model.IconDanmaku,
					}
					if op.SwitchLike == model.SwitchFeedIndexLike {
						item.DescText1 = model.StatString(a.Arc.Stat.Like, "")
						item.DescIcon1 = model.IconLike
						item.DescText2 = model.StatString(a.Arc.Stat.View, "")
						item.DescIcon2 = model.IconPlay
					}
					item.Base.from(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
					item.Args.fromArchiveGRPC(a.Arc, nil)
				case map[int64]*live.Room:
					rm := intfc.(map[int64]*live.Room)
					r, ok := rm[v.ID]
					if !ok || r.LiveStatus != 1 {
						continue
					}
					item = &ThreeItemV2Item{
						DescText1:  model.StatString(r.Online, ""),
						DescIcon1:  model.IconOnline,
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
					item = &ThreeItemV2Item{
						Badge:      "文章",
						BadgeStyle: reasonStyleFrom(model.BgColorTransparentRed, "文章"),
					}
					item.Base.from(op.Plat, op.Build, v.Param, m.ImageURLs[0], m.Title, model.GotoArticle, v.URI, nil)
					if m.Stats != nil {
						item.DescText1 = model.StatString(int32(m.Stats.View), "")
						item.DescIcon1 = model.IconRead
						item.DescText2 = model.StatString(int32(m.Stats.Reply), "")
						item.DescIcon2 = model.IconComment
					}
					item.Args.fromArticle(m)
				default:
					log.Warn("ThreeItemV2 From: unexpected type %T", intfc)
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
			log.Warn("ThreeItemV2 From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	default:
		log.Warn("ThreeItemV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *ThreeItemV2) Get() *Base {
	return c.Base
}

type SmallCoverV4 struct {
	*Base
	CoverBadge     string     `json:"cover_badge,omitempty"`
	Desc           string     `json:"desc,omitempty"`
	TitleRightText string     `json:"title_right_text,omitempty"`
	TitleRightPic  model.Icon `json:"title_right_pic,omitempty"`
	SeasonId       []int64    `json:"season_id,omitempty"`
	Epid           []int64    `json:"epid,omitempty"`
}

func (c *SmallCoverV4) From(main interface{}, op *operate.Card) error {
	emojim := map[uint32]string{
		0: "(´∀｀*)ｳﾌﾌ",
		1: "ヾ( ・∀・)ﾉ",
		2: "(｀･ω･´)ゞ",
		3: "(・∀・)ｲｲ!!",
	}
	//nolint:gosimple
	switch main.(type) {
	case *bangumi.Update:
		title := "你的追番更新啦"
		const (
			_updates = 99
		)
		u := main.(*bangumi.Update)
		if u == nil {
			return newResourceNotExistErr(ResourceBangumiUpdate, 0)
		}
		if u.Updates == 0 {
			return newInvalidResourceErr(ResourceBangumiUpdate, 0, "empty `Updates`")
		}
		title = title + emojim[crc32.ChecksumIEEE([]byte(c.Rcmd.TrackID))%4]
		c.Base.from(op.Plat, op.Build, "", u.SquareCover, title, "", "", nil)
		updates := u.Updates
		if updates > _updates {
			updates = _updates
			c.TitleRightPic = model.IconBomb
		} else {
			c.TitleRightPic = model.IconTV
		}
		c.Desc = u.Title
		c.TitleRightText = strconv.Itoa(updates)
	case *bangumi.Remind:
		const (
			_updates = 99
		)
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
		var title = u.List[0].UpdateTitle
		title = title + emojim[crc32.ChecksumIEEE([]byte(c.Rcmd.TrackID))%4]
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
		c.Base.from(op.Plat, op.Build, "", cover, title, "", uriStr, nil)
		updates := u.Updates
		if updates > _updates {
			updates = _updates
			c.TitleRightPic = model.IconBomb
		} else {
			c.TitleRightPic = model.IconTV
		}
		c.Desc = u.List[0].UpdateDesc
		c.TitleRightText = strconv.Itoa(updates)
		for _, v := range u.List {
			c.SeasonId = append(c.SeasonId, v.SeasonId)
			c.Epid = append(c.Epid, v.Epid)
		}
	case *tunnelgrpc.FeedCard:
		tunnelCard := main.(*tunnelgrpc.FeedCard)
		if tunnelCard == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		c.Base.from(op.Plat, op.Build, op.Param, tunnelCard.Cover, tunnelCard.Title, model.GotoWeb, tunnelCard.Link, nil)
		c.Desc = tunnelCard.Intro
	default:
		log.Warn("SmallCoverV4 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverV4) Get() *Base {
	return c.Base
}

type TwoItemV2 struct {
	*Base
	Items []*TwoItemV2Item `json:"items,omitempty"`
}

type TwoItemV2Item struct {
	Base
	Badge          string       `json:"badge,omitempty"`
	BadgeStyle     *ReasonStyle `json:"badge_style,omitempty"`
	CoverLeftText1 string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon   `json:"cover_left_icon_1,omitempty"`
	DescButton     *Button      `json:"desc_button,omitempty"`
}

func (c *TwoItemV2) From(main interface{}, op *operate.Card) error {
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
		c.Items = make([]*TwoItemV2Item, 0, _limit)
		for _, card := range cs {
			if card == nil || card.LiveStatus != 1 {
				continue
			}
			item := &TwoItemV2Item{
				Badge:          "直播",
				BadgeStyle:     reasonStyleFrom(model.BgColorTransparentRed, "直播"),
				CoverLeftText1: model.StatString(card.Online, ""),
				CoverLeftIcon1: model.IconOnline,
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

func (c *TwoItemV2) Get() *Base {
	return c.Base
}

type MultiItem struct {
	*Base
	MoreURI  string    `json:"more_uri,omitempty"`
	MoreText string    `json:"more_text,omitempty"`
	Items    []Handler `json:"items,omitempty"`
}

//nolint:gocognit
func (c *MultiItem) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch main.(type) {
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		switch op.CardGoto {
		case model.CardGotoTagRcmd, model.CardGotoContentRcmd:
			items := make([]Handler, 0, len(op.Items))
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				var hander Handler
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item := &SmallCoverV2{
						CoverLeftText1: model.StatString(a.Arc.Stat.View, ""),
						CoverLeftIcon1: model.IconPlay,
						CoverLeftText2: model.StatString(a.Arc.Stat.Danmaku, ""),
						CoverLeftIcon2: model.IconDanmaku,
						CoverRightText: model.DurationString(a.Arc.Duration),
						Base:           &Base{CardType: model.SmallCoverV2},
					}
					if op.SwitchLike == model.SwitchFeedIndexLike {
						item.CoverLeftText1 = model.StatString(a.Arc.Stat.Like, "")
						item.CoverLeftIcon1 = model.IconLike
						item.CoverLeftText2 = model.StatString(a.Arc.Stat.View, "")
						item.CoverLeftIcon2 = model.IconPlay
					}
					item.Base.from(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, strconv.FormatInt(a.Arc.Aid, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
					if a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.Arc.RedirectURL != "" {
						// 如果RedirectURL不为空用RedirectURL的数据并且带上TrackID信息
						item.URI = model.FillURI("", 0, 0, a.Arc.RedirectURL, model.URLTrackIDHandler(&ai.Item{TrackID: op.TrackID}))
					}
					item.Args.fromArchiveGRPC(a.Arc, nil)
					if op.Switch == model.SwitchFeedIndexTabThreePoint {
						item.TabThreePointWatchLater()
					}
					item.DescButton = buttonFrom(&ButtonStatus{Text: a.Arc.TypeName}, op.Plat)
					hander = item
				case map[int64]*live.Room:
					rm := intfc.(map[int64]*live.Room)
					r, ok := rm[v.ID]
					if !ok || r.LiveStatus != 1 {
						continue
					}
					item := &SmallCoverV2{
						CoverLeftText1: model.StatString(r.Online, ""),
						CoverLeftIcon1: model.IconOnline,
						Badge:          "直播",
						Base:           &Base{CardType: model.SmallCoverV2},
					}
					item.Base.from(op.Plat, op.Build, v.Param, r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
					// 使用直播接口返回的url
					if r.Link != "" {
						item.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
					}
					item.Args.fromLiveRoom(r)
					item.DescButton = buttonFrom(r, op.Plat)
					hander = item
				case map[int64]*article.Meta:
					mm := intfc.(map[int64]*article.Meta)
					m, ok := mm[v.ID]
					if !ok {
						continue
					}
					if len(m.ImageURLs) == 0 {
						continue
					}
					item := &SmallCoverV2{
						Badge: "文章",
						Base:  &Base{CardType: model.SmallCoverV2},
					}
					item.Base.from(op.Plat, op.Build, v.Param, m.ImageURLs[0], m.Title, model.GotoArticle, strconv.FormatInt(m.ID, 10), nil)
					if m.Stats != nil {
						item.CoverLeftText1 = model.StatString(int32(m.Stats.View), "")
						item.CoverLeftIcon1 = model.IconRead
						item.CoverLeftText2 = model.StatString(int32(m.Stats.Reply), "")
						item.CoverLeftIcon2 = model.IconComment
					}
					item.Args.fromArticle(m)
					item.DescButton = buttonFrom(m.Categories, op.Plat)
					hander = item
				case map[int64]*operate.Card:
					dm := intfc.(map[int64]*operate.Card)
					d, ok := dm[v.ID]
					if !ok {
						continue
					}
					item := &SmallCoverV2{
						CoverLeftText1: model.DownloadString(d.Download),
						Base:           &Base{CardType: model.SmallCoverV2},
					}
					item.Base.from(op.Plat, op.Build, v.Param, d.Coverm[c.Columnm], d.Title, d.Goto, d.URI, nil)
					hander = item
				case map[int64]*bangumi.Season:
					sm := intfc.(map[int64]*bangumi.Season)
					s, ok := sm[v.ID]
					if !ok {
						continue
					}
					item := &SmallCoverV2{
						CoverLeftText1: model.StatString(s.PlayCount, ""),
						CoverLeftIcon1: model.IconPlay,
						CoverLeftText2: model.StatString(s.Favorites, ""),
						CoverLeftIcon2: model.IconFavorite,
						Badge:          s.TypeBadge,
						Desc:           s.UpdateDesc,
						Base:           &Base{CardType: model.SmallCoverV2},
					}
					item.Base.from(op.Plat, op.Build, s.EpisodeID, s.Cover, s.Title, model.GotoBangumi, s.EpisodeID, nil)
					hander = item
				case map[int32]*episodegrpc.EpisodeCardsProto:
					sm := intfc.(map[int32]*episodegrpc.EpisodeCardsProto)
					s, ok := sm[int32(v.ID)]
					if !ok || s.Season == nil {
						continue
					}
					item := &SmallCoverV2{
						Badge: s.Season.SeasonTypeName,
						Desc:  s.Season.NewEpShow,
						Base:  &Base{CardType: model.SmallCoverV2},
					}
					if s.Season.Stat != nil {
						item.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
						item.CoverLeftIcon1 = model.IconPlay
						item.CoverLeftText2 = model.StatString(int32(s.Season.Stat.Follow), "")
						item.CoverLeftIcon2 = model.IconFavorite
					}
					title := s.Season.Title
					if s.ShowTitle != "" {
						title = title + "：" + s.ShowTitle
					}
					item.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.EpisodeId)), s.Cover, title, model.GotoBangumi, strconv.Itoa(int(s.EpisodeId)), nil)
					hander = item
					if s.Url != "" {
						item.URI = s.Url // 直接用pgc接口返回的URL
					}
					item.URI = model.FillURI("", 0, 0, item.URI, model.PGCTrackIDHandler(c.Rcmd))
				case map[int64]*bplus.Picture:
					pm := intfc.(map[int64]*bplus.Picture)
					p, ok := pm[v.ID]
					if !ok {
						continue
					}
					if len(p.Imgs) < 3 {
						hander = &OnePicV2{Base: &Base{CardType: model.OnePicV2}}
					} else {
						hander = &ThreePicV2{Base: &Base{CardType: model.ThreePicV2}}
					}
					//nolint:errcheck
					hander.From(pm, v)
					if !hander.Get().Right {
						continue
					}
					c.Args.fromPicture(p)
				default:
					log.Warn("MultiItem From: unexpected type %T", intfc)
					continue
				}
				if hander != nil {
					items = append(items, hander)
				}
			}
			if len(items) < 2 {
				return newInvalidResourceErr(ResourceItems, 0, "lack of items")
			}
			if len(items)%2 != 0 {
				c.Items = items[:len(items)-1]
			} else {
				c.Items = items
			}
			var title string
			switch op.Goto {
			case model.GotoTag:
				if t, ok := c.Tagm[op.ID]; ok {
					title = t.Name
				}
			default:
				title = op.Title
			}
			c.Base.from(op.Plat, op.Build, op.Param, "", title, "", "", nil)
			c.MoreURI = model.FillURI(op.Goto, 0, 0, op.URI, nil)
			c.MoreText = op.Subtitle
		default:
			log.Warn("MultiItem From: unexpected card_goto %s", op.CardGoto)
			return newUnexpectedCardGotoErr(string(op.CardGoto), "")
		}
	case nil:
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoEntrance:
			c.Items = make([]Handler, 0, len(op.Items))
			for _, v := range op.Items {
				item := &SmallCoverV2{Base: &Base{CardType: model.SmallCoverV2}}
				item.Base.from(op.Plat, op.Build, v.Param, v.Cover, v.Title, v.Goto, v.URI, nil)
				c.Items = append(c.Items, item)
			}
		}
	default:
		log.Warn("MultiItem From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *MultiItem) Get() *Base {
	return c.Base
}

type ThreePicV2 struct {
	*Base
	LeftCover                 string       `json:"left_cover,omitempty"`
	RightCover1               string       `json:"right_cover_1,omitempty"`
	RightCover2               string       `json:"right_cover_2,omitempty"`
	CoverLeftText1            string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1            model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2            string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2            model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverRightText            string       `json:"cover_right_text,omitempty"`
	CoverRightIcon            model.Icon   `json:"cover_right_icon,omitempty"`
	CoverRightBackgroundColor string       `json:"cover_right_background_color,omitempty"`
	Badge                     string       `json:"badge,omitempty"`
	RcmdReason                string       `json:"rcmd_reason,omitempty"`
	DescButton                *Button      `json:"desc_button,omitempty"`
	Desc                      string       `json:"desc,omitempty"`
	Avatar                    *Avatar      `json:"avatar,omitempty"`
	RcmdReasonStyle           *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *ThreePicV2) From(main interface{}, op *operate.Card) error {
	var (
		button interface{}
	)
	if op == nil {
		return newEmptyOPErr()
	}
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
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
		}
		c.LeftCover = p.Imgs[0]
		c.RightCover1 = p.Imgs[1]
		c.RightCover2 = p.Imgs[2]
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.StatString(p.LikeCount, "")
			c.CoverLeftIcon1 = model.IconLike
		} else {
			c.CoverLeftText1 = model.StatString(int32(p.ViewCount), "")
			c.CoverLeftIcon1 = model.IconRead
		}
		if p.ImgCount > 3 {
			c.CoverRightText = model.PictureCountString(p.ImgCount)
			c.CoverRightBackgroundColor = "#66666666"
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReason, c.Desc = rcmdReason(c.Rcmd.RcmdReason, p.NickName, c.IsAttenm[p.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else {
			button = p
			c.Badge = "动态"
		}
		c.Avatar = avatarFrom(&AvatarStatus{Cover: p.FaceImg, Text: p.NickName, Goto: model.GotoDynamicMid, Param: strconv.FormatInt(p.Mid, 10), Type: model.AvatarRound})
		c.Args.fromPicture(p)
	default:
		log.Warn("ThreePicV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.DescButton = buttonFrom(button, op.Plat)
	c.Right = true
	return nil
}

func (c *ThreePicV2) Get() *Base {
	return c.Base
}

type OnePicV2 struct {
	*Base
	CoverLeftText1            string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1            model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverRightText            string       `json:"cover_right_text,omitempty"`
	CoverRightIcon            model.Icon   `json:"cover_right_icon,omitempty"`
	CoverRightBackgroundColor string       `json:"cover_right_background_color,omitempty"`
	Badge                     string       `json:"badge,omitempty"`
	RcmdReason                string       `json:"rcmd_reason,omitempty"`
	Avatar                    *Avatar      `json:"avatar,omitempty"`
	RcmdReasonStyle           *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *OnePicV2) From(main interface{}, op *operate.Card) error {
	var (
		button interface{}
	)
	if op == nil {
		return newEmptyOPErr()
	}
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
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
		}
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.StatString(p.LikeCount, "")
			c.CoverLeftIcon1 = model.IconLike
		} else {
			c.CoverLeftText1 = model.StatString(int32(p.ViewCount), "")
			c.CoverLeftIcon1 = model.IconRead
		}
		if p.ImgCount > 1 {
			c.CoverRightText = model.PictureCountString(p.ImgCount)
			c.CoverRightBackgroundColor = "#66666666"
		}
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			c.RcmdReason, _ = rcmdReason(c.Rcmd.RcmdReason, p.NickName, c.IsAttenm[p.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.RcmdReason, c.Base.Goto, op)
		} else {
			button = p
			c.Badge = "动态"
		}
		c.Avatar = avatarFrom(&AvatarStatus{Cover: p.FaceImg, Text: p.NickName, Goto: model.GotoDynamicMid, Param: strconv.FormatInt(p.Mid, 10), Type: model.AvatarRound})
		c.Args.fromPicture(p)
	default:
		log.Warn("OnePicV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.DescButton = buttonFrom(button, op.Plat)
	c.Right = true
	return nil
}

func (c *OnePicV2) Get() *Base {
	return c.Base
}

type LargeCoverV3 struct {
	*Base
	CoverGif               string       `json:"cover_gif,omitempty"`
	Avatar                 *Avatar      `json:"avatar,omitempty"`
	TopRcmdReasonStyle     *ReasonStyle `json:"top_rcmd_reason_style,omitempty"`
	BadgeStyle             *ReasonStyle `json:"badge_style,omitempty"`
	BottomRcmdReasonStyle  *ReasonStyle `json:"bottom_rcmd_reason_style,omitempty"`
	CoverLeftText1         string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1         model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2         string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2         model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverRightText         string       `json:"cover_right_text,omitempty"`
	Desc                   string       `json:"desc,omitempty"`
	OfficialIcon           model.Icon   `json:"official_icon,omitempty"`
	LeftCoverBadgeNewStyle *ReasonStyle `json:"left_cover_badge_new_style,omitempty"`
}

func (c *LargeCoverV3) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		upID   int64
		avatar *AvatarStatus
		furi   func(uri string) string
		uri    string
	)
	switch main.(type) {
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		intfc, ok := intfcm[op.Goto]
		if !ok {
			return newUnexpectedGotoErr(string(op.Goto), "")
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
			authorface := a.Arc.Author.Face
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			avatar = &AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
			c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
			c.CoverLeftIcon2 = model.IconDanmaku
			c.CoverRightText = model.DurationString(a.Arc.Duration)
			c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			c.Base.PlayerArgs = playerArgsFrom(a)
			furi = model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
			c.Desc = authorname + " · " + model.PubDataString(a.Arc.PubDate.Time())
			upID = a.Arc.Author.Mid
		case map[int32]*episodegrpc.EpisodeCardsProto:
			sm := intfc.(map[int32]*episodegrpc.EpisodeCardsProto)
			s, ok := sm[int32(op.ID)]
			if !ok {
				return newResourceNotExistErr(ResourceEpisode, op.ID)
			}
			if s.Season == nil {
				return newInvalidResourceErr(ResourceEpisode, int64(s.EpisodeId), "empty `Season`")
			}
			c.Goto = model.GotoPGC
			if s.Season.Stat != nil {
				c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
				c.CoverLeftIcon1 = model.IconPlay
				c.CoverLeftText2 = model.StatString(int32(s.Season.Stat.Follow), "")
				c.CoverLeftIcon2 = model.IconFavorite
			}
			avatar = &AvatarStatus{Cover: s.Season.Cover, Goto: model.GotoPGC, Param: strconv.Itoa(int(s.Season.SeasonId)), Type: model.AvatarSquare}
			if s.Season != nil {
				c.Desc = s.Season.NewEpShow
			}
			uri = model.FillURI("", 0, 0, s.Url, model.PGCTrackIDHandler(c.Rcmd)) // 直接用pgc接口返回的URL
		case map[int64]*live.Room:
			rm := intfc.(map[int64]*live.Room)
			r, ok := rm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceRoom, op.ID)
			}
			if r.LiveStatus != 1 {
				return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
			}
			avatar = &AvatarStatus{Cover: r.Face, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound}
			c.CoverLeftText1 = model.StatString(r.Online, "")
			c.CoverLeftIcon1 = model.IconOnline
			c.Args.fromLiveRoom(r)
			c.Base.PlayerArgs = playerArgsFrom(r)
			furi = model.LiveRoomHandler(r, op.Network)
			// 使用直播接口返回的url
			if r.Link != "" {
				uri = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
			}
			upID = r.UID
			c.Desc = r.Uname
		case map[int64]*article.Meta:
			mm := intfc.(map[int64]*article.Meta)
			m, ok := mm[op.ID]
			if !ok {
				return newResourceNotExistErr(ResourceArticle, op.ID)
			}
			if m.Stats != nil {
				c.CoverLeftText1 = model.StatString(int32(m.Stats.View), "")
				c.CoverLeftIcon1 = model.IconRead
				c.CoverLeftText2 = model.StatString(int32(m.Stats.Reply), "")
				c.CoverLeftIcon2 = model.IconComment
			}
			if m.Author != nil {
				avatar = &AvatarStatus{Cover: m.Author.Face, Text: m.Author.Name + "·" + model.PubDataString(m.PublishTime.Time()), Goto: model.GotoMid, Param: strconv.FormatInt(m.Author.Mid, 10), Type: model.AvatarRound}
				upID = m.Author.Mid
				c.Desc = m.Author.Name
			}
			c.Args.fromArticle(m)
		default:
			log.Warn("LargeCoverV3 From: unexpected type %T", intfc)
			return newUnexpectedResourceTypeErr(intfc, "")
		}
		c.Avatar = avatarFrom(avatar)
		_, newRcmd := op.SwitchStyle[model.SwitchNewReason]
		_, newRcmdV2 := op.SwitchStyle[model.SwitchNewReasonV2]
		if newRcmd || newRcmdV2 {
			c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
			if c.Rcmd != nil {
				rcmdReason, _ := rcmdReason(c.Rcmd.RcmdReason, "", c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
				c.BottomRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
			}
		} else {
			if op.Badge != "" {
				c.BottomRcmdReasonStyle = reasonStyleFrom(model.BgColorTransparentRed, op.Badge)
			} else if c.Rcmd != nil {
				rcmdReason, _ := rcmdReason(c.Rcmd.RcmdReason, "", c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
				c.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
			}
		}
		c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, op.Goto, op.URI, furi)
		if uri != "" {
			c.URI = uri
		}
		c.CoverGif = op.GifCover
		c.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
		c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	default:
		log.Warn("LargeCoverV3 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *LargeCoverV3) Get() *Base {
	return c.Base
}

type ThreePicV3 struct {
	*Base
	LeftCover                 string       `json:"left_cover,omitempty"`
	RightCover1               string       `json:"right_cover_1,omitempty"`
	RightCover2               string       `json:"right_cover_2,omitempty"`
	CoverLeftText1            string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1            model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2            string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2            model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverRightText            string       `json:"cover_right_text,omitempty"`
	CoverRightIcon            model.Icon   `json:"cover_right_icon,omitempty"`
	CoverRightBackgroundColor string       `json:"cover_right_background_color,omitempty"`
	Badge                     string       `json:"badge,omitempty"`
	BadgeStyle                *ReasonStyle `json:"badge_style,omitempty"`
	DescButton                *Button      `json:"desc_button,omitempty"`
	RcmdReasonStyle           *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *ThreePicV3) From(main interface{}, op *operate.Card) error {
	var (
		button interface{}
	)
	if op == nil {
		return newEmptyOPErr()
	}
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
		c.Badge = "动态"
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
			c.Badge = "活动"
		}
		c.LeftCover = p.Imgs[0]
		c.RightCover1 = p.Imgs[1]
		c.RightCover2 = p.Imgs[2]
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.StatString(p.LikeCount, "")
			c.CoverLeftIcon1 = model.IconLike
		} else {
			c.CoverLeftText1 = model.StatString(int32(p.ViewCount), "")
			c.CoverLeftIcon1 = model.IconRead
		}
		c.CoverRightText = p.NickName
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			reasonText, _ := rcmdReason(c.Rcmd.RcmdReason, p.NickName, c.IsAttenm[p.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, reasonText, c.Base.Goto, op)
		} else {
			button = p
		}
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, c.Badge)
		c.Args.fromPicture(p)
	default:
		log.Warn("ThreePicV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.DescButton = buttonFrom(button, op.Plat)
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	c.Right = true
	return nil
}

func (c *ThreePicV3) Get() *Base {
	return c.Base
}

type OnePicV3 struct {
	*Base
	CoverLeftText1            string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1            model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverRightText            string       `json:"cover_right_text,omitempty"`
	CoverRightIcon            model.Icon   `json:"cover_right_icon,omitempty"`
	CoverRightBackgroundColor string       `json:"cover_right_background_color,omitempty"`
	Badge                     string       `json:"badge,omitempty"`
	BadgeStyle                *ReasonStyle `json:"badge_style,omitempty"`
	RcmdReasonStyle           *ReasonStyle `json:"rcmd_reason_style,omitempty"`
}

func (c *OnePicV3) From(main interface{}, op *operate.Card) error {
	var (
		button interface{}
	)
	if op == nil {
		return newEmptyOPErr()
	}
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
		c.Badge = "动态"
		if p.JumpUrl != "" && ((model.IsAndroid(op.Plat) && op.Build >= 5510000) || (model.IsIOS(op.Plat) && op.Build > 8960)) {
			c.URI = p.JumpUrl
			c.Badge = "活动"
		}
		if _, ok := op.SwitchStyle[model.SwitchPictureLike]; ok {
			c.CoverLeftText1 = model.StatString(p.LikeCount, "")
			c.CoverLeftIcon1 = model.IconLike
		} else {
			c.CoverLeftText1 = model.StatString(int32(p.ViewCount), "")
			c.CoverLeftIcon1 = model.IconRead
		}
		c.CoverRightText = p.NickName
		if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
			reasonText, _ := rcmdReason(c.Rcmd.RcmdReason, p.NickName, c.IsAttenm[p.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, reasonText, c.Base.Goto, op)
		} else {
			button = p
		}
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, c.Badge)
		c.Args.fromPicture(p)
	default:
		log.Warn("OnePicV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.DescButton = buttonFrom(button, op.Plat)
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	c.Right = true
	return nil
}

func (c *OnePicV3) Get() *Base {
	return c.Base
}

type SmallCoverV7 struct {
	*Base
	Desc         string `json:"desc,omitempty"`
	DestroyCard  int8   `json:"destroy_card"`
	ResourceType string `json:"resource_type,omitempty"`
	GameID       int64  `json:"game_id,omitempty"`
}

func (c *SmallCoverV7) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	c.DestroyCard = 1 // 6.9 版本新字段，不 omitempty，渠道卡依赖渠道响应的字段
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
		c.Desc = vipInfo.Tip
		c.ResourceType = "vip_renew"
	case *tunnelgrpc.FeedCard:
		tunnelCard := main.(*tunnelgrpc.FeedCard)
		if tunnelCard == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		c.Base.from(op.Plat, op.Build, op.Param, tunnelCard.Cover, tunnelCard.Title, model.GotoWeb, tunnelCard.Link, nil)
		c.Goto = model.GotoGame // 曾经渠道卡里只有游戏
		if tunnelCard.Goto != "" {
			c.Goto = model.Gt(tunnelCard.Goto)
		}
		c.ResourceType = cvtTunnelResourceType(tunnelCard.ResourceType)
		switch c.ResourceType {
		case "game_tunnel":
			// 6.9 版本通过 param 兼容游戏 ID
			if (op.Plat == model.PlatIPhone && op.Build >= 10260 && op.Build <= 10270) ||
				(op.Plat == model.PlatAndroid && op.Build >= 6090600 && op.Build <= 6091000) {
				c.Base.Param = strconv.FormatInt(tunnelCard.UniqueId, 10)
			}
			c.GameID = tunnelCard.UniqueId
		}
		c.Desc = tunnelCard.Intro
		c.DestroyCard = int8(tunnelCard.Destroy)
	default:
		log.Warn("SmallCoverV7 From: unexpected card_goto %s", op.CardGoto)
		return newUnexpectedCardGotoErr(string(op.CardGoto), "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverV7) Get() *Base {
	return c.Base
}

type SmallCoverV9 struct {
	*Base
	CoverLeftText1               string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1               model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeft1ContentDescription string       `json:"cover_left_1_content_description,omitempty"`
	CoverLeftText2               string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2               model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverRightText               string       `json:"cover_right_text,omitempty"`
	CoverRightContentDescription string       `json:"cover_right_content_description,omitempty"`
	CoverRightIcon               model.Icon   `json:"cover_right_icon,omitempty"`
	CanPlay                      int32        `json:"can_play,omitempty"`
	Up                           *Up          `json:"up,omitempty"`
	LeftCoverBadgeStyle          *ReasonStyle `json:"left_cover_badge_style,omitempty"`
	LeftBottomRcmdReasonStyle    *ReasonStyle `json:"left_bottom_rcmd_reason_style,omitempty"`
	IsAtten                      bool         `json:"is_atten,omitempty"`
	OfficialIconV2               model.Icon   `json:"official_icon_v2,omitempty"`
	LeftCoverBadgeNewStyle       *ReasonStyle `json:"left_cover_badge_new_style,omitempty"`
	OffBadgeStyle                *ReasonStyle `json:"off_badge_style,omitempty"`
	HideDanmuSwitch              bool         `json:"hide_danmu_switch,omitempty"` // 弹幕开关隐藏
	DisableDanmu                 bool         `json:"disable_danmu,omitempty"`     // 禁用弹幕
}

func (c *SmallCoverV9) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if r.LiveStatus != 1 {
			return newInvalidResourceErr(ResourceRoom, r.RoomID, "LiveStatus: %d", r.LiveStatus)
		}
		if op.CardGoto == model.CardGotoAdLive {
			c.AdInfo = op.AdInfo
		}
		c.Base.from(op.Plat, op.Build, op.Param, r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.CoverLeftText1 = model.StatString(r.Online, "")
		c.CoverLeftIcon1 = model.IconOnline
		c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
		c.CoverRightText = r.AreaV2Name
		c.CoverRightContentDescription = r.AreaV2Name
		c.LeftCoverBadgeNewStyle = setupLiveLeftCoverBadge(r, op)
		c.LeftBottomRcmdReasonStyle = iconBadgeStyleFrom(op.LiveLeftBottomBadgeStyle, 0)
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		c.CanPlay = 1
		c.OfficialIconV2 = model.OfficialIcon(c.Cardm[r.UID])
		c.Up = &Up{
			ID:           r.UID,
			Name:         r.Uname,
			OfficialIcon: model.OfficialIcon(c.Cardm[r.UID]),
		}
		c.Up.Avatar = &Avatar{
			Cover: r.Face,
			URI:   c.URI,
			Event: model.EventMainCard,
		}
		if c.IsAttenm[r.UID] == 1 {
			c.Up.OfficialIcon = model.IconIsAttenm
			c.IsAtten = true
		}
	default:
		log.Warn("SmallCoverV9 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func setupLiveLeftCoverBadge(live *live.Room, op *operate.Card) *ReasonStyle {
	if len(op.LiveLeftCoverBadgeStyle) == 0 {
		return nil
	}
	if len(live.AllPendants) == 0 && live.HotRank <= 0 {
		return nil
	}
	badgeMap := ConvertMap(op.LiveLeftCoverBadgeStyle)
	pendants := ConstructPendants(live)
	var badgeList []*operate.V9LiveLeftCoverBadge
	for _, pendant := range pendants {
		badge, ok := badgeMap[fmt.Sprintf("%s:%s", pendant.Type, pendant.Name)]
		if !ok {
			continue
		}
		badgeList = append(badgeList, badge)
	}
	if len(badgeList) == 0 {
		return nil
	}
	sort.Slice(badgeList, func(i, j int) bool { return badgeList[i].Priority < badgeList[j].Priority })
	newStyle := &ReasonStyle{
		IconURL:      badgeList[0].NewStyleIconURL,
		IconURLNight: badgeList[0].NewStyleIconURLNight,
		IconWidth:    badgeList[0].IconWidth,
		IconHeight:   badgeList[0].IconHeight,
	}
	return newStyle
}

func ConstructPendants(room *live.Room) []*live.Pendants {
	out := make([]*live.Pendants, 0, len(room.AllPendants))
	//nolint:gosimple
	for _, v := range room.AllPendants {
		out = append(out, v)
	}
	if room.HotRank > 0 {
		out = append(out, &live.Pendants{
			Type: "mobile_index_badge",
			Name: "直播热门TOP10",
		})
	}
	return out
}

func ConvertMap(badge []*operate.V9LiveLeftCoverBadge) map[string]*operate.V9LiveLeftCoverBadge {
	out := make(map[string]*operate.V9LiveLeftCoverBadge, len(badge))
	for _, v := range badge {
		out[v.Key] = v
	}
	return out
}

func (c *SmallCoverV9) Get() *Base {
	return c.Base
}

type SmallCoverConvergeV2 struct {
	*Base
	CoverLeftText1    string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1    model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2    string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2    model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverRightText    string       `json:"cover_right_text,omitempty"`
	CoverRightTopText string       `json:"cover_right_top_text,omitempty"`
	RcmdReasonStyle   *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	RcmdReasonStyleV2 *ReasonStyle `json:"rcmd_reason_style_v2,omitempty"`
}

func (c *SmallCoverConvergeV2) From(main interface{}, op *operate.Card) error {
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
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
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
				if view := cinfo.Quality.Play; view > 0 {
					c.CoverLeftText1 = model.Stat64String(view, "")
				}
				if danmaku := cinfo.Quality.Danmu; danmaku > 0 {
					c.CoverLeftText2 = model.Stat64String(danmaku, "")
				}
				if cinfo.Count > 0 {
					c.CoverRightTopText = strconv.Itoa(cinfo.Count) + "个内容"
				}
				if cinfo.Title != "" {
					title = cinfo.Title
				}
			}
		}
		if op.Title != "" {
			title = op.Title
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, title, model.GotoAv, strconv.FormatInt(a.Arc.Aid, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		c.CoverRightText = model.DurationString(a.Arc.Duration)
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
				c.Base.URI = model.FillURI(op.Goto, op.Plat, op.Build, op.Param, model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
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
		log.Warn("SmallCoverConvergeV2 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *SmallCoverConvergeV2) Get() *Base {
	return c.Base
}

type SmallChannelSpecial struct {
	*Base
	BgCover string `json:"bg_cover,omitempty"`
	Desc1   string `json:"desc_1,omitempty"`
	Desc2   string `json:"desc_2,omitempty"`
	Badge   string `json:"badge,omitempty"`
	// RcmdReasonStyle  *ReasonStyle `json:"rcmd_reason_style_1,omitempty"`
	RcmdReasonStyle2 *ReasonStyle `json:"rcmd_reason_style_2,omitempty"`
	BadgeStyle       *ReasonStyle `json:"badge_style,omitempty"`
}

func (c *SmallChannelSpecial) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	if op.Desc != "" {
		c.Desc1 = model.MarkRed(op.Desc)
	}
	if op.Channel != nil {
		c.BgCover = op.Channel.BgCover
		c.Desc2 = op.Channel.Reason
	}
	c.Base.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoChannel, strconv.FormatInt(op.ID, 10), model.ChannelHandler("tab=all"))
	if op.Channel.TabURI != "" {
		c.URI = op.Channel.TabURI
	}
	// c.RcmdReasonStyle = reasonStyleFrom(model.BgColorTransparentRed, "频道", false)
	c.Badge = "频道"
	c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "频道")
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*channelgrpc.ChannelCard:
		chm := main.(map[int64]*channelgrpc.ChannelCard)
		if ch, ok := chm[op.ID]; ok && ch != nil {
			if ch.Subscribed {
				if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
					c.RcmdReasonStyle2 = reasonStyleFrom(model.BgColorLumpOrange, "已订阅")
				} else if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
					c.RcmdReasonStyle2 = reasonStyleFrom(model.BgColorFillingOrange, "已订阅")
				} else {
					c.RcmdReasonStyle2 = reasonStyleFrom(model.BgColorOrange, "已订阅")
				}
			}
			if c.Desc1 == "" {
				var descs []string
				if ch.RCnt != 0 {
					cnt := model.StatString(ch.RCnt, "")
					descs = append(descs, fmt.Sprintf("%s投稿", model.MarkRed(cnt)))
				}
				if ch.FeaturedCnt != 0 {
					cnt := model.StatString(ch.FeaturedCnt, "")
					descs = append(descs, fmt.Sprintf("%s个精选视频", model.MarkRed(cnt)))
				}
				if len(descs) > 0 {
					c.Desc1 = strings.Join(descs, " , ")
				}
			}
		}
	}
	c.Right = true
	return nil
}

func (c *SmallChannelSpecial) Get() *Base {
	return c.Base
}

type LargeCoverV5 struct {
	*Base
	Avatar          *Avatar      `json:"avatar,omitempty"`
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1  model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2  string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2  model.Icon   `json:"cover_left_icon_2,omitempty"`
	RcmdReasonStyle *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	DescButton      *Button      `json:"desc_button,omitempty"`
	BadgeStyle      *ReasonStyle `json:"badge_style,omitempty"`
	OfficialIcon    model.Icon   `json:"official_icon,omitempty"`
	CanPlay         int32        `json:"can_play,omitempty"`
	CoverRightText  string       `json:"cover_right_text,omitempty"`
	// story详情页需要的封面图
	FfCover string `json:"ff_cover,omitempty"`
}

func (c *LargeCoverV5) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		button interface{}
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
		var furi func(uri string) string
		if a.Arc.RedirectURL == "" {
			furi = model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
			c.CanPlay = a.Arc.Rights.Autoplay
			c.Base.PlayerArgs = playerArgsFrom(a)
			if c.Base.PlayerArgs == nil {
				log.Warn("LargeCoverV5 player card aid(%d) can't auto player", a.Arc.Aid)
				return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "PlayerArgs")
			}
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, furi)
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverRightText = model.DurationString(a.Arc.Duration)
		if c.Rcmd != nil {
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorname = au.Name
				}
			}
			rcmdReason, desc := rcmdReason(c.Rcmd.RcmdReason, authorname, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
			if desc != "" {
				button = &ButtonStatus{Text: desc}
			}
		}
		if button == nil {
			if op.Channel != nil && op.Channel.ChannelID != 0 && op.Channel.ChannelName != "" {
				op.Channel.ChannelName = a.Arc.TypeName + " · " + op.Channel.ChannelName
				button = op.Channel
			} else if t, ok := c.Tagm[op.Tid]; ok {
				tag := &taggrpc.Tag{}
				*tag = *t
				tag.Name = a.Arc.TypeName + " · " + tag.Name
				button = tag
			} else {
				button = &ButtonStatus{Text: a.Arc.TypeName}
			}
		}
		c.DescButton = buttonFrom(button, op.Plat)
		c.OfficialIcon = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		if c.Rcmd != nil {
			//nolint:exhaustive
			switch model.Gt(c.Rcmd.JumpGoto) {
			case model.GotoVerticalAv:
				c.FfCover = c.Rcmd.FfCover
				c.Goto = model.GotoVerticalAv
				c.URI = model.FillURI(model.GotoVerticalAv, op.Plat, op.Build, op.Param, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, c.Rcmd, op.Build, op.MobiApp, true))
			}
		}
	default:
		log.Warn("LargeCoverV5 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (c *LargeCoverV5) Get() *Base {
	return c.Base
}

type LargeCoverInline struct {
	*Base
	Avatar                       *Avatar             `json:"avatar,omitempty"`
	OfficialIcon                 model.Icon          `json:"official_icon,omitempty"`
	CoverLeftText1               string              `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1               model.Icon          `json:"cover_left_icon_1,omitempty"`
	CoverLeft1ContentDescription string              `json:"cover_left_1_content_description,omitempty"`
	CoverLeftText2               string              `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2               model.Icon          `json:"cover_left_icon_2,omitempty"`
	CoverLeft2ContentDescription string              `json:"cover_left_2_content_description,omitempty"`
	CoverRightText               string              `json:"cover_right_text,omitempty"`
	CoverRightContentDescription string              `json:"cover_right_content_description,omitempty"`
	RcmdReasonStyle              *ReasonStyle        `json:"rcmd_reason_style,omitempty"`
	CanPlay                      int32               `json:"can_play,omitempty"`
	LeftCoverBadgeNewStyle       *ReasonStyle        `json:"left_cover_badge_new_style,omitempty"`
	BadgeStyle                   *ReasonStyle        `json:"badge_style,omitempty"`
	LikeButton                   *LikeButton         `json:"like_button,omitempty"`
	PlayerWidget                 *InlinePlayerWidget `json:"player_widget,omitempty"`
	OfficialIconV2               model.Icon          `json:"official_icon_v2,omitempty"`
	IsAtten                      bool                `json:"is_atten,omitempty"`
	// story详情页需要的封面图
	FfCover string `json:"ff_cover,omitempty"`
	// 跳转story图标
	GotoIcon              *model.GotoIcon    `json:"goto_icon,omitempty"`
	IsFav                 bool               `json:"is_fav,omitempty"`
	IsHot                 bool               `json:"is_hot,omitempty"`
	SharePlane            *model.SharePlane  `json:"share_plane,omitempty"`
	HideDanmuSwitch       bool               `json:"hide_danmu_switch,omitempty"` // 弹幕开关隐藏，15版本开始只看卡片级开关，不看config内弹幕开关
	DisableDanmu          bool               `json:"disable_danmu,omitempty"`     // 禁用弹幕，15版本开始只看卡片级开关，不看config内弹幕开关
	ExtraURI              string             `json:"extra_uri,omitempty"`
	InlineProgressBar     *InlineProgressBar `json:"inline_progress_bar,omitempty"` // 小电视播放进度icon
	IsCoin                bool               `json:"is_coin,omitempty"`
	RightTopLiveBadge     *LiveBadge         `json:"right_top_live_badge,omitempty"`
	Desc                  string             `json:"desc,omitempty"`
	MultiplyDesc          *MultiplyDesc      `json:"multiply_desc,omitempty"`
	CoverBadge            string             `json:"cover_badge,omitempty"`
	CoverBadgeStyle       *ReasonStyle       `json:"cover_badge_style,omitempty"`
	EnableDoubleClickLike bool               `json:"enable_double_click_like,omitempty"`
	OgvCreativeId         int64              `json:"ogv_creative_id,omitempty"`
	ServerInfo            string             `json:"server_info,omitempty"` // 港澳台tab
}

type MultiplyDesc struct {
	AuthorName string `json:"author_name,omitempty"`
	Extra      string `json:"extra,omitempty"`
	Type       int8   `json:"type,omitempty"`
}

type ReplyButton struct {
	Count     int32       `json:"count,omitempty"`
	ShowCount bool        `json:"show_count,omitempty"`
	Event     model.Event `json:"event,omitempty"`
	EventV2   model.Event `json:"event_v2,omitempty"`
}

type LiveBadge struct {
	LiveStatus int8               `json:"live_status,omitempty"`
	InLive     *LiveBadgeResource `json:"in_live,omitempty"`
}

type LiveBadgeResource struct {
	Text                 string `json:"text,omitempty"`
	AnimationURL         string `json:"animation_url,omitempty"`
	AnimationURLHash     string `json:"animation_url_hash,omitempty"`
	BackgroundColorLight string `json:"background_color_light,omitempty"`
	BackgroundColorNight string `json:"background_color_night,omitempty"`
	AlphaLight           int8   `json:"alpha_light,omitempty"`
	AlphaNight           int8   `json:"alpha_night,omitempty"`
	FontColor            string `json:"font_color,omitempty"`
}

type InlineProgressBar struct {
	IconDrag     string `json:"icon_drag,omitempty"`
	IconDragHash string `json:"icon_drag_hash,omitempty"`
	IconStop     string `json:"icon_stop,omitempty"`
	IconStopHash string `json:"icon_stop_hash,omitempty"`
}

type InlinePlayerWidget struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

func (c *LargeCoverInline) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	var (
		rcmdReasonText string
	)
	if c.Rcmd != nil && c.Rcmd.RcmdReason != nil {
		rcmdReasonText = c.Rcmd.RcmdReason.Content
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, op.ID)
		}
		if !model.AvIsNormalGRPC(a) {
			return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "AvIsNormalGRPC")
		}
		var furi func(uri string) string
		if a.Arc.RedirectURL == "" {
			furi = model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
			c.CanPlay = a.Arc.Rights.Autoplay
			c.Base.PlayerArgs = playerArgsFrom(a)
			if c.Base.PlayerArgs == nil {
				log.Warn("LargeCoverV5 player card aid(%d) can't auto player", a.Arc.Aid)
				return newInvalidResourceErr(ResourceArchive, a.Arc.Aid, "PlayerArgs")
			}
		}
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, furi)
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
		c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
		c.CoverRightText = model.DurationString(a.Arc.Duration)
		c.CoverRightContentDescription = model.DurationContentDescription(a.Arc.Duration)
		c.LikeButton = likeButtonFromGRPC(a.Arc, c.HasLike[a.Arc.Aid], op)
		if c.Rcmd != nil {
			authorname := a.Arc.Author.Name
			if c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorname = au.Name
				}
			}
			rcmdReasonText, _ = rcmdReason(c.Rcmd.RcmdReason, authorname, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
		}
		c.Args.fromArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		c.Args.IsFollow = c.IsAttenm[a.Arc.Author.Mid]
		c.OfficialIcon = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
		c.OfficialIconV2 = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
		if c.IsAttenm[a.Arc.Author.Mid] == 1 {
			c.OfficialIcon = model.IconIsAttenm
			c.IsAtten = true
			c.Avatar = avatarFrom(&AvatarStatus{Cover: a.Arc.Author.Face, Text: a.Arc.Author.Name, Goto: model.GotoMid,
				Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound})
		}
		if model.CardGt(c.Rcmd.Goto) == model.CardGotoInlineAvV2 || isSingleInline(c.Rcmd) {
			if isFav, ok := op.HasFav[a.Arc.Aid]; ok && isFav == 1 {
				c.IsFav = true
			}
			if isCoin, ok := op.HasCoin[a.Arc.Aid]; ok && isCoin > 0 {
				c.IsCoin = true
			}
			if op.HotAidSet.Has(a.Arc.Aid) {
				c.IsHot = true
			}
			shareSubtitle, playNumber := GetShareSubtitle(a.Arc.Stat.View)
			bvid_, _ := GetBvID(a.Arc.Aid)
			c.SharePlane = &model.SharePlane{
				Title:         c.Title,
				ShareSubtitle: shareSubtitle,
				Desc:          a.Arc.Desc,
				Cover:         c.Cover,
				Aid:           a.Arc.Aid,
				Bvid:          bvid_,
				ShareTo:       model.ShareTo,
				Author:        a.Arc.Author.Name,
				AuthorId:      a.Arc.Author.Mid,
				ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%d", c.Rcmd.ID),
				PlayNumber:    playNumber,
			}
			c.settingInlineIcon(op)
			c.settingUGCThreePointPanelMeta(op)
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
		c.UpArgs = upArgsFrom(a.Arc, c.IsAttenm[a.Arc.Author.Mid])
		if isSingleInline(c.Rcmd) {
			c.Desc = a.Arc.Author.Name + " · " + model.PubDataByRequestAt(a.Arc.PubDate.Time(), c.Rcmd.RequestAt())
			c.Avatar = avatarFrom(&AvatarStatus{Cover: a.Arc.Author.Face, Text: a.Arc.Author.Name, Goto: model.GotoMid,
				Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound})
			rcmdReasonText = "" // 单列inline暂时屏蔽推荐理由
		}
	case map[int32]*pgcinline.EpisodeCard:
		eps := main.(map[int32]*pgcinline.EpisodeCard)
		ep, ok := eps[int32(op.ID)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		if ep.Season == nil {
			return newInvalidResourceErr(ResourceEpisode, op.ID, "empty `Season`")
		}
		title := fmt.Sprintf("%s %s", ep.Season.Title, ep.NewDesc)
		c.Base.from(op.Plat, op.Build, op.Param, ep.Cover, title, model.GotoPGC, "", nil)
		c.URI = model.FillURI("", 0, 0, ep.Url, model.PGCTrackIDHandler(c.Rcmd))
		if ep.PlayerInfo != nil {
			c.CanPlay = 1
		}
		if c.Rcmd.NoPlay == 1 {
			c.CanPlay = 0
		}
		if ep.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(ep.Stat.Play), "")
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText2 = model.StatString(int32(ep.Stat.Follow), "")
			c.CoverLeftIcon2 = model.IconFavorite
			c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
			c.CoverLeft2ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon2, c.CoverLeftText2)
		}
		c.LikeButton = likeButtonFromEpisodeCard(ep, c.HasLike[ep.Aid], op)
		c.CoverRightText = model.DurationString(ep.Duration)
		c.CoverRightContentDescription = model.DurationContentDescription(ep.Duration)
		c.Base.PlayerArgs = playerArgsFrom(ep)
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, ep.Season.TypeName)
		c.settingInlineIcon(op)
		c.settingOGVThreePointPanelMeta(op)
		bvid_, _ := GetBvID(ep.Aid)
		shareSubtitle, playNumber := GetShareSubtitle(int32(ep.GetStat().Play))
		c.SharePlane = &model.SharePlane{
			Title:         c.Title,
			ShareSubtitle: shareSubtitle,
			Desc:          ep.GetSeason().GetNewEpShow(),
			Cover:         c.Cover,
			Aid:           ep.Aid,
			Bvid:          bvid_,
			EpId:          ep.EpisodeId,
			SeasonId:      ep.GetSeason().GetSeasonId(),
			ShareTo:       model.ShareTo,
			PlayNumber:    playNumber,
			ShareFrom:     model.InlinePGCShareFrom,
			SeasonTitle:   ep.GetSeason().GetTitle(),
		}
		if ep.Widget != nil {
			c.PlayerWidget = &InlinePlayerWidget{
				Title: ep.Widget.Title,
				Desc:  ep.Widget.Desc,
			}
		}
		if c.Rcmd != nil {
			c.OgvCreativeId = c.Rcmd.CreativeId
		}
	case map[int64]*live.Room:
		rm := main.(map[int64]*live.Room)
		r, ok := rm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, op.ID)
		}
		if op.CardGoto == model.CardGotoAdInlineLive {
			c.AdInfo = op.AdInfo
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
		// 使用直播接口返回的url
		if r.Link != "" {
			c.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(c.Rcmd))
		}
		c.Avatar = avatarFrom(&AvatarStatus{Cover: r.Face, Text: r.Uname, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound})
		c.CoverLeftText1 = model.StatString(r.Online, "")
		c.CoverLeftIcon1 = model.IconOnline
		c.CoverLeft1ContentDescription = model.CoverIconContentDescription(c.CoverLeftIcon1, c.CoverLeftText1)
		c.CoverLeftText2 = r.AreaV2ParentName + " · " + r.AreaV2Name
		rcmdReasonText, _ = rcmdReason(c.Rcmd.RcmdReason, r.Uname, c.IsAttenm[r.UID], c.IsAttenm, c.Rcmd.Goto)
		c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "直播")
		c.CanPlay = 1
		c.Base.PlayerArgs = playerArgsFrom(r)
		c.Args.fromLiveRoom(r)
		c.Args.IsFollow = c.IsAttenm[r.UID]
		c.OfficialIcon = model.OfficialIcon(c.Cardm[r.UID])
		c.OfficialIconV2 = model.OfficialIcon(c.Cardm[r.UID])
		c.settingLiveThreePointPanelMeta(op)
		if c.IsAttenm[r.UID] == 1 {
			c.OfficialIcon = model.IconIsAttenm
			c.IsAtten = true
		}
		c.RightTopLiveBadge = ConstructRightTopLiveBadge(r.LiveStatus)
		c.SharePlane = &model.SharePlane{
			Title:      c.Title,
			Cover:      c.Cover,
			RoomId:     r.RoomID,
			ShareTo:    model.ShareTo,
			Author:     r.Uname,
			AuthorId:   r.UID,
			AreaName:   r.AreaV2Name,
			AuthorFace: r.Face,
		}
	default:
		log.Warn("LargeCoverInline From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReasonText, c.Base.Goto, op)
	c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	if op.Cover != "" {
		c.Cover = op.Cover
	}
	if op.ExtraURI != "" {
		c.ExtraURI = op.ExtraURI
	}
	c.Right = true
	if c.Base.PlayerArgs != nil {
		c.Base.PlayerArgs.ManualPlay = c.Rcmd.ManualInline()
		c.Base.PlayerArgs.HidePlayButton = constructHidePlayButton(c.Base.CardType)
		c.Base.PlayerArgs.ReportHistory = constructInlineReportHistory(c.Rcmd, c.Goto, op.Column)
		c.Base.PlayerArgs.ReportRequiredPlayDuration = constructInlineReportDuration(c.Goto)
		c.Base.PlayerArgs.ReportRequiredTime = constructInlineReportTime(c.Goto)
		return nil
	}
	c.Base.PlayerArgs = &PlayerArgs{
		ManualPlay:                 c.Rcmd.ManualInline(),
		HidePlayButton:             constructHidePlayButton(c.Base.CardType),
		ReportHistory:              constructInlineReportHistory(c.Rcmd, c.Goto, op.Column),
		ReportRequiredPlayDuration: constructInlineReportDuration(c.Goto),
		ReportRequiredTime:         constructInlineReportTime(c.Goto),
	}
	return nil
}

func constructInlineReportTime(gt model.Gt) int64 {
	switch gt {
	case model.GotoAv, model.GotoVerticalAv:
		return model.ReportRequiredTime
	case model.GotoPGC:
	case model.GotoLive:
	default:
	}
	return 0
}

func constructInlineReportDuration(gt model.Gt) int64 {
	switch gt {
	case model.GotoAv, model.GotoVerticalAv:
		return model.ReportRequiredPlayDuration
	case model.GotoPGC:
	case model.GotoLive:
	default:
	}
	return 0
}

func constructInlineReportHistory(rcmd *ai.Item, gt model.Gt, column model.ColumnStatus) int8 {
	if model.Columnm[column] == model.ColumnSvrSingle {
		return model.ReportHistory
	}
	switch gt {
	case model.GotoAv, model.GotoVerticalAv:
		return 1
	case model.GotoPGC:
	case model.GotoLive:
	default:
		log.Error("Failed to match inline goto: %s", gt)
	}
	return 0
}

func isSingleInline(rcmd *ai.Item) bool {
	//nolint:gosimple
	if rcmd.SingleInline > 0 {
		return true
	}
	return false
}

func constructHidePlayButton(cardType model.CardType) bool {
	if cardType == model.LargeCoverSingleV9 {
		return true
	}
	return model.HidePlayButton
}

func (c *LargeCoverInline) Get() *Base {
	return c.Base
}

func (c *LargeCoverInline) settingInlineIcon(op *operate.Card) {
	c.InlineProgressBar = &InlineProgressBar{
		IconDrag:     op.InlinePlayIcon.IconDrag,
		IconDragHash: op.InlinePlayIcon.IconDragHash,
		IconStop:     op.InlinePlayIcon.IconStop,
		IconStopHash: op.InlinePlayIcon.IconStopHash,
	}
}

func (c *LargeCoverInline) settingUGCThreePointPanelMeta(op *operate.Card) {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineUgcShareId  = "tm.recommend.ugc.0"
	)
	if op.InlineThreePoint.PanelType == 0 {
		return
	}
	c.ThreePointMeta = &threePointMeta.PanelMeta{
		PanelType:         int8(op.InlineThreePoint.PanelType),
		ShareOrigin:       _inlineShareOrigin,
		ShareId:           _inlineUgcShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(false, op.NeedSwitchColumnThreePoint, op.Column, op.ReplaceDislikeTitle),
	}
}

func (c *LargeCoverInline) settingOGVThreePointPanelMeta(op *operate.Card) {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineOgvShareId  = "tm.recommend.ogv.0"
	)
	if op.InlineThreePoint.PanelType == 0 {
		return
	}
	c.ThreePointMeta = &threePointMeta.PanelMeta{
		PanelType:         int8(op.InlineThreePoint.PanelType),
		ShareOrigin:       _inlineShareOrigin,
		ShareId:           _inlineOgvShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(true, op.NeedSwitchColumnThreePoint, op.Column, op.ReplaceDislikeTitle),
	}
}

func (c *LargeCoverInline) settingLiveThreePointPanelMeta(op *operate.Card) {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineLiveShareId = "tm.recommend.live.0"
	)
	if op.InlineThreePoint.PanelType == 0 {
		return
	}
	c.ThreePointMeta = &threePointMeta.PanelMeta{
		PanelType:         int8(op.InlineThreePoint.PanelType),
		ShareOrigin:       _inlineShareOrigin,
		ShareId:           _inlineLiveShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(true, op.NeedSwitchColumnThreePoint, op.Column, op.ReplaceDislikeTitle),
	}
}

func ConstructRightTopLiveBadge(liveStatus int8) *LiveBadge {
	const (
		InLiveText                 = "直播中"
		InLiveAnimation            = "https://i0.hdslb.com/bfs/archive/56ba9f2167c9c4e9353eb1a88ece5f7fae550ffc.json"
		InLiveAnimationHash        = "80c789eb839c986e801a638a65827ff3"
		InLiveBackgroundColorLight = "#FB7299"
		InLiveBackgroundColorNight = "#EB7093"
		InLiveAlphaLight           = 100
		InLiveAlphaNight           = 100
		InLiveFontColor            = "#FFFFFF"
	)
	return &LiveBadge{
		LiveStatus: liveStatus,
		InLive: &LiveBadgeResource{
			Text:                 InLiveText,
			AnimationURL:         InLiveAnimation,
			AnimationURLHash:     InLiveAnimationHash,
			BackgroundColorLight: InLiveBackgroundColorLight,
			BackgroundColorNight: InLiveBackgroundColorNight,
			AlphaLight:           InLiveAlphaLight,
			AlphaNight:           InLiveAlphaNight,
			FontColor:            InLiveFontColor,
		},
	}
}

type BigTunnelInline struct {
	Tunnel  *tunnelgrpc.FeedCard   `json:"tunnel"`
	Archive *arcgrpc.ArcPlayer     `json:"archive"`
	PGC     *pgcinline.EpisodeCard `json:"pgc"`
	Live    *live.Room             `json:"live"`
}

type NotifyTunnelItemV1 struct {
	Icon           string                 `json:"icon,omitempty"`
	Title          string                 `json:"title,omitempty"`
	TitleNight     string                 `json:"title_night,omitempty"`
	Subtitle       string                 `json:"subtitle,omitempty"`
	NotificationAt string                 `json:"notification_at,omitempty"`
	URI            string                 `json:"uri,omitempty"`
	Button         NotifyTunnelItemButton `json:"button,omitempty"`
	Param          string                 `json:"param,omitempty"`
	SubGoto        string                 `json:"sub_goto,omitempty"`
	EventType      string                 `json:"event_type,omitempty"`
	ObjectParam    string                 `json:"object_param,omitempty"`
	ObjectSubParam string                 `json:"object_sub_param,omitempty"`
	TitleRightText string                 `json:"title_right_text,omitempty"`
	TitleRightPic  model.Icon             `json:"title_right_pic,omitempty"`
	LiveBadge      *ReasonStyle           `json:"live_badge,omitempty"`
}

func (i *NotifyTunnelItemV1) FromTunnelCard(in *tunnelgrpc.FeedCard) {
	i.Icon = in.Cover
	i.Title = in.Title
	i.TitleNight = in.TitleNight
	i.Subtitle = in.Intro
	i.URI = in.Link
	i.Param = strconv.FormatInt(in.Oid, 10)
	i.SubGoto = in.SubGoTo
	i.EventType = in.EventType
	i.ObjectParam = strconv.FormatInt(in.Param, 10)
	i.ObjectSubParam = strconv.FormatInt(in.SubParam, 10)

	if in.StartTime > 0 {
		startTime := time.Unix(in.StartTime, 0)
		i.NotificationAt = model.PubDataString(startTime)
	}
	if in.TitleRightPic > 0 {
		if in.TitleRightPic > 99 {
			in.TitleRightPic = 99
		}
		i.TitleRightText = strconv.FormatInt(in.TitleRightPic, 10)
		i.TitleRightPic = model.IconTV
	}
	switch in.Badge {
	case model.TunnelBadgeLive:
		i.LiveBadge = &ReasonStyle{
			Text:             "直播中",
			TextColor:        "#FFFFFF",
			BgColor:          "#FB7299",
			BorderColor:      "#FB7299",
			IconURL:          "https://i0.hdslb.com/bfs/aistory/e4d020e11f675f536097c3bf78a5c7e807390799.gif",
			TextColorNight:   "#E5E5E5",
			BgColorNight:     "#BB5B76",
			BorderColorNight: "#BB5B76",
		}
	default:
	}
	if in.Button != nil {
		switch in.Button.Type {
		case "text":
			i.Button = &NotifyTunnelButtonText{
				Type: "text",
				Text: in.Button.Text,
				URI:  in.Button.Link,
			}
		case "game":
			i.Button = &NotifyTunnelButtonGame{
				Type:   "game_button",
				GameID: in.Button.GameId,
				Style:  matchTunnelButtonStyle(in.Button.Style),
			}
		}
	}
}

func matchTunnelButtonStyle(style tunnelcommon.ButtonStyle) int64 {
	const (
		_solid  = 0
		_hollow = 1
	)
	if style == tunnelcommon.ButtonStyleDefault || style == tunnelcommon.ButtonStyleSolid {
		return _solid
	}
	return _hollow
}

type NotifyTunnelItemButton interface {
	ButtonType() string
}

type NotifyTunnelButtonText struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

func (b NotifyTunnelButtonText) ButtonType() string {
	return b.Type
}

type NotifyTunnelButtonGame struct {
	Type   string `json:"type,omitempty"`
	GameID int64  `json:"game_id,omitempty"`
	Style  int64  `json:"style,omitempty"`
}

func (b NotifyTunnelButtonGame) ButtonType() string {
	return b.Type
}

// 单双列使用同一模型，仅 CardType 不同
type UniversalNotifyTunnelV1 struct {
	*Base
	Items []*NotifyTunnelItemV1 `json:"items,omitempty"`
}

func (c *UniversalNotifyTunnelV1) Get() *Base {
	return c.Base
}

func (c *UniversalNotifyTunnelV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*tunnelgrpc.FeedCard:
		tunnelCards := main.(map[int64]*tunnelgrpc.FeedCard)
		msgIDs, err := ConstructMsgIDs(c.Rcmd.MsgIDs)
		if err != nil {
			return newInvalidResourceErr(ResourceAI, 0, "MsgIDs: %q: %+v", c.Rcmd.MsgIDs, err)
		}
		c.Items = make([]*NotifyTunnelItemV1, 0, len(msgIDs))
		for _, oid := range msgIDs {
			tc, ok := tunnelCards[oid]
			if !ok {
				log.Warn("Failed to get tunnel card with id: %d", oid)
				continue
			}
			item := &NotifyTunnelItemV1{}
			item.FromTunnelCard(tc)
			c.Items = append(c.Items, item)
		}
		if len(c.Items) <= 0 {
			return newInvalidResourceErr(ResourceTunnelFeedCard, 0, "none tunnel card exist with msg_ids: %+v", msgIDs)
		}
	default:
		log.Warn("UniversalNotifyTunnelV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func ConstructMsgIDs(msgIDs string) ([]int64, error) {
	var out []int64
	slots := strings.Split(msgIDs, ",")
	for _, slot := range slots {
		oids := strings.Split(slot, "|")
		for _, oidStr := range oids {
			oid, err := strconv.ParseInt(oidStr, 10, 64)
			if err != nil {
				log.Error("Failed to parse msg id: %q, %+v", oidStr, err)
				return nil, err
			}
			out = append(out, oid)
		}
	}
	return out, nil
}

// 单双列使用同一模型，仅 CardType 不同
type UniversalNotifyTunnelLargeV1 struct {
	*Base
	Item *NotifyTunnelLargeItemV1 `json:"item,omitempty"`
}

type NotifyTunnelLargeItemV1 struct {
	NotifyTunnelItemV1
	LargeCover string       `json:"large_cover,omitempty"`
	Type       string       `json:"type,omitempty"`
	InlineAv   *EmbedInline `json:"inline_av,omitempty"`
	InlinePGC  *EmbedInline `json:"inline_pgc,omitempty"`
	InlineLive *EmbedInline `json:"inline_live,omitempty"`
}

func (i *NotifyTunnelLargeItemV1) FromTunnelCard(in *tunnelgrpc.FeedCard) {
	i.NotifyTunnelItemV1.FromTunnelCard(in)
	i.LargeCover = in.LargeImage
	i.Type = "image"
}

func (i *NotifyTunnelLargeItemV1) FromTunnelAndArchive(feedCard *tunnelgrpc.FeedCard, arc *arcgrpc.ArcPlayer, op *operate.Card) error {
	i.NotifyTunnelItemV1.FromTunnelCard(feedCard)
	if err := i.InlineAv.LargeCoverInlineFromArchive(arc, op); err != nil {
		return err
	}
	//ugc 默认不展示 主播头像、昵称、认证、关注状态
	TunnelHide(i.InlineAv)
	i.InlineAv.HideDanmuSwitch = model.TunnelHideDanmuSwitch
	i.InlineAv.DisableDanmu = model.TunnelDisableDanmu
	i.InlineAv.settingInlineIcon(op)
	i.Type = "inline_av"
	i.LargeCover = i.InlineAv.Cover
	uri, err := RawExtraURI(feedCard)
	if err != nil {
		log.Warn("Failed to raw extra uri: %+v", err)
		return nil
	}
	i.InlineAv.ExtraURI = uri
	return nil
}

func (i *NotifyTunnelLargeItemV1) FromTunnelAndPGC(feedCard *tunnelgrpc.FeedCard, pgc *pgcinline.EpisodeCard, op *operate.Card) error {
	i.NotifyTunnelItemV1.FromTunnelCard(feedCard)
	if err := i.InlinePGC.LargeCoverInlineFromPGC(pgc, op); err != nil {
		return err
	}
	TunnelHide(i.InlinePGC)
	i.InlinePGC.HideDanmuSwitch = model.TunnelHideDanmuSwitch
	i.InlinePGC.DisableDanmu = model.TunnelDisableDanmu
	i.InlinePGC.settingInlineIcon(op)
	i.Type = "inline_pgc"
	i.LargeCover = i.InlinePGC.Cover
	uri, err := RawExtraURI(feedCard)
	if err != nil {
		log.Warn("Failed to raw extra uri: %+v", err)
		return nil
	}
	i.InlinePGC.ExtraURI = uri
	return nil
}

func (i *NotifyTunnelLargeItemV1) FromTunnelAndLive(feedCard *tunnelgrpc.FeedCard, live *live.Room, op *operate.Card) error {
	i.NotifyTunnelItemV1.FromTunnelCard(feedCard)
	if err := i.InlineLive.LargeCoverInlineFromLive(live, op); err != nil {
		return err
	}
	//直播卡默认不展示 主播头像、昵称、认证、关注状态
	TunnelHide(i.InlineLive)
	i.InlineLive.HideDanmuSwitch = model.TunnelHideDanmuSwitch
	i.InlineLive.DisableDanmu = model.TunnelDisableDanmu
	i.Type = "inline_live"
	i.LargeCover = i.InlineLive.Cover
	uri, err := RawExtraURI(feedCard)
	if err != nil {
		log.Warn("Failed to raw extra uri: %+v", err)
		return nil
	}
	i.InlineLive.ExtraURI = uri
	return nil
}

func RawExtraURI(feedCard *tunnelgrpc.FeedCard) (string, error) {
	if feedCard.Resource == nil {
		return "", errors.Errorf("tunnel resource is nil: %d", feedCard.Oid)
	}
	if feedCard.Button == nil {
		return "", errors.Errorf("tunnel button is nil: %d", feedCard.Oid)
	}
	if feedCard.Button.Link == "" {
		return "", errors.Errorf("tunnel button link is nil: %d", feedCard.Oid)
	}
	if feedCard.Resource.Target == 2 { // target: 0未知  1播放页   2配置页
		return feedCard.Button.Link, nil
	}
	return "", errors.New("no extra uri")
}

func (c *UniversalNotifyTunnelLargeV1) Get() *Base {
	return c.Base
}

func (c *UniversalNotifyTunnelLargeV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case *tunnelgrpc.FeedCard:
		tunnelCard := main.(*tunnelgrpc.FeedCard)
		if tunnelCard == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		item := &NotifyTunnelLargeItemV1{}
		item.FromTunnelCard(tunnelCard)
		c.Item = item
	case *BigTunnelInline:
		bigTunnel := main.(*BigTunnelInline)
		if bigTunnel == nil || bigTunnel.Tunnel == nil {
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		bigTunnelObject := &ai.BigTunnelObject{}
		if err := json.Unmarshal([]byte(c.Rcmd.BigTunnelObject), &bigTunnelObject); err != nil {
			log.Error("Failed to unmarshal big tunnel object: %+v", errors.WithStack(err))
			return newResourceNotExistErr(ResourceTunnelFeedCard, 0)
		}
		inlineTunnelType := sets.NewString("ugc", "pgc", "live")
		if inlineTunnelType.Has(bigTunnelObject.Type) && !canEnableInlineTunnel(op.MobiApp, int64(op.Build)) {
			return newUnexpectedResourceTypeErr(bigTunnelObject, "ignore inline tunnel to unsupported device: %s, %d", op.MobiApp, op.Build)
		}
		base := new(Base)
		*base = *c.Base
		resourceID, _ := strconv.ParseInt(bigTunnelObject.Resource, 10, 64)
		switch bigTunnelObject.Type {
		case "image":
			item := &NotifyTunnelLargeItemV1{}
			item.FromTunnelCard(bigTunnel.Tunnel)
			c.Item = item
		case "ugc":
			if bigTunnel.Archive == nil {
				return newResourceNotExistErr(ResourceInlineAv, resourceID)
			}
			item := &NotifyTunnelLargeItemV1{
				InlineAv: &EmbedInline{
					LargeCoverInline: LargeCoverInline{
						Base: base,
					},
				},
			}
			if err := item.FromTunnelAndArchive(bigTunnel.Tunnel, bigTunnel.Archive, op); err != nil {
				return err
			}
			if !item.InlineAv.Right {
				return newInvalidResourceErr(ResourceInlineAv, bigTunnel.Archive.Arc.Aid, "inline av right is false")
			}
			c.Item = item
		case "pgc":
			if bigTunnel.PGC == nil {
				return newResourceNotExistErr(ResourceInlinePGC, resourceID)
			}
			item := &NotifyTunnelLargeItemV1{
				InlinePGC: &EmbedInline{
					LargeCoverInline: LargeCoverInline{
						Base: base,
					},
				},
			}
			if err := item.FromTunnelAndPGC(bigTunnel.Tunnel, bigTunnel.PGC, op); err != nil {
				return err
			}
			if !item.InlinePGC.Right {
				return newInvalidResourceErr(ResourceInlinePGC, bigTunnel.PGC.Aid, "inline pgc right is false")
			}
			c.Item = item
		case "live":
			if bigTunnel.Live == nil {
				return newResourceNotExistErr(ResourceInlineLive, resourceID)
			}
			item := &NotifyTunnelLargeItemV1{
				InlineLive: &EmbedInline{
					LargeCoverInline: LargeCoverInline{
						Base: base,
					},
				},
			}
			if err := item.FromTunnelAndLive(bigTunnel.Tunnel, bigTunnel.Live, op); err != nil {
				return err
			}
			if !item.InlineLive.Right {
				return newInvalidResourceErr(ResourceInlineLive, bigTunnel.Live.RoomID, "inline live right is false")
			}
			c.Item = item
		default:
			log.Warn("UniversalNotifyTunnelLargeV1 From: unexpected big tunnel object type %s", bigTunnelObject.Type)
			return newUnexpectedResourceTypeErr(main, "")
		}
	default:
		log.Warn("UniversalNotifyTunnelLargeV1 From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func canEnableInlineTunnel(mobiApp string, build int64) bool {
	return (mobiApp == "android" && build >= 6150000) ||
		(mobiApp == "iphone" && build >= 61500000)
}

type EmbedBannerInline struct {
	EmbedInline
	ServerType int    `json:"server_type"`
	ResourceID int    `json:"resource_id,omitempty"`
	Index      int    `json:"index,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	SrcID      int    `json:"src_id,omitempty"`
	IsAdLoc    bool   `json:"is_ad_loc,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	ID         int64  `json:"id"`
}

type EmbedInline struct {
	LargeCoverInline
}

func TunnelHide(i *EmbedInline) {
	i.Avatar = nil
	i.OfficialIcon = 0
	i.OfficialIconV2 = 0
	i.BadgeStyle = nil
	i.LeftCoverBadgeNewStyle = nil
}

func (i *EmbedInline) LargeCoverInlineFromArchive(arc *arcgrpc.ArcPlayer, op *operate.Card) error {
	i.CardType = ""
	i.CardGoto = ""
	if !model.AvIsNormalGRPC(arc) {
		return newInvalidResourceErr(ResourceArchive, arc.Arc.Aid, "av is not normal")
	}
	var furi func(uri string) string
	if arc.Arc.RedirectURL == "" {
		furi = model.ArcPlayHandler(arc.Arc, model.ArcPlayURL(arc, 0), i.Rcmd.TrackID, nil, op.Build, op.MobiApp, true)
		i.CanPlay = arc.Arc.Rights.Autoplay
		i.Base.PlayerArgs = playerArgsFrom(arc)
		if i.Base.PlayerArgs == nil {
			log.Warn("LargeCoverInline player card aid(%d) can't auto player", arc.Arc.Aid)
			return newInvalidResourceErr(ResourceArchive, arc.Arc.Aid, "av player args is nil")
		}
	}
	i.Base.from(op.Plat, op.Build, strconv.FormatInt(arc.Arc.Aid, 10), arc.Arc.Pic, arc.Arc.Title, model.GotoAv, strconv.FormatInt(arc.Arc.Aid, 10), furi)
	i.Goto = ""
	i.CoverLeftText1 = model.StatString(arc.Arc.Stat.View, "")
	i.CoverLeftIcon1 = model.IconPlay
	i.CoverLeftText2 = model.StatString(arc.Arc.Stat.Danmaku, "")
	i.CoverLeftIcon2 = model.IconDanmaku
	i.CoverRightText = model.DurationString(arc.Arc.Duration)
	i.Args.fromArchiveGRPC(arc.Arc, i.Tagm[op.Tid])
	i.Args.IsFollow = i.IsAttenm[arc.Arc.Author.Mid]
	i.OfficialIcon = model.OfficialIcon(i.Cardm[arc.Arc.Author.Mid])
	i.OfficialIconV2 = model.OfficialIcon(i.Cardm[arc.Arc.Author.Mid])
	if i.IsAttenm[arc.Arc.Author.Mid] == 1 {
		i.OfficialIcon = model.IconIsAttenm
		i.IsAtten = true
		i.Avatar = avatarFrom(&AvatarStatus{Cover: arc.Arc.Author.Face, Text: arc.Arc.Author.Name, Goto: model.GotoMid,
			Param: strconv.FormatInt(arc.Arc.Author.Mid, 10), Type: model.AvatarRound})
	}
	i.UpArgs = upArgsFrom(arc.Arc, i.IsAttenm[arc.Arc.Author.Mid])
	i.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, 0)
	if isFav, ok := op.HasFav[arc.Arc.Aid]; ok && isFav == 1 {
		i.IsFav = true
	}
	if isCoin, ok := op.HasCoin[arc.Arc.Aid]; ok && isCoin > 0 {
		i.IsCoin = true
	}
	i.LikeButton = likeButtonFromGRPC(arc.Arc, i.HasLike[arc.Arc.Aid], op)
	i.Right = true
	if i.Base.PlayerArgs != nil {
		i.Base.PlayerArgs.ManualPlay = i.Rcmd.ManualInline()
		i.Base.PlayerArgs.HidePlayButton = model.HidePlayButton
		i.Base.PlayerArgs.ReportHistory = model.ReportHistory
		i.Base.PlayerArgs.ReportRequiredPlayDuration = model.ReportRequiredPlayDuration
		i.Base.PlayerArgs.ReportRequiredTime = model.ReportRequiredTime
		return nil
	}
	i.Base.PlayerArgs = &PlayerArgs{
		ManualPlay:                 i.Rcmd.ManualInline(),
		HidePlayButton:             model.HidePlayButton,
		ReportHistory:              model.ReportHistory,
		ReportRequiredPlayDuration: model.ReportRequiredPlayDuration,
		ReportRequiredTime:         model.ReportRequiredTime,
	}
	return nil
}

func (i *EmbedInline) LargeCoverInlineFromPGC(ep *pgcinline.EpisodeCard, op *operate.Card) error {
	i.CardType = ""
	i.CardGoto = ""
	if ep.Season == nil {
		return newResourceNotExistErr(ResourceSeason, 0)
	}
	title := fmt.Sprintf("%s %s", ep.Season.Title, ep.NewDesc)
	i.Base.from(op.Plat, op.Build, strconv.FormatInt(int64(ep.EpisodeId), 10), ep.Cover, title, model.GotoPGC, "", nil)
	i.Goto = ""
	i.URI = model.FillURI("", 0, 0, ep.Url, model.PGCTrackIDHandler(i.Rcmd))
	if ep.PlayerInfo != nil {
		i.CanPlay = 1
	}
	if ep.Stat != nil {
		i.CoverLeftText1 = model.StatString(int32(ep.Stat.Play), "")
		i.CoverLeftIcon1 = model.IconPlay
		i.CoverLeftText2 = model.StatString(int32(ep.Stat.Follow), "")
		i.CoverLeftIcon2 = model.IconFavorite
	}
	i.CoverRightText = model.DurationString(ep.Duration)
	i.Base.PlayerArgs = playerArgsFrom(ep)
	i.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, ep.Season.TypeName)
	if ep.Widget != nil {
		i.PlayerWidget = &InlinePlayerWidget{
			Title: ep.Widget.Title,
			Desc:  ep.Widget.Desc,
		}
	}
	i.LikeButton = likeButtonFromEpisodeCard(ep, i.HasLike[ep.Aid], op)
	i.Right = true
	if i.Base.PlayerArgs != nil {
		i.Base.PlayerArgs.ManualPlay = i.Rcmd.ManualInline()
		i.Base.PlayerArgs.HidePlayButton = model.HidePlayButton
		return nil
	}
	i.Base.PlayerArgs = &PlayerArgs{
		ManualPlay:     i.Rcmd.ManualInline(),
		HidePlayButton: model.HidePlayButton,
	}
	return nil
}

func (i *EmbedInline) LargeCoverInlineFromLive(r *live.Room, op *operate.Card) error {
	i.CardType = ""
	i.CardGoto = ""
	i.Base.from(op.Plat, op.Build, strconv.FormatInt(r.RoomID, 10), r.Cover, r.Title, model.GotoLive, strconv.FormatInt(r.RoomID, 10), model.LiveRoomHandler(r, op.Network))
	i.Goto = ""
	// 使用直播接口返回的url
	if r.Link != "" {
		i.URI = model.FillURI("", 0, 0, r.Link, model.URLTrackIDHandler(i.Rcmd))
	}
	i.CoverLeftText1 = model.StatString(r.Online, "")
	i.CoverLeftIcon1 = model.IconOnline
	i.CoverLeftText2 = r.AreaV2Name
	i.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, "直播")
	i.CanPlay = 1
	i.Base.PlayerArgs = playerArgsFrom(r)
	i.Args.fromLiveRoom(r)
	i.Args.IsFollow = i.IsAttenm[r.UID]
	i.Avatar = avatarFrom(&AvatarStatus{Cover: r.Face, Text: r.Uname, Goto: model.GotoMid, Param: strconv.FormatInt(r.UID, 10), Type: model.AvatarRound})
	i.OfficialIcon = model.OfficialIcon(i.Cardm[r.UID])
	i.OfficialIconV2 = model.OfficialIcon(i.Cardm[r.UID])
	if i.IsAttenm[r.UID] == 1 {
		i.OfficialIcon = model.IconIsAttenm
		i.IsAtten = true
	}
	i.RightTopLiveBadge = ConstructRightTopLiveBadge(r.LiveStatus)
	i.Right = true
	if i.Base.PlayerArgs != nil {
		i.Base.PlayerArgs.ManualPlay = i.Rcmd.ManualInline()
		i.Base.PlayerArgs.HidePlayButton = model.HidePlayButton
		return nil
	}
	i.Base.PlayerArgs = &PlayerArgs{
		ManualPlay:     i.Rcmd.ManualInline(),
		HidePlayButton: model.HidePlayButton,
	}
	return nil
}

type BannerV8 struct {
	*Base
	Hash       string        `json:"hash,omitempty"`
	BannerItem []*BannerItem `json:"banner_item,omitempty"`
}

type BannerItem struct {
	Type         string             `json:"type"`
	ResourceID   int64              `json:"resource_id"`
	ID           int64              `json:"id"`
	Index        int64              `json:"index"`
	InlineAv     *EmbedBannerInline `json:"inline_av,omitempty"`
	InlinePGC    *EmbedBannerInline `json:"inline_pgc,omitempty"`
	InlineLive   *EmbedBannerInline `json:"inline_live,omitempty"`
	StaticBanner *banner.Banner     `json:"static_banner,omitempty"`
	AdBanner     *banner.Banner     `json:"ad_banner,omitempty"`
}

type BannerInline struct {
	Archive map[int64]*arcgrpc.ArcPlayer     `json:"archive"`
	PGC     map[int32]*pgcinline.EpisodeCard `json:"pgc"`
	Live    map[int64]*live.Room             `json:"live"`
}

const (
	BannerTypeAd         = "ad"
	BannerTypeAdInline   = "ad_inline"
	BannerTypeStatic     = "static"
	BannerTypeInlineAv   = "inline_av"
	BannerTypeInlinePGC  = "inline_pgc"
	BannerTypeInlineLive = "inline_live"
	InlineTypeAv         = "av"
	InlineTypePGC        = "pgc"
	InlineTypeLive       = "live"
)

func (c *BannerV8) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	//nolint:gosimple
	switch main.(type) {
	case *BannerInline:
		bannerInline := main.(*BannerInline)
		if bannerInline == nil {
			return newResourceNotExistErr(ResourceBannerInline, 0)
		}
		if c.Rcmd.BannerInfo == nil {
			return newInvalidResourceErr(ResourceAI, 0, "empty `BannerInfo`")
		}
		if len(op.Banner) == 0 {
			log.Warn("Banner len is null")
			return newInvalidResourceErr(ResourceOperateCard, 0, "empty `Banner`")
		}
		c.Hash = op.Hash
		c.BannerItem = make([]*BannerItem, 0, len(op.Banner))
		for _, v := range op.Banner {
			banner := &BannerItem{
				ResourceID: int64(v.ResourceID),
				ID:         v.ID,
				Index:      int64(v.Index),
			}
			switch {
			case IsAdBanner(v): // 广告ad
				banner.Type = BannerTypeAd
				banner.AdBanner = v
			case IsAdInlineBanner(v): // 广告inline
				banner.Type = BannerTypeAdInline
				// 低版本降级成ad类型
				if (op.MobiApp == "android" && op.Build < 6290000) || (op.MobiApp == "iphone" && op.Build < 62900000) {
					banner.Type = BannerTypeAd
				}
				banner.AdBanner = v
			case IsStaticBanner(v):
				banner.Type = BannerTypeStatic
				banner.StaticBanner = v
			case IsInlineBanner(v):
				if err := banner.FromInlineBanner(v, bannerInline, c.Base, op); err != nil {
					log.Error("Failed to construct FromInlineBanner: %+v", err)
					banner.FailbackToStatic(v)
				}
			default:
				log.Error("Unrecognized banner type: %+v", v)
				banner.FailbackToStatic(v)
			}
			c.BannerItem = append(c.BannerItem, banner)
		}
	default:
		log.Warn("Banner From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	c.Right = true
	return nil
}

func (banner *BannerItem) FailbackToStatic(v *banner.Banner) {
	banner.InlineAv = nil
	banner.InlinePGC = nil
	banner.InlineLive = nil
	banner.Type = BannerTypeStatic
	banner.StaticBanner = v
}

func (banner *BannerItem) FromInlineBanner(v *banner.Banner, bannerInline *BannerInline, bannerBase *Base, op *operate.Card) error {
	id, err := strconv.ParseInt(v.BannerMeta.InlineId, 10, 64)
	if err != nil {
		return errors.WithStack(err)
	}
	base := new(Base)
	*base = *bannerBase
	switch v.BannerMeta.InlineType {
	case InlineTypeAv:
		arc, ok := bannerInline.Archive[id]
		if !ok {
			return newResourceNotExistErr(ResourceArchive, id)
		}
		if err := banner.FromInlineAvBanner(arc, v, base, op); err != nil {
			return err
		}
		if !banner.InlineAv.Right {
			return newInvalidResourceErr(ResourceInlineAv, id, "inline av right is false")
		}
	case InlineTypePGC:
		pgc, ok := bannerInline.PGC[int32(id)]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, id)
		}
		if err := banner.FromInlinePGCBanner(pgc, v, base, op); err != nil {
			return err
		}
		if !banner.InlinePGC.Right {
			return newInvalidResourceErr(ResourceInlinePGC, id, "inline pgc right is false")
		}
	case InlineTypeLive:
		live, ok := bannerInline.Live[id]
		if !ok {
			return newResourceNotExistErr(ResourceRoom, id)
		}
		if err := banner.FromInlineLiveBanner(live, v, base, op); err != nil {
			return err
		}
		if !banner.InlineLive.Right {
			return newInvalidResourceErr(ResourceInlineLive, id, "inline live right is false")
		}
	default:
		return newUnexpectedResourceTypeErr(v.BannerMeta.InlineType, "")
	}
	return nil
}

func (banner *BannerItem) FromInlineAvBanner(arc *arcgrpc.ArcPlayer, v *banner.Banner, base *Base, op *operate.Card) error {
	banner.Type = BannerTypeInlineAv
	banner.InlineAv = &EmbedBannerInline{
		EmbedInline: EmbedInline{
			LargeCoverInline: LargeCoverInline{Base: base},
		},
	}
	if err := banner.InlineAv.LargeCoverInlineFromArchive(arc, op); err != nil {
		return err
	}
	//ugc 默认不展示 主播头像、昵称、认证、关注状态
	BannerHide(banner.InlineAv)
	SetBannerMeta(banner.InlineAv, v)
	banner.InlineAv.ExtraURI = ExtraURIFromBanner(v)
	banner.InlineAv.Title = v.Title
	banner.InlineAv.Cover = v.Image
	banner.InlineAv.HideDanmuSwitch = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	banner.InlineAv.DisableDanmu = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	return nil
}

func (banner *BannerItem) FromInlinePGCBanner(pgc *pgcinline.EpisodeCard, v *banner.Banner, base *Base, op *operate.Card) error {
	banner.Type = BannerTypeInlinePGC
	banner.InlinePGC = &EmbedBannerInline{
		EmbedInline: EmbedInline{
			LargeCoverInline: LargeCoverInline{Base: base},
		},
	}
	if err := banner.InlinePGC.LargeCoverInlineFromPGC(pgc, op); err != nil {
		return err
	}
	BannerHide(banner.InlinePGC)
	SetBannerMeta(banner.InlinePGC, v)
	banner.InlinePGC.ExtraURI = ExtraURIFromBanner(v)
	banner.InlinePGC.Title = v.Title
	banner.InlinePGC.Cover = v.Image
	banner.InlinePGC.HideDanmuSwitch = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	banner.InlinePGC.DisableDanmu = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	return nil
}

func (banner *BannerItem) FromInlineLiveBanner(live *live.Room, v *banner.Banner, base *Base, op *operate.Card) error {
	banner.Type = BannerTypeInlineLive
	banner.InlineLive = &EmbedBannerInline{
		EmbedInline: EmbedInline{
			LargeCoverInline: LargeCoverInline{Base: base},
		},
	}
	if err := banner.InlineLive.LargeCoverInlineFromLive(live, op); err != nil {
		return err
	}
	BannerHide(banner.InlineLive)
	SetBannerMeta(banner.InlineLive, v)
	banner.InlineLive.ExtraURI = ExtraURIFromBanner(v)
	banner.InlineLive.Title = v.Title
	banner.InlineLive.Cover = v.Image
	banner.InlineLive.HideDanmuSwitch = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	banner.InlineLive.DisableDanmu = AsBannerInlineDanmu(v.InlineBarrageSwitch)
	return nil
}

func ExtraURIFromBanner(banner *banner.Banner) string {
	if banner.BannerMeta.InlineId == "" || banner.InlineUseSame == 2 {
		return ""
	}
	return banner.URI
}

func (c *BannerV8) Get() *Base {
	return c.Base
}

func BannerHide(i *EmbedBannerInline) {
	i.Avatar = nil
	i.OfficialIcon = 0
	i.OfficialIconV2 = 0
	i.BadgeStyle = nil
	i.LeftCoverBadgeNewStyle = nil
	i.CoverLeftIcon1 = 0
	i.CoverLeftIcon2 = 0
	i.CoverLeftText1 = ""
	i.CoverLeftText2 = ""
	i.CoverRightText = ""
	i.PlayerArgs.ReportHistory = 0 // banner需禁止上报历史记录
	i.RcmdReasonStyle = nil
}

func SetBannerMeta(dst *EmbedBannerInline, v *banner.Banner) {
	dst.ServerType = v.ServerType
	dst.ResourceID = v.ResourceID
	dst.Index = v.Index
	dst.ClientIP = v.ClientIP
	dst.SrcID = v.SrcID
	dst.IsAdLoc = v.IsAdLoc
	dst.RequestID = v.RequestID
	dst.ID = v.ID
}

func IsAdBanner(arg *banner.Banner) bool {
	return arg.BannerMeta.Type == "AD_CPM_FRAME" || arg.BannerMeta.Type == "TOP_VIEW"
}

func IsAdInlineBanner(arg *banner.Banner) bool {
	return arg.BannerMeta.Type == "AD_CPM_INLINE"
}

func IsStaticBanner(arg *banner.Banner) bool {
	return arg.BannerMeta.InlineId == "" || arg.BannerMeta.InlineType == ""
}

func IsInlineBanner(arg *banner.Banner) bool {
	return arg.BannerMeta.InlineType != "" && arg.BannerMeta.InlineId != ""
}

type OgvSmallCover struct {
	*Base
	CoverLeftText1               string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1               model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverLeft1ContentDescription string       `json:"cover_left_1_content_description,omitempty"`
	CoverLeftText2               string       `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2               model.Icon   `json:"cover_left_icon_2,omitempty"`
	CoverLeft2ContentDescription string       `json:"cover_left_2_content_description,omitempty"`
	CoverRightText               string       `json:"cover_right_text,omitempty"`
	Subtitle                     string       `json:"subtitle,omitempty"`
	BadgeStyle                   *ReasonStyle `json:"badge_style,omitempty"`
	RcmdReasonStyle              *ReasonStyle `json:"rcmd_reason_style,omitempty"`
	DescButton                   *Button      `json:"desc_button,omitempty"`
	LeftCoverBadgeNewStyle       *ReasonStyle `json:"left_cover_badge_new_style,omitempty"`
	OgvCreativeId                int64        `json:"ogv_creative_id,omitempty"`
}

func (c *OgvSmallCover) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
	switch epCardm := main.(type) {
	case map[int64]*pgccard.EpisodeCard:
		ep, ok := epCardm[op.ID]
		if !ok {
			return newResourceNotExistErr(ResourceEpisode, op.ID)
		}
		if ep.Season == nil || ep.TianmaSmallCardMeta == nil {
			return newInvalidResourceErr(ResourceEpisode, int64(ep.EpisodeId), "Invalid episode card resource season(%+v) cardMeta(%+v)", ep.Season, ep.TianmaSmallCardMeta)
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(ep.EpisodeId)), ep.Cover, ep.TianmaSmallCardMeta.Title, model.GotoBangumi, strconv.Itoa(int(ep.EpisodeId)), nil)
		c.Subtitle = ep.TianmaSmallCardMeta.SubTitle
		if ep.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(ep.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText2 = model.StatString(int32(ep.Season.Stat.Follow), "")
			c.CoverLeftIcon2 = model.IconFavorite
		}
		c.URI = model.FillURI("", 0, 0, ep.Url, model.PGCTrackIDHandler(c.Rcmd))
		if ep.TianmaSmallCardMeta.BadgeInfo != nil {
			c.BadgeStyle = reasonStyleFrom(model.BgColorTransparentRed, ep.TianmaSmallCardMeta.BadgeInfo.Text)
		}
		// 受ai控制的字段
		if c.Rcmd != nil {
			if model.IsValidCover(c.Rcmd.CustomizedCover) {
				c.Cover = c.Rcmd.CustomizedCover
			}
			if c.Rcmd.CustomizedTitle != "" {
				c.Title = c.Rcmd.CustomizedTitle
			}
			if c.Rcmd.CustomizedSubtitle != "" {
				c.Subtitle = c.Rcmd.CustomizedSubtitle
			}
			if c.Rcmd.RcmdReason != nil {
				c.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Rcmd.RcmdReason.Content, c.Base.Goto, op)
			}
			c.LeftCoverBadgeNewStyle = iconBadgeStyleFrom(op, c.Rcmd.CornerMarkId)
			// 获取ogv评分信息，目前仅电影品类
			if c.Rcmd.OgvRightInfo == 1 && ep.Season.RatingInfo != nil && ep.Season.RatingInfo.Score > 0 && ep.Season.SeasonType == 2 {
				c.CoverRightText = fmt.Sprintf("%.1f分", ep.Season.RatingInfo.Score)
			}
			c.OgvCreativeId = c.Rcmd.CreativeId
		}
		if c.RcmdReasonStyle == nil {
			c.DescButton = buttonFrom(&ButtonStatus{Text: ep.TianmaSmallCardMeta.RcmdReason}, op.Plat)
		}
	default:
		log.Warn("OgvSmallCover From: unexpected type %T", main)
		return newUnexpectedResourceTypeErr(main, "")
	}
	// 新增物料
	if op.Title != "" {
		c.Title = op.Title
	}
	if op.Cover != "" {
		c.Cover = op.Cover
	}
	c.Right = true
	return nil
}

func (c *OgvSmallCover) Get() *Base {
	return c.Base
}

type ChannelSmallCoverV1 struct {
	*Base
	Bvid                      string        `json:"bvid,omitempty"`
	CoverLeftText1            string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1            model.Icon    `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2            string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2            model.Icon    `json:"cover_left_icon_2,omitempty"`
	CoverLeftText3            string        `json:"cover_left_text_3,omitempty"`
	LeftText1                 string        `json:"left_text_1,omitempty"`
	CoverRightBackgroundColor string        `json:"cover_right_background_color,omitempty"`
	Badge                     *ChannelBadge `json:"badge,omitempty"`
	Position                  int64         `json:"position,omitempty"`
	// 上报字段
	Sort string `json:"sort"`
	Filt int32  `json:"filt"`
}

func (c *ChannelSmallCoverV1) From(main interface{}, op *operate.Card) error {
	if op == nil {
		return newEmptyOPErr()
	}
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
		var (
			letfText1 string
			position  int64
		)
		if op.Channel != nil {
			// 频道角标逻辑
			if op.Channel.Badges != nil {
				if bdg, ok := op.Channel.Badges[op.ID]; ok {
					if bdg != nil && bdg.Text != "" && bdg.Cover != "" {
						c.Badge = &ChannelBadge{
							Text:    bdg.Text,
							BgCover: bdg.Cover,
						}
					}
				}
			}
			if op.Channel.IsFav != nil {
				if fav, ok := op.Channel.IsFav[op.ID]; ok && fav {
					letfText1 = fmt.Sprintf("已收藏·%s", model.DurationString(a.Arc.Duration))
				}
			}
			if letfText1 == "" {
				if op.Channel.Coins != nil {
					if coin, ok := op.Channel.Coins[op.ID]; ok && coin > 0 {
						letfText1 = fmt.Sprintf("已投币·%s", model.DurationString(a.Arc.Duration))
					}
				}
			}
			position = op.Channel.Position
		}
		position++
		c.Base.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.Param, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		c.Bvid, _ = bvid.AvToBv(op.ID)
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverLeftText3 = model.DurationString(a.Arc.Duration)
		c.LeftText1 = letfText1
		c.Sort = op.Channel.Sort
		c.Filt = op.Channel.Filt
		c.Position = position
		c.Base.Idx = position
	}
	c.Right = true
	return nil
}

func (c *ChannelSmallCoverV1) Get() *Base {
	return c.Base
}

func AsBannerInlineDanmu(inlineBarrageSwitch int64) bool {
	switch inlineBarrageSwitch {
	case 1:
		return true
	case 2:
		return false
	default:
		return true
	}
}

type CardQuality struct {
	Icon model.Icon
	Text string
}

func CastArchiveCustomizedQuality(in *ai.Item, arc *arcgrpc.ArcPlayer) []*CardQuality {
	const (
		// 1：播放；2：弹幕；3：点赞；4：评论
		_view    = 1
		_danmaku = 2
		_like    = 3
		_reply   = 4
	)
	store := []*CardQuality{}
	quality := in.CustomizedQuality
	if len(quality) <= 0 {
		return nil
	}
	if len(in.CustomizedQuality) > 2 {
		quality = in.CustomizedQuality[:2]
	}
	for _, v := range quality {
		switch v {
		case _view:
			store = append(store, &CardQuality{
				Icon: model.IconPlay,
				Text: model.StatString(arc.Arc.Stat.View, ""),
			})
		case _danmaku:
			store = append(store, &CardQuality{
				Icon: model.IconDanmaku,
				Text: model.StatString(arc.Arc.Stat.Danmaku, ""),
			})
		case _like:
			store = append(store, &CardQuality{
				Icon: model.IconLike,
				Text: model.StatString(arc.Arc.Stat.Like, ""),
			})
		case _reply:
			store = append(store, &CardQuality{
				Icon: model.IconComment,
				Text: model.StatString(arc.Arc.Stat.Reply, ""),
			})
		default:
			log.Warn("Unrecognized customized quality: %q", v)
		}
	}
	return store
}

func CastArchiveCustomizedDesc(in *ai.Item, arc *arcgrpc.ArcPlayer, tag *taggrpc.Tag) (*CustomizedButtonMeta, bool) {
	const (
		// 类型：1：分区；2：tag；3：up；4：投稿时间；5：自定义文案
		_descType   = 1
		_descTag    = 2
		_descUP     = 3
		_descUPTime = 4
		_descText   = 5

		// 1：tag；2：up空间 。 字段为0或不存在时，文案不支持跳转
		_linkTag     = 1
		_linkUPSpace = 2
	)
	if in.CustomizedDesc == nil {
		return nil, false
	}
	descList := []string{}
	for _, v := range in.CustomizedDesc.Desc {
		switch v.Type {
		case _descType:
			descList = append(descList, arc.Arc.TypeName)
		case _descTag:
			descList = append(descList, tag.Name)
		case _descUP:
			descList = append(descList, arc.Arc.Author.Name)
		case _descUPTime:
			descList = append(descList, model.PubDataString(arc.Arc.PubDate.Time()))
		case _descText:
			descList = append(descList, v.Text)
		default:
			log.Warn("Unrecognized desc: %+v", v)
		}
	}
	if len(descList) > 2 {
		return nil, false
	}
	descURI := ""
	switch in.CustomizedDesc.LinkType {
	case _linkTag:
		descURI = model.FillURI(model.GotoTag, 0, 0, strconv.FormatInt(tag.Id, 10), nil)
	case _linkUPSpace:
		descURI = model.FillURI(model.GotoMid, 0, 0, strconv.FormatInt(arc.Arc.Author.Mid, 10), nil)
	}
	cb := &CustomizedButtonMeta{
		Text: strings.Join(descList, " · "),
		URI:  descURI,
	}
	return cb, true
}

type SmallCoverV10 struct {
	*Base
	SubTitle                     string          `json:"sub_title,omitempty"`
	CoverGif                     string          `json:"cover_gif,omitempty"`
	CoverLeftText1               string          `json:"cover_left_text_1,omitempty"`
	CoverLeft1ContentDescription string          `json:"cover_left_1_content_description,omitempty"`
	CoverRightText               string          `json:"cover_right_text,omitempty"`
	CoverRightContentDescription string          `json:"cover_right_content_description,omitempty"`
	Avatar                       *Avatar         `json:"avatar,omitempty"`
	DescButton                   *Button         `json:"desc_button,omitempty"`
	BadgeStyle                   *ReasonStyle    `json:"badge_style,omitempty"`
	LeftCoverBadgeNewStyle       *ReasonStyle    `json:"left_cover_badge_new_style,omitempty"`
	GotoIcon                     *model.GotoIcon `json:"goto_icon,omitempty"`
	RcmdReasonStyle              *ReasonStyle    `json:"rcmd_reason_style,omitempty"`
}

func (c *SmallCoverV10) From(_ interface{}, _ *operate.Card) error {
	panic("Exception: card should build by ng")
}

func (c *SmallCoverV10) Get() *Base {
	return c.Base
}

type SmallCoverV11 struct {
	*Base
	FfCover                      string          `json:"ff_cover,omitempty"`
	RcmdReason                   string          `json:"rcmd_reason,omitempty"`
	DescButton                   *Button         `json:"desc_button,omitempty"`
	Desc                         string          `json:"desc,omitempty"`
	CanPlay                      int32           `json:"can_play,omitempty"`
	RcmdReasonStyle              *ReasonStyle    `json:"rcmd_reason_style,omitempty"`
	CoverLeftText1               string          `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1               model.Icon      `json:"cover_left_icon_1,omitempty"`
	CoverLeft1ContentDescription string          `json:"cover_left_1_content_description,omitempty"`
	CoverRightText               string          `json:"cover_right_text,omitempty"`
	CoverRightContentDescription string          `json:"cover_right_content_description,omitempty"`
	GotoIcon                     *model.GotoIcon `json:"goto_icon,omitempty"`
}

func (c *SmallCoverV11) From(_ interface{}, _ *operate.Card) error {
	panic("Exception: card should build by ng")
}

func (c *SmallCoverV11) Get() *Base {
	return c.Base
}

func enablePgcScore(op *operate.Card) bool {
	return op != nil &&
		((op.MobiApp == "android" && op.Build >= 6560000) || (op.MobiApp == "iphone" && op.Build >= 65600000))
}

type LargeCoverV11 struct {
	*Base
	Item []*NavItem `json:"item,omitempty"`
}

type NavItem struct {
	Name string `json:"name,omitempty"`
	Pic  string `json:"pic,omitempty"`
	URI  string `json:"uri,omitempty"`
}

func (c *LargeCoverV11) From(_ interface{}, _ *operate.Card) error {
	panic("Exception: card should build by ng")
}

func (c *LargeCoverV11) Get() *Base {
	return c.Base
}
