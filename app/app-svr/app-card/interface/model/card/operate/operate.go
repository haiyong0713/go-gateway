package operate

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"

	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
)

type Card struct {
	Plat                 int8                          `json:"plat,omitempty"`
	Build                int                           `json:"build,omitempty"`
	Network              string                        `json:"network,omitempty"`
	ID                   int64                         `json:"id,omitempty"`
	Param                string                        `json:"param,omitempty"`
	SubParam             string                        `json:"sub_param,omitempty"`
	CardGoto             model.CardGt                  `json:"card_goto,omitempty"`
	Goto                 model.Gt                      `json:"goto,omitempty"`
	URI                  string                        `json:"uri,omitempty"`
	RedirectURL          string                        `json:"redirect_url,omitempty"`
	Title                string                        `json:"title,omitempty"`
	Desc                 string                        `json:"desc,omitempty"`
	Cover                string                        `json:"cover,omitempty"`
	GifCover             string                        `json:"gif_cover,omitempty"`
	Coverm               map[model.ColumnStatus]string `json:"coverm,omitempty"`
	Avatar               string                        `json:"avatar,omitempty"`
	Download             int32                         `json:"download,omitempty"`
	Badge                string                        `json:"badge,omitempty"`
	Ratio                int                           `json:"ratio,omitempty"`
	Score                int32                         `json:"score,omitempty"`
	Tid                  int64                         `json:"tid,omitempty"`
	Subtitle             string                        `json:"subtitle,omitempty"`
	Limit                int                           `json:"limit,omitempty"`
	Items                []*Card                       `json:"items,omitempty"`
	AdInfo               *cm.AdInfo                    `json:"ad_info,omitempty"`
	Banner               []*banner.Banner              `json:"banner,omitempty"`
	Hash                 string                        `json:"verson,omitempty"`
	TrackID              string                        `json:"trackid,omitempty"`
	FromType             string                        `json:"from_type,omitempty"`
	ShowUGCPay           bool                          `json:"show_ucg_pay,omitempty"`
	Switch               model.Switch                  `json:"switch,omitempty"`
	SwitchLike           model.Switch                  `json:"switch_like,omitempty"`
	SwitchLargeCoverShow model.Switch                  `json:"switch_largecover_show,omitempty"`
	SwitchStyle          map[model.Switch]struct{}     `json:"switch_style,omitempty"`
	Buttons              []*Button                     `json:"buttons,omitempty"`
	MobiApp              string                        `json:"mobi_app,omitempty"`
	ShowHotword          bool                          `json:"show_hotword,omitempty"`
	Channel              *Channel                      `json:"channel,omitempty"`
	CanPlay              bool                          `json:"can_play,omitempty"`
	EntranceItems        []*model.EntranceItem         `json:"-"`
	Share                map[string]bool               `json:"-"`
	// 天马点赞按钮控制
	LikeButtonShowCount  bool                `json:"-"`
	LikeResource         *LikeButtonResource `json:"-"`
	DisLikeResource      *LikeButtonResource `json:"-"`
	LikeNightResource    *LikeButtonResource `json:"-"`
	DisLikeNightResource *LikeButtonResource `json:"-"`
	// 版本控制相关内容
	BuildLimit *BuildLimit `json:"_"`
	// 运营角标日间、夜间 url、原始宽高、展示的高
	IconURL                  string                    `json:"-"`
	IconURLNight             string                    `json:"-"`
	IconWidth                int32                     `json:"-"`
	IconHeight               int32                     `json:"-"`
	RoomID                   int64                     `json:"roomid,omitempty"`
	IsPopular                bool                      `json:"is_popular,omitempty"`
	SvideoShow               bool                      `json:"svideo_show,omitempty"`
	GotoIcon                 map[int64]*model.GotoIcon `json:"-"`
	HasFav                   map[int64]int8            `json:"-"`
	HotAidSet                sets.Int64                `json:"-"`
	ExtraURI                 string                    `json:"-"` // inline 自动定义跳转用字段
	InlinePlayIcon           InlinePlayIcon            `json:"-"`
	HasCoin                  map[int64]int64           `json:"-"`
	LiveLeftCoverBadgeStyle  []*V9LiveLeftCoverBadge   `json:"-"`
	LiveLeftBottomBadgeStyle *LiveBottomBadge          `json:"-"`
	// feature平台
	Feature                    *Feature         `json:"-"`
	InlineThreePoint           InlineThreePoint `json:"-"`
	NeedSwitchColumnThreePoint bool             `json:"-"`
	Column                     cdm.ColumnStatus `json:"-"`
	ReplaceDislikeTitle        bool             `json:"-"`
}

