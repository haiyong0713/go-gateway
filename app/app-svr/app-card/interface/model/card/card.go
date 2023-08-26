package card

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	protov2 "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	"go-gateway/app/app-svr/app-card/interface/model/card/report"
	"go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/interface/model/sets"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
)

var (
	_inlineGotoSet = sets.NewString("inline_av", "inline_av_v2", "inline_pgc", "inline_live")
)

// ButtonStatus is
type ButtonStatus struct {
	Text     string
	Goto     model.Gt
	Param    string
	IsAtten  int8
	Type     model.Type
	Event    model.Event
	EventV2  model.Event
	URI      string `json:"uri,omitempty"`
	Relation *model.Relation
}

type CustomizedButtonMeta struct {
	Text string
	URI  string
}

// AvatarStatus is
type AvatarStatus struct {
	Cover        string
	Text         string
	Goto         model.Gt
	Param        string
	Type         model.Type
	FaceNftNew   int32 // face_nft_new 1 nft头像 0 非nft头像
	DefalutCover int32
}

// Base is
type Base struct {
	CardType        model.CardType                             `json:"card_type,omitempty"`
	CardGoto        model.CardGt                               `json:"card_goto,omitempty"`
	Goto            model.Gt                                   `json:"goto,omitempty"`
	Param           string                                     `json:"param,omitempty"`
	Bvid            string                                     `json:"bvid,omitempty"`
	Cover           string                                     `json:"cover,omitempty"`
	Title           string                                     `json:"title,omitempty"`
	URI             string                                     `json:"uri,omitempty"`
	DescButton      *Button                                    `json:"desc_button,omitempty"`
	ThreePoint      *ThreePoint                                `json:"three_point,omitempty"`
	Args            Args                                       `json:"args,omitempty"`
	PlayerArgs      *PlayerArgs                                `json:"player_args,omitempty"`
	UpArgs          *UpArgs                                    `json:"up_args,omitempty"`
	Idx             int64                                      `json:"idx,omitempty"`
	AdInfo          *cm.AdInfo                                 `json:"ad_info,omitempty"`
	Mask            *Mask                                      `json:"mask,omitempty"`
	PosRecUniqueID  string                                     `json:"pos_rec_unique_id,omitempty"`
	AuthorRelations map[int64]*relationgrpc.InterrelationReply `json:"-"`
	// ===============
	Right    bool                              `json:"-"`
	Rcmd     *ai.Item                          `json:"-"`
	Tagm     map[int64]*taggrpc.Tag            `json:"-"`
	IsAttenm map[int64]int8                    `json:"-"`
	HasLike  map[int64]int8                    `json:"-"`
	Statm    map[int64]*relationgrpc.StatReply `json:"-"`
	Cardm    map[int64]*accountgrpc.Card       `json:"-"`
	CardLen  int                               `json:"-"`
	Columnm  model.ColumnStatus                `json:"-"`
	// ===============
	FromType          string                    `json:"from_type,omitempty"`
	ThreePointV2      []*ThreePointV2           `json:"three_point_v2,omitempty"`
	ThreePointV3      []*ThreePointV3           `json:"three_point_v3,omitempty"`
	TrackID           string                    `json:"track_id,omitempty"`
	CmInfo            *cm.CmInfo                `json:"cm_info,omitempty"`
	ThreePointMeta    *threePointMeta.PanelMeta `json:"three_point_meta,omitempty"` // inline三点需求
	TalkBack          string                    `json:"talk_back,omitempty"`
	OgvCreativeId     int64                     `json:"ogv_creative_id,omitempty"`
	MaterialId        int64                     `json:"material_id,omitempty"`
	DislikeReportData string                    `json:"dislike_report_data,omitempty"`
}

// ThreePoint is
type ThreePoint struct {
	DislikeReasons []*DislikeReason `json:"dislike_reasons,omitempty"`
	Feedbacks      []*DislikeReason `json:"feedbacks,omitempty"`
	WatchLater     int8             `json:"watch_later,omitempty"`
}

// Mask is
type Mask struct {
	Avatar *Avatar `json:"avatar,omitempty"`
	Button *Button `json:"button,omitempty"`
}

// ThreePointV2 is
type ThreePointV2 struct {
	Title    string           `json:"title,omitempty"`
	Subtitle string           `json:"subtitle,omitempty"`
	Reasons  []*DislikeReason `json:"reasons,omitempty"`
	Type     string           `json:"type"`
	ID       int64            `json:"id,omitempty"`
	Toast    string           `json:"toast,omitempty"`
	Icon     string           `json:"icon,omitempty"`
}

func (c *Base) from(plat int8, build int, param, cover, title string, gt model.Gt, uri string, f func(uri string) string) {
	c.URI = model.FillURI(gt, plat, build, uri, f)
	c.Cover = cover
	c.Title = title
	if gt != "" {
		c.Goto = gt
	} else {
		c.Goto = model.Gt(c.CardGoto)
	}
	c.Param = param
}

func (c *Base) fillRcmdMeta(rcmd *ai.Item) {
	c.TrackID = rcmd.TrackID
	c.PosRecUniqueID = rcmd.PosRecUniqueID
	c.OgvCreativeId = rcmd.CreativeId
	c.MaterialId = rcmd.CreativeId
	c.DislikeReportData = report.BuildDislikeReportData(rcmd.CreativeId, rcmd.PosRecUniqueID)
}

// Handler is
type Handler interface {
	From(main interface{}, op *operate.Card) error
	Get() *Base
}

// Handle is
func Handle(plat int8, cardGoto model.CardGt, cardType model.CardType, column model.ColumnStatus, rcmd *ai.Item, tagm map[int64]*taggrpc.Tag, isAttenm, hasLike map[int64]int8,
	statm map[int64]*relationgrpc.StatReply, cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply) (hander Handler) {
	if model.IsPad(plat) {
		// 安卓 pad 也按照 ipad 的逻辑来
		return ipadHandle(cardGoto, cardType, rcmd, nil, isAttenm, nil, statm, cardm, authorRelations)
	}
	switch column {
	case model.ColumnSvrSingle, model.ColumnUserSingle:
		return singleHandle(cardGoto, cardType, rcmd, tagm, isAttenm, hasLike, statm, cardm, authorRelations)
	default:
		return doubleHandle(cardGoto, cardType, rcmd, tagm, isAttenm, hasLike, statm, cardm, authorRelations)
	}
}

// SwapTwoItem is
func SwapTwoItem(rs []Handler, i Handler) (is []Handler) {
	is = append(rs, rs[len(rs)-1])
	is[len(is)-2] = i
	return
}

func SwapThreeItem(rs []Handler, i Handler) (is []Handler) {
	is = append(rs, rs[len(rs)-1])
	is[len(is)-2] = i
	is[len(is)-3], is[len(is)-2] = is[len(is)-2], is[len(is)-3]
	return
}

func SwapFourItem(rs []Handler, i Handler) (is []Handler) {
	is = append(rs, rs[len(rs)-1])
	is[len(is)-2] = i
	is[len(is)-3], is[len(is)-2] = is[len(is)-2], is[len(is)-3]
	is[len(is)-4], is[len(is)-3] = is[len(is)-3], is[len(is)-4]
	return
}

// TopBottomRcmdReason is
func TopBottomRcmdReason(r *ai.RcmdReason, isAtten int8, userIsAtten map[int64]int8, gotoType string) (topRcmdReason, bottomRcomdReason string) {
	if r == nil {
		if _inlineGotoSet.Has(gotoType) {
			return
		}
		if isAtten == 1 {
			bottomRcomdReason = "已关注"
		}
		return
	}
	switch r.Style {
	case 3:
		if isAtten != 1 {
			return
		}
		bottomRcomdReason = r.Content
	case 4:
		attention, ok := userIsAtten[r.FollowedMid]
		if !ok || attention != 1 {
			return
		}
		topRcmdReason = "关注的人赞过"
	case 5:
		bottomRcomdReason = r.Content
	default:
		topRcmdReason = r.Content
	}
	return
}

