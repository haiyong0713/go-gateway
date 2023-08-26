package show

import "go-common/library/time"

const (
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone = int8(1)
	// PlatIPad is int8 for ipad.
	PlatIPad = int8(2)
	// PlatWPhone is int8 for wphone.
	PlatWPhone = int8(3)
	// PlatAndroidG is int8 for Android Global.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// PlatAndroidB is int8 for android_b
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for iphone_b
	PlatIPhoneB = int8(10)
	// PlatIPadHD is int8 for ipadHD.
	PlatIPadHD = int8(20)
	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)

	SModuleNormal   = 1
	SModuleTypeMine = 1
)

// SidebarLimit
type SidebarLimit struct {
	SID        int64  `gorm:"column:s_id" json:"s_id,omitempty"`
	Conditions string `gorm:"column:conditions" json:"conditions"`
	Build      int64  `gorm:"column:build" json:"build"`
	Name       string `gorm:"column:name" json:"name,omitempty"`
	Plat       int32  `gorm:"column:plat" json:"plat,omitempty"`
	ModuleName string `gorm:"column:sm_name" json:"sm_name,omitempty"`
}

// TableName .
func (a SidebarLimit) TableName() string {
	return "sidebar_limit"
}

// SidebarWithLimit
type SidebarWithLimit struct {
	SID        int64           `json:"sid"`
	Name       string          `json:"name"`
	Plat       int32           `json:"plat"`
	Limit      []*SidebarLimit `json:"limit"`
	ModuleName string          `json:"sm_name"`
}

// SidebarModule
type SidebarModule struct {
	ID              int64  `gorm:"column:id" json:"id"`
	MType           int32  `gorm:"column:mtype" json:"mtype"`
	Plat            int32  `gorm:"column:plat" json:"plat"`
	Name            string `gorm:"column:name" json:"name"`
	Title           string `gorm:"column:title" json:"title"`
	Style           int32  `gorm:"column:style" json:"style"`
	Rank            int32  `gorm:"column:rank" json:"rank"`
	ButtonName      string `gorm:"column:button_name" json:"button_name"`
	ButtonURL       string `gorm:"column:button_url" json:"button_url"`
	ButtonIcon      string `gorm:"column:button_icon" json:"button_icon"`
	ButtonStyle     int32  `gorm:"column:button_style" json:"button_style"`
	WhiteURL        string `gorm:"column:white_url" json:"white_url"`
	TitleColor      string `gorm:"column:title_color" json:"title_color"`
	Subtitle        string `gorm:"column:subtitle" json:"subtitle"`
	SubtitleURL     string `gorm:"column:subtitle_url" json:"subtitle_url"`
	SubtitleColor   string `gorm:"column:subtitle_color" json:"subtitle_color"`
	Background      string `gorm:"column:background" json:"background"`
	BackgroundColor string `gorm:"column:background_color" json:"background_color"`
	State           int32  `gorm:"column:state" json:"state"`
	AuditShow       int32  `gorm:"column:audit_show" json:"audit_show"`
	IsMng           int32  `gorm:"column:is_mng" json:"is_mng"`
	OpStyleType     int32  `gorm:"column:op_style_type" json:"op_style_type"`
	OpLoadCondition string `gorm:"column:op_load_condition" json:"op_load_condition"`
	AreaPolicy      int64  `gorm:"column:area_policy" json:"area_policy"`
	ShowPurposed    int32  `gorm:"column:show_purposed" json:"show_purposed"`
}

const (
	//ModuleStyle_Icon = 1
	//ModuleStyle_List = 2
	ModuleStyle_OP             = 3
	ModuleStyle_Launcher       = 4
	ModuleStyle_Classification = 5
	ModuleStyle_Recommend_tab  = 6
	ModuleStyle_Publish_bubble = 8
)

// SidebarModuleSimple
type SidebarModuleSimple struct {
	ID              int64  `gorm:"column:id" json:"id"`
	Name            string `gorm:"column:name" json:"name"`
	MType           int32  `gorm:"column:mtype" json:"mtype"`
	Style           int32  `gorm:"column:style" json:"style"`
	OpLoadCondition string `gorm:"column:op_load_condition" json:"op_load_condition"`
	OpStyleType     int32  `gorm:"column:op_style_type" json:"op_style_type"`
	AreaPolicy      int64  `gorm:"column:area_policy" json:"area_policy"`
	ShowPurposed    int32  `gorm:"column:show_purposed" json:"show_purposed"`
}

// SaveModuleParam
type SaveModuleParam struct {
	ID              int64  `form:"id"`
	MType           int32  `form:"mtype" validate:"min=1"`
	Plat            int32  `form:"plat" validate:"min=0"`
	Name            string `form:"name"  validate:"required"`
	Title           string `form:"title"`
	Style           int32  `form:"style" validate:"min=1"`
	Rank            int32  `form:"rank"`
	ButtonName      string `form:"button_name"`
	ButtonURL       string `form:"button_url"`
	ButtonIcon      string `form:"button_icon"`
	ButtonStyle     int32  `form:"button_style"`
	WhiteURL        string `form:"white_url"`
	TitleColor      string `form:"title_color"`
	Subtitle        string `form:"subtitle"`
	SubtitleURL     string `form:"subtitle_url"`
	SubtitleColor   string `form:"subtitle_color"`
	Background      string `form:"background"`
	BackgroundColor string `form:"background_color"`
	AuditShow       int32  `form:"audit_show"`
	IsMng           int32  `form:"is_mng"`
	OpStyleType     int32  `form:"op_style_type"`
	OpLoadCondition string `form:"op_load_condition"`
	AreaPolicy      int64  `form:"area_policy"`
	ShowPurposed    int32  `form:"show_purposed"`
}