type InlineThreePoint struct {
	PanelType int
}

type Feature struct {
	c       context.Context
	keyName string
}

type LiveBottomBadge struct {
	Text             string `json:"text,omitempty"`
	TextColor        string `json:"text_color,omitempty"`
	BgColor          string `json:"bg_color,omitempty"`
	BorderColor      string `json:"border_color,omitempty"`
	TextColorNight   string `json:"text_color_night,omitempty"`
	BgColorNight     string `json:"bg_color_night,omitempty"`
	BorderColorNight string `json:"border_color_night,omitempty"`
	BgStyle          int8   `json:"bg_style,omitempty"`
	IconURL          string `json:"icon_url,omitempty"`
}

type V9LiveLeftCoverBadge struct {
	Key                  string `json:"key,omitempty"`
	IconBgURL            string `json:"icon_bg_url,omitempty"`
	Text                 string `json:"text,omitempty"`
	TextColor            string `json:"text_color,omitempty"`
	Priority             int64  `json:"priority,omitempty"`
	TextLen              int8   `json:"text_len,omitempty"`
	IconWidth            int32  `json:"icon_width,omitempty"`
	IconHeight           int32  `json:"icon_height,omitempty"`
	NewStyleIconURL      string `json:"new_icon_url,omitempty"`
	NewStyleIconURLNight string `json:"new_icon_url_night,omitempty"`
}

type InlinePlayIcon struct {
	IconDrag     string `json:"icon_drag,omitempty"`
	IconDragHash string `json:"icon_drag_hash,omitempty"`
	IconStop     string `json:"icon_stop,omitempty"`
	IconStopHash string `json:"icon_stop_hash,omitempty"`
}

