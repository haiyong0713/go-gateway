package v2

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"

	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	api "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
)

func singleHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, tagm map[int64]*cardm.Tag, isAttenm, hasLike map[int64]int8, statm map[int64]*relationgrpc.StatReply,
	cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply, adInfo *api.AdInfo) (hander Handler) {
	base := &api.Base{CardType: cardType, CardGoto: cardGoto}
	if rcmd != nil && rcmd.Idx == 5 { //热门只有第五张卡会是广告卡
		base.AdInfo = adInfo
	}
	card := &Card{
		Base:       base,
		CardCommon: &CardCommon{Rcmd: rcmd, Tagm: tagm, IsAttenm: isAttenm, HasLike: hasLike, Statm: statm, Cardm: cardm, Columnm: model.ColumnSvrSingle, AuthorRelations: authorRelations},
	}
	switch cardType {
	case model.LargeCoverV1:
		hander = &LargeCoverV1{Card: card, Item: &api.LargeCoverV1{Base: base}}
	case model.LargeCoverV4:
		hander = &LargeCoverV4{Card: card, Item: &api.LargeCoverV4{Base: base}}
	case model.SmallCoverV5:
		hander = &SmallCoverV5{Card: card, Item: &api.SmallCoverV5{Base: base}}
	case model.PopularTopEntrance:
		hander = &PopularTopEntrance{Card: card, Item: &api.PopularTopEntrance{Base: base}}
	case model.SmallCoverV5Ad:
		hander = &SmallCoverV5Ad{Card: card, Item: &api.SmallCoverV5Ad{Base: base}}
	default:
		//nolint:exhaustive
		switch cardGoto {
		case model.CardGotoAv, model.CardGotoBangumi, model.CardGotoLive, model.CardGotoPlayer, model.CardGotoPlayerLive, model.CardGotoChannelRcmd, model.CardGotoUpRcmdAv,
			model.CardGotoPGC, model.CardGotoPlayerBangumi, model.CardGotoAvConverge, model.CardGotoSpecialB:
			base.CardType = model.LargeCoverV1
			hander = &LargeCoverV1{Card: card, Item: &api.LargeCoverV1{Base: base}}
		case model.CardGotoUpRcmdSingle:
			base.CardType = model.RcmdOneItem
			hander = &RcmdOneItem{Card: card, Item: &api.RcmdOneItem{Base: base}}
		case model.CardGotoConverge, model.CardGotoRank, model.CardGotoConvergeAi:
			base.CardType = model.ThreeItemV1
			hander = &ThreeItemV1{Card: card, Item: &api.ThreeItemV1{Base: base}}
		case model.CardGotoEventTopic:
			base.CardType = model.MiddleCoverV3
			hander = &MiddleCoverV3{Card: card, Item: &api.MiddleCoverV3{Base: base}}
		case model.CardGotoReadCard:
			base.CardType = model.SmallCoverV5
			hander = &SmallCoverV5{Card: card, Item: &api.SmallCoverV5{Base: base}}
		}
	}
	return
}

type LargeCoverV1 struct {
	*Card
	Item *api.LargeCoverV1
}

