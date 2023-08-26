package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

// CardGt is
type CardGt string

// CardType is
type CardType string

// ColumnStatus is
type ColumnStatus int8

// Gt is
type Gt string

// Icon is
type Icon int8

// Type is
type Type int8

// BlurStatus is
type BlurStatus int8

// Event is
type Event string

// CoverColor is
type CoverColor string

// Switch is
type Switch string

const (
	//player_info 修改层级版本限制
	PlayerIOSBuild      = 8400
	PlayerIOSBBuild     = 7370
	PlayerIPadHDBuild   = 12080
	PlayerAndroidBuild  = 5385000
	PlayerAndroidIBuild = 2020000

	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone = int8(1)
	// PlatIPad is int8 for ipad.
	PlatIPad = int8(2)
	// PlatWPhone is int8 for wphone.
	PlatWPhone = int8(3)
	// PlatAndroidG is int8 for Android Googleplay.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD = int8(20)
	// PlatAndroidB is int8 for android_b
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for iphone_b
	PlatIPhoneB = int8(10)
	// PlatAndroidTVYST is int8 for AndroidTV_YST Global.
	PlatAndroidTVYST = int8(12)
	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)

	CardGotoAv                     = CardGt("av")
	CardGotoBangumi                = CardGt("bangumi")
	CardGotoLive                   = CardGt("live")
	CardGotoArticle                = CardGt("article")
	CardGotoAudio                  = CardGt("audio")
	CardGotoRank                   = CardGt("rank")
	CardGotoConverge               = CardGt("converge")
	CardGotoConvergeAi             = CardGt("converge_ai")
	CardGotoDownload               = CardGt("download")
	CardGotoMoe                    = CardGt("moe")
	CardGotoBanner                 = CardGt("banner")
	CardGotoPGC                    = CardGt("pgc")
	CardGotoSpecial                = CardGt("special")
	CardGotoSpecialB               = CardGt("special_b")
	CardGotoSubscribe              = CardGt("subscribe")
	CardGotoBangumiRcmd            = CardGt("bangumi_rcmd")
	CardGotoUpRcmdAv               = CardGt("up_rcmd_av")
	CardGotoChannelRcmd            = CardGt("channel_rcmd")
	CardGotoLiveUpRcmd             = CardGt("live_up_rcmd")
	CardGotoAdAv                   = CardGt("ad_av")
	CardGotoAdPlayer               = CardGt("ad_player")
	CardGotoAdWeb                  = CardGt("ad_web")
	CardGotoAdWebS                 = CardGt("ad_web_s")
	CardGotoAdLive                 = CardGt("ad_live")
	CardGotoAdInlineAv             = CardGt("ad_inline_av")
	CardGotoPlayer                 = CardGt("player")
	CardGotoPlayerLive             = CardGt("player_live")
	CardGotoPlayerOGV              = CardGt("player_ogv")     //运营TAB的PGC播放卡
	CardGotoPlayerBangumi          = CardGt("player_bangumi") //天马首页的PGC播放卡
	CardGotoArticleS               = CardGt("article_s")
	CardGotoSpecialS               = CardGt("special_s")
	CardGotoShoppingS              = CardGt("shopping_s")
	CardGotoGameDownloadS          = CardGt("game_download_s")
	CardGotoGame                   = CardGt("game")
	CardGotoTopstick               = CardGt("topstick")
	CardGotoSearchSubscribe        = CardGt("search_subscribe")
	CardGotoPicture                = CardGt("picture")
	CardGotoInterest               = CardGt("interest")
	CardGotoFollowMode             = CardGt("follow_mode")
	CardGotoVipRenew               = CardGt("vip_renew")
	CardGotoAvConverge             = CardGt("av_converge")
	CardGotoIntroduction           = CardGt("introduction")
	CardGotoMultilayerConverge     = CardGt("multilayer_converge")
	CardGotoChannelNew             = CardGt("channel_new")
	CardGotoSpecialChannel         = CardGt("special_channel")
	CardGotoChannelNewDetail       = CardGt("channel_new_detail")
	CardGotoChannelNewDetailCustom = CardGt("channel_new_detail_custom")
	CardGotoTunnel                 = CardGt("tunnel")
	CardGotoNewTunnel              = CardGt("new_tunnel")
	CardGotoBigTunnel              = CardGt("big_tunnel")
	CardGotoChannelNewDetailRank   = CardGt("channel_new_detail_rank")
	CardGotoChannelScaned          = CardGt("channel_scaned")
	CardGotoChannelRcmdV2          = CardGt("channel_rcmd_v2")
	CardGotoChannelOGV             = CardGt("channel_ogv")
	CardGotoChannelOGVLarge        = CardGt("channel_ogv_large")
	CardGotoInlineAv               = CardGt("inline_av")
	CardGotoInlineAvV2             = CardGt("inline_av_v2")
	CardGotoReadCard               = CardGt("read_card")
	CardGotoAiStory                = CardGt("ai_story")
	CardGotoVerticalAv             = CardGt("vertical_av")
	CardGotoInlinePGC              = CardGt("inline_pgc")
	CardGotoInlineLive             = CardGt("inline_live")
	CardGotoAdInlineGesture        = CardGt("ad_inline_gesture")
	CardGotoAdInline360            = CardGt("ad_inline_360")
	CardGotoAdInlineLive           = CardGt("ad_inline_live")
	CardGotoAdWebGif               = CardGt("ad_web_gif")
	CardGotoAdInlineChoose         = CardGt("ad_inline_choose")
	CardGotoAdDynamic              = CardGt("ad_dynamic")
	CardGotoAdInlineChooseTeam     = CardGt("ad_inline_choose_team")
	CardGotoAdWebGifReservation    = CardGt("ad_web_gif_reservation")
	CardGotoAdPlayerReservation    = CardGt("ad_player_reservation")
	CardGotoAdInline3D             = CardGt("ad_inline_3d")
	CardGotoAdInline3DV2           = CardGt("ad_inline_3d_v2")
	CardGotoAdInlinePgc            = CardGt("ad_inline_ogv")
	CardGotoAdPgc                  = CardGt("ad_ogv")
	CardGotoAdInlineEggs           = CardGt("ad_inline_eggs")

	// operate tab
	CardGotoEntrance      = CardGt("entrance")
	CardGotoContentRcmd   = CardGt("content_rcmd")
	CardGotoTagRcmd       = CardGt("tag_rcmd")
	CardGotoNews          = CardGt("news")
	CardGotoChannelSquare = CardGt("channel_square")
	CardGotoPgcsRcmd      = CardGt("pgcs_rcmd")
	CardGotoUpRcmdS       = CardGt("up_rcmd_s")
	CardGotoSearchUpper   = CardGt("search_upper")
	CardGotoUpRcmdNew     = CardGt("up_rcmd_new")
	CardGotoUpRcmdSingle  = CardGt("up_rcmd_new_single")
	CardGotoUpRcmdNewV2   = CardGt("up_rcmd_new_v2")
	CardGotoEventTopic    = CardGt("event_topic")
	CardGotoVip           = CardGt("vip")
	CardGotoHotAV         = CardGt("hot_player_av")
	// single card
	LargeCoverV1              = CardType("large_cover_v1")
	LargeCoverV4              = CardType("large_cover_v4")
	PopularTopEntrance        = CardType("popular_top_entrance")
	SmallCoverV1              = CardType("small_cover_v1")
	MiddleCoverV1             = CardType("middle_cover_v1")
	ThreeItemV1               = CardType("three_item_v1")
	ThreeItemHV1              = CardType("three_item_h_v1")
	ThreeItemHV3              = CardType("three_item_h_v3")
	TwoItemV1                 = CardType("two_item_v1")
	BannerV1                  = CardType("banner_v1")
	BannerV4                  = CardType("banner_v4")
	CmV1                      = CardType("cm_v1")
	CmSingleV9                = CardType("cm_single_v9")
	CmSingleV7                = CardType("cm_single_v7")
	TopStick                  = CardType("top_stick")
	ChannelSquare             = CardType("channel_square")
	ThreeItemHV4              = CardType("three_item_h_v4")
	UpRcmdCover               = CardType("up_rcmd_cover")
	ThreeItemAll              = CardType("three_item_all")
	TwoItemHV1                = CardType("two_item_h_v1")
	OnePicV1                  = CardType("one_pic_v1")
	ThreePicV1                = CardType("three_pic_v1")
	SmallCoverV5              = CardType("small_cover_v5")
	SmallCoverV5Ad            = CardType("small_cover_v5_ad")
	SmallCoverH5              = CardType("small_cover_h5")
	SmallCoverH6              = CardType("small_cover_h6")
	SmallCoverH7              = CardType("small_cover_h7")
	OptionsV1                 = CardType("options_v1")
	RcmdOneItem               = CardType("rcmd_one_item")
	MiddleCoverV3             = CardType("middle_cover_v3")
	Select                    = CardType("select")
	SmallCoverV6              = CardType("small_cover_v6")
	SmallCoverV8              = CardType("small_cover_v8")
	Introduction              = CardType("introduction")
	SmallCoverConvergeV1      = CardType("small_cover_c_v1")
	ChannelNew                = CardType("channel_new")
	LargeCoverChannle         = CardType("large_cover_channel")
	ChannelThreeItemHV1       = CardType("channel_three_item_h_v1")
	ChannelThreeItemHV2       = CardType("channel_three_item_h_v2")
	ChannelScaned             = CardType("channel_scaned")
	ChannelRcmdV2             = CardType("channel_rcmd_v2")
	ChannelOGV                = CardType("channel_ogv")
	ChannelOGVLarge           = CardType("channel_ogv_large")
	StorysV1                  = CardType("storys_v1")
	NotifyTunnelSingleV1      = CardType("notify_tunnel_single_v1")
	NotifyTunnelLargeSingleV1 = CardType("notify_tunnel_large_single_v1")
	BannerSingleV8            = CardType("banner_single_v8")
	BannerV4169               = CardType("banner_v4_169")
	// double card
	SmallCoverV2         = CardType("small_cover_v2")
	SmallCoverV3         = CardType("small_cover_v3")
	MiddleCoverV2        = CardType("middle_cover_v2")
	LargeCoverV2         = CardType("large_cover_v2")
	LargeCoverV3         = CardType("large_cover_v3")
	ThreeItemHV2         = CardType("three_item_h_v2")
	ThreeItemV2          = CardType("three_item_v2")
	TwoItemV2            = CardType("two_item_v2")
	SmallCoverV4         = CardType("small_cover_v4")
	BannerV2             = CardType("banner_v2")
	BannerV5             = CardType("banner_v5")
	BannerV8             = CardType("banner_v8")
	BannerV5169          = CardType("banner_v5_169")
	CmV2                 = CardType("cm_v2")
	CmDoubleV9           = CardType("cm_double_v9")
	CmDoubleV7           = CardType("cm_double_v7")
	News                 = CardType("news")
	MultiItem            = CardType("multi_item")
	MultiItemH           = CardType("multi_item_h")
	ThreePicV2           = CardType("three_pic_v2")
	ThreePicV3           = CardType("three_pic_v3")
	OptionsV2            = CardType("options_v2")
	OnePicV2             = CardType("one_pic_v2")
	OnePicV3             = CardType("one_pic_v3")
	VipV1                = CardType("vip_v1")
	SmallCoverV7         = CardType("small_cover_v7")
	SmallCoverV9         = CardType("small_cover_v9")
	SmallCoverV10        = CardType("small_cover_v10")
	SmallCoverV11        = CardType("small_cover_v11")
	SmallCoverConvergeV2 = CardType("small_cover_c_v2")
	SmallCoverChannle    = CardType("small_cover_channel")
	LargeCoverV5         = CardType("large_cover_v5")
	StorysV2             = CardType("storys_v2")
	LargeCoverV6         = CardType("large_cover_v6")
	LargeCoverV7         = CardType("large_cover_v7")
	LargeCoverV8         = CardType("large_cover_v8")
	LargeCoverV9         = CardType("large_cover_v9")
	LargeCoverSingleV9   = CardType("large_cover_single_v9")
	NotifyTunnelV1       = CardType("notify_tunnel_v1")
	NotifyTunnelLargeV1  = CardType("notify_tunnel_large_v1")
	CmSingleV1           = CardType("cm_single_v1")
	OgvSmallCover        = CardType("ogv_small_cover")
	LargeCoverSingleV7   = CardType("large_cover_single_v7")
	LargeCoverSingleV8   = CardType("large_cover_single_v8")
	// double card new
	SmallCoverV2New         = CardType("small_cover_v2_new")
	SmallCoverV3New         = CardType("small_cover_v3_new")
	MiddleCoverV2New        = CardType("middle_cover_v2_new")
	LargeCoverV2New         = CardType("large_cover_v2_new")
	LargeCoverV3New         = CardType("large_cover_v3_new")
	ThreeItemHV2New         = CardType("three_item_h_v2_new")
	ThreeItemV2New          = CardType("three_item_v2_new")
	TwoItemV2New            = CardType("two_item_v2_new")
	SmallCoverV4New         = CardType("small_cover_v4_new")
	CoverOnlyV2New          = CardType("cover_only_v2_new")
	BannerV2New             = CardType("banner_v2_new")
	BannerV5New             = CardType("banner_v5_new")
	CmV2New                 = CardType("cm_v2_new")
	NewsNew                 = CardType("news_new")
	MultiItemNew            = CardType("multi_item_new")
	MultiItemHNew           = CardType("multi_item_h_new")
	ThreePicV3New           = CardType("three_pic_v3_new")
	OptionsV2New            = CardType("options_v2_new")
	OnePicV3New             = CardType("one_pic_v3_new")
	VipV1New                = CardType("vip_v1_new")
	SmallCoverV7New         = CardType("small_cover_v7_new")
	SmallCoverV9New         = CardType("small_cover_v9_new")
	SelectNew               = CardType("select_new")
	SmallCoverConvergeV2New = CardType("small_cover_c_v2_new")
	SmallCoverChannleNew    = CardType("small_cover_channel_new")
	ChannelSmallCoverV1     = CardType("channel_small_cover_v1")
	// ipad card
	BannerV3     = CardType("banner_v3")
	BannerV6     = CardType("banner_v6")
	BannerIPadV8 = CardType("banner_ipad_v8")

	// vertical card
	VerticalSmallCoverV2  = CardType("vertical_small_cover_v2")
	VerticalLargeCoverV7  = CardType("vertical_large_cover_v7")
	VerticalLargeCoverV9  = CardType("vertical_large_cover_v9")
	VerticalLargeCoverV11 = CardType("vertical_large_cover_v11")

	ColumnDefault    = ColumnStatus(0)
	ColumnSvrSingle  = ColumnStatus(1)
	ColumnSvrDouble  = ColumnStatus(2)
	ColumnUserSingle = ColumnStatus(3)
	ColumnUserDouble = ColumnStatus(4)

	GotoWeb                 = Gt("web")
	GotoAv                  = Gt("av")
	GotoAvAd                = Gt("av_ad")
	GotoBangumi             = Gt("bangumi")
	GotoLive                = Gt("live")
	GotoGame                = Gt("game")
	GotoArticle             = Gt("article")
	GotoArticleTag          = Gt("article_tag")
	GotoAudio               = Gt("audio")
	GotoAudioTag            = Gt("audio_tag")
	GotoSong                = Gt("song")
	GotoAlbum               = Gt("album")
	GotoClip                = Gt("clip")
	GotoDaily               = Gt("daily")
	GotoTag                 = Gt("tag")
	GotoMid                 = Gt("mid")
	GotoDynamicMid          = Gt("dynamic_mid")
	GotoConverge            = Gt("converge")
	GotoRank                = Gt("rank")
	GotoLiveTag             = Gt("live_tag")
	GotoPGC                 = Gt("pgc")
	GotoTopstick            = Gt("topstick")
	GotoSpecial             = Gt("special")
	GotoSubscribe           = Gt("subscribe")
	GotoPicture             = Gt("picture")
	GotoPictureTag          = Gt("picture_tag")
	GotoVip                 = Gt("vip")
	GotoPlaylist            = Gt("playlist")
	GotoAvConverge          = Gt("av_converge")
	GotoMultilayerConverge  = Gt("multilayer_converge")
	GotoChannel             = Gt("channel")
	GotoChannelDetailCustom = Gt("channel_detail_custom")
	GotoChannelDetailRank   = Gt("channel_detail_rank")
	GotoHotPage             = Gt("hot_page")
	GotoHotPlayerAv         = Gt("hot_player_av")
	GotoFeedLive            = Gt("feed_live")
	GotoVerticalAv          = Gt("vertical_av")
	GotoVerticalAvV2        = Gt("vertical_av_v2")
	GotoNavigation          = Gt("navigation")

	IconPlay           = Icon(1)
	IconOnline         = Icon(2)
	IconDanmaku        = Icon(3)
	IconFavorite       = Icon(4)
	IconStar           = Icon(5)
	IconRead           = Icon(6)
	IconComment        = Icon(7)
	IconLocation       = Icon(8)
	IconHeadphone      = Icon(9)
	IconRank           = Icon(10)
	IconGoldMedal      = Icon(11)
	IconSilverMedal    = Icon(12)
	IconBronzeMedal    = Icon(13)
	IconTV             = Icon(14)
	IconBomb           = Icon(15)
	IconRoleYellow     = Icon(16)
	IconRoleBlue       = Icon(17)
	IconRoleVipRed     = Icon(18)
	IconRoleYearVipRed = Icon(19)
	IconLike           = Icon(20)
	IconRoleBigYellow  = Icon(21)
	IconRoleBigBlue    = Icon(22)
	IconRoleCoin       = Icon(23)
	IconIsAttenm       = Icon(24)
	IconUp             = Icon(25)
	IconLiveWatched    = Icon(30)
	IconLiveOnline     = Icon(31)

	AvatarRound  = Type(0)
	AvatarSquare = Type(1)

	ButtonGrey  = Type(1)
	ButtonTheme = Type(2)

	BlurNo  = BlurStatus(0)
	BlurYes = BlurStatus(1)

	EventUpFollow         = Event("up_follow")
	EventChannelSubscribe = Event("channel_subscribe")
	EventUpClick          = Event("up_click")
	EventChannelClick     = Event("channel_click")
	EventButtonClick      = Event("button_click")
	EventGameClick        = Event("game_click")
	EventlikeClick        = Event("like_click")
	EventMainCard         = Event("main_card")
	EventReplyClick       = Event("reply_click")

	EventV2UpFollow         = Event("up-follow")
	EventV2ChannelSubscribe = Event("channel-subscribe")
	EventV2UpClick          = Event("up-click")
	EventV2ChannelClick     = Event("channel")
	EventV2ButtonClick      = Event("button")
	EventV2GameClick        = Event("button")
	EventV2likeClick        = Event("button")
	EventV2ReplyClick       = Event("button")

	PurpleCoverBadge = CoverColor("purple")

	BgColorOrange            = int8(0)
	BgColorTransparentOrange = int8(1)
	BgColorBlue              = int8(2)
	BgColorRed               = int8(3)
	BgTransparentTextOrange  = int8(4)
	BgColorPurple            = int8(5)
	BgColorTransparentRed    = int8(6)
	BgColorFillingOrange     = int8(7)
	BgColorYellow            = int8(8)
	BgColorLumpOrange        = int8(9)
	BgColorContourOrange     = int8(10)

	BgStyleFill              = int8(1)
	BgStyleStroke            = int8(2)
	BgStyleFillAndStroke     = int8(3)
	BgStyleNoFillAndNoStroke = int8(4)

	SwitchFeedIndexLike          = Switch("天马卡片好评数替换弹幕数")
	SwitchFeedIndexTabThreePoint = Switch("运营tab稿件卡片三点稍后再看")
	SwitchCooperationHide        = Switch("cooperation_hide")
	SwitchCooperationShow        = Switch("cooperation_show")
	SwitchLargeCoverHideAll      = Switch("largecover_hit_all")
	SwitchLargeCoverShowAll      = Switch("largecover_show_all")
	SwitchLargeCoverShowBottom   = Switch("largecover_show_bottom")
	SwitchFeedNewLive            = Switch("new_live")
	SwitchPictureLike            = Switch("picture_like")
	SwitchSpecialInfo            = Switch("special_info")
	SwitchNewReason              = Switch("new_reason")
	SwitchNewReasonV2            = Switch("new_reason_v2")
	SwitchPGCHideSubtitle        = Switch("pgc_hide_subtitle")

	// 热门显示up主信息abtest
	HotCardStyleOld    = int8(0)
	HotCardStyleShowUp = int8(1)
	HotCardStyleHideUp = int8(2)

	// 三点面板
	ThreePointWatchLater     = "watch_later"
	ThreePointDislike        = "dislike"
	ThreePointFeedback       = "feedback"
	ThreePointLike           = "like"
	ThreePointWhyContent     = "why_content"
	ThreePointSwitchToSingle = "switch_to_single"
	ThreePointSwitchToDouble = "switch_to_double"

	// 新旧频道类型
	OldChannel = 1
	NewChannel = 2

	ShortLinkHost = "https://b23.tv"

	//大会员铭牌过期图标
	VipLabelExpire  = "https://i0.hdslb.com/bfs/vip/label_overdue.png"
	VipStatusExpire = 0

	// inline 隐藏播放按钮
	HidePlayButton        = false
	TunnelHideDanmuSwitch = false
	TunnelDisableDanmu    = false
	BannerHideDanmuSwitch = false
	BannerDisableDanmu    = false

	SingleInlineV1 = 1 // 单列inline 1.0

	AIStoryIconType   = 1
	AIUpIconType      = 2
	AIUpStoryIconType = 3

	InlinePGCShareFrom       = "ogv_tianma_double_inline_normal_share"
	SingleInlinePGCShareFrom = "ogv_tianma_single_inline_normal_share"

	FlagConfirm = 1
	FlagCancel  = 2

	IconWatchLater     = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/NyPAqcn0QF.png"
	IconSwitchToSingle = "https://i0.hdslb.com/bfs/activity-plat/static/20210527/0977767b2e79d8ad0a36a731068a83d7/FwEKasbQt9.png"
	IconSwitchToDouble = "https://i0.hdslb.com/bfs/activity-plat/static/20210527/0977767b2e79d8ad0a36a731068a83d7/kaqQ3RjNwF.png"

	ReportHistory              = 1
	ReportRequiredPlayDuration = 10
	ReportRequiredTime         = 10

	// ogv新卡实验
	OgvCustomizedType = 1

	TunnelBadgeLive = 1

	FfCoverFromFeed         = "feed"
	FfCoverFromStory        = "story"
	FfCoverFromDynamicStory = "dynamic"
	FfCoverFromSpaceStory   = "space"
	//av inline progress icon
	InlineIconDrag     = "http://i0.hdslb.com/bfs/archive/c1461e2c6ca97783ac0298b6ebb2d85d94b8f37c.json"
	InlineIconDragHash = "31df8ce99de871afaa66a7a78f44deec"
	InlineIconStop     = "http://i0.hdslb.com/bfs/archive/6ee2f9b016f20714705cb5b8f15da1446587d172.json"
	InlineIconStopHash = "5648c2926c1c93eb2d30748994ba7b96"
)

