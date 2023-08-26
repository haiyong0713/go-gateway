package view

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	v12 "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"

	"go-common/library/log"
	"go-common/library/stat/prom"
	xtime "go-common/library/time"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	replyApi "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	musicApi "git.bilibili.co/bapis/bapis-go/crm/service/music-publicity-interface/toplist"
	mngApi "git.bilibili.co/bapis/bapis-go/manager/service/active"
	ogvgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	v1 "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	resApiV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/ad"
	"go-gateway/app/app-svr/app-view/interface/model/bangumi"
	"go-gateway/app/app-svr/app-view/interface/model/creative"
	"go-gateway/app/app-svr/app-view/interface/model/elec"
	"go-gateway/app/app-svr/app-view/interface/model/game"
	"go-gateway/app/app-svr/app-view/interface/model/live"
	"go-gateway/app/app-svr/app-view/interface/model/special"
	"go-gateway/app/app-svr/app-view/interface/model/tag"
	ahApi "go-gateway/app/app-svr/archive-honor/service/api"
	"go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	resApi "go-gateway/app/app-svr/resource/service/api/v1"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	steinsApi "go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

// View struct
type View struct {
	*ViewStatic
	// owner_ext
	OwnerExt OwnerExt `json:"owner_ext"`
	// now user
	ReqUser *viewApi.ReqUser `json:"req_user,omitempty"`
	// tag info
	Tag     []*tag.Tag                `json:"tag,omitempty"`
	DescTag []*tag.Tag                `json:"desc_tag,omitempty"`
	TIcon   map[string]*viewApi.TIcon `json:"t_icon,omitempty"`
	// movie
	Movie *bangumi.Movie `json:"movie,omitempty"`
	// bangumi
	Season *bangumi.Season `json:"season,omitempty"`
	// bp
	Bp json.RawMessage `json:"bp,omitempty"`
	// elec
	Elec     *elec.NewInfo     `json:"elec,omitempty"`
	ElecRank *viewApi.ElecRank `json:"elec_rank,omitempty"`
	// history
	History *viewApi.History `json:"history,omitempty"`
	// audio
	Audio *Audio `json:"audio,omitempty"`
	// contribute data
	Contributions []*Contribution `json:"contributions,omitempty"`
	// relate data
	Relates      []*Relate     `json:"relates,omitempty"`
	RelatesInfoc *RelatesInfoc `json:"-"`
	ReturnCode   string        `json:"-"`
	UserFeature  string        `json:"-"`
	IsRec        int8          `json:"-"`
	// dislike reason
	Dislikes  []*Dislike       `json:"dislike_reasons,omitempty"`
	DislikeV2 *viewApi.Dislike `json:"dislike_reasons_v2,omitempty"`
	// dm
	DMSeg int `json:"dm_seg,omitempty"`
	// paster
	Paster *resmdl.Paster `json:"paster,omitempty"`
	// player_icon
	PlayerIcon *resmdl.PlayerIcon `json:"player_icon,omitempty"`
	// vip_active
	VIPActive string `json:"vip_active,omitempty"`
	// cm
	CMs []*CM `json:"cms,omitempty"`
	// cm config
	CMConfig *CMConfig `json:"cm_config,omitempty"`
	// asset
	Asset       *Asset         `json:"asset,omitempty"`
	ActivityURL string         `json:"activity_url,omitempty"`
	Bgm         []*viewApi.Bgm `json:"bgm,omitempty"`
	Staff       []*Staff       `json:"staff,omitempty"`
	ArgueMsg    string         `json:"argue_msg,omitempty"`
	ShortLink   string         `json:"short_link,omitempty"`
	// AI experiments
	PlayParam int             `json:"play_param"` // 1=play automatically the relates, 0=not
	Label     *viewApi.Label  `json:"label,omitempty"`
	PvFeature json.RawMessage `json:"-"`
	// ugc season info
	UgcSeason *UgcSeason `json:"ugc_season,omitempty"` // ugc season info
	// config
	Config *Config `json:"config,omitempty"`
	// subtitle
	ShareSubtitle string                `json:"share_subtitle,omitempty"`
	Interaction   *viewApi.Interaction  `json:"interaction,omitempty"`
	Honor         *viewApi.Honor        `json:"honor,omitempty"`
	RelateTab     []*viewApi.RelateTab  `json:"relate_tab,omitempty"`
	TabInfo       []*TabInfo            `json:"-"` //相关推荐上报用
	CustomConfig  *viewApi.CustomConfig `json:"-"`
	BvID          string                `json:"bvid,omitempty"`
	ForbidRec     int64                 `json:"-"`
	UpAct         *viewApi.UpAct        `json:"up_act,omitempty"`
	// grpc接口用
	CMConfigNew       *viewApi.CMConfig          `json:"-"`
	CMSNew            []*viewApi.CM              `json:"-"`
	IPadCM            *viewApi.CmIpad            `json:"-"`
	ViewTab           *viewApi.Tab               `json:"-"`
	Rank              *viewApi.Rank              `json:"-"`
	TfPanelCustomized *viewApi.TFPanelCustomized `json:"-"`
	ZoneID            int64                      `json:"-"`
	BadgeUrl          string                     `json:"-"`
	ReplyStyle        *viewApi.ReplyStyle        `json:"-"`
	DescV2            []*viewApi.DescV2          `json:"-"`
	//用户装扮
	UserGarb *viewApi.UserGarb          `json:"-"`
	MngAct   *mngApi.CommonActivityResp `json:"-"`
	// 直播预约条
	LiveOrderInfo *viewApi.LiveOrderInfo `json:"live_order_info,omitempty"`
	//必剪-贴纸
	Sticker []*viewApi.ViewMaterial `json:"sticker,omitempty"`
	//点赞场景化定制配置
	LikeCustom *viewApi.LikeCustom `json:"like_custom,omitempty"`
	//投币定制
	CoinCustom *viewApi.CoinCustom `json:"-"`
	//一键三连动画
	UpLikeImg *viewApi.UpLikeImg `json:"-"`
	//新cell
	SpecialCell *viewApi.SpecialCell `json:"-"`
	//在看人数-简介内-左下角-特殊弹幕弹幕-三点里的开关面板4个地方是否展示
	Online *viewApi.Online `json:"-"`
	//商业框下条
	CmUnderPlayer *types.Any `json:"cm_under_player,omitempty"`
	//必剪-潮点视频
	VideoSource []*viewApi.ViewMaterial `json:"video_source,omitempty"`
	//首映资源
	PremiereResource *viewApi.PremiereResource `json:"premiere_resource"`
	//标签
	SpecialCellNew []*viewApi.SpecialCell `json:"-"`
	// 播放页展示标签 :是否要重新获取
	RefreshSpecialCell bool `json:"-"`
	//半屏幕icon
	MaterialLeft *viewApi.MaterialLeft `json:"-"`
	//笔记数量
	NotesCount int64 `json:"-"`
	//首映风控状态
	PremiereRiskStatus bool `json:"-"`
	// 是否是契约者
	IsContractor bool `json:"-"`
	// 是否是老粉
	IsOldFans bool `json:"-"`
	// 是否是硬核粉丝
	IsHardCoreFans bool `json:"-"`
	//客户端是否拉起浮层
	ClientAction *viewApi.PullClientAction `json:"client_action,omitempty"`
	//相关推荐分页参数
	Next string `json:"-"`
	//点赞icon
	LikeAnimation *viewApi.LikeAnimation `json:"-"`
	//是否是运营配置点赞动画
	IsLikeAnimation bool `json:"-"`
	//运营配置点赞动画
	OperationLikeAnimation string `json:"-"`
	//页面刷新
	RefreshPage *viewApi.RefreshPage `json:"-"`
}

// https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001420736
type RelatesInfoc struct {
	AdCode string
	AdNum  string
	PKCode string
}

func (ri *RelatesInfoc) SetAdCode(code string) {
	ri.AdCode = code
}

func (ri *RelatesInfoc) SetAdNum(num string) {
	if ri.AdNum == "" {
		ri.AdNum = num
		return
	}
	ri.AdNum += "," + num
}

func (ri *RelatesInfoc) SetPKCode(code string) {
	s := strings.Split(code, ",")
	if len(s) > 0 {
		ri.PKCode = s[0]
		prom.BusinessInfoCount.Incr(code)
	}
}

var PkCode = map[int]string{
	0:  "0,默认和未知",
	1:  "1,有运营-不可调整优先级",
	2:  "2,有运营-无广告返回",
	3:  "3,有运营-概率避让运营",
	4:  "4,有运营-概率优先-展现广告",
	5:  "5,有运营-概率优先-去重丢弃",
	6:  "6,无运营-避让商单",
	7:  "7,无运营-去重丢弃",
	8:  "8,无运营-广告展现",
	9:  "9,无运营-广告库存",
	10: "10,无运营-推荐&&其他",
}

const (
	AdFirstForRelate0  = "0,默认和未知"
	AdFirstForRelate1  = "1,有运营-不可调整优先级"
	AdFirstForRelate2  = "2,有运营-无广告返回"
	AdFirstForRelate3  = "3,有运营-概率避让运营"
	AdFirstForRelate4  = "4,有运营-概率优先-展现广告"
	AdFirstForRelate5  = "5,有运营-概率优先-去重丢弃"
	AdFirstForRelate6  = "6,无运营-避让商单"
	AdFirstForRelate7  = "7,无运营-去重丢弃"
	AdFirstForRelate8  = "8,无运营-广告展现"
	AdFirstForRelate9  = "9,无运营-广告库存"
	AdFirstForRelate10 = "10,无运营-推荐&&其他"
)

type Config struct {
	RelatesTitle       string                  `json:"relates_title,omitempty"`
	AutoplayCountdown  int                     `json:"autoplay_countdown,omitempty"`
	AutoplayDesc       string                  `json:"autoplay_desc,omitempty"`
	PageRefresh        int                     `json:"page_refresh,omitempty"`
	ShareStyle         int                     `json:"share_style,omitempty"`
	RelatesStyle       int                     `json:"relates_style,omitempty"`
	RelateGifExp       int                     `json:"relate_gif_exp,omitempty"`
	EndPageHalf        int                     `json:"end_page_half,omitempty"`
	EndPageFull        int                     `json:"end_page_full,omitempty"`
	AutoSwindow        bool                    `json:"auto_swindow,omitempty"`
	PopupInfo          bool                    `json:"popup_info,omitempty"`
	AbTestSmallWindow  string                  `json:"abtest_small_window,omitempty"`
	RecThreePointStyle int32                   `json:"rec_three_point_style"`
	IsAbsoluteTime     bool                    `json:"is_absolute_time"`
	NewSwindow         bool                    `json:"new_swindow,omitempty"`
	FeedStyle          string                  `json:"feed_style"`
	FeedPopUp          bool                    `json:"has_guide"`
	FeedHasNext        bool                    `json:"feed_has_next"`
	RelatesBiserial    bool                    `json:"relates_biserial,omitempty"`
	ListenerConfig     *viewApi.ListenerConfig `json:"-"`
	LocalPlay          int32                   `json:"local_play"`
	//命中实验竖屏稿件全屏默认进Story
	PlayStory bool `json:"-"`
	//当前视频竖屏稿件全屏进Story
	ArcPlayStory bool `json:"-"`
	//切story新icon样式
	StoryIcon string `json:"-"`
	//命中实验横屏稿件全屏默认进Story
	LandscapeStory bool `json:"-"`
	//当前视频横屏稿件全屏进Story
	ArcLandscapeStory bool `json:"-"`
	// 命中实验横屏稿件全屏默认进Story icon
	LandscapeIcon string `json:"-"`
	//听视频按钮
	ShowListenButton bool `json:"-"`
}