type LikeButtonResource struct {
	URL  string `json:"url,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type BuildLimit struct {
	IsAndroid              bool
	IsIphone               bool
	IsIPad                 bool
	HotCardOptimizeAndroid int
	HotCardOptimizeIPhone  int
	HotCardOptimizeIPad    int
}

type Channel struct {
	ChannelID      int64                   `json:"channel_id,omitempty"`      // 新频道ID
	ChannelName    string                  `json:"channel_name,omitempty"`    // 新频道名字
	LastUpTime     int64                   `json:"last_update_ts,omitempty"`  // 最后一次更新时间
	UpCnt          int32                   `json:"updated_rs_num,omitempty"`  // 更新的优质资源数
	TodayCnt       int32                   `json:"today_rs_num,omitempty"`    // 今日投稿数
	FeatureCnt     int32                   `json:"featured_cnt,omitempty"`    // 精选视频数
	OfficiaLVerify int32                   `json:"official_verify,omitempty"` // 是否官方
	CType          int32                   `json:"ctype,omitempty"`           //  1:老频道/老tag, 2:新频道
	Badges         map[int64]*ChannelBadge `json:"badge,omitempty"`           // 角标字段
	CustomDesc     string                  `json:"custom_jump,omitempty"`     // 自定义卡片跳转描述
	CustomURI      string                  `json:"custom_uri,omitempty"`      // 自定义卡片跳转URI
	Coins          map[int64]int64         `json:"is_coin,omitempty"`         // 稿件投币状态
	IsFav          map[int64]bool          `json:"is_fav,omitempty"`          // 稿件收藏状态
	IsAtten        bool                    `json:"is_atten,omitempty"`        // 是否订阅
	AttenCnt       int32                   `json:"atten_cnt,omitempty"`       // 订阅数
	// special_channel manager/v2
	BgCover string `json:"bg_cover,omitempty"` // 底图
	Reason  string `json:"reason,omitempty"`   // 推荐理由
	TabURI  string `json:"tab_uri,omitempty"`  // 强制跳转地址
	// rank
	RankType int32 `json:"rank_type,omitempty"` // 排行榜类型:1播放;4收藏;5投币
	// infoc
	Position int64 `json:"position,omitempty"` // 上报用物料位置
	// switch
	HasMore bool `json:"has_more"` // OGV卡查看更多开关
	HasFold bool `json:"has_fold"` // OGV卡折叠开关
	// 上报字段
	Sort string `json:"sort"`
	Filt int32  `json:"filt"`
}

type ChannelBadge struct {
	Text  string `json:"text,omitempty"`
	Cover string `json:"cover,omitempty"`
}

type Button struct {
	Text  string `json:"text,omitempty"`
	Event string `json:"event,omitempty"`
}

func (c *Card) From(cardGoto model.CardGt, id int64, tid int64, plat int8, build int, mobiApp string) {
	c.CardGoto = cardGoto
	c.ID = id
	c.Tid = tid
	c.Goto = model.Gt(cardGoto)
	c.Param = strconv.FormatInt(id, 10)
	c.URI = strconv.FormatInt(id, 10)
	c.Plat = plat
	c.Build = build
	c.MobiApp = mobiApp
}

func (c *Card) FromDev(mobiApp string, plat int8, build int) {
	if c == nil {
		return
	}
	c.MobiApp = mobiApp
	c.Plat = plat
	c.Build = build
}

func (c *Card) FromSwitch(sw model.Switch) {
	c.SwitchLike = sw
}

func (c *Card) FromDownload(o *Download) {
	c.CardGoto = model.CardGotoDownload
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Coverm = map[model.ColumnStatus]string{model.ColumnSvrSingle: o.Cover, model.ColumnSvrDouble: o.DoubleCover}
	c.Title = o.Title
	c.Goto = model.OperateType[o.URLType]
	c.URI = o.URLValue
	c.Avatar = o.Icon
	c.Download = o.Number
	c.Desc = o.Desc
}

func (c *Card) FromSpecial(o *Special) {
	const (
		// https://www.tapd.cn/20095661/prong/stories/view/1120095661001172604 按照客户端的设计图里面的固定高
		_showHeight = float64(21)
	)
	c.ID, _ = strconv.ParseInt(o.ReValue, 10, 64)
	c.CardGoto = model.CardGotoSpecial
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Coverm = map[model.ColumnStatus]string{model.ColumnSvrSingle: o.SingleCover, model.ColumnSvrDouble: o.Cover}
	c.GifCover = o.GifCover
	c.Title = o.Title
	c.Goto = model.OperateType[o.ReType]
	c.URI = o.ReValue
	c.Desc = o.Desc
	c.Badge = o.Badge
	c.IconURL = o.PowerPicSun
	c.IconURLNight = o.PowerPicNight
	if o.Size == "1020x300" {
		c.Ratio = 34
	} else if o.Size == "1020x378" {
		c.Ratio = 27
	}
	if o.PowerPicWidth > 0 && o.PowerPicHeight > 0 {
		c.IconHeight = int32(_showHeight)
		c.IconWidth = int32(math.Floor((o.PowerPicWidth / o.PowerPicHeight) * _showHeight))
	}
}

func (c *Card) FromAppSpecialCard(o *resourceV2grpc.AppSpecialCard) {
	const (
		// https://www.tapd.cn/20095661/prong/stories/view/1120095661001172604 按照客户端的设计图里面的固定高
		_showHeight = float64(21)
	)
	c.ID, _ = strconv.ParseInt(o.ReValue, 10, 64)
	c.CardGoto = model.CardGotoSpecial
	c.Param = strconv.FormatInt(o.Id, 10)
	c.Coverm = map[model.ColumnStatus]string{model.ColumnSvrSingle: o.Scover, model.ColumnSvrDouble: o.Cover}
	c.GifCover = o.Gifcover
	c.Title = o.Title
	c.Goto = model.OperateType[int(o.ReType)]
	c.URI = o.ReValue
	c.Desc = o.Desc
	c.Badge = o.Corner
	c.IconURL = o.PowerPicSun
	c.IconURLNight = o.PowerPicNight
	if o.Size_ == "1020x300" {
		c.Ratio = 34
	} else if o.Size_ == "1020x378" {
		c.Ratio = 27
	}
	if o.Width > 0 && o.Height > 0 {
		c.IconHeight = int32(_showHeight)
		c.IconWidth = int32(math.Floor(float64(o.Width) / float64(o.Height) * _showHeight))
	}
}

func (c *Card) FromSpecialB(o *Special) {
	const (
		// https://www.tapd.cn/20095661/prong/stories/view/1120095661001172604 按照客户端的设计图里面的固定高
		_showHeight = float64(21)
	)
	c.ID, _ = strconv.ParseInt(o.ReValue, 10, 64)
	c.CardGoto = model.CardGotoSpecialB
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Cover = o.Cover
	c.GifCover = o.GifCover
	c.Title = o.Title
	c.Goto = model.OperateType[o.ReType]
	c.URI = o.ReValue
	c.Desc = o.Desc
	c.Badge = o.Badge
	c.IconURL = o.PowerPicSun
	c.IconURLNight = o.PowerPicNight
	if o.PowerPicWidth > 0 && o.PowerPicHeight > 0 {
		c.IconHeight = int32(_showHeight)
		c.IconWidth = int32(math.Floor((o.PowerPicWidth / o.PowerPicHeight) * _showHeight))
	}
}

func (c *Card) FromSpecialChannel(o *Special) {
	c.ID, _ = strconv.ParseInt(o.ReValue, 10, 64)
	c.CardGoto = model.CardGotoSpecialChannel
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Title = o.Title
	c.Cover = o.Cover
	c.Desc = o.Desc
	c.Goto = model.GotoChannel
	c.Channel = &Channel{
		BgCover: o.BgCover,
		Reason:  o.Reason,
		TabURI:  o.TabURI,
	}
}

func (c *Card) FromTopstick(o *Special) {
	c.CardGoto = model.CardGotoTopstick
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Title = o.Title
	c.Goto = model.OperateType[o.ReType]
	c.URI = o.ReValue
	c.Desc = o.Desc
	c.Badge = o.Badge
}

func (c *Card) FromFollow(o *Follow) {
	switch o.Type {
	case "upper", "channel_three":
		var contents []*struct {
			Ctype  string `json:"ctype,omitempty"`
			Cvalue int64  `json:"cvalue,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &contents); err != nil {
			log.Error("%+v", err)
			return
		}
		items := make([]*Card, 0, len(contents))
		for _, content := range contents {
			var gt model.Gt
			switch content.Ctype {
			case "mid":
				gt = model.GotoMid
			case "channel_id":
				gt = model.GotoTag
			default:
				continue
			}
			items = append(items, &Card{ID: content.Cvalue, Goto: gt, Param: strconv.FormatInt(content.Cvalue, 10), URI: strconv.FormatInt(content.Cvalue, 10)})
		}
		if len(items) < 3 {
			return
		}
		c.Items = items
		c.CardGoto = model.CardGotoSubscribe
		c.Title = o.Title
		c.Param = strconv.FormatInt(o.ID, 10)
	case "channel_single":
		var content struct {
			Aid       int64 `json:"aid"`
			ChannelID int64 `json:"channel_id"`
		}
		if err := json.Unmarshal(o.Content, &content); err != nil {
			log.Error("%+v", err)
			return
		}
		c.CardGoto = model.CardGotoChannelRcmd
		c.Title = o.Title
		c.ID = content.Aid
		c.Tid = content.ChannelID
		c.Goto = model.GotoAv
		c.Param = strconv.FormatInt(o.ID, 10)
		c.URI = strconv.FormatInt(content.Aid, 10)
	}
}

