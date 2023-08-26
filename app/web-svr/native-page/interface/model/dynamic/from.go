package dynamic

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/web-svr/native-page/interface/api"
	gamdl "go-gateway/app/web-svr/native-page/interface/model/game"
	lmdl "go-gateway/app/web-svr/native-page/interface/model/like"
	"go-gateway/app/web-svr/native-page/interface/model/pgc"
	bvsafe "go-gateway/pkg/idsafe/bvid"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	xtime "go-common/library/time"

	arccli "go-gateway/app/app-svr/archive/service/api"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	"go-common/library/log"

	bvidmdl "go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

// Item .
type Item struct {
	Goto   string `json:"goto,omitempty"`
	Type   int64  `json:"type,omitempty"`
	Param  string `json:"param,omitempty"`
	ItemID int64  `json:"item_id,omitempty"`
	Ukey   string `json:"ukey,omitempty"`
	// click
	Width      int64   `json:"width,omitempty"`
	Length     int64   `json:"length,omitempty"`
	Image      string  `json:"image,omitempty"`
	UnImage    string  `json:"un_image,omitempty"`
	Leftx      int64   `json:"leftx,omitempty"`
	Lefty      int64   `json:"lefty,omitempty"`
	URI        string  `json:"uri,omitempty"`
	IosURI     string  `json:"ios_uri,omitempty"`
	AndroidURI string  `json:"android_uri,omitempty"`
	Content    string  `json:"content,omitempty"`
	Subtitle   string  `json:"subtitle,omitempty"`
	Title      string  `json:"title,omitempty"`
	Item       []*Item `json:"item,omitempty"`
	Bar        string  `json:"bar,omitempty"`
	// 动态卡片接口 数据透传
	DyCard *DyCard `json:"dy_card,omitempty"`
	//点赞相关信息
	Liked           *Liked        `json:"liked,omitempty"`
	IsGap           int32         `json:"is_gap,omitempty"`
	IsFeed          int64         `json:"is_feed,omitempty"`
	Button          *Button       `json:"button,omitempty"`
	CoverLeftText1  string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1  string        `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2  string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2  string        `json:"cover_left_icon_2,omitempty"`
	CoverRightText  string        `json:"cover_right_text,omitempty"`
	Badge           *ReasonStyle  `json:"badge,omitempty"`
	Repost          *Repost       `json:"repost,omitempty"`
	Fid             int64         `json:"fid,omitempty"`
	Duration        string        `json:"duration,omitempty"`
	Danmaku         string        `json:"danmaku,omitempty"`
	View            string        `json:"view,omitempty"`
	Dimension       *Dimension    `json:"dimension,omitempty"`
	ResourceInfo    *ResourceInfo `json:"resource_info,omitempty"`
	Stime           int64         `json:"stime,omitempty"`
	Setting         *Setting      `json:"setting,omitempty"`
	Color           *Color        `json:"color,omitempty"`
	ImagesUnion     *ImagesUnion  `json:"images_union,omitempty"`
	ClickExt        *ClickExt     `json:"click_ext,omitempty"`
	FontSize        int64         `json:"font_size,omitempty"`
	FontType        string        `json:"font_type,omitempty"`
	Num             int64         `json:"num,omitempty"`
	TargetNum       int64         `json:"target_num,omitempty"`
	CurrentTab      string        `json:"current_tab,omitempty"`
	SimpleImages    *SimpleImages `json:"simple_images,omitempty"`
	Style           string        `json:"style,omitempty"`
	Images          []*api.Image  `json:"images,omitempty"`
	UrlExt          *UrlExt       `json:"url_ext,omitempty"`
	Meta            string        `json:"meta,omitempty"`
	Caption         string        `json:"caption,omitempty"`
	Attr            int64         `json:"attr,omitempty"`
	Positions       *Positions    `json:"positions,omitempty"`
	RcmdContent     *RcmdContent  `json:"rcmd_content,omitempty"`
	DisplayType     string        `json:"display_type,omitempty"`
	UserInfo        *UserInfo     `json:"user_info,omitempty"`
	ScrollType      int32         `json:"scroll_type,omitempty"`
	BgStyle         int64         `json:"background_style,omitempty"`
	IndicatorStyle  int64         `json:"indicator_style,omitempty"`
	MutexUkeys      []string      `json:"mutex_ukeys,omitempty"`
	ButtonType      string        `json:"button_type,omitempty"`
	CurrentTabIndex int32         `json:"current_tab_index,omitempty"`
	Time            string        `json:"time,omitempty"`
	SponsorTitle    string        `json:"sponsor_title,omitempty"`
	NewactFeatures  []*Item       `json:"newact_features,omitempty"`
	TableAttrs      []*TableAttr  `json:"table_attrs,omitempty"`
	Header          *Item         `json:"header,omitempty"`
	Status          string        `json:"status,omitempty"`
}

type TableAttr struct {
	Ratio     int64  `json:"ratio,omitempty"`
	TextAlign string `json:"text_align,omitempty"`
}

type UserInfo struct {
	Mid          int64                `json:"mid,omitempty"`
	Name         string               `json:"name,omitempty"`
	Face         string               `json:"face,omitempty"`
	OfficialInfo accgrpc.OfficialInfo `json:"official_info,omitempty"`
	Vip          accgrpc.VipInfo      `json:"vip,omitempty"`
}

func (myinfo *UserInfo) formatRole() {
	roleSpe := int32(7)
	if myinfo.OfficialInfo.Role == roleSpe {
		myinfo.OfficialInfo.Role = 1
	}
}

// UrlExt .
type UrlExt struct {
	Fid          int64           `json:"fid,omitempty"`
	Type         int32           `json:"type,omitempty"`
	SortType     int64           `json:"sort_type,omitempty"`
	Num          int64           `json:"num,omitempty"`
	Category     int64           `json:"category,omitempty"`
	IDs          []*ResourcesIDs `json:"ids,omitempty"`
	HasMore      bool            `json:"has_more,omitempty"`
	Types        string          `json:"types,omitempty"`
	SourceID     string          `json:"source_id,omitempty"`
	SeasonID     int64           `json:"season_id,omitempty"`
	ConfModuleID int64           `json:"conf_module_id,omitempty"`
	Sid          int64           `json:"sid,omitempty"`
	Counter      string          `json:"counter,omitempty"`
}

type ResourcesIDs struct {
	ID          int64        `json:"id,omitempty"`
	Type        int32        `json:"type,omitempty"`
	FID         int64        `json:"fid,omitempty"`
	RcmdContent *RcmdContent `json:"rcmd_content,omitempty"` //编辑推荐内容
}

type SimpleImages struct {
	ButtonImage string `json:"button_image,omitempty"`
	LayerImage  string `json:"layer_image,omitempty"`
}

type Positions struct {
	Position1 Position `json:"position1"` //属性展示位置1
	Position2 Position `json:"position2"` //属性展示位置2
	Position3 Position `json:"position3"` //属性展示位置3
	Position4 Position `json:"position4"` //属性展示位置4
	Position5 Position `json:"position5"` //属性展示位置5
}

type Position string

func (p *Position) FormatRankArc(typ string, rankItem *actGRPC.RankResult, arc *actGRPC.ArchiveInfo) {
	switch typ {
	case api.PosComprehensive:
		*p = Position(rankItem.ShowScore)
	case api.PosLike:
		*p = Position(statString(int64(arc.Like), "点赞"))
	case api.PosView:
		*p = Position(statString(int64(arc.View), "观看"))
	case api.PosShare:
		*p = Position(statString(int64(arc.Share), "分享"))
	case api.PosCoin:
		*p = Position(statString(int64(arc.Coin), "投币"))
	default:
		*p = ""
	}
}

type MixFolder struct {
	Fid         int64        `json:"fid,omitempty"`
	RcmdContent *RcmdContent `json:"rcmd_content,omitempty"` //编辑推荐内容
}

type RcmdContent struct {
	TopContent      string `json:"top_content,omitempty"`    //顶部推荐语
	TopFontColor    string `json:"top_font_color,omitempty"` //顶部字体颜色
	TopIcon         string `json:"top_icon,omitempty"`
	BottomContent   string `json:"bottom_content,omitempty"`    //底部推荐语
	BottomFontColor string `json:"bottom_font_color,omitempty"` //底部字体颜色
	BottomIcon      string `json:"bottom_icon,omitempty"`
	MiddleIcon      string `json:"middle_icon,omitempty"` //排行榜icon
}

// ClickExt .
type ClickExt struct {
	FID        int64        `json:"fid,omitempty"`
	Tip        string       `json:"tip,omitempty"`
	Images     *ImagesUnion `json:"images,omitempty"`
	RuleID     int64        `json:"rule_id,omitempty"`
	Dimension  int64        `json:"dimension,omitempty"`
	CurrentNum int64        `json:"current_num,omitempty"`
	DisplayNum string       `json:"display_num,omitempty"`
	Style      string       `json:"style,omitempty"`
	IsFollow   bool         `json:"is_follow,omitempty"`
	TargetNum  int64        `json:"target_num,omitempty"`
	BgColor    string       `json:"bg_color,omitempty"`
	IsDisplay  bool         `json:"is_display,omitempty"`
	// 当前状态 0 无资格或无奖励，1 未领取，2 已领取
	CurrentState int8 `json:"current_state,omitempty"`
}

type ImagesUnion struct {
	UnSelect        *Image     `json:"un_select,omitempty"` //未选中
	Select          *Image     `json:"select,omitempty"`    //选中
	Lock            *Image     `json:"lock,omitempty"`
	UnfinishedImage *Image     `json:"unfinished_image,omitempty"`
	FinishedImage   *Image     `json:"finished_image,omitempty"`
	OptionalImage   *Image     `json:"optional_image,omitempty"`
	ShareImageInfo  *api.Image `json:"share_image_info,omitempty"`
	Event           *Image     `json:"event,omitempty"`  //赛事图片
	Button          *Image     `json:"button,omitempty"` //按钮图片

}

type CarouselImage struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Length      int64  `json:"length"`
	Width       int64  `json:"width"`
}

type IconRemark struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Content     string `json:"content"`
}

type Image struct {
	Image  string `json:"image,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Uri    string `json:"uri,omitempty"`
}