// Staff from cooperation
type Staff struct {
	Mid            int64  `json:"mid,omitempty"`
	Title          string `json:"title,omitempty"`
	Face           string `json:"face,omitempty"`
	Name           string `json:"name,omitempty"`
	OfficialVerify struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify"`
	Vip struct {
		Type          int             `json:"vipType"`
		DueDate       int64           `json:"vipDueDate"`
		DueRemark     string          `json:"dueRemark"`
		AccessStatus  int             `json:"accessStatus"`
		VipStatus     int             `json:"vipStatus"`
		VipStatusWarn string          `json:"vipStatusWarn"`
		ThemeType     int             `json:"themeType"`
		Label         accApi.VipLabel `json:"label"`
	} `json:"vip"`
	Attention  int   `json:"attention"`
	LabelStyle int32 `json:"label_style,omitempty"`
}

// ViewStatic struct
type ViewStatic struct {
	*api.Arc
	Pages []*Page `json:"pages,omitempty"`
}

// Page struct
type Page struct {
	*api.Page
	Metas      []*Meta            `json:"metas"`
	DMLink     string             `json:"dmlink"`
	Audio      *Audio             `json:"audio,omitempty"`
	DM         *dmApi.SubjectInfo `json:"dm,omitempty"`
	DlTitle    string             `json:"download_title,omitempty"`
	DlSubtitle string             `json:"download_subtitle,omitempty"`
}

// Meta struct
type Meta struct {
	Quality int    `json:"quality"`
	Format  string `json:"format"`
	Size    int64  `json:"size"`
}

// CM struct
type CM struct {
	RequestID string     `json:"request_id,omitempty"`
	RscID     int64      `json:"rsc_id,omitempty"`
	SrcID     int64      `json:"src_id,omitempty"`
	IsAdLoc   bool       `json:"is_ad_loc,omitempty"`
	IsAd      bool       `json:"is_ad,omitempty"`
	CmMark    int        `json:"cm_mark,omitempty"`
	ClientIP  string     `json:"client_ip,omitempty"`
	Index     int        `json:"index,omitempty"`
	AdInfo    *ad.AdInfo `json:"ad_info,omitempty"`
}

// CMConfig struct
type CMConfig struct {
	AdsControl  json.RawMessage `json:"ads_control,omitempty"`
	MonitorInfo json.RawMessage `json:"monitor_info,omitempty"`
}

// Dislike struct
type Dislike struct {
	ID   int    `json:"reason_id"`
	Name string `json:"reason_name"`
}

// OwnerExt struct
type OwnerExt struct {
	OfficialVerify struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify,omitempty"`
	Live *live.Live `json:"live,omitempty"`
	Vip  struct {
		Type          int             `json:"vipType"`
		DueDate       int64           `json:"vipDueDate"`
		DueRemark     string          `json:"dueRemark"`
		AccessStatus  int             `json:"accessStatus"`
		VipStatus     int             `json:"vipStatus"`
		VipStatusWarn string          `json:"vipStatusWarn"`
		ThemeType     int             `json:"themeType"`
		Label         accApi.VipLabel `json:"label"`
	} `json:"vip"`
	Assists  []int64 `json:"assists"`
	Fans     int     `json:"fans"`
	ArcCount string  `json:"arc_count"`
}

//
type ArcsPlayer struct {
	//aid
	Aid int64 `protobuf:"varint,1,opt,name=aid,proto3" json:"aid,omitempty"`
	//cid - 秒开地址
	Cid int64  `json:"cid,omitempty"`
	URI string `json:"uri,omitempty"`
}

// Relate struct
type Relate struct {
	Aid         int64             `json:"aid,omitempty"`
	Pic         string            `json:"pic,omitempty"`
	Title       string            `json:"title,omitempty"`
	Author      *api.Author       `json:"owner,omitempty"`
	Stat        api.Stat          `json:"stat,omitempty"`
	Duration    int64             `json:"duration,omitempty"`
	Goto        string            `json:"goto,omitempty"`
	Param       string            `json:"param,omitempty"`
	URI         string            `json:"uri,omitempty"`
	JumpURL     string            `json:"jump_url,omitempty"`
	Rating      float64           `json:"rating,omitempty"`
	Reserve     string            `json:"reserve,omitempty"`
	From        string            `json:"from,omitempty"`
	Desc        string            `json:"desc,omitempty"`
	RcmdReason  string            `json:"rcmd_reason,omitempty"`
	Badge       string            `json:"badge,omitempty"`
	Cid         int64             `json:"cid,omitempty"`
	SeasonType  int32             `json:"season_type,omitempty"`
	RatingCount int32             `json:"rating_count,omitempty"`
	Bage        string            `json:"bage,omitempty"`
	TagName     string            `json:"tag_name,omitempty"`
	PackInfo    *viewApi.PackInfo `json:"pack_info,omitempty"`
	Notice      *viewApi.Notice   `json:"notice,omitempty"`
	// cm ad
	AdIndex      int                  `json:"ad_index,omitempty"`
	CmMark       int                  `json:"cm_mark,omitempty"`
	SrcID        int64                `json:"src_id,omitempty"`
	RequestID    string               `json:"request_id,omitempty"`
	CreativeID   int64                `json:"creative_id,omitempty"`
	CreativeType int64                `json:"creative_type,omitempty"`
	Type         int                  `json:"type,omitempty"`
	Cover        string               `json:"cover,omitempty"`
	ButtonTitle  string               `json:"button_title,omitempty"`
	View         int                  `json:"view,omitempty"`
	Danmaku      int                  `json:"danmaku,omitempty"`
	IsAd         bool                 `json:"is_ad,omitempty"`
	IsAdLoc      bool                 `json:"is_ad_loc,omitempty"`
	AdCb         string               `json:"ad_cb,omitempty"`
	ShowURL      string               `json:"show_url,omitempty"`
	ClickURL     string               `json:"click_url,omitempty"`
	ClientIP     string               `json:"client_ip,omitempty"`
	Extra        json.RawMessage      `json:"extra,omitempty"`
	Button       *viewApi.Button      `json:"button,omitempty"`
	CardIndex    int                  `json:"card_index,omitempty"`
	Source       string               `json:"-"`
	AvFeature    json.RawMessage      `json:"-"`
	TrackID      string               `json:"trackid"`
	NewCard      int                  `json:"new_card,omitempty"`
	ReasonStyle  *viewApi.ReasonStyle `json:"rcmd_reason_style,omitempty"`
	CoverGif     string               `json:"cover_gif,omitempty"`
	CM           *viewApi.CM          `json:"-"`
	// game
	ReserveStatus     int64                  `json:"-"`
	ReserveStatusText string                 `json:"reserve_status_text,omitempty"`
	RcmdReasonExtra   string                 `json:"rcmd_reason_extra,omitempty"`
	RecThreePoint     *viewApi.RecThreePoint `json:"rec_three_point,omitempty"`
	//运营卡: 投放ID
	UniqueId string `json:"uniq_id,omitempty"`
	//运营卡: 物料ID
	MaterialId     int64         `json:"material_id,omitempty"`
	FromSourceType int64         `json:"from_source_type"`
	FromSourceId   string        `json:"from_source_id"`
	Dimension      api.Dimension `json:"dimension"`
	//粉标
	BadgeStyle *viewApi.ReasonStyle
	//强化角标
	PowerIconStyle *viewApi.PowerIconStyle
	//dislike上报
	DislikeReportData string `json:"dislike_report_data,omitempty"`
	//游戏榜单
	RankInfo   *game.RankInfo `json:"rank_info"`
	FirstFrame string         `json:"first_frame"`
}

// Contribution struct
type Contribution struct {
	Aid    int64      `json:"aid,omitempty"`
	Pic    string     `json:"pic,omitempty"`
	Title  string     `json:"title,omitempty"`
	Author api.Author `json:"owner,omitempty"`
	Stat   api.Stat   `json:"stat,omitempty"`
	CTime  xtime.Time `json:"ctime,omitempty"`
}

// Audio struct
type Audio struct {
	Title    string `json:"title"`
	Cover    string `json:"cover_url"`
	SongID   int    `json:"song_id"`
	Play     int    `json:"play_count"`
	Reply    int    `json:"reply_count"`
	UpperID  int    `json:"upper_id"`
	Entrance string `json:"entrance"`
	SongAttr int    `json:"song_attr"`
}

// VipPlayURL playurl token struct.
type VipPlayURL struct {
	From  string `json:"from"`
	Ts    int64  `json:"ts"`
	Aid   int64  `json:"aid"`
	Cid   int64  `json:"cid"`
	Mid   int64  `json:"mid"`
	VIP   int    `json:"vip"`
	SVIP  int    `json:"svip"`
	Owner int    `json:"owner"`
	Fcs   string `json:"fcs"`
}

// NewRelateRec struct
type NewRelateRec struct {
	TrackID           string          `json:"trackid"`
	Oid               int64           `json:"id"`
	Source            string          `json:"source"`
	AvFeature         json.RawMessage `json:"av_feature"`
	Goto              string          `json:"goto"`
	Title             string          `json:"title"`
	IsDalao           int8            `json:"is_dalao"`
	RcmdReason        *RcmdReason     `json:"rcmd_reason,omitempty"`
	CoverGif          string          `json:"cover_gif"`
	UniqueId          string          `json:"uniq_id"`
	MaterialId        int64           `json:"creative_id"`
	CustomizedTitle   string          `json:"customized_title"`
	CustomizedCover   string          `json:"customized_cover"`
	CustomizedOgvDesc string          `json:"customized_ogv_desc"`
	FromSourceType    int64           `json:"from_source_type"`
	FromSourceId      string          `json:"from_source_id"`
	Pos               int64           `json:"pos"`
	IsOgvEff          int64           `json:"is_ogv_eff"`
}

type RcmdReason struct {
	Content    string `json:"content,omitempty"`
	Style      int    `json:"style,omitempty"`
	CornerMark int8   `json:"corner_mark,omitempty"`
}

