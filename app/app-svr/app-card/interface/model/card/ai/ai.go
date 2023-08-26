package ai

import (
	"encoding/json"
	"time"

	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/archive/service/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	OgvWatched = 2
)

type AiItem interface {
	TrackId() string
}

type Item struct {
	PosRecID       int64           `json:"dalao_uniq_id,omitempty"`
	PosRecUniqueID string          `json:"dalao_new_uniq_id,omitempty"`
	ID             int64           `json:"id,omitempty"`
	Idx            int64           `json:"idx,omitempty"`
	TrackID        string          `json:"trackid,omitempty"`
	Name           string          `json:"name,omitempty"`
	Goto           string          `json:"goto,omitempty"`
	Tid            int64           `json:"tid,omitempty"`
	From           int8            `json:"from,omitempty"`
	Source         string          `json:"source,omitempty"`
	AvFeature      json.RawMessage `json:"av_feature,omitempty"`
	Config         *Config         `json:"config,omitempty"`
	RcmdReason     *RcmdReason     `json:"rcmd_reason,omitempty"`
	StatType       int8            `json:"stat_type,omitempty"`
	Style          int8            `json:"style,omitempty"`
	ConvergeInfo   *ConvergeInfo   `json:"converge_info,omitempty"`
	StoryInfo      *StoryInfo      `json:"story_info,omitempty"`
	StoryParam     string          `json:"story_param,omitempty"`
	// extra
	Archive       *api.Arc         `json:"archive,omitempty"`
	Tag           *taggrpc.Tag     `json:"tag,omitempty"`
	Ad            *cm.AdInfo       `json:"-"`
	Ads           []*cm.AdInfo     `json:"-"`
	Banners       []*banner.Banner `json:"-"`
	Version       string           `json:"-"`
	HideButton    bool             `json:"-"`
	CornerMark    int8             `json:"corner_mark,omitempty"`
	CoverGif      string           `json:"cover_gif,omitempty"`
	JumpID        int64            `json:"jumpid,omitempty"`
	JumpGoto      string           `json:"jumpgoto,omitempty"`
	ConvergeParam string           `json:"converge_param,omitempty"`
	CardType      string           `json:"-"`
	DynamicCover  int32            `json:"-"`
	FfCover       string           `json:"ff_cover,omitempty"`
	MsgIDs        string           `json:"msg_ids,omitempty"` // 当 goto 为 new_tunnel 时使用
	// ai ad
	BizIdx            int32           `json:"biz_idx,omitempty"`
	StaticCover       int             `json:"static_cover,omitempty"`
	IconType          int             `json:"icon_type,omitempty"`
	CustomizedTitle   string          `json:"customized_title,omitempty"`
	CustomizedCover   string          `json:"customized_cover,omitempty"`
	CustomizedOGVDesc string          `json:"customized_ogv_desc,omitempty"`
	BannerInfo        *BannerInfo     `json:"banner_info,omitempty"`
	BigTunnelObject   string          `json:"big_tunnel_object,omitempty"`
	StoryDislike      int8            `json:"story_dislike,omitempty"`
	CustomizedQuality []int64         `json:"customized_quality,omitempty"`
	CustomizedDesc    *CustomizedDesc `json:"customized_desc,omitempty"`
	SingleInline      int8            `json:"-"`
	HideDuration      int8            `json:"hide_duration,omitempty"`
	SingleAdNew       int8            `json:"single_ad_new,omitempty"`
	cardStatus        struct {
		allowGIF                bool
		requestAt               time.Time
		ad                      *cm.AdInfo
		manualInline            int8
		gotoStoryDislikeReason  string
		adPkCode                string
		dynamicCover            int32
		singleInlineDbClickLike bool
		doubleInlineDbClickLike bool
		ogvNewCardHasScore      bool
		isPlaylist              bool
		allowGameBadge          bool
	}
	OgvNewStyle          int32              `json:"ogv_new_style,omitempty"`
	CornerMarkId         int8               `json:"corner_mark_id,omitempty"`
	CustomizedSubtitle   string             `json:"customized_subtitle,omitempty"`
	OgvRightInfo         int8               `json:"ogv_right_info,omitempty"`
	Epid                 int32              `json:"epid,omitempty"`
	OgvDislikeInfo       int64              `json:"ogv_dislike_info,omitempty"`
	NoPlay               int64              `json:"no_play,omitempty"`
	OgvCreativeId        int64              `json:"ogv_creative_id,omitempty"`
	IsStarlightLive      int64              `json:"is_starlight_live,omitempty"`
	StarlightOrderID     int64              `json:"starlight_order_id,omitempty"`
	SingleSpecialInfo    *SingleSpecialInfo `json:"single_special_info,omitempty"`
	OGVDescPriority      int8               `json:"ogv_desc_priority,omitempty"`
	AvDislikeInfo        int8               `json:"av_dislike_info,omitempty"`
	LiveCornerMark       int64              `json:"live_corner_mark,omitempty"`
	CreativeId           int64              `json:"creative_id,omitempty"`
	StoryCover           string             `json:"story_cover,omitempty"`
	StNewCover           int8               `json:"st_new_cover,omitempty"`
	LiveInlineDanmu      int8               `json:"live_inline_danmu,omitempty"`
	LiveInlineLight      int8               `json:"live_inline_light,omitempty"`
	LiveInlineLightDanmu int8               `json:"live_inline_light_danmu,omitempty"`
	TeenagerExempt       int8               `json:"teenager_exempt,omitempty"`
	RelationDislike      int8               `json:"relation_dislike,omitempty"`
	RelationDislikeText  string             `json:"relation_dislike_text,omitempty"`
}

