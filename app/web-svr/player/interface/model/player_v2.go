package model

import (
	"encoding/json"
	"time"

	resmdl "go-gateway/app/app-svr/resource/service/model"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	dmapi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
)

type PlayerV2 struct {
	// 视频相关
	Aid      int64  `json:"aid"`       // aid
	Bvid     string `json:"bvid"`      // bvid
	AllowBp  bool   `json:"allow_bp"`  // 是否允许承包
	NoShare  bool   `json:"no_share"`  // 是否禁止分享
	Cid      int64  `json:"cid"`       // cid，原chatid
	MaxLimit int64  `json:"max_limit"` // 弹幕数限制
	PageNo   int32  `json:"page_no"`   // 当前分p是第几p，原pid
	HasNext  bool   `json:"has_next"`  // 当前分p是否有下一p
	// ip 相关
	IPInfo *PlayerIPInfo `json:"ip_info"` // 用户ip相关信息
	// 登陆用户相关
	LoginMid     int64            `json:"login_mid"`      // 登录用户id，原login和user
	LoginMidHash string           `json:"login_mid_hash"` // 登录用户id加密后hash串，原user_hash
	IsOwner      bool             `json:"is_owner"`       // 登陆用户是否该视频up，原IsAdmin
	Name         string           `json:"name"`           // 登陆用户名
	Permission   string           `json:"permission"`     // 登陆用户rank值，特殊权限相关
	LevelInfo    accapi.LevelInfo `json:"level_info"`     // 登陆用户等级信息
	Vip          accapi.VipInfo   `json:"vip"`            // 登陆用户vip信息
	AnswerStatus int32            `json:"answer_status"`  // 登陆用户答题状态，针对未转正用户
	BlockTime    int64            `json:"block_time"`     // 登陆用户封禁时间
	Role         string           `json:"role"`           // 用户身份,多个时用逗号隔开（第一位:是否是当前视频up主协管）
	// 播放进度相关
	LastPlayTime int64 `json:"last_play_time"` // 上次观看进度
	LastPlayCid  int64 `json:"last_play_cid"`  // 上次观看cid
	// 其它
	NowTime         int64                 `json:"now_time"`                // 请求服务器时间
	OnlineCount     int64                 `json:"online_count"`            // 在线人数
	DmMask          *PlayerDmMask         `json:"dm_mask,omitempty"`       // 蒙板
	Subtitle        *dmapi.VideoSubtitles `json:"subtitle,omitempty"`      // 字幕
	PlayerIcon      *resmdl.PlayerIcon    `json:"player_icon,omitempty"`   // 进度条icon
	ViewPoints      []*Point              `json:"view_points"`             // 高能看点
	IsUgcPayPreview bool                  `json:"is_ugc_pay_preview"`      // 是否为ugc付费预览视频
	PreviewToast    string                `json:"preview_toast"`           // ugc付费预览toast 用 | 分割2个文案
	Interaction     *Interaction          `json:"interaction,omitempty"`   // 互动视频相关数据
	Pugv            *PlayerPugv           `json:"pugv,omitempty"`          // pugv相关
	PcdnLoader      json.RawMessage       `json:"pcdn_loader"`             // pcdn loader
	Options         *Option               `json:"options"`                 // 其他选项 包含is_360字段，表示视频是否为全景视频
	GuideAttention  []*GuideAttention     `json:"guide_attention"`         // 关注引导卡
	JumpCard        []*JumpCard           `json:"jump_card"`               // 跳转卡
	OperationCard   []*OperationCard      `json:"operation_card"`          // 运营卡
	OnlineSwitch    map[string]string     `json:"online_switch,omitempty"` // 在线开关 subtitle_submit_switch
	Fawkes          *Fawkes               `json:"fawkes,omitempty"`
	ShowSwitch      *ShowSwitch           `json:"show_switch"`
	BgmInfo         *BgmInfo              `json:"bgm_info"` // bgm音乐信息
}

type ShowSwitch struct {
	LongProgress bool `json:"long_progress"`
}

type PlayerIPInfo struct {
	IP       string `json:"ip"`
	ZoneIP   string `json:"zone_ip"`
	ZoneID   int64  `json:"zone_id"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
}

type PlayerPugv struct {
	WatchStatus  int32 `json:"watch_status"`  // 试看状态(1=整集试看 2=非试看 3=5分钟试看)
	PayStatus    int32 `json:"pay_status"`    // 购买状态(1=购买 2=未购买)
	SeasonStatus int32 `json:"season_status"` // 系列状态(1=上架 2=下架)
}

type PlayerDmMask struct {
	Cid     int64  `json:"cid"`
	Plat    int32  `json:"plat"`
	Fps     int32  `json:"fps"`
	Time    int64  `json:"time"`
	MaskUrl string `json:"mask_url"`
}

type PlayerV2Arg struct {
	Aid          int64     `form:"aid"`
	Bvid         string    `form:"bvid"`
	Cid          int64     `form:"cid" validate:"min=1"`
	GraphVersion int64     `form:"graph_version"`
	SeasonID     int64     `form:"season_id"`
	EpID         int64     `form:"ep_id"`
	Buvid        string    `form:"-"`
	Refer        string    `form:"-"`
	InnerSign    string    `form:"-"`
	CdnIP        string    `form:"-"`
	Now          time.Time `form:"-"`
	FawkesAppKey string    `form:"fawkes_app_key" default:"web_main"`
	FawkesEnv    string    `form:"fawkes_env" default:"prod"`
}

type Fawkes struct {
	ConfigVersion int64 `json:"config_version"`
	FFVersion     int64 `json:"ff_version"`
}

type PlayerCardClickArg struct {
	ID      int64  `form:"id" validate:"min=1"`
	OidType int    `form:"oid_type" validate:"min=1,max=2"`
	Oid     int64  `form:"oid" validate:"min=1"`
	Pid     int64  `form:"pid"`
	Action  int64  `form:"action" validate:"min=0,max=1"`
	Buvid   string `form:"-"`
}
