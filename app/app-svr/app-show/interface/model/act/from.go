package act

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actapi "git.bilibili.co/bapis/bapis-go/activity/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	livefeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"go-common/library/log"
	xtime "go-common/library/time"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model"
	bgmmdl "go-gateway/app/app-svr/app-show/interface/model/bangumi"
	busmdl "go-gateway/app/app-svr/app-show/interface/model/business"
	"go-gateway/app/app-svr/app-show/interface/model/dynamic"
	gamdl "go-gateway/app/app-svr/app-show/interface/model/game"
	"go-gateway/app/app-svr/app-show/interface/model/pgc"
	arccli "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/app/web-svr/native-page/interface/api"
	bvsafe "go-gateway/pkg/idsafe/bvid"
)

// Item .
type Item struct {
	Goto   string `json:"goto,omitempty"`
	Param  string `json:"param,omitempty"`
	ItemID int64  `json:"item_id,omitempty"`
	Ukey   string `json:"ukey,omitempty"`
	// click
	Width       int64   `json:"width,omitempty"`
	Length      int64   `json:"length,omitempty"`
	Image       string  `json:"image,omitempty"`
	ImageType   int32   `json:"image_type,omitempty"`
	UnImage     string  `json:"un_image,omitempty"`
	UnImageType int32   `json:"un_image_type,omitempty"`
	ShareImage  string  `json:"share_image,omitempty"`
	ShareType   int     `json:"share_type,omitempty"`
	Leftx       int64   `json:"leftx,omitempty"`
	Lefty       int64   `json:"lefty,omitempty"`
	URI         string  `json:"uri,omitempty"`
	Content     string  `json:"content,omitempty"`
	HeadURI     string  `json:"head_uri,omitempty"`
	Title       string  `json:"title,omitempty"`
	Subtitle    string  `json:"subtitle,omitempty"`
	Item        []*Item `json:"item,omitempty"`
	ChildItem   []*Item `json:"-"` //低版本兼容逻辑，inline_tab的子组件
	// 动态卡片接口 数据透传
	DyCard *dynamic.DyCard `json:"dy_card,omitempty"`
	//点赞相关信息
	Liked            *Liked                 `json:"liked,omitempty"`
	IsGap            int32                  `json:"is_gap,omitempty"`
	IsFeed           int64                  `json:"is_feed,omitempty"`
	HasLive          int64                  `json:"has_live,omitempty"`
	IsDisplay        int64                  `json:"is_display,omitempty"` //tab组件是否展示展开收起按钮
	UrlExt           *UrlExt                `json:"url_ext,omitempty"`
	Color            *Color                 `json:"color,omitempty"`
	Setting          *Setting               `json:"setting,omitempty"`
	UserInfo         *UserInfo              `json:"user_info,omitempty"`
	ClickExt         *ClickExt              `json:"click_ext,omitempty"`
	Bar              string                 `json:"-"`
	CoverLeftText1   string                 `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1   cardmdl.Icon           `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2   string                 `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2   cardmdl.Icon           `json:"cover_left_icon_2,omitempty"`
	CoverRightText   string                 `json:"cover_right_text,omitempty"`
	CoverRightText1  string                 `json:"cover_right_text_1,omitempty"`
	Badge            *ReasonStyle           `json:"badge,omitempty"`
	Repost           *Repost                `json:"repost,omitempty"`
	OptionalImage    string                 `json:"opt_image,omitempty"`
	OptionalImage2   string                 `json:"opt_image_2,omitempty"`
	CoverLeftText3   string                 `json:"cover_left_text_3,omitempty"`
	Rights           *ArcRights             `json:"rights,omitempty"`
	Dimension        *ArcDimension          `json:"dimension,omitempty"`
	LiveCard         *livefeed.LiveCardInfo `json:"live_card,omitempty"`
	Icon             *Icon                  `json:"icon,omitempty"`
	Text             *Text                  `json:"text,omitempty"`
	Positions        *Positions             `json:"positions,omitempty"`
	Share            *Share                 `json:"share,omitempty"`
	Images           []*api.Image           `json:"images,omitempty"`
	ImagesUnion      *ImagesUnion           `json:"images_union,omitempty"`
	Type             string                 `json:"type,omitempty"`
	IosURI           string                 `json:"ios_uri,omitempty"`
	AndroidURI       string                 `json:"android_uri,omitempty"`
	ContentStyle     int64                  `json:"content_style,omitempty"` // 内容样式
	ScrollType       int32                  `json:"scroll_type,omitempty"`   // 滚动方向
	BgStyle          int64                  `json:"background_style,omitempty"`
	IndicatorStyle   int64                  `json:"indicator_style,omitempty"`
	Num              int64                  `json:"num,omitempty"`
	DisplayNum       string                 `json:"display_num,omitempty"`
	TargetNum        int64                  `json:"target_num,omitempty"`
	TargetDisplayNum string                 `json:"target_display_num,omitempty"`
	FontSize         int64                  `json:"font_size,omitempty"`
	CurrentTabIndex  int32                  `json:"current_tab_index,omitempty"`
	FontType         string                 `json:"font_type,omitempty"`
	TabConf          *TabConf               `json:"tab_conf,omitempty"`
	LayerImage       string                 `json:"layer_image,omitempty"`
	ButtonImage      string                 `json:"button_image,omitempty"`
	Style            string                 `json:"style,omitempty"`
	ShareImageInfo   *api.Image             `json:"share_image_info,omitempty"`
	MutexUkeys       []string               `json:"mutex_ukeys,omitempty"`
	Time             string                 `json:"time,omitempty"`
	SponsorTitle     string                 `json:"sponsor_title,omitempty"`
	NewactFeatures   []*Item                `json:"newact_features,omitempty"`
	TableAttrs       []*TableAttr           `json:"table_attrs,omitempty"`
	Header           *Item                  `json:"header,omitempty"`
	Status           string                 `json:"status,omitempty"`
}

type TableAttr struct {
	Ratio     int64  `json:"ratio,omitempty"`
	TextAlign string `json:"text_align,omitempty"`
}

type CarouselImage struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Length      int64  `json:"length"`
	Width       int64  `json:"width"`
	api.ConfSet
}

type IconRemark struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Content     string `json:"content"`
}

type ImagesUnion struct {
	UnSelect *Image `json:"un_select,omitempty"` //未选中
	Select   *Image `json:"select,omitempty"`    //选中
	Event    *Image `json:"event,omitempty"`     //赛事图片
	Button   *Image `json:"button,omitempty"`    //按钮图片
}

type Image struct {
	Image  string `json:"image"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Uri    string `json:"uri,omitempty"`
}

type Share struct {
	ShareOrigin  string `json:"share_origin,omitempty"`
	ShareType    int32  `json:"share_type,omitempty"`
	DisplayLater bool   `json:"display_later,omitempty"`
	Oid          int64  `json:"oid,omitempty"`
	Sid          string `json:"sid,omitempty"`
	ShareTitle   string `json:"share_title,omitempty"`
	ShareImage   string `json:"share_image,omitempty"`
	ShareURL     string `json:"share_url,omitempty"`
	ShareCaption string `json:"share_caption,omitempty"`
}

type Icon struct {
	TopIcon    string `json:"top_icon,omitempty"`    //顶部推荐icon
	BottomIcon string `json:"bottom_icon,omitempty"` //底部推荐icon
	MiddleIcon string `json:"middle_icon,omitempty"` //中间icon
}

type Text struct {
	TopContent    string `json:"top_content,omitempty"`    //顶部推荐语
	BottomContent string `json:"bottom_content,omitempty"` //底部推荐语
}

type Positions struct {
	Position1 Position `json:"position1"` //属性展示位置1
	Position2 Position `json:"position2"` //属性展示位置2
	Position3 Position `json:"position3"` //属性展示位置3
	Position4 Position `json:"position4"` //属性展示位置4
	Position5 Position `json:"position5"` //属性展示位置5
}

type Position string

