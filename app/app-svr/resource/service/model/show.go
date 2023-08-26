package model

import (
	"encoding/json"
	"fmt"
	"strconv"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
)

var (
	IsUp = fmt.Errorf("IsUp")
)

const (
	mtype_Home           = 2
	recommendTopTabStyle = 6
)

// Card audio card struct
type Card struct {
	ID        int    `json:"-"`
	Tab       int    `json:"-"`
	RegionID  int    `json:"-"`
	Type      int    `json:"-"`
	Title     string `json:"-"`
	Cover     string `json:"-"`
	Rtype     int    `json:"-"`
	Rvalue    string `json:"-"`
	PlatVer   string `json:"-"`
	Plat      int8   `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
	TypeStr   string `json:"-"`
	Goto      string `json:"-"`
	Param     string `json:"-"`
	URI       string `json:"-"`
	Desc      string `json:"-"`
	TagID     int    `json:"-"`
}

// Content audio content struct
type Content struct {
	ID     int    `json:"-"`
	Module int    `json:"-"`
	RecID  int    `json:"-"`
	Type   int8   `json:"-"`
	Value  string `json:"-"`
	Title  string `json:"-"`
	TagID  int    `json:"-"`
}

// PlatLimit audio plat limit struct
type PlatLimit struct {
	Plat      int8   `json:"plat"`
	Build     int    `json:"build"`
	Condition string `json:"conditions"`
}

// ShowItem audio show item struct
type ShowItem struct {
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	URI    string `json:"uri"`
	NewURI string `json:"-"`
	Param  string `json:"param"`
	Goto   string `json:"goto"`
	// up
	Mid            int64           `json:"mid,omitempty"`
	Name           string          `json:"name,omitempty"`
	Face           string          `json:"face,omitempty"`
	Follower       int             `json:"follower,omitempty"`
	Attribute      int             `json:"attribute,omitempty"`
	OfficialVerify *OfficialVerify `json:"official_verify,omitempty"`
	// stat
	Play    int `json:"play,omitempty"`
	Danmaku int `json:"danmaku,omitempty"`
	Reply   int `json:"reply,omitempty"`
	Fav     int `json:"favourite,omitempty"`
	// movie and bangumi badge
	Status    int8   `json:"status,omitempty"`
	CoverMark string `json:"cover_mark,omitempty"`
	// ranking
	Pts      int64       `json:"pts,omitempty"`
	Children []*ShowItem `json:"children,omitempty"`
	// av
	PubDate xtime.Time `json:"pubdate"`
	// av stat
	Duration int64 `json:"duration,omitempty"`
	// region
	Rid   int    `json:"rid,omitempty"`
	Rname string `json:"rname,omitempty"`
	Reid  int    `json:"reid,omitempty"`
	//new manager
	Desc  string `json:"desc,omitempty"`
	Stime string `json:"stime,omitempty"`
	Etime string `json:"etime,omitempty"`
	Like  int    `json:"like,omitempty"`
}

// OfficialVerify audio verify struct
type OfficialVerify struct {
	Type int    `json:"type"`
	Desc string `json:"desc"`
}

// Head audio struct
type Head struct {
	CardID    int         `json:"card_id,omitempty"`
	Title     string      `json:"title,omitempty"`
	Cover     string      `json:"cover,omitempty"`
	Type      string      `json:"type,omitempty"`
	Date      int64       `json:"date,omitempty"`
	Plat      int8        `json:"-"`
	Build     int         `json:"-"`
	Condition string      `json:"-"`
	URI       string      `json:"uri,omitempty"`
	Goto      string      `json:"goto,omitempty"`
	Param     string      `json:"param,omitempty"`
	Body      []*ShowItem `json:"body,omitempty"`
}

// CardPlatChange audio card change plat
func (c *Card) CardPlatChange() (platlinits []*PlatLimit) {
	platlinits = platJSONChange(c.PlatVer)
	return
}

// platJSONChange json change plat build condition
func platJSONChange(jsonStr string) (platlinits []*PlatLimit) {
	var tmp []struct {
		Plat      string `json:"plat"`
		Build     string `json:"build"`
		Condition string `json:"conditions"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &tmp); err == nil {
		for _, limit := range tmp {
			platlinit := &PlatLimit{}
			switch limit.Plat {
			case "0": // resource android
				platlinit.Plat = PlatAndroid
			case "1": // resource iphone
				platlinit.Plat = PlatIPhone
			case "2": // resource pad
				platlinit.Plat = PlatIPad
			case "5": // resource iphone_i
				platlinit.Plat = PlatIPhoneI
			case "8": // resource android_i
				platlinit.Plat = PlatAndroidI
			}
			platlinit.Build, _ = strconv.Atoi(limit.Build)
			platlinit.Condition = limit.Condition
			platlinits = append(platlinits, platlinit)
		}
	}
	return
}

// FromArchivePB from archive archive.
func (i *ShowItem) FromArchivePB(a *api.Arc) {
	i.Title = a.Title
	i.Cover = a.Pic
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.URI = FillURI(GotoAv, i.Param)
	i.Goto = GotoAv
	i.Play = int(a.Stat.View)
	i.Danmaku = int(a.Stat.Danmaku)
	i.Name = a.Author.Name
	i.Reply = int(a.Stat.Reply)
	i.Fav = int(a.Stat.Fav)
	i.PubDate = a.PubDate
	i.Rid = int(a.TypeID)
	i.Rname = a.TypeName
	i.Duration = a.Duration
	i.Like = int(a.Stat.Like)
	if a.Access > 0 {
		i.Play = 0
	}
}

