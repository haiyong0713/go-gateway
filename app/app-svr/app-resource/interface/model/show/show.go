package show

import (
	"strconv"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/tab"
	resource "go-gateway/app/app-svr/resource/service/model"

	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
	resourceBapi "git.bilibili.co/bapis/bapis-go/resource/service"
)

const (
	BusGame             = "game"
	ClearVer            = "clear_ver"
	TopActivityFission  = 1
	TopActivityMng      = 2
	_attrTabExtIsFollow = 3
)

type Show struct {
	Tab     []*Tab   `json:"tab,omitempty"`
	Top     []*Tab   `json:"top,omitempty"`
	Bottom  []*Tab   `json:"bottom,omitempty"`
	Config  *Config  `json:"config,omitempty"`
	TopMore []*Tab   `json:"top_more,omitempty"`
	TopLeft *TopLeft `json:"top_left,omitempty"`
}

type TopLeft struct {
	//实验情况 0-在实验组 1-不在实验组
	Exp int64 `json:"exp"`
	//头像特殊标记
	HeadTag string `json:"head_tag"`
	//跳转链接
	Url string `json:"url"`
	//落地页 1-【我的】页 2-story
	Goto int64 `json:"goto"`
	//看一看的背景图
	StoryBackgroundImage string `json:"story_background_image"`
	//看一看的前景图
	StoryForegroundImage string `json:"story_foreground_image"`
	//听一听的背景图
	ListenBackgroundImage string `json:"listen_background_image"`
	//听一听的前景图
	ListenForegroundImage string `json:"listen_foreground_image"`
}

type TabsFilter []*Tab

func (ts TabsFilter) Filter(fn func(tab *Tab) bool) []*Tab {
	var tmpTabs []*Tab
	for _, v := range ts {
		if fn(v) {
			continue
		}
		tmpTabs = append(tmpTabs, v)
	}
	return tmpTabs
}

type Tab struct {
	ID              int64        `json:"id,omitempty"`
	Icon            string       `json:"icon,omitempty"`
	IconSelected    string       `json:"icon_selected,omitempty"`
	Name            string       `json:"name,omitempty"`
	URI             string       `json:"uri,omitempty"`
	TabID           string       `json:"tab_id,omitempty"`
	Color           string       `json:"color,omitempty"`
	Pos             int          `json:"pos,omitempty"`
	DefaultSelected int          `json:"default_selected,omitempty"`
	Module          int          `json:"-"`
	ModuleStr       string       `json:"-"`
	Plat            int8         `json:"-"`
	Group           string       `json:"-"`
	Language        string       `json:"-"`
	Red             string       `json:"-"`
	Animate         string       `json:"-"`
	RedDot          *RedDot      `json:"red_dot,omitempty"`
	AnimateIcon     *AnimateIcon `json:"animate_icon,omitempty"`
	Extension       *Extension   `json:"extension,omitempty"`
	Area            string       `json:"-"`
	ShowPurposed    int32        `json:"-"`
	AreaPolicy      int64        `json:"-"`
	WhiteURL        string       `json:"-"`
	WhiteURLShow    int32        `json:"-"`
	//对话框items
	DialogItems []*DialogItems `json:"dialog_items,omitempty"`
	Type        int64          `json:"type,omitempty"`
	ModuleStyle int32          `json:"-"`
	//异形图配置
	OpIcon        *OpIcon          `json:"-"`
	PublishBubble []*PublishBubble `json:"publish_bubble,omitempty"`
}

type PublishBubble struct {
	ID   int64  `json:"id"`
	Url  string `json:"url"`
	Icon string `json:"icon"`
}

type DialogItems struct {
	ID     int64   `json:"id,omitempty"`
	Name   string  `json:"name,omitempty"`
	Icon   string  `json:"icon,omitempty"`
	Uri    string  `json:"uri,omitempty"`
	OpIcon *OpIcon `json:"op_icon,omitempty"`
}

type OpIcon struct {
	Icon string `json:"icon,omitempty"`
	Id   int64  `json:"id,omitempty"`
	//图片失效时间
	FTime int64 `json:"ftime,omitempty"`
	//图片生效时间
	ETime int64 `json:"etime,omitempty"`
}