func (p *Position) FormatRankArc(typ string, rankItem *actapi.RankResult, arc *actapi.ArchiveInfo) {
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

func (p *Position) FormatArc(typ string, arc *arccli.Arc, viewedArcs map[int64]struct{}) {
	switch typ {
	case api.PosUp:
		*p = Position(arc.GetAuthor().Name)
	case api.PosView:
		*p = Position(statString(int64(arc.GetStat().View), "观看"))
	case api.PosPubTime:
		*p = Position(pubDataString(arc.GetPubDate().Time()))
	case api.PosLike:
		*p = Position(statString(int64(arc.GetStat().Like), "点赞"))
	case api.PosDanmaku:
		*p = Position(statString(int64(arc.GetStat().Danmaku), "弹幕"))
	case api.PosViewStat:
		*p = ""
		if _, ok := viewedArcs[arc.GetAid()]; ok {
			*p = "已观看"
		}
	default:
		*p = ""
	}
}

func (p *Position) FormatArt(typ string, meta *artmdl.Meta) {
	switch typ {
	case api.PosUp:
		*p = Position(meta.Author.Name)
	case api.PosView:
		*p = Position(statString(meta.Stats.View, "观看"))
	case api.PosLike:
		*p = Position(statString(meta.Stats.Like, "点赞"))
	case api.PosPubTime:
		*p = Position(pubDataString(meta.PublishTime.Time()))
	default:
		*p = ""
	}
}

func (p *Position) FormatEp(typ string, ep *bgmmdl.EpPlayer) {
	switch typ {
	case api.PosDuration:
		*p = Position(durationString(ep.Duration))
	case api.PosView:
		*p = Position(statString(ep.Stat.Play, "观看"))
	case api.PosFollow:
		*p = Position(statString(ep.Stat.Follow, "追剧"))
	default:
		*p = ""
	}
}

type ModulesReply struct {
	Card map[int64]*Item `json:"card"`
}

// ClickExt .
type ClickExt struct {
	FID int64 `json:"fid,omitempty"`
	// 是否预约 是否追番，追剧
	IsFollow bool       `json:"is_follow,omitempty"`
	Tip      *TipCancel `json:"tip,omitempty"`
	Goto     string     `json:"goto,omitempty"`
	// 当前状态
	CurrentState int8 `json:"current_state,omitempty"`
	// 预埋按钮list
	List         []*ClickList `json:"list,omitempty"`
	Num          int64        `json:"num,omitempty"`
	DisplayNum   string       `json:"-"`
	TargetNum    int64        `json:"target_num,omitempty"`
	FollowText   string       `json:"follow_text,omitempty"`
	FollowIcon   string       `json:"follow_icon,omitempty"`
	UnFollowText string       `json:"un_follow_text,omitempty"`
	UnFollowIcon string       `json:"un_follow_icon,omitempty"`
	ActionType   string       `json:"action_type,omitempty"`
	UnactionType string       `json:"unaction_type,omitempty"`
	URL          string       `json:"url,omitempty"`
	Icon         string       `json:"icon,omitempty"`
	ItemID       int64        `json:"item_id,omitempty"`
	GroupID      int64        `json:"group_id,omitempty"`
	// toast
	CancelDisable bool   `json:"cancel_disable,omitempty"`
	Toast         string `json:"toast,omitempty"`
}

type ClickList struct {
	Image string `json:"image,omitempty"`
	// 领取组件state
	State int8 `json:"state"`
	// 按钮交互类型
	Interaction int8 `json:"interaction,omitempty"`
	//文案
	Text string `json:"text,omitempty"`
}

func (i *ClickExt) FromClickReceive(v *api.NativeClick, state int8) {
	i.FID = v.ForeignID
	switch {
	case v.IsUpAppointment():
		i.Goto = GotoClickUpAppointment
	default:
		i.Goto = GotoClickPendant
	}
	i.CurrentState = state
	i.List = []*ClickList{
		{
			State:       0, // 0 无资格或无奖励，1 未领取，2 已领取
			Image:       v.UnfinishedImage,
			Interaction: 3, // 按钮交互状态  （1）不可点击；（2）点击后直接刷新状态；（3）点击后toast提示并刷新；（4）点击后先弹窗询问，确定后刷新状态
		},
		{
			State:       1,
			Image:       v.OptionalImage,
			Interaction: 3,
		},
		{
			State:       2,
			Image:       v.FinishedImage,
			Interaction: 3,
		},
	}
}

// TipCancel .
type TipCancel struct {
	Msg       string `json:"msg,omitempty"`
	ThinkMsg  string `json:"think_msg,omitempty"`
	SureMsg   string `json:"sure_msg,omitempty"`
	FollowMsg string `json:"follow_msg,omitempty"`
	CancelMsg string `json:"cancel_msg,omitempty"`
}

func (i *TipCancel) FromVoteTip(tip string) {
	i.Msg = fmt.Sprintf("是否取消%s？", tip)
	i.SureMsg = "确定"
	i.ThinkMsg = "取消"
}

func (i *TipCancel) FromTip(tip string) {
	i.Msg = fmt.Sprintf("确定取消%s吗?", tip)
	i.SureMsg = fmt.Sprintf("取消%s", tip)
	i.ThinkMsg = "再想想"
	i.FollowMsg = fmt.Sprintf("%s成功", tip)
	i.CancelMsg = fmt.Sprintf("取消%s成功", tip)
}

func (i *TipCancel) FromCancelTip(tip string) {
	i.Msg = fmt.Sprintf("是否确认取消%s？", tip)
	i.SureMsg = "确认"
	i.ThinkMsg = "取消"
	i.FollowMsg = fmt.Sprintf("%s成功，会在开始时提醒您", tip)
	i.CancelMsg = fmt.Sprintf("已取消%s", tip)
}

// Setting .
type Setting struct {
	AutoPlay        bool `json:"auto_play,omitempty"`
	DisplayTitle    bool `json:"display_title"` //false,也下发，兼容直播卡客户端没有对nil做处理
	DisplayOp       bool `json:"display_op"`
	DisplayNum      bool `json:"display_num"`
	DisplayNodeNum  bool `json:"display_node_num"`
	DisplayDesc     bool `json:"display_desc"`
	IsFaseAway      bool `json:"is_fase_away,omitempty"`  //轮播组件滑出屏幕后顶栏配置样式消失
	IsFollowTab     bool `json:"is_follow_tab,omitempty"` //首页顶栏跟随图片变化
	ShareImage      bool `json:"share_image,omitempty"`
	UnAllowClick    bool `json:"un_allow_click,omitempty"` //tab不支持点击
	TabStyle        int  `json:"tab_style,omitempty"`      //tab组件样式 0颜色 1：图片
	SyncHoverButton bool `json:"sync_hover_button,omitempty"`
	IsHighlight     bool `json:"is_highlight,omitempty"`
	HiddenReserve   bool `json:"hidden_reserve,omitempty"` //是否隐藏预约数
}

// Color .
type Color struct {
	BgColor                string `json:"bg_color,omitempty"`
	TitleColor             string `json:"title_color,omitempty"`
	MoreColor              string `json:"more_color,omitempty"`
	FontColor              string `json:"font_color,omitempty"`
	SelectBgColor          string `json:"select_bg_color,omitempty"`
	SelectFontColor        string `json:"select_font_color,omitempty"`
	NtSelectBgColor        string `json:"nt_select_bg_color,omitempty"`
	NtSelectFontColor      string `json:"nt_select_font_color,omitempty"`
	NtBgColor              string `json:"nt_bg_color,omitempty"`
	NtFontColor            string `json:"nt_font_color,omitempty"`
	MoreFontColor          string `json:"more_font_color,omitempty"`
	TitleBgColor           string `json:"title_bg_color,omitempty"`
	DisplayColor           string `json:"display_color,omitempty"`
	TopFontColor           string `json:"top_font_color,omitempty"`    //顶部字体颜色
	BottomFontColor        string `json:"bottom_font_color,omitempty"` //底部字体颜色
	TopColor               string `json:"top_color,omitempty"`
	PanelBgColor           string `json:"panel_bg_color,omitempty"`             //展开面板背景色
	PanelSelectFontColor   string `json:"panel_select_font_color,omitempty"`    //展开面板字体选中色
	PanelNtSelectFontColor string `json:"panel_nt_select_font_color,omitempty"` //展开面板字体未选中色
	PanelSelectColor       string `json:"panel_select_color,omitempty"`         //展开面板选中色
	FillColor              string `json:"fill_color,omitempty"`
	TimelineColor          string `json:"timeline_color,omitempty"`    //时间轴颜色
	SubtitleColor          string `json:"subtitle_color,omitempty"`    //副标题文字色-三列   推荐语文字色-单列
	SupernatantColor       string `json:"supernatant_color,omitempty"` //浮层标题文字色
	BorderColor            string `json:"border_color,omitempty"`      //边框颜色
	StatusColor            string `json:"status_color,omitempty"`      //状态颜色
}

// UrlExt .
type UrlExt struct {
	Sid      int64  `json:"sid,omitempty"`
	SortType int32  `json:"sort_type,omitempty"`
	TopicID  int64  `json:"topic_id,omitempty"`
	Types    string `json:"types,omitempty"`
	Sortby   int32  `json:"sortby,omitempty"`
	//修护客户端二级列表页面inline播放bug ios build 9120,9150 ,9160 ,9170 ,9180
	RemoteFrom   string `json:"remote_from,omitempty"`
	ConfModuleID int64  `json:"conf_module_id,omitempty"`
	Offset       int64  `json:"offset,omitempty"`
	LastIndex    int64  `json:"last_index,omitempty"`
	UpperMid     int64  `json:"upper_mid,omitempty"`
	ScenaryFrom  string `json:"scenary_from,omitempty"`
	Goto         string `json:"goto,omitempty"`
}

// DynamicMore .
type DynamicMore struct {
	TopicID int64  `json:"topic_id"`
	Sort    string `json:"sort"`
	Name    string `json:"name"`
	Offset  string `json:"offset"`
	PageID  int64  `json:"page_id"`
}

type Liked struct {
	Sid          int64 `json:"sid"`
	Lid          int64 `json:"lid"`
	Score        int64 `json:"score"`
	HasLiked     int64 `json:"has_liked"`
	DisplayScore bool  `json:"display_score"`
}

type UserInfo struct {
	Mid          int64                    `json:"mid"`
	Name         string                   `json:"name"`
	Face         string                   `json:"face"`
	Url          string                   `json:"url"`
	OfficialInfo accountgrpc.OfficialInfo `json:"official_info,omitempty"`
	Vip          accountgrpc.VipInfo      `json:"vip,omitempty"`
}

// nolint:gomnd
func (myinfo *UserInfo) formatRole() {
	if myinfo.OfficialInfo.Role == 7 {
		myinfo.OfficialInfo.Role = 1
	}
}

type MixFolder struct {
	Fid         int64        `json:"fid"`
	RcmdContent *RcmdContent `json:"rcmd_content"` //编辑推荐内容
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

type RcmdContent struct {
	TopContent      string `json:"top_content"`       //顶部推荐语
	TopFontColor    string `json:"top_font_color"`    //顶部字体颜色
	BottomContent   string `json:"bottom_content"`    //底部推荐语
	BottomFontColor string `json:"bottom_font_color"` //底部字体颜色
	MiddleIcon      string `json:"middle_icon"`       //排行榜icon
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
func (i *Item) FromVideoLike(card *dynamic.DyCard, itemObj *actapi.ItemObj) {
	i.Goto = GotoVideoLike
	i.DyCard = card
	i.Liked = &Liked{Sid: itemObj.Item.Sid, Lid: itemObj.Item.ID, HasLiked: itemObj.HasLiked}
	if itemObj.Score == -1 {
		i.Liked.DisplayScore = false
	} else {
		i.Liked.DisplayScore = true
		i.Liked.Score = itemObj.Score
	}

}

func (i *Item) FromVideoMore(mou *api.NativeModule, offset, pageID int64, dyOffset string, isAvMore bool) {
	if isAvMore {
		i.Goto = GotoCardMore
	} else {
		i.Goto = GotoVideoMore
	}
	i.Title = "查看更多"
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("dy_offset", dyOffset)
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	i.URI = "bilibili://following/activity_detail/" + strconv.FormatInt(mou.ID, 10) + "?" + params.Encode()
}

func (i *Item) FromTimelineMore(mou *api.NativeModule, offset int64) {
	i.Goto = GotoTimelineMore
	i.fromSupernatantMore(mou, offset)
}

func (i *Item) FromOgvSeasonMore(mou *api.NativeModule, offset int64) {
	i.Goto = GotoOgvSeasonMore
	i.fromSupernatantMore(mou, offset)
}

func (i *Item) fromSupernatantMore(mou *api.NativeModule, offset int64) {
	i.Title = mou.Remark
	if mou.Remark == "" {
		i.Title = "查看更多"
	}
	i.Content = mou.Title //浮层标题
	//LastIndex 最后一个事件卡片的位置
	i.UrlExt = &UrlExt{LastIndex: offset, ConfModuleID: mou.ID}
}

func (i *Item) FromTimelineExpand(mou *api.NativeModule) {
	i.Goto = GotoTimelineExpand
	i.Title = mou.Remark
	if mou.Remark == "" {
		i.Title = "展开"
	}
	i.Subtitle = "收起"
}

func (i *Item) FromVideo(card *dynamic.DyCard) {
	i.Goto = GotoVideo
	i.DyCard = card
}

func (i *Item) FromNewVideoCard(c *arccli.Arc, firstPlay *arccli.PlayerInfo, build int64, mobiApp string) {
	i.Goto = GotoNewUgcVideo
	i.CoverLeftText1 = cardmdl.DurationString(c.Duration)      //播放时长
	i.CoverLeftText2 = statString(int64(c.Stat.View), "观看")    // 播放数
	i.CoverLeftText3 = statString(int64(c.Stat.Danmaku), "弹幕") //弹幕数
	i.Image = c.Pic                                            //封面
	i.Title = c.Title                                          //标题
	i.ItemID = c.Aid
	i.URI = cardmdl.FillURI(cardmdl.GotoAv, 0, 0, strconv.FormatInt(c.Aid, 10),
		cardmdl.ArcPlayHandler(c, firstPlay, "", nil, int(build), mobiApp, true))
	i.Rights = &ArcRights{UgcPay: c.Rights.UGCPay, IsCooperation: c.Rights.IsCooperation, IsPgc: c.AttrVal(arccli.AttrBitIsPGC) == arccli.AttrYes}
	playExt := firstPlay.GetPlayerExtra().GetDimension()
	if playExt != nil {
		i.Dimension = &ArcDimension{Width: playExt.Width, Rotate: playExt.Rotate, Height: playExt.Height}
	}
	i.Repost = &Repost{
		Aid:        c.Aid,
		Cid:        c.FirstCid,
		AuthorName: c.Author.Name,
	}
	i.Type = TypeUgc
}

func (i *Item) FromNewEPCard(c *bgmmdl.EpPlayer) {
	i.Goto = GotoNewPgcVideo
	i.Title = c.ShowTitle //标题
	i.Image = c.Cover     //封面
	i.ItemID = c.EpID
	i.URI = c.Uri                                              //ep一定会返回，不需要兜底逻辑
	i.CoverLeftText1 = cardmdl.DurationString(c.Duration)      //播放时长
	i.CoverLeftText2 = statString(int64(c.Stat.Play), "观看")    // 播放数
	i.CoverLeftText3 = statString(int64(c.Stat.Danmaku), "弹幕") //弹幕数
	i.Badge = &ReasonStyle{Text: c.Season.TypeName}
	i.Repost = &Repost{
		BizType:    strconv.Itoa(BizOgvType),
		SeasonType: strconv.FormatInt(c.Season.Type, 10),
		Aid:        c.AID,
		Cid:        c.CID,
		EpId:       c.EpID,
		IsPreview:  int32(c.IsPreview),
		SeasonId:   c.Season.SeasonID,
	}
	i.Type = TypeOgv
}

func (i *Item) FromVideoCard(card *dynamic.DyCard, isSingle bool) {
	if isSingle {
		i.Goto = GotoVideoSingle
	} else {
		i.Goto = GotoVideoDouble
	}
	i.DyCard = &dynamic.DyCard{Card: card.Card, Desc: card.Desc}
}

func (i *Item) FromNewVideoActModule(mou *api.NativeModule) {
	i.Goto = GotoNewVideoModule
	if mou == nil {
		return
	}
	i.FromCommonModule(mou)
}

func (i *Item) FromSortModule(videoAct *api.VideoAct) {
	i.Goto = GotoSortModule
	if videoAct == nil || len(videoAct.SortList) == 0 {
		return
	}
	for _, va := range videoAct.SortList {
		te := &Item{}
		te.FromSort(va)
		i.Item = append(i.Item, te)
	}
}

func (i *Item) FromSort(sort *api.NativeVideoExt) {
	i.Goto = GotoSortTab
	i.ItemID = sort.SortType
	i.Param = strconv.FormatInt(sort.SortType, 10)
	if sort.Category == 1 {
		i.Title = sort.SortName
	} else {
		if sort.IsCtimeType() {
			i.Title = "时间"
		} else if sort.IsStochasticType() {
			i.Title = "随机"
		} else if sort.IsEsLikesType() {
			i.Title = "热度"
		} else {
			i.Title = "分数"
		}
	}
}

func (i *Item) FromDynamicMore(mou *api.NativeModule, ext *DynamicMore) {
	i.Goto = GotoDynamicMore
	i.Title = "查看更多"
	params := url.Values{}
	params.Set("title", mou.Title)
	params.Set("sort", ext.Sort)
	params.Set("name", ext.Name)
	params.Set("module_id", strconv.FormatInt(mou.ID, 10))
	params.Set("page_id", strconv.FormatInt(ext.PageID, 10))
	params.Set("sortby", strconv.FormatInt(int64(mou.DySort), 10))
	params.Set("offset", ext.Offset)
	i.URI = "bilibili://following/topic_content_list/" + strconv.FormatInt(ext.TopicID, 10) + "?" + params.Encode()
}

func (i *Item) FromDynamic(card *dynamic.DyCard) {
	i.Goto = GotoDynamic
	i.DyCard = card
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
			i.URI = "bilibili://following/activity_landing/" + strconv.FormatInt(act.ID, 10)
		}
	} else {
		i.URI = "bilibili://pegasus/channel/" + strconv.FormatInt(act.ForeignID, 10) + "?type=topic"
	}
}

func (i *Item) FromActCapsuleItem(card *api.NativePageCard) {
	i.ItemID = card.Id
	i.Title = card.Title
	i.URI = card.SkipURL
}

func (i *Item) FromDynamicModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoDynamicModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Title = mou.Title
	i.IsFeed = mou.IsAttrLast()
	i.UrlExt = ext
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, DisplayColor: ryColors.DisplayColor}
}