//nolint:gocognit
func (c *LargeCoverV1) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	var (
		button    interface{}
		isBangumi bool
		upID      int64
		avatar    *cardm.AvatarStatus
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok || !model.AvIsNormalGRPC(a) {
			return
		}
		c.Card.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		if isBangumi = a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && op.RedirectURL != ""; isBangumi {
			c.Uri = op.RedirectURL
			c.Goto = model.GotoBangumi
			c.CardGoto = model.CardGotoBangumi
		}
		c.Item.CoverLeftText_1 = model.DurationString(a.Arc.Duration)
		c.Item.CoverLeftText_2 = model.ArchiveViewString(a.Arc.Stat.View)
		c.Item.CoverLeftText_3 = model.DanmakuString(a.Arc.Stat.Danmaku)
		if op.SwitchLike == model.SwitchFeedIndexLike {
			c.Item.CoverLeftText_2 = model.LikeString(a.Arc.Stat.Like)
			c.Item.CoverLeftText_3 = model.ArchiveViewString(a.Arc.Stat.View)
		}
		switch op.CardGoto {
		case model.CardGotoAv, model.CardGotoUpRcmdAv, model.CardGotoPlayer, model.CardGotoAvConverge:
			var (
				authorface = a.Arc.Author.Face
				authorname = a.Arc.Author.Name
			)
			if a.Arc.Author.Name != "" {
				if op.Switch != model.SwitchCooperationHide {
					authorname = unionAuthorGRPC(a)
				}
			}
			if (authorface == "" || authorname == "") && c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			avatar = &cardm.AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
			if c.Rcmd != nil && c.Rcmd.RcmdReason != nil && c.Rcmd.RcmdReason.Style == 3 && c.IsAttenm[a.Arc.Author.Mid] == 1 {
				c.Item.Desc = authorname
			} else {
				c.Item.Desc = authorname + " · " + model.PubDataString(a.Arc.PubDate.Time())
			}
			if op.CardGoto == model.CardGotoUpRcmdAv {
				button = &cardm.ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
			} else {
				if op.Channel != nil && op.Channel.ChannelID != 0 && op.Channel.ChannelName != "" {
					op.Channel.ChannelName = a.Arc.TypeName + " · " + op.Channel.ChannelName
					button = op.Channel
				} else if t, ok := c.Tagm[op.Tid]; ok {
					button = t
				} else {
					button = &cardm.ButtonStatus{Text: a.Arc.TypeName}
				}
				c.Card.maskFromGRPC(a.Arc)
			}
			if !isBangumi {
				c.Base.PlayerArgs = playerArgsFrom(a.Arc)
			}
			if op.CardGoto == model.CardGotoPlayer && c.Base.PlayerArgs == nil {
				log.Warn("player card aid(%d) can't auto player", a.Arc.Aid)
				return
			}
			c.Card.fromArgsArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			upID = a.Arc.Author.Mid
			//nolint:staticcheck
			if !model.IsPad(op.Plat) {
				// if c.HasLike != nil {
				// 	// c.LikeButton = likeButtonFromGRPC(a.Arc, c.HasLike[a.Aid], true)
				// }
			}
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
					if !ok || !model.AvIsNormalGRPC(ac) {
						return
					}
					urlf = model.ArcPlayHandler(ac.Arc, model.ArcPlayURL(ac, 0), op.TrackID, nil, op.Build, op.MobiApp, true)
					c.Base.PlayerArgs = playerArgsFrom(ac.Arc)
				case model.GotoAvConverge:
					c.Base.PlayerArgs = nil
					c.Param = strconv.FormatInt(c.Rcmd.ID, 10)
					urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build)
				default:
					urlf = model.TrackIDHandler(op.TrackID, c.Rcmd, op.Plat, op.Build)
				}
				c.Goto = jumpGoto
				c.Base.Uri = model.FillURI(jumpGoto, op.Plat, op.Build, op.Param, urlf)
				c.Card.fromArgsAvConverge(c.CardCommon.Rcmd)
			}
			c.Item.CoverGif = op.GifCover
		case model.CardGotoChannelRcmd:
			t, ok := c.Tagm[op.Tid]
			if !ok {
				return
			}
			avatar = &cardm.AvatarStatus{Cover: t.Cover, Goto: model.GotoTag, Param: strconv.FormatInt(t.ID, 10), Type: model.AvatarSquare}
			c.Item.Desc = model.SubscribeString(int32(t.Count.Atten))
			button = &cardm.ButtonStatus{Goto: model.GotoTag, Param: strconv.FormatInt(t.ID, 10), IsAtten: t.IsAtten}
			c.Base.PlayerArgs = playerArgsFrom(a.Arc)
			c.Card.maskFromGRPC(a.Arc)
			c.Card.fromArgsArchiveGRPC(a.Arc, c.Tagm[op.Tid])
		case model.CardGotoAdAv:
			c.Item.AdInfo = adInfoChangeProto(c.Rcmd.Ad)
			avatar = &cardm.AvatarStatus{Cover: a.Arc.Author.Face, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
			c.Item.Desc = a.Arc.Author.Name + " · " + model.PubDataString(a.Arc.PubDate.Time())
			button = c.Tagm[op.Tid]
			if (op.MobiApp == "iphone" && op.Build > 8430) || (op.MobiApp == "android" && op.Build > 5395000) {
				c.Base.PlayerArgs = playerArgsFrom(a.Arc)
				if c.Base.PlayerArgs == nil {
					log.Warn("player card ad aid(%d) can't auto player", a.Arc.Aid)
					return
				}
			}
			c.Card.fromArgsArchiveGRPC(a.Arc, c.Tagm[op.Tid])
			upID = a.Arc.Author.Mid
		default:
			log.Warn("LargeCoverV1 From: unexpected card_goto %s", op.CardGoto)
			return
		}
		if op.CardGoto != model.CardGotoAvConverge || op.Goto != model.GotoConverge {
			c.Item.CanPlay = a.Arc.Rights.Autoplay
		}
		if a.Arc.Rights.UGCPay == 1 && op.ShowUGCPay {
			c.Item.CoverBadge_2 = "付费"
		}
	default:
		log.Warn("LargeCoverV1 From: unexpected type %T", main)
		return
	}
	if (op.CardGoto != model.CardGotoSpecialB && c.Rcmd != nil) || (op.CardGoto == model.CardGotoSpecialB && c.Item.CoverBadge == "" && c.Rcmd != nil) {
		var isShowV2 bool
		c.Item.TopRcmdReason, c.Item.BottomRcmdReason = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[upID], c.IsAttenm, c.Rcmd.Goto)
		//nolint:exhaustive
		switch op.CardGoto {
		case model.CardGotoAvConverge:
			if isShowV2 = cardm.IsShowRcmdReasonStyleV2(c.Rcmd); isShowV2 {
				c.Item.Desc = ""
			}
		}
		if _, ok := op.SwitchStyle[model.SwitchFeedNewLive]; ok {
			if c.Item.BottomRcmdReason != "" {
				c.Item.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Item.BottomRcmdReason, c.Base.Goto, op)
			} else {
				c.Item.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Item.TopRcmdReason, c.Base.Goto, op)
			}
			c.Item.TopRcmdReason = ""
			c.Item.BottomRcmdReason = ""
		} else {
			if isShowV2 {
				c.Item.RcmdReasonStyleV2 = reasonStyleFromV2(c.Rcmd, c.Item.TopRcmdReason, c.Base.Goto, op.Plat, op.Build)
				c.Item.TopRcmdReason = ""
				c.Item.BottomRcmdReason = ""
			} else {
				c.Item.TopRcmdReasonStyle = topReasonStyleFrom(c.Rcmd, c.Item.TopRcmdReason, c.Base.Goto, op)
				c.Item.BottomRcmdReasonStyle = bottomReasonStyleFrom(c.Rcmd, c.Item.BottomRcmdReason, c.Base.Goto, op)
			}
		}
	}
	c.Item.OfficialIcon = model.OfficialIcon(c.Cardm[upID])
	if _, ok := op.SwitchStyle[model.SwitchFeedNewLive]; !ok {
		c.Item.Avatar = avatarFrom(avatar)
	}
	if c.Rcmd == nil || !c.Rcmd.HideButton {
		c.Item.DescButton = buttonFrom(button, op.Plat)
	}
	c.Right = true
}