// FillBuildURI fill url by plat build
func (h *Head) FillBuildURI(plat int8, build int) {
	switch h.Goto {
	case GotoDaily:
		if (plat == PlatIPhone && build > 6670) || (plat == PlatAndroid && build > 5250000) {
			h.URI = "bilibili://pegasus/list/daily/" + h.Param
		}
	}
}

// SideBars for side bars
type SideBars struct {
	SideBar []*SideBar                `json:"sidebar,omitempty"`
	Limit   map[int64][]*SideBarLimit `json:"limit,omitempty"`
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
	Build        int        `json:"-"`
	Conditions   string     `json:"-"`
	OnlineTime   xtime.Time `json:"online_time"`
	NeedLogin    int8       `json:"need_login,omitempty"`
	WhiteURL     string     `json:"white_url,omitempty"`
	Menu         int8       `json:"menu,omitempty"`
	LogoSelected string     `json:"logo_selected,omitempty"`
	TabID        string     `json:"tab_id,omitempty"`
	Red          string     `json:"red_dot_url,omitempty"`
	Language     string     `json:"language,omitempty"`
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
}

// SideBarLimit side bar limit
type SideBarLimit struct {
	ID        int64  `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
}

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

const (
	//ModuleStyle_Icon = 1
	//ModuleStyle_List = 2
	ModuleStyle_OP = 3
	//ModuleStyle_Launcher = 4
	ModuleStyle_Classification = 5
)

type DynamicConf struct {
	Param        string `json:"param"`
	Name         string `json:"name"`
	DefaultIcon  int64  `json:"default_icon"`
	SelectedIcon int64  `json:"selected_icon"`
}

func InformationRegionChange(cardPosition int32) (resionID int32) {
	var regionID = map[int32]int32{
		1: 202, // 资讯-首页
		3: 203, // 资讯-热点
		4: 204, // 资讯-环球
		5: 205, // 资讯-社会
		6: 206, // 资讯-综合
	}
	return regionID[cardPosition]
}

const (
	InitTabExtKey  = "tab_ext_%d_%d"
	TabExtCacheKey = "tab_ext"
	ClearVer       = "clear_ver"
	//attribute bit
	AttrYes = int64(1)
	// attribute bit
	AttrBitImage          = uint(0)
	AttrBitColor          = uint(1)
	AttrBitBgImage        = uint(2)
	AttrBitFollowBusiness = uint(3)
)

// MenuTabExt .
type MenuExt struct {
	*TabExt
	Limit []*TabLimit
}

// MenuTabExt .
type TabExt struct {
	ID             int64      `json:"id"`
	Type           int64      `json:"type"`
	TabID          int64      `json:"tab_id"`
	Attribute      int64      `json:"attribute"`
	InactiveIcon   string     `json:"inactive_icon"`
	Inactive       int64      `json:"inactive"`
	InactiveType   int64      `json:"inactive_type"`
	ActiveIcon     string     `json:"active_icon"`
	Active         int64      `json:"active"`
	ActiveType     int64      `json:"active_type"`
	TabTopColor    string     `json:"tab_top_color"`
	TabMiddleColor string     `json:"tab_middle_color"`
	TabBottomColor string     `json:"tab_bottom_color"`
	BgImage1       string     `json:"bg_image1"`
	BgImage2       string     `json:"bg_image2"`
	FontColor      string     `json:"font_color"`
	BarColor       int64      `json:"bar_color"`
	State          int64      `json:"state"`
	Stime          xtime.Time `json:"stime"`
	Etime          xtime.Time `json:"etime"`
	Ver            string     `json:"ver"`
}

// AttrVal get attr val by bit.
func (a *TabExt) AttrVal(bit uint) int64 {
	return (a.Attribute >> bit) & int64(1)
}

// MenuTabLimit .
type TabLimit struct {
	ID         int64  `json:"id"`
	Type       int64  `json:"type"`
	TID        int64  `json:"t_id"`
	Plat       int64  `json:"plat"`
	Build      int64  `json:"build"`
	Conditions string `json:"conditions"`
	State      int64  `json:"state"`
}

// WebRcmdCard
type WebRcmdCard struct {
	ID      int64      `form:"id" gorm:"column:id" json:"id"`
	Type    int32      `form:"type" gorm:"column:type" json:"type"`
	Title   string     `form:"title" gorm:"column:title" json:"title"`
	Desc    string     `form:"desc" gorm:"column:desc" json:"desc"`
	Cover   string     `form:"cover" gorm:"column:cover" json:"cover"`
	ReType  int32      `form:"re_type" gorm:"column:re_type" json:"re_type"`
	ReValue string     `form:"re_value" gorm:"column:re_value" json:"re_value"`
	Person  string     `form:"person" gorm:"column:person" json:"person"`
	Deleted int64      `form:"deleted" gorm:"column:deleted" json:"deleted"`
	Ctime   xtime.Time `form:"ctime" gorm:"column:ctime" json:"ctime"`
	Mtime   xtime.Time `form:"mtime" gorm:"column:mtime" json:"mtime"`
	Image   string     `form:"image" gorm:"column:-" json:"image"`
}