func (i *Item) FromResourceModule(c context.Context, featureCfg *conf.Feature, mou *api.NativeModule, mobiApp string, build int64) {
	i.Goto = GotoResourceModule
	if mou == nil {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Title = mou.Title
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, MoreColor: mou.MoreColor, FontColor: mou.FontColor, TitleBgColor: ryColors.TitleBgColor, DisplayColor: ryColors.DisplayColor}
	if !IsNewFeed(c, featureCfg, mobiApp, build) { //低版本需要下发more_color
		i.Color.MoreColor = ryColors.TitleBgColor
	}
}

func (i *Item) FromTimelineModule(mou *api.NativeModule) {
	i.Goto = GotoTimelineModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleBgColor: ryColors.TitleBgColor, TimelineColor: ryColors.TimelineColor}
}

func (i *Item) FromTitleConf(mou *api.NativeModule) {
	i.Title = mou.Title
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, SupernatantColor: ryColors.SupernatantColor}
}

func (i *Item) FromOgvSeasonModule(mou *api.NativeModule) {
	i.Goto = GotoOgvSeasonModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{
		BgColor:          mou.BgColor,               //背景色
		TitleBgColor:     ryColors.TitleBgColor,     //卡片背景色-单列
		MoreColor:        mou.MoreColor,             //查看更多按钮色
		FontColor:        mou.FontColor,             //查看更多文字色
		TitleColor:       mou.TitleColor,            //剧集标题色-三列
		DisplayColor:     ryColors.DisplayColor,     //文字标题文字色
		SupernatantColor: ryColors.SupernatantColor, //浮层标题文字色
	}
}

func (i *Item) FromReplyModule(mou *api.NativeModule) {
	i.Goto = GotoReplyModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.UrlExt = &UrlExt{
		Sid:      mou.Fid,           //评论id
		SortType: int32(mou.AvSort), //评论类型
		//UpperMid:0, //评论mid,暂时不下发，客户端需具备透传能力
	}
}

func (i *Item) FromEditorModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoEditorModule
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

func (i *Item) FromVideoModule(mou *api.NativeModule, ext *UrlExt) {
	i.Goto = GotoVideoModule
	if mou != nil {
		i.ItemID = mou.ID
		i.Param = strconv.FormatInt(mou.ID, 10)
		i.Ukey = mou.Ukey
		i.Title = mou.Title
		i.IsFeed = mou.IsAttrLast()
		i.UrlExt = ext
		i.Bar = mou.Bar
		ryColors := mou.ColorsUnmarshal()
		i.Color = &Color{BgColor: mou.BgColor, FontColor: mou.FontColor, MoreColor: mou.MoreColor, TitleColor: mou.TitleColor, DisplayColor: ryColors.DisplayColor}
	}
}

func (i *Item) FromVideoAvidModule(mou *api.NativeModule, isSingle bool) {
	if isSingle {
		i.Goto = GotoAvIDSingleModule
	} else {
		i.Goto = GotoAvIDDoubleModule
	}
	i.FromCommonModule(mou)
}

func (i *Item) FromVideoActModule(mou *api.NativeModule, isSingle bool) {
	if isSingle {
		i.Goto = GotoActSingleModule
	} else {
		i.Goto = GotoActDoubleModule
	}
	i.FromCommonModule(mou)
}

func (i *Item) FromVideoDynModule(mou *api.NativeModule, isSingle bool) {
	if isSingle {
		i.Goto = GotoDynSingleModule
	} else {
		i.Goto = GotoDynDoubleModule
	}
	i.FromCommonModule(mou)
}

func (i *Item) FromCommonModule(mou *api.NativeModule) {
	if mou != nil {
		i.ItemID = mou.ID
		i.Param = strconv.FormatInt(mou.ID, 10)
		i.Ukey = mou.Ukey
		i.Title = mou.Title
		ryColors := mou.ColorsUnmarshal()
		i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, MoreColor: mou.MoreColor, FontColor: mou.FontColor, DisplayColor: ryColors.DisplayColor}
		if (mou.IsAttrAutoPlay() == api.AttrModuleYes) || !(mou.IsAttrHideTitle() == api.AttrModuleYes) {
			i.Setting = &Setting{AutoPlay: mou.IsAttrAutoPlay() == api.AttrModuleYes, DisplayTitle: !(mou.IsAttrHideTitle() == api.AttrModuleYes)}
		}
		i.Bar = mou.Bar
	}
}