func (c *LargeCoverV1) Get() *Card {
	return c.Card
}

type SmallCoverV5 struct {
	*Card
	Item *api.SmallCoverV5
}

// From is
//
//nolint:gocognit
func (c *SmallCoverV5) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	var (
		button     interface{}
		rcmdReason string
		avatar     *cardm.AvatarStatus
	)
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[op.ID]
		if !ok || !model.AvIsNormalGRPC(a) {
			return
		}
		c.Card.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		// TAPD https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001236037
		if (a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.Arc.RedirectURL != "") && ((op.Build > 9160 && op.Plat == model.PlatIPhone) || (op.Build > 5530500 && op.Plat == model.PlatAndroid)) {
			c.Uri = a.Arc.RedirectURL
			if op.TrackID != "" {
				//nolint:gosimple
				if strings.Index(c.Uri, "?") == -1 {
					c.Uri = fmt.Sprintf("%s?trackid=%s", c.Uri, op.TrackID)
				} else {
					c.Uri = fmt.Sprintf("%s&trackid=%s", c.Uri, op.TrackID)
				}
			}
		}
		c.Item.CoverRightText_1 = model.DurationString(a.Arc.Duration)
		c.Item.CoverRightTextContentDescription = model.DurationContentDescription(a.Arc.Duration)
		if c.Rcmd != nil {
			rcmdReason, _ = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[a.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.Item.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
		}
		switch op.CardGoto {
		case model.CardGotoAv:
			var (
				authorface = a.Arc.Author.Face
				authorname = a.Arc.Author.Name
			)
			if (authorface == "" || authorname == "") && c.Cardm != nil {
				if au, ok := c.Cardm[a.Arc.Author.Mid]; ok {
					authorface = au.Face
					authorname = au.Name
				}
			}
			switch c.Rcmd.Style {
			case model.HotCardStyleShowUp:
				c.Item.Up = &api.Up{
					Id:   a.Arc.Author.Mid,
					Name: authorname,
				}
				if stat, ok := c.Statm[a.Arc.Author.Mid]; ok {
					c.Item.Up.Desc = model.AttentionString(int32(stat.Follower))
				}
				avatar = &cardm.AvatarStatus{Cover: authorface, Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), Type: model.AvatarRound}
				c.Item.Up.Avatar = avatarFrom(avatar)
				c.Item.Up.OfficialIcon = model.OfficialIcon(c.Cardm[a.Arc.Author.Mid])
				c.Item.RightDesc_1 = model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.PubDataString(a.Arc.PubDate.Time())
				c.Item.RightIcon_1 = model.IconPlay
				button = &cardm.ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Arc.Author.Mid, 10), IsAtten: c.IsAttenm[a.Arc.Author.Mid]}
				c.Item.Up.DescButton = buttonFrom(button, op.Plat)
				if a.Arc.Rights.IsCooperation > 0 {
					c.Item.Up.Cooperation = "等联合创作"
				}
			default:
				c.Item.CoverGif = c.Rcmd.CoverGif
				if op.Switch != model.SwitchCooperationHide {
					if a.Arc.Author.Name != "" {
						c.Item.RightDesc_1 = unionAuthorGRPC(a)
						c.Item.RightIcon_1 = model.IconUp
					}
				} else {
					if authorname != "" {
						c.Item.RightDesc_1 = authorname
						c.Item.RightIcon_1 = model.IconUp
					}
				}
				c.Item.RightDesc_1ContentDescription = model.CoverIconContentDescription(c.Item.RightIcon_1,
					c.Item.RightDesc_1)
				// log warn
				if c.Item.RightDesc_1 == "" {
					prom.BusinessErrCount.Incr("card_av_name_err")
					log.Warn("CardGotoAv name not exist(%v,%d)", a.Arc.Author, a.Arc.Aid)
				}
				prom.BusinessErrCount.Incr("card_av_total")
				c.Item.RightDesc_2 = model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.PubDataString(a.Arc.PubDate.Time())
				c.Item.RightIcon_2 = model.IconPlay
			}
			if op.ShowHotword {
				func() {
					if op.SvideoShow { // 联播页
						c.Item.LeftCornerMarkStyle = &api.ReasonStyle{
							IconBgUrl: op.Cover,
							Text:      op.Subtitle,
						}
						c.Item.Uri = op.RedirectURL
						return
					}
					c.Item.HotwordEntrance = &api.HotwordEntrance{ // 热门热点
						HotwordId: op.Tid,
						Icon:      op.Cover,
						H5Url:     op.RedirectURL,
						HotText:   op.Subtitle,
					}
				}()
			}
			if (op.Build > 9040 && model.IsIPhone(op.Plat)) || (op.Build > 5510300 && op.MobiApp == "android") {
				aid, _ := strconv.ParseInt(c.Param, 10, 64)
				bvid, _ := cardm.GetBvIDStr(c.Param)
				shareSubtitle, playNumber := getShareSubtitle(a.Arc.Stat.View, a.Arc.Author.Name)
				c.Item.ThreePointV4 = &api.ThreePointV4{
					SharePlane: &api.SharePlane{
						Title:         c.Title,
						ShareSubtitle: shareSubtitle,
						Desc:          a.Arc.Desc,
						Cover:         c.Cover,
						Aid:           aid,
						Bvid:          bvid,
						ShareTo:       op.Share,
						Author:        a.Arc.Author.Name,
						AuthorId:      a.Arc.Author.Mid,
						ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%s", op.Param),
						PlayNumber:    playNumber,
						FirstCid:      a.Arc.FirstCid,
					},
					WatchLater: &api.WatchLater{
						Aid:  aid,
						Bvid: bvid,
					},
				}
			}
		default:
			log.Warn("SmallCoverV5 From: unexpected type %T", main)
			return
		}
	case map[int64]*livexroom.Infos:
		lv := main.(map[int64]*livexroom.Infos)
		l, ok := lv[op.RoomID]
		if !ok || l.Status.LiveStatus != 1 {
			return
		}
		var cover = l.Show.Cover
		if op.Cover != "" {
			cover = op.Cover
		}
		if l.Show != nil {
			c.Card.from(op.Plat, op.Build, strconv.FormatInt(l.RoomId, 10), cover, l.Show.Title, model.GotoLive, strconv.FormatInt(l.RoomId, 10), nil)
		}
		if c.Rcmd != nil {
			rcmdReason, _ = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[l.Uid], c.IsAttenm, c.Rcmd.Goto)
			c.Item.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
		}
		if l.Show != nil {
			c.Item.RightDesc_2 = model.LiveOnlineString(int32(l.Show.PopularityCount))
			c.Item.RightIcon_2 = model.IconOnline
		}
		c.Base.PlayerArgs = playerArgsFrom(l)
		c.Item.CornerMarkStyle = reasonStyleFrom(model.BgColorRed, "直播中")
		c.Item.CornerMarkStyle.LeftIconType = "liveIcon"
		au, ok := c.Cardm[l.Uid]
		if !ok {
			return
		}
		c.Item.RightDesc_1 = au.Name
		c.Item.RightIcon_1 = model.IconUp
		c.Base.Args = &api.Args{
			Type:   0,
			UpId:   l.Uid,
			UpName: au.Name,
			Rid:    int32(op.RoomID),
		}
		c.Base.CardGoto = model.CardGotoLive
		if (op.Build >= 9275 && model.IsIPhone(op.Plat)) || (op.Build >= 5560000 && op.MobiApp == "android") {
			c.Item.ThreePointV4 = &api.ThreePointV4{
				SharePlane: &api.SharePlane{
					Title:     c.Title,
					Cover:     c.Cover,
					Aid:       op.RoomID,
					ShareTo:   op.Share,
					Author:    au.Name,
					AuthorId:  l.Uid,
					ShortLink: c.Uri,
				},
				WatchLater: nil,
			}
			if l.Show != nil {
				c.Item.ThreePointV4.SharePlane.PlayNumber = model.LiveOnlineString(int32(l.Show.PopularityCount))
			}
			if l.Area != nil {
				c.Item.ThreePointV4.SharePlane.Desc = l.Area.AreaName
			}
		}
	case map[int64]*article.Meta:
		mm := main.(map[int64]*article.Meta)
		m, ok := mm[op.ID]
		if !ok {
			return
		}
		if len(m.ImageURLs) == 0 {
			return
		}
		cover := m.ImageURLs[0]
		if op.Cover != "" {
			cover = op.Cover
		}
		c.Card.from(op.Plat, op.Build, op.Param, cover, m.Title, model.GotoArticle, strconv.FormatInt(m.ID, 10), nil)
		if m.Author != nil {
			c.Item.RightDesc_1 = m.Author.Name
			c.Item.RightIcon_1 = model.IconUp
		}
		if m.Stats != nil {
			c.Item.RightDesc_2 = model.ArticleViewString(m.Stats.View) + " · " + model.PubDataString(m.PublishTime.Time())
			c.Item.RightIcon_2 = model.IconRead
		}
		if c.Rcmd != nil {
			rcmdReason, _ = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[m.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
			c.Item.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
		}
		c.Item.CornerMarkStyle = reasonStyleFrom(model.BgColorRed, "专栏")
		c.Base.Args = fromArticle(m)
	case map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList:
		lv := main.(map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList)
		l, ok := lv[op.RoomID]
		if !ok || l == nil || l.LiveStatus != 1 {
			return
		}
		var cover = l.Cover
		if op.Cover != "" {
			cover = op.Cover
		}
		if l.Title != "" {
			c.Card.from(op.Plat, op.Build, strconv.FormatInt(l.RoomId, 10), cover, l.Title, model.GotoLive, strconv.FormatInt(l.RoomId, 10), nil)
		}
		if c.Rcmd != nil {
			rcmdReason, _ = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[l.Uid], c.IsAttenm, c.Rcmd.Goto)
			c.Item.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
		}
		c.Item.RightDesc_2, c.Item.RightIcon_2 = func() (string, model.Icon) {
			if !enableLiveWatched(op.MobiApp, op.Build, op.Plat) {
				return model.LiveOnlineString(int32(l.PopularityCount)), model.IconOnline
			}
			if l.WatchedShow != nil && l.WatchedShow.Switch {
				return l.WatchedShow.TextLarge, model.IconLiveWatched
			}
			return model.LiveOnlineString(int32(l.PopularityCount)), model.IconLiveOnline
		}()
		c.Base.PlayerArgs = playerArgsFrom(l)
		c.Item.CornerMarkStyle = reasonStyleFrom(model.BgColorRed, "直播中")
		c.Item.CornerMarkStyle.LeftIconType = "liveIcon"
		au, ok := c.Cardm[l.Uid]
		if !ok {
			return
		}
		c.Item.RightDesc_1 = au.Name
		c.Item.RightIcon_1 = model.IconUp
		c.Base.Args = &api.Args{
			Type:   0,
			UpId:   l.Uid,
			UpName: au.Name,
			Rid:    int32(op.RoomID),
		}
		c.Base.CardGoto = model.CardGotoLive
		if (op.Build >= 9275 && model.IsIPhone(op.Plat)) || (op.Build >= 5560000 && op.MobiApp == "android") {
			c.Item.ThreePointV4 = &api.ThreePointV4{
				SharePlane: &api.SharePlane{
					Title:     c.Title,
					Cover:     c.Cover,
					Aid:       op.RoomID,
					ShareTo:   op.Share,
					Author:    au.Name,
					AuthorId:  l.Uid,
					ShortLink: c.Uri,
				},
				WatchLater: nil,
			}
			if l.WatchedShow == nil || !l.WatchedShow.Switch {
				c.Item.ThreePointV4.SharePlane.PlayNumber = model.LiveOnlineString(int32(l.PopularityCount))
			} else {
				c.Item.ThreePointV4.SharePlane.PlayNumber = l.WatchedShow.TextLarge
			}
			c.Item.ThreePointV4.SharePlane.Desc = l.AreaName
		}
	default:
		log.Warn("SmallCoverV5 From: unexpected type %T", main)
		return
	}
	c.Right = true
}