// Extension .
type Extension struct {
	// 未激活状态资源
	InactiveIcon string `json:"inactive_icon,omitempty"`
	// 未激活状态动画类型
	Inactive int64 `json:"inactive,omitempty"`
	// 未激活状态动画控制
	InactiveType int64 `json:"inactive_type,omitempty"`
	// 激活状态资源
	ActiveIcon string `json:"active_icon,omitempty"`
	// 激活状态动画类型
	Active int64 `json:"active,omitempty"`
	// 激活状态动画控制
	ActiveType int64 `json:"active_type,omitempty"`
	// 背景色
	BgColor string `json:"bg_color,omitempty"`
	// 文本高亮色
	FontColor string `json:"font_color,omitempty"`
	// 状态栏颜色
	BarColor int64 `json:"bar_color,omitempty"`
	// tab头部色值
	TabTopColor string `json:"tab_top_color,omitempty"`
	// tab底部色值
	TabBottomColor string `json:"tab_bottom_color,omitempty"`
	// tab中间颜色值
	TabMiddleColor string `json:"tab_middle_color,omitempty"`
	// 背景图片1
	BgImage1 string `json:"bg_image_1,omitempty"`
	// 背景图片2
	BgImage2 string `json:"bg_image_2,omitempty"`
	// tab运营资源点击
	Click *Click `json:"click,omitempty"`
	// 开关配置，true：开启跟随业务方模式 false:不开启
	IsFollowBusiness bool `json:"is_follow_business,omitempty"`
}

type Click struct {
	Ver  string `json:"ver,omitempty"`
	ID   int64  `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type Limit struct {
	ID        int64  `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
}

func (ext *Extension) BuildExt(menuExt *resourceBapi.TabExt) {
	ext.ActiveType = menuExt.ActiveType
	ext.Active = menuExt.Active
	ext.ActiveIcon = menuExt.ActiveIcon
	ext.InactiveType = menuExt.InactiveType
	ext.InactiveIcon = menuExt.InactiveIcon
	ext.Inactive = menuExt.Inactive
	if menuExt.Click != nil {
		ext.Click = &Click{ID: menuExt.Click.Id, Ver: menuExt.Click.Ver, Type: menuExt.Click.Type}
	}
	// 低版本需要bgcolor
	ext.BgColor = menuExt.TabBottomColor
	ext.FontColor = menuExt.FontColor
	ext.BarColor = menuExt.BarColor
	ext.TabBottomColor = menuExt.TabBottomColor
	ext.TabMiddleColor = menuExt.TabMiddleColor
	ext.TabTopColor = menuExt.TabTopColor
	ext.BgImage1 = menuExt.BgImage1
	ext.BgImage2 = menuExt.BgImage2
	ext.IsFollowBusiness = attrVal(menuExt.Attribute, _attrTabExtIsFollow)
}

func attrVal(attribute int64, bit uint) bool {
	return (attribute>>bit)&int64(1) == 1
}

type Config struct {
	NoLoginAvatar     string `json:"no_login_avatar,omitempty"`
	NoLoginAvatarType int    `json:"no_login_avatar_type"`
	PopupStyle        int8   `json:"popup_style"`
}

func (t *Tab) TabChange(rsb *resource.SideBar, abtest map[string]string, defaultTab map[string]*Tab) (ok bool) {
	var (
		_top    = 10
		_tab    = 8
		_bottom = 9
	)
	t.ID = rsb.ID
	t.Icon = rsb.Logo
	t.IconSelected = rsb.LogoSelected
	t.Name = rsb.Name
	t.URI = rsb.Param
	t.Module = rsb.Module
	t.Plat = rsb.Plat
	t.Language = rsb.Language
	t.Red = rsb.Red
	t.Animate = rsb.Animate
	t.Area = rsb.Area
	t.ShowPurposed = rsb.ShowPurposed
	t.AreaPolicy = rsb.AreaPolicy
	t.WhiteURL = rsb.WhiteURL
	t.WhiteURLShow = rsb.WhiteURLShow
	switch t.Module {
	case _top:
		t.ModuleStr = model.ModuleTop
	case _tab:
		t.ModuleStr = model.ModuleTab
		t.Icon = ""
		t.IconSelected = ""
	case _bottom:
		t.ModuleStr = model.ModuleBottom
	default:
		return false
	}
	if len(abtest) > 0 {
		if groups, ok := abtest[t.URI]; ok {
			t.Group = groups
		}
	}
	if len(defaultTab) > 0 {
		if dt, ok := defaultTab[t.URI]; ok && dt != nil {
			t.DefaultSelected = dt.DefaultSelected
			t.TabID = dt.TabID
		}
		if rsb.TabID != "" {
			t.TabID = rsb.TabID
		}
	}
	return true
}

func (t *Tab) TabMenuChange(m *tab.Menu) {
	t.TabID = strconv.FormatInt(m.TabID, 10)
	t.Name = m.Name
	t.Color = m.Color
	t.ID = m.ID
	t.ModuleStr = "tab"
	if m.CType == tab.ActPageType {
		t.URI = model.FillURI(model.GotoActPageTab, strconv.FormatInt(t.ID, 10), nil)
	} else {
		t.URI = model.FillURI(model.GotoPegasusTab, strconv.FormatInt(t.ID, 10), model.PegasusHandler(m))
	}
}

// SectionURL is
type SectionURL struct {
	ID  int64
	URL string
}