// FromInlineTabModule .
func (i *Item) FromInlineTabModule(mou *api.NativeModule, items, child []*Item) {
	i.Goto = GotoInlineTabModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Item = items
	i.ChildItem = child //低版本兼容逻辑
}

func (i *Item) FromSelectModule(mou *api.NativeModule, items, child []*Item) {
	i.Goto = GotoSelectModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Item = items
	i.ChildItem = child //低版本兼容逻辑
}

func (i *Item) FormatSelect(mou *api.NativeModule) {
	i.Goto = GotoSelect
	i.IsDisplay = 1
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

func (i *Item) FormatInline(c context.Context, featureCfg *conf.Feature, mou *api.NativeModule, mobApp string, build int64) {
	i.Goto = GotoInlineTab
	i.IsDisplay = mou.IsAttrDisplayButton()
	i.Title = mou.Title
	// 0,2:颜色 1:图片
	i.Setting = &Setting{}
	if mou.AvSort == 1 {
		i.Setting.TabStyle = 1
		i.Image = mou.Meta
		if mou.Meta != "" && mou.Width > 0 && mou.Length > 0 {
			i.Width = mou.Width
			i.Length = mou.Length
		} else { //图片为空也需要下发
			i.Width = 1125
			i.Length = 120
		}
	}
	if IsVersion615Low(c, featureCfg, mobApp, build) || mou.AvSort == 0 || mou.AvSort == 2 { //低版本 或者选择颜色
		i.Color = &Color{BgColor: mou.BgColor, SelectFontColor: mou.MoreColor, NtSelectFontColor: mou.FontColor}
	}
}

func (i *Item) FromActModule(mou *api.NativeModule) {
	i.Goto = GotoActModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.IsFeed = mou.IsAttrLast()
	i.Title = mou.Title
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
	i.Bar = mou.Bar
}

func (i *Item) FromActCapsuleModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoActCapsuleModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor}
	i.Bar = mou.Bar
	i.Item = items
}

func (i *Item) FromTitleImage(mou *api.NativeModule) {
	i.Goto = GotoTitleImage
	i.Image = mou.Meta
}

// FromGame
func (i *Item) FromGame(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoGameModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = items
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromGameExt(mix *api.NativeMixtureExt, act *gamdl.Item) {
	i.Goto = GotoGame
	i.ItemID = act.GameBaseId
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
}

func (i *Item) FromHoverButton(mou *api.NativeModule, items []*Item, confSort *api.ConfSort) {
	i.Goto = GotoHoverButtonModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Item = items
	i.MutexUkeys = confSort.MUkeys
}

// Recommend
func (i *Item) FromRecommend(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoRecommendModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	i.Item = items
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

// CarouselImg
func (i *Item) FromCarouselImgModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoCarouselImgModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor}
	i.Item = items
	i.Bar = mou.Bar
}

// CarouselWord
func (i *Item) FromCarouselWordModule(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoCarouselWordModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Color = &Color{BgColor: mou.BgColor}
	i.Item = items
	i.Bar = mou.Bar
}

// Icon
func (i *Item) FromIcon(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoIconModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Color = &Color{
		BgColor:   mou.BgColor,
		FontColor: mou.FontColor,
	}
	i.Item = items
	i.Bar = mou.Bar
}

func (i *Item) FromRecommendRankExt(mix *actapi.RankResult, ext *ClickExt, display bool, img string) {
	i.Goto = GotoRecommend
	//用户头像图标
	if mix.Account != nil {
		i.UserInfo = &UserInfo{
			Mid:  mix.Account.MID,
			Name: mix.Account.Name,
			Face: mix.Account.Face,
		}
		vl := mix.Account.Vip.Label
		vipLabel := accountgrpc.VipLabel{
			Path:        vl.Path,
			LabelTheme:  vl.LabelTheme,
			TextColor:   vl.TextColor,
			BgStyle:     vl.BgStyle,
			BgColor:     vl.BgColor,
			BorderColor: vl.BorderColor,
		}
		i.UserInfo.Vip = accountgrpc.VipInfo{
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
		i.UserInfo.OfficialInfo = accountgrpc.OfficialInfo{
			Role:  mix.Account.Official.Role,
			Title: mix.Account.Official.Title,
			Desc:  mix.Account.Official.Desc,
			Type:  mix.Account.Official.Type,
		}
		// 认证信息转换
		i.UserInfo.formatRole()
	}
	//空间跳转地址
	i.URI = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", mix.Account.MID)
	// 推荐理由
	if display {
		i.Title = mix.ShowScore
	}
	if img != "" {
		i.Icon = &Icon{MiddleIcon: img}
	}
	i.ClickExt = ext
}

func (i *Item) FromRecommendExt(mix *api.NativeMixtureExt, act *accountgrpc.Card, ext *ClickExt) {
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
	i.URI = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", act.Mid)
	// 推荐理由
	i.Title = mix.Reason
	i.ClickExt = ext
}

func (i *Item) FromRcmdVerticalModule(mou *api.NativeModule) {
	i.Goto = GotoRcmdVerticalMou
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromRcmdVertical(items []*Item) {
	i.Goto = GotoRcmdVertical
	i.Item = items
}

func (i *Item) FromRcmdVerticalItem(mixExt *api.NativeMixtureExt, account *accountgrpc.Card, clickExt *ClickExt) {
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
	i.Type = TypePersonalSpace
	i.URI = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", account.Mid)
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
			i.Type = TypeURI
			i.URI = rcmdExt.URI
		}
	}
}

func (i *Item) FromCarouselImg(mou *api.NativeModule) {
	i.Goto = GotoCarouselImg
	i.ContentStyle = mou.AvSort
	i.Color = &Color{SelectBgColor: mou.MoreColor, BgColor: mou.TitleColor}
	i.Setting = &Setting{AutoPlay: mou.IsAttrAutoPlay() == api.AttrModuleYes, IsFollowTab: mou.IsAttrDisplayNum() == api.AttrModuleYes, IsFaseAway: mou.IsAttrDisplayDesc() == api.AttrModuleYes}
}

func (i *Item) FromCarouselImgItem(cv *CarouselImage) {
	i.Image = cv.ImgUrl
	i.URI = cv.RedirectUrl
	i.Length = cv.Length
	i.Width = cv.Width
	i.TabConf = TabConfJoin(&cv.ConfSet)
}

func (i *Item) FromCarouselWord(mou *api.NativeModule) {
	i.Goto = GotoCarouselWord
	i.ContentStyle = mou.AvSort
	i.ScrollType = mou.DySort
	i.Color = &Color{FontColor: mou.FontColor, BgColor: mou.TitleColor}
}

func (i *Item) FromCarouselWordItem(cvStr string) {
	i.Content = cvStr
}

func (i *Item) FromIconExt(ext *IconRemark) {
	i.Image = ext.ImgUrl
	i.URI = ext.RedirectUrl
	i.Content = ext.Content
}

func (i *Item) FromPartExt(part *api.NativeParticipationExt, image, url, joinType string) {
	i.Goto = GotoPartModule
	i.Image = image
	i.URI = url
	i.Title = part.Title
	i.ItemID = int64(part.MType)
	i.Content = joinType
}