func (c *Card) FromConverge(o *Converge) {
	c.CardGoto = model.CardGotoConverge
	c.Param = strconv.FormatInt(o.ID, 10)
	c.Coverm = map[model.ColumnStatus]string{model.ColumnSvrSingle: o.Cover, model.ColumnSvrDouble: o.Cover}
	c.Title = o.Title
	c.Goto = model.OperateType[o.ReType]
	c.URI = o.ReValue
	var contents []*struct {
		Ctype  string `json:"ctype,omitempty"`
		Cvalue string `json:"cvalue,omitempty"`
	}
	if err := json.Unmarshal(o.Content, &contents); err != nil {
		log.Error("%+v", err)
		return
	}
	c.Items = make([]*Card, 0, len(contents))
	for _, content := range contents {
		var (
			gt     model.Gt
			cardGt model.CardGt
		)
		id, _ := strconv.ParseInt(content.Cvalue, 10, 64)
		if id == 0 {
			continue
		}
		switch content.Ctype {
		case "0":
			gt = model.GotoAv
			cardGt = model.CardGotoAv
		case "1":
			gt = model.GotoLive
			cardGt = model.CardGotoLive
		case "2":
			gt = model.GotoArticle
			cardGt = model.CardGotoArticleS
		default:
			continue
		}
		c.Items = append(c.Items, &Card{ID: id, CardGoto: cardGt, Goto: gt, Param: content.Cvalue, URI: content.Cvalue})
	}
}