// Asset .
type Asset struct {
	Paid  int8  `json:"paid"`
	Price int64 `json:"price"`
	Msg   struct {
		Desc1 string `json:"desc1"`
		Desc2 string `json:"desc2"`
	} `json:"msg"`
	PreviewMsg struct {
		Desc1 string `json:"desc1"`
		Desc2 string `json:"desc2"`
	} `json:"preview_msg"`
}

// 相关推荐负反馈文案
const (
	//"反馈"文案
	_recFeedbackTitle    = "反馈"
	_recFeedbackSubTitle = "（选择后将优化此类推荐）"
	_recFeedbackText     = "将优化此类推荐"

	//"我不想看"文案
	_recDislikeTitle    = "我不想看"
	_recDislikeSubTitle = "（选择后将减少相似推荐）"
	_recDislikeText     = "将减少相似推荐"

	//"关闭个性推荐模式"文案
	_recClosedText             = "操作成功"
	_recFeedbackClosedSubTitle = "（选择后将优化此类推荐）"
	_recDislikeClosedSubTitle  = ""
	_recClosedToast            = "将在开启个性化推荐后生效"

	//"不感兴趣"文案
	_recNoInterestingTitle = "不感兴趣"
)

var disLikeReason = map[string]map[string][]*viewApi.DislikeReasons{
	"av": {
		"feedback": {
			&viewApi.DislikeReasons{Id: 1, Name: "恐怖血腥"},
			&viewApi.DislikeReasons{Id: 2, Name: "色情低俗"},
			&viewApi.DislikeReasons{Id: 3, Name: "封面恶心"},
			&viewApi.DislikeReasons{Id: 4, Name: "标题党/封面党"},
		},
		"dislike": {
			&viewApi.DislikeReasons{Id: 1, Name: "不感兴趣"},
		},
	},
	"bangumi": {
		"feedback": {
			&viewApi.DislikeReasons{Id: 3, Name: "封面恶心"},
			&viewApi.DislikeReasons{Id: 4, Name: "标题党/封面党"},
		},
		"dislike": {
			&viewApi.DislikeReasons{Id: 10, Name: "看过了"},
			&viewApi.DislikeReasons{Id: 1, Name: "不感兴趣"},
		},
	},
	"bangumi-ep": {
		"feedback": {
			&viewApi.DislikeReasons{Id: 3, Name: "封面恶心"},
			&viewApi.DislikeReasons{Id: 4, Name: "标题党/封面党"},
		},
		"dislike": {
			&viewApi.DislikeReasons{Id: 10, Name: "看过了"},
			&viewApi.DislikeReasons{Id: 1, Name: "不感兴趣"},
		},
	},
	"game": {
		"feedback": {},
		"dislike":  {},
	},
	"special": {
		"feedback": {},
		"dislike":  {},
	},
	"order": {
		"feedback": {},
		"dislike":  {},
	},
}

type InitTag struct {
	Config             *Config
	ActivityURL        string
	Tag                []*tag.Tag //tag为外层ai使用，不可随便删除
	TIcon              map[string]*viewApi.TIcon
	ViewTab            *viewApi.Tab
	SpecialCell        *viewApi.SpecialCell
	DescTag            []*tag.Tag
	SpecialCellNew     []*viewApi.SpecialCell
	RefreshSpecialCell bool
	MaterialLeft       *viewApi.MaterialLeft
	NotesCount         int64
	ClientAction       *viewApi.PullClientAction
}

// FromAv func
func (r *Relate) FromAv(a *api.Arc, from, trackid, coverGif string, ap *api.PlayerInfo, cooperation, ogvURL bool, build int, mobiApp string) {
	if a == nil {
		return
	}
	r.Aid = a.Aid
	r.Title = a.Title
	r.Pic = a.Pic
	r.Author = &a.Author
	r.Stat = a.Stat
	r.Duration = a.Duration
	r.Cid = a.FirstCid
	r.Goto = model.GotoAv
	r.FirstFrame = a.GetFirstFrame()
	r.Param = strconv.FormatInt(a.Aid, 10)
	needPlayInfo := true
	if ap == nil {
		needPlayInfo = false
	}
	r.URI = model.FillURI(r.Goto, r.Param, cdm.ArcPlayHandler(a, ap, trackid, nil, build, mobiApp, needPlayInfo))
	if ogvURL && a.RedirectURL != "" && a.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
		r.JumpURL = fillURIHandler(a.RedirectURL, r.Goto, trackid, from, 0)
	}
	r.From = from
	if a.AttrVal(api.AttrBitIsCooperation) == api.AttrYes && r.Author != nil && r.Author.Name != "" && cooperation {
		r.Author.Name = r.Author.Name + " 等联合创作"
	}
	r.CoverGif = coverGif
	r.Dimension = a.Dimension
}

// FromGame func
//
//nolint:gomnd
func (r *Relate) FromGame(c context.Context, featureCfg *conf.Feature, i *game.Info, from string, plat int8, build, gamecardStyleExp int) {
	if i.GameLink == "" {
		return
	}
	if plat == model.PlatIPhone && build > 8740 || plat == model.PlatAndroid && build > 5455000 {
		r.Title = i.GameName
	} else {
		r.Title = "相关游戏：" + i.GameName
	}
	r.Pic = i.GameIcon
	r.Rating = i.Grade
	r.ReserveStatus = int64(i.GameStatus)
	if i.GameStatus == 1 || i.GameStatus == 2 {
		var reserve string
		if i.BookNum < 10000 {
			reserve = strconv.FormatInt(i.BookNum, 10) + "人预约"
		} else {
			reserve = strconv.FormatFloat(float64(i.BookNum)/10000, 'f', 1, 64) + "万人预约"
		}
		r.Reserve = reserve
	}
	r.Goto = model.GotoGame
	r.URI = model.FillURI(r.Goto, i.GameLink, nil)
	r.Param = strconv.FormatInt(i.GameBaseID, 10)
	r.Button = &viewApi.Button{Title: "进入", Uri: r.URI}
	r.From = from
	r.Badge = "游戏"
	r.TagName = i.GameTags
	if i.GiftTitle != "" && i.GiftURL != "" {
		r.PackInfo = &viewApi.PackInfo{
			Title: i.GiftTitle,
			Uri:   i.GiftURL,
		}
	}
	if i.NoticeTitle != "" && i.Notice != "" {
		r.Notice = &viewApi.Notice{
			Title: i.NoticeTitle,
			Desc:  i.Notice,
		}
	}
	r.NewCard = gamecardStyleExp
}

// FromSpecial func
func (r *Relate) FromSpecial(sp *special.Card, from string, gifExp int, rec *NewRelateRec) {
	r.Title = sp.Title
	r.Pic = sp.Cover
	r.Goto = model.GotoSpecial
	// TODO FUCK game
	r.URI = model.FillURI(model.OperateType[sp.ReType], sp.ReValue, nil)
	if sp.Url != "" {
		r.URI = sp.Url
	}
	if r.URI != "" {
		if model.OperateType[sp.ReType] == model.GotoEP || model.OperateType[sp.ReType] == model.GotoBangumi {
			r.URI = fillURIHandler(r.URI, r.Goto, rec.TrackID, "", 0)
		}
	}
	r.Desc = sp.Desc
	r.Param = strconv.FormatInt(sp.ID, 10)
	r.RcmdReason = sp.Badge
	r.From = from
	if gifExp == 1 {
		r.CoverGif = sp.GifCover
	}
	//特殊小卡角标
	r.Badge = sp.Badge
	//特殊小卡"推荐原因"
	if rec.RcmdReason != nil {
		r.RcmdReasonExtra = rec.RcmdReason.Content
	}
	//粉标
	r.BadgeStyle = reasonStyle(model.BgColorTransparentRed, sp.Badge)
}

// FromOperate func
func (r *Relate) FromOperate(c context.Context, featureCfg *conf.Feature, i *NewRelateRec, a *api.Arc, info *game.Info,
	sp *special.Card, from, trackid string, cooperation, ogvURL bool, plat int8, build, gamecardStyleExp, gifExp int,
	ban *v1.CardInfoProto, bangumiAvId int64, ms map[int64]*resApiV2.Material) {
	switch i.Goto {
	case model.GotoAv:
		r.FromAv(a, from, trackid, "", nil, cooperation, ogvURL, 0, "")
	case model.GotoGame:
		r.FromGame(c, featureCfg, info, from, plat, build, gamecardStyleExp)
	case model.GotoSpecial:
		r.FromSpecial(sp, from, gifExp, i)
	case model.GotoBangumi:
		r.FromBangumi(ban, bangumiAvId, from, i)
	}

	if r.Title == "" {
		r.Title = i.Title
	}
	if i.RcmdReason != nil && i.RcmdReason.Content != "" {
		r.RcmdReason = i.RcmdReason.Content
	}
	if i.UniqueId != "" {
		r.UniqueId = i.UniqueId
	}
	if i.MaterialId > 0 && ms != nil {
		if _, ok := ms[i.MaterialId]; ok {
			if ms[i.MaterialId].Title != "" && ms[i.MaterialId].Cover != "" && ms[i.MaterialId].Desc != "" {
				r.MaterialId = i.MaterialId
				r.Title = ms[i.MaterialId].Title
				r.Pic = ms[i.MaterialId].Cover
				r.Desc = ms[i.MaterialId].Desc
			}
		}
	}
	r.FromSourceType = i.FromSourceType
	r.FromSourceId = i.FromSourceId
}

func (r *Relate) RecThreePointStyle(rec *NewRelateRec, authorName string, feedStyle string) {
	if rec == nil {
		return
	}
	r.RecThreePoint = NewRecThreePoint(rec)
	feedback := disLikeReason[rec.Goto]["feedback"]
	dislike := disLikeReason[rec.Goto]["dislike"]
	if rec.Goto == model.GotoAv && (feedStyle == "" || feedStyle == "default") {
		dislike = append([]*viewApi.DislikeReasons{
			{Id: 11, Name: "相关性低"},
		}, dislike...)
	}
	if rec.Goto == model.GotoAv && authorName != "" {
		dislike = append([]*viewApi.DislikeReasons{
			{Id: 4, Name: "UP主：" + authorName},
		}, dislike...)
	}
	r.RecThreePoint.Feedback.DislikeReason = feedback
	r.RecThreePoint.Dislike.DislikeReason = dislike
}