func (i *Item) FromClick(mou *api.NativeModule, act []*Item) {
	i.Goto = GotoClick
	if mou.IsBaseBottomButton() {
		i.Goto = GotoBottomButton
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.IsFeed = mou.IsAttrLast()
	i.Bar = mou.Bar
	i.Item = make([]*Item, 0, 1)
	temp := &Item{}
	temp.FromClickBackground(mou)
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
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
}

func (i *Item) FromClickBackground(mou *api.NativeModule) {
	i.Goto = GotoClickBack
	i.Width = mou.Width
	i.Length = mou.Length
	i.Image = mou.Meta
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
}

func (i *Item) FromArea(c context.Context, featureCfg *conf.Feature, area *api.NativeClick, ext *ClickExt, mou *api.NativeModule, mobiApp string, build int64) {
	switch {
	case area.IsReserve(), area.IsActReserve(), area.IsFollow(), area.IsCatchUp(), area.IsBuyCoupon(), area.IsCartoon():
		i.Goto = GotoClickButton
	case area.IsPendant(), area.IsUpAppointment():
		i.Goto = GotoClickButtonV2
	case area.IsProgress(), area.IsStaticProgress():
		i.Goto = GotoClickProgress
		i.ItemID = area.ID
	case area.IsRedirect():
		i.Goto = GotoClickButtonV3
	default:
		i.Goto = GotoClickArea
		i.Type = TypeClickArea
	}
	i.Width = area.Width
	i.Length = area.Length
	i.Lefty = area.Lefty
	i.Leftx = area.Leftx
	i.URI = area.Link
	switch {
	case area.IsLayerImage():
		i.Type = TypeClickImage
		_ = setAreaTip(area, i)
		areaExt, _ := setAreaExt(area, i, mou)
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
		i.Type = TypeClickLink
		_ = setAreaTip(area, i)
		_, _ = setAreaExt(area, i, mou)
	case area.IsAPP():
		i.Type = TypeClickAPP
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
		i.Color = &Color{FontColor: progressExt.FontColor}
		if ext != nil {
			i.Num = ext.Num
			i.DisplayNum = strconv.FormatInt(ext.Num, 10)
			if ext.DisplayNum != "" {
				i.DisplayNum = ext.DisplayNum
			}
			i.TargetNum = ext.TargetNum
			i.TargetDisplayNum = ProgressStatString(ext.TargetNum)
		}
		i.FontType = progressExt.FontType
		i.Type = progressExt.DisplayType
	case area.IsRedirect():
		i.Image = area.OptionalImage
	case area.IsOnlyImage():
		i.ButtonImage = area.OptionalImage
	default:
		// 完成态图片
		i.Image = area.FinishedImage
		i.UnImage = area.UnfinishedImage
	}
	// 低版本兼容 fix-click图片漂移问题
	if i.Goto == GotoClickArea && i.ButtonImage == "" {
		if feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.ClickArea, nil) {
			i.ButtonImage = "https://i0.hdslb.com/bfs/activity-plat/static/20201223/b39c4b95b3f5be9176bb18f203331ce1/dwKriVTuzW.png"
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

func setAreaTip(area *api.NativeClick, item *Item) error {
	if area.Tip == "" {
		return nil
	}
	areaTip := new(api.ClickTip)
	if err := json.Unmarshal([]byte(area.Tip), areaTip); err != nil {
		log.Error("Fail to unmarshal areaTip, tip=%+v error=%+v", area.Tip, err)
		return err
	}
	if item.Color == nil {
		item.Color = &Color{}
	}
	item.Color.TopColor = areaTip.TopColor
	item.Color.TitleColor = areaTip.TitleColor
	item.Title = areaTip.Title
	return nil
}

func setAreaExt(area *api.NativeClick, item *Item, mou *api.NativeModule) (*api.ClickExt, error) {
	if area.Ext == "" {
		return &api.ClickExt{}, nil
	}
	areaExt := new(api.ClickExt)
	if err := json.Unmarshal([]byte(area.Ext), areaExt); err != nil {
		log.Error("Fail to unmarshal areaExt, ext=%+v error=%+v", area.Ext, err)
		return nil, err
	}
	item.ButtonImage = areaExt.ButtonImage
	item.Style = areaExt.Style
	if areaExt.Style == StyleImage {
		item.LayerImage = areaExt.LayerImage
	}
	if mou.IsAttrShareImage() == api.AttrModuleYes && areaExt.ShareImage != nil && areaExt.ShareImage.Image != "" {
		if areaExt.ShareImage != nil {
			item.ShareImageInfo = areaExt.ShareImage
			// 转为kb
			item.ShareImageInfo.Size_ = int64(math.Floor(float64(item.ShareImageInfo.Size_)/1024 + 0.5))
		}
		if item.Setting == nil {
			item.Setting = &Setting{}
		}
		item.Setting.ShareImage = true
		if item.Share == nil {
			item.Share = &Share{}
		}
		item.Share.ShareOrigin = ShareLongPress
	}
	return areaExt, nil
}

func unmarshalAreaImage(img string) *api.Image {
	image := new(api.Image)
	if err := json.Unmarshal([]byte(img), image); err == nil && image.Image != "" {
		return image
	}
	return nil
}

func (i *Item) FromBannerModule(mou *api.NativeModule, page *api.NativePage, info *UserInfo) {
	i.Goto = GotoBannerModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	temp := &Item{}
	temp.FromBannerCard(mou, page, info)
	i.Item = append(i.Item, temp)
	i.Bar = mou.Bar
}

func (i *Item) FromBannerCard(mou *api.NativeModule, page *api.NativePage, info *UserInfo) {
	i.Goto = GotoBannerCard
	i.Image = mou.Meta
	if info != nil {
		i.UserInfo = info
		i.URI = fmt.Sprintf("bilibili://space/%d", info.Mid)
		i.Content = ActStatContent
	}
	i.Title = page.Title
}

func (i *Item) FromNavigation(c context.Context, featureCfg *conf.Feature, mou *api.NativeModule, mobiApp string, build int64) {
	i.Goto = GotoNavigationModule
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	tmpColor := &Color{}
	// 导航组件日间和夜间色用，分隔; 颜色为空且版本低于64的时候 传默认值
	if !IsNewFeed(c, featureCfg, mobiApp, build) {
		tmpColor.BgColor = "#FFFFFF"
		tmpColor.NtBgColor = "#282828"
		tmpColor.FontColor = "#505050"
		tmpColor.NtFontColor = "#828282"
		tmpColor.SelectBgColor = "#FFFFFF"
		tmpColor.NtSelectBgColor = "#282828"
		tmpColor.SelectFontColor = "#FB7299"
		tmpColor.NtSelectFontColor = "#BB5B76"
	}
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

func (i *Item) FromStatementModule(mou *api.NativeModule) {
	i.Goto = GotoStatementModule
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	temp := &Item{}
	temp.FromStatement(mou)
	i.Item = append(i.Item, temp)
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *Item) FromStatement(mou *api.NativeModule) {
	i.Goto = GotoStatement
	i.Content = mou.Remark
	i.IsDisplay = mou.IsAttrStatementDisplayButton()
}

func (i *Item) FromTitleName(mou *api.NativeModule) {
	i.Goto = GotoTitleName
	i.Title = mou.Caption
}

func (i *Item) FromSingleDynModule(mou *api.NativeModule, from string, card *dynamic.DyCard) {
	i.Goto = GotoSingleDynModule
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.ItemID = mou.ID
	i.Ukey = mou.Ukey
	temp := &Item{}
	temp.FromCard(from)
	i.Item = append(i.Item, temp)
	temp1 := &Item{}
	temp1.FromDynamic(card)
	i.Item = append(i.Item, temp1)
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor}
}

func (i *SingleDyn) FromSingleDynTm(card *dynamic.DyCard) {
	i.Title = dynamic.FeedStat
	i.DyCard = card
}

func (i *Item) FromCard(from string) {
	i.Goto = GotoFromCard
	i.Title = from
}

func (i *Item) FromTimelineArc(c *arccli.Arc) {
	i.Goto = GotoTimelineResource
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.URI = model.FillURI(model.GotoAv, strconv.FormatInt(c.GetAid(), 10), model.AvHandler(c))
	i.Positions = new(Positions)
	i.Positions.Position1.FormatArc(api.PosUp, c, nil)
	i.Positions.Position2.FormatArc(api.PosView, c, nil)
	i.Repost = &Repost{BizType: strconv.Itoa(BizUgcType)}
	i.Type = TypeUgc
}

func (i *Item) FromTimelineArt(c *artmdl.Meta) {
	i.Goto = GotoTimelineResource
	i.Title = c.Title //标题
	if len(c.ImageURLs) >= 1 {
		i.Image = c.ImageURLs[0] //封面
	}
	i.ItemID = c.ID
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/cv%d", c.ID)
	i.Positions = new(Positions)
	i.Positions.Position1.FormatArt(api.PosUp, c)
	i.Positions.Position2.FormatArt(api.PosView, c)
	i.Badge = &ReasonStyle{Text: "文章", BgColor: "#FB7299"}
	i.Repost = &Repost{BizType: strconv.Itoa(BizArtType)}
	i.Type = TypeArticle
}

// FromTimelinePic 图.
func (i *Item) FromTimelinePic(c *api.MixReason) {
	i.Goto = GotoTimelinePic
	i.Title = c.Title //标题
	i.Image = c.Image
	i.URI = c.Url
	i.Width = int64(c.Width)
	i.Length = int64(c.Length)
}

// FromTimeline 图文.
func (i *Item) FromTimeline(obj interface{}) {
	i.Goto = GotoTimelineMix
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
	}
}

// FromTimelineText 文.
func (i *Item) FromTimelineText(c *api.MixReason) {
	i.Goto = GotoTimelineText
	i.Title = c.Title       //标题
	i.Subtitle = c.SubTitle //副标题
	i.Content = c.Desc      //正文
	i.URI = c.Url
}

func (i *Item) FromTimelineFormatHead(stime xtime.Time, confSort *api.ConfSort) {
	i.Goto = GotoTimelineHead
	y, m, d := stime.Time().Date()
	h, min, sec := stime.Time().Clock()
	//精确到 0:年 1:月 2: 日 3:时 4:分 5:秒
	switch confSort.TimeSort {
	case api.TimeSortMonth:
		i.Title = fmt.Sprintf("%d年%d月", y, m)
	case api.TimeSortDay:
		i.Title = fmt.Sprintf("%d年%d月%d日", y, m, d)
	case api.TimeSortHour:
		i.Title = fmt.Sprintf("%d年%d月%d日 %02d时", y, m, d, h)
	case api.TimeSortMin:
		i.Title = fmt.Sprintf("%d年%d月%d日 %02d:%02d", y, m, d, h, min)
	case api.TimeSortSec:
		i.Title = fmt.Sprintf("%d年%d月%d日 %02d:%02d:%02d", y, m, d, h, min, sec)
	default:
		i.Title = fmt.Sprintf("%d年", y)
	}
}

func (i *Item) FromTimelineHead(c *api.MixReason) {
	i.Goto = GotoTimelineHead
	i.Title = c.Name
}

func (i *Item) FromOgvSeason(mou *api.NativeModule, ep *pgcAppGrpc.SeasonCardInfoProto, defaultTitle string) {
	i.ItemID = int64(ep.SeasonId)
	i.Param = strconv.FormatInt(int64(ep.SeasonId), 10)
	if mou.IsCardThree() { //三列卡
		i.Goto = GotoOgvSeasoThree
		if ep.Stat == nil {
			//追番/追剧数
			i.CoverLeftText1 = "-"
		} else {
			i.CoverLeftText1 = ep.Stat.FollowView
		}
		//推荐语 运营配置 > 副标题 > 更新集数（更新集数逻辑参见下方）
		if mou.IsAttrDisplayDesc() == api.AttrModuleYes {
			i.Content = defaultTitle
			if i.Content == "" {
				i.Content = ep.RecommendView
			}
		}
	} else {
		i.Goto = GotoOgvSeasoOne
		i.Icon = new(Icon)
		i.Positions = new(Positions)
		if ep.Stat != nil {
			//观看数
			i.Positions.Position2 = Position(statString(ep.Stat.View, "观看"))
			//追番/追剧数
			i.Positions.Position3 = Position(ep.Stat.FollowView)
		} else {
			//观看数
			i.Positions.Position2 = Position("-观看")
		}
		if ep.NewEp != nil {
			//更新集数
			i.Positions.Position4 = Position(ep.NewEp.IndexShow)
		}
		//评分
		if mou.IsAttrDisplayNum() == api.AttrModuleYes && ep.Rating != nil {
			i.CoverRightText = fmt.Sprintf("%0.1f", ep.Rating.Score)
			i.CoverRightText1 = "分"
		}
		//推荐语
		if mou.IsAttrDisplayRecommend() == api.AttrModuleYes {
			i.Content = defaultTitle // 运营配置 > 副标题
			if i.Content == "" {
				i.Content = ep.Subtitle
			}
			if i.Content != "" {
				i.Icon.BottomIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
			}
		}
	}
	ryColors := mou.ColorsUnmarshal()
	if ryColors.SubtitleColor != "" {
		//副标题文字色-三列   推荐语文字色-单列
		i.Color = &Color{SubtitleColor: ryColors.SubtitleColor}
	}
	//追番按钮
	if ep.FollowInfo != nil {
		i.ClickExt = &ClickExt{
			FID:          int64(ep.SeasonId),
			Goto:         GotoClickPgc,
			FollowText:   ep.FollowInfo.FollowText,
			FollowIcon:   ep.FollowInfo.FollowIcon,
			UnFollowText: ep.FollowInfo.UnfollowText,
			UnFollowIcon: ep.FollowInfo.UnfollowIcon,
		}
		if ep.FollowInfo.IsFollow == 1 {
			i.ClickExt.IsFollow = true
		}
	}
	i.Image = ep.Cover //封面
	if mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes && ep.BadgeInfo != nil {
		//付费角标
		i.Badge = &ReasonStyle{Text: ep.BadgeInfo.Text, BgColor: ep.BadgeInfo.BgColor, BgColorNight: ep.BadgeInfo.BgColorNight}
	}
	//标题
	i.Title = ep.Title
	i.Repost = &Repost{
		BizType:    strconv.Itoa(BizOgvType),
		SeasonType: strconv.FormatInt(int64(ep.SeasonType), 10),
		SeasonId:   int64(ep.SeasonId),
	}
	i.URI = ep.GetUrl()
	i.Type = TypeOgv
}

func (i *Item) FromResourceProduct(c *busmdl.ProductItem) {
	i.Goto = GotoResource
	i.Title = c.Title    //标题
	i.Image = c.ImageURL //封面
	i.ItemID = c.ItemID
	i.URI = c.LinkURL
	i.Repost = &Repost{BizType: strconv.Itoa(BizBusinessCommodity)}
	i.Type = TypeCommodity
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
	i.CoverLeftIcon1 = cardmdl.IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = cardmdl.IconDanmaku
	i.CoverLeftText2 = statString(item.Dm, "")
	i.CoverRightText = cardmdl.DurationString(item.PlayLen)
	i.Repost = &Repost{BizType: strconv.Itoa(BizUgcType)}
	i.Type = TypeUgc
}

func buildWidSeason(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.CoverLeftIcon1 = cardmdl.IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = cardmdl.IconFavorite
	i.CoverLeftText2 = statString(item.Follow, "")
	i.Repost = &Repost{BizType: strconv.Itoa(BizSeasonType)}
	i.Type = TypeSeason
}

func buildWidOgv(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.CoverLeftIcon1 = cardmdl.IconPlay
	i.CoverLeftText1 = statString(item.Play, "")
	i.CoverLeftIcon2 = cardmdl.IconDanmaku
	i.CoverLeftText2 = statString(item.Dm, "")
	i.CoverRightText = cardmdl.DurationString(item.PlayLen)
	i.Repost = &Repost{BizType: strconv.Itoa(BizOgvType)}
	i.Type = TypeOgv
}

func buildWidWeb(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizWebType)}
	i.Type = TypeWeb
}

