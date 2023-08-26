package splash_screen

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/api"
)

const (
	IsDeleted            = 1
	NotDeleted           = 0
	ShowModeForceOrder   = 1   // 强制-顺序
	ShowModeForceRate    = 2   // 强制-概率
	ShowModeDefaultOrder = 3   // 默认-顺序
	ShowModeDefaultRate  = 4   // 默认-概率
	ShowModeSelect       = 5   // 用户自选
	AuditStatePass       = 1   // 审核通过
	AuditStateCancel     = 0   // 审核取消，拒绝
	AuditStateOffline    = 2   // 手动下线
	ImageModeHalfScreen  = 1   // 物料模式-半屏
	ImageModeFullScreen  = 2   // 物料模式-全屏
	LogoShow             = 0   // LOGO-显示
	LogoHide             = 1   // LOGO-隐藏
	LogoModePink         = 1   // 物料-粉色LOGO
	LogoModeWhite        = 2   // 物料-白色LOGO
	LogoModeUser         = 3   // 物料-用户自定义LOGO
	ActionLogBusiness    = 209 // 行为日志对应的id
)

// 图片物料
type SplashScreenImage struct {
	ID              int64      `json:"id" form:"id" gorm:"column:id"`
	ImageName       string     `json:"img_name" form:"img_name" gorm:"column:img_name"`
	ImageUrl        string     `json:"img_url" form:"img_url" gorm:"column:img_url"`
	Mode            int        `json:"mode" form:"mode" gorm:"column:mode"`
	ImageUrlNormal  string     `json:"img_url_normal" form:"img_url_normal" gorm:"column:img_url_normal"`
	ImageUrlFull    string     `json:"img_url_full" form:"img_url_full" gorm:"column:img_url_full"`
	ImageUrlPad     string     `json:"img_url_pad" form:"img_url_pad" gorm:"column:img_url_pad"`
	LogoHideFlag    int        `json:"-" form:"logo_hide" gorm:"column:logo_hide"`
	LogoShowFlag    int        `json:"logo_show" form:"logo_show" gorm:"-"`
	LogoMode        int        `json:"logo_mode" form:"logo_mode" gorm:"column:logo_mode"`
	LogoImageUrl    string     `json:"logo_img_url" form:"logo_img_url" gorm:"column:logo_img_url"`
	InitialPushTime xtime.Time `json:"initial_push_time" form:"initial_push_time" gorm:"column:initial_push_time"`
	CategoryIDs     []int64    `json:"category_ids" form:"-" gorm:"-"`
	IsDeleted       int        `json:"-" form:"-" gorm:"column:is_deleted"`
	CUser           string     `json:"c_user" form:"-" gorm:"column:c_user"`
	MUser           string     `json:"m_user" form:"-" gorm:"column:m_user"`
	CTime           xtime.Time `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime           xtime.Time `json:"mtime" form:"-" gorm:"column:mtime"`
}

func (*SplashScreenImage) TableName() string {
	return "splash_screen_images"
}

type LogoConfig struct {
	Show     bool   `json:"show"`
	Mode     int    `json:"mode"`
	ImageUrl string `json:"img_url"`
}

type FullScreenImageUrl struct {
	Normal string `json:"normal"`
	Full   string `json:"full"`
	Pad    string `json:"pad"`
}

type SplashScreenImageForGateway struct {
	ID                 int64               `json:"id"`
	ImageName          string              `json:"img_name"`
	ImageUrl           string              `json:"img_url"`
	FullScreenImageUrl *FullScreenImageUrl `json:"full_screen_img_url"`
	Mode               int                 `json:"mode"`
	LogoConfig         *LogoConfig         `json:"logo_config"`
	InitialPushTime    int64               `json:"initial_push_time"`
	KeepNewDays        int32               `json:"keep_new_days"`
	CategoryIDs        []int64             `json:"category_ids"`
}

// ImageMap 配置物料信息Map，以id为key
type ImageMap map[int64]*SplashScreenImageForGateway

// GatewayConfig 给网关的配置
type GatewayConfig struct {
	ImageMap              ImageMap              `json:"img_map"`
	DefaultConfig         *SplashScreenConfig   `json:"default_config"`
	BaseDefaultConfig     *SplashScreenConfig   `json:"base_default_config"`
	SelectConfig          *SplashScreenConfig   `json:"select_config"`
	PrepareDefaultConfigs []*SplashScreenConfig `json:"prepare_default_configs"`
	PrepareSelectConfigs  []*SplashScreenConfig `json:"prepare_select_configs"`
	Categories            []*Category           `json:"categories"`
}

type Pager struct {
	Total int32 `json:"total"`
	Ps    int32 `json:"size"`
	Pn    int32 `json:"num"`
}

type SplashConfigListWithPager struct {
	List  []*SplashScreenConfig `json:"list"`
	Pager *Pager                `json:"page"`
}

// 闪屏配置
type SplashScreenConfig struct {
	ID            int64      `json:"id" form:"id" gorm:"column:id"`
	STime         xtime.Time `json:"stime" form:"stime" gorm:"column:stime"`
	ETime         xtime.Time `json:"etime" form:"etime" gorm:"column:etime"`
	IsImmediately int        `json:"immediately" form:"immediately" gorm:"column:immediately"`
	// state 表示状态：
	// 0、待通过
	// 1、待生效
	// 2、已失效
	// 3、生效中
	State          api.SplashScreenConfigState_Enum       `json:"state" form:"-" gorm:"-"`
	AuditState     api.SplashScreenConfigAuditStatus_Enum `json:"audit_state" form:"-" gorm:"column:audit_state"`
	ShowMode       int                                    `json:"show_mode" form:"show_mode" gorm:"column:show_mode"`
	ForceShowTimes int32                                  `json:"force_show_times" form:"force_show_times" gorm:"column:force_show_times"`
	ConfigJson     string                                 `json:"config_json" form:"config_json" gorm:"column:config_json"`
	IsDeleted      int                                    `json:"-" form:"-" gorm:"column:is_deleted"`
	CUser          string                                 `json:"c_user" form:"-" gorm:"column:c_user"`
	MUser          string                                 `json:"m_user" form:"-" gorm:"column:m_user"`
	CTime          xtime.Time                             `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime          xtime.Time                             `json:"mtime" form:"-" gorm:"column:mtime"`
}

func (*SplashScreenConfig) TableName() string {
	return "splash_screen"
}

// 配置详情
type ConfigDetail struct {
	Position int   `json:"position"`
	Rate     int   `json:"rate"`
	ImgId    int64 `json:"img_id"`
	Sort     int64 `json:"sort,omitempty"`
}