func NewRecThreePoint(rec *NewRelateRec) *viewApi.RecThreePoint {
	recThreePoint := &viewApi.RecThreePoint{}
	feedback := &viewApi.RecDislike{}
	disLike := &viewApi.RecDislike{}
	if rec.Goto == model.GotoAv && rec.Oid != 0 {
		recThreePoint.WatchLater = true
	}
	//ogv卡 + ugc卡
	if rec.Goto == model.GotoAv || rec.Goto == model.GotoBangumi || rec.Goto == model.GotoBangumiEp {
		//反馈
		feedback = &viewApi.RecDislike{
			Title:           _recFeedbackTitle,
			SubTitle:        _recFeedbackSubTitle,
			ClosedSubTitle:  _recFeedbackClosedSubTitle,
			PasteText:       _recFeedbackText,
			ClosedPasteText: _recClosedText,
			Toast:           _recClosedText,
			ClosedToast:     _recClosedText,
		}
		//我不想看
		disLike = &viewApi.RecDislike{
			Title:           _recDislikeTitle,
			SubTitle:        _recDislikeSubTitle,
			ClosedSubTitle:  _recDislikeClosedSubTitle,
			PasteText:       _recDislikeText,
			ClosedPasteText: _recClosedText,
			Toast:           _recClosedText,
			ClosedToast:     _recClosedToast,
		}
	}
	//游戏 + 特殊小卡 + 商单游戏卡
	if rec.Goto == model.GotoGame || rec.Goto == model.GotoSpecial || rec.Goto == model.GotoOrder {
		//反馈
		feedback = &viewApi.RecDislike{
			PasteText:       _recDislikeText,
			ClosedPasteText: _recClosedText,
			Toast:           _recClosedText,
			ClosedToast:     _recClosedText,
		}
		//我不想看
		disLike = &viewApi.RecDislike{
			Title:           _recNoInterestingTitle,
			PasteText:       _recDislikeText,
			ClosedPasteText: _recClosedText,
			Toast:           _recClosedText,
			ClosedToast:     _recClosedToast,
		}
	}
	recThreePoint.Dislike = disLike
	recThreePoint.Feedback = feedback
	return recThreePoint
}

// FromCM func
func (r *Relate) FromCM(ad *ad.AdInfo) {
	r.AdIndex = ad.Index
	r.CmMark = ad.CmMark
	r.SrcID = ad.Source
	r.RequestID = ad.RequestID
	r.CreativeID = ad.CreativeID
	r.CreativeType = ad.CreativeType
	r.Type = ad.CardType
	r.URI = ad.URI
	r.Param = ad.Param
	r.Goto = model.GotoCm
	r.View = ad.View
	r.Danmaku = ad.Danmaku
	r.Stat = ad.Stat
	r.Author = &ad.Author
	r.Duration = ad.Duration
	r.IsAd = ad.IsAd
	r.IsAdLoc = ad.IsAdLoc
	r.AdCb = ad.AdCb
	r.ClientIP = ad.ClientIP
	r.Extra = ad.Extra
	r.CardIndex = ad.CardIndex
	if ad.CreativeContent != nil {
		r.Aid = ad.CreativeContent.VideoID
		r.Cover = ad.CreativeContent.ImageURL
		r.Title = ad.CreativeContent.Title
		r.ButtonTitle = ad.CreativeContent.ButtonTitle
		r.Desc = ad.CreativeContent.Desc
		r.ShowURL = ad.CreativeContent.ShowURL
		r.ClickURL = ad.CreativeContent.ClickURL
	}
}

// FromCM func
func (c *CM) FromCM(ad *ad.AdInfo) {
	c.RequestID = ad.RequestID
	c.RscID = ad.Resource
	c.SrcID = ad.Source
	c.IsAd = ad.IsAd
	c.IsAdLoc = ad.IsAdLoc
	c.Index = ad.Index
	c.CmMark = ad.CmMark
	c.ClientIP = ad.ClientIP
	c.AdInfo = ad
}

// FromBangumi func
func (r *Relate) FromBangumi(ban *v1.CardInfoProto, aid int64, from string, rec *NewRelateRec) {
	r.Title = ban.Title
	r.Pic = ban.NewEp.Cover
	//如果ai下发title（标题）、cover（封面）则优先使用ai的结果
	if rec.CustomizedTitle != "" {
		r.Title = rec.CustomizedTitle
	}
	if rec.CustomizedCover != "" {
		r.Pic = rec.CustomizedCover
	}
	r.Stat = api.Stat{
		Danmaku: int32(ban.Stat.Danmaku),
		View:    int32(ban.Stat.View),
		Fav:     int32(ban.Stat.Follow),
	}
	r.Goto = model.GotoBangumi
	r.Param = strconv.FormatInt(int64(ban.SeasonId), 10)
	r.URI = model.FillURI(r.Goto, r.Param, nil)
	if aid != 0 && r.URI != "" {
		r.URI = fillURIHandler(r.URI, r.Goto, r.TrackID, from, aid)
	}
	r.From = from
	r.SeasonType = ban.SeasonType
	r.Badge = ban.SeasonTypeName
	r.Desc = ban.NewEp.IndexShow
	if ban.Rating != nil {
		r.Rating = float64(ban.Rating.Score)
		r.RatingCount = ban.Rating.Count
	}
	//粉标
	r.BadgeStyle = reasonStyle(model.BgColorTransparentRed, ban.SeasonTypeName)
}

func (r *Relate) FromBangumiEpOperate(banEp *ogvgrpc.EpisodeCard, aid int64, from string, rec *NewRelateRec, ms map[int64]*resApiV2.Material,
	ogvMaterial map[int64]*v12.EpMaterial) {
	r.FromBangumiEp(banEp, aid, from, rec, ms, ogvMaterial)
	//如果是后台配置，替换uri
	r.URI = model.FillURI(r.Goto, r.Param, nil)
	if aid != 0 && r.URI != "" {
		r.URI = fillURIHandler(r.URI, r.Goto, r.TrackID, from, aid)
	}
	if rec.UniqueId != "" {
		r.UniqueId = rec.UniqueId
	}
	r.MaterialId = rec.MaterialId
	r.FromSourceId = rec.FromSourceId
	r.FromSourceType = rec.FromSourceType
}

func (r *Relate) FromBangumiEp(ep *ogvgrpc.EpisodeCard, aid int64, from string, rec *NewRelateRec, ms map[int64]*resApiV2.Material, ogvMaterial map[int64]*v12.EpMaterial) {
	r.Title = ep.GetUgcRelatedRcmdCardMeta().GetTitle()
	r.Desc = ep.GetUgcRelatedRcmdCardMeta().GetSubTitle()
	r.Pic = ep.Cover
	//source='pgc'、goto='bangumi-ep' 会有customized字段
	if _, ok := ms[rec.MaterialId]; ok { //非算法卡则拿运营后台的数据
		if ms[rec.MaterialId].GetCover() != "" {
			r.Pic = ms[rec.MaterialId].Cover
		}
		if ms[rec.MaterialId].GetTitle() != "" {
			r.Title = ms[rec.MaterialId].Title
			r.Desc = ms[rec.MaterialId].Desc
		}
		//如果ai下发OgvDesc（推荐原因）则优先使用ai的结果
		if rec.CustomizedOgvDesc != "" {
			r.Desc = rec.CustomizedOgvDesc
		}
	}
	if rec.Source == "pgc" { //算法卡则拿ogvMaterial的数据
		if _, ok := ogvMaterial[rec.MaterialId]; ok {
			if ogvMaterial[rec.MaterialId].Cover != "" {
				r.Pic = ogvMaterial[rec.MaterialId].Cover
			}
			if ogvMaterial[rec.MaterialId].Title != "" {
				r.Title = ogvMaterial[rec.MaterialId].Title
				if rec.CustomizedOgvDesc != "" {
					r.Desc = rec.CustomizedOgvDesc
				}
			}
		}
	}
	if rec.Source == "da" && rec.IsOgvEff == 1 { //效率池
		if _, ok := ogvMaterial[rec.MaterialId]; ok {
			if ogvMaterial[rec.MaterialId].GetCover() != "" {
				r.Pic = ogvMaterial[rec.MaterialId].Cover
			}
			if ogvMaterial[rec.MaterialId].GetTitle() != "" {
				r.Title = ogvMaterial[rec.MaterialId].Title
				r.Desc = ogvMaterial[rec.MaterialId].Desc
				if rec.CustomizedOgvDesc != "" {
					r.Desc = rec.CustomizedOgvDesc
				}
			}
		}
	}
	r.Stat = api.Stat{
		Danmaku: int32(ep.GetSeason().GetStat().GetDanmaku()),
		View:    int32(ep.GetSeason().GetStat().GetView()),
		Fav:     int32(ep.GetSeason().GetStat().GetFollow()),
	}
	r.MaterialId = rec.MaterialId
	r.Goto = model.GotoBangumiEp
	r.Param = strconv.FormatInt(int64(ep.EpisodeId), 10)
	r.URI = model.FillURI(r.Goto, r.Param, nil)
	if aid != 0 && r.URI != "" {
		r.URI = fillURIHandler(r.URI, r.Goto, r.TrackID, from, aid)
	}
	r.From = from
	r.SeasonType = ep.GetSeason().GetSeasonType()
	r.Badge = ep.GetSeason().GetSeasonTypeName()
	r.Rating = float64(ep.GetSeason().GetRatingInfo().GetScore())
	r.RatingCount = ep.GetSeason().GetRatingInfo().GetCount()
	//粉标
	r.BadgeStyle = reasonStyle(model.BgColorTransparentRed, ep.GetSeason().GetSeasonTypeName())
}

// fillURIHandler func
func fillURIHandler(uri, gt, trackId, from string, aid int64) string {
	if gt == "" {
		return uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		log.Error("fillURIHandler url.Parse error(%v)", err)
		return uri
	}
	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Error("fillURIHandler url.ParseQuery error(%v)", err)
		return uri
	}
	params.Set("goto", gt)
	switch gt {
	case model.GotoAv:
		if from == model.FromOperation {
			params.Set("card_type", from)
		} else {
			params.Set("card_type", "ai")
		}
		if trackId != "" {
			params.Set("trackid", trackId)
		}
	case model.GotoBangumi:
		params.Set("from_av", strconv.FormatInt(aid, 10))
		if trackId != "" {
			params.Set("trackid", trackId)
		}
		if from == model.FromOperation {
			params.Set("card_type", from)
		} else {
			params.Set("card_type", "ai")
		}
	case model.GotoBangumiEp:
		params.Set("from_av", strconv.FormatInt(aid, 10))
		if trackId != "" {
			params.Set("trackid", trackId)
		}
		if from == model.FromOperation {
			params.Set("card_type", from)
		} else {
			params.Set("card_type", "ai")
		}
	case model.GotoSpecial:
		if trackId != "" {
			params.Set("trackid", trackId)
		}
	default:
		log.Error("unrecognized goto(%s)", gt)
	}
	paramStr := params.Encode()
	if strings.IndexByte(paramStr, '+') > -1 {
		paramStr = strings.Replace(paramStr, "+", "%20", -1)
	}
	u.RawQuery = paramStr
	return u.String()
}

// TripleParam struct
type TripleParam struct {
	MobiApp   string `form:"mobi_app"`
	Build     string `form:"build"`
	AID       int64  `form:"aid" validate:"min=1"`
	Ak        string `form:"access_key"`
	From      string `form:"from"`
	Device    string `form:"device"`
	Platform  string `form:"platform"`
	Appkey    string `form:"appkey"`
	Spmid     string `form:"spmid"`
	FromSpmid string `form:"from_spmid"`
	TrackID   string `form:"track_id"`
	Goto      string `form:"goto"`
}