func (c *Card) FromRank(os []*rank.Rank) {
	c.CardGoto = model.CardGotoRank
	c.Goto = model.GotoRank
	c.Items = make([]*Card, 0, len(os))
	for _, o := range os {
		c.Items = append(c.Items, &Card{Goto: model.GotoAv, ID: o.Aid, Param: strconv.FormatInt(o.Aid, 10), URI: strconv.FormatInt(o.Aid, 10), Score: o.Score})
	}
}

func (c *Card) FromActive(o *Active) {
	switch o.Type {
	case "live", "player_live", "converge", "special", "archive", "player":
		var id int64
		if err := json.Unmarshal(o.Content, &id); err != nil {
			log.Error("%+v", err)
			return
		}
		if id < 1 {
			return
		}
		c.ID = id
		c.Param = strconv.FormatInt(id, 10)
		switch o.Type {
		case "live":
			c.CardGoto = model.CardGotoPlayerLive
		case "converge":
			c.CardGoto = model.CardGotoConverge
		case "special":
			c.CardGoto = model.CardGotoSpecial
		case "archive":
			c.CardGoto = model.CardGotoPlayer
		}
	case "basic", "content_rcmd":
		var basic struct {
			Type     string `json:"type,omitempty"`
			Title    string `json:"title,omitempty"`
			Subtitle string `json:"subtitle,omitempty"`
			Sublink  string `json:"sublink,omitempty"`
			Content  []*struct {
				LinkType  string `json:"link_type,omitempty"`
				LinkValue string `json:"link_value,omitempty"`
			} `json:"content,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &basic); err != nil {
			log.Error("%+v", err)
			return
		}
		items := make([]*Card, 0, len(basic.Content))
		for _, c := range basic.Content {
			typ, _ := strconv.Atoi(c.LinkType)
			id, _ := strconv.ParseInt(c.LinkValue, 10, 64)
			ri := &Card{Goto: model.OperateType[typ], ID: id, Param: c.LinkValue}
			if ri.Goto != "" {
				items = append(items, ri)
			}
		}
		if len(items) == 0 {
			return
		}
		c.Items = items
		c.Title = basic.Title
		c.Subtitle = basic.Subtitle
		c.URI = basic.Sublink
		c.CardGoto = model.CardGotoContentRcmd
	case "shortcut", "entrance", "banner":
		var card struct {
			Type     string      `json:"type,omitempty"`
			CardItem []*CardItem `json:"card_item,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &card); err != nil {
			log.Error("%+v", err)
			return
		}
		items := make([]*Card, 0, len(card.CardItem))
		sort.Sort(CardItems(card.CardItem))
		for _, v := range card.CardItem {
			typ, _ := strconv.Atoi(v.LinkType)
			id, _ := strconv.ParseInt(v.LinkValue, 10, 64)
			item := &Card{Goto: model.OperateType[typ], ID: id, Param: v.LinkValue, URI: v.LinkValue, Title: v.Title, Cover: v.Cover}
			if item.Goto != "" {
				items = append(items, item)
			}
		}
		if len(items) == 0 {
			return
		}
		c.Items = items
		switch o.Type {
		case "shortcut", "entrance":
			c.CardGoto = model.CardGotoEntrance
		case "banner":
			c.CardGoto = model.CardGotoBanner
		}
	case "common", "background":
		c.Title = o.Name
		c.Cover = o.Background
	case "tag", "tag_rcmd":
		var tag struct {
			AidStr    string `json:"aid,omitempty"`
			Type      string `json:"type,omitempty"`
			NumberStr string `json:"number,omitempty"`
			Tid       int64  `json:"-"`
			Number    int    `json:"-"`
		}
		if err := json.Unmarshal(o.Content, &tag); err != nil {
			log.Error("%+v", err)
			return
		}
		tag.Tid, _ = strconv.ParseInt(tag.AidStr, 10, 64)
		tag.Number, _ = strconv.Atoi(tag.NumberStr)
		if tag.Tid == 0 {
			return
		}
		c.ID = tag.Tid
		c.Limit = tag.Number
		c.Goto = model.GotoTag
		c.CardGoto = model.CardGotoTagRcmd
		c.Subtitle = "查看更多"
	case "news":
		var news struct {
			Title string `json:"title,omitempty"`
			Body  string `json:"body,omitempty"`
			Link  string `json:"link,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &news); err != nil {
			log.Error("%+v", err)
			return
		}
		if news.Body == "" {
			return
		}
		c.Title = news.Title
		c.Desc = news.Body
		c.URI = news.Link
		c.Goto = model.GotoWeb
		c.CardGoto = model.CardGotoNews
	case "vip":
		c.Goto = model.GotoVip
		c.CardGoto = model.CardGotoVip
	}
	c.Title = o.Title
	c.Param = strconv.FormatInt(o.ID, 10)
}

func (c *Card) FromAdAv(o *cm.AdInfo) {
	c.CardGoto = model.CardGotoAdAv
	c.AdInfo = o
}

func (c *Card) FromAdLive(o *cm.AdInfo) {
	c.CardGoto = model.CardGotoAdLive
	c.AdInfo = o
}

func (c *Card) FromAdLiveInLine(o *cm.AdInfo) {
	c.CardGoto = model.CardGotoAdInlineLive
	c.AdInfo = o
}

func (c *Card) FromActiveBanner(os []*Active, hash string, isNewBanner bool) (cardType cdm.CardType) {
	c.Banner = make([]*banner.Banner, 0, len(os))
	for _, o := range os {
		var tmpCover string
		if isNewBanner && o.BigCover != "" {
			tmpCover = o.BigCover
			cardType = cdm.BannerV5
		} else if o.Cover != "" {
			tmpCover = o.Cover
		}
		if tmpCover == "" {
			continue
		}
		banner := &banner.Banner{ID: o.Pid, Title: o.Title, Image: tmpCover, URI: model.FillURI(o.Goto, 0, 0, o.Param, nil)}
		c.Banner = append(c.Banner, banner)
	}
	c.CardGoto = model.CardGotoBanner
	c.Hash = hash
	return
}

func (c *Card) FromBanner(os []*banner.Banner, hash string) {
	if len(os) == 0 {
		return
	}
	c.Banner = os
	c.CardGoto = model.CardGotoBanner
	c.Hash = hash
}

func (c *Card) FromCardSet(o *CardSet) {
	switch o.Type {
	case "pgcs_rcmd":
		var contents []*struct {
			ID interface{} `json:"id,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &contents); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, content := range contents {
			var cid int64
			switch v := content.ID.(type) {
			case string:
				cid, _ = strconv.ParseInt(v, 10, 64)
			case float64:
				cid = int64(v)
			}
			item := &Card{ID: cid, Goto: model.GotoPGC}
			c.Items = append(c.Items, item)
		}
		c.Title = o.Title
		c.Param = strconv.FormatInt(o.ID, 10)
		c.CardGoto = model.CardGotoPgcsRcmd
	case "up_rcmd_new", "up_rcmd_new_single":
		var contents []*struct {
			ID interface{} `json:"id,omitempty"`
		}
		if err := json.Unmarshal(o.Content, &contents); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, content := range contents {
			var aid int64
			switch v := content.ID.(type) {
			case string:
				aid, _ = strconv.ParseInt(v, 10, 64)
			case float64:
				aid = int64(v)
			}
			item := &Card{ID: aid, Goto: model.GotoAv}
			c.Items = append(c.Items, item)
		}
		c.Title = "新星卡片"
		c.Desc = o.Title
		c.Param = strconv.FormatInt(o.Value, 10)
		c.ID = o.Value
		switch o.Type {
		case "up_rcmd_new":
			c.CardGoto = model.CardGotoUpRcmdNew
		case "up_rcmd_new_single":
			c.CardGoto = model.CardGotoUpRcmdSingle
		}
	}
}

