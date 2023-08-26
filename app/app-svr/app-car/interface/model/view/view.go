package view

import (
	"strconv"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/archive/service/api"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

const (
	// 详情页分P展示样式
	GridStyle       = "grid"
	HorizontalStyle = "horizontal"
	// 视频类型
	BiliType    = "bili"
	BangumiType = "bangumi"
)

type ViewParam struct {
	model.DeviceInfo
	Aid       int64  `form:"aid"`
	SeasonID  int64  `form:"season_id"`
	AccessKey string `form:"access_key"`
}

type View struct {
	Arc          *api.Arc      `json:"arc,omitempty"`
	PGC          *ViewPGC      `json:"pgc,omitempty"`
	VideoType    string        `json:"video_type,omitempty"`
	History      *History      `json:"history,omitempty"`
	Pages        []*Page       `json:"pages,omitempty"`
	PagesStyle   string        `json:"pages_style,omitempty"`
	Owner        *ViewOwner    `json:"owner,omitempty"`
	Button       *Button       `json:"button,omitempty"`
	ButtonV2     *Button       `json:"button_v2,omitempty"`
	QRCode       string        `json:"qr_code,omitempty"`
	Introduction *Introduction `json:"introduction,omitempty"`
	ReqUser      *ReqUser      `json:"req_user,omitempty"`
}

type ViewOwner struct {
	Mid  int64  `json:"mid,omitempty"`
	Name string `json:"name,omitempty"`
	Face string `json:"face,omitempty"`
	Fans string `json:"fans,omitempty"`
	URI  string `json:"uri,omitempty"`
	// h5
	RequestParam *card.RequestParam `json:"request_param,omitempty"`
	SourceType   string             `json:"source_type,omitempty"`
	Relation     *model.Relation    `json:"relation,omitempty"`
}

type ViewPGC struct {
	Title       string     `json:"title,omitempty"`
	SeasonTitle string     `json:"season_title,omitempty"`
	Cover       string     `json:"cover,omitempty"`
	Detail      string     `json:"detail,omitempty"`
	SeasonID    int64      `json:"season_id,omitempty"`
	BadgeInfo   *BadgeInfo `json:"badge_info,omitempty"`
	SeasonType  int        `json:"season_type,omitempty"`
	Stat        *ViewStat  `json:"stat,omitempty"`
}

type ViewStat struct {
	Reply int64 `json:"reply,omitempty"`
}

type History struct {
	Cid      int64 `json:"cid,omitempty"`
	Epid     int64 `json:"ep_id,omitempty"`
	Progress int64 `json:"progress,omitempty"`
	ViewAt   int64 `json:"view_at,omitempty"`
}

type Page struct {
	Aid       int64      `json:"aid,omitempty"`
	Cid       int64      `json:"cid,omitempty"`
	EpID      int64      `json:"ep_id,omitempty"`
	Desc      string     `json:"desc,omitempty"`
	Dimension *Dimension `json:"dimension,omitempty"`
	Duration  int64      `json:"duration,omitempty"`
	From      string     `json:"from,omitempty"`
	Title     string     `json:"title,omitempty"`
	Part      string     `json:"part,omitempty"`
	BadgeInfo *BadgeInfo `json:"badge_info,omitempty"`
	ShareURL  string     `json:"share_url,omitempty"`
	Stat      *ViewStat  `json:"stat,omitempty"`
}

type Dimension struct {
	Height int64 `json:"height,omitempty"`
	Width  int64 `json:"width,omitempty"`
	Rotate int64 `json:"rotate,omitempty"`
}

type BadgeInfo struct {
	Text      string `json:"text,omitempty"`
	TextColor string `json:"text_color,omitempty"`
	BgColor   string `json:"bg_color,omitempty"`
}

type Button struct {
	Type     string          `json:"type,omitempty"`
	Text     string          `json:"text,omitempty"`
	Relation *model.Relation `json:"relation,omitempty"`
	Selected int8            `json:"selected,omitempty"`
}

type Introduction struct {
	Title  string `json:"title,omitempty"`
	Info   string `json:"info,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Rating string `json:"rating,omitempty"`
	Cover  string `json:"cover,omitempty"`
}

type SilverEventCtx struct {
	Action      string `json:"action,omitempty"`
	Aid         int64  `json:"avid,omitempty"`
	UpID        int64  `json:"up_mid,omitempty"`
	Mid         int64  `json:"mid"`
	PubTime     string `json:"pubtime,omitempty"`
	LikeSource  string `json:"like_source,omitempty"`
	Buvid       string `json:"buvid,omitempty"`
	Ip          string `json:"ip,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Ctime       string `json:"ctime,omitempty"`
	Api         string `json:"api,omitempty"`
	Origin      string `json:"origin,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Build       string `json:"build,omitempty"`
	Token       string `json:"token,omitempty"`
	ItemType    string `json:"item_type,omitempty"`
	ShareSource string `json:"share_source,omitempty"`
	Title       string `json:"title,omitempty"`
	PlayNum     int64  `json:"play_num,omitempty"`
	CoinNum     int64  `json:"coin_num,omitempty"`
	*VipEventCtx
}

type VipEventCtx struct {
	SubScene    string `json:"subscene,omitempty"`
	ActivityUID string `json:"activity_uid,omitempty"`
}

type ReqUser struct {
	Like  int  `json:"like"` // 必须要下发
	IsFav bool `json:"is_fav,omitempty"`
}

func (p *Page) FromPageArc(v *api.Page) {
	p.Cid = v.Cid
	p.Desc = v.Desc
	p.Dimension = &Dimension{
		Width:  v.Dimension.Width,
		Height: v.Dimension.Height,
		Rotate: v.Dimension.Rotate,
	}
	p.Duration = v.Duration
	p.From = v.From
	p.Title = v.Part
	p.Part = v.Part
}

func (v *View) FromViewPGC(b *bangumi.View) {
	v.PGC = &ViewPGC{
		Title:       b.Title,
		SeasonTitle: b.SeasonTitle,
		Cover:       b.Cover,
		Detail:      b.Detail,
		SeasonID:    b.SeasonID,
		SeasonType:  b.Type,
	}
	// 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
	if b.Type == 1 || b.Type == 4 {
		// 如果当前PGC第一个分P是番剧类型则全部展示宫格
		v.PagesStyle = GridStyle
	}
	if b.BadgeInfo != nil && b.BadgeInfo.Text != "" {
		v.PGC.BadgeInfo = reasonStyleFrom(model.PGCBageType[b.BadgeType], b.Badge)
	}
	if b.UserStatus != nil && b.UserStatus.Progress != nil {
		v.History = &History{
			Epid:     b.UserStatus.Progress.LastEpID,
			Progress: b.UserStatus.Progress.LastTime,
		}
	}
	if b.Stat != nil {
		v.PGC.Stat = &ViewStat{
			Reply: b.Stat.Reply,
		}
	}
}

func (p *Page) FromPagePgc(e *bangumi.Episodes) {
	p.From = "bangumi"
	p.EpID = e.ID
	p.Aid = e.Aid
	p.Cid = e.Cid
	p.Dimension = &Dimension{
		Width:  e.Dimension.Width,
		Height: e.Dimension.Height,
		Rotate: e.Dimension.Rotate,
	}
	p.Title = e.Title
	p.Part = e.LongTitle
	if e.LongTitle == "" {
		p.Part = e.Title
	}
	if e.BadgeInfo != nil && e.BadgeInfo.Text != "" {
		p.BadgeInfo = reasonStyleFrom(model.PGCBageType[e.BadgeType], e.Badge)
	}
	p.ShareURL = model.FillURI(model.GotoWebPGC, 0, 0, strconv.FormatInt(e.ID, 10), nil)
	p.Stat = &ViewStat{
		Reply: e.Stat.Reply,
	}
	p.Duration = e.Duration
}

func (v *View) FromViewOwner(plat int8, build int, stat int64) {
	v.Arc.Access = 0
	v.Arc.Attribute = 0
	v.Arc.AttributeV2 = 0
	v.Owner = &ViewOwner{
		Mid:  v.Arc.Author.Mid,
		Name: v.Arc.Author.Name,
		Face: v.Arc.Author.Face,
		Fans: model.FanString(int32(stat)),
		URI:  model.FillURI(model.GotoSpace, plat, build, strconv.FormatInt(v.Arc.Author.Mid, 10), nil),
	}
}

func (v *View) FromButtonV2(gt string, selected int8, relations map[int64]*relationgrpc.InterrelationReply) {
	switch gt {
	case model.GotoUp:
		if v.Arc == nil {
			return
		}
		v.ButtonV2 = &Button{
			Type:     gt,
			Selected: selected,
			Relation: model.RelationChange(v.Arc.Author.Mid, relations),
		}
	case model.GotoPGC:
		v.ButtonV2 = &Button{
			Type:     gt,
			Selected: selected,
		}
	}
}

func (v *View) FromButtonArc(upMid int64, gt string, relations map[int64]*relationgrpc.InterrelationReply) {
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	// 2表示关注，6表示双向关注
	if rel.Attribute == 2 || rel.Attribute == 6 {
		return
	}
	v.Button = &Button{
		Type: gt,
		Text: model.ButtonText[gt],
	}
	if rel.IsFollowed {
		v.Button.Text = "互粉"
	}
}

func (v *View) FromButtonPgc(mid int64, state int8, gt string) {
	// 未登录已追番则不展示按钮
	if mid == 0 {
		return
	}
	if state != 0 {
		return
	}
	v.Button = &Button{
		Type: gt,
		Text: model.ButtonText[gt],
	}
}

func reasonStyleFrom(style string, text string) *BadgeInfo {
	if text == "" {
		return nil
	}
	res := &BadgeInfo{
		Text: text,
	}
	switch style {
	case model.BgColorRed:
		res.TextColor = "#FFFFFF"
		res.BgColor = "#FF5377"
	case model.BgColorBlue:
		res.TextColor = "#FFFFFF"
		res.BgColor = "#20AAE2"
	case model.BgColorYellow:
		res.TextColor = "#7E2D11"
		res.BgColor = "#FFB112"
	default:
		return nil
	}
	return res
}

func (i *Introduction) FromIntroductionArc(arc *api.Arc, desc string) {
	i.Title = arc.Title
	i.Desc = arc.Desc
	// 长简介
	if desc != "" {
		i.Desc = desc
	}
	splice := " "
	// 播放
	i.Info = model.ViewInfo(i.Info, model.StatString(arc.Stat.View, "播放"), splice)
	// 弹幕
	i.Info = model.ViewInfo(i.Info, model.StatString(arc.Stat.Danmaku, "弹幕"), splice)
	// 发布时间
	i.Info = model.ViewInfo(i.Info, model.PubDataString(arc.PubDate.Time()), splice)
	// bvid
	if bvid, err := model.GetBvID(arc.Aid); err == nil {
		i.Info = model.ViewInfo(i.Info, bvid, splice)
	}
}

func (i *Introduction) FromIntroductionPGC(b *bangumi.View) {
	i.Title = b.Title
	i.Cover = b.Cover
	splice1 := "\n"
	splice2 := " "
	splice3 := " | "
	var (
		desc1, desc2, desc3 string
	)
	if b.Publish != nil {
		// 发布时间
		if pubData, err := model.TimeToUnix(b.Publish.PubTime); err == nil {
			desc1 = model.ViewInfo(desc1, pubData.Format("2006"), splice2)
		}
	}
	// PGC类型
	if pgcType, err := model.PGCTypeValue(b.Type); err == nil {
		desc1 = model.ViewInfo(desc1, pgcType, splice2)
	}
	// 地区
	if len(b.Areas) > 0 {
		desc1 = model.ViewInfo(desc1, b.Areas[0].Name, splice3)
	}
	// 是否完结
	if b.Publish != nil {
		if b.Publish.IsFinish == 1 {
			desc2 = model.ViewInfo(desc2, model.BangumiTotalCountString(strconv.Itoa(b.Total), b.Publish.IsFinish), splice2)
		} else {
			if b.NewEP != nil {
				desc2 = model.ViewInfo(desc2, model.BangumiTotalCountString(b.NewEP.Title, b.Publish.IsFinish), splice2)
			}
		}
	}
	// 播放信息
	if b.Stat != nil {
		desc3 = model.ViewInfo(desc3, model.StatString64(b.Stat.Views, "播放"), splice2)
		if b.Stat.Followers != "" {
			// 有系列时，返回系列追番量，没有，则显示season追番量
			desc3 = model.ViewInfo(desc3, b.Stat.Followers, splice2)
		}
	}
	// 第一行
	i.Info = model.ViewInfo(i.Info, desc1, splice1)
	// 第二行
	i.Info = model.ViewInfo(i.Info, desc2, splice1)
	// 第三行
	i.Info = model.ViewInfo(i.Info, desc3, splice1)
	if b.Rating != nil {
		i.Rating = model.BangumiRating(b.Rating.Score, "")
	}
}