func ImageChange(in *api.ImageComm) *Image {
	if in != nil {
		return &Image{
			Image:  in.Image,
			Width:  int(in.Width),
			Height: int(in.Height),
		}
	}
	return nil
}

// Color .
type Color struct {
	BgColor                string `json:"bg_color,omitempty"`
	SelectFontColor        string `json:"select_font_color,omitempty"`
	NtSelectFontColor      string `json:"nt_select_font_color,omitempty"`
	TopColor               string `json:"top_color,omitempty"`
	TitleColor             string `json:"title_color,omitempty"`
	FontColor              string `json:"font_color,omitempty"`
	TopFontColor           string `json:"top_font_color,omitempty"`             //顶部字体颜色
	PanelSelectColor       string `json:"panel_select_color,omitempty"`         //展开面板选中色
	PanelBgColor           string `json:"panel_bg_color,omitempty"`             //展开面板背景色
	PanelSelectFontColor   string `json:"panel_select_font_color,omitempty"`    //展开面板字体选中色
	PanelNtSelectFontColor string `json:"panel_nt_select_font_color,omitempty"` //展开面板字体未选中色
	DisplayColor           string `json:"display_color,omitempty"`
	MoreColor              string `json:"more_color,omitempty"`
	TitleBgColor           string `json:"title_bg_color,omitempty"`
	SelectBgColor          string `json:"select_bg_color,omitempty"`
	TimelineColor          string `json:"timeline_color,omitempty"`
	NtBgColor              string `json:"nt_bg_color,omitempty"`
	NtFontColor            string `json:"nt_font_color,omitempty"`
	NtSelectBgColor        string `json:"nt_select_bg_color,omitempty"`
	SupernatantColor       string `json:"supernatant_color,omitempty"`
	SubtitleColor          string `json:"subtitle_color,omitempty"`
	FillColor              string `json:"fill_color,omitempty"`
	BorderColor            string `json:"border_color,omitempty"` //边框颜色
	StatusColor            string `json:"status_color,omitempty"` //状态颜色
}

// Setting .
type Setting struct {
	UnAllowClick    bool  `json:"un_allow_click,omitempty"` //tab不支持点击
	TabStyle        int   `json:"tab_style,omitempty"`      //tab组件样式 0颜色 1：图片
	IsDisplay       int64 `json:"is_display,omitempty"`     //inline-tab是否展示展开收起按
	ShareImage      bool  `json:"share_image,omitempty"`    //click-浮层 长按保存图片
	DisplayTitle    bool  `json:"display_title,omitempty"`
	AutoPlay        bool  `json:"auto_play,omitempty"`
	DisplayMore     bool  `json:"display_more,omitempty"`
	ArtDisplay      bool  `json:"art_display,omitempty"`
	ArcDisplay      bool  `json:"arc_display,omitempty"`
	PgcDisplay      bool  `json:"pgc_display,omitempty"`
	LiveType        int32 `json:"live_type,omitempty"` //主播未开播，0:隐藏卡片 1:直播间
	DisplayNum      bool  `json:"display_num,omitempty"`
	DisplayNodeNum  bool  `json:"display_node_num,omitempty"`
	DisplayDesc     bool  `json:"display_desc,omitempty"`
	SyncHoverButton bool  `json:"sync_hover_button,omitempty"`
	IsHighlight     bool  `json:"is_highlight,omitempty"`
	DisplayOp       bool  `json:"display_op,omitempty"`
	HiddenReserve   bool  `json:"hidden_reserve,omitempty"`
}

type Button struct {
	FollowText    string `json:"follow_text,omitempty"`
	FollowIcon    string `json:"follow_icon,omitempty"`
	IsFollow      int32  `json:"is_follow,omitempty"`
	UnFollowText  string `json:"un_follow_text,omitempty"`
	UnFollowIcon  string `json:"un_follow_icon,omitempty"`
	FollowToast   string `json:"follow_toast,omitempty"`
	UnFollowToast string `json:"un_follow_toast,omitempty"`
	Goto          string `json:"goto,omitempty"`
	Icon          string `json:"icon,omitempty"`
}

type ResourceInfo struct {
	Up       string `json:"up,omitempty"`
	View     string `json:"view,omitempty"`
	PubTime  string `json:"pub_time,omitempty"`
	Like     string `json:"like,omitempty"`
	Danmaku  string `json:"danmaku,omitempty"`
	Duration string `json:"duration,omitempty"`
	Follow   string `json:"follow,omitempty"`
	Season   string `json:"season,omitempty"`
}

type Dimension struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
	Rotate int64 `json:"rotate"`
}

type Repost struct {
	BizType    string `json:"biz_type"`
	SeasonType string `json:"season_type"`
	NewBizType string `json:"new_biz_type"` //与客户端上报逻辑保持一致
}

type ReasonStyle struct {
	Text         string `json:"text,omitempty"`
	BgColor      string `json:"bg_color,omitempty"`
	BgColorNight string `json:"bg_color_night,omitempty"`
}

type Liked struct {
	Sid          int64 `json:"sid"`
	Lid          int64 `json:"lid"`
	Score        int64 `json:"score"`
	HasLiked     int64 `json:"has_liked"`
	DisplayScore bool  `json:"display_score"`
}

func MixFolderUnmarshal(reason string) *MixFolder {
	if reason == "" {
		return nil
	}
	mixFold := &MixFolder{}
	if err := json.Unmarshal([]byte(reason), mixFold); err == nil {
		return mixFold
	}
	return nil
}

func splitRuleIDName(s string) (int64, string, error) {
	if s == "" {
		return 0, "", errors.New("ruleIDName is empty")
	}
	parts := strings.SplitN(s, "_", 2)
	ruleID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}
	ruleName := ""
	ruleMinLen := 2
	if len(parts) >= ruleMinLen {
		ruleName = parts[1]
	}
	return ruleID, ruleName, nil
}

func ExtractProgressParamFromClick(click *api.NativeClick) (sid, ruleID, dimension int64, err error) {
	tmpDimension, err := strconv.ParseInt(click.FinishedImage, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	dimension = tmpDimension
	ruleID, _, err = splitRuleIDName(click.OptionalImage)
	if err != nil {
		return 0, 0, 0, err
	}
	if click.ForeignID == 0 {
		return 0, 0, 0, errors.New("ForeignID is empty")
	}
	sid = click.ForeignID
	return sid, ruleID, dimension, nil
}

func unmarshalAreaImage(img string) *api.Image {
	image := new(api.Image)
	if err := json.Unmarshal([]byte(img), image); err == nil && image.Image != "" {
		return image
	}
	return nil
}

func (i *Item) setAreaTip(area *api.NativeClick) {
	areaTip := new(api.ClickTip)
	if err := json.Unmarshal([]byte(area.Tip), areaTip); err != nil {
		log.Error("Fail to unmarshal areaTip, tip=%+v error=%+v", area.Tip, err)
		return
	}
	if i.Color == nil {
		i.Color = &Color{}
	}
	i.Color.TopColor = areaTip.TopColor
	i.Color.TitleColor = areaTip.TitleColor
	i.Title = areaTip.Title
}

func (i *Item) setAreaExt(area *api.NativeClick, mou *api.NativeModule) *api.ClickExt {
	areaExt := new(api.ClickExt)
	if err := json.Unmarshal([]byte(area.Ext), areaExt); err != nil {
		log.Error("Fail to unmarshal areaExt, ext=%+v error=%+v", area.Ext, err)
		return nil
	}
	if i.SimpleImages == nil {
		i.SimpleImages = &SimpleImages{}
	}
	i.SimpleImages.ButtonImage = areaExt.ButtonImage
	i.Style = areaExt.Style
	if areaExt.Style == StyleImage {
		i.SimpleImages.LayerImage = areaExt.LayerImage
	}
	if mou.IsAttrShareImage() == api.AttrModuleYes && areaExt.ShareImage != nil && areaExt.ShareImage.Image != "" {
		if i.ImagesUnion == nil {
			i.ImagesUnion = &ImagesUnion{}
		}
		i.ImagesUnion.ShareImageInfo = areaExt.ShareImage
		if i.Setting == nil {
			i.Setting = &Setting{}
		}
		i.Setting.ShareImage = true
	}
	return areaExt
}

func (i *Item) FromNewEditorModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoNewEditorModule
	if mou == nil {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Title = mou.Title
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor}
	if mou.IsAttrLast() == api.AttrModuleYes && ext != nil {
		i.UrlExt = ext
		i.IsFeed = mou.IsAttrLast()
	}
}