// TableName .
func (a SidebarModule) TableName() string {
	return "sidebar_module"
}

// SidebarLimitORM
type SidebarLimitORM struct {
	SID        int64  `gorm:"column:s_id" json:"s_id,omitempty"`
	Conditions string `gorm:"column:conditions" json:"conditions"`
	Build      int64  `gorm:"column:build" json:"build"`
}

// TableName .
func (a *SidebarLimitORM) TableName() string {
	return "sidebar_limit"
}

type SidebarORM struct {
	ID         int64     `gorm:"column:id" json:"id" form:"id"`
	Plat       int32     `gorm:"column:plat" json:"plat" form:"plat"`
	Module     int64     `gorm:"column:module" json:"module" form:"module"`
	Name       string    `gorm:"column:name" json:"name" form:"name"`
	Logo       string    `gorm:"column:logo" json:"logo" form:"logo"`
	Param      string    `gorm:"column:param" json:"param" form:"param"`
	Rank       int64     `gorm:"column:rank" json:"rank" form:"rank"`
	OnlineTime time.Time `gorm:"column:online_time" json:"online_time" form:"online_time"`
	Tip        int32     `gorm:"column:tip" json:"tip"`
	State      int32     `gorm:"state" json:"state"`
	LogoWhite  string    `gorm:"column:logo_white" json:"logo_white" form:"logo_white"`
	NeedLogin  int32     `gorm:"column:need_login" json:"need_login" form:"need_login"`
	WhiteURL   string    `gorm:"column:white_url" json:"white_url" form:"white_url"`
	//Menu            int32     `gorm:"column:menu" json:"menu"`
	LogoSelected    string    `gorm:"column:logo_selected" json:"logo_selected" form:"logo_selected"`
	TabID           string    `gorm:"column:tab_id" json:"tab_id" form:"tab_id"`
	Red             string    `gorm:"column:red_dot_url" json:"red_dot_url" form:"red_dot_url"`
	Lang            int64     `gorm:"column:lang_id" json:"lang_id" form:"lang_id"`
	GlobalRed       string    `gorm:"column:global_red_dot" json:"global_red_dot"`
	RedLimit        int64     `gorm:"column:red_dot_limit" json:"red_dot_limit"`
	Animate         string    `gorm:"column:animate" json:"animate" form:"animate"`
	WhiteURLShow    int64     `gorm:"column:white_url_show" json:"white_url_show" form:"white_url_show"`
	ShowPurposed    int32     `gorm:"column:show_purposed" json:"show_purposed" form:"show_purposed"`
	AreaPolicy      int64     `gorm:"column:area_policy" json:"area_policy" form:"area_policy"`
	RedDotForNew    int32     `gorm:"column:red_dot_for_new" json:"red_dot_for_new" form:"red_dot_for_new"`
	OpLoadCondition string    `gorm:"column:op_load_condition" json:"op_load_condition" form:"op_load_condition"`
	OfflineTime     time.Time `gorm:"column:offline_time" json:"offline_time" form:"offline_time"`
	OpTitle         string    `gorm:"column:op_title" json:"op_title" form:"op_title"`
	OpSubTitle      string    `gorm:"column:op_sub_title" json:"op_sub_title" form:"op_sub_title"`
	OpTitleIcon     string    `gorm:"column:op_title_icon" json:"op_title_icon" form:"op_title_icon"`
	OpLinkText      string    `gorm:"column:op_link_text" json:"op_link_text" form:"op_link_text"`
	OpLinkIcon      string    `gorm:"column:op_link_icon" json:"op_link_icon" form:"op_link_icon"`
	OpLinkType      int32     `gorm:"column:op_link_type" json:"op_link_type" form:"op_link_type"`
	OpFansLimit     int64     `gorm:"column:op_fans_limit" json:"op_fans_limit" form:"op_fans_limit"`
	GrayToken       string    `gorm:"column:gray_token" json:"gray_token" form:"gray_token"`
	DynamicConfUrl  string    `gorm:"column:dynamic_conf_url" json:"dynamic_conf_url" form:"dynamic_conf_url"` // 动态配置读取url
	// 运营条颜色信息
	OpTitleColor         string `gorm:"column:op_title_color" json:"op_title_color" form:"op_title_color"`
	OpBackgroundColor    string `gorm:"column:op_background_color" json:"op_background_color" form:"op_background_color"`
	OpLinkContainerColor string `gorm:"column:op_link_container_color" json:"op_link_container_color" form:"op_link_container_color"`
	TusValue             string `gorm:"column:tus_value" json:"tus_value" form:"tus_value"`
}

// SidebarORM
type SidebarEntity struct {
	SidebarORM
	Limit             []*SidebarLimit `json:"limit"`
	LimitStr          string          `form:"limit" json:"limit_str,omitempty"`
	ModuleStyle       int32           `form:"style" json:"style,omitempty"`
	ModuleOpStyleType int32           `form:"op_style_type" json:"op_style_type,omitempty"`
}

func (sidebarItem *SidebarORM) TableName() string {
	return "sidebar"
}

const (
	SidebarORM_State_Deleted = -1
	SidebarORM_State_Banned  = 0
	SidebarORM_State_Normal  = 1
)

// 查询二级模块list用
type ModuleItemListReq struct {
	Plat   int32 `form:"plat"`
	Build  int64 `form:"build"`
	Module int64 `form:"module"`
	Lang   int64 `form:"lang_id"`
}