func (c *Card) FromFollowMode(title, desc string, button []string) {
	c.Title = title
	if c.Title == "" {
		c.Title = "启用首页推荐 - 关注模式（内测版）"
	}
	c.Desc = desc
	if c.Desc == "" {
		c.Desc = "我们根据你对bilibili推荐的反馈，为你定制了关注模式。开启后，仅为你显示关注UP主更新的视频哦。尝试体验一下？"
	}
	if len(button) == 2 {
		c.Buttons = []*Button{
			{Text: button[0], Event: "close"},
			{Text: button[1], Event: "follow_mode"},
		}
	} else {
		c.Buttons = []*Button{
			{Text: "暂不需要", Event: "close"},
			{Text: "立即开启", Event: "follow_mode"},
		}
	}
	c.CardGoto = model.CardGotoFollowMode
}

func (c *Card) FromEventTopic(o *EventTopic) {
	if o.ShowTitle == 1 {
		c.Title = o.Title
	}
	c.Desc = o.Desc
	c.Cover = o.Cover
	switch o.ReType {
	case 1:
		c.Goto = model.Gt("topic")
	case 2:
		c.Goto = model.Gt("broadcast")
	case 3:
		c.Goto = model.Gt("channel")
	}
	c.Param = strconv.FormatInt(o.ID, 10)
	c.URI = o.ReValue
	c.Badge = o.Corner
}