func (p *Position) FormatArc(typ string, arc *arccli.Arc) {
	switch typ {
	case api.PosUp:
		*p = Position(arc.GetAuthor().Name)
	case api.PosView:
		*p = Position(statString(int64(arc.GetStat().View), "观看"))
	case api.PosPubTime:
		*p = Position(cardmdl.PubDataString(arc.GetPubDate().Time()))
	case api.PosLike:
		*p = Position(statString(int64(arc.GetStat().Like), "点赞"))
	case api.PosDanmaku:
		*p = Position(statString(int64(arc.GetStat().Danmaku), "弹幕"))
	default:
		*p = ""
	}
}

func (p *Position) FormatEp(typ string, ep *lmdl.EpPlayer) {
	switch typ {
	case api.PosDuration:
		*p = Position(cardmdl.DurationString(ep.Duration))
	case api.PosView:
		*p = Position(statString(ep.Stat.Play, "观看"))
	case api.PosFollow:
		*p = Position(statString(ep.Stat.Follow, "追剧"))
	default:
		*p = ""
	}
}

func (i *Item) FromEditorEp(ep *lmdl.EpPlayer, display bool, mou *api.NativeModule, rcmd *RcmdContent, posConf string) {
	posMeta := new(Positions)
	if err := json.Unmarshal([]byte(posConf), posMeta); err != nil {
		log.Warn("[FromEditorEp] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
	}
	i.Positions = new(Positions)
	i.Positions.Position1.FormatEp(string(posMeta.Position1), ep)
	i.Positions.Position2.FormatEp(string(posMeta.Position2), ep)
	i.Positions.Position3.FormatEp(string(posMeta.Position3), ep)
	i.Positions.Position4.FormatEp(string(posMeta.Position4), ep)
	i.Positions.Position5.FormatEp(string(posMeta.Position5), ep)
	i.fromEditorCard(mou, rcmd)
	// PGC不下发三点操作
	if i.Setting != nil {
		i.Setting.DisplayOp = false
	}
	i.Title = ep.ShowTitle
	i.Image = ep.Cover
	i.ItemID = ep.EpID
	i.URI = ep.Uri
	if display {
		i.Badge = &ReasonStyle{Text: ep.Season.TypeName}
	}
	i.Repost = &Repost{NewBizType: strconv.Itoa(api.BizOgvType), BizType: strconv.Itoa(api.MixEpidType), SeasonType: strconv.FormatInt(ep.Season.Type, 10)}
}

func (i *Item) FromNewEditorArc(mou *api.NativeModule, arc *arccli.Arc, display bool, rcmd *RcmdContent, f *favmdl.Folder) {
	posMeta := new(Positions)
	if err := json.Unmarshal([]byte(mou.TName), posMeta); err != nil {
		log.Error("[FromEditorArc] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
	}
	i.Positions = new(Positions)
	i.Positions.Position1.FormatArc(string(posMeta.Position1), arc)
	i.Positions.Position2.FormatArc(string(posMeta.Position2), arc)
	i.Positions.Position3.FormatArc(string(posMeta.Position3), arc)
	i.Positions.Position4.FormatArc(string(posMeta.Position4), arc)
	i.Positions.Position5.FormatArc(string(posMeta.Position5), arc)
	i.fromEditorCard(mou, rcmd)
	i.Title = arc.Title
	i.Image = arc.Pic
	i.ItemID = arc.Aid
	bvid, _ := bvidmdl.AvToBv(arc.Aid)
	if f == nil {
		i.URI = fmt.Sprintf("https://www.bilibili.com/video/%s", bvid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType), NewBizType: strconv.Itoa(api.BizUgcType)}
	} else {
		i.URI = fmt.Sprintf("https://m.bilibili.com/playlist/pl%d?bvid=%s&oid=%d", f.Mlid, bvid, arc.Aid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixFolder), NewBizType: strconv.Itoa(api.BizUgcType)}
		i.Fid = f.Mlid
	}
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
}

func (i *Item) FromEditorRankArc(rankItem *actGRPC.RankResult, mou *api.NativeModule, arc *actGRPC.ArchiveInfo, display bool, rcmd *RcmdContent) {
	posMeta := new(Positions)
	if err := json.Unmarshal([]byte(mou.TName), posMeta); err != nil {
		log.Error("[FromEditorArc] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
	}
	i.Positions = new(Positions)
	i.Positions.Position1.FormatRankArc(string(posMeta.Position1), rankItem, arc)
	i.Positions.Position2.FormatRankArc(string(posMeta.Position2), rankItem, arc)
	i.Positions.Position3.FormatRankArc(string(posMeta.Position3), rankItem, arc)
	i.Positions.Position4.FormatRankArc(string(posMeta.Position4), rankItem, arc)
	i.Positions.Position5.FormatRankArc(string(posMeta.Position5), rankItem, arc)
	i.fromEditorCard(mou, rcmd)
	i.Title = arc.Title
	i.Image = arc.Pic
	i.ItemID, _ = bvsafe.BvToAv(arc.BvID)
	i.URI = arc.ShowLink
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType), NewBizType: strconv.Itoa(api.BizUgcType)}
}

func (i *Item) fromEditorCard(mou *api.NativeModule, rcmd *RcmdContent) {
	i.Goto = GotoNewEditor
	i.Setting = &Setting{
		DisplayOp: mou.IsAttrDisplayOp() == api.AttrModuleYes,
	}
	// 内容推荐
	if rcmd != nil {
		i.RcmdContent = &RcmdContent{}
		if rcmd.TopContent != "" {
			i.RcmdContent.TopFontColor = rcmd.TopFontColor
			i.RcmdContent.TopIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/q~1vlO6h25.png"
			i.RcmdContent.TopContent = rcmd.TopContent
		}
		if rcmd.BottomContent != "" {
			i.RcmdContent.BottomFontColor = rcmd.BottomFontColor
			i.RcmdContent.BottomIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
			i.RcmdContent.BottomContent = rcmd.BottomContent
		}
		if rcmd.MiddleIcon != "" {
			i.RcmdContent.MiddleIcon = rcmd.MiddleIcon
		}
	}
}

func (i *Item) FromMVote(area *api.NativeClick, ext *ClickExt) {
	i.Goto = GotoVoteArea
	i.Width = area.Width
	i.Length = area.Length
	i.Lefty = area.Lefty
	i.Leftx = area.Leftx
	i.Type = area.Type
	i.ClickExt = ext
	i.ItemID = area.ID
	i.Param = strconv.FormatInt(area.ID, 10)
	i.Ukey = area.ExtUnmarshal().Ukey
}

func (i *Item) FromMVoteProcess(area *api.NativeClick, firExRankInfo map[int64]*actGRPC.ExternalRankInfo, mou *api.NativeModule) {
	i.FromMVote(area, nil)
	i.Style = area.ExtUnmarshal().Style
	i.Setting = &Setting{IsDisplay: mou.IsAttrDisplayNum()}
	jn := int64(0)
	for _, v := range area.ExtUnmarshal().Items {
		if v == nil {
			continue
		}
		tmp := &Item{Color: &Color{BgColor: v.BgColor}}
		if val, ok := firExRankInfo[jn]; ok && val != nil {
			tmp.Num = val.Vote
		}
		jn++
		i.Item = append(i.Item, tmp)
	}
}

func (i *Item) FromUpVoteProgress(area *api.NativeClick, options []*dyncommongrpc.VoteOptionInfo, mou *api.NativeModule) {
	areaExt := area.ExtUnmarshal()
	i.FromMVote(area, nil)
	i.Style = areaExt.Style
	i.Setting = &Setting{IsDisplay: mou.IsAttrDisplayNum()}
	var no int64
	for _, v := range areaExt.Items {
		if v == nil || no >= VoteOptionNum {
			continue
		}
		item := &Item{Color: &Color{BgColor: v.BgColor}}
		item.Num = int64(options[no].Cnt)
		i.Item = append(i.Item, item)
		no++
	}
}