// TripleRes struct
type TripleRes struct {
	Like     bool  `json:"like"`
	Coin     bool  `json:"coin"`
	Fav      bool  `json:"fav"`
	Prompt   bool  `json:"prompt"`
	Multiply int64 `json:"multiply"`
	UpID     int64 `json:"-"`
}

// ShareIcon .
type ShareIcon struct {
	ShareChannel string `json:"share_channel"`
}

// Videoshot videoshot
type Videoshot struct {
	*api.VideoShot
	Points []*creative.Points `json:"points,omitempty"`
}

// DislikeReasons .
func (v *View) DislikeReasons(c context.Context, featureCfg *conf.Feature, mobiApp, device string, build int, disableRcmd int) {
	const (
		_noSeason = 1
		_region   = 2
		_channel  = 3
		_upper    = 4
		_tagMAX   = 2
	)
	var (
		taginfo *tag.Tag
	)

	dislikeText := "选择不想看的原因，减少相似内容推荐"
	//关闭个性化推荐时，展示文案
	if disableRcmd == 1 {
		dislikeText = "选择不想看的原因"
	}

	v.DislikeV2 = &viewApi.Dislike{
		Title: dislikeText,
		// Subtitle: "(选择后将减少相似内容推荐)",
	}
	if v.Author.Name != "" {
		v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, &viewApi.DislikeReasons{Id: _upper, Name: "UP主:" + v.Author.Name, Mid: v.Author.Mid})
	}
	// if v.TypeName != "" {
	// 	v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, &DislikeReasons{ID: _region, Name: "分区:" + v.TypeName, RID: v.TypeID})
	// }
	if len(v.Tag) > 0 {
		for i, t := range v.Tag {
			v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, &viewApi.DislikeReasons{Id: _channel, Name: "频道:" + t.Name, TagId: t.TagID})
			if i == 0 {
				taginfo = t
			}
			if i+1 >= _tagMAX {
				break
			}
		}
	}
	var dislike *viewApi.DislikeReasons
	if feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.Dislike, &feature.OriginResutl{
		MobiApp:    mobiApp,
		Device:     device,
		Build:      int64(build),
		BuildLimit: (mobiApp == "iphone" && build > 8700) || (mobiApp == "android" && build > 5445000),
	}) {
		dislike = &viewApi.DislikeReasons{Id: _noSeason, Name: "我不想看这个内容", Mid: v.Author.Mid, Rid: v.TypeID}
	} else {
		dislike = &viewApi.DislikeReasons{Id: _noSeason, Name: "不感兴趣", Mid: v.Author.Mid, Rid: v.TypeID}
		subtitle := "(选择后将减少相似内容推荐)"
		//关闭个性化推荐时，展示文案
		if disableRcmd == 1 {
			subtitle = ""
		}
		v.DislikeV2.Subtitle = subtitle
	}
	if taginfo != nil {
		dislike.TagId = taginfo.TagID
	}
	v.DislikeV2.Reasons = append(v.DislikeV2.Reasons, dislike)
	//descTag字段优先级高于v.Tag
	//灰度过程v.Tag和v.DescTag会共存，最终标签都会收进简介里descTag：https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661002448049
	if (len(v.DescTag) > 0 || v.SpecialCell != nil || len(v.SpecialCellNew) > 0) && len(v.Tag) > 0 {
		v.Tag = nil //清空v.Tag
	}
}

// RelateRes is
type RelateRes struct {
	Code              int             `json:"code"`
	Data              []*NewRelateRec `json:"data"`
	UserFeature       string          `json:"user_feature"`
	PlayParam         int             `json:"play_param"`
	PvFeature         json.RawMessage `json:"pv_feature"`
	AutoplayCountdown int             `json:"autoplay_countdown"`
	ReturnPage        int             `json:"return_page_exp"`
	AutoplayToast     string          `json:"autoplay_toast"`
	GamecardStyleExp  int             `json:"gamecard_style_exp"`
	RecReasonExp      int             `json:"rec_reason_exp"`
	TabInfo           []*TabInfo      `json:"tabinfo"`
	GifExp            int             `json:"gif_exp"` //gif封面实验 1=命中
	FeedStyle         string          `json:"feed_style"`
	FeedPopUp         int             `json:"has_guide"`
}

type RelateResV2 struct {
	BizData          *BizData        `json:"biz_data"`
	Code             int             `json:"code"`
	BizPkCode        int             `json:"biz_pk_code"`
	DalaoExp         int             `json:"dalao_exp"`
	Data             []*NewRelateRec `json:"data"`
	DislikeExp       int             `json:"dislike_exp"`
	UserFeature      string          `json:"user_feature"`
	PlayParam        int             `json:"play_param"`
	PvFeature        json.RawMessage `json:"pv_feature"`
	GamecardStyleExp int             `json:"gamecard_style_exp"` //是否在详情页展示游戏卡片的新样式, 已推全，固定为1
	GifExp           int             `json:"gif_exp"`            //gif封面实验 1=命中
	RecReasonExp     int             `json:"rec_reason_exp"`
	ArcCommercial    *Commercial     `json:"arc_commercial"`
	BizAdvNum        string          `json:"biz_adv_num"`
	FeedStyle        string          `json:"feed_style"`
	FeedPopUp        int             `json:"has_guide"`
	LocalPlay        int             `json:"local_play"`   //local_play=1，代表命中原地播放实验，即用户点击相关推荐列表后不更新当前feed流
	Next             string          `json:"next"`         //分页结束标识
	Refreshable      int             `json:"refreshable"`  //是否支持顶部刷新
	RefreshIcon      int             `json:"refresh_icon"` //悬浮按钮icon类型
	RefreshText      string          `json:"refresh_text"` //icon 文本
	RefreshShow      float32         `json:"refresh_show"` //悬浮按钮出现时机
}

type BizData struct {
	Code int     `json:"code"`
	Data *AdData `json:"data"`
}

type AdData struct {
	RequestID  string                         `json:"request_id,omitempty"`
	AdsControl string                         `json:"ads_control,omitempty"`
	AdsInfo    map[int32]map[int32]*AdsInfoV2 `json:"ads_info,omitempty"`
	TabInfo    *AdTabInfo                     `json:"tab_info,omitempty"`
}

type AdTabInfo struct {
	TabName string `json:"tab_name"`
	Extra   string `json:"extra"`
	//商业tab新样式改版, 1=旧样式, 2=新样式
	TabVersion int32 `json:"tab_version"`
}

type AdsInfoV2 struct {
	AvId           int64  `json:"av_id"`
	IsAd           bool   `json:"is_ad"`
	CardIndex      int32  `json:"card_index"`
	CardType       int32  `json:"card_type"`
	SourceContents string `json:"source_contents"`
}

type Commercial struct {
	Code int `json:"code"`
	Data struct {
		Aid    int64 `json:"aid"`
		GameId int64 `json:"gameid"`
	} `json:"data"`
	Message string `json:"message"`
}

//nolint:gomnd
func (v *View) SubTitleChange() {
	if v.Stat.View > 100000 {
		tmp := strconv.FormatFloat(float64(v.Stat.View)/10000, 'f', 1, 64)
		v.ShareSubtitle = "已观看" + strings.TrimSuffix(tmp, ".0") + "万次"
	}
}

// TripleParam struct
type MaterialParam struct {
	MobiApp  string `form:"mobi_app"`
	Build    int32  `form:"build"`
	Device   string `form:"device"`
	Platform string `form:"platform"`
	AID      int64  `form:"aid" validate:"min=1"`
	CID      int64  `form:"cid" validate:"min=1"`
}

// MaterialRes struct
type MaterialRes struct {
	ID       int64  `json:"id"`
	Icon     string `json:"icon,omitempty"`
	URL      string `json:"url,omitempty"`
	Typ      int32  `json:"type"`
	Name     string `json:"name,omitempty"`
	BgColor  string `json:"bg_color,omitempty"`
	BgPic    string `json:"bg_pic,omitempty"`
	JumpType int32  `json:"jump_type,omitempty"`
}

// Ai推荐Request
type RecommendReq struct {
	Aid          int64  `json:"aid"`         //稿件id
	Mid          int64  `json:"mid"`         //用户id
	ZoneId       int64  `json:"zone_id"`     //地区标示id
	Build        int    `json:"build"`       //客户端版本号
	ParentMode   int    `json:"parent_mode"` //家长模式标记
	AutoPlay     int    `json:"auto_play"`   //view接口请求是否是自动播放行为触发的(ai那边已经下线)
	IsAct        int    `json:"is_act"`      //是否是活动页稿件
	Buvid        string `json:"buvid"`       //设备号id
	SourcePage   string `json:"source_page"` //请求相关推荐来源的页面
	TrackId      string `json:"track_id"`    //请求view接口时的trackid（标记用户的上一次请求）
	Cmd          string `json:"cmd"`
	TabId        string `json:"tab_id"`
	Plat         int8   `json:"plat"`          //平台
	AdExp        int8   `json:"ad_exp"`        //广告商单实验，广告商单是否由AI侧返回
	IsAd         int8   `json:"is_ad"`         //是否可出广告
	IsCommercial int8   `json:"is_commercial"` //是否可出商单
	AdResource   string `json:"ad_resource"`   //商业位次id
	MobileApp    string `json:"mobile_app"`    //客户端类型
	AdExtra      string `json:"ad_extra"`      //客户端透传的广告参数
	AvRid        int32  `json:"av_rid"`        //二级分区id
	AvPid        int32  `json:"av_pid"`        //一级分区id
	AvTid        string `json:"av_tid"`        //tag id list
	AvUpId       int64  `json:"av_up_id"`      //up主id
	FromSpmid    string `json:"from_spmid"`    //上级页面
	Spmid        string `json:"spmid"`         //当前页
	Network      string `json:"network"`       //网络
	RequestType  string `json:"request_type"`  //app端
	Device       string `json:"device"`        //设备
	PageVersion  string `json:"page_version"`  // 播放页页面版本
	AdFrom       string `json:"ad_from"`       //from
	AdTab        bool   `json:"ad_tab"`        //是否是商业tab
	DisableRcmd  int    `json:"disable_rcmd"`  //关闭个性化推荐，1关闭
	RecStyle     int    `json:"rec_style"`     //是否是新的样式 1-是 0-不是
	DeviceType   int64  `json:"device_type"`   //设备类型：透传给ai新激活设备是否出广告
	PageIndex    int64  `json:"page_index"`    //相关推荐请求页数
	DisplayId    int64  `json:"display_id"`    //表示当前第几刷的请求
	SessionId    string `json:"session_id"`    //唯一标识一个播放详情页，标识一个连播页面
	Ip           string `json:"ip"`            //ip
	Copyright    int32  `json:"copyright"`     //转载类型： 1=原创  2=转载 0=历史上可能遗留的脏数据
	InfeedPlay   int32  `json:"in_feed_play"`  //原地播放场景
	IsArcPay     int8   `json:"is_arc_pay"`    //1-付费视频
	IsFreeWatch  int8   `json:"is_free_watch"` //1-付费视频中免费观看
	IsUpBlue     int8   `json:"is_up_blue"`    //1-蓝v
	RefreshType  int32  `json:"refresh_type"`  //记录刷新类型
	RefreshNum   int32  `json:"refresh_num"`   //	记录同一详情页/session下刷新次数
}