func buildWidOgvFilm(i *Item, item *pgcAppGrpc.QueryWidItem) {
	if i == nil || item == nil {
		return
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizOgvFilmType)}
	i.Type = TypeOgvFilm
}

func (i *Item) FromResourceArc(c *arccli.Arc, display bool, f *favmdl.Folder) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	if f == nil {
		if c.AttrVal(arccli.AttrBitIsPGC) == arccli.AttrYes && c.RedirectURL != "" {
			i.URI = c.RedirectURL
		} else {
			i.URI = model.FillURI(model.GotoAv, strconv.FormatInt(c.GetAid(), 10), model.AvHandler(c))
		}
	} else {
		i.URI = fmt.Sprintf("bilibili://music/playlist/playpage/%d?avid=%d&oid=%d&page_type=4", f.Mlid, c.Aid, c.Aid)
	}
	i.CoverRightText = cardmdl.DurationString(c.Duration)
	i.CoverLeftText1 = statString(int64(c.Stat.View), "")
	i.CoverLeftIcon1 = cardmdl.IconPlay
	i.CoverLeftText2 = statString(int64(c.Stat.Danmaku), "")
	i.CoverLeftIcon2 = cardmdl.IconDanmaku
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizUgcType)}
	i.Type = TypeUgc
}

func (i *Item) FromResourceLive(c *playgrpc.RoomList, ctx context.Context, featureCfg *conf.Feature) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	i.Image = c.Icon  //封面
	i.ItemID = c.RoomId
	i.URI = fmt.Sprintf("https://live.bilibili.com/%d", c.RoomId)
	i.CoverRightText = c.UserName
	i.CoverLeftText1, i.CoverLeftIcon1 = func() (string, cardmdl.Icon) {
		if !feature.GetBuildLimit(ctx, featureCfg.FeatureBuildLimit.LiveWatched, nil) {
			return c.Online, cardmdl.IconOnline
		}
		if c.WatchedShow != nil && c.WatchedShow.Switch {
			return c.WatchedShow.TextLarge, cardmdl.IconLiveWatched
		}
		return c.Online, cardmdl.IconLiveOnline
	}()
	if c.Pendant != "" {
		i.Badge = &ReasonStyle{Text: c.Pendant}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizLiveType)}
	i.Type = TypeLive
}

func (i *Item) FromEditorRankArc(rankItem *actapi.RankResult, mou *api.NativeModule, arc *actapi.ArchiveInfo, display bool, rcmd *RcmdContent) {
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
	i.FromEditorCommon(mou, rcmd)
	i.Title = arc.Title
	i.Image = arc.Pic
	i.ItemID, _ = bvsafe.BvToAv(arc.BvID)
	i.URI = arc.ShowLink
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizUgcType)}
	i.Type = TypeUgc
	i.Share = &Share{
		ShareOrigin:  ShareUgc,
		ShareType:    ShareTypeActivity,
		DisplayLater: true,
		Oid:          i.ItemID,
	}
}

func (i *Item) FromEditorArc(arc *arccli.Arc, display bool, f *favmdl.Folder, mou *api.NativeModule, rcmd *RcmdContent, mobiApp, device string, viewedArcs map[int64]struct{}) {
	posMeta := new(Positions)
	if err := json.Unmarshal([]byte(mou.TName), posMeta); err != nil {
		log.Error("[FromEditorArc] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
	}
	i.Positions = new(Positions)
	i.Positions.Position1.FormatArc(string(posMeta.Position1), arc, viewedArcs)
	i.Positions.Position2.FormatArc(string(posMeta.Position2), arc, viewedArcs)
	i.Positions.Position3.FormatArc(string(posMeta.Position3), arc, viewedArcs)
	i.Positions.Position4.FormatArc(string(posMeta.Position4), arc, viewedArcs)
	i.Positions.Position5.FormatArc(string(posMeta.Position5), arc, viewedArcs)
	i.FromEditorCommon(mou, rcmd)
	i.Title = arc.GetTitle()
	i.Image = arc.GetPic()
	i.ItemID = arc.GetAid()
	if f == nil || (mobiApp != "" && model.IsIPad(model.Plat(mobiApp, device))) {
		i.URI = model.FillURI(model.GotoAv, strconv.FormatInt(arc.GetAid(), 10), model.AvHandler(arc))
	} else {
		i.URI = fmt.Sprintf("bilibili://music/playlist/playpage/%d?avid=%d&oid=%d&page_type=4", f.Mlid, arc.GetAid(), arc.GetAid())
	}
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizUgcType)}
	i.Type = TypeUgc
	i.Share = &Share{
		ShareOrigin:  ShareUgc,
		ShareType:    ShareTypeActivity,
		DisplayLater: true,
		Oid:          arc.GetAid(),
	}
}

func (i *Item) FromEditorArt(meta *artmdl.Meta, artDisplay bool, mou *api.NativeModule, rcmd *RcmdContent) {
	posMeta := new(Positions)
	if err := json.Unmarshal([]byte(mou.TName), posMeta); err != nil {
		log.Warn("[FromEditorArt] json.Unmarshal(%+v) error(%+v)", mou.Meta, err)
	}
	i.Positions = new(Positions)
	i.Positions.Position1.FormatArt(string(posMeta.Position1), meta)
	i.Positions.Position2.FormatArt(string(posMeta.Position2), meta)
	i.Positions.Position3.FormatArt(string(posMeta.Position3), meta)
	i.Positions.Position4.FormatArt(string(posMeta.Position4), meta)
	i.Positions.Position5.FormatArt(string(posMeta.Position5), meta)
	i.FromEditorCommon(mou, rcmd)
	i.Title = meta.Title
	if len(meta.ImageURLs) >= 1 {
		i.Image = meta.ImageURLs[0]
	}
	i.ItemID = meta.ID
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/cv%d", meta.ID)
	if artDisplay {
		i.Badge = &ReasonStyle{Text: "文章"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizArtType)}
	i.Type = TypeArticle
	i.Share = &Share{
		ShareOrigin:  ShareArticle,
		ShareType:    ShareTypeActivity,
		DisplayLater: false,
		Oid:          meta.ID,
	}
}

func (i *Item) FromEditorEp(ep *bgmmdl.EpPlayer, display bool, mou *api.NativeModule, rcmd *RcmdContent, posConf string) {
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
	i.FromEditorCommon(mou, rcmd)
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
	i.Repost = &Repost{BizType: strconv.Itoa(BizOgvType), SeasonType: strconv.FormatInt(ep.Season.Type, 10)}
	i.Type = TypeOgv
	i.Share = &Share{
		ShareType:    ShareTypeActivity,
		DisplayLater: false,
	}
}

func (i *Item) FromEditorCommon(mou *api.NativeModule, rcmd *RcmdContent) {
	i.Goto = GotoEditor
	i.Setting = &Setting{
		DisplayOp: mou.IsAttrDisplayOp() == api.AttrModuleYes,
	}
	// 内容推荐
	if rcmd != nil {
		i.Color = new(Color)
		i.Icon = new(Icon)
		i.Text = new(Text)
		if rcmd.TopContent != "" {
			i.Color.TopFontColor = rcmd.TopFontColor
			i.Icon.TopIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/q~1vlO6h25.png"
			i.Text.TopContent = rcmd.TopContent
		}
		if rcmd.BottomContent != "" {
			i.Color.BottomFontColor = rcmd.BottomFontColor
			i.Icon.BottomIcon = "http://i0.hdslb.com/bfs/activity-plat/static/20200619/ce4d241380919d495e1e6f11992d3e0f/rYGIuJ~Ii4.png"
			i.Text.BottomContent = rcmd.BottomContent
		}
		if rcmd.MiddleIcon != "" {
			i.Icon.MiddleIcon = rcmd.MiddleIcon
		}
	}
}

func ProgressStatString(number int64) string {
	if number == 0 {
		return "0"
	}
	return statString(number, "")
}

// nolint:gomnd
func statString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