func (i *Item) FromArea(area *api.NativeClick, ext *ClickExt, mou *api.NativeModule) {
	switch {
	case area.IsReserve(), area.IsActReserve(), area.IsFollow(), area.IsCatchUp(), area.IsBuyCoupon(), area.IsCartoon():
		i.Goto = GotoClickButton
	case area.IsPendant(), area.IsUpAppointment():
		i.Goto = GotoClickButtonV2
	case area.IsProgress():
		i.Goto = GotoClickProgress
		i.Ukey = area.UnfinishedImage
	case area.IsStaticProgress():
		i.Goto = GotoClickStaticProgress
	case area.IsRedirect():
		i.Goto = GotoClickButtonV3
	default:
		i.Goto = GotoClickArea
	}
	i.Width = area.Width
	i.Length = area.Length
	i.Lefty = area.Lefty
	i.Leftx = area.Leftx
	i.URI = area.Link
	i.Type = area.Type
	switch {
	case area.IsLayerImage():
		i.setAreaTip(area)
		areaExt := i.setAreaExt(area, mou)
		// 无图片也照样下发卡片
		i.Images = make([]*api.Image, 0, 3)
		if image := unmarshalAreaImage(area.UnfinishedImage); image != nil {
			i.Images = append(i.Images, image)
		}
		if image := unmarshalAreaImage(area.FinishedImage); image != nil {
			i.Images = append(i.Images, image)
		}
		if image := unmarshalAreaImage(area.OptionalImage); image != nil {
			i.Images = append(i.Images, image)
		}
		if areaExt != nil && len(areaExt.Images) > 0 {
			i.Images = append(i.Images, areaExt.Images...)
		}
	case area.IsLayerLink(), area.IsLayerInterface():
		i.setAreaTip(area)
		i.setAreaExt(area, mou)
	case area.IsAPP():
		i.IosURI = area.UnfinishedImage
		i.AndroidURI = area.FinishedImage
	case area.IsPendant(), area.IsUpAppointment():
		// 不下发图片
	case area.IsProgress(), area.IsStaticProgress():
		var progressExt = &api.ClickTip{}
		if err := json.Unmarshal([]byte(area.Tip), progressExt); err != nil {
			log.Error("Fail to unmarshal progressExt, progressExt=%+v error=%+v", area.Tip, err)
		}
		i.FontSize = progressExt.FontSize
		i.FontType = progressExt.FontType
		i.Color = &Color{FontColor: progressExt.FontColor}
		if ext != nil {
			if ext.DisplayNum == "" {
				ext.DisplayNum = strconv.FormatInt(ext.CurrentNum, 10)
			}
			i.TargetNum = ext.TargetNum
		}
		i.DisplayType = progressExt.DisplayType
		i.Ukey = area.UnfinishedImage
	case area.IsRedirect(), area.IsOnlyImage():
		i.Image = area.OptionalImage
	default:
		// 完成态图片
		if area.FinishedImage != "" && area.UnfinishedImage != "" {
			if i.ImagesUnion == nil {
				i.ImagesUnion = &ImagesUnion{}
			}
			i.ImagesUnion.FinishedImage = &Image{Image: area.FinishedImage}
			i.ImagesUnion.UnfinishedImage = &Image{Image: area.UnfinishedImage}
		}
	}
	i.ClickExt = ext
	i.ItemID = area.ID
	i.Param = strconv.FormatInt(area.ID, 10)
	func() {
		if area.Ext == "" {
			return
		}
		areaExt := new(api.ClickExt)
		if err := json.Unmarshal([]byte(area.Ext), areaExt); err != nil {
			log.Error("Fail to unmarshal areaExt, ext=%+v error=%+v", area.Ext, err)
			return
		}
		if i.Setting == nil {
			i.Setting = &Setting{}
		}
		i.Setting.SyncHoverButton = areaExt.SynHover
		i.Ukey = areaExt.Ukey
	}()
}

func (i *Item) FromClick(mou *api.NativeModule, act []*Item) {
	i.Goto = GotoClick
	if mou.IsBaseBottomButton() {
		i.Goto = GotoBottomButton
	}
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Item = make([]*Item, 0, 1)
	temp := &Item{}
	temp.FromClickBackground(mou)
	i.Item = append(i.Item, temp)
	if len(act) > 0 {
		i.Item[0].Item = act
	}
}

func (i *Item) FromVote(mou *api.NativeModule, act []*Item) {
	i.Goto = GotoVote
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Item = make([]*Item, 0, 1)
	temp := &Item{}
	temp.FromVoteBackground(mou)
	i.Item = append(i.Item, temp)
	if len(act) > 0 {
		i.Item[0].Item = act
	}
}

func (i *Item) FromVoteBackground(mou *api.NativeModule) {
	i.Goto = GotoVoteBack
	i.Width = mou.Width
	i.Length = mou.Length
	i.Image = mou.Meta
}

func (i *Item) FromClickBackground(mou *api.NativeModule) {
	i.Goto = GotoClickBack
	i.Width = mou.Width
	i.Length = mou.Length
	i.Image = mou.Meta
}

func (i *Item) FromVideoMore(mou *api.NativeModule, offset, pageID int64, dyOffset string) {
	i.Goto = GotoVideoMore
	i.Title = "查看更多"
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	params.Set("dy_offset", dyOffset)
	i.URI = "bilibili://following/activity_detail/" + strconv.FormatInt(mou.ID, 10) + "?" + params.Encode()
}

func (i *Item) FromUgcVideo(c *arccli.Arc, bvid string) {
	i.Goto = GotoNewUgcVideo
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.Param = bvid
	i.URI = fmt.Sprintf("bilibili://video/%d", c.Aid)
	i.Duration = cardmdl.DurationString(c.Duration) //时长
	i.View = statString(int64(c.Stat.View), "观看")
	i.Danmaku = statString(int64(c.Stat.Danmaku), "弹幕")
	i.Dimension = &Dimension{
		Width:  c.Dimension.Width,
		Height: c.Dimension.Height,
		Rotate: c.Dimension.Rotate,
	}
}

func (i *Item) FromPgcVideo(c *lmdl.EpPlayer) {
	i.Goto = GotoNewPgcVideo
	i.Title = c.ShowTitle //标题
	i.Image = c.Cover     //封面
	i.ItemID = c.EpID
	i.URI = c.Uri //ep一定会返回，不需要兜底逻辑
	i.Duration = cardmdl.DurationString(c.Duration)
	i.View = statString(int64(c.Stat.Play), "观看")
	i.Danmaku = statString(int64(c.Stat.Danmaku), "弹幕")
	i.Badge = &ReasonStyle{Text: c.Season.TypeName}
	i.Dimension = &Dimension{
		Width:  c.Dimension.Width,
		Height: c.Dimension.Height,
		Rotate: c.Dimension.Rotate,
	}
}

func (i *Item) FromDynamicMore(topicID, pageID int64, mou *api.NativeModule, sort, topicName, offset string) {
	i.Goto = GotoDynamicMore
	i.Title = "查看更多"
	params := url.Values{}
	params.Set("title", mou.Title)
	params.Set("sort", sort)
	params.Set("name", topicName)
	params.Set("module_id", strconv.FormatInt(mou.ID, 10))
	params.Set("sortby", strconv.FormatInt(int64(mou.DySort), 10))
	params.Set("offset", offset)
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	i.URI = "bilibili://following/topic_content_list/" + strconv.FormatInt(topicID, 10) + "?" + params.Encode()
}

func (i *Item) FromActs(act *api.NativePage) {
	i.Goto = GotoAct
	i.Image = act.ShareImage
	i.Title = act.Title
	i.Content = act.ShareTitle
	if act.IsOnline() {
		if act.SkipURL != "" {
			i.URI = act.SkipURL
		} else {
			i.URI = "https://www.bilibili.com/blackboard/dynamic/" + strconv.FormatInt(act.ID, 10)
		}
	} else {
		i.URI = "bilibili://pegasus/channel/" + strconv.FormatInt(act.ForeignID, 10) + "?type=topic"
	}
}

func (i *Item) FromEditorArc(arc *arccli.Arc, display bool, f *favmdl.Folder, rcmd *RcmdContent) {
	i.Goto = GotoEditor
	var bvid string
	if bvidStr, err := bvidmdl.AvToBv(arc.Aid); err == nil && bvidStr != "" {
		bvid = bvidStr
	}
	// 内容推荐
	if rcmd != nil {
		i.RcmdContent = &RcmdContent{}
		if rcmd.TopContent != "" {
			i.RcmdContent.TopFontColor = rcmd.TopFontColor
			i.RcmdContent.TopIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/q~1vlO6h25.png"
			i.RcmdContent.TopContent = rcmd.TopContent
		}
		if rcmd.BottomContent != "" {
			i.RcmdContent.BottomFontColor = rcmd.BottomFontColor
			i.RcmdContent.BottomIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
			i.RcmdContent.BottomContent = rcmd.BottomContent
		}
	}
	i.resourceCommon(arc, display, bvid, f)
}

func (i *Item) resourceCommon(c *arccli.Arc, display bool, bvid string, f *favmdl.Folder) {
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.Param = bvid
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	if f == nil {
		i.URI = fmt.Sprintf("bilibili://video/%d", c.Aid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType), NewBizType: strconv.Itoa(api.BizUgcType)}
	} else {
		i.URI = fmt.Sprintf("bilibili://music/playlist/playpage/%d?avid=%d", f.Mlid, c.Aid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixFolder), NewBizType: strconv.Itoa(api.BizUgcType)}
		i.Fid = f.Mlid
	}
	i.ResourceInfo = &ResourceInfo{
		Up:      c.GetAuthor().Name,
		View:    statString(int64(c.GetStat().View), "观看"),
		PubTime: cardmdl.PubDataString(c.GetPubDate().Time()),
		Like:    statString(int64(c.GetStat().Like), "点赞"),
		Danmaku: statString(int64(c.GetStat().Danmaku), "弹幕"),
	}
}

func (i *Item) FromResourceArc(c *arccli.Arc, display bool, bvid string, f *favmdl.Folder) {
	i.Goto = GotoResource
	i.CoverRightText = cardmdl.DurationString(c.Duration)
	i.CoverLeftText1 = statString(int64(c.Stat.View), "")
	i.CoverLeftIcon1 = IconPlay
	i.CoverLeftText2 = statString(int64(c.Stat.Danmaku), "")
	i.CoverLeftIcon2 = IconDanmaku
	i.resourceCommon(c, display, bvid, f)
}

