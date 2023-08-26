package v2

import (
	"strconv"

	"go-common/library/log"

	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	api "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
)

// Handler is
type Card struct {
	*CardCommon
	*api.Base
}

type CardCommon struct {
	Right           bool                                       `json:"-"`
	Rcmd            *ai.Item                                   `json:"-"`
	Tagm            map[int64]*cardm.Tag                       `json:"-"`
	IsAttenm        map[int64]int8                             `json:"-"`
	HasLike         map[int64]int8                             `json:"-"`
	Statm           map[int64]*relationgrpc.StatReply          `json:"-"`
	Cardm           map[int64]*accountgrpc.Card                `json:"-"`
	CardLen         int                                        `json:"-"`
	Columnm         model.ColumnStatus                         `json:"-"`
	AuthorRelations map[int64]*relationgrpc.InterrelationReply `json:"-"`
}

type Handler interface {
	From(main interface{}, op *operate.Card)
	Get() *Card
}

// Handle is
func Handle(plat int8, cardGoto model.CardGt, cardType model.CardType, column model.ColumnStatus, rcmd *ai.Item, tagm map[int64]*cardm.Tag, isAttenm, hasLike map[int64]int8,
	statm map[int64]*relationgrpc.StatReply, cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply, adInfo *api.AdInfo) (hander Handler) {
	if model.IsPad(plat) {
		return ipadHandle(cardGoto, cardType, rcmd, nil, isAttenm, nil, statm, cardm, authorRelations)
	}
	switch column {
	case model.ColumnSvrSingle, model.ColumnUserSingle:
		return singleHandle(cardGoto, cardType, rcmd, tagm, isAttenm, hasLike, statm, cardm, authorRelations, adInfo)
	default:
		return nil
	}
}

func AddCard(card interface{}) (res *api.Card) {
	//nolint:gosimple
	switch card.(type) {
	case *SmallCoverV5:
		res = &api.Card{
			Item: &api.Card_SmallCoverV5{
				SmallCoverV5: card.(*SmallCoverV5).Item,
			},
		}
	case *SmallCoverV5Ad:
		res = &api.Card{
			Item: &api.Card_SmallCoverV5Ad{
				SmallCoverV5Ad: card.(*SmallCoverV5Ad).Item,
			},
		}
	case *LargeCoverV1:
		res = &api.Card{
			Item: &api.Card_LargeCoverV1{
				LargeCoverV1: card.(*LargeCoverV1).Item,
			},
		}
	case *ThreeItemV1:
		res = &api.Card{
			Item: &api.Card_ThreeItemV1{
				ThreeItemV1: card.(*ThreeItemV1).Item,
			},
		}
	case *MiddleCoverV3:
		res = &api.Card{
			Item: &api.Card_MiddleCoverV3{
				MiddleCoverV3: card.(*MiddleCoverV3).Item,
			},
		}
	case *LargeCoverV4:
		res = &api.Card{
			Item: &api.Card_LargeCoverV4{
				LargeCoverV4: card.(*LargeCoverV4).Item,
			},
		}
	case *PopularTopEntrance:
		res = &api.Card{
			Item: &api.Card_PopularTopEntrance{
				PopularTopEntrance: card.(*PopularTopEntrance).Item,
			},
		}
	case *RcmdOneItem:
		res = &api.Card{
			Item: &api.Card_RcmdOneItem{
				RcmdOneItem: card.(*RcmdOneItem).Item,
			},
		}
	}
	return
}

func (c *Card) from(plat int8, build int, param, cover, title string, gt model.Gt, uri string, f func(uri string) string) {
	if c == nil || c.Base == nil {
		return
	}
	c.Base.Uri = model.FillURI(gt, plat, build, uri, f)
	c.Base.Cover = cover
	c.Base.Title = title
	if gt != "" {
		c.Base.Goto = gt
	} else {
		c.Base.Goto = model.Gt(c.Base.CardGoto)
	}
	c.Base.Param = param
}

func fromBase(plat int8, build int, param, cover, title string, gt model.Gt, uri string, f func(uri string) string) (b *api.Base) {
	b = &api.Base{
		Uri:   model.FillURI(gt, plat, build, uri, f),
		Cover: cover,
		Title: title,
		Param: param,
		Goto:  gt,
	}
	return
}

