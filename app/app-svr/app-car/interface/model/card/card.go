package card

import (
	"encoding/json"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"net/url"
	"strconv"
	"strings"

	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// Base is
type Base struct {
	// app
	CardType model.CardType `json:"card_type,omitempty"`
	CardGoto model.CardGt   `json:"card_goto,omitempty"`
	Goto     string         `json:"goto,omitempty"`
	Param    string         `json:"param,omitempty"`
	Bvid     string         `json:"bvid,omitempty"`
	Cover    string         `json:"cover,omitempty"`
	Title    string         `json:"title,omitempty"`
	URI      string         `json:"uri,omitempty"`
	Args     *Args          `json:"args,omitempty"`
	Pos      int            `json:"position,omitempty"`
	// h5
	Otype        string        `json:"otype,omitempty"`
	Oid          string        `json:"oid,omitempty"`
	Cid          string        `json:"cid,omitempty"`
	RequestParam *RequestParam `json:"request_param,omitempty"`
	SourceType   string        `json:"source_type,omitempty"`
	// ===============
	Rcmd      *ai.Item   `json:"-"`
	Materials *Materials `json:"-"`
	// ===============
	FromType string `json:"from_type,omitempty"`
	// 过滤原因
	Filter string `json:"-"`
}

type RequestParam struct {
	Keyword    string `json:"keyword,omitempty"`
	Param      string `json:"param,omitempty"`
	Cid        string `json:"cid,omitempty"`
	Rid        string `json:"rid,omitempty"`
	FollowType string `json:"follow_type,omitempty"`
	Vmid       int64  `json:"vmid,omitempty"`
	FavID      int64  `json:"fav_id,omitempty"`
}

type Page struct {
	Position       int    `json:"position,omitempty"`
	Pn             int    `json:"pn,omitempty"`
	Ps             int    `json:"ps,omitempty"`
	HistoryOffset  string `json:"history_offset,omitempty"`
	UpdateBaseline string `json:"update_baseline,omitempty"`
	UpdateNum      int64  `json:"update_num,omitempty"`
	Max            int64  `json:"max,omitempty"`
	MaxTP          int32  `json:"max_tp,omitempty"`
	Oid            int64  `json:"oid,omitempty"`
}

type CardParam struct {
	Plat         int8
	Mid          int64
	Pos          int
	FromType     string
	MobiApp      string
	Build        int
	IsPlayer     bool
	IsBackUpCard bool
}

type Materials struct {
	ViewReplym         map[int64]*arcgrpc.ViewReply
	EpisodeCardsProtom map[int32]*episodegrpc.EpisodeCardsProto
	Seams              map[int32]*seasongrpc.CardInfoProto
	Arcs               map[int64]*arcgrpc.Arc
	Epms               map[int32]*episodegrpc.EpisodeCardsProto
	ArcPlayers         map[int64]*arcgrpc.ArcPlayer
	Animem             map[int32]*cardappgrpc.CardSeasonProto
	Bangumim           map[int32]*bangumi.Module
	EpInlinem          map[int32]*pgcinline.EpisodeCard
	// 插入逻辑需要的字段
	Prune *Prune
}

func (c *Base) from(plat int8, build int, param, cid, cover, title string, gt string, uri string, f func(uri string) string) {
	c.URI = model.FillURI(gt, plat, build, uri, f)
	c.Cover = cover
	c.Title = title
	if gt != "" {
		c.Goto = gt
	} else {
		c.Goto = string(c.CardGoto)
	}
	c.Param = param
	c.Cid = cid
}

func (c *Base) fromH5(oid, cid, cover, title, gt string) {
	c.Cover = cover
	c.Title = title
	if gt != "" {
		c.Otype = gt
	} else {
		c.Otype = string(c.CardGoto)
	}
	c.Oid = oid
	c.Cid = cid
}

// Handler is
type Handler interface {
	From(main interface{}, op *operate.Card) bool
	Get() *Base
}

// Handle is
func Handle(plat int8, cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, materials *Materials) (hander Handler) {
	switch plat {
	case model.PlatH5:
		return h5Handle(cardGoto, cardType, rcmd, materials)
	}
	return singleHandle(cardGoto, cardType, rcmd, materials)
}

// Args is
type Args struct {
	Type   int8   `json:"type,omitempty"`
	UpID   int64  `json:"up_id,omitempty"`
	UpName string `json:"up_name,omitempty"`
	Aid    int64  `json:"aid,omitempty"`
}

// ReasonStyle reason style
type ReasonStyle struct {
	Text string `json:"text,omitempty"`
	// 白天模式
	TextColor   string `json:"text_color,omitempty"`
	BgColor     string `json:"bg_color,omitempty"`
	BorderColor string `json:"border_color,omitempty"`
	// fill：填充、stroke：描边、fill_stroke：填充+描边、no_fill_stroke：背景不填充 + 背景不描边
	BgStyle string `json:"bg_style,omitempty"`
}

func reasonStyleFrom(style string, text string) *ReasonStyle {
	if text == "" {
		return nil
	}
	res := &ReasonStyle{
		Text: text,
	}
	switch style {
	case model.BgColorRed:
		res.TextColor = "#FFFFFF"
		res.BgColor = "#FF5377"
		res.BorderColor = "#FF5377"
		res.BgStyle = model.BgStyleFill
	case model.BgColorBlue:
		res.TextColor = "#FFFFFF"
		res.BgColor = "#20AAE2"
		res.BorderColor = "#20AAE2"
		res.BgStyle = model.BgStyleFill
	case model.BgColorYellow:
		res.TextColor = "#7E2D11"
		res.BgColor = "#FFB112"
		res.BorderColor = "#FFB112"
		res.BgStyle = model.BgStyleFill
	default:
		return nil
	}
	return res
}

type HistoryArgs struct {
	Cid      int64  `json:"cid,omitempty"`
	Duration int64  `json:"duration,omitempty"`
	Mid      int64  `json:"mid,omitempty"`
	Name     string `json:"name,omitempty"`
	Page     int32  `json:"page,omitempty"`
	Progress int64  `json:"progress,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Business string `json:"business,omitempty"`
}

type HistoryDT struct {
	Icon string `json:"icon"`
	Type int32  `json:"type"`
}

type Prune struct {
	// PGC
	SeasonID int64 `json:"season_id,omitempty"`
	// His
	Business string `json:"business,omitempty"`
	Oid      int64  `json:"oid,omitempty"`
	Cid      int64  `json:"cid,omitempty"`
	Epid     int64  `json:"epid,omitempty"`
	// popular、search
	Goto    string `json:"goto,omitempty"`
	ID      int64  `json:"id,omitempty"`
	ChildID int64  `json:"child_id,omitempty"`
	// dynamic
	DynamicID int64 `json:"dynamic_id,omitempty"`
	Dtype     int64 `json:"dtype,omitempty"`
	Drid      int64 `json:"drid,omitempty"`
}

func (c *Base) FromRequestParam(main interface{}, cid, rid int64, entrance, followtype, keyword string) {
	c.RequestParam = &RequestParam{
		FollowType: followtype,
		Keyword:    keyword,
	}
	c.SourceType = entrance
	// 特殊参数用于进入接口列表做插入逻辑使用
	if main != nil {
		b, _ := json.Marshal(main)
		paramStr := url.QueryEscape(string(b))
		if strings.IndexByte(paramStr, '+') > -1 {
			paramStr = strings.Replace(paramStr, "+", "%20", -1)
		}
		c.RequestParam.Param = paramStr
	}
	if rid > 0 && entrance == model.EntranceRegion {
		c.RequestParam.Rid = strconv.FormatInt(rid, 10)
	}
}