func (c *SmallCoverV5) Get() *Card {
	return c.Card
}

type SmallCoverV5Ad struct {
	*Card
	Item *api.SmallCoverV5Ad
}

func (c *SmallCoverV5Ad) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	var rcmdReason string
	arcPlayerMap, ok := main.(map[int64]*arcgrpc.ArcPlayer)
	if !ok {
		return
	}
	arcPlayer, ok := arcPlayerMap[op.ID]
	if !ok || !model.AvIsNormalGRPC(arcPlayer) {
		return
	}
	c.Card.from(op.Plat, op.Build, op.Param, arcPlayer.Arc.Pic, arcPlayer.Arc.Title, model.GotoAvAd, op.URI, model.ArcPlayHandler(arcPlayer.Arc, model.ArcPlayURL(arcPlayer, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
	c.Item.CoverRightText_1 = model.DurationString(arcPlayer.Arc.Duration)
	c.Item.CoverRightTextContentDescription = model.DurationContentDescription(arcPlayer.Arc.Duration)
	if c.Rcmd != nil {
		rcmdReason, _ = cardm.TopBottomRcmdReason(c.Rcmd.RcmdReason, c.IsAttenm[arcPlayer.Arc.Author.Mid], c.IsAttenm, c.Rcmd.Goto)
		c.Item.RcmdReasonStyle = topReasonStyleFrom(c.Rcmd, rcmdReason, c.Base.Goto, op)
	}
	var (
		authorface = arcPlayer.Arc.Author.Face
		authorname = arcPlayer.Arc.Author.Name
	)
	if (authorface == "" || authorname == "") && c.Cardm != nil {
		if au, ok := c.Cardm[arcPlayer.Arc.Author.Mid]; ok {
			authorface = au.Face
			authorname = au.Name
		}
	}
	c.Item.CoverGif = c.Rcmd.CoverGif
	if op.Switch != model.SwitchCooperationHide {
		if arcPlayer.Arc.Author.Name != "" {
			c.Item.RightDesc_1 = unionAuthorGRPC(arcPlayer)
			c.Item.RightIcon_1 = model.IconUp
		}
	} else {
		if authorname != "" {
			c.Item.RightDesc_1 = authorname
			c.Item.RightIcon_1 = model.IconUp
		}
	}
	c.Item.RightDesc_1ContentDescription = model.CoverIconContentDescription(c.Item.RightIcon_1,
		c.Item.RightDesc_1)
	if c.Item.RightDesc_1 == "" {
		log.Warn("CardGotoAv name not exist(%v,%d)", arcPlayer.Arc.Author, arcPlayer.Arc.Aid)
	}
	c.Item.RightDesc_2 = model.ArchiveViewString(arcPlayer.Arc.Stat.View) + " · " + model.PubDataString(arcPlayer.Arc.PubDate.Time())
	c.Item.RightIcon_2 = model.IconPlay
	if op.ShowHotword {
		func() {
			if op.SvideoShow { // 联播页
				c.Item.LeftCornerMarkStyle = &api.ReasonStyle{
					IconBgUrl: op.Cover,
					Text:      op.Subtitle,
				}
				c.Item.Uri = op.RedirectURL
				return
			}
			c.Item.HotwordEntrance = &api.HotwordEntrance{ // 热门热点
				HotwordId: op.Tid,
				Icon:      op.Cover,
				H5Url:     op.RedirectURL,
				HotText:   op.Subtitle,
			}
		}()
	}
	aid, _ := strconv.ParseInt(c.Param, 10, 64)
	bvid, _ := cardm.GetBvIDStr(c.Param)
	shareSubtitle, playNumber := getShareSubtitle(arcPlayer.Arc.Stat.View, arcPlayer.Arc.Author.Name)
	c.Item.ThreePointV4 = &api.ThreePointV4{
		SharePlane: &api.SharePlane{
			Title:         c.Title,
			ShareSubtitle: shareSubtitle,
			Desc:          arcPlayer.Arc.Desc,
			Cover:         c.Cover,
			Aid:           aid,
			Bvid:          bvid,
			ShareTo:       op.Share,
			Author:        arcPlayer.Arc.Author.Name,
			AuthorId:      arcPlayer.Arc.Author.Mid,
			ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%s", op.Param),
			PlayNumber:    playNumber,
			FirstCid:      arcPlayer.Arc.FirstCid,
		},
		WatchLater: &api.WatchLater{
			Aid:  aid,
			Bvid: bvid,
		},
	}
	c.Right = true
}

func (c *SmallCoverV5Ad) Get() *Card {
	return c.Card
}

type RcmdOneItem struct {
	*Card
	Item *api.RcmdOneItem
}

func (c *RcmdOneItem) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	au, ok := c.Cardm[op.ID]
	if !ok {
		return
	}
	c.Item.Base.CardGoto = model.CardGt(model.Gt(model.CardGotoUpRcmdNewV2))
	c.Card.from(op.Plat, op.Build, op.Param, au.Face, au.Name, model.GotoMid, op.Param, nil)
	button := &cardm.ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(op.ID, 10), IsAtten: c.IsAttenm[op.ID], Relation: model.RelationChange(op.ID, c.AuthorRelations)}
	c.Item.DescButton = buttonFrom(button, op.Plat)
	if op.Desc != "" {
		if (op.BuildLimit.IsAndroid && op.Build >= op.BuildLimit.HotCardOptimizeAndroid) ||
			(op.BuildLimit.IsIphone && op.Build >= op.BuildLimit.HotCardOptimizeIPhone) ||
			(op.BuildLimit.IsIPad && op.Build >= op.BuildLimit.HotCardOptimizeIPad) {
			c.Item.TopRcmdReasonStyle = reasonStyleFrom(model.BgColorLumpOrange, op.Desc)
		} else {
			c.Item.TopRcmdReasonStyle = reasonStyleFrom(model.BgColorOrange, op.Desc)
		}
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		item := op.Items[0]
		am := main.(map[int64]*arcgrpc.ArcPlayer)
		a, ok := am[item.ID]
		if !ok {
			return
		}
		c.Item.Item = &api.SmallCoverRcmdItem{
			Title:                            a.Arc.Title,
			Cover:                            a.Arc.Pic,
			CoverGif:                         c.Rcmd.CoverGif,
			Uri:                              model.FillURI(model.GotoAv, 0, 0, strconv.FormatInt(item.ID, 10), model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), "", nil, op.Build, op.MobiApp, false)),
			Param:                            strconv.FormatInt(item.ID, 10),
			Goto:                             string(model.GotoAv),
			CoverRightText_1:                 model.DurationString(a.Arc.Duration),
			CoverRightTextContentDescription: model.DurationContentDescription(a.Arc.Duration),
			RightDesc_1:                      a.Arc.Author.Name,
			RightIcon_1:                      model.IconUp,
			RightDesc_2:                      model.ArchiveViewString(a.Arc.Stat.View) + " · " + model.PubDataString(a.Arc.PubDate.Time()),
			RightIcon_2:                      model.IconPlay,
		}
		c.Item.Item.RightDesc_1ContentDescription = model.CoverIconContentDescription(c.Item.Item.RightIcon_1,
			c.Item.Item.RightDesc_1)
		aid := item.ID
		bvid, _ := cardm.GetBvIDStr(strconv.FormatInt(item.ID, 10))
		shareSubtitle, playNumber := getShareSubtitle(a.Arc.Stat.View, a.Arc.Author.Name)
		c.Item.ThreePointV4 = &api.ThreePointV4{
			SharePlane: &api.SharePlane{
				Title:         a.Arc.Title,
				ShareSubtitle: shareSubtitle,
				Desc:          a.Arc.Desc,
				Cover:         a.Arc.Pic,
				Aid:           aid,
				Bvid:          bvid,
				ShareTo:       op.Share,
				Author:        a.Arc.Author.Name,
				AuthorId:      a.Arc.Author.Mid,
				ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%d", item.ID),
				PlayNumber:    playNumber,
				FirstCid:      a.Arc.FirstCid,
			},
			WatchLater: &api.WatchLater{
				Aid:  aid,
				Bvid: bvid,
			},
		}
	default:
		log.Warn("RcmdOneItem From: unexpected type %T", main)
		return
	}
	c.Right = true
}