func (c *Card) FromConvergeAi(o *ai.ConvergeInfo, id int64) {
	c.CardGoto = model.CardGotoConvergeAi
	c.Param = strconv.FormatInt(id, 10)
	c.Title = o.Title
	c.Goto = model.Gt(model.CardGotoConverge)
	// c.Goto = gt
	for _, item := range o.Items {
		var (
			cardGt model.CardGt
			gt     = model.Gt(item.Goto)
		)
		switch model.Gt(item.Goto) {
		case model.GotoAv:
			cardGt = model.CardGotoAv
		default:
			continue
		}
		c.Items = append(c.Items, &Card{ID: item.ID, CardGoto: cardGt, Goto: gt, Param: strconv.FormatInt(item.ID, 10), URI: strconv.FormatInt(item.ID, 10)})
	}
}

func (c *Card) FromVipRenew(v *vipgrpc.TipsRenewReply) {
	if v == nil || v.Title == "" {
		return
	}
	if v.ImgUri != "" {
		c.Cover = v.ImgUri
	} else {
		c.Cover = "https://i0.hdslb.com/bfs/archive/1b8deb69e4a9effc8f3be24107f925480afe3ade.png"
	}
	c.Title = strings.Replace(v.Title, "[", " \u003cem class=\"keyword\"\u003e", 1)
	c.Title = strings.Replace(c.Title, "]", "\u003c/em\u003e ", 1)
}

func (c *Card) FromAvConverge(o *ai.Item, aid int64) {
	if (o.JumpID == 0 && o.JumpGoto != string(model.GotoHotPage)) || o.Goto == "" {
		c.Param = strconv.FormatInt(o.ID, 10)
		c.Goto = model.GotoAvConverge
	} else {
		c.Param = strconv.FormatInt(o.JumpID, 10)
		c.Goto = model.Gt(o.JumpGoto)
	}
	c.ID = aid
	c.TrackID = o.TrackID
}

func (c *Card) FromAvConvergeCard(o *Card) {
	c.Param = o.Param
	c.ID = o.ID
	c.Goto = o.Goto
	c.TrackID = o.TrackID
}