func (c *Card) maskFromGRPC(a *arcgrpc.Arc) {
	if a == nil {
		return
	}
	c.Mask = &api.Mask{}
	c.Mask.Avatar = avatarFrom(&cardm.AvatarStatus{Cover: a.Author.Face, Text: a.Author.Name, Goto: model.GotoMid, Param: strconv.FormatInt(a.Author.Mid, 10), Type: model.AvatarRound})
	c.Mask.Button = buttonFrom(&cardm.ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Author.Mid, 10), IsAtten: c.IsAttenm[a.Author.Mid]}, 0)
}

func topReasonStyleFrom(rcmd *ai.Item, text string, _ model.Gt, op *operate.Card) (res *api.ReasonStyle) {
	if text == "" || rcmd == nil {
		return
	}
	var (
		style, bgstyle int8
	)
	if style = rcmd.CornerMark; style == 0 {
		if rcmd.RcmdReason != nil {
			if rcmd.RcmdReason.Content == "" {
				style = 0
			} else {
				style = rcmd.RcmdReason.CornerMark
			}
		}
	}
	if op != nil {
		if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
			style = 5
		}
	}
	switch style {
	case 0, 2:
		bgstyle = model.BgColorOrange
	case 1:
		bgstyle = model.BgColorTransparentOrange
	case 3:
		bgstyle = model.BgTransparentTextOrange
	case 4:
		bgstyle = model.BgColorRed
	case 5:
		bgstyle = model.BgColorFillingOrange
	case 6:
		bgstyle = model.BgColorLumpOrange
	case 7:
		bgstyle = model.BgColorContourOrange
	default:
		bgstyle = model.BgColorOrange
	}
	res = reasonStyleFrom(bgstyle, text)
	return
}

func playerArgsFrom(v interface{}) (playerArgs *api.PlayerArgs) {
	//nolint:gosimple
	switch v.(type) {
	case *arcgrpc.Arc:
		a := v.(*arcgrpc.Arc)
		if a == nil || (a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && a.Rights.Autoplay != 1) || (a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes) {
			return
		}
		playerArgs = &api.PlayerArgs{Aid: a.Aid, Cid: a.FirstCid, Type: model.GotoAv}
	case *live.Room:
		r := v.(*live.Room)
		if r == nil || r.LiveStatus != 1 {
			return
		}
		playerArgs = &api.PlayerArgs{RoomId: r.RoomID, IsLive: 1, Type: model.GotoLive}
	case *livexroom.Infos:
		r := v.(*livexroom.Infos)
		if r == nil || r.Status.LiveStatus != 1 {
			return
		}
		playerArgs = &api.PlayerArgs{RoomId: r.RoomId, IsLive: 1, Type: model.GotoLive}
	case *roomgategrpc.EntryRoomInfoResp_EntryList:
		r := v.(*roomgategrpc.EntryRoomInfoResp_EntryList)
		if r == nil || r.LiveStatus != 1 {
			return
		}
		playerArgs = &api.PlayerArgs{RoomId: r.RoomId, IsLive: 1, Type: model.GotoLive}
	case *bangumi.EpPlayer:
		ep := v.(*bangumi.EpPlayer)
		if ep == nil {
			return
		}
		playerArgs = &api.PlayerArgs{Aid: ep.AID, Cid: ep.CID, EpId: ep.EpID, IsPreview: ep.IsPreview, Type: model.GotoBangumi, Duration: ep.Duration, SubType: ep.Season.Type, SeasonId: ep.Season.SeasonID}
	case nil:
	default:
		log.Warn("playerArgsFrom: unexpected type %T", v)
	}
	return
}

func (c *Card) fromArgsArchiveGRPC(a *arcgrpc.Arc, t *cardm.Tag) {
	if c.Args == nil {
		c.Args = &api.Args{}
	}
	if a != nil {
		c.Args.Aid = a.Aid
		c.Args.UpId = a.Author.Mid
		c.Args.UpName = a.Author.Name
		c.Args.Rid = a.TypeID
		c.Args.Rname = a.TypeName
	}
	if t != nil {
		c.Args.Tid = t.ID
		c.Args.Tname = t.Name
	}
}