func (c *RcmdOneItem) Get() *Card {
	return c.Card
}

type ThreeItemV1 struct {
	*Card
	Item *api.ThreeItemV1
}

//nolint:gocognit
func (c *ThreeItemV1) From(main interface{}, op *operate.Card) {
	if c.Item == nil {
		return
	}
	switch main.(type) {
	case map[model.Gt]interface{}:
		intfcm := main.(map[model.Gt]interface{})
		if op == nil {
			return
		}
		switch op.CardGoto {
		case model.CardGotoRank:
			const (
				_title = "全站排行榜"
				_limit = 3
			)
			c.Card.from(op.Plat, op.Build, "0", "", _title, "", "", nil)
			// c.TitleIcon = model.IconRank
			c.Item.MoreUri = model.FillURI(op.Goto, 0, 0, op.URI, nil)
			c.Item.MoreText = "查看更多"
			c.Item.Items = make([]*api.ThreeItemV1Item, 0, _limit)
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
					item := &api.ThreeItemV1Item{
						CoverLeftText: model.DurationString(a.Arc.Duration),
						Desc_1:        model.ScoreString(v.Score),
						Base:          fromBase(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true)),
					}
					item.Args = fromArchiveGRPC(a.Arc, nil)
					c.Item.Items = append(c.Item.Items, item)
					if len(c.Item.Items) == _limit {
						break
					}
				}
			}
			if len(c.Item.Items) < _limit {
				return
			}
			c.Item.Items[0].CoverLeftIcon = model.IconGoldMedal
			c.Item.Items[1].CoverLeftIcon = model.IconSilverMedal
			c.Item.Items[2].CoverLeftIcon = model.IconBronzeMedal
		case model.CardGotoConverge, model.CardGotoConvergeAi:
			limit := 3
			if op.Coverm[c.Columnm] != "" {
				limit = 2
			}
			c.Card.from(op.Plat, op.Build, op.Param, op.Coverm[c.Columnm], op.Title, op.Goto, op.URI, nil)
			c.Item.MoreUri = model.FillURI(model.GotoConverge, 0, 0, op.Param, nil)
			c.Item.MoreText = "查看更多"
			//nolint:exhaustive
			switch op.CardGoto {
			case model.CardGotoConvergeAi:
				limit = 2
				c.Base.CardGoto = model.CardGotoConverge
				if len(op.Items) <= limit {
					c.Item.MoreText = ""
					c.Item.MoreUri = ""
				}
			}
			c.Item.Items = make([]*api.ThreeItemV1Item, 0, len(op.Items))
			for _, v := range op.Items {
				if v == nil {
					continue
				}
				intfc, ok := intfcm[v.Goto]
				if !ok {
					continue
				}
				var item *api.ThreeItemV1Item
				//nolint:gosimple
				switch intfc.(type) {
				case map[int64]*arcgrpc.ArcPlayer:
					am := intfc.(map[int64]*arcgrpc.ArcPlayer)
					a, ok := am[v.ID]
					if !ok || !model.AvIsNormalGRPC(a) {
						continue
					}
					item = &api.ThreeItemV1Item{
						CoverLeftText: model.DurationString(a.Arc.Duration),
						Desc_1:        model.ArchiveViewString(a.Arc.Stat.View),
						Desc_2:        model.DanmakuString(a.Arc.Stat.Danmaku),
						Base:          fromBase(op.Plat, op.Build, v.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, v.URI, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true)),
					}
					if op.SwitchLike == model.SwitchFeedIndexLike {
						item.Desc_1 = model.LikeString(a.Arc.Stat.Like)
						item.Desc_2 = model.ArchiveViewString(a.Arc.Stat.View)
					}
					item.Args = fromArchiveGRPC(a.Arc, nil)
				case map[int64]*live.Room:
					rm := intfc.(map[int64]*live.Room)
					r, ok := rm[v.ID]
					if !ok || r.LiveStatus != 1 {
						continue
					}
					item = &api.ThreeItemV1Item{
						Desc_1: model.LiveOnlineString(r.Online),
						Badge:  "直播",
						Base:   fromBase(op.Plat, op.Build, v.Param, r.Cover, r.Title, model.GotoLive, v.URI, model.LiveRoomHandler(r, op.Network)),
					}
					item.Args = fromLiveRoom(r)
				case map[int64]*article.Meta:
					mm := intfc.(map[int64]*article.Meta)
					m, ok := mm[v.ID]
					if !ok {
						continue
					}
					if len(m.ImageURLs) == 0 {
						continue
					}
					item = &api.ThreeItemV1Item{
						Badge: "文章",
						Base:  fromBase(op.Plat, op.Build, v.Param, m.ImageURLs[0], m.Title, model.GotoArticle, v.URI, nil),
					}
					if m.Stats != nil {
						item.Desc_1 = model.ArticleViewString(m.Stats.View)
						item.Desc_2 = model.ArticleReplyString(m.Stats.Reply)
					}
					item.Args = fromArticle(m)
				default:
					log.Warn("ThreeItemV1 From: unexpected type %T", intfc)
					continue
				}
				c.Item.Items = append(c.Item.Items, item)
				if len(c.Item.Items) == limit {
					break
				}
			}
			if len(c.Item.Items) < limit {
				return
			}
		default:
			log.Warn("ThreeItemV1 From: unexpected card_goto %s", op.CardGoto)
			return
		}
	default:
		log.Warn("ThreeItemV1 From: unexpected type %T", main)
		return
	}
	c.Right = true
}