// Button is
type Button struct {
	Text     string          `json:"text,omitempty"`
	Param    string          `json:"param,omitempty"`
	URI      string          `json:"uri,omitempty"`
	Event    model.Event     `json:"event,omitempty"`
	Selected int32           `json:"selected,omitempty"`
	Type     model.Type      `json:"type,omitempty"`
	EventV2  model.Event     `json:"event_v2,omitempty"`
	Relation *model.Relation `json:"relation,omitempty"`
}

func buttonFrom(v interface{}, plat int8) (button *Button) {
	//nolint:gosimple
	switch v.(type) {
	case *taggrpc.Tag:
		t := v.(*taggrpc.Tag)
		if t != nil {
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    t.Name,
				URI:     model.FillURI(model.GotoTag, 0, 0, strconv.FormatInt(t.Id, 10), nil),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *CustomizedButtonMeta:
		cb := v.(*CustomizedButtonMeta)
		if cb != nil {
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    cb.Text,
				URI:     cb.URI,
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
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    name,
				URI:     model.FillURI(model.GotoAudioTag, 0, 0, "", model.AudioTagHandler(ctgs)),
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
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    name,
				URI:     model.FillURI(model.GotoArticleTag, 0, 0, "", model.ArticleTagHandler(ctgs, plat)),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *live.Room:
		r := v.(*live.Room)
		if r != nil {
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    r.AreaV2Name,
				URI:     model.FillURI(model.GotoLiveTag, 0, 0, strconv.FormatInt(r.AreaV2ParentID, 10), model.LiveRoomTagHandler(r)),
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *live.Card:
		card := v.(*live.Card)
		if card != nil {
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    card.Uname,
				URI:     model.FillURI(model.GotoMid, 0, 0, strconv.FormatInt(card.UID, 10), nil),
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
			var uri string
			if p.IsNewChannel {
				for _, tf := range p.TopicInfos {
					if tf.TopicName == p.Topics[0] {
						uri = tf.TopicLink
					}
				}
			}
			if uri == "" {
				uri = model.FillURI(model.GotoPictureTag, 0, 0, p.Topics[0], nil)
			}
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    p.Topics[0],
				URI:     uri,
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *ButtonStatus:
		b := v.(*ButtonStatus)
		if b != nil {
			//nolint:ineffassign
			event, _ := model.ButtonEvent[b.Goto]
			eventV2, ok := model.ButtonEventV2[b.Goto]
			if ok {
				button = &Button{
					Text:     model.ButtonText[b.Goto],
					Param:    b.Param,
					Event:    event,
					Selected: int32(b.IsAtten),
					Type:     model.ButtonTheme,
					EventV2:  eventV2,
					Relation: b.Relation,
				}
			} else {
				button = &Button{
					Text:  b.Text,
					Param: b.Param,
					URI:   model.FillURI(b.Goto, 0, 0, b.Param, nil),
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
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    ep.Season.TypeName,
				URI:     ep.RegionURI,
				Event:   model.EventChannelClick,
				EventV2: model.EventV2ChannelClick,
			}
		}
	case *operate.Channel:
		channel := v.(*operate.Channel)
		if channel != nil {
			button = &Button{
				Type:    model.ButtonGrey,
				Text:    channel.ChannelName,
				URI:     model.FillURI(model.GotoChannel, 0, 0, strconv.FormatInt(channel.ChannelID, 10), nil),
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

// Avatar is
type Avatar struct {
	Cover        string       `json:"cover,omitempty"`
	Text         string       `json:"text,omitempty"`
	URI          string       `json:"uri,omitempty"`
	Type         model.Type   `json:"type,omitempty"`
	Event        model.Event  `json:"event,omitempty"`
	EventV2      model.Event  `json:"event_v2,omitempty"`
	DefalutCover int32        `json:"defalut_cover,omitempty"`
	UpID         int64        `json:"up_id,omitempty"`
	FaceNftNew   int32        `json:"face_nft_new,omitempty"`
	NftFaceIcon  *NftFaceIcon `json:"nft_face_icon,omitempty"` // nft角标展示信息
}

type NftFaceIcon struct {
	RegionType int32  `json:"region_type"` // nft所属区域 0 默认 1 大陆 2 港澳台
	Icon       string `json:"icon"`        // 角标链接
	ShowStatus int32  `json:"show_status"` // 展示状态 0:默认 1:放大20% 2:原图大小
}

func avatarFrom(status *AvatarStatus) (avatar *Avatar) {
	if status == nil {
		return
	}
	avatar = &Avatar{
		Cover:        status.Cover,
		Text:         status.Text,
		URI:          model.FillURI(status.Goto, 0, 0, status.Param, nil),
		Type:         status.Type,
		Event:        model.AvatarEvent[status.Goto],
		EventV2:      model.AvatarEventV2[status.Goto],
		DefalutCover: status.DefalutCover,
		FaceNftNew:   status.FaceNftNew,
	}
	if status.Goto == model.GotoMid {
		avatar.UpID, _ = strconv.ParseInt(status.Param, 10, 64)
	}
	return
}

// DislikeReason is
type DislikeReason struct {
	ID    int64  `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Toast string `json:"toast,omitempty"`
}

const (
	_noSeason           = 1
	_region             = 2
	_channel            = 3
	_upper              = 4
	_gotoStoryDislikeID = 9
	// dislikeExp
	_newDislike         = 1
	_dislikeReasonToast = "将减少相似内容推荐"
	_feedbackToast      = "将优化首页此类内容"

	_ogvAddFeedbacks = 1
	_watched         = 10
	_moreContent     = 12
	_repeatedRcmd    = 13
	_relation        = 16
)

// ThreePointFrom is
//
//nolint:gocognit
func (c *Base) ThreePointFrom(mobiApp string, build int, dislikeExp int, abtest *feed.Abtest, column model.ColumnStatus, isCloseRcmd int) {
	dislikeSubTitle := "(选择后将减少相似内容推荐)"
	dislikeReasonToast := _dislikeReasonToast
	if isCloseRcmd == 1 {
		dislikeSubTitle = ""
		dislikeReasonToast = "将在开启个性化推荐后生效"
	}
	// abtest
	switch dislikeExp {
	case _newDislike:
		//nolint:exhaustive
		switch c.CardGoto {
		case model.CardGotoLive, model.CardGotoPicture, model.CardGotoArticleS, model.CardGotoInlineLive:
			c.ThreePoint = &ThreePoint{}
			dislikeReasons := make([]*DislikeReason, 0, 4)
			if c.Args.UpName != "" {
				dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _upper, Name: "UP主:" + c.Args.UpName, Toast: dislikeReasonToast})
			}
			//nolint:exhaustive
			switch c.CardGoto {
			case model.CardGotoLive, model.CardGotoInlineLive:
				if c.Args.Tname != "" {
					dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _region, Name: "分区:" + c.Args.Tname, Toast: dislikeReasonToast})
				}
			case model.CardGotoPicture:
				if c.Args.Tname != "" {
					dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _channel, Name: "话题:" + c.Args.Tname, Toast: dislikeReasonToast})
				}
			}
			c.ThreePoint.DislikeReasons = append(dislikeReasons, &DislikeReason{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast})
			if CanEnableSwitchColumnThreePoint(mobiApp, build, abtest) && c.ThreePointMeta == nil {
				c.ThreePointV2 = append(c.ThreePointV2, constructSwitchColumnThreePoint(column))
			}
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: DislikeTitle(abtest), Subtitle: dislikeSubTitle, Reasons: c.ThreePoint.DislikeReasons, Type: model.ThreePointDislike})
			return
		}
	}
	// abtest
	switch c.CardGoto {
	case model.CardGotoBanner, model.CardGotoRank, model.CardGotoConverge, model.CardGotoBangumiRcmd, model.CardGotoInterest, model.CardGotoFollowMode, model.CardGotoVip, model.CardGotoTunnel:
		return
	case model.CardGotoAv, model.CardGotoPlayer, model.CardGotoUpRcmdAv, model.CardGotoChannelRcmd, model.CardGotoAvConverge, model.CardGotoMultilayerConverge, model.CardGotoInlineAv, model.CardGotoVerticalAv, model.CardGotoInlineAvV2:
		c.ThreePoint = &ThreePoint{}
		dislikeReasons := make([]*DislikeReason, 0, 4)
		if c.Args.UpName != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _upper, Name: "UP主:" + c.Args.UpName, Toast: dislikeReasonToast})
		}
		if c.Args.Rname != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _region, Name: "分区:" + c.Args.Rname, Toast: dislikeReasonToast})
		}
		if c.Args.Tname != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _channel, Name: "频道:" + c.Args.Tname, Toast: dislikeReasonToast})
		}
		if CanEnableAvDislikeInfo(c.Rcmd) {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _moreContent, Name: "此类内容过多", Toast: dislikeReasonToast})
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _repeatedRcmd, Name: "推荐过", Toast: dislikeReasonToast})
		}
		c.ThreePoint.DislikeReasons = append(dislikeReasons, &DislikeReason{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast})
		ReplaceStoryDislikeReason(c.ThreePoint.DislikeReasons, c.Rcmd)
		c.ThreePoint.Feedbacks = []*DislikeReason{{ID: 1, Name: "恐怖血腥", Toast: _feedbackToast}, {ID: 2, Name: "色情低俗", Toast: _feedbackToast}, {ID: 3, Name: "封面恶心", Toast: _feedbackToast}, {ID: 4, Name: "标题党/封面党", Toast: _feedbackToast}}
		c.ThreePoint.WatchLater = 1
		if c.CardGoto != model.CardGotoAvConverge && c.CardGoto != model.CardGotoMultilayerConverge {
			if c.CardGoto != model.CardGotoVerticalAv && resolveWatchLaterOnThreePoint(mobiApp, build, c.ThreePointMeta) {
				// 跳转story与新三点面板不出稍后再看
				c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "添加至稍后再看", Type: model.ThreePointWatchLater, Icon: model.IconWatchLater})
			}
			if CanEnableSwitchColumnThreePoint(mobiApp, build, abtest) && c.ThreePointMeta == nil {
				c.ThreePointV2 = append(c.ThreePointV2, constructSwitchColumnThreePoint(column))
			}
			if mobiApp != "android_i" { //国际版暂不支持稿件反馈
				c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "反馈", Subtitle: "(选择后将优化首页此类内容)", Reasons: c.ThreePoint.Feedbacks, Type: model.ThreePointFeedback})
			}
		}
		c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: DislikeTitle(abtest),
			Subtitle: dislikeSubTitle, Reasons: c.ThreePoint.DislikeReasons, Type: model.ThreePointDislike})
	case model.CardGotoAiStory:
		if CanEnableSwitchColumnThreePoint(mobiApp, build, abtest) && c.ThreePointMeta == nil {
			c.ThreePointV2 = append(c.ThreePointV2, constructSwitchColumnThreePoint(column))
		}
		if c.Title != "" || (mobiApp == "android" && build >= 609000) || (mobiApp == "iphone" && build > 10240) {
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "不感兴趣", Type: model.ThreePointDislike, ID: _noSeason})
		}
	case model.CardGotoBangumi, model.CardGotoPGC, model.CardGotoSpecialS:
		c.ThreePoint = &ThreePoint{}
		dislikeReasons := []*DislikeReason{{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast}}
		c.ThreePoint.DislikeReasons = dislikeReasons
		if CanEnableSwitchColumnThreePoint(mobiApp, build, abtest) && c.ThreePointMeta == nil {
			c.ThreePointV2 = append(c.ThreePointV2, constructSwitchColumnThreePoint(column))
		}
		if c.Rcmd != nil && CanEnableOGVFeedback(c.Rcmd, mobiApp, column) {
			if c.Rcmd.OgvDislikeInfo == ai.OgvWatched {
				c.ThreePoint.DislikeReasons = []*DislikeReason{
					{ID: _watched, Name: "看过了", Toast: dislikeReasonToast},
					{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast},
				}
			}
			c.ThreePoint.Feedbacks = []*DislikeReason{
				{ID: 3, Name: "封面恶心", Toast: _feedbackToast},
				{ID: 4, Name: "标题党/封面党", Toast: _feedbackToast}}
			c.ThreePointV2 = append(c.ThreePointV2,
				&ThreePointV2{
					Title:    "反馈",
					Subtitle: "(选择后将优化首页此类内容)",
					Reasons:  c.ThreePoint.Feedbacks,
					Type:     model.ThreePointFeedback})
			c.ThreePointV2 = append(c.ThreePointV2,
				&ThreePointV2{
					Title:    DislikeTitle(abtest),
					Subtitle: dislikeSubTitle,
					Reasons:  c.ThreePoint.DislikeReasons,
					Type:     model.ThreePointDislike,
				})
			return
		}
		// 老客户端用reason个数判断是否显示单条，为了兼容1标题+1理由 和 1单条理由的形式，改了不感兴趣的格式
		if (mobiApp == "iphone" && build > 8470) || (mobiApp == "android" && build > 5405000) || (mobiApp == "ipad" && build > 12080) || (mobiApp == "iphone_b" && build >= 8000) || (mobiApp == "android_i" && build >= 3000500) || (mobiApp == "android_b" && build >= 5370100) {
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "不感兴趣", Type: model.ThreePointDislike, ID: _noSeason})
		} else {
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Reasons: dislikeReasons, Type: model.ThreePointDislike})
		}
	default:
		c.ThreePoint = &ThreePoint{}
		dislikeReasons := []*DislikeReason{{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast}}
		c.ThreePoint.DislikeReasons = dislikeReasons
		if CanEnableSwitchColumnThreePoint(mobiApp, build, abtest) && c.ThreePointMeta == nil {
			c.ThreePointV2 = append(c.ThreePointV2, constructSwitchColumnThreePoint(column))
		}
		// 老客户端用reason个数判断是否显示单条，为了兼容1标题+1理由 和 1单条理由的形式，改了不感兴趣的格式
		if (mobiApp == "iphone" && build > 8470) || (mobiApp == "android" && build > 5405000) || (mobiApp == "ipad" && build > 12080) || (mobiApp == "iphone_b" && build >= 8000) || (mobiApp == "android_i" && build >= 3000500) || (mobiApp == "android_b" && build >= 5370100) {
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "不感兴趣", Type: model.ThreePointDislike, ID: _noSeason})
		} else {
			c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Reasons: dislikeReasons, Type: model.ThreePointDislike})
		}
	}
}

func DislikeTitle(abtest *feed.Abtest) string {
	const _dislikeWatch = 1
	if abtest != nil && abtest.DislikeText == _dislikeWatch {
		return "我不想看"
	}
	return "不感兴趣"
}

func CanEnableOGVFeedback(rcmd *ai.Item, mobiApp string, column model.ColumnStatus) bool {
	if rcmd.OgvDislikeInfo < _ogvAddFeedbacks {
		return false
	}
	if mobiApp != "android" && mobiApp != "iphone" {
		return false
	}
	if model.Columnm[column] != model.ColumnSvrDouble {
		return false
	}
	return true
}

func CanEnableAvDislikeInfo(rcmd *ai.Item) bool {
	if rcmd == nil || rcmd.AvDislikeInfo == 0 {
		return false
	}
	return true
}

func CanEnableSwitchColumnThreePoint(mobiApp string, build int, abtest *feed.Abtest) bool {
	if abtest == nil {
		return false
	}
	if !abtest.CanSupportThreePoint() {
		return false
	}
	if (mobiApp == "iphone" && build >= 62800000) || (mobiApp == "android" && build >= 6280000) {
		return true
	}
	return false
}

func constructSwitchColumnThreePoint(column model.ColumnStatus) *ThreePointV2 {
	title := ""
	type_ := ""
	toast := ""
	subTitle := "(首页模式)"
	icon := ""
	switch column {
	case model.ColumnSvrDouble, model.ColumnDefault, model.ColumnUserDouble:
		title = "切换至单列"
		type_ = model.ThreePointSwitchToSingle
		toast = "已成功切换至单列模式"
		icon = model.IconSwitchToSingle
	case model.ColumnSvrSingle, model.ColumnUserSingle:
		title = "切换至双列"
		type_ = model.ThreePointSwitchToDouble
		toast = "已成功切换至双列模式"
		icon = model.IconSwitchToDouble
	default:
	}
	return &ThreePointV2{
		Title:    title,
		Type:     type_,
		Toast:    toast,
		Subtitle: subTitle,
		Icon:     icon,
	}
}

func resolveWatchLaterOnThreePoint(mobiApp string, build int, meta *threePointMeta.PanelMeta) bool {
	if meta == nil {
		return true
	}
	if (mobiApp == "iphone" && build >= 62600000) || (mobiApp == "android" && build >= 6260000) {
		return false
	}
	return true
}

func ReplaceStoryDislikeReason(reasons []*DislikeReason, rcmd *ai.Item) {
	for _, reason := range reasons {
		if reason.ID == _region && canReplaceStoryDislikeReason(rcmd) {
			reason.ID = _gotoStoryDislikeID
			reason.Name = rcmd.GotoStoryDislikeReason()
			reason.Toast = "将减少竖屏模式推荐"
		}
		if reason.ID == _channel && canReplaceRelationsDislikeReason(rcmd) {
			reason.ID = _relation
			reason.Name = rcmd.RelationDislikeText
		}
	}
}

func canReplaceStoryDislikeReason(rcmd *ai.Item) bool {
	return rcmd != nil && rcmd.StoryDislike == 1 &&
		rcmd.JumpGoto == string(model.GotoVerticalAv)
}

func canReplaceRelationsDislikeReason(rcmd *ai.Item) bool {
	return rcmd != nil && rcmd.Goto == "av" && rcmd.RelationDislike == 1 && rcmd.RelationDislikeText != ""
}

type ThreePointV3 struct {
	Title         string           `json:"title,omitempty"`
	SelectedTitle string           `json:"selected_title,omitempty"`
	Subtitle      string           `json:"subtitle,omitempty"`
	Reasons       []*DislikeReason `json:"reasons,omitempty"`
	Type          string           `json:"type"`
	ID            int64            `json:"id,omitempty"`
	Selected      int8             `json:"selected,omitempty"`
	Icon          string           `json:"icon,omitempty"`
	SelectedIcon  string           `json:"selected_icon,omitempty"`
	URL           string           `json:"url,omitempty"`
	DefaultID     int              `json:"default_id,omitempty"`
}

func (c *Base) ThreePointFromV3(mobiApp string, build int, dislikeExp int) {
	const (
		_noSeason      = 1
		_region        = 2
		_channel       = 3
		_upper         = 4
		_whyContentURL = ""
		// icon
		_feedbackIcon     = "https://i0.hdslb.com/bfs/archive/127bab8a72f065b38a74b7cd76f5b112edda7026.png"
		_watchLater       = "https://i0.hdslb.com/bfs/archive/de81a42978c1afbe2d2625e1f0ffa6f181852ca8.png"
		_likeIcon         = "https://i0.hdslb.com/bfs/archive/8b95bd3decc7b327482629d48cd6f2365305f14b.png"
		_likeSelectedIcon = "https://i0.hdslb.com/bfs/archive/e62edd4c4c56d0c2a795236933700d011101c991.png"
		_dislikeIcon      = "https://i0.hdslb.com/bfs/archive/50495ad7f33ccffba988f49679dcd85666f38ddd.png"
		_whyContentIcon   = "https://i0.hdslb.com/bfs/archive/54316c67cf4afe51d1f7990dbbeedbf53a87a20e.png"
		// dislikeExp
		_newDislike = 1
	)
	var (
		_feedbackURL = fmt.Sprintf("https://www.bilibili.com/appeal/?avid=%s", c.Param)
	)
	// abtest
	switch dislikeExp {
	case _newDislike:
		//nolint:exhaustive
		switch c.CardGoto {
		case model.CardGotoLive, model.CardGotoPicture, model.CardGotoArticleS:
			dislikeReasons := make([]*DislikeReason, 0, 2)
			if c.Args.UpName != "" {
				dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _upper, Name: "不想看UP主:" + c.Args.UpName})
			}
			//nolint:exhaustive
			switch c.CardGoto {
			case model.CardGotoLive:
				if c.Args.Tname != "" {
					dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _region, Name: "不想看分区:" + c.Args.Tname})
				}
			case model.CardGotoPicture:
				if c.Args.Tname != "" {
					dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _channel, Name: "不想看话题:" + c.Args.Tname})
				}
			}
			c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "我不想看这个内容", Subtitle: "为了解你的喜好，请告诉我们原因",
				Reasons: dislikeReasons, Type: model.ThreePointDislike, Icon: _dislikeIcon, DefaultID: _noSeason})
			return
		}
	}
	// abtest
	switch c.CardGoto {
	case model.CardGotoBanner, model.CardGotoRank, model.CardGotoConverge, model.CardGotoBangumiRcmd, model.CardGotoInterest, model.CardGotoFollowMode, model.CardGotoVip, model.CardGotoTunnel:
		return
	case model.CardGotoAv, model.CardGotoPlayer, model.CardGotoUpRcmdAv, model.CardGotoChannelRcmd, model.CardGotoAvConverge, model.CardGotoMultilayerConverge, model.CardGotoInlineAv:
		dislikeReasons := make([]*DislikeReason, 0, 2)
		if c.Args.UpName != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _upper, Name: "不想看UP主:" + c.Args.UpName})
		}
		if c.Args.Tname != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _channel, Name: "不想看频道:" + c.Args.Tname})
		}
		if c.CardGoto != model.CardGotoAvConverge && c.CardGoto != model.CardGotoMultilayerConverge {
			c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "稍后再看", Type: model.ThreePointWatchLater, Icon: _watchLater})
			//nolint:staticcheck
			if c.CardGoto == model.CardGotoAv {
				// var (
				// 	selected int8
				// )
				// if c.Rcmd != nil && c.HasLike != nil {
				// 	switch c.HasLike[c.Rcmd.ID] {
				// 	case 1:
				// 		selected = 1
				// 	}
				// }
				// c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "点赞", SelectedTitle: "已点赞", Type: model.ThreePointLike, Icon: _likeIcon, SelectedIcon: _likeSelectedIcon, Selected: selected})
			}
			c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "举报", Type: model.ThreePointFeedback, Icon: _feedbackIcon, URL: _feedbackURL})
			// c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "我为什么会看到这个内容", Type: model.ThreePointWhyContent})
		}
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "我不想看这个内容", Subtitle: "为了解你的喜好，请告诉我们原因",
			Reasons: dislikeReasons, Type: model.ThreePointDislike, Icon: _dislikeIcon, DefaultID: _noSeason})
	default:
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "我不想看这个内容", Subtitle: "为了解你的喜好，请告诉我们原因",
			Type: model.ThreePointDislike, DefaultID: _noSeason, Icon: _dislikeIcon})
	}
}

// ThreePointChannel is
func (c *Base) ThreePointChannel() {
	const (
		_noSeason = 1
		_upper    = 4
	)
	if c.CardGoto == model.CardGotoAv || c.CardGoto == model.CardGotoPlayer || c.CardGoto == model.CardGotoUpRcmdAv {
		c.ThreePoint = &ThreePoint{}
		if c.Args.UpName != "" {
			c.ThreePoint.DislikeReasons = append(c.ThreePoint.DislikeReasons, &DislikeReason{ID: _upper, Name: "UP主:" + c.Args.UpName, Toast: _dislikeReasonToast})
		}
		c.ThreePoint.DislikeReasons = append(c.ThreePoint.DislikeReasons, &DislikeReason{ID: _noSeason, Name: "不感兴趣", Toast: _dislikeReasonToast})
		c.ThreePoint.WatchLater = 1
		c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "添加至稍后再看", Type: model.ThreePointWatchLater})
		c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "不感兴趣", Subtitle: "(选择后将减少相似内容推荐)", Reasons: c.ThreePoint.DislikeReasons, Type: model.ThreePointDislike})
	}
}

func (c *Base) ThreePointChannelV3() {
	const (
		_noSeason = 1
		_upper    = 4
		// icon
		_dislikeIcon = "https://i0.hdslb.com/bfs/archive/50495ad7f33ccffba988f49679dcd85666f38ddd.png"
		_watchLater  = "https://i0.hdslb.com/bfs/archive/de81a42978c1afbe2d2625e1f0ffa6f181852ca8.png"
	)
	//nolint:exhaustive
	switch c.CardGoto {
	case model.CardGotoAv, model.CardGotoPlayer, model.CardGotoUpRcmdAv:
		dislikeReasons := make([]*DislikeReason, 0, 1)
		if c.Args.UpName != "" {
			dislikeReasons = append(dislikeReasons, &DislikeReason{ID: _upper, Name: "不想看UP主:" + c.Args.UpName, Toast: _dislikeReasonToast})
		}
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "稍后再看", Type: model.ThreePointWatchLater, Icon: _watchLater})
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "我不想看这个内容", Subtitle: "为了解你的喜好，请告诉我们原因",
			Reasons: dislikeReasons, Type: model.ThreePointDislike, Icon: _dislikeIcon, DefaultID: _noSeason})
	}
}

// ThreePointWatchLater is
func (c *Base) ThreePointWatchLater() {
	if c.CardGoto == model.CardGotoAv || c.CardGoto == model.CardGotoPlayer || c.CardGoto == model.CardGotoUpRcmdAv || c.Goto == model.GotoAv {
		c.ThreePoint = &ThreePoint{}
		c.ThreePoint.WatchLater = 1
		c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "添加至稍后再看", Type: model.ThreePointWatchLater})
	}
}

// ThreePointWatchLater is
func (c *Base) ThreePointWatchLaterV3() {
	const (
		_watchLater = "https://i0.hdslb.com/bfs/archive/de81a42978c1afbe2d2625e1f0ffa6f181852ca8.png"
	)
	if c.CardGoto == model.CardGotoAv || c.CardGoto == model.CardGotoPlayer || c.CardGoto == model.CardGotoUpRcmdAv || c.Goto == model.GotoAv {
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "稍后再看", Type: model.ThreePointWatchLater, Icon: _watchLater})
	}
}

// TabThreePointWatchLater is
func (c *Base) TabThreePointWatchLater() {
	if c.Goto == model.GotoAv && c.CardGoto != model.CardGotoPlayer {
		c.ThreePoint = &ThreePoint{}
		c.ThreePoint.WatchLater = 1
		c.ThreePointV2 = append(c.ThreePointV2, &ThreePointV2{Title: "添加至稍后再看", Type: model.ThreePointWatchLater})
	}
}

// TabThreePointWatchLater is
func (c *Base) TabThreePointWatchLaterV3() {
	const (
		_watchLater = "https://i0.hdslb.com/bfs/archive/de81a42978c1afbe2d2625e1f0ffa6f181852ca8.png"
	)
	if c.Goto == model.GotoAv && c.CardGoto != model.CardGotoPlayer {
		c.ThreePointV3 = append(c.ThreePointV3, &ThreePointV3{Title: "稍后再看", Type: model.ThreePointWatchLater, Icon: _watchLater})
	}
}

// Args is
type Args struct {
	Type            int8    `json:"type,omitempty"`
	UpID            int64   `json:"up_id,omitempty"`
	UpName          string  `json:"up_name,omitempty"`
	Rid             int32   `json:"rid,omitempty"`
	Rname           string  `json:"rname,omitempty"`
	Tid             int64   `json:"tid,omitempty"`
	Tname           string  `json:"tname,omitempty"`
	TrackID         string  `json:"track_id,omitempty"`
	State           string  `json:"state,omitempty"`
	ConvergeType    int32   `json:"converge_type,omitempty"`
	Aid             int64   `json:"aid,omitempty"`
	Duration        int64   `json:"duration,omitempty"`
	ReportExtraInfo *Report `json:"report_extra_info,omitempty"`
	RoomID          int64   `json:"room_id,omitempty"`
	Online          int32   `json:"online,omitempty"`
	IsFollow        int8    `json:"is_follow,omitempty"`
}

type Report struct {
	DynamicActivity string `json:"dynamic_activity,omitempty"`
}

func (c *Args) fromShopping(s *show.Shopping) {
	c.Type = s.Type
}

func (c *Args) fromArchiveGRPC(a *arcgrpc.Arc, t *taggrpc.Tag) {
	if a != nil {
		c.Aid = a.Aid
		c.UpID = a.Author.Mid
		c.UpName = a.Author.Name
		c.Rid = a.TypeID
		c.Rname = a.TypeName
	}
	if t != nil {
		c.Tid = t.Id
		c.Tname = t.Name
	}
}

func (c *Args) fromLiveRoom(r *live.Room) {
	if r == nil {
		return
	}
	c.UpID = r.UID
	c.UpName = r.Uname
	c.Rid = int32(r.AreaV2ParentID)
	c.Rname = r.AreaV2ParentName
	c.Tid = r.AreaV2ID
	c.Tname = r.AreaV2Name
	c.RoomID = r.RoomID
	c.Online = r.Online
}

func (c *Args) fromLiveUp(card *live.Card) {
	if card == nil {
		return
	}
	c.UpID = card.UID
	c.UpName = card.Uname
}

func (c *Args) fromAudio(a *audio.Audio) {
	if a == nil {
		return
	}
	c.Type = a.Type
	if len(a.Ctgs) != 0 {
		c.Rid = int32(a.Ctgs[0].ItemID)
		c.Rname = a.Ctgs[0].ItemVal
		if len(a.Ctgs) > 1 {
			c.Tid = a.Ctgs[1].ItemID
			c.Tname = a.Ctgs[1].ItemVal
		}
	}
}

func (c *Args) fromArticle(m *article.Meta) {
	if m == nil {
		return
	}
	if m.Author != nil {
		c.UpID = m.Author.Mid
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
}

func (c *Args) fromAvConverge(r *ai.Item) {
	if r == nil {
		return
	}
	var (
		_stateAv       = "typeA"
		_stateConverge = "typeB"
	)
	c.TrackID = r.TrackID
	if r.ConvergeInfo != nil {
		c.ConvergeType = r.ConvergeInfo.ConvergeType
	}
	//nolint:exhaustive
	switch model.Gt(r.JumpGoto) {
	case model.GotoAv:
		c.State = _stateAv
	case model.GotoConverge:
		c.State = _stateConverge
	}
}

func (c *Args) fromPicture(b *bplus.Picture) {
	if b == nil {
		return
	}
	c.UpID = b.Mid
	c.UpName = b.NickName
	if len(b.TopicInfos) > 0 {
		c.Tid = b.TopicInfos[0].TopicID
		c.Tname = b.TopicInfos[0].TopicName
	}
	for _, v := range b.TopicInfos {
		if v.IsActivity == 1 {
			c.ReportExtraInfo = &Report{DynamicActivity: v.TopicName}
			break
		}
	}
}

// PlayerArgs is
type PlayerArgs struct {
	IsLive                     int8     `json:"is_live,omitempty"`
	Aid                        int64    `json:"aid,omitempty"`
	Cid                        int64    `json:"cid,omitempty"`
	SubType                    int32    `json:"sub_type,omitempty"`
	RoomID                     int64    `json:"room_id,omitempty"`
	EpID                       int64    `json:"ep_id,omitempty"`
	IsPreview                  int32    `json:"is_preview,omitempty"`
	Type                       model.Gt `json:"type"`
	Duration                   int64    `json:"duration,omitempty"`
	SeasonID                   int64    `json:"season_id,omitempty"`
	ManualPlay                 int8     `json:"manual_play,omitempty"`
	HidePlayButton             bool     `json:"hide_play_button,omitempty"`
	ReportHistory              int8     `json:"report_history,omitempty"`
	ReportRequiredPlayDuration int64    `json:"report_required_play_duration,omitempty"`
	ReportRequiredTime         int64    `json:"report_required_time,omitempty"`
	ContentMode                int64    `json:"content_mode,omitempty"` // 0,默认，按比例完整展示，会留空；1，按比例撑满，多余部分裁剪
}

func playerArgsFrom(v interface{}) (playerArgs *PlayerArgs) {
	//nolint:gosimple
	switch v.(type) {
	case *arcgrpc.ArcPlayer:
		a := v.(*arcgrpc.ArcPlayer)
		if a == nil || (a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && a.Arc.Rights.Autoplay != 1) ||
			(a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.Arc.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes) {
			return
		}
		playerArgs = &PlayerArgs{Aid: a.Arc.Aid, Cid: a.DefaultPlayerCid, Type: model.GotoAv, Duration: a.Arc.Duration}
	case *live.Room:
		r := v.(*live.Room)
		if r == nil {
			return
		}
		playerArgs = &PlayerArgs{RoomID: r.RoomID, IsLive: 1, Type: model.GotoLive}
	case *bangumi.EpPlayer:
		ep := v.(*bangumi.EpPlayer)
		if ep == nil {
			return
		}
		playerArgs = &PlayerArgs{Aid: ep.AID, Cid: ep.CID, EpID: ep.EpID, IsPreview: ep.IsPreview, Type: model.GotoBangumi, Duration: ep.Duration, SubType: ep.Season.Type, SeasonID: ep.Season.SeasonID}
	case *pgcinline.EpisodeCard:
		ep := v.(*pgcinline.EpisodeCard)
		if ep == nil || ep.Season == nil {
			return
		}
		playerArgs = &PlayerArgs{Aid: ep.Aid, Cid: ep.Cid, EpID: int64(ep.EpisodeId), IsPreview: ep.IsPreview, Type: model.GotoBangumi, Duration: ep.Duration, SubType: ep.Season.Type, SeasonID: int64(ep.Season.SeasonId)}
	case nil:
	default:
		log.Warn("playerArgsFrom: unexpected type %T", v)
	}
	return
}

type UpArgs struct {
	UpID     int64  `json:"up_id,omitempty"`
	UpName   string `json:"up_name,omitempty"`
	UpFace   string `json:"up_face,omitempty"`
	Selected int8   `json:"selected,omitempty"`
}

func upArgsFrom(v interface{}, state int8) (upArgs *UpArgs) {
	//nolint:gosimple
	switch v.(type) {
	case *arcgrpc.Arc:
		a := v.(*arcgrpc.Arc)
		if a == nil {
			return
		}
		upArgs = &UpArgs{UpID: a.Author.Mid, UpName: a.Author.Name, UpFace: a.Author.Face, Selected: state}
	default:
		log.Warn("upArgsFrom: unexpected type %T", v)
	}
	return
}

// rcmdReason
func rcmdReason(r *ai.RcmdReason, name string, isAtten int8, userIsAtten map[int64]int8, gotoType string) (rcmdReason, desc string) {
	// "rcmd_reason":{"content":"已关注","font":1,"grounding":"yellow","id":3,"position":"bottom","style":3}
	if r == nil {
		if _inlineGotoSet.Has(gotoType) {
			return
		}
		if isAtten == 1 {
			rcmdReason = "已关注"
			desc = name
		}
		return
	}
	switch r.Style {
	case 3:
		if isAtten != 1 {
			return
		}
		rcmdReason = r.Content
		desc = name
	case 4:
		// https://info.bilibili.co/pages/viewpage.action?pageId=4551903
		// style 4 动态图文卡片样式，额外添加字段【 "followed_mid" : 123 】，保存的是【召回此动态卡片的关注用户mid】，服务端需要根据此mid拿用户昵称
		if uIsAtten, ok := userIsAtten[r.FollowedMid]; !ok || uIsAtten != 1 {
			return
		}
		if r.Content == "" {
			r.Content = "关注的人赞过"
		}
		rcmdReason = r.Content
	case 6:
		rcmdReason = r.Content
		desc = name
	default:
		rcmdReason = r.Content
	}
	return
}

// ReasonStyle reason style
type ReasonStyle struct {
	Text string `json:"text,omitempty"`
	// 白天模式
	TextColor   string `json:"text_color,omitempty"`
	BgColor     string `json:"bg_color,omitempty"`
	BorderColor string `json:"border_color,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	// 夜间模式
	TextColorNight   string `json:"text_color_night,omitempty"`
	BgColorNight     string `json:"bg_color_night,omitempty"`
	BorderColorNight string `json:"border_color_night,omitempty"`
	IconURLNight     string `json:"icon_night_url,omitempty"`
	// 1:填充 2:描边 3:填充 + 描边 4:背景不填充 + 背景不描边
	BgStyle   int8        `json:"bg_style,omitempty"`
	URI       string      `json:"uri,omitempty"`
	IconBGURL string      `json:"icon_bg_url,omitempty"`
	Event     model.Event `json:"event,omitempty"`
	EventV2   model.Event `json:"event_v2,omitempty"`
	// 文案右边的小箭头，默认0或者字段不下发，右边的跳转箭头为白色，1是展示橙色橙色
	RightIconType int8 `json:"right_icon_type,omitempty"`
	// 运营角标日间、夜间 原始宽高、展示的高
	IconWidth  int32 `json:"icon_width,omitempty"`
	IconHeight int32 `json:"icon_height,omitempty"`
	TextLen    int8  `json:"text_len,omitempty"`
}

func topReasonStyleFrom(rcmd *ai.Item, text string, _ model.Gt, op *operate.Card) (res *ReasonStyle) {
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
		} else if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
			style = 6
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
	default:
		bgstyle = model.BgColorOrange
	}
	res = reasonStyleFrom(bgstyle, text)
	return
}

func topReasonStyleFromV2(rcmd *ai.Item, text string, _ model.Gt) (res *ReasonStyle) {
	if text == "" || rcmd == nil {
		return
	}
	bgstyle := model.BgColorFillingOrange
	res = reasonStyleFrom(bgstyle, text)
	return
}

func topReasonStyleFromV3(rcmd *ai.Item, text string, _ model.Gt) (res *ReasonStyle) {
	if text == "" || rcmd == nil {
		return
	}
	bgstyle := model.BgColorLumpOrange
	res = reasonStyleFrom(bgstyle, text)
	return
}

func reasonStyleFromV2(rcmd *ai.Item, text string, gt model.Gt, plat int8, build int) (res *ReasonStyle) {
	if rcmd.RcmdReason == nil || text == "" || rcmd.RcmdReason.JumpGoto == "" {
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
			res.IconURL = "https://i0.hdslb.com/bfs/archive/6983dc5b73d32a8241421a7b25f78c855b8e0362.png"
		case model.GotoPlaylist:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/b985518f35ad2b12eec34d4b3b6dca33df2b85a2.png"
		case model.GotoTag:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/c735ef9f33feb19f52bce852355e72a6b367c466.png"
		case model.GotoHotPage:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/e257e216b965905774b1ef9d526dfd2d8dae02f2.png"
		}
	case 1:
		//nolint:exhaustive
		switch model.Gt(rcmd.RcmdReason.JumpGoto) {
		case model.GotoAvConverge, model.GotoMultilayerConverge:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/8ba6d17e066f6ad3497e071abe654615fb073726.png"
		case model.GotoPlaylist:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/e0bd607cb58289fb32866e19ec207efae23657de.png"
		case model.GotoTag:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/0d1185d6ceca3de0e0a99bdce0479838265adcc3.png"
		case model.GotoHotPage:
			res.IconURL = "https://i0.hdslb.com/bfs/archive/d0f105544b1b0df77e8795ae297a72e627642b43.png"
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
	res.URI = model.FillURI(jumpGoto, 0, 0, strconv.FormatInt(rcmd.RcmdReason.JumpID, 10), urlf)
	return
}

func reasonStyleFromV3(rcmd *ai.Item, text string, gt model.Gt, plat int8, build int) (res *ReasonStyle) {
	if rcmd.RcmdReason == nil || text == "" || rcmd.RcmdReason.JumpGoto == "" {
		return
	}
	res = topReasonStyleFromV2(rcmd, text, gt)
	var (
		urlf             func(uri string) string
		_rightIconOrange = int8(1) // 带icon的推荐理由最右边的小箭头、展示橙色
	)
	res.RightIconType = _rightIconOrange
	//nolint:exhaustive
	switch model.Gt(rcmd.RcmdReason.JumpGoto) {
	case model.GotoAvConverge, model.GotoMultilayerConverge:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/c705fe617230bcfce0234c8a7323deb68350a209.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/6632d96c3504047c04a0118b5e6bae95241f7680.png"
	case model.GotoPlaylist:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/87182dd087b82d2e07928046d5a78311982ce04d.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/14e626a61674081aabf8e1ea77f2c9122077951a.png"
	case model.GotoTag:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/5df817cb2ca2c325534d940c3d1381bf6ac78258.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/1a3f717a993ae671b31c18b58c6d20a305097141.png"
	case model.GotoHotPage:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/5c34304eff82a432c8320a67d2977fece9d500a5.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/c303e397d707c84e736c31e8871f57206c7778b0.png"
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
	res.URI = model.FillURI(jumpGoto, 0, 0, strconv.FormatInt(rcmd.RcmdReason.JumpID, 10), urlf)
	return
}

func reasonStyleFromV4(rcmd *ai.Item, text string, gt model.Gt, plat int8, build int) (res *ReasonStyle) {
	if rcmd.RcmdReason == nil || text == "" || rcmd.RcmdReason.JumpGoto == "" {
		return
	}
	res = topReasonStyleFromV3(rcmd, text, gt)
	var (
		urlf             func(uri string) string
		_rightIconOrange = int8(1) // 带icon的推荐理由最右边的小箭头、展示橙色
	)
	res.RightIconType = _rightIconOrange
	//nolint:exhaustive
	switch model.Gt(rcmd.RcmdReason.JumpGoto) {
	case model.GotoAvConverge, model.GotoMultilayerConverge:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/09fea8d3ed60aae6f5f7e8147fce12379dbff726.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/d4c711c39d5c95b29f5730389536952bdb2b7e9c.png"
	case model.GotoPlaylist:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/b568780ef7310490d16a2c7a257f7542a98e59f0.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/e1a96822f3b6908253761fbdbadb65d354b26ebd.png"
	case model.GotoTag:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/703c23ab27abf70bd03167a9606c9ba1e9895caa.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/32499393ba3be99d342811bdf8ff597ac128fc77.png"
	case model.GotoHotPage:
		res.IconURL = "https://i0.hdslb.com/bfs/archive/c9ca993374ecef309c63044e1cd135977fb3ae88.png"
		res.IconURLNight = "https://i0.hdslb.com/bfs/archive/615ddf4dd574ab54d0cdb2b07de6fddad887b277.png"
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
	res.URI = model.FillURI(jumpGoto, 0, 0, strconv.FormatInt(rcmd.RcmdReason.JumpID, 10), urlf)
	return
}

func IsShowRcmdReasonStyleV2(rcmd *ai.Item) bool {
	if rcmd.RcmdReason != nil && rcmd.RcmdReason.JumpID > 0 {
		var style int8
		if style = rcmd.CornerMark; style == 0 {
			if rcmd.RcmdReason.Content != "" {
				style = rcmd.RcmdReason.CornerMark
			}
		}
		switch style {
		case 0, 1, 2:
			return true
		}
	}
	return false
}

func bottomReasonStyleFrom(rcmd *ai.Item, text string, _ model.Gt, op *operate.Card) (res *ReasonStyle) {
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
		} else if _, ok := op.SwitchStyle[model.SwitchNewReasonV2]; ok {
			style = 6
		}
	}
	switch style {
	case 1:
		bgstyle = model.BgColorTransparentOrange
	case 3:
		bgstyle = model.BgTransparentTextOrange
	case 5:
		bgstyle = model.BgColorFillingOrange
	case 6:
		bgstyle = model.BgColorLumpOrange
	default:
		bgstyle = model.BgColorOrange
	}
	res = reasonStyleFrom(bgstyle, text)
	return
}

func iconBadgeStyleFrom(v interface{}, style int8) (res *ReasonStyle) {
	//nolint:gosimple
	switch v.(type) {
	case *ReasonStyle:
		r := v.(*ReasonStyle)
		if r != nil && r.Text != "" {
			res = reasonStyleFrom(style, r.Text)
			res.IconURL = r.IconURL
		}
	case *operate.Card:
		o := v.(*operate.Card)
		if o.IconURL != "" && o.IconURLNight != "" && o.IconWidth > 0 && o.IconHeight > 0 {
			res = &ReasonStyle{
				IconURL:      o.IconURL,
				IconURLNight: o.IconURLNight,
				IconWidth:    o.IconWidth,
				IconHeight:   o.IconHeight,
			}
		}
	case *operate.LiveBottomBadge:
		l := v.(*operate.LiveBottomBadge)
		res = &ReasonStyle{
			Text:             l.Text,
			TextColor:        l.TextColor,
			BgColor:          l.BgColor,
			BorderColor:      l.BorderColor,
			IconURL:          l.IconURL,
			TextColorNight:   l.TextColorNight,
			BgColorNight:     l.BgColorNight,
			BorderColorNight: l.BorderColorNight,
			BgStyle:          l.BgStyle,
		}

	default:
		log.Warn("iconBadgeStyleFrom: unexpected type %T", v)
	}
	return
}

func reasonStyleFrom(style int8, text string) (res *ReasonStyle) {
	if text == "" {
		return
	}
	res = &ReasonStyle{
		Text: text,
	}
	switch style {
	case model.BgColorOrange: //defalut
		// 白天
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FFFB9E60"
		res.BorderColor = "#FFFB9E60"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#E5E5E5"
		res.BgColorNight = "#BC7A4F"
		res.BorderColorNight = "#BC7A4F"
	case model.BgColorTransparentOrange:
		// 白天
		res.TextColor = "#FFFB9E60"
		res.BorderColor = "#FFFB9E60"
		res.BgStyle = model.BgStyleStroke
		// 夜间
		res.TextColorNight = "#BC7A4F"
		res.BorderColorNight = "#BC7A4F"
	case model.BgColorBlue:
		res.TextColor = "#FF23ADE5"
		res.BgColor = "#3323ADE5"
		res.BorderColor = "#3323ADE5"
		res.BgStyle = model.BgStyleFill
	case model.BgColorRed:
		// 白天
		res.TextColor = "#FFFFFF"
		res.BgColor = "#FB7299"
		res.BorderColor = "#FB7299"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#E5E5E5"
		res.BgColorNight = "#BB5B76"
		res.BorderColorNight = "#BB5B76"
	case model.BgTransparentTextOrange:
		res.TextColor = "#FFFB9E60"
		res.BgStyle = model.BgStyleNoFillAndNoStroke
	case model.BgColorPurple:
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FF7D75F2"
		res.BorderColor = "#FF7D75F2"
		res.BgStyle = model.BgStyleFill
	case model.BgColorTransparentRed:
		// 白天
		res.TextColor = "#FB7299"
		res.BorderColor = "#FB7299"
		res.BgStyle = model.BgStyleStroke
		// 夜间
		res.TextColorNight = "#BB5B76"
		res.BorderColorNight = "#BB5B76"
	case model.BgColorFillingOrange:
		// 白天
		res.TextColor = "#FF6633"
		res.BgColor = "#FFF1ED"
		res.BorderColor = "#FFF1ED"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#BF5330"
		res.BgColorNight = "#BFB5B2"
		res.BorderColorNight = "#BFB5B2"
	case model.BgColorYellow:
		// 白天
		res.TextColor = "#FFFFFF"
		res.BgColor = "#FAAB4B"
		res.BorderColor = "#FAAB4B"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#E5E5E5"
		res.BgColorNight = "#BA833F"
		res.BorderColorNight = "#BA833F"
	case model.BgColorLumpOrange:
		// 白天
		res.TextColor = "#FF6633"
		res.BgColor = "#FFF1ED"
		res.BorderColor = "#FFF1ED"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#BF5330"
		res.BgColorNight = "#3D2D29"
		res.BorderColorNight = "#3D2D29"
	}
	return
}

func unionAuthorGRPC(a *arcgrpc.ArcPlayer, upName string) (name string) {
	if upName == "" {
		upName = a.Arc.Author.Name
	}
	if a.Arc.Rights.IsCooperation == 1 {
		name = upName + " 等联合创作"
		return
	}
	name = upName
	return
}

type LikeButton struct {
	Aid                  int64               `json:"aid,omitempty"`
	Count                int32               `json:"count,omitempty"`
	ShowCount            bool                `json:"show_count,omitempty"`
	Selected             int8                `json:"selected,omitempty"`
	Event                model.Event         `json:"event,omitempty"`
	EventV2              model.Event         `json:"event_v2,omitempty"`
	LikeResource         *LikeButtonResource `json:"like_resource,omitempty"`
	DisLikeResource      *LikeButtonResource `json:"dislike_resource,omitempty"`
	LikeNightResource    *LikeButtonResource `json:"like_night_resource,omitempty"`
	DisLikeNightResource *LikeButtonResource `json:"dislike_night_resource,omitempty"`
}

type LikeButtonResource struct {
	URL  string `json:"url,omitempty"`
	Hash string `json:"content_hash,omitempty"`
}

func likeButtonFromGRPC(a *arcgrpc.Arc, state int8, op *operate.Card) (button *LikeButton) {
	var (
		selected int8
	)
	switch state {
	case 1:
		selected = 1
	}
	button = &LikeButton{
		Aid:      a.Aid,
		Selected: selected,
		Count:    a.Stat.Like,
		Event:    model.EventlikeClick,
		EventV2:  model.EventV2likeClick,
	}
	if op != nil {
		button.ShowCount = op.LikeButtonShowCount
		button.LikeResource = likeButtonResourceFrom(op.LikeResource)
		button.DisLikeResource = likeButtonResourceFrom(op.DisLikeResource)
		button.LikeNightResource = likeButtonResourceFrom(op.LikeNightResource)
		button.DisLikeNightResource = likeButtonResourceFrom(op.DisLikeNightResource)
	}
	return
}

func likeButtonFromEpisodeCard(ep *pgcinline.EpisodeCard, state int8, op *operate.Card) (button *LikeButton) {
	var (
		selected int8
	)
	switch state {
	case 1:
		selected = 1
	}
	button = &LikeButton{
		Aid:      ep.Aid,
		Selected: selected,
		Count:    int32(ep.Stat.Like),
		Event:    model.EventlikeClick,
		EventV2:  model.EventV2likeClick,
	}
	if op != nil {
		button.ShowCount = op.LikeButtonShowCount
		button.LikeResource = likeButtonResourceFrom(op.LikeResource)
		button.DisLikeResource = likeButtonResourceFrom(op.DisLikeResource)
		button.LikeNightResource = likeButtonResourceFrom(op.LikeNightResource)
		button.DisLikeNightResource = likeButtonResourceFrom(op.DisLikeNightResource)
	}
	return
}

func likeButtonResourceFrom(b *operate.LikeButtonResource) (res *LikeButtonResource) {
	if b == nil {
		return
	}
	res = &LikeButtonResource{
		URL:  b.URL,
		Hash: b.Hash,
	}
	return
}

func (c *Base) MaskFromGRPC(a *arcgrpc.Arc) {
	if a == nil {
		return
	}
	c.Mask = &Mask{}
	c.Mask.Avatar = avatarFrom(&AvatarStatus{Cover: a.Author.Face, Text: a.Author.Name, Goto: model.GotoMid, Param: strconv.FormatInt(a.Author.Mid, 10), Type: model.AvatarRound})
	c.Mask.Button = buttonFrom(&ButtonStatus{Goto: model.GotoMid, Param: strconv.FormatInt(a.Author.Mid, 10), IsAtten: c.IsAttenm[a.Author.Mid]}, 0)
}

// ChannelBadge for new channel badge
type ChannelBadge struct {
	Text    string `json:"text,omitempty"`
	BgCover string `json:"icon_bg_url,omitempty"`
}

func GetBvIDStr(input string) (bid string, err error) {
	var aid int64
	if aid, err = GetAvID(input); err != nil {
		return "", fmt.Errorf("视频ID非法！")
	}
	if bid, err = bvid.AvToBv(aid); err != nil {
		return "", fmt.Errorf("视频ID非法！")
	}
	return
}

func GetAvID(input string) (aid int64, err error) {
	if aid, err = strconv.ParseInt(input, 10, 64); err != nil {
		if aid, err = bvid.BvToAv(input); err != nil {
			return 0, fmt.Errorf("视频ID非法！")
		}
	}
	return
}

func GetBvID(input int64) (bid string, err error) {
	if bid, err = bvid.AvToBv(input); err != nil {
		return "", fmt.Errorf("视频ID非法！")
	}
	return
}

func cvtTunnelResourceType(in string) string {
	if in == "game" {
		return "game_tunnel" // 曾经和客户端约定的游戏预约卡为 game_tunnel
	}
	return in
}

func GetShareSubtitle(view int32) (shareSubtitle, playNumber string) {
	tmp := strconv.FormatFloat(float64(view)/10000, 'f', 1, 64)
	if view > 100000 {
		shareSubtitle = "已观看" + strings.TrimSuffix(tmp, ".0") + "万次"
	}
	playNumber = strings.TrimSuffix(tmp, ".0") + "万次"
	return
}

// Tag .
type Tag struct {
	ID           int64      `json:"tag_id"`
	Name         string     `json:"tag_name"`
	Cover        string     `json:"cover"`
	HeadCover    string     `json:"head_cover"`
	Content      string     `json:"content"`
	ShortContent string     `json:"short_content"`
	Type         int8       `json:"type"`
	State        int8       `json:"state"`
	CTime        xtime.Time `json:"ctime"`
	MTime        xtime.Time `json:"-"`
	// tag count
	Count struct {
		View  int `json:"view"`
		Use   int `json:"use"`
		Atten int `json:"atten"`
	} `json:"count"`
	// subscriber
	IsAtten int8 `json:"is_atten"`
	// archive_tag
	Role      int8  `json:"-"`
	Likes     int64 `json:"likes"`
	Hates     int64 `json:"hates"`
	Attribute int8  `json:"attribute"`
	Liked     int8  `json:"liked"`
	Hated     int8  `json:"hated"`
	ExtraAttr int32 `json:"extra_attr"`
}

func CheckMidMaxInt32(mid int64) bool {
	if mid > math.MaxInt32 {
		return true
	}
	return false
}

func CheckMidMaxInt32Version(dev device.Device) bool {
	return (dev.RawMobiApp == "android" && dev.Build < 6500000) ||
		(dev.RawMobiApp == "iphone" && dev.Build < 65000000) ||
		(dev.RawMobiApp == "ipad" && dev.Build < 33000000) ||
		(dev.RawMobiApp == "iphone_i" && dev.Build < 65000000) ||
		(dev.RawMobiApp == "android_i" && dev.Build < 6500000) ||
		(dev.RawMobiApp == "android_b" && dev.Build < 6500000) ||
		(dev.RawMobiApp == "iphone_b" && dev.Build < 65000000) ||
		(dev.RawMobiApp == "android_hd" && dev.Build < 1070000)
}
func GetInlineProgressBar() *protov2.InlineProgressBar {
	return &protov2.InlineProgressBar{
		IconDrag:     model.InlineIconDrag,
		IconDragHash: model.InlineIconDragHash,
		IconStop:     model.InlineIconStop,
		IconStopHash: model.InlineIconStopHash,
	}
}
func ConstructSharePlane(in *arcgrpc.Arc) *protov2.SharePlane {
	shareSubtitle, playNumber := GetShareSubtitle(in.Stat.View)
	bvid, _ := GetBvID(in.Aid)
	return &protov2.SharePlane{
		Title:         in.Title,
		ShareSubtitle: shareSubtitle,
		Desc:          in.Desc,
		Cover:         in.Pic,
		Aid:           in.Aid,
		Bvid:          bvid,
		ShareTo:       model.ShareTo,
		Author:        in.Author.Name,
		AuthorId:      in.Author.Mid,
		ShortLink:     fmt.Sprintf(model.ShortLinkHost+"/av%d", in.Aid),
		PlayNumber:    playNumber,
		FirstCid:      in.FirstCid,
	}
}