func fromArchiveGRPC(a *arcgrpc.Arc, t *cardm.Tag) (args *api.Args) {
	args = &api.Args{}
	if a != nil {
		args.Aid = a.Aid
		args.UpId = a.Author.Mid
		args.UpName = a.Author.Name
		args.Rid = a.TypeID
		args.Rname = a.TypeName
	}
	if t != nil {
		args.Tid = t.ID
		args.Tname = t.Name
	}
	return
}

func (c *Card) fromArgsAvConverge(r *ai.Item) {
	if r == nil {
		return
	}
	if c.Args == nil {
		c.Args = &api.Args{}
	}
	var (
		_stateAv       = "typeA"
		_stateConverge = "typeB"
	)
	c.Args.TrackId = r.TrackID
	if r.ConvergeInfo != nil {
		c.Args.ConvergeType = r.ConvergeInfo.ConvergeType
	}
	//nolint:exhaustive
	switch model.Gt(r.JumpGoto) {
	case model.GotoAv:
		c.Args.State = _stateAv
	case model.GotoConverge:
		c.Args.State = _stateConverge
	}
}

func reasonStyleFrom(style int8, text string) (res *api.ReasonStyle) {
	res = &api.ReasonStyle{
		Text: text,
	}
	switch style {
	case model.BgColorOrange: //defalut
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FFFB9E60"
		res.BorderColor = "#FFFB9E60"
		res.BgStyle = int32(model.BgStyleFill)
	case model.BgColorTransparentOrange:
		res.TextColor = "#FFFB9E60"
		res.BorderColor = "#FFFB9E60"
		res.BgStyle = int32(model.BgStyleStroke)
	case model.BgColorBlue:
		res.TextColor = "#FF23ADE5"
		res.BgColor = "#3323ADE5"
		res.BorderColor = "#3323ADE5"
		res.BgStyle = int32(model.BgStyleFill)
	case model.BgColorRed:
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FFFB7299"
		res.BorderColor = "#FFFB7299"
		res.BgStyle = int32(model.BgStyleFill)
		// 夜间
		res.TextColorNight = "#E5E5E5"
		res.BgColorNight = "#BB5B76"
		res.BorderColorNight = "#BB5B76"
	case model.BgTransparentTextOrange:
		res.TextColor = "#FFFB9E60"
		res.BgStyle = int32(model.BgStyleNoFillAndNoStroke)
	case model.BgColorPurple:
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FF7D75F2"
		res.BorderColor = "#FF7D75F2"
		res.BgStyle = int32(model.BgStyleFill)
	case model.BgColorTransparentRed:
		res.TextColor = "#FFFB7299"
		res.BorderColor = "#FFFB7299"
		res.BgStyle = int32(model.BgStyleStroke)
	case model.BgColorFillingOrange:
		// 白天
		res.TextColor = "#FFFA8E57"
		res.BgColor = "#FFFFF1EA"
		res.BorderColor = "#FFFFF1EA"
		res.BgStyle = int32(model.BgStyleFill)
		// 夜间
		res.TextColorNight = "#FFBA6B45"
		res.BgColorNight = "#FF3B352E"
		res.BorderColorNight = "#FF3B352E"
	case model.BgColorLumpOrange:
		//白天
		res.TextColor = "#FF6633"
		res.BgColor = "#FFF1ED"
		res.BorderColor = "#FFF1ED"
		res.BgStyle = int32(model.BgStyleFill)
		// 夜间
		res.TextColorNight = "#BF5330"
		res.BgColorNight = "#3D2D29"
		res.BorderColorNight = "#3D2D29"
	case model.BgColorContourOrange:
		// 白天
		res.TextColor = "#FF6633"
		res.BorderColor = "#FF6633"
		res.BgStyle = int32(model.BgStyleStroke)
		// 夜间
		res.TextColorNight = "#BF5330"
		res.BorderColorNight = "#BF5330"
	}
	return
}

func avatarFrom(status *cardm.AvatarStatus) (avatar *api.Avatar) {
	if status == nil {
		return
	}
	avatar = &api.Avatar{
		Cover:        status.Cover,
		Text:         status.Text,
		Uri:          model.FillURI(status.Goto, 0, 0, status.Param, nil),
		Type:         status.Type,
		Event:        model.AvatarEvent[status.Goto],
		EventV2:      model.AvatarEventV2[status.Goto],
		DefalutCover: status.DefalutCover,
	}
	return
}