//nolint:gomnd
func MaterialName(typ int32, name string) (string, error) {
	switch typ {
	case 1:
		return "活动:" + name, nil
	case 2:
		return "BGM:" + name, nil
	case 3:
		return "特效:" + name, nil
	case 4: //B剪
		return name, nil
	case 5: //视频模板（拍同款）
		return name, nil
	case 6: //合拍
		return name, nil
	default:
		return "", errors.New(fmt.Sprintf("material unknown type(%d)", typ))
	}
}

// ShareParam struct
type ShareParam struct {
	MobiApp      string `form:"mobi_app"`
	Build        int64  `form:"build"`
	Device       string `form:"device"`
	Platform     string `form:"platform"`
	From         string `form:"from"`
	ShareChannel string `form:"share_channel"`
	AID          int64  `form:"aid"`
	ShareTraceID string `form:"share_trace_id" validate:"required"`
	SeasonID     int64  `form:"season_id"`
	EpID         int64  `form:"ep_id"`
	OID          int64  `form:"oid"`            // 6.7版本接入直播分享，开始使用oid+type组合 aid可以不必传
	Type         string `form:"type"`           // 直播-live 稿件-av
	UpID         int64  `form:"up_id"`          // up主mid
	ParentAreaID int64  `form:"parent_area_id"` // 直播一级分区
	AreaID       int64  `form:"area_id"`        // 直播二级分区
	IsMelloi     string `form:"is_melloi"`      // 来自melloi
	Spmid        string `form:"spmid"`
	FromSpmid    string `form:"from_spmid"`
	AppKey       string `form:"appkey"`
}

// RelateTabParam struct
type RelateTabParam struct {
	MobiApp     string `form:"mobi_app"`
	Build       int    `form:"build"`
	Device      string `form:"device"`
	Platform    string `form:"platform"`
	From        string `form:"from"`
	FromTrackID string `form:"from_trackid"`
	FromAv      int64  `form:"from_av" validate:"required"`
	TabID       string `form:"tabid" validate:"required"`
	Qn          int    `form:"qn"`
	Fnver       int    `form:"fnver"`
	Fnval       int    `form:"fnval"`
	ForceHost   int    `form:"force_host"`
	Fourk       int    `form:"fourk"`
	NetType     int32
	TfType      int32
}

// TabInfo
type TabInfo struct {
	ID   string `json:"tabid"`
	IDx  int    `json:"tabidx"`
	Desc string `json:"tabdesc"`
}

// RelateTabRes
type RelateTabRes struct {
	Relates []*Relate `json:"relates"`
}

// RelateInfoc
type RelateInfoc struct {
	UserFeature string          `json:"user_feature"`
	PlayParam   int             `json:"play_param"`
	PvFeature   json.RawMessage `json:"pv_feature"`
	ReturnCode  string          `json:"return_code"`
}

// RelateConf
type RelateConf struct {
	AutoplayCountdown int    `json:"autoplay_countdown"`
	ReturnPage        int    `json:"return_page"`
	GamecardStyleExp  int    `json:"gamecard_style_exp"`
	AutoplayToast     string `json:"autoplay_toast"`
	HasDalao          int    `json:"has_dalao"`
	RelatesStyle      int    `json:"relates_style"`
	GifExp            int    `json:"gif_exp"`
	//相关推荐三点类型：0-旧样式 1-新样式
	RecThreePointStyle int32   `json:"rec_three_point_style"`
	FeedStyle          string  `json:"feed_style"`
	FeedPopUp          int     `json:"has_guide"`
	FeedHasNext        bool    `json:"feed_has_next"`
	LocalPlay          int     `json:"local_play"`
	Next               string  `json:"next"`
	Refreshable        int     `json:"refreshable"`  //是否支持顶部刷新
	RefreshIcon        int     `json:"refresh_icon"` //悬浮按钮icon类型
	RefreshText        string  `json:"refresh_text"` //icon 文本
	RefreshShow        float32 `json:"refresh_show"` //悬浮按钮出现时机
}

func ArchivePage(in *steinsApi.Page) (out *api.Page) {
	out = new(api.Page)
	out.Cid = in.Cid
	out.Page = in.Page
	out.From = in.From
	out.Part = in.Part
	out.Duration = in.Duration
	out.Vid = in.Vid
	out.Desc = in.Desc
	out.WebLink = in.WebLink
	out.Dimension = api.Dimension{
		Width:  in.Dimension.Width,
		Height: in.Dimension.Height,
		Rotate: in.Dimension.Rotate,
	}
	return
}

//nolint:gomnd
func (r *Relate) ReasonStyleFrom(rcmd *RcmdReason, isNewColor bool) {
	if rcmd == nil || rcmd.Content == "" {
		return
	}
	const (
		_isAtten = 1 // 已关注
	)
	style := model.BgColorOrange
	if isNewColor {
		style = model.BgLightColoredOrange
	}
	r.ReasonStyle = reasonStyle(style, rcmd.Content)
	switch rcmd.Style {
	case 3:
		r.ReasonStyle.Selected = _isAtten
	}
}

func reasonStyle(style int32, text string) (res *viewApi.ReasonStyle) {
	res = &viewApi.ReasonStyle{
		Text: text,
	}
	switch style {
	case model.BgColorOrange:
		// 白天
		res.TextColor = "#FFFFFF"
		res.BgColor = "#FB9E60"
		res.BorderColor = "#FB9E60"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#E5E5E5"
		res.BgColorNight = "#BC7A4F"
		res.BorderColorNight = "#BC7A4F"
	case model.BgLightColoredOrange:
		// 白天
		res.TextColor = "#FF6633"
		res.BgColor = "#FFF1ED"
		res.BorderColor = "#FFF1ED"
		res.BgStyle = model.BgStyleFill
		// 夜间
		res.TextColorNight = "#BF5330"
		res.BgColorNight = "#3D2D29"
		res.BorderColorNight = "#3D2D29"
	case model.BgColorTransparentRed:
		// 白天
		res.TextColor = "#FB7299"
		res.BgColor = "#FB7299"
		res.BorderColor = "#FB7299"
		res.BgStyle = model.BgStyleStroke
		// 夜间
		res.TextColorNight = "#BB5B76"
		res.BgColorNight = "#BB5B76"
		res.BorderColorNight = "#BB5B76"
	}

	return
}

// VideoShotParam struct
type VideoShotParam struct {
	MobiApp string `form:"mobi_app"`
	Build   int32  `form:"build"`
	AID     int64  `form:"aid" validate:"min=1"`
	CID     int64  `form:"cid" validate:"min=1"`
}

func FromOwnerExt(in OwnerExt) (out *viewApi.OnwerExt) {
	var (
		ownerLive *viewApi.Live
		vip       *viewApi.Vip
	)
	if in.Live != nil {
		ownerLive = &viewApi.Live{
			Mid:        in.Live.Mid,
			Roomid:     in.Live.RoomID,
			Uri:        in.Live.URI,
			EndpageUri: in.Live.EndPageUri,
		}
	}
	vip = &viewApi.Vip{
		Type:          int32(in.Vip.Type),
		DueDate:       in.Vip.DueDate,
		DueRemark:     in.Vip.DueRemark,
		AccessStatus:  int32(in.Vip.AccessStatus),
		VipStatus:     int32(in.Vip.VipStatus),
		VipStatusWarn: in.Vip.VipStatusWarn,
		ThemeType:     int32(in.Vip.ThemeType),
		Label: &viewApi.VipLabel{
			Path:       in.Vip.Label.Path,
			Text:       in.Vip.Label.Text,
			LabelTheme: in.Vip.Label.LabelTheme,
		},
	}
	out = &viewApi.OnwerExt{
		OfficialVerify: &viewApi.OfficialVerify{
			Type: int32(in.OfficialVerify.Type),
			Desc: in.OfficialVerify.Desc,
		},
		Live:     ownerLive,
		Vip:      vip,
		Assists:  in.Assists,
		Fans:     int64(in.Fans),
		ArcCount: in.ArcCount,
	}
	return
}

func FromPages(in []*Page) (out []*viewApi.ViewPage) {
	for _, p := range in {
		if p == nil {
			continue
		}
		var (
			au *viewApi.Audio
			dm *viewApi.DM
		)
		if p.Audio != nil {
			au = &viewApi.Audio{
				Title:      p.Audio.Title,
				CoverUrl:   p.Audio.Cover,
				SongId:     int64(p.Audio.SongID),
				PlayCount:  int64(p.Audio.Play),
				ReplyCount: int64(p.Audio.Reply),
				UpperId:    int64(p.Audio.UpperID),
				Entrance:   p.Audio.Entrance,
				SongAttr:   int64(p.Audio.SongAttr),
			}
		}
		if p.DM != nil {
			dm = &viewApi.DM{
				Closed:   p.DM.Closed,
				RealName: p.DM.RealName,
				Count:    p.DM.Count,
			}
		}
		out = append(out, &viewApi.ViewPage{
			Page:             p.Page,
			Audio:            au,
			Dm:               dm,
			DownloadTitle:    p.DlTitle,
			DownloadSubtitle: p.DlSubtitle,
		})
	}
	return
}

func FromTag(in []*tag.Tag) (out []*viewApi.Tag) {
	for _, v := range in {
		out = append(out, &viewApi.Tag{
			Id:      v.TagID,
			Name:    v.Name,
			Likes:   v.Likes,
			Hates:   v.Hates,
			Liked:   v.Liked,
			Hated:   v.Hated,
			Uri:     v.URI,
			TagType: v.TagType,
		})
	}
	return
}