func pubDataString(t time.Time) (s string) {
	if t.IsZero() {
		return
	}
	now := time.Now()
	sub := now.Sub(t)
	if sub < time.Minute {
		s = "刚刚"
		return
	}
	if sub < time.Hour {
		s = strconv.FormatFloat(sub.Minutes(), 'f', 0, 64) + "分钟前"
		return
	}
	if sub < 24*time.Hour {
		s = strconv.FormatFloat(sub.Hours(), 'f', 0, 64) + "小时前"
		return
	}
	if now.Year() == t.Year() {
		if now.YearDay()-t.YearDay() == 1 {
			s = "昨天"
			return
		}
		s = t.Format("01-02")
		return
	}
	s = t.Format("2006-01-02")
	return
}

// nolint:gomnd
func durationString(second int64) (s string) {
	var hour, min, sec int
	if second < 1 {
		return
	}
	d, err := time.ParseDuration(strconv.FormatInt(second, 10) + "s")
	if err != nil {
		log.Error("%+v", err)
		return
	}
	r := strings.NewReplacer("h", ":", "m", ":", "s", ":")
	ts := strings.Split(strings.TrimSuffix(r.Replace(d.String()), ":"), ":")
	if len(ts) == 1 {
		sec, _ = strconv.Atoi(ts[0])
	} else if len(ts) == 2 {
		min, _ = strconv.Atoi(ts[0])
		sec, _ = strconv.Atoi(ts[1])
	} else if len(ts) == 3 {
		hour, _ = strconv.Atoi(ts[0])
		min, _ = strconv.Atoi(ts[1])
		sec, _ = strconv.Atoi(ts[2])
	}
	if hour == 0 {
		s = fmt.Sprintf("%d:%02d", min, sec)
		return
	}
	s = fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
	return
}

func (i *Item) FromResourceEp(c *bgmmdl.EpPlayer, display bool) {
	i.Goto = GotoResource
	i.Title = c.ShowTitle //标题
	i.Image = c.Cover     //封面
	i.ItemID = c.EpID
	i.URI = c.Uri //ep一定会返回，不需要兜底逻辑
	i.CoverRightText = cardmdl.DurationString(c.Duration)
	i.CoverLeftText1 = statString(int64(c.Stat.Play), "")
	i.CoverLeftIcon1 = cardmdl.IconPlay
	i.CoverLeftText2 = statString(int64(c.Stat.Follow), "")
	i.CoverLeftIcon2 = cardmdl.IconFavorite
	if display {
		i.Badge = &ReasonStyle{Text: c.Season.TypeName}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizOgvType), SeasonType: strconv.FormatInt(c.Season.Type, 10)}
	i.Type = TypeOgv
}

// FromResourceArt .
func (i *Item) FromResourceArt(c *artmdl.Meta, artDisplay bool) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	if len(c.ImageURLs) >= 1 {
		i.Image = c.ImageURLs[0] //封面
	}
	i.ItemID = c.ID
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/cv%d", c.ID)
	if c.Stats != nil {
		i.CoverLeftText1 = statString(int64(c.Stats.View), "")
		i.CoverLeftText2 = statString(int64(c.Stats.Reply), "")
	}
	i.CoverLeftIcon1 = cardmdl.IconRead
	i.CoverLeftIcon2 = cardmdl.IconComment
	if artDisplay {
		i.Badge = &ReasonStyle{Text: "文章"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(BizArtType)}
	i.Type = TypeArticle
}

func (i *Item) FromLiveModule(mou *api.NativeModule) {
	i.Goto = GotoLiveModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, FontColor: mou.FontColor, DisplayColor: ryColors.DisplayColor}
	i.Setting = &Setting{DisplayTitle: !(mou.IsAttrHideTitle() == api.AttrModuleYes)}
	i.Bar = mou.Bar
}

// FromLive
func (i *Item) FromLive(mou *api.NativeModule, card *livefeed.LiveCardInfo) {
	i.Goto = GotoLive
	i.ItemID = mou.Fid
	i.Param = strconv.FormatInt(mou.Fid, 10)
	if card.LiveStatus == 1 && mou.TName != "" {
		card.Cover = mou.TName
	}
	if mou.Stime < card.LastEndTime {
		i.HasLive = 1
	}
	i.LiveCard = card
}

func (i *Item) FromProgressModule(mou *api.NativeModule, item []*Item) {
	i.Goto = GotoProgressModule
	if mou == nil {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Title = mou.Title
	i.Bar = mou.Bar
	i.Color = &Color{BgColor: mou.BgColor}
	i.Item = item
}

// nolint:gomnd
func (i *Item) FromProgress(mou *api.NativeModule, group *actapi.ActivityProgressGroup) {
	i.Goto = GotoProgress
	i.ItemID = mou.ID
	switch mou.AvSort {
	case ProgStyleRound:
		i.Type = TypeRound
	case ProgStyleRectangle:
		i.Type = TypeRectangle
	case ProgStyleNode:
		i.Type = TypeNode
	}
	i.BgStyle, _ = strconv.ParseInt(mou.MoreColor, 10, 64)
	i.IndicatorStyle, _ = strconv.ParseInt(mou.TitleColor, 10, 64)
	switch mou.Length {
	case 1:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/I2CHCvboo.png"
	case 2:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/bxL-soEal.png"
	case 3:
		i.Image = "https://i0.hdslb.com/bfs/activity-plat/static/20200811/8a3e1fa14e30dc3be9c5324f604e5991/vniA~jzxp.png"
	}
	i.Color = &Color{FillColor: mou.FontColor}
	i.Setting = &Setting{
		DisplayNum:     mou.IsAttrDisplayNum() == api.AttrModuleYes,
		DisplayNodeNum: mou.IsAttrDisplayNodeNum() == api.AttrModuleYes,
		DisplayDesc:    mou.IsAttrDisplayDesc() == api.AttrModuleYes,
	}
	i.Num = group.Total
	i.DisplayNum = ProgressStatString(group.Total)
	func() {
		items := make([]*Item, 0, len(group.Nodes))
		for _, v := range group.Nodes {
			item := &Item{
				Title:      v.Desc,
				Num:        v.Val,
				DisplayNum: ProgressStatString(v.Val),
			}
			items = append(items, item)
		}
		i.Item = items
	}()
}

func TabConfJoin(cv *api.ConfSet) *TabConf {
	if cv == nil {
		return nil
	}
	if cv.BgType == api.BgTypeColor {
		return &TabConf{
			TabBottomColor: cv.TabBottomColor,
			TabMiddleColor: cv.TabMiddleColor,
			TabTopColor:    cv.TabTopColor,
			FontColor:      cv.FontColor,
			BarType:        cv.BarType,
		}
	} else if cv.BgType == api.BgTypeImage {
		return &TabConf{
			BgImage1:  cv.BgImage1,
			BgImage2:  cv.BgImage2,
			FontColor: cv.FontColor,
			BarType:   cv.BarType,
		}
	}
	return nil
}

// ChooseMenu 过滤menu页面支持的组件.
func ChooseMenu(bases []*api.Module) (needBase []*api.Module) {
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		tt := v.NativeModule
		// 过滤支持的组件类型
		switch {
		// 滑轮，图标
		case tt.IsClick(), tt.IsAct(), tt.IsActCapsule(), tt.IsNewVideoDyn(), tt.IsNewVideoAct(), tt.IsNewVideoID(), tt.IsStatement(),
			tt.IsRecommend(), tt.IsResourceID(), tt.IsResourceDyn(), tt.IsResourceAct(), tt.IsResourceOrigin(), tt.IsLive(), tt.IsEditor(), tt.IsEditorOrigin(),
			tt.IsCarouselImg(), tt.IsCarouselWord(), tt.IsIcon(), tt.IsProgress(), tt.IsTimelineSource(), tt.IsTimelineIDs(),
			tt.IsRcmdVertical(), tt.IsOgvSeasonSource(), tt.IsOgvSeasonID(), tt.IsReply(), tt.IsCarouselSource(), tt.IsRcmdSource(), tt.IsRcmdVerticalSource(), tt.IsGame(), tt.IsReserve(), tt.IsVote(),
			tt.IsInlineTab(), tt.IsSelect(), tt.IsNavigation(), tt.IsMatchMedal(), tt.IsMatchEvent():
			needBase = append(needBase, v)
		default:
			continue
		}
	}
	return
}

// ChooseOgv 过滤ogv页面支持的组件.
func ChooseOgv(bases []*api.Module) (needBase []*api.Module) {
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		tt := v.NativeModule
		// 过滤支持的组件类型
		switch {
		// 版头,投稿，tab，导航,动态列表
		case tt.IsBaseHead(), tt.IsPart(), tt.IsInlineTab(), tt.IsNavigation(), tt.IsDynamic(), tt.IsVideo():
			continue
		default:
			needBase = append(needBase, v)
		}
	}
	return
}

func ChooseBottom(bases []*api.Module) (needBase []*api.Module) {
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		tt := v.NativeModule
		// 过滤支持的组件类型
		switch {
		// 版头,投稿，评论
		case tt.IsBaseHead(), tt.IsPart(), tt.IsReply():
			continue
		default:
			needBase = append(needBase, v)
		}
	}
	return
}

func ChooseUgc(bases []*api.Module) []*api.Module {
	needBase := make([]*api.Module, 0, len(bases))
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		tt := v.NativeModule
		switch {
		// 自定义点击、资源小卡、相关活动、文本组件、编辑推荐卡片、推荐组件、直播大卡、视频大卡、时间轴组件
		case tt.IsClick(), tt.IsResourceAct(), tt.IsResourceDyn(), tt.IsResourceID(), tt.IsResourceOrigin(), tt.IsAct(), tt.IsActCapsule(), tt.IsStatement(),
			tt.IsEditor(), tt.IsEditorOrigin(), tt.IsRecommend(), tt.IsRcmdVertical(), tt.IsLive(),
			tt.IsVideoAct(), tt.IsVideoAvid(), tt.IsVideoDyn(), tt.IsNewVideoAct(), tt.IsNewVideoDyn(), tt.IsNewVideoID(),
			tt.IsCarouselImg(), tt.IsCarouselWord(), tt.IsIcon(), tt.IsProgress(), tt.IsTimelineIDs(), tt.IsTimelineSource(), tt.IsOgvSeasonSource(), tt.IsOgvSeasonID(),
			tt.IsCarouselSource(), tt.IsRcmdSource(), tt.IsRcmdVerticalSource(), tt.IsGame(), tt.IsReserve(), tt.IsVote():
			needBase = append(needBase, v)
		}
	}
	return needBase
}