func buttonFrom(v interface{}, plat int8) (button *api.Button) {
	//nolint:gosimple
	switch v.(type) {
	case *cardm.Tag:
		t := v.(*cardm.Tag)
		if t != nil {
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    t.Name,
				Uri:     model.FillURI(model.GotoTag, 0, 0, strconv.FormatInt(t.ID, 10), nil),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case []*audio.Ctg:
		ctgs := v.([]*audio.Ctg)
		if len(ctgs) > 1 {
			var name string
			if ctgs[0] != nil {
				name = ctgs[0].ItemVal
				if ctgs[1] != nil {
					name += " · " + ctgs[1].ItemVal
				}
			}
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    name,
				Uri:     model.FillURI(model.GotoAudioTag, 0, 0, "", model.AudioTagHandler(ctgs)),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case []*article.Category:
		ctgs := v.([]*article.Category)
		if len(ctgs) > 1 {
			var name string
			if ctgs[0] != nil {
				name = ctgs[0].Name
				if ctgs[1] != nil {
					name += " · " + ctgs[1].Name
				}
			}
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    name,
				Uri:     model.FillURI(model.GotoArticleTag, 0, 0, "", model.ArticleTagHandler(ctgs, plat)),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *live.Room:
		r := v.(*live.Room)
		if r != nil {
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    r.AreaV2Name,
				Uri:     model.FillURI(model.GotoLiveTag, 0, 0, strconv.FormatInt(r.AreaV2ParentID, 10), model.LiveRoomTagHandler(r)),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *live.Card:
		card := v.(*live.Card)
		if card != nil {
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    card.Uname,
				Uri:     model.FillURI(model.GotoMid, 0, 0, strconv.FormatInt(card.UID, 10), nil),
				Event:   model.EventUpClick,
				EventV2: model.EventV2UpClick,
			}
		}
	case *bplus.Picture:
		p := v.(*bplus.Picture)
		if p != nil {
			if len(p.Topics) == 0 {
				return
			}
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    p.Topics[0],
				Uri:     model.FillURI(model.GotoPictureTag, 0, 0, p.Topics[0], nil),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *cardm.ButtonStatus:
		b := v.(*cardm.ButtonStatus)
		if b != nil {
			//nolint:ineffassign
			event, _ := model.ButtonEvent[b.Goto]
			eventV2, ok := model.ButtonEventV2[b.Goto]
			if ok {
				button = &api.Button{
					Text:     model.ButtonText[b.Goto],
					Param:    b.Param,
					Event:    event,
					Selected: int32(b.IsAtten),
					Type:     model.ButtonTheme,
					EventV2:  eventV2,
				}
				if b.Relation != nil {
					const _follow = 2
					button.Relation = &api.Relation{
						Status:     int32(b.Relation.Status),
						IsFollow:   int32(b.Relation.IsFollow),
						IsFollowed: int32(b.Relation.IsFollowed),
					}
					switch b.Relation.Status {
					case _follow:
						button.Selected = _follow
					}
				}
			} else {
				button = &api.Button{
					Text:  b.Text,
					Param: b.Param,
					Uri:   model.FillURI(b.Goto, 0, 0, b.Param, nil),
				}
				if b.Event != "" {
					button.Event = b.Event
					button.EventV2 = b.EventV2
				} else {
					button.Event = model.EventChannelClick
					button.EventV2 = model.EventV2ChannelClick
				}
				if b.Type != 0 {
					button.Type = b.Type
				} else {
					button.Type = model.ButtonGrey
				}
			}
		}
	case *bangumi.EpPlayer:
		ep := v.(*bangumi.EpPlayer)
		if ep != nil && ep.Season != nil {
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    ep.Season.TypeName,
				Uri:     ep.RegionURI,
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *operate.Channel:
		channel := v.(*operate.Channel)
		if channel != nil {
			button = &api.Button{
				Type:    model.ButtonGrey,
				Text:    channel.ChannelName,
				Uri:     model.FillURI(model.GotoChannel, 0, 0, strconv.FormatInt(channel.ChannelID, 10), nil),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case nil:
	default:
		log.Warn("buttonFrom: unexpected type %T", v)
	}
	return
}

func unionAuthorGRPC(a *arcgrpc.ArcPlayer) (name string) {
	if a.Arc.Rights.IsCooperation == 1 {
		name = a.Arc.Author.Name + " 等联合创作"
		return
	}
	name = a.Arc.Author.Name
	return
}

func adInfoChangeProto(ad *cm.AdInfo) (res *api.AdInfo) {
	res = &api.AdInfo{
		CreativeId:   ad.CreativeID,
		CreativeType: ad.CreativeType,
		CardType:     ad.CardType,
		AdCb:         ad.AdCb,
		Resource:     ad.Resource,
		Source:       ad.Source,
		RequestId:    ad.RequestID,
		IsAd:         ad.IsAd,
		CmMark:       ad.CmMark,
		Index:        ad.Index,
		IsAdLoc:      ad.IsAdLoc,
		CardIndex:    ad.CardIndex,
		ClientIp:     ad.ClientIP,
	}
	if ad.CreativeContent != nil {
		res.CreativeContent = &api.CreativeContent{
			Title:       ad.CreativeContent.Title,
			Description: ad.CreativeContent.Desc,
			VideoId:     ad.CreativeContent.VideoID,
			Username:    ad.CreativeContent.UserName,
			ImageUrl:    ad.CreativeContent.ImageURL,
			ImageMd5:    ad.CreativeContent.ImageMD5,
			LogUrl:      ad.CreativeContent.LogURL,
			LogMd5:      ad.CreativeContent.LogMD5,
			Url:         ad.CreativeContent.URL,
			ClickUrl:    ad.CreativeContent.ClickURL,
			ShowUrl:     ad.CreativeContent.ShowURL,
		}
	}
	return
}

func reasonStyleFromV2(rcmd *ai.Item, text string, gt model.Gt, plat int8, build int) (res *api.ReasonStyle) {
	if rcmd.RcmdReason == nil || text == "" || rcmd.RcmdReason.JumpID == 0 || rcmd.RcmdReason.JumpGoto == "" {
		return
	}
	res = topReasonStyleFrom(rcmd, text, gt, nil)
	var (
		style int8
		urlf  func(uri string) string
	)
	if style = rcmd.CornerMark; style == 0 {
		if rcmd.RcmdReason.Content != "" {
			style = rcmd.RcmdReason.CornerMark
		}
	}
	switch style {
	case 0, 2:
		//nolint:exhaustive
		switch model.Gt(rcmd.RcmdReason.JumpGoto) {
		case model.GotoAvConverge, model.GotoMultilayerConverge:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/6983dc5b73d32a8241421a7b25f78c855b8e0362.png"
		case model.GotoPlaylist:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/b985518f35ad2b12eec34d4b3b6dca33df2b85a2.png"
		case model.GotoTag:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/c735ef9f33feb19f52bce852355e72a6b367c466.png"
		}
	case 1:
		//nolint:exhaustive
		switch model.Gt(rcmd.RcmdReason.JumpGoto) {
		case model.GotoAvConverge, model.GotoMultilayerConverge:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/8ba6d17e066f6ad3497e071abe654615fb073726.png"
		case model.GotoPlaylist:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/e0bd607cb58289fb32866e19ec207efae23657de.png"
		case model.GotoTag:
			res.IconUrl = "https://i0.hdslb.com/bfs/archive/0d1185d6ceca3de0e0a99bdce0479838265adcc3.png"
		}
	}
	jumpGoto := model.Gt(rcmd.RcmdReason.JumpGoto)
	//nolint:exhaustive
	switch model.Gt(rcmd.RcmdReason.JumpGoto) {
	case model.GotoAvConverge, model.GotoMultilayerConverge:
		jumpGoto = model.GotoAvConverge
		urlf = model.TrackIDHandler(rcmd.TrackID, rcmd, plat, build)
	}
	res.Event = model.EventButtonClick
	res.EventV2 = model.EventV2ButtonClick
	res.Uri = model.FillURI(jumpGoto, 0, 0, strconv.FormatInt(rcmd.RcmdReason.JumpID, 10), urlf)
	return
}

func bottomReasonStyleFrom(rcmd *ai.Item, text string, _ model.Gt, op *operate.Card) (res *api.ReasonStyle) {
	if text == "" || rcmd == nil {
		return
	}
	var (
		style, bgstyle int8
	)
	if style = rcmd.CornerMark; style == 0 {
		if rcmd.RcmdReason != nil {
			if rcmd.RcmdReason.Content == "" {
				style = 0
			} else {
				style = rcmd.RcmdReason.CornerMark
			}
		}
	}
	if op != nil {
		if _, ok := op.SwitchStyle[model.SwitchNewReason]; ok {
			style = 5
		}
	}
	switch style {
	case 1:
		bgstyle = model.BgColorTransparentOrange
	case 3:
		bgstyle = model.BgTransparentTextOrange
	case 5:
		bgstyle = model.BgColorFillingOrange
	default:
		bgstyle = model.BgColorOrange
	}
	res = reasonStyleFrom(bgstyle, text)
	return
}

func fromLiveRoom(r *live.Room) (c *api.Args) {
	if r == nil {
		return
	}
	c = &api.Args{
		UpId:   r.UID,
		UpName: r.Uname,
		Rid:    int32(r.AreaV2ParentID),
		Rname:  r.AreaV2ParentName,
		Tid:    r.AreaV2ID,
		Tname:  r.AreaV2Name,
	}
	return
}

func fromArticle(m *article.Meta) (c *api.Args) {
	if m == nil {
		return
	}
	c = &api.Args{}
	if m.Author != nil {
		c.UpId = m.Author.Mid
		c.UpName = m.Author.Name
	}
	if len(m.Categories) != 0 {
		if m.Categories[0] != nil {
			c.Rid = int32(m.Categories[0].ID)
			c.Rname = m.Categories[0].Name
		}
		if len(m.Categories) > 1 {
			if m.Categories[1] != nil {
				c.Tid = m.Categories[1].ID
				c.Tname = m.Categories[1].Name
			}
		}
	}
	return
}

// ThreePointWatchLater is
func (c *Card) ThreePointWatchLater(op *operate.Card) {
	if op != nil && ((op.Build >= 9275 && model.IsIPhone(op.Plat)) || (op.Build >= 5560000 && op.MobiApp == "android")) {
		//nolint:exhaustive
		switch c.CardGoto {
		case model.CardGotoReadCard:
			c.ThreePointV4 = &api.ThreePointV4{
				SharePlane: &api.SharePlane{
					Title:     c.Title,
					Cover:     c.Cover,
					Aid:       op.ID,
					ShareTo:   op.Share,
					Author:    c.Args.UpName,
					AuthorId:  c.Args.UpId,
					ShortLink: c.shareURL(strconv.FormatInt(op.ID, 10)),
				},
				WatchLater: nil,
			}
			return
		}
	}
	if c.CardGoto == model.CardGotoAv || c.CardGoto == model.CardGotoPlayer || c.CardGoto == model.CardGotoUpRcmdAv || c.Goto == model.GotoAv {
		c.ThreePoint = &api.ThreePoint{}
		c.ThreePoint.WatchLater = 1
		c.ThreePointV2 = append(c.ThreePointV2, &api.ThreePointV2{Title: "添加至稍后再看", Type: model.ThreePointWatchLater})
	}
}

// ThreePointWatchLaterV3 is
func (c *Card) ThreePointWatchLaterV3() {
	const (
		_watchLater = "https://i0.hdslb.com/bfs/archive/de81a42978c1afbe2d2625e1f0ffa6f181852ca8.png"
	)
	if c.CardGoto == model.CardGotoAv || c.CardGoto == model.CardGotoPlayer || c.CardGoto == model.CardGotoUpRcmdAv || c.Goto == model.GotoAv {
		c.ThreePointV3 = append(c.ThreePointV3, &api.ThreePointV3{Title: "稍后再看", Type: model.ThreePointWatchLater, Icon: _watchLater})
	}
}

func (c *Card) shareURL(param string) (url string) {
	//nolint:exhaustive
	switch c.CardGoto {
	case model.CardGotoReadCard:
		url = "https://www.bilibili.com/read/cv" + param
	}
	return
}