func FromRelates(in []*Relate) (out []*viewApi.Relate) {
	for _, v := range in {
		if v == nil {
			continue
		}
		tmp := &viewApi.Relate{
			Aid:               v.Aid,
			Pic:               v.Pic,
			Title:             v.Title,
			Author:            v.Author,
			Stat:              &v.Stat,
			Duration:          v.Duration,
			Goto:              v.Goto,
			Param:             v.Param,
			Uri:               v.URI,
			JumpUrl:           v.JumpURL,
			Rating:            v.Rating,
			Reserve:           v.Reserve,
			From:              v.From,
			Desc:              v.Desc,
			RcmdReason:        v.RcmdReason,
			Badge:             v.Badge,
			Cid:               v.Cid,
			SeasonType:        v.SeasonType,
			RatingCount:       v.RatingCount,
			TagName:           v.TagName,
			PackInfo:          v.PackInfo,
			Notice:            v.Notice,
			Button:            v.Button,
			Trackid:           v.TrackID,
			NewCard:           int32(v.NewCard),
			RcmdReasonStyle:   v.ReasonStyle,
			CoverGif:          v.CoverGif,
			Cm:                v.CM,
			ReserveStatus:     v.ReserveStatus,
			RcmdReasonExtra:   v.RcmdReasonExtra,
			RecThreePoint:     v.RecThreePoint,
			UniqueId:          v.UniqueId,
			MaterialId:        v.MaterialId,
			FromSourceType:    v.FromSourceType,
			FromSourceId:      v.FromSourceId,
			Dimension:         &v.Dimension,
			BadgeStyle:        v.BadgeStyle,
			PowerIconStyle:    v.PowerIconStyle,
			ReserveStatusText: v.ReserveStatusText,
			Cover:             v.Cover,
			DislikeReportData: v.DislikeReportData,
			FirstFrame:        v.FirstFrame,
		}
		if v.RankInfo != nil {
			tmp.RankInfoGame = &viewApi.RankInfo{
				IconUrlNight:   v.RankInfo.SearchNightIconUrl,
				IconUrlDay:     v.RankInfo.SearchDayIconUrl,
				BkgNightColor:  v.RankInfo.SearchBkgNightColor,
				BkgDayColor:    v.RankInfo.SearchBkgDayColor,
				FontNightColor: v.RankInfo.SearchFontNightColor,
				FontDayColor:   v.RankInfo.SearchFontDayColor,
				RankContent:    v.RankInfo.RankContent,
				RankLink:       v.RankInfo.RankLink,
			}
		}
		out = append(out, tmp)
	}
	return
}

func FromStaff(in []*Staff) (out []*viewApi.Staff) {
	for _, v := range in {
		if v == nil {
			continue
		}
		out = append(out, &viewApi.Staff{
			Mid:   v.Mid,
			Title: v.Title,
			Face:  v.Face,
			Name:  v.Name,
			OfficialVerify: &viewApi.OfficialVerify{
				Type: int32(v.OfficialVerify.Type),
				Desc: v.OfficialVerify.Desc,
			},
			Vip: &viewApi.Vip{
				Type:          int32(v.Vip.Type),
				DueDate:       v.Vip.DueDate,
				DueRemark:     v.Vip.DueRemark,
				AccessStatus:  int32(v.Vip.AccessStatus),
				VipStatus:     int32(v.Vip.VipStatus),
				VipStatusWarn: v.Vip.VipStatusWarn,
				ThemeType:     int32(v.Vip.ThemeType),
				Label: &viewApi.VipLabel{
					Path:       v.Vip.Label.Path,
					Text:       v.Vip.Label.Text,
					LabelTheme: v.Vip.Label.LabelTheme,
				},
			},
			Attention:  int32(v.Attention),
			LabelStyle: v.LabelStyle,
		})
	}
	return
}

// FromHonor func
func FromHonor(in *ahApi.Honor) (out *viewApi.Honor) {
	if in == nil {
		return
	}
	out = &viewApi.Honor{
		Url:            in.Url,
		Icon:           model.HonorIcon[in.Type],
		IconNight:      model.HonorIconNight[in.Type],
		Text:           in.Desc,
		TextExtra:      model.HonorTextExtra[in.Type],
		TextColor:      model.HonorTextColor[in.Type],
		TextColorNight: model.HonorTextColorNight[in.Type],
		BgColor:        model.HonorBgColor[in.Type],
		BgColorNight:   model.HonorBgColorNight[in.Type],
		UrlText:        model.HonorURLText[in.Type],
	}
	return
}

// FromReplyHonor func
func FromReplyHonor(in *replyApi.ArchiveHonorResp) (out *viewApi.Honor) {
	if in == nil {
		return
	}
	out = &viewApi.Honor{
		Url:            in.ArchiveHonor.Url,
		Icon:           in.ArchiveHonor.Icon,
		IconNight:      in.ArchiveHonor.IconNight,
		Text:           in.ArchiveHonor.Text,
		TextExtra:      in.ArchiveHonor.TextExtra,
		TextColor:      in.ArchiveHonor.TextColor,
		TextColorNight: in.ArchiveHonor.TextColorNight,
		BgColor:        in.ArchiveHonor.BgColor,
		BgColorNight:   in.ArchiveHonor.BgColorNight,
		UrlText:        in.ArchiveHonor.UrlText,
	}
	return
}

// FromMusicHonor func
func FromMusicHonor(in *musicApi.ToplistEntranceReply) (out *viewApi.Honor) {
	if in == nil {
		return
	}
	out = &viewApi.Honor{
		Url:            in.ArcHonor.Url,
		Icon:           in.ArcHonor.Icon,
		IconNight:      in.ArcHonor.IconNight,
		Text:           in.ArcHonor.Text,
		TextExtra:      in.ArcHonor.TextExtra,
		TextColor:      in.ArcHonor.TextColor,
		TextColorNight: in.ArcHonor.TextColorNight,
		BgColor:        in.ArcHonor.BgColor,
		BgColorNight:   in.ArcHonor.BgColorNight,
		UrlText:        in.ArcHonor.UrlText,
	}
	return
}

func FromUgcSeason(in *UgcSeason) (out *viewApi.UgcSeason) {
	if in == nil {
		return
	}
	out = &viewApi.UgcSeason{
		Id:    in.Id,
		Title: in.Title,
		Cover: in.Cover,
		Intro: in.Intro,
		Stat: &viewApi.UgcSeasonStat{
			SeasonId: in.Stat.SeasonID,
			View:     in.Stat.View,
			Danmaku:  in.Stat.Danmaku,
			Reply:    in.Stat.Reply,
			Fav:      in.Stat.Fav,
			Coin:     in.Stat.Coin,
			Share:    in.Stat.Share,
			NowRank:  in.Stat.NowRank,
			HisRank:  in.Stat.HisRank,
		},
		LabelText:           in.LabelText,
		LabelTextColor:      in.LabelTextColor,
		LabelBgColor:        in.LabelBgColor,
		LabelTextNightColor: in.LabelTextNightColor,
		LabelBgNightColor:   in.LabelBgNightColor,
		DescRight:           in.DescRight,
		EpCount:             in.EpCount,
		SeasonType:          in.SeasonType,
		ShowContinualButton: in.ShowContinualButton,
		Activity:            in.Activity,
		EpNum:               in.EpNum,
		SeasonPay:           in.SeasonPay,
		SeasonAbility:       in.SeasonAbility,
		GoodsInfo: &viewApi.GoodsInfo{
			GoodsId:    in.GoodsInfo.GoodsId,
			Category:   in.GoodsInfo.Category,
			GoodsPrice: in.GoodsInfo.GoodsPrice,
			PayState:   in.GoodsInfo.PayState,
			GoodsName:  in.GoodsInfo.GoodsName,
			PriceFmt:   in.GoodsInfo.PriceFmt,
		},
		PayButton: &viewApi.ButtonStyle{
			Text:           in.PayButton.Text,
			TextColor:      in.PayButton.TextColor,
			TextColorNight: in.PayButton.TextColorNight,
			BgColor:        in.PayButton.BgColor,
			BgColorNight:   in.PayButton.BgColorNight,
			JumpLink:       "",
		},
		LabelTextNew: in.LabelTextNew,
	}
	for _, s := range in.Sections {
		if s == nil {
			continue
		}
		var tmpEp []*viewApi.Episode
		for _, e := range s.Episodes {
			if e == nil {
				continue
			}
			var (
				tmpPage       *api.Page
				tmpStat       *api.Stat
				tmpAuthor     *api.Author
				tmpBadgeStyle *viewApi.BadgeStyle
			)
			if e.ArcPage != nil {
				tmpPage = &api.Page{
					Cid:      e.ArcPage.Cid,
					Page:     e.ArcPage.Page,
					From:     e.ArcPage.From,
					Part:     e.ArcPage.Part,
					Duration: e.ArcPage.Duration,
					Vid:      e.ArcPage.Vid,
					Desc:     e.ArcPage.Desc,
					WebLink:  e.ArcPage.WebLink,
					Dimension: api.Dimension{
						Width:  e.ArcPage.Dimension.Width,
						Height: e.ArcPage.Dimension.Height,
						Rotate: e.ArcPage.Dimension.Rotate,
					},
				}
			}
			if e.Stat != nil {
				tmpStat = &api.Stat{
					Aid:     e.Stat.Aid,
					View:    e.Stat.View,
					Danmaku: e.Stat.Danmaku,
					Reply:   e.Stat.Reply,
					Fav:     e.Stat.Fav,
					Coin:    e.Stat.Coin,
					Share:   e.Stat.Share,
					NowRank: e.Stat.NowRank,
					HisRank: e.Stat.HisRank,
					Like:    e.Stat.Like,
					DisLike: e.Stat.DisLike,
				}
			}
			if e.Author != nil {
				tmpAuthor = &api.Author{
					Mid:  e.Author.Mid,
					Name: e.Author.Name,
					Face: e.Author.Face,
				}
			}
			if e.BadgeStyle != nil {
				tmpBadgeStyle = &viewApi.BadgeStyle{
					Text:             e.BadgeStyle.Text,
					TextColor:        e.BadgeStyle.TextColor,
					TextColorNight:   e.BadgeStyle.TextColorNight,
					BgColor:          e.BadgeStyle.BgColor,
					BgColorNight:     e.BadgeStyle.BgColorNight,
					BorderColor:      e.BadgeStyle.BorderColor,
					BorderColorNight: e.BadgeStyle.BorderColorNight,
					BgStyle:          e.BadgeStyle.BgStyle,
				}
			}
			tmpEp = append(tmpEp, &viewApi.Episode{
				Id:             e.Id,
				Aid:            e.Aid,
				Cid:            e.Cid,
				Title:          e.Title,
				Cover:          e.Cover,
				CoverRightText: e.CoverRightText,
				Page:           tmpPage,
				Stat:           tmpStat,
				Bvid:           e.BvID,
				Author:         tmpAuthor,
				AuthorDesc:     e.AuthorDesc,
				BadgeStyle:     tmpBadgeStyle,
				NeedPay:        e.NeedPay,
				EpisodePay:     e.EpisodePay,
				FreeWatch:      e.FreeWatch,
				FirstFrame:     e.FirstFrame,
			})
		}
		out.Sections = append(out.Sections, &viewApi.Section{
			Id:       s.Id,
			Title:    s.Title,
			Type:     s.Type,
			Episodes: tmpEp,
		})
	}
	return
}