func (i *Item) FromResourceEp(c *lmdl.EpPlayer, disploy bool) {
	i.Goto = GotoResource
	i.Title = c.ShowTitle //标题
	i.Image = c.Cover     //封面
	i.ItemID = c.EpID
	i.URI = c.Uri //ep一定会返回，不需要兜底逻辑
	i.CoverRightText = cardmdl.DurationString(c.Duration)
	i.CoverLeftText1 = statString(int64(c.Stat.Play), "")
	i.CoverLeftIcon1 = IconPlay
	// 追番数等待业务方返回
	i.CoverLeftText2 = statString(int64(c.Stat.Follow), "")
	i.CoverLeftIcon2 = IconFavorite
	if disploy {
		i.Badge = &ReasonStyle{Text: c.Season.TypeName}
	}
	i.Repost = &Repost{NewBizType: strconv.Itoa(api.BizOgvType), BizType: strconv.Itoa(api.MixEpidType), SeasonType: strconv.FormatInt(c.Season.Type, 10)}
	i.ResourceInfo = &ResourceInfo{
		View:     statString(c.Stat.Play, "观看"),
		Duration: cardmdl.DurationString(c.Duration),
		Follow:   statString(c.Stat.Follow, "追剧"),
	}
}

func (i *Item) FromResourceLive(c *playgrpc.RoomList) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	i.Image = c.Icon  //封面
	i.ItemID = c.RoomId
	i.URI = fmt.Sprintf("https://live.bilibili.com/%d", c.RoomId)
	i.CoverRightText = c.UserName
	i.CoverLeftText1 = c.Online
	i.CoverLeftIcon1 = IconLive
	if c.WatchedShow != nil {
		i.CoverLeftIcon1 = c.WatchedShow.Icon
		if c.WatchedShow.Switch {
			i.CoverLeftText1 = c.WatchedShow.TextLarge
		}
	}
	if c.Pendant != "" {
		i.Badge = &ReasonStyle{Text: c.Pendant}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixLive), NewBizType: strconv.Itoa(api.BizLive)}
}

// FromResourceArt .
func (i *Item) FromResourceArt(c *artmdl.Meta, artDisplay bool) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	if len(c.ImageURLs) >= 1 {
		i.Image = c.ImageURLs[0] //封面
	}
	i.ItemID = c.ID
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/mobile/%d", c.ID)
	if c.Stats != nil {
		i.CoverLeftText1 = statString(int64(c.Stats.View), "")
		i.CoverLeftText2 = statString(int64(c.Stats.Reply), "")
	}
	i.CoverLeftIcon1 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/UKWvn8PP.png"
	i.CoverLeftIcon2 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/Epgv08nd.png"
	if artDisplay {
		i.Badge = &ReasonStyle{Text: "文章"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixCvidType), NewBizType: strconv.Itoa(api.BizArtType)}
	i.ResourceInfo = &ResourceInfo{
		Up:      c.Author.Name,
		View:    statString(c.Stats.View, "观看"),
		PubTime: cardmdl.PubDataString(c.PublishTime.Time()),
		Like:    statString(c.Stats.Like, "点赞"),
	}
}

func statString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	tenThousand := int64(10000)
	if number < tenThousand {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	hundredMillion := int64(100000000)
	if number < hundredMillion {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

func (i *Item) FromOgvSeasonMore(mou *api.NativeModule) {
	i.Goto = GotoOgvSeasonMore
	i.Title = mou.Remark
	if mou.Remark == "" {
		i.Title = "查看更多"
	}
}

func (i *Item) FromOgvSeasonModule(mou *api.NativeModule) {
	i.Goto = GotoOgvSeasonModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Ukey = mou.Ukey
	i.Color = &Color{
		BgColor:          mou.BgColor,               //背景色
		TitleBgColor:     ryColors.TitleBgColor,     //卡片背景色-单列
		MoreColor:        mou.MoreColor,             //查看更多按钮色
		FontColor:        mou.FontColor,             //查看更多文字色
		TitleColor:       mou.TitleColor,            //剧集标题色-三列
		DisplayColor:     ryColors.DisplayColor,     //文字标题文字色
		SupernatantColor: ryColors.SupernatantColor, //浮层标题文字色
		SubtitleColor:    ryColors.SubtitleColor,
	}
}

// FromOgvSeason .
func (i *Item) FromOgvSeason(mou *api.NativeModule, ep *pgcAppGrpc.SeasonCardInfoProto, defaultTitle string) {
	if mou.IsCardThree() { //三列卡
		i.Goto = GotoOgvSeasoThree
		//追番/追剧数
		if ep.Stat != nil {
			i.CoverLeftText1 = ep.Stat.FollowView
		} else {
			i.CoverLeftText1 = "-"
		}
		//副标题
		if mou.IsAttrDisplayDesc() == api.AttrModuleYes {
			i.Content = defaultTitle
			if i.Content == "" {
				i.Content = ep.RecommendView // 副标题 > 更新集数
			}
		}
	} else {
		i.Goto = GotoOgvSeasoOne
		i.ResourceInfo = &ResourceInfo{}
		if ep.Stat != nil {
			i.ResourceInfo.View = statString(ep.Stat.View, "观看") //观看数
			i.ResourceInfo.Follow = ep.Stat.FollowView           //追番/追剧数
		} else {
			i.ResourceInfo.View = statString(0, "观看") //观看数
		}
		if ep.NewEp != nil {
			i.ResourceInfo.Season = ep.NewEp.IndexShow //追番/追剧数
		}
		//评分
		if mou.IsAttrDisplayNum() == api.AttrModuleYes {
			i.CoverRightText = fmt.Sprintf("%0.1f分", ep.Rating.Score)
		}
		if mou.IsAttrDisplayRecommend() == api.AttrModuleYes {
			i.Content = defaultTitle
			if i.Content == "" {
				i.Content = ep.Subtitle //运营配置 > 副标题
			}
		}
	}
	i.ItemID = int64(ep.SeasonId) // ssid
	i.Param = strconv.FormatInt(int64(ep.SeasonId), 10)
	//追番按钮
	if ep.FollowInfo != nil {
		i.Button = &Button{
			IsFollow:      ep.FollowInfo.IsFollow,
			FollowIcon:    ep.FollowInfo.FollowIcon,
			FollowText:    ep.FollowInfo.FollowText,
			UnFollowIcon:  ep.FollowInfo.UnfollowIcon,
			UnFollowText:  ep.FollowInfo.UnfollowText,
			FollowToast:   ep.FollowInfo.FollowToast,
			UnFollowToast: ep.FollowInfo.UnfollowToast,
		}
	}
	i.Image = ep.Cover //封面
	if mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes && ep.BadgeInfo != nil {
		//付费角标
		i.Badge = &ReasonStyle{Text: ep.BadgeInfo.Text, BgColor: ep.BadgeInfo.BgColor, BgColorNight: ep.BadgeInfo.BgColorNight}
	}
	i.Repost = &Repost{
		BizType:    strconv.Itoa(api.MixEpidType),
		SeasonType: strconv.FormatInt(int64(ep.SeasonType), 10),
		NewBizType: strconv.Itoa(api.BizOgvType),
	}
	//标题
	i.Title = ep.Title
	i.URI = ep.GetUrl()
}

func (i *Item) FromResourceProduct(c *ProductItem) {
	i.Goto = GotoResource
	i.Title = c.Title    //标题
	i.Image = c.ImageURL //封面
	i.ItemID = c.ItemID
	i.URI = c.LinkURL
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixProduct), NewBizType: strconv.Itoa(api.BizBusinessCommodity)}
}

func (i *Item) FromResourceWidItem(item *pgcAppGrpc.QueryWidItem) {
	i.Goto = GotoResource
	i.ItemID = int64(item.Id)
	i.Title = item.Title
	i.Image = item.Cover
	i.URI = item.Link
	switch item.Type {
	case pgc.WidItemTypeUgc:
		buildWidUgc(i, item)
	case pgc.WidItemTypeSeason:
		buildWidSeason(i, item)
	case pgc.WidItemTypeOgv:
		buildWidOgv(i, item)
	case pgc.WidItemTypeWeb:
		buildWidWeb(i, item)
	case pgc.WidItemTypeOgvFilm:
		buildWidOgvFilm(i, item)
	}
}

func buildWidUgc(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.CoverLeftIcon1 = IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = IconDanmaku
	i.CoverLeftText2 = statString(item.Dm, "")
	i.CoverRightText = cardmdl.DurationString(item.PlayLen)
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType), NewBizType: strconv.Itoa(api.BizUgcType)}
}

func buildWidSeason(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.CoverLeftIcon1 = IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = IconFavorite
	i.CoverLeftText2 = statString(item.Follow, "")
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixOgvSsid), NewBizType: strconv.Itoa(api.BizSeasonType)}
}

func buildWidOgv(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.CoverLeftIcon1 = IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = IconDanmaku
	i.CoverLeftText2 = statString(item.Dm, "")
	i.CoverRightText = cardmdl.DurationString(item.PlayLen)
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixEpidType), NewBizType: strconv.Itoa(api.BizOgvType)}
}

func buildWidWeb(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixWeb), NewBizType: strconv.Itoa(api.BizWebType)}
}

func buildWidOgvFilm(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixOgvFilm), NewBizType: strconv.Itoa(api.BizOgvFilmType)}
}

// FromInlineTabModule .
func (i *Item) FromInlineTabModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoInlineTabModule
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Item = items
}

func (i *Item) FromNewIDsModule(mou *api.NativeModule, ids []*ResourcesIDs, hasMore bool) {
	i.Goto = GotoNewVideoModule
	i.UrlExt = &UrlExt{Category: mou.Category, IDs: ids, HasMore: hasMore}
	i.Num = mou.Num
	i.FromCommonModule(mou)
}