type Red struct {
	RedDot bool   `json:"red_dot,omitempty"`
	Type   int8   `json:"type"`
	Number int64  `json:"number,omitempty"`
	Icon   string `json:"icon"`
}

type RedDot struct {
	Type   int8  `json:"type"`
	Number int64 `json:"number,omitempty"`
}

type AnimateIcon struct {
	Icon string `json:"icon,omitempty"`
	Json string `json:"json,omitempty"`
}

type TabBubble struct {
	Key   string     `json:"key,omitempty"`
	ID    int64      `json:"id,omitempty"`
	Title string     `json:"title,omitempty"`
	Cover string     `json:"cover,omitempty"`
	Param string     `json:"param,omitempty"`
	URI   string     `json:"uri,omitempty"`
	STime xtime.Time `json:"stime,omitempty"`
	ETime xtime.Time `json:"etime,omitempty"`
}

// SkinReply .
type SkinReply struct {
	// 运营类皮肤
	CommonEquip *SkinConf `json:"common_equip,omitempty"`
	// 装扮皮肤
	UserEquip *SkinConf `json:"user_equip,omitempty"`
	// 颜色主题
	SkinColors []*SkinColor `json:"skin_colors,omitempty"`
	// 加载动画
	LoadEquip *LoadEquip `json:"load_equip,omitempty"`
}

// LoadEquip .
type LoadEquip struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Ver        int64  `json:"ver,omitempty"`
	LoadingUrl string `json:"loading_url,omitempty"`
}

type SkinConf struct {
	ID         int64                 `json:"id,omitempty"`
	Name       string                `json:"name,omitempty"`
	Preview    string                `json:"preview,omitempty"`
	Ver        int64                 `json:"ver,omitempty"`
	PackageUrl string                `json:"package_url,omitempty"`
	PackageMd5 string                `json:"package_md5,omitempty"`
	Data       *garbmdl.UserSkinData `json:"data,omitempty"`
	Conf       *SingleConf           `json:"conf,omitempty"`
}

// SingleConf .
type SingleConf struct {
	Alias     string     `json:"alias,omitempty"`
	Attribute int64      `json:"attribute,omitempty"`
	STime     xtime.Time `json:"stime,omitempty"`
	ETime     xtime.Time `json:"etime,omitempty"`
}

// SkinColor .
type SkinColor struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	IsFree    bool   `json:"is_free,omitempty"`
	Price     int64  `json:"price,omitempty"`
	IsBought  bool   `json:"is_bought,omitempty"`
	Status    int64  `json:"status,omitempty"`
	BuyTime   int64  `json:"buy_time,omitempty"`
	DueTime   int64  `json:"due_time,omitempty"`
	ColorName string `json:"color_name,omitempty"`
}

// TopActivityReply
type TopActivityReq struct {
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Build    int64  `form:"build"`
	Platform string `form:"platform"`
}

// TopActivityReply
type TopActivityReply struct {
	Online *TopOnline `json:"online"`
	Hash   string     `json:"hash"`
}

type TopOnline struct {
	Icon     string      `json:"icon,omitempty"`
	Uri      string      `json:"uri,omitempty"`
	UniqueID string      `json:"unique_id,omitempty"`
	Interval int64       `json:"interval,omitempty"`
	Animate  *TopAnimate `json:"animate,omitempty"`
	RedDot   *RedDot     `json:"red_dot,omitempty"`
	Type     int32       `json:"type"`
	Name     string      `json:"name,omitempty"`
}

type TopAnimate struct {
	Icon string `json:"icon,omitempty"`
	Json string `json:"json,omitempty"`
	Svg  string `json:"svg,omitempty"`
	Loop int32  `json:"loop,omitempty"`
}

type PopUp struct {
	*PopUpReport
	AutoClose     bool   `json:"auto_close"`
	AutoCloseTime int    `json:"auto_close_time"`
	ReportData    string `json:"report_data"`
}

type PopUpReport struct {
	ID     int64  `json:"id"`
	Pic    string `json:"pic"`
	Detail string `json:"detail"`
	Link   string `json:"link"`
}

type TabsV2Params struct {
	Channel       string `form:"channel"`
	MobiApp       string `form:"mobi_app"`
	Build         int64  `form:"build"`
	Device        string `form:"device"`
	Lang          string `form:"lang"`
	Platform      string `form:"platform"`
	TeenagersMode int64  `form:"teenagers_mode"`
	LessonsMode   int64  `form:"lessons_mode"`
	Slocale       string `form:"s_locale"`
	Clocale       string `form:"c_locale"`
	DisableRcmd   int64  `form:"disable_rcmd"`
}

type VIVOPopularBadgeReply struct {
	HotUpdateInterval     int64  `json:"hot_update_interval"`
	KeywordUpdateInterval int64  `json:"keyword_update_interval"`
	Slogan                string `json:"slogan"`
	SearchDefaultWord     string `json:"search_default_word"`
}