var (
	OperateType = map[int]Gt{
		0:  GotoWeb,
		1:  GotoGame,
		2:  GotoAv,
		3:  GotoBangumi,
		4:  GotoLive,
		6:  GotoArticle,
		7:  GotoDaily,
		8:  GotoAudio,
		9:  GotoSong,
		10: GotoAlbum,
		11: GotoClip,
		12: GotoSpecial,
		13: GotoPicture,
		14: GotoPGC,
	}

	Columnm = map[ColumnStatus]ColumnStatus{
		ColumnDefault:    ColumnSvrDouble,
		ColumnSvrSingle:  ColumnSvrSingle,
		ColumnSvrDouble:  ColumnSvrDouble,
		ColumnUserSingle: ColumnSvrSingle,
		ColumnUserDouble: ColumnSvrDouble,
	}

	AvatarEvent = map[Gt]Event{
		GotoMid:        EventUpClick,
		GotoTag:        EventChannelClick,
		GotoDynamicMid: EventUpClick,
		GotoPGC:        EventUpClick,
	}

	AvatarEventV2 = map[Gt]Event{
		GotoMid:        EventV2UpClick,
		GotoTag:        EventV2ChannelClick,
		GotoDynamicMid: EventV2UpClick,
		GotoPGC:        EventV2UpClick,
	}

	ButtonEvent = map[Gt]Event{
		GotoMid: EventUpFollow,
		GotoTag: EventChannelSubscribe,
	}

	ButtonEventV2 = map[Gt]Event{
		GotoMid: EventV2UpFollow,
		GotoTag: EventV2ChannelSubscribe,
	}

	ButtonText = map[Gt]string{
		GotoMid: "+ 关注",
		GotoTag: "订阅",
	}

	ShareTo = map[string]bool{
		"weibo":         true,
		"wechat":        true,
		"wechatmonment": true,
		"qq":            true,
		"qzone":         true,
		"copy":          true,
		"more":          true,
		"dynamic":       true,
		"im":            true,
	}
	LiveRoomTagHandler = func(r *live.Room) func(uri string) string {
		return func(uri string) string {
			if r == nil {
				return ""
			}
			return fmt.Sprintf("%s?parent_area_id=%d&parent_area_name=%s&area_id=%d&area_name=%s", uri, r.AreaV2ParentID, url.QueryEscape(r.AreaV2ParentName), r.AreaV2ID, url.QueryEscape(r.AreaV2Name))
		}
	}
	LiveEntryRoomTagHandler = func(r *livexroomgate.EntryRoomInfoResp_EntryList) func(uri string) string {
		return func(uri string) string {
			if r == nil {
				return ""
			}
			return fmt.Sprintf("%s?parent_area_id=%d&parent_area_name=%s&area_id=%d&area_name=%s", uri, 0, url.QueryEscape(r.ParentAreaName), 0, url.QueryEscape(r.AreaName))
		}
	}
	AudioTagHandler = func(c []*audio.Ctg) func(uri string) string {
		return func(uri string) string {
			var schema string
			if len(c) != 0 {
				schema = c[0].Schema
				if len(c) > 1 {
					schema = c[1].Schema
				}
			}
			return schema
		}
	}
	LiveUpHandler = func(card *live.Card) func(uri string) string {
		return func(uri string) string {
			if card == nil {
				return uri
			}
			return fmt.Sprintf("%s?broadcast_type=%d", uri, card.BroadcastType)
		}
	}
	LiveRoomHandler = func(r *live.Room, network string) func(uri string) string {
		return func(uri string) string {
			if r == nil {
				return uri
			}
			params := url.Values{}
			params.Set("broadcast_type", strconv.Itoa(r.BroadcastType))
			if network != "" && r.PlayurlH264 != "" && len(r.AcceptQuality) > 0 && r.CurrentQuality != 0 && r.CurrentQn != 0 {
				if r.PlayurlH265 != "" {
					params.Set("playurl_h265", r.PlayurlH265)
				}
				params.Set("playurl_h264", r.PlayurlH264)
				params.Set("platform_network_status", network)
				acceptq, _ := json.Marshal(r.AcceptQuality)
				params.Set("accept_quality", string(acceptq))
				params.Set("current_quality", strconv.Itoa(r.CurrentQuality))
				params.Set("current_qn", strconv.Itoa(r.CurrentQn))
				qualityDesc, _ := json.Marshal(r.QualityDescription)
				params.Set("quality_description", string(qualityDesc))
				if r.ExtraParameter != "" {
					params.Set("extra_parameter", r.ExtraParameter)
				}
			}
			return fmt.Sprintf("%s?%s", uri, params.Encode())
		}
	}
	LiveEntryHandler = func(l *livexroomgate.EntryRoomInfoResp_EntryList, entryFrom string) func(uri string) string {
		return func(uri string) string {
			if l == nil {
				return uri
			}
			if entryFrom != "" {
				entryURI, ok := l.JumpUrl[entryFrom]
				if ok {
					return entryURI
				}
			}
			if l.LiveScreenType == 0 || l.LiveScreenType == 1 {
				return fmt.Sprintf("%s?broadcast_type=%d", uri, l.LiveScreenType)
			}
			return uri
		}
	}

	ArcPlayHandler = func(a *arcgrpc.Arc, ap *arcgrpc.PlayerInfo, trackID string, r *ai.Item, build int, mobiApp string, needPlayerInfo bool) func(uri string) string {
		buildOldPlayerFunc := func(p *arcgrpc.BvcVideoItem) (po *OldPlayerInfo) {
			if p == nil {
				return nil
			}
			po = &OldPlayerInfo{
				Cid:                p.Cid,
				ExpireTime:         p.ExpireTime,
				SupportQuality:     p.SupportQuality,
				SupportFormats:     p.SupportFormats,
				SupportDescription: p.SupportDescription,
				Quality:            int(p.Quality),
				URL:                p.Url,
				VideoCodecid:       p.VideoCodecid,
				VideoProject:       p.VideoProject,
				Fnver:              int(p.Fnver),
				Fnval:              int(p.Fnval),
				Dash:               p.Dash,
				NoRexcode:          p.NoRexcode,
			}
			po.FileInfo = make(map[int][]*OldPlayerFileInfo)
			for qn, f := range p.FileInfo {
				if f == nil {
					continue
				}
				for _, v := range f.Infos {
					if v == nil {
						continue
					}
					po.FileInfo[int(qn)] = append(po.FileInfo[int(qn)], &OldPlayerFileInfo{
						FileSize:   int64(v.Filesize),
						TimeLength: int64(v.Timelength),
					})
				}
			}
			return
		}
		var player string
		if ap != nil && ap.Playurl != nil && needPlayerInfo {
			var bs []byte
			if PlayerInfoNew(build, mobiApp) {
				bs, _ = json.Marshal(ap.Playurl)
			} else {
				bs, _ = json.Marshal(buildOldPlayerFunc(ap.Playurl))
			}
			player = string(bs)
		}
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if ap == nil || ap.Playurl == nil {
				params.Set("player_width", strconv.FormatInt(a.Dimension.Width, 10))
				params.Set("player_height", strconv.FormatInt(a.Dimension.Height, 10))
				params.Set("player_rotate", strconv.FormatInt(a.Dimension.Rotate, 10))
			} else {
				params.Set("cid", strconv.FormatInt(int64(ap.Playurl.Cid), 10))
				if player != "" {
					params.Set("player_preload", player)
				}
				if ap.PlayerExtra.GetDimension().GetHeight() != 0 || ap.PlayerExtra.GetDimension().GetWidth() != 0 {
					params.Set("player_width", strconv.FormatInt(ap.PlayerExtra.Dimension.GetWidth(), 10))
					params.Set("player_height", strconv.FormatInt(ap.PlayerExtra.Dimension.GetHeight(), 10))
					params.Set("player_rotate", strconv.FormatInt(ap.PlayerExtra.Dimension.GetRotate(), 10))
				}
				if ap.GetPlayerExtra().GetProgress() > 0 {
					params.Set("history_progress", strconv.FormatInt(ap.PlayerExtra.Progress, 10))
				}
			}
			if trackID != "" {
				params.Set("trackid", trackID)
			}
			if r != nil {
				if r.ConvergeParam != "" {
					params.Set("converge_param", r.ConvergeParam)
				}
				if r.StoryParam != "" {
					params.Set("story_param", r.StoryParam)
				}
				if r.Goto == "av" && r.JumpGoto == "" && r.TeenagerExempt != 0 {
					bizExtra := &BizExtra{
						TeenagerExempt: r.TeenagerExempt,
					}
					beByte, _ := json.Marshal(bizExtra)
					params.Set("biz_extra", string(beByte))
				}
			}
			// 拜年祭活动合集
			if a.AttrValV2(arcgrpc.AttrBitV2ActSeason) == arcgrpc.AttrYes && a.SeasonTheme != nil {
				params.Set("is_festival", "1")
				params.Set("bg_color", a.SeasonTheme.BgColor)
				params.Set("selected_bg_color", a.SeasonTheme.SelectedBgColor)
				params.Set("text_color", a.SeasonTheme.TextColor)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	ArticleTagHandler = func(c []*article.Category, plat int8) func(uri string) string {
		return func(uri string) string {
			var (
				rid int64
				tid int64
			)
			if len(c) > 1 {
				if c[0] != nil {
					rid = c[0].ID
				}
				if c[1] != nil {
					tid = c[1].ID
				}
			}
			if rid != 0 && tid != 0 {
				return fmt.Sprintf("bilibili://article/category/%d?sec_cid=%d", rid, tid)
			}
			return ""
		}
	}

	TrackIDHandler = func(trackID string, r *ai.Item, plat int8, build int) func(uri string) string {
		return func(uri string) string {
			var uriStr string
			if trackID != "" {
				uriStr = fmt.Sprintf("%s?trackid=%s", uri, trackID)
			}
			if r != nil {
				if r.ConvergeParam != "" {
					var convergeParam = r.ConvergeParam
					if plat == PlatIPhone && build > 8740 || plat == PlatAndroid && build > 5455000 {
						convergeParam = url.QueryEscape(convergeParam)
					}
					if uriStr == "" {
						uriStr = fmt.Sprintf("%s?converge_param=%s", uri, convergeParam)
					} else {
						uriStr = fmt.Sprintf("%s&converge_param=%s", uriStr, convergeParam)
					}
				}
				if r.ConvergeInfo != nil {
					if uriStr == "" {
						uriStr = fmt.Sprintf("%s?converge_type=%d", uri, r.ConvergeInfo.ConvergeType)
					} else {
						uriStr = fmt.Sprintf("%s&converge_type=%d", uriStr, r.ConvergeInfo.ConvergeType)
					}
				}
			}
			if uriStr != "" {
				return uriStr
			}
			return uri
		}
	}

	ChannelHandler = func(tab string) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?%s", uri, tab)
		}
	}

	BcupURIHandler = func() func(uri string) string {
		return func(uri string) string {
			if uri == "" {
				return uri
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("BcupURIHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("BcupURIHandler url.ParseQuery error(%v)", err)
				return uri
			}
			params.Set("auto_media_playback", "1")
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	PGCTrackIDHandler = func(r *ai.Item) func(uri string) string {
		return func(uri string) string {
			if uri == "" || r == nil || r.TrackID == "" {
				return uri
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("PGCTrackIDHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("PGCTrackIDHandler url.ParseQuery error(%v)", err)
				return uri
			}
			params.Set("goto", r.Goto)
			params.Set("trackid", r.TrackID)
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	URLTrackIDHandler = func(r ai.AiItem) func(uri string) string {
		return func(uri string) string {
			if uri == "" || r == nil {
				return uri
			}
			trackID := r.TrackId()
			if trackID == "" {
				return uri
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("URLTrackIDHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("URLTrackIDHandler url.ParseQuery error(%v)", err)
				return uri
			}
			params.Set("trackid", trackID)
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	URLLiveHandler = func(r *ai.Item, jumpFrom string) func(uri string) string {
		return func(uri string) string {
			if uri == "" || r == nil {
				return uri
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("URLLiveStarJumpFromHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("URLLiveStarJumpFromHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if r.TrackID != "" {
				params.Set("trackid", r.TrackID)
			}
			if jf := params.Get("jump_from"); r.IsStarlightLive > 0 && jf == "" {
				params.Set("jump_from", jumpFrom)
			}
			if from := params.Get("from"); r.IsStarlightLive > 0 && from == "" {
				params.Set("from", jumpFrom)
			}
			if lf := params.Get("live_from"); r.IsStarlightLive > 0 && lf == "" {
				params.Set("live_from", jumpFrom)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	GameHandler = func(r *ai.Item, sourceFrom string) func(uri string) string {
		return func(uri string) string {
			if uri == "" || r == nil {
				return uri
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("GameHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("GameHandler url.ParseQuery error(%v)", err)
				return uri
			}
			params.Set("sourceFrom", sourceFrom)
			if r.TrackID != "" {
				params.Set("trackid", r.TrackID)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
)

type OldPlayerInfo struct {
	Cid                uint32                       `json:"cid"`
	ExpireTime         uint64                       `json:"expire_time,omitempty"`
	FileInfo           map[int][]*OldPlayerFileInfo `json:"file_info"`
	SupportQuality     []uint32                     `json:"support_quality"`
	SupportFormats     []string                     `json:"support_formats"`
	SupportDescription []string                     `json:"support_description"`
	Quality            int                          `json:"quality"`
	URL                string                       `json:"url,omitempty"`
	VideoCodecid       uint32                       `json:"video_codecid"`
	VideoProject       bool                         `json:"video_project"`
	Fnver              int                          `json:"fnver"`
	Fnval              int                          `json:"fnval"`
	Dash               *arcgrpc.ResponseDash        `json:"dash,omitempty"`
	NoRexcode          int32                        `json:"no_rexcode,omitempty"`
}

type OldPlayerFileInfo struct {
	TimeLength int64  `json:"timelength"`
	FileSize   int64  `json:"filesize"`
	Ahead      string `json:"ahead,omitempty"`
	Vhead      string `json:"vhead,omitempty"`
	URL        string `json:"url,omitempty"`
	Order      int64  `json:"order,omitempty"`
}

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidI || plat == PlatAndroidB
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI || plat == PlatIPhoneB
}

// IsIPhone check plat is iphone.
func IsIPhone(plat int8) bool {
	return plat == PlatIPhone || plat == PlatIPhoneI
}

// IsIPad check plat is pad.
func IsIPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD
}

func IsAndroidPad(plat int8) bool {
	return plat == PlatAndroidHD
}

func IsPad(plat int8) bool {
	return IsIPad(plat) || IsAndroidPad(plat) || (plat == PlatWPhone)
}

// IsAndroidI check plat is android_i.
func IsAndroidI(plat int8) bool {
	return plat == PlatAndroidI
}

// IsIPhoneB check plat is iphone_b
func IsIPhoneB(plat int8) bool {
	return plat == PlatIPhoneB
}

// IsIPadHD check plat is iPadHD
func IsIPadHD(plat int8) bool {
	return plat == PlatIpadHD
}

// AdAvIsNormal check advert archive normal.
var AdAvIsNormalGRPC = AvIsNormalGRPC

func AvIsNormalGRPC(a *arcgrpc.ArcPlayer) bool {
	if a == nil || a.Arc == nil {
		return false
	}
	return a.Arc.IsNormal()
}

func AvIsCharging(a *arcgrpc.ArcPlayer) bool {
	if a == nil || a.Arc == nil {
		return false
	}
	return a.Arc.AttrValV2(arcgrpc.AttrBitV2Charging) == arcgrpc.AttrYes
}

// FillURI deal app schema.
//
//nolint:gocognit
func FillURI(gt Gt, plat int8, build int, param string, f func(uri string) string) (uri string) {
	switch gt {
	case GotoAv:
		if param != "" {
			uri = "bilibili://video/" + param
		}
	case GotoAvAd:
		if param != "" {
			uri = "bilibili://video/" + param
		}
	case GotoLive:
		if param != "" {
			if (plat == PlatAndroid && ((build > 5375000 && build < 5385000) || build > 5395000)) || (plat == PlatIPhone && build > 8430) {
				uri = "https://live.bilibili.com/"
			} else {
				uri = "bilibili://live/"
			}
			uri += param
		}
	case GotoFeedLive:
		if param != "" {
			uri = "https://live.bilibili.com/feed/" + param
		}
	case GotoBangumi:
		if param != "" {
			uri = "https://www.bilibili.com/bangumi/play/ep" + param
		}
	case GotoPGC:
		if param != "" {
			uri = "https://www.bilibili.com/bangumi/play/ss" + param
		}
	case GotoArticle:
		if param != "" {
			uri = "bilibili://article/" + param
		}
	case GotoArticleTag:
		// TODO fuck article
	case GotoGame:
		// TODO fuck game
		if param != "" {
			uri = "bilibili://game_center/detail?id=" + param + "&sourceType=adPut"
		}
	case GotoAudio:
		if param != "" {
			uri = "bilibili://music/menu/detail/" + param
		}
	case GotoSong:
		if param != "" {
			uri = "bilibili://music/detail/" + param
		}
	case GotoAudioTag:
		// uri = "bilibili://music/menus/menu?itemId=(请求所需参数)&cateId=(请求所需参数)&itemVal=(分类的标题value)"
	case GotoDaily:
		if param != "" {
			uri = "bilibili://pegasus/list/daily/" + param
		}
	case GotoAlbum:
		if param != "" {
			uri = "bilibili://album/" + param
		}
	case GotoClip:
		if param != "" {
			uri = "bilibili://clip/" + param
		}
	case GotoTag:
		if param != "" {
			uri = "bilibili://pegasus/channel/" + param
		}
	case GotoMid:
		if param != "" {
			uri = "bilibili://space/" + param
		}
	case GotoDynamicMid:
		if param != "" {
			uri = "bilibili://space/" + param + "?defaultTab=dynamic"
		}
	case GotoRank:
		uri = "bilibili://rank/"
	case GotoConverge:
		if param != "" {
			uri = "bilibili://pegasus/converge/" + param
		}
	case GotoAvConverge, GotoMultilayerConverge:
		if param != "" {
			uri = "bilibili://pegasus/ai/converge/" + param
		}
	case GotoLiveTag:
		uri = "https://live.bilibili.com/app/area"
	case GotoWeb:
		uri = param
	case GotoPicture:
		uri = "bilibili://following/detail/" + param
	case GotoPictureTag:
		uri = "bilibili://pegasus/channel/0/?name=" + param + "&type=topic"
	case GotoPlaylist:
		uri = "bilibili://music/playlist/playpage/" + param
	case GotoChannel:
		uri = "bilibili://pegasus/channel/v2/" + param
	case GotoHotPage:
		uri = "bilibili://pegasus/hotpage"
	case GotoVerticalAv:
		if param != "" {
			uri = "bilibili://story/" + param
		}
	default:
		uri = param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// DurationString duration to string
func DurationString(second int64) (s string) {
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

// StatString Stat to string
func StatString(number int32, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// StatString Stat to string
func Stat64String(number int64, suffix string) (s string) {
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

// ArchiveViewString ArchiveView to string
func ArchiveViewString(number int32) string {
	const _suffix = "观看"
	return StatString(number, _suffix)
}

// ArchiveViewString ArchiveView to string
func ArchiveView64String(number int64) string {
	const _suffix = "观看"
	return Stat64String(number, _suffix)
}

func ArchiveViewShareString(number int32) string {
	const (
		_suffix  = "已观看"
		_suffix2 = "万次"
	)
	if number <= 100000 {
		return ""
	}
	tmp := strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
	return _suffix + strings.TrimSuffix(tmp, ".0") + _suffix2
}

// DanmakuString Danmaku to string
func DanmakuString(number int32) string {
	const _suffix = "弹幕"
	return StatString(number, _suffix)
}

// DanmakuString Danmaku to string
func Danmaku64String(number int64) string {
	const _suffix = "弹幕"
	return Stat64String(number, _suffix)
}

// LikeString Danmaku to string
func LikeString(number int32) string {
	const _suffix = "点赞"
	return StatString(number, _suffix)
}

// BangumiFavString BangumiFav to string
func BangumiFavString(number, tp int32) string {
	var suffix = "追番"
	switch tp {
	case 2, 3, 5, 7:
		suffix = "追剧"
	}
	return StatString(number, suffix)
}

// LiveOnlineString online to string
func LiveOnlineString(number int32) string {
	const _suffix = "人气"
	return StatString(number, _suffix)
}

// FanString fan to string
func FanString(number int32) string {
	const _suffix = "粉丝"
	return StatString(number, _suffix)
}

// AttentionString fan to string
func AttentionString(number int32) string {
	const _suffix = "人关注"
	return StatString(number, _suffix)
}

// AudioDescString audio to string
func AudioDescString(firstSong string, total int) (desc1, desc2 string) {
	desc1 = firstSong
	if total == 1 {
		desc2 = "共1首歌曲"
		return
	}
	desc2 = "...共" + strconv.Itoa(total) + "首歌曲"
	return
}

// AudioTotalStirng audioTotal to string
func AudioTotalStirng(total int) string {
	if total == 0 {
		return ""
	}
	return strconv.Itoa(total) + "首歌曲"
}

// AudioBadgeString audioBadge to string
func AudioBadgeString(number int8) string {
	if number == 5 {
		return "专辑"
	}
	return "歌单"
}

// AudioPlayString audioPlay to string
func AudioPlayString(number int32) string {
	const _suffix = "收听"
	return StatString(number, _suffix)
}

// AudioFavString audioFav to string
func AudioFavString(numbber int32) string {
	const _suffix = "收藏"
	return StatString(numbber, _suffix)
}

// DownloadString download to string
func DownloadString(number int32) string {
	if number == 0 {
		return ""
	}
	const _suffix = "下载"
	return StatString(number, _suffix)
}

// ArticleViewString articleView to string
func ArticleViewString(number int64) string {
	const _suffix = "阅读"
	return StatString(int32(number), _suffix)
}

// PictureViewString pictureView to string
func PictureViewString(number int64) string {
	const _suffix = "浏览"
	return StatString(int32(number), _suffix)
}

// ArticleReplyString articleReply to string
func ArticleReplyString(number int64) string {
	const _suffix = "评论"
	return StatString(int32(number), _suffix)
}

// SubscribeString subscribe to string
func SubscribeString(number int32) string {
	const _suffix = "人已订阅"
	return StatString(number, _suffix)
}

// RecommendString recommend to string
func RecommendString(like, dislike int32) string {
	rcmd := like / (like + dislike) * 100
	if rcmd != 0 {
		return strconv.Itoa(int(rcmd)) + "%的人推荐"
	}
	return ""
}

// ShoppingDuration shopping duration
func ShoppingDuration(stime, etime string) string {
	if stime == "" && etime == "" {
		return ""
	}
	return stime + " - " + etime
}

// ScoreString is
func ScoreString(number int32) string {
	const _prefix = "综合评分："
	score := StatString(number, "")
	if score != "" {
		return _prefix + score
	}
	return _prefix + "-"
}

// ShoppingCover is
func ShoppingCover(cover string) string {
	if strings.HasPrefix(cover, "http:") || strings.HasPrefix(cover, "https:") {
		return cover
	}
	return "http:" + cover
}

// PubDataString is.
func PubDataString(t time.Time) (s string) {
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

// PubDataByRequestAt is.
func PubDataByRequestAt(pubData, requestAt time.Time) (s string) {
	if pubData.IsZero() {
		return
	}
	sub := requestAt.Sub(pubData)
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
	if requestAt.Year() == pubData.Year() {
		if requestAt.YearDay()-pubData.YearDay() == 1 {
			s = "昨天"
			return
		}
		s = pubData.Format("01-02")
		return
	}
	s = pubData.Format("2006-01-02")
	return
}

// PictureCountString is.
func PictureCountString(count int) string {
	return strconv.Itoa(count) + "P"
}

// OfficialIcon return 认证图标（1 UP 主认证，2 身份认证）黄标，（3 企业认证，4 政府认证，5 媒体认证，6 其他认证）蓝标
func OfficialIcon(cd *accountgrpc.Card) (icon Icon) {
	if cd == nil {
		return
	}
	switch cd.Official.Type {
	case 0:
		icon = IconRoleYellow
	case 1:
		icon = IconRoleBlue
	}
	return
}

// OfficialIconV2 return 认证图标（1 UP 主认证，2 身份认证）黄标，（3 企业认证，4 政府认证，5 媒体认证，6 其他认证）蓝标
func OfficialIconV2(cd *accountgrpc.Card) (icon Icon) {
	if cd == nil {
		return
	}
	switch cd.Official.Type {
	case 0:
		icon = IconRoleBigYellow
	case 1:
		icon = IconRoleBigBlue
	}
	return
}

func PlayerInfoNew(build int, mobiApp string) bool {
	switch mobiApp {
	case "iphone":
		return build > PlayerIOSBuild
	case "ipad":
		return build > PlayerIPadHDBuild
	case "android":
		return build > PlayerAndroidBuild
	case "iphone_b":
		return build > PlayerIOSBBuild
	case "android_i":
		return build > PlayerAndroidIBuild
	default:
		log.Warn("unknow player mobiApp(%s) build(%d)", mobiApp, build)
		return false
	}
}

func MarkRed(str string) (red string) {
	red = `<em class="keyword">` + str + `</em>`
	return
}

type EntranceItem struct {
	Goto         string  `json:"goto,omitempty"`
	Icon         string  `json:"icon,omitempty"`
	Title        string  `json:"title,omitempty"`
	ModuleId     string  `json:"module_id,omitempty"`
	Uri          string  `json:"uri,omitempty"`
	EntranceId   int64   `json:"entrance_id,omitempty"`
	Bubble       *Bubble `json:"bubble,omitempty"`
	EntranceType int32   `json:"entrance_type,omitempty"`
}

type Bubble struct {
	BubbleContent string `json:"bubble_content,omitempty"`
	Version       int32  `json:"version,omitempty"`
	Stime         int64  `json:"stime,omitempty"`
}

// 新接口转老关注关系
func RelationOldChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (isAttenm int8) {
	const (
		_follow = 1
	)
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	switch rel.Attribute {
	case 2, 6: // 用户关注UP主
		isAttenm = _follow
	}
	return
}

type Relation struct {
	Status     int32 `json:"status,omitempty"`
	IsFollow   int32 `json:"is_follow,omitempty"`
	IsFollowed int32 `json:"is_followed,omitempty"`
}

// 互相关注关系转换
func RelationChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (r *Relation) {
	const (
		// state使用
		_statenofollow      = 1
		_statefollow        = 2
		_statefollowed      = 3
		_statemutualConcern = 4
		// 关注关系
		_follow = 1
	)
	r = &Relation{
		Status: _statenofollow,
	}
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	switch rel.Attribute {
	case 2, 6: // 用户关注UP主
		r.Status = _statefollow
		r.IsFollow = _follow
	}
	if rel.IsFollowed { // UP主关注用户
		r.Status = _statefollowed
		r.IsFollowed = _follow
	}
	if r.IsFollow == _follow && r.IsFollowed == _follow { // 用户和UP主互相关注
		r.Status = _statemutualConcern
	}
	return
}

// ShowLive 是否展示直播入口 https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001302130
func ShowLive(mobiApp, device string, build int) bool {
	if mobiApp == "iphone_b" && build < 9360 || mobiApp == "android_b" && build < 5400000 {
		return false
	}
	return true
}

func ShowLiveV2(c context.Context, key string, req *feature.OriginResutl) bool {
	//nolint:gosimple
	if feature.GetBuildLimit(c, key, req) {
		return false
	}
	return true
}

type GotoIcon struct {
	IconURL      string `json:"icon_url,omitempty"`
	IconNightURL string `json:"icon_night_url,omitempty"`
	IconWidth    int64  `json:"icon_width,omitempty"`
	IconHeight   int64  `json:"icon_height,omitempty"`
}

func FillGotoIcon(iconType int, gotoIcon map[int64]*GotoIcon) *GotoIcon {
	icon, ok := gotoIcon[int64(iconType)]
	if !ok {
		return nil
	}
	return icon
}

func IsValidCover(cover string) bool {
	if cover == "" {
		return false
	}
	if !strings.HasPrefix(cover, "http://") && !strings.HasPrefix(cover, "https://") {
		log.Error("invalid cover: %q", cover)
		return false
	}
	_, err := url.Parse(cover)
	if err != nil {
		log.Error("Failed to parse cover: %s, %+v", cover, err)
		return false
	}
	return true
}

func ArcPlayURL(arc *arcgrpc.ArcPlayer, cid int64) *arcgrpc.PlayerInfo {
	if cid == 0 {
		cid = arc.DefaultPlayerCid
	}
	playerInfo, ok := arc.PlayerInfo[cid]
	if !ok {
		return nil
	}
	return playerInfo
}

type SharePlane struct {
	Title         string `json:"title,omitempty"`
	ShareSubtitle string `json:"share_subtitle,omitempty"`
	Desc          string `json:"desc,omitempty"`
	Cover         string `json:"cover,omitempty"`
	Aid           int64  `json:"aid,omitempty"`
	Bvid          string `json:"bvid,omitempty"`
	// 分享的渠道如："weibo": true
	ShareTo     map[string]bool `json:"share_to,omitempty"`
	Author      string          `json:"author,omitempty"`
	AuthorId    int64           `json:"author_id,omitempty"`
	ShortLink   string          `json:"short_link,omitempty"`
	PlayNumber  string          `json:"play_number,omitempty"`
	RoomId      int64           `json:"room_id,omitempty"`
	EpId        int32           `json:"ep_id,omitempty"`
	AreaName    string          `json:"area_name,omitempty"`
	AuthorFace  string          `json:"author_face,omitempty"`
	SeasonId    int32           `json:"season_id,omitempty"`
	ShareFrom   string          `json:"share_from,omitempty"`
	SeasonTitle string          `json:"season_title,omitempty"`
}

// https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661002259690
func CoverIconContentDescription(icon Icon, text string) string {
	if icon == 0 || text == "" {
		return ""
	}
	iconText := ""
	switch icon {
	case IconPlay:
		iconText = "观看"
	case IconOnline, IconLiveOnline:
		iconText = "人气值"
	case IconDanmaku:
		iconText = "弹幕"
	case IconFavorite:
		iconText = "追番"
	case IconRead:
		iconText = "阅读"
	case IconComment:
		iconText = "评论"
	case IconLike:
		return "点赞"
	case IconUp:
		return fmt.Sprintf("%s%s", "up主", text)
	case IconLiveWatched:
		return "看过"
	default:
		return ""
	}
	return fmt.Sprintf("%s%s", text, iconText)
}

// DurationContentDescription duration to talkBack
func DurationContentDescription(second int64) (s string) {
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
		s = fmt.Sprintf("%d分钟%02d秒", min, sec)
		return
	}
	s = fmt.Sprintf("%d小时%02d分钟%02d秒", hour, min, sec)
	return
}

func OgvCoverRightText(rcmd *ai.Item, epcard *pgccard.EpisodeCard, enablePgcScore bool) string {
	if rcmd.OgvHasScore() && epcard.Season.RatingInfo != nil && epcard.Season.RatingInfo.Score > 0 &&
		enablePgcScore {
		return fmt.Sprintf("%.1f分", epcard.Season.RatingInfo.Score)
	}
	return ""
}

func TalkBackCardType(goto_ Gt) string {
	switch goto_ {
	case GotoAv:
		return "视频"
	case GotoBangumi:
		return "番剧"
	case GotoLive:
		return "直播"
	case GotoArticle:
		return "专栏"
	case GotoGame:
		return "游戏"
	case GotoPGC:
		return "影视"
	case GotoVerticalAv, GotoVerticalAvV2:
		return "竖版视频"
	default:
		return ""
	}
}

type BizExtra struct {
	TeenagerExempt int8 `json:"teenager_exempt"`
}