func ChoosePlayer(bases []*api.Module) []*api.Module {
	needBase := make([]*api.Module, 0, len(bases))
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		tt := v.NativeModule
		switch {
		// 自定义点击、资源小卡、视频卡、视频卡-新、推荐组件、推荐-竖卡组件、相关活动、文本组件、编辑推荐卡、时间轴组件
		case tt.IsClick(), tt.IsResourceAct(), tt.IsResourceDyn(), tt.IsResourceID(), tt.IsResourceOrigin(),
			tt.IsVideoAct(), tt.IsVideoAvid(), tt.IsVideoDyn(), tt.IsNewVideoAct(), tt.IsNewVideoDyn(), tt.IsNewVideoID(),
			tt.IsRecommend(), tt.IsRcmdVertical(), tt.IsAct(), tt.IsActCapsule(), tt.IsStatement(), tt.IsEditor(), tt.IsEditorOrigin(), tt.IsTimelineSource(), tt.IsTimelineIDs(),
			tt.IsCarouselSource(), tt.IsRcmdSource(), tt.IsRcmdVerticalSource(), tt.IsGame(), tt.IsReserve(), tt.IsVote():
			needBase = append(needBase, v)
		}
	}
	return needBase
}

func ChooseLiveTab(bases []*api.Module) []*api.Module {
	modules := make([]*api.Module, 0, len(bases))
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		nm := v.NativeModule
		if nm.IsNavigation() || nm.IsInlineTab() || nm.IsSelect() || nm.IsOgvSeasonID() || nm.IsOgvSeasonSource() || nm.IsBaseHoverButton() {
			continue
		}
		modules = append(modules, v)
	}
	return modules
}

func ChooseNewact(bases []*api.Module) []*api.Module {
	modules := make([]*api.Module, 0, len(bases))
	for _, v := range bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		nm := v.NativeModule
		if nm.IsNewactStatementModule() || nm.IsNewactAwardModule() {
			modules = append(modules, v)
		}
	}
	return modules
}

func RemoveDuplicates(ids []int64) []int64 {
	checkMap := make(map[int64]struct{})
	var rly []int64
	for _, v := range ids {
		if _, ok := checkMap[v]; ok {
			continue
		}
		rly = append(rly, v)
		checkMap[v] = struct{}{}
	}
	return rly
}

func SetUpCurrentActPage(in []*api.Module, pageID int64) {
	var actPage *api.ActPage
	for _, v := range in {
		if v == nil || v.NativeModule == nil {
			continue
		}
		if !v.NativeModule.IsActCapsule() || v.ActPage == nil {
			continue
		}
		if actPage == nil {
			actPage = v.ActPage
		}
		DelCurrentActPage(v.ActPage, pageID)
	}
	if actPage == nil {
		return
	}
	items := append([]*api.ActPageItem{{PageID: pageID}}, actPage.List...)
	actPage.List = items
}

func DelCurrentActPage(actPage *api.ActPage, pageID int64) {
	if actPage == nil || len(actPage.List) == 0 {
		return
	}
	items := make([]*api.ActPageItem, 0, len(actPage.List))
	for _, v := range actPage.List {
		if v.PageID == pageID {
			continue
		}
		items = append(items, v)
	}
	actPage.List = items
}

func (i *Item) FromReserve(mou *api.NativeModule, items []*Item) {
	i.Goto = GotoReserveModule
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Bar = mou.Bar
	i.Item = items
	ryColors := mou.ColorsUnmarshal()
	i.Color = &Color{BgColor: mou.BgColor, TitleColor: mou.TitleColor, TitleBgColor: ryColors.TitleBgColor}
}

// isDisplayReserve 是否展示预约数
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

func (i *Item) FromReserveExt(mix *api.NativeMixtureExt, act *ReserveRly, displayUp, build, mid int64, mobiApp string) {
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
	//跳转空间
	i.URI = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", act.Item.Upmid)
	var (
		hiddenReserve bool
	)
	switch act.DisplayType {
	case ReserveDisplayA:
		switch act.Item.Type {
		case actapi.UpActReserveRelationType_Archive:
			i.Title = act.Item.Title
			i.Content = "视频预约"
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		case actapi.UpActReserveRelationType_Live:
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		case actapi.UpActReserveRelationType_Course:
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 开售", reserveTime(act.Item.LivePlanStartTime))
			i.Num = act.Item.Total
			i.Subtitle = "人已预约"
		default:
		}
		tip := &TipCancel{}
		tip.FromCancelTip("预约")
		if act.Item.IsFollow == 1 {
			i.ClickExt = &ClickExt{FollowText: "已预约", UnFollowText: "预约", IsFollow: true, Goto: GotoClickReserve, FID: mix.ForeignID, Tip: tip, Icon: ReserveIcon}
		} else {
			i.ClickExt = &ClickExt{FollowText: "已预约", UnFollowText: "预约", Goto: GotoClickReserve, FID: mix.ForeignID, Tip: tip, Icon: ReserveIcon}
		}
	case ReserveDisplayC:
		switch act.Item.Type {
		case actapi.UpActReserveRelationType_Archive:
			i.Title = act.Item.Title
			//获取观看人数
			var (
				views int32
				url   string
			)
			if act.Arc != nil {
				views = act.Arc.Stat.View
				url = cardmdl.FillURI(cardmdl.GotoAv, 0, 0, strconv.FormatInt(act.Arc.Aid, 10),
					cardmdl.ArcPlayHandler(act.Arc, nil, "", nil, int(build), mobiApp, false))
			}
			i.Content = "视频预约"
			i.Num = int64(views)
			i.Subtitle = "观看"
			i.ClickExt = &ClickExt{FollowText: "去观看", Goto: GotoClickURL, URL: url}
		case actapi.UpActReserveRelationType_Live:
			i.Title = act.Item.Title
			//获取直播人数
			var (
				popularCount int64
				url          string
			)
			if act.Live != nil {
				if act.Live.SessionInfoPerLive != nil {
					popularCount = act.Live.SessionInfoPerLive.PopularityCount
				}
				url = act.Live.JumpUrl[LiveEnterFrom]
			}
			i.Setting = &Setting{IsHighlight: true}
			i.Content = "直播中"
			i.Num = popularCount
			i.Subtitle = "人气"
			i.ClickExt = &ClickExt{FollowText: "去观看", Goto: GotoClickURL, URL: url}
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
		case actapi.UpActReserveRelationType_Course:
			i.Title = act.Item.Title
			i.Num = act.Item.OidView
			i.Subtitle = "人看过"
			i.ClickExt = &ClickExt{FollowText: "去观看", Goto: GotoClickURL, URL: act.Item.BaseJumpUrl}
		default:
		}
	case ReserveDisplayD:
		if act.Item.Type == actapi.UpActReserveRelationType_Live {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			// 满足预约数条件
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		}
		if act.Live != nil && act.Live.SessionInfoPerLive != nil {
			url := cardmdl.FillURI(cardmdl.GotoAv, 0, 0, act.Live.SessionInfoPerLive.Bvid, nil)
			i.ClickExt = &ClickExt{FollowText: "看回放", Goto: GotoClickURL, URL: url}
		}
	case ReserveDisplayE:
		if act.Item.Type == actapi.UpActReserveRelationType_Live {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 直播", reserveTime(act.Item.LivePlanStartTime))
			//满足预约数条件
			hiddenReserve = !isDisplayReserve(act.Item.Total, act.Item.ReserveTotalShowLimit, mid, act.Item.Upmid)
			i.Num = act.Item.Total
			i.Subtitle = "人预约"
		} else if act.Item.Type == actapi.UpActReserveRelationType_Course {
			i.Title = act.Item.Title
			i.Content = fmt.Sprintf("%s 开售", reserveTime(act.Item.LivePlanStartTime))
			i.Num = act.Item.Total
			i.Subtitle = "人已预约"
		}
		i.ClickExt = &ClickExt{FollowText: "已结束", Goto: GotoClickUnable, Tip: &TipCancel{Msg: "不在预约时间"}}
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

func (i *Item) FromMVote(area *api.NativeClick, ext *ClickExt) {
	i.Width = area.Width
	i.Length = area.Length
	i.Lefty = area.Lefty
	i.Leftx = area.Leftx
	switch {
	case area.IsVoteButton():
		i.Goto = GotoVoteButton
	case area.IsVoteProcess():
		i.Goto = GotoVoteProcess
	case area.IsVoteUser():
		i.Goto = GotoVoteUser
	default:
	}
	i.ClickExt = ext
	i.ItemID = area.ID
	i.Param = strconv.FormatInt(area.ID, 10)
	i.Ukey = area.ExtUnmarshal().Ukey
}

func (i *Item) FromMVoteProcess(area *api.NativeClick, firExRankInfo map[int64]*actapi.ExternalRankInfo, mou *api.NativeModule) {
	i.FromMVote(area, nil)
	i.Style = area.ExtUnmarshal().Style
	i.IsDisplay = mou.IsAttrDisplayNum()
	jn := int64(0)
	for _, v := range area.ExtUnmarshal().Items {
		if v == nil {
			continue
		}
		tmp := &Item{Color: &Color{BgColor: v.BgColor}}
		tmp.ClickExt = &ClickExt{FID: mou.Fid, GroupID: mou.ConfUnmarshal().Sid}
		if val, ok := firExRankInfo[jn]; ok && val != nil {
			tmp.ClickExt.ItemID = val.SourceItemId
			tmp.ClickExt.Num = val.Vote
		}
		jn++
		i.Item = append(i.Item, tmp)
	}
}

func (i *Item) FromVote(mou *api.NativeModule, act []*Item) {
	i.Goto = GotoVoteModule
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

func (i *Item) FromUpVoteProgress(area *api.NativeClick, options []*dyncommongrpc.VoteOptionInfo, mou *api.NativeModule) {
	areaExt := area.ExtUnmarshal()
	i.FromMVote(area, nil)
	i.Style = areaExt.Style
	i.IsDisplay = mou.IsAttrDisplayNum()
	var no int64
	for _, v := range areaExt.Items {
		if v == nil || no >= VoteOptionNum {
			continue
		}
		item := &Item{Color: &Color{BgColor: v.BgColor}}
		item.ClickExt = &ClickExt{
			FID:    mou.Fid,
			ItemID: int64(options[no].OptIdx),
			Num:    int64(options[no].Cnt),
		}
		i.Item = append(i.Item, item)
		no++
	}
}