func FromConfig(in *Config) (out *viewApi.Config) {
	if in == nil {
		return
	}
	out = &viewApi.Config{
		RelatesTitle:       in.RelatesTitle,
		RelatesStyle:       int32(in.RelatesStyle),
		RelateGifExp:       int32(in.RelateGifExp),
		EndPageHalf:        int32(in.EndPageHalf),
		EndPageFull:        int32(in.EndPageFull),
		AutoSwindow:        in.AutoSwindow,
		PopupInfo:          in.PopupInfo,
		AbtestSmallWindow:  in.AbTestSmallWindow,
		RecThreePointStyle: in.RecThreePointStyle,
		IsAbsoluteTime:     in.IsAbsoluteTime,
		NewSwindow:         in.NewSwindow,
		RelatesFeedStyle:   in.FeedStyle,
		RelatesFeedPopup:   in.FeedPopUp,
		RelatesHasNext:     in.FeedHasNext,
		RelatesBiserial:    in.RelatesBiserial,
		ListenerConf:       in.ListenerConfig,
		LocalPlay:          in.LocalPlay,
		PlayStory:          in.PlayStory,
		ArcPlayStory:       in.ArcPlayStory,
		StoryIcon:          in.StoryIcon,
		LandscapeStory:     in.LandscapeStory,
		LandscapeIcon:      in.LandscapeIcon,
		ArcLandscapeStory:  in.ArcLandscapeStory,
		ShowListenButton:   in.ShowListenButton,
	}
	return
}

func FromPlayerIcon(in *resmdl.PlayerIcon) (out *viewApi.PlayerIcon) {
	if in == nil {
		return
	}
	out = &viewApi.PlayerIcon{
		Url1:         in.URL1,
		Url2:         in.URL2,
		Hash1:        in.Hash1,
		Hash2:        in.Hash2,
		DragRightPng: in.DragRightPng,
		MiddlePng:    in.MiddlePng,
		DragLeftPng:  in.DragLeftPng,
	}
	if in.DragData != nil {
		out.DragData = &viewApi.IconData{
			MetaJson:  in.DragData.MetaJson,
			SpritsImg: in.DragData.SpritsImg,
		}
	}
	if in.NoDragData != nil {
		out.NodragData = &viewApi.IconData{
			MetaJson:  in.NoDragData.MetaJson,
			SpritsImg: in.NoDragData.SpritsImg,
		}
	}
	return
}

func FromSeason(in *bangumi.Season) (out *viewApi.Season) {
	if in == nil {
		return
	}
	seasonID, _ := strconv.ParseInt(in.SeasonID, 10, 64)
	isFinish, _ := strconv.ParseInt(in.IsFinish, 10, 64)
	newestEpID, _ := strconv.ParseInt(in.NewestEpID, 10, 64)
	totalCnt, _ := strconv.ParseInt(in.TotalCount, 10, 64)
	weekday, _ := strconv.ParseInt(in.Weekday, 10, 64)
	out = &viewApi.Season{
		AllowDownload: in.AllowDownload,
		SeasonId:      seasonID,
		IsJump:        int32(in.IsJump),
		Title:         in.Title,
		Cover:         in.Cover,
		IsFinish:      int32(isFinish),
		NewestEpId:    newestEpID,
		NewestEpIndex: in.NewestEpIndex,
		TotalCount:    totalCnt,
		Weekday:       int32(weekday),
		OgvPlayurl:    in.OGVPlayURL,
	}
	if in.UserSeason != nil {
		out.UserSeason = &viewApi.UserSeason{
			Attention: in.UserSeason.Attention,
		}
	}
	if in.Player != nil {
		out.Player = &viewApi.SeasonPlayer{
			Aid:  in.Player.Aid,
			Vid:  in.Player.Vid,
			Cid:  in.Player.Cid,
			From: in.Player.From,
		}
	}
	return
}

func FromPlayerCustomizedPanel(inRes *resApi.GetPlayerCustomizedPanelV2Rep) *viewApi.TFPanelCustomized {
	if inRes.GetItem() == nil {
		return nil
	}
	in := inRes.GetItem()
	out := &viewApi.TFPanelCustomized{
		RightBtnImg:       in.BtnImg,
		RightBtnText:      in.BtnText,
		RightBtnTextColor: in.TextColor,
		RightBtnLink:      in.Link,
		SubPanel:          map[string]*viewApi.SubTFPanel{},
	}
	allPanels := map[string]*resApi.PlayerPanel{}
	for _, v := range in.Panels {
		v.Operator = strings.ToLower(v.Operator)
		if v.Operator != "" {
			v.DisplayStage = fmt.Sprintf("%s_%s", v.DisplayStage, v.Operator)
		}
		allPanels[v.DisplayStage] = v
	}
	beforePlay, ok := allPanels["before_play"]
	if ok {
		out.RightBtnImg = beforePlay.BtnImg
		out.RightBtnText = beforePlay.BtnText
		out.RightBtnTextColor = beforePlay.TextColor
		out.RightBtnLink = beforePlay.Link
		out.MainLabel = beforePlay.Label
		out.Operator = beforePlay.Operator
	}
	for stage, panel := range allPanels {
		out.SubPanel[stage] = &viewApi.SubTFPanel{
			RightBtnImg:       panel.BtnImg,
			RightBtnText:      panel.BtnText,
			RightBtnTextColor: panel.TextColor,
			RightBtnLink:      panel.Link,
			MainLabel:         panel.Label,
			Operator:          panel.Operator,
		}
	}
	return out
}

// LikeNoLoginParam is
type LikeNoLoginParam struct {
	MobiApp   string `form:"mobi_app"`
	Build     string `form:"build"`
	Aid       int64  `form:"aid" validate:"min=1"`
	From      string `form:"from"`
	FromSpmid string `form:"from_spmid"`
	Like      int32  `form:"like"`
	Action    string `form:"action" validate:"required"` //未登录点赞来源点赞or三连
	OgvType   int64  `form:"ogv_type"`
	Device    string `form:"device"`
	Platform  string `form:"platform"`
	Appkey    string `form:"appkey"`
	Spmid     string `form:"spmid"`
	TrackID   string `form:"track_id"`
	Goto      string `form:"goto"`
}

// LikeNoLoginRes is
type LikeNoLoginRes struct {
	Toast     string `json:"toast"`
	NeedLogin int    `json:"need_login"`
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
	ItemType    string `json:"item_type,omitempty"`
	ShareSource string `json:"share_source,omitempty"`
	Title       string `json:"title,omitempty"`
	PlayNum     int64  `json:"play_num,omitempty"`
	CoinNum     int64  `json:"coin_num,omitempty"`
	Token       string `json:"token,omitempty"`
}

type UserActInfoc struct {
	Buvid    string
	Build    string
	Client   string
	Ip       string
	Uid      int64
	Aid      int64
	Mid      int64
	Sid      string
	Refer    string
	Url      string
	From     string
	ItemID   string
	ItemType string
	Action   string
	ActionID string
	Ua       string
	Ts       string
	Extra    string
	IsRisk   string
}

type VideoOnlineParam struct {
	Aid   int64 `form:"aid" validate:"min=1"`
	Cid   int64 `form:"cid" validate:"min=1"`
	Buvid string
	Mid   int64
	//在线人数展示的场景
	//0:ugc横屏, 1:ugc竖屏, 2:story
	Scene int64 `form:"scene"`
}

type VideoOnlineRes struct {
	Online struct {
		TotalText string `json:"total_text"`
	} `json:"online"`
	LikeSwitch bool `json:"like_switch"`
}

type VideoDownloadReq struct {
	*viewApi.ShortFormVideoDownloadReq
	TfType int32
}

type VideoDownloadReply struct {
	*viewApi.ShortFormVideoDownloadReply
}

// dmVoteReq struct
type DmVoteReq struct {
	MobiApp       string `form:"mobi_app"`
	Build         int32  `form:"build"`
	Device        string `form:"device"`
	Platform      string `form:"platform"`
	AID           int64  `form:"aid" validate:"min=1"`
	CID           int64  `form:"cid" validate:"min=1"`
	TeenagersMode int32  `form:"teenagers_mode"`
	LessonsMode   int32  `form:"lessons_mode"`
	Vote          int32  `form:"vote" validate:"min=1"`
	VoteID        int64  `form:"vote_id" validate:"min=1"`
	Progress      int32  `form:"progress" validate:"min=1"`
	Mid           int64
	Buvid         string
}

type DmVoteReply struct {
	Vote *VoteReply `json:"vote,omitempty"`
	Dm   *DmReply   `json:"dm,omitempty"`
}

type VoteReply struct {
	Uid  int64 `json:"uid"`
	Type int32 `json:"type"`
}

type DmReply struct {
	DmID    int64  `json:"dm_id"`
	DmIDStr string `json:"dm_id_str"`
	Visible bool   `json:"visible"`
	Action  string `json:"action"`
}

type BiJianMaterialReq struct {
	Type int64 `json:"type"`
	Ids  int64 `json:"ids"`
	Biz  int64 `json:"biz"`
}

type BiJianMaterialReply struct {
	Code int `json:"code"`
	Data struct {
		List []*PointMaterial `json:"list"`
	} `json:"data"`
}

type PointMaterial struct {
	DownloadUrl string `json:"download_url"`
}

type StatReq struct {
	Aid int64 `form:"aid" validate:"min=1"`
}

type StatReply struct {
	Stat struct {
		Like int64 `json:"like"`
	} `json:"stat"`
}

type DmExtra struct {
	ReserveType int32 `json:"reserve_type"`
	ReserveId   int64 `json:"reserve_id"`
}

type MaterialResArr []*MaterialRes

func (m MaterialResArr) Len() int {
	return len(m)
}
func (m MaterialResArr) Less(i, j int) bool {
	return m[i].Typ > m[j].Typ
}
func (m MaterialResArr) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type BizExtra struct {
	AdPlayPage     int8 `json:"ad_play_page"`
	TeenagerExempt int  `json:"teenager_exempt"`
}

type ContinuousInfo struct {
	Ip          string `json:"ip"`            //ip信息
	Now         string `json:"now"`           //请求时间
	Api         string `json:"api"`           //接口api
	Buvid       string `json:"buvid"`         //buvid
	Mid         string `json:"mid"`           //mid
	Client      string `json:"client"`        //客户端
	MobiApp     string `json:"mobi_app"`      //mobi
	From        string `json:"from"`          //
	ShowList    string `json:"show_list"`     //卡片list json格式
	IsRec       int    `json:"is_rec"`        //标识是否推荐结果
	Build       string `json:"build"`         //客户端build号
	ReturnCode  string `json:"return_code"`   //推荐服务返回值
	DeviceId    string `json:"device_id"`     //设备id
	Network     string `json:"network"`       //网络
	TrackId     string `json:"track_id"`      //ai返回的track_id
	FromTrackId string `json:"from_track_id"` //来源的trackid
	Spmid       string `json:"spmid"`         //当前页面的spmid
	FromSpmid   string `json:"from_spmid"`    //来源页面的spmid
	UserFeature string `json:"user_feature"`  //用户特征（ai reponse中user_feature字段透传）
	DisplayId   string `json:"display_id"`    //标记session内的刷次
	FromAv      string `json:"from_av"`       //aid
}

type InspirationMaterial struct {
	Title         string `json:"title"`
	Url           string `json:"url"`
	InspirationId int64  `json:"inspiration_id"`
}