// Icon
func (i *Item) FromIcon(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoIconModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Color = &Color{
		BgColor:   mou.BgColor,
		FontColor: mou.FontColor,
	}
	i.Item = items
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
}

func (i *Item) FromCarouselImg(mou *api.NativeModule) {
	i.Goto = GotoCarouselImg
	i.Style = strconv.FormatInt(mou.AvSort, 10)
	i.Color = &Color{SelectBgColor: mou.MoreColor}
	i.Setting = &Setting{AutoPlay: mou.IsAttrAutoPlay() == api.AttrModuleYes}
}

func (i *Item) FromCarouselImgItem(cv *CarouselImage) {
	i.Image = cv.ImgUrl
	i.URI = cv.RedirectUrl
	i.Length = cv.Length
	i.Width = cv.Width
}

// CarouselImg
func (i *Item) FromCarouselImgModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoCarouselImgModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Color = &Color{BgColor: mou.BgColor}
	i.Ukey = mou.Ukey
	i.Item = items
	i.Bar = mou.Bar
}

func (i *Item) FromCarouselWord(mou *api.NativeModule) {
	i.Goto = GotoCarouselWord
	i.Style = strconv.FormatInt(mou.AvSort, 10)
	i.ScrollType = mou.DySort
	i.Color = &Color{FontColor: mou.FontColor, BgColor: mou.TitleColor}
}

func (i *Item) FromCarouselWordItem(cvStr string) {
	i.Content = cvStr
}

// CarouselWord
func (i *Item) FromCarouselWordModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoCarouselWordModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Color = &Color{BgColor: mou.BgColor}
	i.Item = items
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
}

func (i *Item) FromIconExt(list []*api.NativeMixtureExt) {
	i.Goto = GotoIcon
	for _, v := range list {
		if v == nil {
			continue
		}
		ext := &IconRemark{}
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("FromIconExt Fail to unmarshal iconExt, iconExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &Item{
			Image:   ext.ImgUrl,
			URI:     ext.RedirectUrl,
			Content: ext.Content,
		}
		i.Item = append(i.Item, item)
	}
}

func (i *Item) FromNewDynModule(mou *api.NativeModule) {
	i.Goto = GotoNewVideoModule
	i.UrlExt = &UrlExt{Fid: mou.Fid, Type: 8, Num: mou.Num, Category: mou.Category}
	i.FromCommonModule(mou)
}

func (i *Item) FromResourceIDsModule(mou *api.NativeModule, ids []*ResourcesIDs, hasMore bool) {
	i.UrlExt = &UrlExt{Category: mou.Category, IDs: ids, HasMore: hasMore}
	i.Num = mou.Num
	i.FromResourceCommon(mou)
}

func (i *Item) FromTimelineModule(mou *api.NativeModule) {
	i.Goto = GotoTimelineModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Ukey = mou.Ukey
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleBgColor: ryColors.TitleBgColor, TimelineColor: ryColors.TimelineColor}
}

func (i *Item) FromActCapsuleItem(card *api.NativePageCard) {
	i.ItemID = card.Id
	i.Title = card.Title
	i.URI = card.SkipURL
}

func (i *Item) FromActModule(mou *api.NativeModule) {
	i.Goto = GotoActModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.IsFeed = mou.IsAttrLast()
	i.Title = mou.Title
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
	i.Bar = mou.Bar
	i.Ukey = mou.Ukey
}

func (i *Item) FromActCapsuleModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoActCapsuleModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Color = &Color{BgColor: mou.BgColor}
	i.Bar = mou.Bar
	i.Item = items
}

func FromTimelineFormatHead(stime xtime.Time, timeSort int64) string {
	var title string
	y, m, d := stime.Time().Date()
	h, min, sec := stime.Time().Clock()
	//精确到 0:年 1:月 2: 日 3:时 4:分 5:秒
	switch timeSort {
	case api.TimeSortMonth:
		title = fmt.Sprintf("%d年%d月", y, m)
	case api.TimeSortDay:
		title = fmt.Sprintf("%d年%d月%d日", y, m, d)
	case api.TimeSortHour:
		title = fmt.Sprintf("%d年%d月%d日 %02d时", y, m, d, h)
	case api.TimeSortMin:
		title = fmt.Sprintf("%d年%d月%d日 %02d:%02d", y, m, d, h, min)
	case api.TimeSortSec:
		title = fmt.Sprintf("%d年%d月%d日 %02d:%02d:%02d", y, m, d, h, min, sec)
	default:
		title = fmt.Sprintf("%d年", y)
	}
	return title
}

// FromTimeline 图文.
func (i *Item) FromTimeline(obj interface{}, first string) {
	i.Goto = GotoTimelineMix
	i.CoverLeftText1 = first
	switch c := obj.(type) {
	case *api.MixReason:
		i.Title = c.Title       //标题
		i.Subtitle = c.SubTitle //副标题
		i.Content = c.Desc      //正文
		i.Image = c.Image
		i.URI = c.Url
	case *populargrpc.TimeEvent:
		i.Content = c.Title //正文
		i.Image = c.Pic
		i.URI = c.JumpLink
		i.Stime = int64(c.Stime)
		i.ItemID = c.EventId

	}
}

func (i *Item) FromTimelineArc(c *arccli.Arc, first string) {
	i.Goto = GotoTimelineResource
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.CoverLeftText1 = first
	bvid, _ := bvidmdl.AvToBv(c.Aid)
	i.URI = fmt.Sprintf("https://www.bilibili.com/video/%s", bvid)
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType), NewBizType: strconv.Itoa(api.BizUgcType)}
	i.ResourceInfo = &ResourceInfo{
		Up:   c.GetAuthor().Name,
		View: statString(int64(c.GetStat().View), "观看"),
	}
}

func (i *Item) FromTimelineArt(c *artmdl.Meta, first string) {
	i.Goto = GotoTimelineResource
	i.Title = c.Title //标题
	if len(c.ImageURLs) >= 1 {
		i.Image = c.ImageURLs[0] //封面
	}
	i.ItemID = c.ID
	i.CoverLeftText1 = first
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/cv%d", c.ID)
	i.Badge = &ReasonStyle{Text: "文章", BgColor: "#FB7299"}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixCvidType), NewBizType: strconv.Itoa(api.BizArtType)}
	i.ResourceInfo = &ResourceInfo{
		Up:   c.Author.Name,
		View: statString(c.Stats.View, "观看"),
	}
}

// FromTimelinePic 图.
func (i *Item) FromTimelinePic(c *api.MixReason, first string) {
	i.Goto = GotoTimelinePic
	i.Title = c.Title //标题
	i.Image = c.Image
	i.URI = c.Url
	i.Width = int64(c.Width)
	i.Length = int64(c.Length)
	i.CoverLeftText1 = first
}

// FromTimelineText 文.
func (i *Item) FromTimelineText(c *api.MixReason, first string) {
	i.Goto = GotoTimelineText
	i.Title = c.Title       //标题
	i.Subtitle = c.SubTitle //副标题
	i.Content = c.Desc      //正文
	i.URI = c.Url
	i.CoverLeftText1 = first
}

func (i *Item) FromTitleImage(mou *api.NativeModule) {
	i.Goto = GotoTitleImage
	i.Image = mou.Meta
}

func (i *Item) FromRcmdVerticalModule(mou *api.NativeModule) {
	i.Goto = GotoRcmdVerticalMou
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromRcmdVertical(items []*Item) {
	i.Goto = GotoRcmdVertical
	i.Item = items
}

func (i *Item) FromRcmdVerticalItem(mixExt *api.NativeMixtureExt, account *accgrpc.Card, clickExt *ClickExt) {
	i.ClickExt = clickExt
	i.UserInfo = &UserInfo{
		Mid:          account.GetMid(),
		Name:         account.GetName(),
		Face:         account.GetFace(),
		OfficialInfo: account.GetOfficial(),
		Vip:          account.GetVip(),
	}
	// 认证信息转换
	i.UserInfo.formatRole()
	// 默认个人空间
	i.URI = fmt.Sprintf("https://m.bilibili.com/space/%d?defaultTab=dynamic", account.Mid)
	if mixExt.Reason != "" {
		rcmdExt := new(struct {
			Reason string `json:"reason"` //推荐理由
			URI    string `json:"uri"`    //链接
		})
		if err := json.Unmarshal([]byte(mixExt.Reason), rcmdExt); err != nil {
			log.Error("Fail to unmarshal mixExt.Reason, reason=%s", mixExt.Reason)
			return
		}
		i.Title = rcmdExt.Reason
		if rcmdExt.URI != "" {
			i.URI = rcmdExt.URI
		}
	}
}

// Recommend
func (i *Item) FromReserve(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoReserveModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Item = items
	i.Ukey = mou.Ukey
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, TitleBgColor: ryColors.TitleBgColor}
}