type SingleSpecialInfo struct {
	SpID   int64  `json:"sp_id,omitempty"`
	SpType string `json:"sp_type,omitempty"`
}

type CustomizedDescItem struct {
	Type int64  `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

func (i *Item) TrackId() string {
	if i != nil {
		return i.TrackID
	}
	return ""
}

type CustomizedDesc struct {
	Desc     []*CustomizedDescItem `json:"desc,omitempty"`
	LinkType int64                 `json:"link_type,omitempty"`
}

type BigTunnelObject struct {
	Type     string `json:"type"`
	Resource string `json:"resource"`
}

func (i *Item) SingleInlineDbClickLike() bool {
	return i.cardStatus.singleInlineDbClickLike
}

func (i *Item) SetSingleInlineDbClickLike(in bool) {
	i.cardStatus.singleInlineDbClickLike = in
}

func (i *Item) DoubleInlineDbClickLike() bool {
	return i.cardStatus.doubleInlineDbClickLike
}

func (i *Item) SetDoubleInlineDbClickLike(in bool) {
	i.cardStatus.doubleInlineDbClickLike = in
}

func (i *Item) OgvHasScore() bool {
	return i.cardStatus.ogvNewCardHasScore
}

func (i *Item) SetOgvHasScore(in bool) {
	i.cardStatus.ogvNewCardHasScore = in
}

func (i *Item) IsPlaylist() bool {
	return i.cardStatus.isPlaylist
}

func (i *Item) SetIsPlaylist(in bool) {
	i.cardStatus.isPlaylist = in
}

func (i *Item) DynamicCoverInfoc() int32 {
	return i.cardStatus.dynamicCover
}

func (i *Item) SetDynamicCoverInfoc(in int32) {
	i.cardStatus.dynamicCover = in
}

func (i *Item) AdPKCode() string {
	return i.cardStatus.adPkCode
}

func (i *Item) SetAdPKCode(in string) {
	i.cardStatus.adPkCode = in
}

func (i *Item) CardStatusAd() *cm.AdInfo {
	return i.cardStatus.ad
}

func (i *Item) SetCardStatusAd(ad *cm.AdInfo) {
	i.cardStatus.ad = ad
}

func (i *Item) SetRequestAt(requestAt time.Time) {
	i.cardStatus.requestAt = requestAt
}

func (i *Item) RequestAt() time.Time {
	if i == nil {
		return time.Now()
	}
	if i.cardStatus.requestAt.IsZero() {
		return time.Now()
	}
	return i.cardStatus.requestAt
}

func (i *Item) SetAllowGIF() {
	i.cardStatus.allowGIF = true
}

func (i *Item) AllowGIF() bool {
	return i.cardStatus.allowGIF
}

func (i *Item) SetAllowGameBadge(in bool) {
	i.cardStatus.allowGameBadge = in
}

func (i *Item) AllowGameBadge() bool {
	return i.cardStatus.allowGameBadge
}

func (i *Item) SetManualInline(in int8) {
	i.cardStatus.manualInline = in
}

func (i *Item) ManualInline() int8 {
	return i.cardStatus.manualInline
}

func (i *Item) GotoStoryDislikeReason() string {
	return i.cardStatus.gotoStoryDislikeReason
}

func (i *Item) SetGotoStoryDislikeReason(in string) {
	i.cardStatus.gotoStoryDislikeReason = in
}

type BannerInfoItem struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	InlineType string `json:"inline_type"`
	InlineID   string `json:"inline_id"`
}

type BannerInfo struct {
	Items []*BannerInfoItem `json:"items"`
}

type Interest struct {
	CateID int64           `json:"cateid,omitempty"`
	Text   string          `json:"text,omitempty"`
	Items  []*InterestItem `json:"items,omitempty"`
}

type InterestItem struct {
	SubCateID int64  `json:"sub_cateid,omitempty"`
	Text      string `json:"text,omitempty"`
}

type Config struct {
	URI   string `json:"uri,omitempty"`
	Title string `json:"title,omitempty"`
	Cover string `json:"cover,omitempty"`
}

type RcmdReason struct {
	ID           int    `json:"id,omitempty"`
	Content      string `json:"content,omitempty"`
	BgColor      string `json:"bg_color,omitempty"`
	IconLocation string `json:"icon_location,omitempty"`
	Style        int    `json:"style,omitempty"`
	Font         int    `json:"font,omitempty"`
	Position     string `json:"position,omitempty"`
	Grounding    string `json:"grounding,omitempty"`
	CornerMark   int8   `json:"corner_mark,omitempty"`
	FollowedMid  int64  `json:"followed_mid,omitempty"`
	JumpID       int64  `json:"jumpid,omitempty"`
	JumpGoto     string `json:"jumpgoto,omitempty"`
}

type ConvergeInfo struct {
	Title        string `json:"title,omitempty"`
	Count        int    `json:"count,omitempty"`
	ConvergeType int32  `json:"converge_type,omitempty"`
	Quality      struct {
		Play  int64 `json:"play,omitempty"`
		Danmu int64 `json:"danmu,omitempty"`
	} `json:"quality,omitempty"`
	Items []*Item `json:"items,omitempty"`
}

type ConvergeInfoV2 struct {
	*ConvergeInfo
	UserFeature json.RawMessage `json:"user_feature"`
	Desc        string          `json:"desc,omitempty"`
}

type StoryInfo struct {
	Title string      `json:"title,omitempty"`
	Items []*SubItems `json:"items,omitempty"`
}

type SubItems struct {
	ID      int64  `json:"id,omitempty"`
	Goto    string `json:"goto,omitempty"`
	Cover   string `json:"cover,omitempty"`
	FfCover string `json:"ff_cover,omitempty"`
	Tid     int64  `json:"tid,omitempty"`
	// 详情页里面的字段
	TrackID        string `json:"trackid,omitempty"`
	Source         string `json:"source,omitempty"`
	AvFeature      string `json:"av_feature,omitempty"`
	StoryParam     string `json:"story_param,omitempty"`
	AdvertiseType  int32  `json:"advertise_type,omitempty"`
	StoryUpMid     int64  `json:"story_up_mid,omitempty"`
	EpID           int32  `json:"epid,omitempty"`
	HasIcon        int8   `json:"has_icon,omitempty"`
	IconType       string `json:"icon_type,omitempty"`
	IconID         int64  `json:"icon_id,omitempty"`
	IconTitle      string `json:"icon_title,omitempty"`
	PosRecUniqueId string `json:"dalao_uniq_id,omitempty"`
	PosRecTitle    string `json:"dalao_title,omitempty"`
	cardStatus     struct {
		disableRcmd      bool
		pgcAid           int64
		uid              int64
		liveAttentionExp int
		videoMode        int
		mobiApp          string
		build            int
	}
	HasTopic       int64  `json:"has_topic,omitempty"`
	TopicID        int64  `json:"topic_id,omitempty"`
	TopicTitle     string `json:"topic_title,omitempty"`
	OGVStyle       int8   `json:"ogv_style,omitempty"`
	HighlightStart int64  `json:"highlight_start,omitempty"` //单位: s
	DislikeStyle   int8   `json:"dislike_style,omitempty"`
	MutualReason   string `json:"mutual_reason,omitempty"`
	ExtraJson      string `json:"extra_json,omitempty"`
}

func (i *SubItems) DisableRcmd() bool {
	return i.cardStatus.disableRcmd
}

func (i *SubItems) SetDisableRcmd(in int) {
	i.cardStatus.disableRcmd = in == 1
}

func (i *SubItems) PgcAid() int64 {
	return i.cardStatus.pgcAid
}

func (i *SubItems) SetPgcAid(in int64) {
	i.cardStatus.pgcAid = in
}

func (i *SubItems) Uid() int64 {
	return i.cardStatus.uid
}

func (i *SubItems) SetUid(in int64) {
	i.cardStatus.uid = in
}

func (i *SubItems) LiveAttentionExp() int {
	return i.cardStatus.liveAttentionExp
}

func (i *SubItems) SetLiveAttentionExp(in int) {
	i.cardStatus.liveAttentionExp = in
}

func (i *SubItems) VideoMode() int {
	return i.cardStatus.videoMode
}

func (i *SubItems) SetVideoMode(in int) {
	i.cardStatus.videoMode = in
}

func (i *SubItems) MobiApp() string {
	return i.cardStatus.mobiApp
}

func (i *SubItems) SetMobiApp(in string) {
	i.cardStatus.mobiApp = in
}

func (i *SubItems) Build() int {
	return i.cardStatus.build
}

func (i *SubItems) SetBuild(in int) {
	i.cardStatus.build = in
}

func (i *SubItems) TrackId() string {
	if i != nil {
		return i.TrackID
	}
	return ""
}

type StoryView struct {
	Data        []*SubItems `json:"data,omitempty"`
	UserFeature string      `json:"user_feature,omitempty"`
	StoryBiz    *StoryBiz   `json:"story_biz,omitempty"`
}

type StoryBiz struct {
	Code int64         `json:"code"`
	Data *StoryBizData `json:"data"`
}

type StoryBizData struct {
	*cm.StoryAdResource
}

type BizData struct {
	// 用于展现日志上报返回的库存和广告数
	BizResult string `json:"biz_result,omitempty"`
	// 至多有两个选中的广告，如果没有ad_contents，那就是一个库存
	AdSelected []*cm.AdResource `json:"ad_selected,omitempty"`
	// 被抛弃掉的广告结构，用于上报Databus
	AdDiscarded []*cm.AdResource `json:"ad_discarded,omitempty"`
	// 未填充广告的库存卡片
	Stocks []*cm.AdResource `json:"stocks,omitempty"`
}