func (c *ThreeItemV1) Get() *Card {
	return c.Card
}

type MiddleCoverV3 struct {
	*Card
	Item *api.MiddleCoverV3
}

func (c *MiddleCoverV3) From(main interface{}, op *operate.Card) {
	if op == nil {
		return
	}
	c.Card.from(op.Plat, op.Build, op.Param, op.Cover, op.Title, model.GotoWeb, op.URI, nil)
	c.Item.Goto = op.Goto
	if op.Badge != "" {
		c.Item.CoverBadgeStyle = reasonStyleFrom(model.BgColorPurple, op.Badge)
	}
	c.Item.Desc1 = op.Desc
	c.Right = true
}

func (c *MiddleCoverV3) Get() *Card {
	return c.Card
}

type LargeCoverV4 struct {
	*Card
	Item *api.LargeCoverV4
}

func (c *LargeCoverV4) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	rid, err := strconv.ParseInt(op.SubParam, 10, 64)
	if err != nil {
		return
	}
	//nolint:gosimple
	switch main.(type) {
	case map[int64]*arcgrpc.
		ArcPlayer:
		am := main.(map[int64]*arcgrpc.ArcPlayer)

		a, ok := am[rid]
		if !ok || !model.AvIsNormalGRPC(a) {
			return
		}
		c.Card.from(op.Plat, op.Build, op.Param, a.Arc.Pic, a.Arc.Title, model.GotoAv, op.SubParam, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), op.TrackID, nil, op.Build, op.MobiApp, true))
		c.Item.CoverLeftText_1 = model.DurationString(a.Arc.Duration)
		c.Item.CoverLeftText_2 = model.ArchiveViewString(a.Arc.Stat.View)
		c.Item.CoverLeftText_3 = model.DanmakuString(a.Arc.Stat.Danmaku)
		c.Base.PlayerArgs = playerArgsFrom(a.Arc)
		c.Item.Up = &api.Up{
			Id:   a.Arc.Author.Mid,
			Name: a.Arc.Author.Name,
		}
		c.Item.ShareSubtitle, c.Item.PlayNumber = getShareSubtitle(a.Arc.Stat.View, a.Arc.Author.Name)
		c.Base.Args = fromArchiveGRPC(a.Arc, nil)
	}
	c.Item.SubParam = op.SubParam
	c.CardGoto = model.CardGotoHotAV
	c.Goto = model.GotoHotPlayerAv
	c.Title = op.Desc
	if op.CanPlay {
		c.Item.CanPlay = 1
	}
	c.Item.Bvid, _ = cardm.GetBvIDStr(op.SubParam)
	c.Right = true
}