// Recommend
func (i *Item) FromGame(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoGameModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Item = items
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func isDisplayReserve(total, limit, mid, upmid int64) bool {
	// limit无限制 或者 发起人mid为0
	if limit <= 0 || upmid == 0 {
		return true
	}
	if mid == upmid {
		return true
	}
	if total >= limit {
		return true
	}
	return false
}

func (i *Item) FromReserveExt(mix *api.NativeMixtureExt, act *ReserveRly, displayUp, mid int64) {
	i.Goto = GotoReserve
	i.Param = fmt.Sprintf("%d", mix.ForeignID)
	if displayUp == 1 && act.Account != nil {
		i.UserInfo = &UserInfo{
			Mid:  act.Account.GetMid(),
			Name: act.Account.GetName(),
			Face: act.Account.GetFace(),
		}
	}
	if act.Item == nil {
		return
	}
	var hiddenReserve bool
	switch act.DisplayType {
	case ReserveDisplayA:
		switch act.Item.Type {
		case actGRPC.UpActReserveRelationType_Archive:
			i.Title = act.Item.Title
			i.Content = "视频预约"
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		case actGRPC.UpActReserveRelationType_Live:
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		case actGRPC.UpActReserveRelationType_Course:
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 开售", reserveTime(act.Item.LivePlanStartTime))
			i.Num = act.Item.Total
			i.Subtitle = "人已预约"
		default:
		}
		if act.Item.IsFollow == 1 {
			i.Button = &Button{FollowText: "已预约", IsFollow: 1, Goto: GotoClickReserve, Icon: IconReserve}
		} else {
			i.Button = &Button{FollowText: "预约", Goto: GotoClickReserve, Icon: IconReserve}
		}
	case ReserveDisplayC:
		switch act.Item.Type {
		case actGRPC.UpActReserveRelationType_Archive:
			i.Title = act.Item.Title
			//获取观看人数
			var views int32
			if act.Arc != nil {
				views = act.Arc.Stat.View
			}
			i.Content = "视频预约"
			i.Num = int64(views)
			i.Subtitle = "观看"
		case actGRPC.UpActReserveRelationType_Live:
			i.Title = act.Item.Title
			//获取直播人数
			var popularCount int64
			if act.Live != nil && act.Live.SessionInfoPerLive != nil {
				popularCount = act.Live.SessionInfoPerLive.PopularityCount
			}
			i.Content = "直播中"
			i.Num = popularCount
			i.Subtitle = "人气"
			i.Setting = &Setting{IsHighlight: true}
			func() {
				if act.Live == nil || act.Live.SessionInfoPerLive == nil || act.Live.SessionInfoPerLive.WatchedShow == nil {
					return
				}
				if !act.Live.SessionInfoPerLive.WatchedShow.Switch {
					return
				}
				i.Num = act.Live.SessionInfoPerLive.WatchedShow.Num
				i.Subtitle = "人看过"
			}()
		case actGRPC.UpActReserveRelationType_Course:
			i.Title = act.Item.Title
			i.Num = act.Item.OidView
			i.Subtitle = "人看过"
		default:
		}
		i.Button = &Button{FollowText: "去观看", Goto: GotoClickURL}
	case ReserveDisplayD:
		if act.Item.Type == actGRPC.UpActReserveRelationType_Live {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		}
		i.Button = &Button{FollowText: "看回放", Goto: GotoClickURL}
	case ReserveDisplayE:
		if act.Item.Type == actGRPC.UpActReserveRelationType_Live {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		} else if act.Item.Type == actGRPC.UpActReserveRelationType_Course {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 开售", reserveTime(act.Item.LivePlanStartTime))
			i.Num = act.Item.Total
			i.Subtitle = "人已预约"
		}
		i.Button = &Button{FollowText: "已结束", Goto: GotoClickUnable}
	default:
	}
	if hiddenReserve {
		if i.Setting == nil {
			i.Setting = &Setting{}
		}
		i.Setting.HiddenReserve = hiddenReserve
	}
}

func reserveTime(in xtime.Time) string {
	nowT := time.Now()
	if in.Time().Format("20060102") == nowT.Format("20060102") {
		return "今天 " + in.Time().Format("15:04")
	}
	atnowT := time.Now().AddDate(0, 0, 1)
	if in.Time().Format("20060102") == atnowT.Format("20060102") {
		return "明天 " + in.Time().Format("15:04")
	}
	if in.Time().Year() == nowT.Year() { //同一个自然年
		return in.Time().Format("01月02日 15:04")
	}
	return in.Time().Format("2006年01月02日 15:04")
}

func (i *Item) FromGameExt(mix *api.NativeMixtureExt, act *gamdl.Item) {
	i.Goto = GotoGame
	i.Param = fmt.Sprintf("%d", act.GameBaseId)
	i.Image = act.GameIcon
	i.Title = act.GameName
	i.URI = act.GameLink
	i.Subtitle = act.GameSubtitle
	mixReason := mix.RemarkUnmarshal()
	if mixReason.Desc != "" {
		i.Subtitle = mixReason.Desc
	}
	i.Content = strings.Join(act.GameTags, "/")
	i.Button = &Button{FollowText: "下载"}
}

func (i *Item) FromHoverButton(mou *api.NativeModule, items []*Item, confSort *api.ConfSort) {
	i.Goto = GotoHoverButtonModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Item = items
	i.MutexUkeys = confSort.MUkeys
}

func (i *Item) FromRecommendRankExt(mix *actGRPC.RankResult, ext *ClickExt, display bool, img string) {
	i.Goto = GotoRecommend
	//用户头像图标
	if mix.Account != nil {
		i.UserInfo = &UserInfo{
			Mid:  mix.Account.MID,
			Name: mix.Account.Name,
			Face: mix.Account.Face,
		}
		vl := mix.Account.Vip.Label
		vipLabel := accgrpc.VipLabel{
			Path:        vl.Path,
			LabelTheme:  vl.LabelTheme,
			TextColor:   vl.TextColor,
			BgStyle:     vl.BgStyle,
			BgColor:     vl.BgColor,
			BorderColor: vl.BorderColor,
		}
		i.UserInfo.Vip = accgrpc.VipInfo{
			Type:               mix.Account.Vip.Type,
			Status:             mix.Account.Vip.Status,
			DueDate:            mix.Account.Vip.DueDate,
			VipPayType:         mix.Account.Vip.VipPayType,
			ThemeType:          mix.Account.Vip.ThemeType,
			AvatarSubscript:    mix.Account.Vip.AvatarSubscript,
			NicknameColor:      mix.Account.Vip.NicknameColor,
			Role:               mix.Account.Vip.Role,
			Label:              vipLabel,
			AvatarSubscriptUrl: mix.Account.Vip.AvatarSubscriptUrl,
		}
		i.UserInfo.OfficialInfo = accgrpc.OfficialInfo{
			Role:  mix.Account.Official.Role,
			Title: mix.Account.Official.Title,
			Desc:  mix.Account.Official.Desc,
			Type:  mix.Account.Official.Type,
		}
		// 认证信息转换
		i.UserInfo.formatRole()
	}
	//空间跳转地址
	i.URI = fmt.Sprintf("https://m.bilibili.com/space/%d?defaultTab=dynamic", mix.Account.MID)
	// 推荐理由
	if display {
		i.Title = mix.ShowScore
	}
	if img != "" {
		i.Image = img
	}
	i.ClickExt = ext
}

func (i *Item) FromRecommendExt(mix *api.NativeMixtureExt, act *accgrpc.Card, ext *ClickExt) {
	i.Goto = GotoRecommend
	//用户头像图标
	i.UserInfo = &UserInfo{
		Mid:          act.Mid,
		Name:         act.Name,
		Face:         act.Face,
		OfficialInfo: act.Official,
		Vip:          act.Vip,
	}
	// 认证信息转换
	i.UserInfo.formatRole()
	//空间跳转地址
	i.URI = fmt.Sprintf("https://m.bilibili.com/space/%d?defaultTab=dynamic", act.Mid)
	// 推荐理由
	if mix != nil {
		i.Title = mix.Reason
	}
	i.ClickExt = ext
}

// Recommend
func (i *Item) FromRecommend(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoRecommendModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	i.Item = items
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromTitleName(mou *api.NativeModule) {
	i.Goto = GotoTitleName
	i.Title = mou.Caption
}

func (i *Item) FromTimelineExpand(mou *api.NativeModule) {
	i.Goto = GotoTimelineExpand
	i.Title = mou.Remark
	if mou.Remark == "" {
		i.Title = "展开"
	}
	i.Subtitle = "收起"
}

func (i *Item) FromTimelineMore(mou *api.NativeModule) {
	i.Goto = GotoTimelineMore
	i.Title = mou.Remark
	if mou.Remark == "" {
		i.Title = "查看更多"
	}
}

func (i *Item) FromEditorOriginModule(mou *api.NativeModule) {
	confSort := mou.ConfUnmarshal()
	i.UrlExt = &UrlExt{Category: mou.Category, Fid: mou.Fid, Type: int32(confSort.RdbType), ConfModuleID: mou.ID, Sid: confSort.Sid, Counter: confSort.Counter}
	i.Num = mou.Num
	i.FromEditorCommon(mou)
}

func (i *Item) FromEditorModule(mou *api.NativeModule, ids []*ResourcesIDs, hasMore bool) {
	confSort := mou.ConfUnmarshal()
	i.UrlExt = &UrlExt{Category: mou.Category, IDs: ids, HasMore: hasMore, Sid: confSort.Sid, Counter: confSort.Counter}
	i.Num = mou.Num
	i.FromEditorCommon(mou)
}

func (i *Item) FromEditorCommon(mou *api.NativeModule) {
	i.Goto = GotoEditorModule
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Title = mou.Title
	i.Bar = mou.Bar
	i.Meta = mou.Meta
	i.Caption = mou.Caption
	if mou.TName != "" {
		posMeta := new(Positions)
		if err := json.Unmarshal([]byte(mou.TName), posMeta); err != nil { //错误降级处理
			log.Error("[FromEditorArc] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
		}
		i.Positions = posMeta
	}
	i.Attr = mou.Attribute
	i.Color = &Color{BgColor: mou.BgColor}
}

func (i *Item) FromResourceOriginModule(mou *api.NativeModule, sortType int64) {
	confSort := mou.ConfUnmarshal()
	i.UrlExt = &UrlExt{SourceID: mou.TName, Type: int32(confSort.RdbType), Num: mou.Num, Category: mou.Category}
	if confSort.RdbType == api.RDBLive {
		i.UrlExt.SortType = sortType
	}
	i.FromResourceCommon(mou)
}

func (i *Item) FromNavigation(mou *api.NativeModule) {
	i.Goto = GotoNavigationModule
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	tmpColor := &Color{}
	if mou.BgColor != "" {
		bgColors := strings.Split(mou.BgColor, ",")
		for n, v := range bgColors {
			switch n {
			case 0:
				tmpColor.BgColor = v
			case 1:
				tmpColor.NtBgColor = v
			default:
				break
			}
		}
	}
	if mou.FontColor != "" {
		fontColors := strings.Split(mou.FontColor, ",")
		for n, v := range fontColors {
			switch n {
			case 0:
				tmpColor.FontColor = v
			case 1:
				tmpColor.NtFontColor = v
			default:
				break
			}
		}
	}
	if mou.TitleColor != "" {
		selectBgColors := strings.Split(mou.TitleColor, ",")
		for n, v := range selectBgColors {
			switch n {
			case 0:
				tmpColor.SelectBgColor = v
			case 1:
				tmpColor.NtSelectBgColor = v
			default:
				break
			}
		}
	}
	if mou.MoreColor != "" {
		selectFontColors := strings.Split(mou.MoreColor, ",")
		for n, v := range selectFontColors {
			switch n {
			case 0:
				tmpColor.SelectFontColor = v
			case 1:
				tmpColor.NtSelectFontColor = v
			default:
				break
			}

		}
	}
	i.Item = append(i.Item, &Item{
		Color: tmpColor,
		Goto:  GotoNavigation,
	})
}

func (i *Item) FromResourceRoleModule(mou *api.NativeModule) {
	i.UrlExt = &UrlExt{Fid: mou.Length, SeasonID: mou.Width, Num: mou.Num, Category: mou.Category}
	i.FromResourceCommon(mou)
}

func (i *Item) FromResourceActModule(mou *api.NativeModule, types int32) {
	i.UrlExt = &UrlExt{Fid: mou.Fid, Type: types, Num: mou.Num, Category: mou.Category}
	i.FromResourceCommon(mou)
}

func (i *Item) FromResourceModule(mou *api.NativeModule, types int32) {
	i.UrlExt = &UrlExt{Fid: mou.Fid, Type: types, Num: mou.Num, Category: mou.Category}
	i.FromResourceCommon(mou)
}

func (i *Item) FromResourceCommon(mou *api.NativeModule) {
	i.Goto = GotoResourceModule
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Title = mou.Title
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Meta = mou.Meta
	i.Caption = mou.Caption
	i.Attr = mou.Attribute
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, MoreColor: mou.MoreColor, FontColor: mou.FontColor, TitleBgColor: ryColors.TitleBgColor, DisplayColor: ryColors.DisplayColor}
}

func (i *Item) FromNewVideoActModule(mou *api.NativeModule, sortType int32) {
	i.Goto = GotoNewVideoModule
	i.UrlExt = &UrlExt{Fid: mou.Fid, Type: sortType, Num: mou.Num, Category: mou.Category}
	i.FromCommonModule(mou)
}

func (i *Item) FromCommonModule(mou *api.NativeModule) {
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Title = mou.Title
	i.Ukey = mou.Ukey
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, MoreColor: mou.MoreColor, FontColor: mou.FontColor, DisplayColor: ryColors.DisplayColor}
	//i.Setting = &Setting{
	//	AutoPlay:     mou.IsAttrAutoPlay() == api.AttrModuleYes,
	//	DisplayTitle: !(mou.IsAttrHideTitle() == api.AttrModuleYes),
	//	DisplayMore:  mou.IsAttrHideMore() != api.AttrModuleYes,
	//}
	i.Attr = mou.Attribute
	i.Bar = mou.Bar
	i.Meta = mou.Meta
	i.Caption = mou.Caption

}

func (i *Item) FromStatementModule(mou *api.NativeModule) {
	i.Goto = GotoStatementModule
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Content = mou.Remark
	i.Setting = &Setting{IsDisplay: mou.IsAttrStatementDisplayButton()}
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromReplyModule(mou *api.NativeModule) {
	i.Goto = GotoReplyModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.UrlExt = &UrlExt{
		Fid:  mou.Fid,           //评论id
		Type: int32(mou.AvSort), //评论类型
		//UpperMid:0, //评论mid,暂时不下发，客户端需具备透传能力
	}
}

func (i *Item) FromLiveModule(mou *api.NativeModule) {
	i.Goto = GotoLiveModule
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, FontColor: mou.FontColor, DisplayColor: ryColors.DisplayColor}
	i.Setting = &Setting{LiveType: mou.LiveType}
	i.Bar = mou.Bar
	i.Meta = mou.Meta
	i.Caption = mou.Caption
	i.Attr = mou.Attribute
	i.Image = mou.TName
	i.Fid = mou.Fid
}

func (i *Item) FormatSelect(mou *api.NativeModule) {
	i.Goto = GotoSelect
	i.Setting = &Setting{IsDisplay: 1}
	i.Title = mou.Title
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{
		BgColor:                mou.BgColor,
		TopFontColor:           ryColors.SelectColor,
		PanelSelectColor:       ryColors.NotSelectColor,
		PanelBgColor:           ryColors.PanelBgColor,
		PanelSelectFontColor:   ryColors.PanelSelectColor,
		PanelNtSelectFontColor: ryColors.PanelNotSelectColor,
	}
}

func (i *Item) FromSelectModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoSelectModule
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Item = items
}

