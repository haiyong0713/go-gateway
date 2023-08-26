package section

import xtime "go-common/library/time"

const (
	mtype_Home           = 2
	recommendTopTabStyle = 6
)

// ModuleInfo
type ModuleInfo struct {
	ID              int64  `json:"id"`
	Plat            int32  `json:"plat"`
	Title           string `json:"title,omitempty"`
	Style           int32  `json:"style,omitempty"`
	ButtonName      string `json:"button_name,omitempty"`
	ButtonURL       string `json:"button_url,omitempty"`
	ButtonIcon      string `json:"button_icon,omitempty"`
	ButtonStyle     int32  `json:"button_style,omitempty"`
	WhiteURL        string `json:"white_url,omitempty"`
	TitleColor      string `json:"title_color,omitempty"`
	Subtitle        string `json:"subtitle,omitempty"`
	SubtitleURL     string `json:"subtitle_url,omitempty"`
	SubtitleColor   string `json:"subtitle_color,omitempty"`
	Background      string `json:"background,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	MType           int32  `json:"mtype"`
	AuditShow       int32  `json:"audit_show"`
	IsMng           int32  `json:"is_mng"`
	OpStyleType     int32  `json:"op_style_type"`
	OpLoadCondition string `json:"op_load_condition"`
}

func (m *ModuleInfo) IsRecommendTab() bool {
	return m.MType == mtype_Home && m.Style == recommendTopTabStyle
}

type SectionReq struct {
	Plat       int32
	Build      int32
	Mid        int64
	Lang       string
	Channel    string
	Ip         string
	IsUploader bool
	IsLiveHost bool
	FansCount  int64
	Buvid      string
}

type DynamicConf struct {
	Param        string `json:"param"`
	Name         string `json:"name"`
	DefaultIcon  int64  `json:"default_icon"`
	SelectedIcon int64  `json:"selected_icon"`
}

// SideBar for side bar
type SideBar struct {
	ID           int64      `json:"id,omitempty"`
	Tip          int        `json:"tip,omitempty"`
	Rank         int        `json:"rank,omitempty"`
	Logo         string     `json:"logo,omitempty"`
	LogoWhite    string     `json:"logo_white,omitempty"`
	Name         string     `json:"name,omitempty"`
	Param        string     `json:"param,omitempty"`
	Module       int        `json:"module,omitempty"`
	Plat         int8       `json:"-"`
	OnlineTime   xtime.Time `json:"online_time"`
	NeedLogin    int8       `json:"need_login,omitempty"`
	WhiteURL     string     `json:"white_url,omitempty"`
	Menu         int8       `json:"menu,omitempty"`
	LogoSelected string     `json:"logo_selected,omitempty"`
	TabID        string     `json:"tab_id,omitempty"`
	Red          string     `json:"red_dot_url,omitempty"`
	Language     string     `json:"language,omitempty"`
	LanguageID   int64      `json:"language_id,omitempty"`
	GlobalRed    int8       `json:"global_red_dot,omitempty"`
	RedLimit     int64      `json:"red_dot_limit,omitempty"`
	Animate      string     `json:"animate,omitempty"`
	WhiteURLShow int32      `json:"white_url_show"`
	Area         string     `json:"area"`
	ShowPurposed int32      `json:"show_purposed"`
	AreaPolicy   int64      `json:"area_policy"`
	RedDotForNew bool       `json:"red_dot_for_new"`
	// APP管理-我的页新增运营位配置
	OpLoadCondition string     `json:"op_load_condition"`
	OfflineTime     xtime.Time `json:"offline_time"`
	OpTitle         string     `json:"op_title"`
	OpSubTitle      string     `json:"op_sub_title"`
	OpTitleIcon     string     `json:"op_title_icon"`
	OpLinkText      string     `json:"op_link_text"`
	OpLinkIcon      string     `json:"op_link_icon"`
	OpLinkType      int32      `json:"op_link_type"`
	OpFansLimit     int64      `json:"op_fans_limit"`
	GrayToken       string     `json:"gray_token"`
	DynamicConfUrl  string     `json:"dynamic_conf_url"`
	//运营条
	OpTitleColor         string `json:"op_title_color"`
	OpBackgroundColor    string `json:"op_background_color"`
	OpLinkContainerColor string `json:"op_link_container_color"`
	TusValue             string `json:"tus_value"`
}

// SideBarLimit side bar limit
type SideBarLimit struct {
	SideBarID int64  `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
}

func ValidBuild(srcBuild, cfgBuild int64, cfgCond string) bool {
	if cfgBuild == 0 && cfgCond != "" { //配置0表示对所有版本生效
		return true
	}
	switch cfgCond {
	case "gt":
		if cfgBuild < srcBuild {
			return true
		}
	case "lt":
		if cfgBuild > srcBuild {
			return true
		}
	case "eq":
		if cfgBuild == srcBuild {
			return true
		}
	case "ne":
		if cfgBuild != srcBuild {
			return true
		}
	default:
		return false
	}
	return false
}