func getShareSubtitle(view int32, authorName string) (shareSubtitle, playNumber string) {
	tmp := strconv.FormatFloat(float64(view)/10000, 'f', 1, 64)
	shareSubtitle = fmt.Sprintf("UP主：%s", authorName)
	if view > 100000 {
		shareSubtitle = fmt.Sprintf("%s\n%s万播放", shareSubtitle, strings.TrimSuffix(tmp, ".0"))
	}
	playNumber = strings.TrimSuffix(tmp, ".0") + "万次"
	return
}

func (c *LargeCoverV4) Get() *Card {
	return c.Card
}

type PopularTopEntrance struct {
	*Card
	Item *api.PopularTopEntrance
}

func (c *PopularTopEntrance) From(main interface{}, op *operate.Card) {
	if op == nil || c.Item == nil {
		return
	}
	c.Card.from(op.Plat, op.Build, op.Param, "", op.Title, model.GotoAv, op.URI, nil)
	for _, item := range op.EntranceItems {
		c.Item.Items = append(c.Item.Items, &api.EntranceItem{
			Goto:         item.Goto,
			Icon:         item.Icon,
			Title:        item.Title,
			ModuleId:     item.ModuleId,
			Uri:          item.Uri,
			EntranceId:   item.EntranceId,
			EntranceType: item.EntranceType,
			Bubble: &api.Bubble{
				BubbleContent: item.Bubble.BubbleContent,
				Version:       item.Bubble.Version,
				Stime:         item.Bubble.Stime,
			},
		})
	}
	c.Title = op.Desc
}

func (c *PopularTopEntrance) Get() *Card {
	return c.Card
}

func enableLiveWatched(mobiApp string, build int, plat int8) bool {
	return (mobiApp == "android" && build >= 6610000) ||
		(mobiApp == "iphone" && model.IsIPhone(plat) && build >= 66600000) ||
		(mobiApp == "iphone" && model.IsIPad(plat) && build >= 66200000) ||
		(mobiApp == "ipad" && build >= 33600000)
}