func (i *Item) FormatInline(mou *api.NativeModule) {
	i.Goto = GotoInlineTab
	i.Title = mou.Title
	// 0,2:颜色 1:图片
	i.Setting = &Setting{IsDisplay: mou.IsAttrDisplayButton()}
	if mou.AvSort == 1 {
		i.Setting.TabStyle = 1
		i.Image = mou.Meta
		if mou.Meta != "" && mou.Width > 0 && mou.Length > 0 {
			i.Width = mou.Width
			i.Length = mou.Length
		} else {
			i.Width = 1125
			i.Length = 120
		}
	}
	if mou.AvSort == 0 || mou.AvSort == 2 { //低版本 或者选择颜色
		i.Color = &Color{BgColor: mou.BgColor, SelectFontColor: mou.MoreColor, NtSelectFontColor: mou.FontColor}
	}
}

func (i *Item) FromDynamicModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoDynamicModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Fid = mou.Fid
	i.Title = mou.Title
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, DisplayColor: ryColors.DisplayColor}
	i.Meta = mou.Meta
	i.Caption = mou.Caption
	i.UrlExt = ext
	i.Num = mou.Num
	i.Ukey = mou.Ukey
}

func (i *Item) FromProgressModule(mou *api.NativeModule, item []*Item) {
	i.Goto = GotoProgressModule
	if mou == nil {
		return
	}
	i.Ukey = mou.Ukey
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Title = mou.Title
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor}
	i.Item = item
}

func (i *Item) FromProgress(mou *api.NativeModule, group *actGRPC.ActivityProgressGroup) {
	i.Goto = GotoProgress
	i.ItemID = mou.ID
	switch mou.AvSort {
	case ProgStyleRound:
		i.DisplayType = DisTypeRound
	case ProgStyleRectangle:
		i.DisplayType = DisTypeRectangle
	case ProgStyleNode:
		i.DisplayType = DisTypeNode
	}
	i.BgStyle, _ = strconv.ParseInt(mou.MoreColor, 10, 64)
	i.IndicatorStyle, _ = strconv.ParseInt(mou.TitleColor, 10, 64)
	var (
		one   = int64(1)
		two   = int64(2)
		three = int64(3)
	)
	switch mou.Length {
	case one:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/I2CHCvboo.png"
	case two:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/bxL-soEal.png"
	case three:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/vniA~jzxp.png"
	}
	i.Color = &Color{FillColor: mou.FontColor}
	i.Setting = &Setting{
		DisplayNum:     mou.IsAttrDisplayNum() == api.AttrModuleYes,
		DisplayNodeNum: mou.IsAttrDisplayNodeNum() == api.AttrModuleYes,
		DisplayDesc:    mou.IsAttrDisplayDesc() == api.AttrModuleYes,
	}
	i.Num = group.Total
	func() {
		items := make([]*Item, 0, len(group.Nodes))
		for _, v := range group.Nodes {
			item := &Item{
				Title: v.Desc,
				Num:   v.Val,
			}
			items = append(items, item)
		}
		i.Item = items
	}()

}

func (i *Item) FromVideoModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoVideoModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Title = mou.Title
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, DisplayColor: ryColors.DisplayColor}
	i.Meta = mou.Meta
	i.Caption = mou.Caption
	i.UrlExt = ext
	i.Num = mou.Num
	i.Ukey = mou.Ukey
}

func (i *Item) FromBaseHead(mou *api.NativeModule) {
	i.Title = mou.Title
	i.Color = &Color{BgColor: mou.BgColor}
	i.Attr = mou.Attribute
}
