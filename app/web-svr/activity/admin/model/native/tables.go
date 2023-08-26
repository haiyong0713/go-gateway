package native

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/time"
)

const (
	Choice      = "PICKED"
	_commonType = 0
	_choiceType = 1
)

// tab
type TabData struct {
	Title         string    `json:"title" gorm:"column:title"`
	Stime         time.Time `json:"stime" gorm:"column:stime;default:0000-00-00 00:00:00" time_format:"2006-01-02 15:04:05"`
	Etime         time.Time `json:"etime" gorm:"column:etime;default:0000-00-00 00:00:00" time_format:"2006-01-02 15:04:05"`
	State         int8      `json:"state" gorm:"column:state"`
	Operator      string    `json:"operator" gorm:"column:operator"`
	BgType        int8      `json:"bg_type" gorm:"column:bg_type"`
	BgImg         string    `json:"bg_img" gorm:"column:bg_img"`
	BgColor       string    `json:"bg_color" gorm:"column:bg_color"`
	IconType      int8      `json:"icon_type" gorm:"column:icon_type"`
	ActiveColor   string    `json:"active_color" gorm:"column:active_color"`
	InactiveColor string    `json:"inactive_color" gorm:"column:inactive_color"`
}

type Tab struct {
	TabData
	ID      int32     `json:"id" gorm:"column:id;primary_key"`
	Creator string    `json:"creator" gorm:"column:creator"`
	Ctime   time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime   time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type TabList struct {
	Total int32  `json:"total"`
	List  []*Tab `json:"list"`
}

// tab module
type TabModuleData struct {
	Title       string `json:"title" gorm:"column:title"`
	TabId       int32  `json:"tab_id" gorm:"column:tab_id"`
	State       int8   `json:"state" gorm:"column:state"`
	Operator    string `json:"operator" gorm:"column:operator"`
	ActiveImg   string `json:"active_img" gorm:"column:active_img"`
	InactiveImg string `json:"inactive_img" gorm:"column:inactive_img"`
	Category    int8   `json:"category" gorm:"column:category"`
	Pid         int32  `json:"pid" gorm:"column:pid"`
	Url         string `json:"url" gorm:"column:url"`
	Rank        int8   `json:"rank" gorm:"column:rank"`
}

type TabModule struct {
	TabModuleData
	ID    int32     `json:"id" gorm:"column:id;primary_key"`
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// NatPage .
type NatPage struct {
	PageParam
	Ctime time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type NatPageExt struct {
	*NatPage
	PageDyn *PageDynRly `json:"page_dyn"`
}

type PageDynRly struct {
	*PageDyn
	TopicInfo []*TopicInfo `json:"topic_info,omitempty"`
}

type TopicInfo struct {
	Tid  int64  `json:"tid"`
	Name string `json:"name"`
}

// ModuleData .
type ModuleData struct {
	NatModule
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// NatModule .
type NatModule struct {
	ID         int64  `json:"id"             gorm:"column:id"`
	Category   int    `json:"category"       gorm:"column:category"`
	FID        int64  `json:"f_id"           gorm:"column:f_id"`
	NativeID   int64  `json:"native_id"      gorm:"column:native_id"`
	Rank       int    `json:"rank"           gorm:"column:rank"`
	Meta       string `json:"meta"           gorm:"column:meta"`
	Width      int    `json:"width"          gorm:"column:width"`
	Length     int    `json:"length"         gorm:"column:length"`
	Num        int    `json:"num"            gorm:"column:num"`
	Title      string `json:"title"          gorm:"column:title"`
	State      int    `json:"state"          gorm:"column:state"`
	DySort     int    `json:"dy_sort"        gorm:"column:dy_sort"`
	Ukey       string `json:"ukey"           gorm:"column:ukey"`
	Attribute  int64  `json:"attribute"      gorm:"column:attribute"`
	BgColor    string `json:"bg_color"       gorm:"column:bg_color"`
	TitleColor string `json:"title_color"    gorm:"column:title_color"`
	MoreColor  string `json:"more_color"     gorm:"column:more_color"`
	TName      string `json:"t_name"         gorm:"column:t_name"`
	CardStyle  int    `json:"card_style"     gorm:"column:card_style"`
	AvSort     int    `json:"av_sort"        gorm:"column:av_sort"`
	FontColor  string `json:"font_color"     gorm:"column:font_color"`
	Remark     string `json:"remark" gorm:"column:remark"`
	Caption    string `json:"caption"  gorm:"column:caption"`
	PType      int    `json:"p_type" gorm:"column:p_type"`
	Bar        string `json:"bar"  gorm:"column:bar"`
	LiveType   int    `json:"live_type" gorm:"column:live_type"`
	Stime      int64  `json:"stime" gorm:"column:stime"`
	Etime      int64  `json:"etime" gorm:"column:etime"`
	Colors     string `json:"colors" gorm:"column:colors"`       //json
	ConfSort   string `json:"conf_sort" gorm:"column:conf_sort"` //组件公共分类类配置
}

type Sort struct {
	MoreSort int64 `json:"more_sort"` //查看更多方式 0:跳转二级页面 1:浮层 3:下拉展示
	TimeSort int64 `json:"time_sort"` //精确到 0:年 1:月 2: 日 3:时 4:分 5:秒
	Axis     int64 `json:"axis"`      //时间轴节点类型 0:文本 1:时间节点
}

// 字段color的json格式
type Colors struct {
	DisplayColor        string `json:"display_color,omitempty"`          //文字标题字体色
	TitleBgColor        string `json:"title_bg_color,omitempty"`         //卡片背景色
	SelectColor         string `json:"select_color,omitempty"`           // 选中色
	NotSelectColor      string `json:"not_select_color,omitempty"`       // 未选中色
	PanelBgColor        string `json:"panel_bg_color,omitempty"`         // 展开面板背景色
	PanelSelectColor    string `json:"panel_select_color,omitempty"`     //展开面板选中色
	PanelNotSelectColor string `json:"panel_not_select_color,omitempty"` //展开面板未选中色
	TimelineColor       string `json:"timeline_color,omitempty"`         //时间轴色
}

type ConfLive struct {
	TitleImage   string `json:"title_image"` //组件图片标题
	Category     int    `json:"category"`
	FID          int64  `json:"f_id"`      //直播id
	TName        string `json:"t_name"`    //封面标题
	LiveType     int    `json:"live_type"` // 主播未开播，0:隐藏卡片 1:直播间
	Attribute    int64  `json:"attribute"`
	BgColor      string `json:"bg_color"`
	FontColor    string `json:"font_color"`
	Caption      string `json:"caption"` //组件文字标题
	Bar          string `json:"bar"`
	Stime        int64  `json:"stime"`
	Etime        int64  `json:"etime"`
	DisplayColor string `json:"display_color"` //文字标题-文字颜色
}

// PubData .
type PubData struct {
	ModuleID int64 `gorm:"column:module_id"`
	State    int   `gorm:"column:state"`
}

// Click .
type Click struct {
	PubData
	LeftX           int    `gorm:"column:left_x"`
	LeftY           int    `gorm:"column:left_y"`
	Width           int    `gorm:"column:width"`
	Length          int    `gorm:"column:length"`
	Link            string `gorm:"column:link"`
	Type            int    `gorm:"column:type"`
	ForeignID       int64  `gorm:"column:foreign_id"`
	UnfinishedImage string `gorm:"column:unfinished_image"`
	FinishedImage   string `gorm:"column:finished_image"`
	Tip             string `gorm:"column:tip"`
	OptionalImage   string `gorm:"column:optional_image"`
}

// Act .
type Act struct {
	PubData
	Rank   int   `gorm:"column:rank"`
	PageID int64 `gorm:"column:page_id"`
}

// DynamicExt .
type DynamicExt struct {
	PubData
	SelectType int64 `gorm:"column:select_type"`
	ClassType  int64 `gorm:"class_type"`
	ClassID    int64 `gorm:"class_id"`
}

// VideoExt .
type VideoExt struct {
	PubData
	SortType int64 `gorm:"column:sort_type"`
	Rank     int   `gorm:"column:rank"`
}

// ParticipationExt .
type ParticipationExt struct {
	PubData
	MType     int    `json:"m_type" gorm:"column:m_type"`
	Image     string `json:"image" gorm:"column:image"`
	Title     string `json:"title" gorm:"column:title"`
	Rank      int    `gorm:"column:rank"`
	ForeignID int    `json:"foreign_id" gorm:"column:foreign_id"`
	UpType    int    `json:"up_type" gorm:"column:up_type"`
}

// NatData .
type NatData struct {
	Structure *Structure `json:"structure"`
	Modules   *Modules   `json:"modules"`
}

// 以下为解析json的结构体
// JsonData .
type JsonData struct {
	Structure *Structure            `json:"structure"`
	Modules   map[string]*ModuleCfg `json:"modules"`
	PType     int                   `json:"p_type"`
	Base      *BaseJson             `json:"base"`
	MoreBases []*BaseJson           `json:"more_bases"`
}

type BaseJson struct {
	Children []string              `json:"children"`
	Modules  map[string]*ModuleCfg `json:"modules"`
	PType    int                   `json:"p_type"`
}

type ModuleCfg struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Config json.RawMessage `json:"config"`
}

// Structure .
type Structure struct {
	Root Root `json:"root"`
}

// Root .
type Root struct {
	ID       string   `json:"id"`
	Children []string `json:"children"`
}

// Modules
type Modules struct {
	NatClick   []*NatClick
	NatAct     []*NatAct
	NatDynamic []*NatDynamic
}

// Native click .
type NatClick struct {
	Type   string    `json:"type"`
	Config ConfClick `json:"config"`
}

type NatAct struct {
	Type   string   `json:"type"`
	Config *ConfAct `json:"config"`
}

type NatDynamic struct {
	Type   string       `json:"type"`
	Config *ConfDynamic `json:"config"`
}

type BannerImage struct {
	Image string `json:"image"`
	Bar   string `json:"bar"`
}

type PageBase struct {
	BgColor string `json:"bg_color"`
}

type HeadBase struct {
	BgColor   string `json:"bg_color"`
	Title     string `json:"title"`
	Attribute int64  `json:"attribute"`
}

type Navigation struct {
	BgColor         string `json:"bg_color"`
	FontColor       string `json:"font_color"`
	SelectBgColor   string `json:"select_bg_color"`   // `json:"title_color"`
	SelectFontColor string `json:"select_font_color"` //more_color
}

type InlineTab struct {
	BgColor     string       `json:"bg_color"`
	SelectColor string       `json:"select_color"`
	FontColor   string       `json:"font_color"`
	Attribute   int64        `json:"attribute"`
	IDs         []*InlineIDs `json:"ids"`
}

// InlineIDs .
type InlineIDs struct {
	ID int64 `json:"id"`
}

type Select struct {
	BgColor             string       `json:"bg_color"`
	SelectColor         string       `json:"select_color"`
	NotSelectColor      string       `json:"not_select_color"`
	PanelBgColor        string       `json:"panel_bg_color"`
	PanelSelectColor    string       `json:"panel_select_color"`
	PanelNotSelectColor string       `json:"panel_not_select_color"`
	Attribute           int64        `json:"attribute"`
	IDs                 []*SelectIDs `json:"ids"`
}

// SelectIDs .
type SelectIDs struct {
	ID int64 `json:"id"`
}

// CheckColor 导航组件色值 日间，夜间.
func (m *Navigation) CheckColor() bool {
	if m.BgColor != "" {
		bgColors := strings.Split(m.BgColor, ",")
		if len(bgColors) < 2 {
			return false
		}
	}
	if m.FontColor != "" {
		fontColors := strings.Split(m.FontColor, ",")
		if len(fontColors) < 2 {
			return false
		}
	}
	if m.SelectBgColor != "" {
		selectBgColors := strings.Split(m.SelectBgColor, ",")
		if len(selectBgColors) < 2 {
			return false
		}
	}
	if m.SelectFontColor != "" {
		selectFontColors := strings.Split(m.SelectFontColor, ",")
		if len(selectFontColors) < 2 {
			return false
		}
	}
	return true
}

type StatementModule struct {
	Remark     string `json:"remark"`
	Bar        string `json:"bar"`
	BgColor    string `json:"bg_color"`
	TitleColor string `json:"title_color"`
	Attribute  int64  `json:"attribute"`
}

type ConfClick struct {
	Image  string   `json:"image"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Areas  []*Areas `json:"areas"`
	Bar    string   `json:"bar"`
}

type Areas struct {
	X               int    `json:"x"`
	Y               int    `json:"y"`
	W               int    `json:"w"`
	H               int    `json:"h"`
	Link            string `json:"link"`
	Type            int    `json:"type"`
	ForeignID       int64  `json:"foreign_id"`
	UnfinishedImage string `json:"unfinished_image"`
	FinishedImage   string `json:"finished_image"`
	Tip             string `json:"tip"`
	OptionalImage   string `json:"optional_image"`
	Images          [3]struct {
		Image  string `json:"image"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"images"` //浮层类型的图片
	TopColor      string `json:"top_color"` //顶栏颜色
	Title         string `json:"title"`
	TitleColor    string `json:"title_color"`
	IosLink       string `json:"ios_link"`
	AndroidLink   string `json:"android_link"`
	StatRule      int64  `json:"stat_rule"` //统计规则
	RuleName      string `json:"rule_name"` //规则名称
	StatDimension int64  `json:"stat_dimension"`
	Num           int64  `json:"num"`
	AlertNum      int64  `json:"alert_num"`
	FontSize      int64  `json:"font_size"`
	FontColor     string `json:"font_color"`
	UKey          string `json:"ukey"`
}

func (a *Areas) IsTypeCustom() bool {
	return a.Type > 0 && a.Type < 10
}

func (a *Areas) IsTypeRedirect() bool {
	return a.Type == 5
}

func (a *Areas) IsTypeLayer() bool {
	return a.Type >= 10 && a.Type < 20
}

func (a *Areas) IsTypeAPP() bool {
	return a.Type == 20
}

func (a *Areas) IsTypeProgress() bool {
	return a.Type == 30
}

type ConfAct struct {
	TitleImage string  `json:"title_image"`
	Attribute  int64   `json:"attribute"`
	BgColor    string  `json:"bg_color"`
	TitleColor string  `json:"title_color"`
	Acts       []*Acts `json:"acts"`
	Caption    string  `json:"caption"`
	Bar        string  `json:"bar"`
}

type Acts struct {
	PageID int64 `json:"page_id"`
}

type ConfDynamic struct {
	TitleImage    string `json:"title_image"`
	ShowNum       int    `json:"show_num"`
	NextTopicName string `json:"next_topic_name"`
	Pattern       int    `json:"pattern"`
	SelectType    string `json:"select_type"`
	ClassType     int64  `json:"class_type"`
	ClassID       int64  `json:"class_id"`
	SourceID      int64  `json:"source_id"`
	SourceType    string `json:"source_type"`
	SortType      string `json:"sort_type"`
	DySort        int    `json:"dy_sort"`
	Attribute     int64  `json:"attribute"`
	Caption       string `json:"caption"`
	Bar           string `json:"bar"`
	BgColor       string `json:"bg_color"`
	DisplayColor  string `json:"display_color"`
}

type ConfArchive struct {
	TitleImage   string         `json:"title_image"`
	Category     int            `json:"category"`
	AvSort       int            `json:"av_sort"` //老视频卡需要
	FID          int64          `json:"f_id"`
	TName        string         `json:"t_name"`
	Num          int            `json:"num"`
	ObjIDs       []int64        `json:"obj_ids"` //老视频卡需要
	Bvids        []string       `json:"bvids"`   //老视频卡需要
	Title        string         `json:"title"`
	SortType     string         `json:"sort_type"`
	CardStyle    int            `json:"card_style"`
	Attribute    int64          `json:"attribute"`
	BgColor      string         `json:"bg_color"`
	TitleColor   string         `json:"title_color"`
	MoreColor    string         `json:"more_color"`
	FontColor    string         `json:"font_color"`
	Caption      string         `json:"caption"`
	Bar          string         `json:"bar"`
	IDs          []*ResourceIDs `json:"ids"` //新视频卡需要
	DisplayColor string         `json:"display_color"`
}

type ConfTimeline struct {
	TitleImage    string         `json:"title_image"`    //图片标题
	Category      int            `json:"category"`       //类型
	Caption       string         `json:"caption"`        //文字标题
	Bar           string         `json:"bar"`            //导航标题
	TimeSort      int64          `json:"time_sort"`      //精确到 0:年 1:月 2: 日 3:时 4:分 5:秒
	Num           int            `json:"num"`            //外显数量
	Title         string         `json:"title"`          //浮层标题
	Axis          int64          `json:"axis"`           //时间轴节点类型 0:文本 1:时间节点
	Remark        string         `json:"remark"`         //查看更多文案
	MoreSort      int64          `json:"more_sort"`      //查看更多方式 0:跳转二级页面 1:浮层 3:下拉展示
	BgColor       string         `json:"bg_color"`       //背景色
	TimelineColor string         `json:"timeline_color"` //时间轴色
	TitleBgColor  string         `json:"title_bg_color"` //卡片背景色
	FID           int64          `json:"f_id"`           //资讯数据源id
	IDs           []*ResourceIDs `json:"ids"`
}

type ConfResource struct {
	TitleImage   string         `json:"title_image"`
	Category     int            `json:"category"`
	FID          int64          `json:"f_id"`
	TName        string         `json:"t_name"`
	Num          int            `json:"num"`
	IDs          []*ResourceIDs `json:"ids"`
	Title        string         `json:"title"`
	SortType     string         `json:"sort_type"`
	Attribute    int64          `json:"attribute"`
	BgColor      string         `json:"bg_color"`
	TitleColor   string         `json:"title_color"`
	MoreColor    string         `json:"more_color"`
	TitleBgColor string         `json:"title_bg_color"`
	FontColor    string         `json:"font_color"`
	Caption      string         `json:"caption"`
	Bar          string         `json:"bar"`
	DynStyle     string         `json:"dyn_style"` //动态类型：视频，文章与动态组件的type保持一致
	DisplayColor string         `json:"display_color"`
	RoleID       int64          `json:"role_id"`
	SeasonID     int64          `json:"season_id"`
}

// ResourceIDs .
type ResourceIDs struct {
	Type        int          `json:"type"`
	ID          int64        `json:"id"`
	Bvid        string       `json:"bvid"`
	Fid         int64        `json:"fid"`
	RcmdContent *RcmdContent `json:"rcmd_content"`      //编辑推荐内容
	Content     *MixContent  `json:"content,omitempty"` //组件内容
}

func (v *NatModule) ToLive(a *ConfLive, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = LiveModule // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.BgColor = a.BgColor
	v.FontColor = a.FontColor
	v.Meta = a.TitleImage
	v.FID = a.FID
	v.TName = a.TName
	v.Attribute = a.Attribute
	v.LiveType = a.LiveType
	v.Stime = a.Stime
	v.Etime = a.Etime
	v.Bar = a.Bar
	v.Caption = a.Caption
	colStr, _ := json.Marshal(&Colors{DisplayColor: a.DisplayColor})
	v.Colors = string(colStr)
}

// ToInlineTab
func (v *NatModule) ToInlineTab(a *InlineTab, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = InlineTabModule // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.BgColor = a.BgColor
	v.FontColor = a.FontColor   //未选中色
	v.MoreColor = a.SelectColor //选中色
	v.Attribute = a.Attribute
}

// ToSelect
func (v *NatModule) ToSelect(a *Select, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = SelectModule // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.BgColor = a.BgColor //背景色
	colStr, _ := json.Marshal(&Colors{
		PanelBgColor:        a.PanelBgColor,        // 展开面板背景色
		SelectColor:         a.SelectColor,         // 顶部字体色
		NotSelectColor:      a.NotSelectColor,      // 展开面板选中背景色
		PanelSelectColor:    a.PanelSelectColor,    // 展开面板字体选中色
		PanelNotSelectColor: a.PanelNotSelectColor, // 展开面板字体未选中色
	})
	v.Colors = string(colStr)
	v.Attribute = a.Attribute
}

func (v *NatModule) ToNavigation(a *Navigation, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = NavigationModule // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.BgColor = a.BgColor
	v.FontColor = a.FontColor
	v.TitleColor = a.SelectBgColor
	v.MoreColor = a.SelectFontColor
}

// ToBanner  .
func (v *NatModule) ToBanner(a *BannerImage, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = BannerModule // 组件类别
	v.Meta = a.Image
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.Bar = a.Bar
}

// ToHead  .
func (v *NatModule) ToHead(a *HeadBase, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = HeadModule // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.BgColor = a.BgColor
	v.Title = a.Title
	v.Attribute = a.Attribute
}

// ToStatement .
func (v *NatModule) ToStatement(a *StatementModule, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = StatementsModule // 组件类别
	v.Remark = a.Remark
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
	v.Bar = a.Bar
	v.BgColor = a.BgColor
	v.TitleColor = a.TitleColor
	v.Attribute = a.Attribute
}

// ToSingleDynamic  .
func (v *NatModule) ToSingleDynamic(nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = SingleDynamic // 组件类别
	v.Ukey = ukey
	v.Rank = order
	v.PType = pType
	v.State = 1
}

type ConfParticipation struct {
	Button []*ParticipationExt `json:"button"`
}

type ConfRecommned struct {
	TitleImage string      `json:"title_image"`
	RecoUsers  []*RecoUser `json:"users"`
	Bar        string      `json:"bar"`
	BgColor    string      `json:"bg_color"`
	TitleColor string      `json:"title_color"`
}

type CarouselImg struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Length      int64  `json:"length"`
	Width       int64  `json:"width"`
}

type CarouselWord struct {
	Content string `json:"content"`
}

type ConfCarousel struct {
	Category       int             `json:"category"`
	ContentStyle   int             `json:"content_style"`
	Attribute      int64           `json:"attribute"`
	BgColor        string          `json:"bg_color"`
	CardColor      string          `json:"card_color"`
	IndicatorColor string          `json:"indicator_color"`
	TitleImage     string          `json:"title_image"`
	ImgList        []*CarouselImg  `json:"img_list"`
	FontColor      string          `json:"font_color"`
	ScrollType     int             `json:"scroll_type"`
	WordList       []*CarouselWord `json:"word_list"`
}

type IconImg struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Content     string `json:"content"`
}

type ConfIcon struct {
	BgColor   string     `json:"bg_color"`
	FontColor string     `json:"font_color"`
	ImgList   []*IconImg `json:"img_list"`
}

type RecoUser struct {
	MID     int64  `json:"mid"`
	Content string `json:"content"`
	URI     string `json:"uri"`
}

// ToMclick to Module click .
func (v *NatModule) ToMclick(a *ConfClick, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = ClickModule // 组件类别
	v.Width = a.Width
	v.Length = a.Height
	v.State = 1
	v.Rank = order
	v.Meta = a.Image
	v.Ukey = ukey
	v.PType = pType
	v.Bar = a.Bar
}

// // ToMdynamic to Module dynamic
func (v *NatModule) ToMdynamic(a *ConfDynamic, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	if a.Pattern == 0 { // 组件类别
		v.Category = DynmaicModule
	} else if a.Pattern == 1 {
		v.Category = VideoModule
	}
	v.FID = a.SourceID
	v.Meta = a.TitleImage
	v.Num = a.ShowNum
	v.Title = a.NextTopicName
	v.State = 1
	v.Rank = order
	v.DySort = a.DySort
	v.Ukey = ukey
	v.Attribute = a.Attribute
	v.Caption = a.Caption
	v.PType = pType
	v.Bar = a.Bar
	v.BgColor = a.BgColor
	colStr, _ := json.Marshal(&Colors{DisplayColor: a.DisplayColor})
	v.Colors = string(colStr)
}

// ToMact to module act .
func (v *NatModule) ToMact(act *ConfAct, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = ActModule // 组件类别
	v.State = 1            // 有效
	v.Rank = order
	v.Meta = act.TitleImage
	v.Ukey = ukey
	v.Attribute = act.Attribute
	v.BgColor = act.BgColor
	v.TitleColor = act.TitleColor
	v.Caption = act.Caption
	v.PType = pType
	v.Bar = act.Bar
}

// ToTimeline to module archive.
func (v *NatModule) ToTimeline(arc *ConfTimeline, nativeID int64, order, pType int, ukey string) {
	v.Meta = arc.TitleImage
	v.Category = arc.Category
	v.Caption = arc.Caption
	v.Bar = arc.Bar
	v.Num = arc.Num
	v.Title = arc.Title
	v.Remark = arc.Remark
	confSort, _ := json.Marshal(&Sort{MoreSort: arc.MoreSort, TimeSort: arc.TimeSort, Axis: arc.Axis})
	v.ConfSort = string(confSort)
	v.BgColor = arc.BgColor
	colStr, _ := json.Marshal(&Colors{TitleBgColor: arc.TitleBgColor, TimelineColor: arc.TimelineColor})
	v.Colors = string(colStr)
	v.NativeID = nativeID
	v.Rank = order
	v.State = 1
	v.Ukey = ukey
	v.PType = pType
	v.FID = arc.FID
}

// ToMArchive to module archive.
func (v *NatModule) ToMArchive(arc *ConfArchive, nativeID int64, order, pType int, ukey string) {
	v.Category = arc.Category
	v.FID = arc.FID
	v.NativeID = nativeID
	v.Rank = order
	v.Num = arc.Num
	v.Title = arc.Title
	v.State = 1
	v.Ukey = ukey
	v.Attribute = arc.Attribute
	v.BgColor = arc.BgColor
	v.TitleColor = arc.TitleColor
	v.MoreColor = arc.MoreColor
	v.TName = arc.TName
	v.CardStyle = arc.CardStyle
	v.AvSort = arc.AvSort
	v.FontColor = arc.FontColor
	v.Meta = arc.TitleImage
	v.Caption = arc.Caption
	v.PType = pType
	v.Bar = arc.Bar
	colStr, _ := json.Marshal(&Colors{DisplayColor: arc.DisplayColor})
	v.Colors = string(colStr)
}

// ToMArchive to module archive.
func (v *NatModule) ToMResource(arc *ConfResource, nativeID int64, order, pType int, ukey string) {
	v.Category = arc.Category
	v.FID = arc.FID
	v.NativeID = nativeID
	v.Rank = order
	v.Num = arc.Num
	v.Title = arc.Title
	v.State = 1
	v.Ukey = ukey
	v.Attribute = arc.Attribute
	v.BgColor = arc.BgColor       //背景色
	v.TitleColor = arc.TitleColor //标题文本色
	v.MoreColor = arc.MoreColor   //查看更多按钮色
	v.FontColor = arc.FontColor
	v.TName = arc.TName
	v.Meta = arc.TitleImage
	v.Caption = arc.Caption
	v.PType = pType
	v.Bar = arc.Bar
	colStr, _ := json.Marshal(&Colors{DisplayColor: arc.DisplayColor, TitleBgColor: arc.TitleBgColor})
	v.Colors = string(colStr)
	if arc.Category == ResourceRoleModule {
		v.Length = int(arc.RoleID)  //角色id
		v.Width = int(arc.SeasonID) //剧集id
	}
}

// ToMPart to module Participation
func (v *NatModule) ToMPart(part *ConfParticipation, nativeID int64, order, pType int, ukey string) {
	v.Category = ParticipationModule
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Num = len(part.Button)
	v.State = 1
	v.Ukey = ukey
}

// ToMRecommend to module Recomment
func (v *NatModule) ToMRecommend(recomment *ConfRecommned, nativeID int64, order, pType int, ukey string) {
	v.Category = RecommentModule
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Num = len(recomment.RecoUsers)
	v.Meta = recomment.TitleImage
	v.State = 1
	v.Ukey = ukey
	v.Bar = recomment.Bar
	v.BgColor = recomment.BgColor
	v.TitleColor = recomment.TitleColor
}

func (v *NatModule) ToMRcmdVertical(rcmd *ConfRecommned, nativeID int64, order, pType int, ukey string) {
	v.Category = RcmdVerticalModule
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Num = len(rcmd.RecoUsers)
	v.Meta = rcmd.TitleImage
	v.State = 1
	v.Ukey = ukey
	v.Bar = rcmd.Bar
	v.BgColor = rcmd.BgColor
	v.TitleColor = rcmd.TitleColor
}

func (v *NatModule) ToEditor(conf *ConfEditor, nativeID int64, order, pType int, ukey string) {
	v.Category = EditorModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.Attribute = conf.Attribute
	if position, err := json.Marshal(conf.Positions); err == nil {
		v.TName = string(position)
	}
	v.BgColor = conf.BgColor
	v.Num = 40
}

func (v *NatModule) ToMCarousel(carousel *ConfCarousel, nativeID int64, order, pType int, ukey string) {
	v.Category = carousel.Category
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.AvSort = carousel.ContentStyle      //内容样式
	v.Attribute = carousel.Attribute      //是否自动轮播
	v.BgColor = carousel.BgColor          //背景色
	v.TitleColor = carousel.CardColor     //卡片背景
	v.MoreColor = carousel.IndicatorColor //指示符颜色
	v.Meta = carousel.TitleImage          //组件标题
	v.FontColor = carousel.FontColor      //文字色
	v.DySort = carousel.ScrollType        //滚动方向
}

func (v *NatModule) ToMIcon(icon *ConfIcon, nativeID int64, order, pType int, ukey string) {
	v.Category = IconModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.BgColor = icon.BgColor     //背景色
	v.FontColor = icon.FontColor //文字色
}

func (v *NatModule) ToProgress(conf *ConfProgress, nativeID int64, order, pType int, ukey string) {
	v.Category = ProgressModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.Attribute = conf.Attribute
	v.BgColor = conf.BgColor
	v.FID = conf.Fid
	v.FontColor = conf.FilledColor                            //填充色
	v.AvSort = int(conf.ProgressStyle)                        //1 圆角；2 矩形；3 分节
	v.TitleColor = strconv.FormatInt(conf.IndicatorStyle, 10) //达成态（进度条）：1 纯色填充；2 纹理颜色填充
	v.MoreColor = strconv.FormatInt(conf.BackgroundStyle, 10) //未达成态（进度槽）：1 描边；2 填充
	v.Length = int(conf.Texture)                              //纹理类型：1 纹理一；2 纹理二；3 纹理三
	v.Width = int(conf.StatDimension)                         //统计维度：0 当前用户行为；1 整体规则数据
	v.Colors = joinRuleIDName(conf.StatRule, conf.RuleName)   //统计规则、规则名称
}

// ToClick .
func (v *Click) ToClick(n *ConfClick, moduleID int64, a *Areas) {
	v.ModuleID = moduleID
	v.State = 1 // 组件生效
	v.Width = n.Width
	v.Length = n.Height // 点击区域的高
	v.LeftX = a.X
	v.LeftY = a.Y
	v.Width = a.W
	v.Length = a.H
	v.Link = a.Link
	v.Type = a.Type
	v.ForeignID = a.ForeignID
	switch {
	case a.IsTypeLayer():
		if image, err := json.Marshal(a.Images[0]); err == nil {
			v.UnfinishedImage = string(image)
		}
		if image, err := json.Marshal(a.Images[1]); err == nil {
			v.FinishedImage = string(image)
		}
		if image, err := json.Marshal(a.Images[2]); err == nil {
			v.OptionalImage = string(image)
		}
		var layer = struct {
			TopColor   string `json:"top_color"`
			Title      string `json:"title"`
			TitleColor string `json:"title_color"`
		}{
			TopColor:   a.TopColor,
			Title:      a.Title,
			TitleColor: a.TitleColor,
		}
		if layerBytes, err := json.Marshal(layer); err == nil {
			v.Tip = string(layerBytes)
		}
	case a.IsTypeAPP():
		v.UnfinishedImage = a.IosLink
		v.FinishedImage = a.AndroidLink
	case a.IsTypeProgress():
		v.UnfinishedImage = a.UKey
		v.FinishedImage = strconv.FormatInt(a.StatDimension, 10)
		v.OptionalImage = joinRuleIDName(a.StatRule, a.RuleName)
		var ext = struct {
			FontSize  int64  `json:"font_size"`
			FontColor string `json:"font_color"`
			Num       int64  `json:"num"`
		}{
			FontSize:  a.FontSize,
			FontColor: a.FontColor,
			Num:       a.Num,
		}
		if extBytes, err := json.Marshal(ext); err == nil {
			v.Tip = string(extBytes)
		}
	default:
		v.UnfinishedImage = a.UnfinishedImage
		v.FinishedImage = a.FinishedImage
		v.Tip = a.Tip
		v.OptionalImage = a.OptionalImage
	}
}

// ToAct .
func (v *Act) ToAct(moduleID int64, a *Acts, order int) {
	v.ModuleID = moduleID
	v.State = 1 // 组件生效
	v.Rank = order
	v.PageID = a.PageID
}

// ToDynamicExt .
func (v *DynamicExt) ToDynamicExt(selectType, moduleID, classID int64, isChoice bool) {
	v.ModuleID = moduleID
	v.State = 1
	// 精选模式
	if isChoice {
		v.ClassType = _choiceType
		v.ClassID = classID
	} else {
		//普通模式
		v.ClassType = _commonType
		v.SelectType = selectType
	}
}

// ToVideoExt .
func (v *VideoExt) ToVideoExt(moduleID int64, sortType int64, order int) {
	v.ModuleID = moduleID
	v.State = 1
	v.SortType = sortType
	v.Rank = order
}

// ToParticipation .
func (v *ParticipationExt) ToParticipation(moduleID int64, order int) {
	v.ModuleID = moduleID
	v.State = 1
	v.Rank = order + 1
}

type ConfEditor struct {
	Category  int            `json:"category"`
	Attribute int64          `json:"attribute"` //bit第10位，是否展示三点操作
	IDs       []*ResourceIDs `json:"ids"`       //资源
	Positions *Positions     `json:"positions"` //属性展示位置
	BgColor   string         `json:"bg_color"`  //卡片背景色
}

type Positions struct {
	Position1 string `json:"position1"` //位置1
	Position2 string `json:"position2"` //位置2
	Position3 string `json:"position3"` //位置3
	Position4 string `json:"position4"` //位置4
	Position5 string `json:"position5"` //位置5
}

type RcmdContent struct {
	TopContent      string `json:"top_content"`       //顶部推荐语
	TopFontColor    string `json:"top_font_color"`    //顶部字体颜色
	BottomContent   string `json:"bottom_content"`    //底部推荐语
	BottomFontColor string `json:"bottom_font_color"` //底部字体颜色
}

type ConfProgress struct {
	ProgressStyle   int64           `json:"progress_style"`
	IndicatorStyle  int64           `json:"indicator_style"`
	BackgroundStyle int64           `json:"background_style"`
	BgColor         string          `json:"bg_color"`
	FilledColor     string          `json:"filled_color"`
	Texture         int64           `json:"texture"`
	Attribute       int64           `json:"attribute"`
	StatDimension   int64           `json:"stat_dimension"`
	Fid             int64           `json:"fid"`
	StatRule        int64           `json:"stat_rule"`
	RuleName        string          `json:"rule_name"`
	AlertNum        int64           `json:"alert_num"`
	Nodes           []*ProgressNode `json:"nodes"`
}

type ProgressNode struct {
	Name string `json:"name"`
	Num  int64  `json:"num"`
}

type MixContent struct {
	Stime    int64  `json:"stime,omitempty"`     //时间轴组件-时间控件
	Title    string `json:"title,omitempty"`     //时间轴组件-主标题
	SubTitle string `json:"sub_title,omitempty"` //时间轴组件-副标题
	Desc     string `json:"desc,omitempty"`      //时间轴组件-描述
	Image    string `json:"image,omitempty"`     //时间轴组件-图片
	Url      string `json:"url,omitempty"`       //时间轴组件-跳转连接
	Name     string `json:"name,omitempty"`      //时间轴组件-阶段名
	Width    int32  `json:"width,omitempty"`     //时间轴组件-图片宽
	Length   int32  `json:"length,omitempty"`    //时间轴组件-图片长
}

// TableName native_page .
func (NatPage) TableName() string {
	return "native_page"
}

// TableName native_module .
func (NatModule) TableName() string {
	return "native_module"
}

// TableName native_click .
func (Click) TableName() string {
	return "native_click"
}

// TableName native_act .
func (Act) TableName() string {
	return "native_act"
}

// TableName native_dynamic_ext .
func (DynamicExt) TableName() string {
	return "native_dynamic_ext"
}

// TableName native_video_ext .
func (VideoExt) TableName() string {
	return "native_video_ext"
}

// TableName native_participation_ext
func (ParticipationExt) TableName() string {
	return "native_participation_ext"
}

type NatTsPage struct {
	ID        int64     `json:"id"`
	State     int       `json:"state"`
	Pid       int64     `json:"pid"`
	Ctime     time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	Title     string    `json:"title"`
	ForeignID int64     `json:"foreign_id"`
	Msg       string    `json:"msg"`
}

// NatModule .
type NatTsModule struct {
	ID       int64     `json:"id"             gorm:"column:id"`
	Category int       `json:"category"       gorm:"column:category"`
	TsID     int64     `json:"ts_id"      gorm:"column:ts_id"`
	Rank     int       `json:"rank"           gorm:"column:rank"`
	Meta     string    `json:"meta"           gorm:"column:meta"`
	Width    int       `json:"width"          gorm:"column:width"`
	Length   int       `json:"length"         gorm:"column:length"`
	State    int       `json:"state"          gorm:"column:state"`
	Ukey     string    `json:"ukey"           gorm:"column:ukey"`
	Remark   string    `json:"remark" gorm:"column:remark"`
	PType    int       `json:"p_type" gorm:"column:p_type"`
	Ctime    time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime    time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type NtCloudInfo struct {
	//通过审核=pass；结束活动=finished ；不通过审核=failed
	Status       string `json:"status"`
	TopicID      int64  `json:"topic_id"`
	ActivityName string `json:"activity_name"`
	ApplyDate    int32  `json:"apply_date"`
	BeginDate    int32  `json:"begin_date"`
	Ctime        int64  `json:"ctime"`
	Mid          int64  `json:"mid"`
}

func joinRuleIDName(ruleID int64, ruleName string) string {
	return fmt.Sprintf("%d_%s", ruleID, ruleName)
}
