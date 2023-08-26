package native

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	Choice         = "PICKED"
	_commonType    = 0
	_choiceType    = 1
	TsAutoAudit    = "auto" //自动审核通过
	_commonJump    = 0      //普通跳转
	_redirect      = 5      //跳转链接
	_onlyImage     = 7      //仅展示图片
	_layerImage    = 10     //图片模式
	_app           = 20     //拉起APP
	_interface     = 21     //点击区域-接口模式
	_progress      = 30     //进度数据
	_staticProcess = 31     //进度数据-静态
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
	ActOrigin string    `json:"act_origin" gorm:"column:act_origin"`
	Ctime     time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type NatPageExt struct {
	*NatPage
	PageDyn *PageDynRly `json:"page_dyn"`
	Ext     *PageExt    `json:"ext"`
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
	MoreSort int64 `json:"more_sort,omitempty"` //查看更多方式 0:跳转二级页面 1:浮层 3:下拉展示
	TimeSort int64 `json:"time_sort,omitempty"` //精确到 0:年 1:月 2: 日 3:时 4:分 5:秒
	Axis     int64 `json:"axis,omitempty"`      //时间轴节点类型 0:文本 1:时间节点
	RDBType  int32 `json:"rdb_type,omitempty"`  //资源小卡-外接数据源类型
	MseeType int32 `json:"msee_type"`           //入站必刷类型
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
	SupernatantColor    string `json:"supernatant_color,omitempty"`      //浮层标题文字色
	SubtitleColor       string `json:"subtitle_color,omitempty"`         //副标题文字色-三列   推荐语文字色-单列
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
	Ext             string `gorm:"column:ext"`
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
	SortType int64  `gorm:"column:sort_type"`
	SortName string `gorm:"column:sort_name"`
	Rank     int    `gorm:"column:rank"`
	Category int    `gorm:"column:category"`
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
	Ext       string `json:"ext" gorm:"column:ext"`
	NewTid    int64  `json:"new_tid" gorm:"-"`
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
	AvSort      int          `json:"av_sort"` //tab组件样式  0,2:颜色 1:图片
	Image       string       `json:"image"`   //背景图片
	Width       int          `json:"width"`
	Height      int          `json:"height"`
	Attribute   int64        `json:"attribute"`
	IDs         []*InlineIDs `json:"ids"`
}

// InlineIDs .
type InlineIDs struct {
	ID               int64      `json:"id"`
	Type             string     `json:"type"`              //每周必看:week
	LocationKey      string     `json:"location_key"`      //type:每周必看,对应的期数
	DisplayType      int        `json:"display_type"`      //0:无要求 1.解锁后展示
	DisplayCondition int        `json:"display_condition"` //0:无要求 1:时间 2.预约数据源
	Stime            int64      `json:"stime"`             //时间类型 开始时间
	UnLock           int        `json:"un_lock"`           //未解锁时 1:不展示 2:不可点
	Tip              string     `json:"tip"`               //提示文案
	UnImage          *ImageComm `json:"un_image"`          //未解锁态图片
	SelectImage      *ImageComm `json:"select_image"`      //选中态图片
	UnSelectImage    *ImageComm `json:"un_select_image"`   //未选中态图片
	DefType          int        `json:"def_type"`          //默认tab选择模式 0:无需处理 1:默认生效 2:定时生效
	DStime           int64      `json:"d_stime"`           //默认tab:定时生效开始时间 [1,2)
	DEtime           int64      `json:"d_etime"`           //默认tab:定时生效结束时间 [1,2)
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
	ID          int64  `json:"id"`
	Type        string `json:"type"`         //每周必看:week
	LocationKey string `json:"location_key"` //type:每周必看,对应的期数
	DefType     int    `json:"def_type"`     //默认tab选择模式 0:无需处理 1:默认生效 2:定时生效
	DStime      int64  `json:"d_stime"`      //默认tab:定时生效开始时间
	DEtime      int64  `json:"d_etime"`      //默认tab:定时生效结束时间
}

// CheckColor 导航组件色值 日间，夜间.
func (m *Navigation) CheckColor() bool {
	if m.BgColor != "" {
		bgColors := strings.Split(m.BgColor, ",")
		if len(bgColors) < _colorLen {
			return false
		}
	}
	if m.FontColor != "" {
		fontColors := strings.Split(m.FontColor, ",")
		if len(fontColors) < _colorLen {
			return false
		}
	}
	if m.SelectBgColor != "" {
		selectBgColors := strings.Split(m.SelectBgColor, ",")
		if len(selectBgColors) < _colorLen {
			return false
		}
	}
	if m.SelectFontColor != "" {
		selectFontColors := strings.Split(m.SelectFontColor, ",")
		if len(selectFontColors) < _colorLen {
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
	Image            string   `json:"image"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	Areas            []*Areas `json:"areas"`
	Bar              string   `json:"bar"`
	Attribute        int64    `json:"attribute"`
	DisplayType      int      `json:"display_type"`      //0:无要求 1.解锁后展示
	DisplayCondition int      `json:"display_condition"` //0:无要求 1:时间 2.预约数据源
	Stime            int64    `json:"stime"`             //时间类型 开始时间
}

type ConfVote struct {
	Image         string   `json:"image"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	Bar           string   `json:"bar"`
	Sort          int      `json:"sort"`     //0:二选一模式
	FID           int64    `json:"fid"`      //数据源id
	GroupID       int64    `json:"group_id"` //数据组
	UKey          string   `json:"ukey"`
	Areas         []*Areas `json:"areas"`          //数据展示区域,有顺序
	FinishedImage string   `json:"finished_image"` //完成态图片
	Attribute     int64    `json:"attribute"`
	SourceType    string   `json:"source_type"` //数据源类型：act_vote 活动投票；up_vote UP主投票
}

type Areas struct {
	X               int               `json:"x"`
	Y               int               `json:"y"`
	W               int               `json:"w"`
	H               int               `json:"h"`
	Link            string            `json:"link"`
	Type            int               `json:"type"`
	ForeignID       int64             `json:"foreign_id"`
	UnfinishedImage string            `json:"unfinished_image"`
	FinishedImage   string            `json:"finished_image"`
	Tip             string            `json:"tip"`
	OptionalImage   string            `json:"optional_image"`
	Images          []*api.Image      `json:"images"`    //浮层类型的图片
	TopColor        string            `json:"top_color"` //顶栏颜色
	Title           string            `json:"title"`
	TitleColor      string            `json:"title_color"`
	IosLink         string            `json:"ios_link"`
	AndroidLink     string            `json:"android_link"`
	StatDimension   int64             `json:"stat_dimension"`
	Num             int64             `json:"num"`
	FontSize        int64             `json:"font_size"`
	FontColor       string            `json:"font_color"`
	UKey            string            `json:"ukey"`
	FontType        string            `json:"font_type"`
	DisplayType     string            `json:"display_type"`
	LayerImage      string            `json:"layer_image"`  //浮层图片
	ButtonImage     string            `json:"button_image"` //按钮图片
	ShareImage      *api.Image        `json:"share_image"`  //分享图片
	Style           string            `json:"style"`        //浮层样式
	PSort           int32             `json:"p_sort"`
	Activity        string            `json:"activity"`
	Counter         string            `json:"counter"`
	LotteryID       string            `json:"lottery_id"`
	StatPeriodicity string            `json:"stat_periodicity"` //任务统计：统计周期 单日:daily 累计：total
	GroupID         int64             `json:"group_id"`
	NodeID          int64             `json:"node_id"`
	DisplayMode     int64             `json:"display_mode"`      //展示模式：0 无要求 1 解锁后展示
	UnlockCondition int64             `json:"unlock_condition"`  //解锁条件：0 无要求；1 时间；2 预约/积分进度
	Stime           int64             `json:"stime"`             //时间解锁-开始时间
	SyncHoverButton bool              `json:"sync_hover_button"` //是否同步悬浮按钮
	Sid             int64             `json:"sid"`               //数据源id
	Items           []*api.OptionItem `json:"items"`             //投票组件：选项颜色
	Rank            int64             `json:"rank"`              //投票组件-顺序字段
	ID              string            `json:"id"`                //点击区域唯一id
	NewTid          int64             `json:"new_tid"`           //新话题id
	UpType          int64             `json:"up_type"`           //页面选择：0 上传；1 排序
}

func (a *Areas) IsTypeCustom() bool {
	return a.Type > _commonJump && a.Type < _layerImage
}

func (a *Areas) IsTypeRedirect() bool {
	return a.Type == _redirect
}

func (a *Areas) IsTypeOnlyImage() bool {
	return a.Type == _onlyImage
}

func (a *Areas) IsTypeLayer() bool {
	return a.Type >= _layerImage && a.Type < _app
}

func (a *Areas) IsTypeAPP() bool {
	return a.Type == _app
}

func (a *Areas) IsTypeInterface() bool {
	return a.Type == _interface
}

func (a *Areas) IsTypeProgress() bool {
	return a.Type == _progress
}

func (a *Areas) IsTypeStaticProgress() bool {
	return a.Type == _staticProcess
}

func (a *Areas) IsTypeVoteProgress() bool {
	return a.Type == api.VoteProcess
}

func (a *Areas) IsTypeVoteButton() bool {
	return a.Type == api.VoteButton
}

func (a *Areas) IsTypeVoteUser() bool {
	return a.Type == api.VoteUser
}

func (a *Areas) IsTypePublishBtn() bool {
	return a.Type == api.ClickPublishBtn
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

type ConfActCapsule struct {
	Caption string  `json:"caption"` //文字标题
	Bar     string  `json:"bar"`     //导航标题
	BgColor string  `json:"bg_color"`
	Acts    []*Acts `json:"acts"`
}

type Acts struct {
	PageID int64 `json:"page_id"`
}

type ConfGame struct {
	TitleImage string   `json:"title_image"`
	BgColor    string   `json:"bg_color"`
	TitleColor string   `json:"title_color"`
	Games      []*Games `json:"games"`
	Caption    string   `json:"caption"`
	Bar        string   `json:"bar"`
}

type ConfReserve struct {
	TitleImage   string  `json:"title_image"`
	BgColor      string  `json:"bg_color"`       //背景色
	TitleBgColor string  `json:"title_bg_color"` //卡片背景色
	TitleColor   string  `json:"title_color"`    //文字色
	Sids         []int64 `json:"sids"`
	Caption      string  `json:"caption"`
	Bar          string  `json:"bar"`
	Attribute    int64   `json:"attribute"`
}

type Games struct {
	GameID  int64  `json:"game_id"`
	Content string `json:"content"`
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

type ConfReply struct {
	FID    int64 `json:"f_id"`    //评论id
	AvSort int   `json:"av_sort"` // 评论类型 活动：4
}

type ConfOgvSeason struct {
	TitleImage       string         `json:"title_image"`       //图片标题
	Category         int            `json:"category"`          //类型
	Caption          string         `json:"caption"`           //文字标题
	Bar              string         `json:"bar"`               //导航标题
	CardStyle        int            `json:"card_style"`        //排列样式 1:单列 3:三列
	BgColor          string         `json:"bg_color"`          //背景色
	TitleBgColor     string         `json:"title_bg_color"`    //卡片背景色-单列
	MoreColor        string         `json:"more_color"`        //查看更多按钮色
	FontColor        string         `json:"font_color"`        //查看更多文字色
	TitleColor       string         `json:"title_color"`       //剧集标题色-三列
	DisplayColor     string         `json:"display_color"`     //文字标题文字色
	SubtitleColor    string         `json:"subtitle_color"`    //副标题文字色-三列   推荐语文字色-单列
	SupernatantColor string         `json:"supernatant_color"` //浮层标题文字色
	Num              int            `json:"num"`               //外显数量
	Title            string         `json:"title"`             //浮层标题
	Remark           string         `json:"remark"`            //查看更多文案
	FID              int64          `json:"f_id"`              //片单id
	Attribute        int64          `json:"attribute"`         // 3:是否不展示查看更多 0：否 1是 7:ogv剧集卡是否展示付费角标 11:ogv是否展示评分 14:是否展示推荐语
	IDs              []*ResourceIDs `json:"ids"`
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
	SortList     []*SortList    `json:"sort_list"`
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
	RDBType      int32          `json:"rdb_type"` // 外接数据源类型
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

type SortList struct {
	SortName string `json:"sort_name"`
	SortType int64  `json:"sort_type"`
	Category int    `json:"category"` // 1:自定义
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
	v.AvSort = a.AvSort //tab组件样式  0,2:颜色 1:图片
	v.Meta = a.Image    //背景图
	v.Width = a.Width
	v.Length = a.Height
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
	if len(colStr) <= _colLen { //数据库有长度限制
		v.Colors = string(colStr)
	}
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
	Category   int            `json:"category"`
	TitleImage string         `json:"title_image"`
	RecoUsers  []*RecoUser    `json:"users"`
	IDs        []*ResourceIDs `json:"ids"` //资源
	Bar        string         `json:"bar"`
	BgColor    string         `json:"bg_color"`
	TitleColor string         `json:"title_color"`
	Fid        int64          `json:"fid"`
	SortType   string         `json:"sort_type"`
	SourceType string         `json:"source_type"`
	Num        int            `json:"num"`
	Attribute  int64          `json:"attribute"`
}

type CarouselImg struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Length      int64  `json:"length"`
	Width       int64  `json:"width"`
	// 首页配置
	ConfSet
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
	Fid            int64           `json:"fid"`         //数据源id
	SortType       string          `json:"sort_type"`   //排序：random 随机；ctime 创建时间
	SourceType     string          `json:"source_type"` //数据源类型：act_up 活动的up主；
	Length         int             `json:"length"`
	Width          int             `json:"width"`
	Bar            string          `json:"bar"`
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
	Bar       string     `json:"bar"`
}

type RecoUser struct {
	MID     int64  `json:"mid"`
	Content string `json:"content"`
	URI     string `json:"uri"`
}

type ConfSet struct {
	//首页tab相关配置
	BgType         int    `json:"bg_type,omitempty"`          //背景配置模式 1:颜色 2:图片
	TabTopColor    string `json:"tab_top_color,omitempty"`    //顶栏头部色值
	TabMiddleColor string `json:"tab_middle_color,omitempty"` //中间色值
	TabBottomColor string `json:"tab_bottom_color,omitempty"` //tab栏底部色值
	FontColor      string `json:"font_color,omitempty"`       //tab文本高亮色值
	BarType        int    `json:"bar_type,omitempty"`         //系统状态栏色值 1:白色 0:默认黑色
	BgImage1       string `json:"bg_image_1,omitempty"`       //背景图1
	BgImage2       string `json:"bg_image_2,omitempty"`       //背景图2
	//inline 内嵌页配置
	DT     int    `json:"dt,omitempty"`      //0:无要求 1.解锁后展示
	DC     int    `json:"dc,omitempty"`      //0:无要求 1:时间 2.预约数据源
	Stime  int64  `json:"stime,omitempty"`   //时间类型 开始时间
	UnLock int    `json:"un_lock,omitempty"` //未解锁时 1:不展示 2:不可点
	Tip    string `json:"tip,omitempty"`     //提示文案
	//inline 内嵌页配置
}

func (v *NatModule) ToMVote(a *ConfVote, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = api.VoteModule // 组件类别
	v.Width = a.Width
	v.Length = a.Height
	v.State = 1
	v.Rank = order
	v.Meta = a.Image
	v.Ukey = ukey
	v.PType = pType
	v.Bar = a.Bar
	v.DySort = a.Sort
	v.Attribute = a.Attribute
	v.FID = a.FID
	func() {
		confSort, err := json.Marshal(&api.ConfSort{Image: a.FinishedImage, Sid: a.GroupID, SourceType: a.SourceType})
		if err != nil {
			log.Error("Fail to marshal confSort, ToMVote=%+v error=%+v", a, err)
			return
		}
		v.ConfSort = string(confSort)
	}()
}

// ToMclick to Module click .
func (v *NatModule) ToMclick(a *ConfClick, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = ClickModule // 组件类别
	v.Attribute = a.Attribute
	v.Width = a.Width
	v.Length = a.Height
	v.State = 1
	v.Rank = order
	v.Meta = a.Image
	v.Ukey = ukey
	v.PType = pType
	v.Bar = a.Bar
	//解锁模式
	v.AvSort = a.DisplayType      //0:无要求 1.解锁后展示
	v.DySort = a.DisplayCondition //0:无要求 1:时间 2.预约数据源
	v.Stime = a.Stime             //时间类型 开始时间
}

func (v *NatModule) ToMbottomButton(a *ConfBottomButton, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = api.BottomButtonModule // 组件类别
	v.Attribute = a.Attribute
	v.Width = a.Width
	v.Length = a.Height
	v.State = 1
	v.Rank = order
	v.Meta = a.Image
	v.Ukey = ukey
	v.PType = pType
	v.Bar = a.Bar
	//解锁模式
	v.AvSort = a.DisplayType      //0:无要求 1.解锁后展示
	v.DySort = a.DisplayCondition //0:无要求 1:时间 2.预约数据源
	v.Stime = a.Stime             //时间类型 开始时间
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

func (v *NatModule) ToActCapsule(act *ConfActCapsule, nativeID int64, order, pType int, ukey string) {
	v.NativeID = nativeID
	v.Category = ActCapsuleModule
	v.State = 1
	v.Rank = order
	v.Ukey = ukey
	v.BgColor = act.BgColor
	v.Caption = act.Caption
	v.Bar = act.Bar
	v.PType = pType
}

func (v *NatModule) ToReply(arc *ConfReply, nativeID int64, order, pType int, ukey string) {
	v.Category = ReplyModule
	v.NativeID = nativeID
	v.Rank = order
	v.State = 1
	v.Ukey = ukey
	v.PType = pType
	v.FID = arc.FID
	v.AvSort = arc.AvSort
}

// ToOgvSeason to module archive.
func (v *NatModule) ToOgvSeason(arc *ConfOgvSeason, nativeID int64, order, pType int, ukey string) {
	v.Meta = arc.TitleImage
	v.Category = arc.Category
	v.Caption = arc.Caption
	v.Bar = arc.Bar
	v.Num = arc.Num
	v.Title = arc.Title
	v.Remark = arc.Remark
	v.BgColor = arc.BgColor
	v.FontColor = arc.FontColor
	v.MoreColor = arc.MoreColor
	v.TitleColor = arc.TitleColor
	colStr, _ := json.Marshal(&Colors{TitleBgColor: arc.TitleBgColor, DisplayColor: arc.DisplayColor, SupernatantColor: arc.SupernatantColor, SubtitleColor: arc.SubtitleColor})
	if len(colStr) <= _colLen {
		v.Colors = string(colStr)
	}
	v.NativeID = nativeID
	v.Rank = order
	v.State = 1
	v.Ukey = ukey
	v.PType = pType
	v.FID = arc.FID
	v.Attribute = arc.Attribute
	v.CardStyle = arc.CardStyle
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
	if len(confSort) <= _colLen { //数据库有长度限制
		v.ConfSort = string(confSort)
	}
	v.BgColor = arc.BgColor
	colStr, _ := json.Marshal(&Colors{TitleBgColor: arc.TitleBgColor, TimelineColor: arc.TimelineColor})
	if len(colStr) <= _colLen { //数据库有长度限制
		v.Colors = string(colStr)
	}
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
	v.TName = arc.TName // 外接数据源id-string类型,话题活动类型-话题名
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
	if arc.Category == ResourceDataOriginModule {
		confSort, _ := json.Marshal(&Sort{RDBType: arc.RDBType})
		v.ConfSort = string(confSort)
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

// ToMGame to module ToMReserve
func (v *NatModule) ToMReserve(ga *ConfReserve, nativeID int64, order, pType int, ukey string) {
	v.Category = api.ModuleReserve
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Meta = ga.TitleImage
	v.State = 1
	v.Ukey = ukey
	v.Bar = ga.Bar
	v.BgColor = ga.BgColor
	v.TitleColor = ga.TitleColor
	colStr, _ := json.Marshal(&Colors{TitleBgColor: ga.TitleBgColor})
	if len(colStr) <= _colLen { //数据库有长度限制
		v.Colors = string(colStr)
	}
	v.Caption = ga.Caption
	v.Attribute = ga.Attribute
}

// ToMGame to module ToMGame
func (v *NatModule) ToMGame(ga *ConfGame, nativeID int64, order, pType int, ukey string) {
	v.Category = api.ModuleGame
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Meta = ga.TitleImage
	v.State = 1
	v.Ukey = ukey
	v.Bar = ga.Bar
	v.BgColor = ga.BgColor
	v.TitleColor = ga.TitleColor
	v.Caption = ga.Caption
}

// ToMRecommend to module Recomment
func (v *NatModule) ToMRecommend(recomment *ConfRecommned, nativeID int64, order, pType int, ukey string) {
	v.Category = recomment.Category
	v.PType = pType
	v.NativeID = nativeID
	v.Rank = order
	v.Num = recomment.Num
	if len(recomment.RecoUsers) > 0 {
		v.Num = len(recomment.RecoUsers)
	}
	v.Meta = recomment.TitleImage
	v.State = 1
	v.Ukey = ukey
	v.Bar = recomment.Bar
	v.BgColor = recomment.BgColor
	v.TitleColor = recomment.TitleColor
	if v.Category == RcmdSourceModule {
		v.FID = recomment.Fid
		if recomment.SourceType == api.SourceTypeRank {
			v.Num = recomment.Num
			v.Attribute = recomment.Attribute
		}
		func() {
			confSort, err := json.Marshal(&api.ConfSort{SortType: recomment.SortType, SourceType: recomment.SourceType})
			if err != nil {
				log.Error("Fail to marshal confSort, recommendConf=%+v error=%+v", recomment, err)
				return
			}
			v.ConfSort = string(confSort)
		}()
	}
}

func (v *NatModule) ToMRcmdVertical(rcmd *ConfRecommned, nativeID int64, order, pType int, ukey string) {
	v.Category = rcmd.Category
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
	if v.Category == RcmdVerticalSourceModule {
		v.FID = rcmd.Fid
		func() {
			confSort, err := json.Marshal(&api.ConfSort{SortType: rcmd.SortType, SourceType: rcmd.SourceType})
			if err != nil {
				log.Error("Fail to marshal confSort, rcmdVerticalConf=%+v error=%+v", rcmd, err)
				return
			}
			v.ConfSort = string(confSort)
		}()
	}
}

func (v *NatModule) ToEditor(conf *ConfEditor, nativeID int64, order, pType int, ukey string) {
	var confSort *api.ConfSort
	if conf.Category == api.ModuleEditorOrigin {
		v.Category = api.ModuleEditorOrigin
		v.FID = conf.FID
		if confSort == nil {
			confSort = &api.ConfSort{}
		}
		confSort.RdbType = int64(conf.RDBType)
		confSort.MseeType = conf.MustseeType
	} else {
		v.Category = api.ModuleEditor
	}
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.Attribute = conf.Attribute
	if position, err := json.Marshal(conf.Positions); err == nil {
		v.TName = string(position)
	}
	func() {
		if conf.PointSid == 0 || conf.Counter == "" {
			return
		}
		if confSort == nil {
			confSort = &api.ConfSort{}
		}
		confSort.Sid = conf.PointSid
		confSort.Counter = conf.Counter
	}()
	func() {
		if confSort == nil {
			return
		}
		cs, err := json.Marshal(confSort)
		if err != nil {
			log.Error("Fail to marshal ConfSort, ConfSort=%+v error=%+v", confSort, err)
			return
		}
		v.ConfSort = string(cs)
	}()
	v.BgColor = conf.BgColor
	v.Num = 40
	if conf.RDBType == api.RDBRank {
		v.Num = conf.Num
	}
	v.Bar = conf.Bar
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
	v.Bar = carousel.Bar
	if carousel.Category == CarouselSourceModule {
		v.Length = carousel.Length
		v.Width = carousel.Width
		v.FID = carousel.Fid
		func() {
			confSort, err := json.Marshal(&api.ConfSort{SortType: carousel.SortType, SourceType: carousel.SourceType})
			if err != nil {
				log.Error("Fail to marshal confSort, carouselConf=%+v error=%+v", carousel, err)
				return
			}
			v.ConfSort = string(confSort)
		}()
	}
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
	v.Bar = icon.Bar
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
	v.Width = int(conf.GroupID)                               //节点组id
	v.Bar = conf.Bar
}

func (v *NatModule) ToHoverButton(conf *ConfHoverButton, nativeID int64, order, pType int, ukey string) {
	v.Category = HoverButtonModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.ForeignId
	v.FontColor = conf.UnfinishedImage
	v.TitleColor = conf.FinishedImage
	v.MoreColor = conf.Image
	v.Colors = conf.Link
	func() {
		confSort := &api.ConfSort{BtType: conf.ButtonType, Hint: conf.SuccessHint}
		if conf.MutexUkeys != "" {
			confSort.MUkeys = strings.Split(conf.MutexUkeys, ",")
		}
		raw, err := json.Marshal(confSort)
		if err != nil {
			log.Error("Fail to marshal confSort, hoverButton=%+v error=%+v", confSort, err)
			return
		}
		v.ConfSort = string(raw)
	}()
}

func (v *NatModule) ToNewactHeader(conf *ConfNewactHeader, nativeID int64, order, pType int, ukey string) {
	v.Category = api.ModuleNewactHeader
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.Fid
}

func (v *NatModule) ToNewactAward(conf *ConfNewactAward, nativeID int64, order, pType int, ukey string) {
	v.Category = api.ModuleNewactAward
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.Fid
}

func (v *NatModule) ToNewactStatement(conf *ConfNewactStatement, nativeID int64, order, pType int, ukey string) {
	v.Category = api.ModuleNewactStatement
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.Fid
	func() {
		confSort, err := json.Marshal(&api.ConfSort{StatementType: conf.Type})
		if err != nil {
			log.Error("Fail to marshal confSort, conf=%+v error=%+v", conf, err)
			return
		}
		v.ConfSort = string(confSort)
	}()
}

func (v *NatModule) ToMatchMedal(conf *ConfMatchMedal, nativeID int64, order, pType int, ukey string) {
	v.Category = api.MatchMedalModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.Fid
	v.BgColor = conf.BgColor
}

func (v *NatModule) ToMatchEvent(conf *ConfMatchEvent, nativeID int64, order, pType int, ukey string) {
	v.Category = api.MatchEventModule
	v.NativeID = nativeID
	v.Rank = order
	v.PType = pType
	v.Ukey = ukey
	v.State = 1
	v.FID = conf.Fid
	v.BgColor = conf.BgColor
}

// ToVote .
func (v *Click) ToVote(moduleID int64, a *Areas) {
	v.ModuleID = moduleID
	v.State = 1 // 组件生效
	v.LeftX = a.X
	v.LeftY = a.Y
	v.Width = a.W
	v.Length = a.H
	v.Type = a.Type
	clickExt := &api.ClickExt{Ukey: a.ID}
	switch {
	case a.IsTypeVoteButton():
		v.UnfinishedImage = a.UnfinishedImage
	case a.IsTypeVoteProgress():
		clickExt.Items = a.Items
		clickExt.Style = a.Style //circle | square
	default:
	}
	func() {
		if clickExt == nil {
			return
		}
		extBytes, err := json.Marshal(clickExt)
		if err != nil {
			log.Error("Fail to marshal clickExt, clickExt=%+v error=%+v", clickExt, err)
			return
		}
		v.Ext = string(extBytes)
	}()
}

// ToClick .
func (v *Click) ToClick(moduleID int64, a *Areas, width, height int) {
	v.ModuleID = moduleID
	v.State = 1 // 组件生效
	v.Width = width
	v.Length = height // 点击区域的高
	v.LeftX = a.X
	v.LeftY = a.Y
	v.Width = a.W
	v.Length = a.H
	v.Link = a.Link
	v.Type = a.Type
	v.ForeignID = a.ForeignID
	clickExt := &api.ClickExt{Ukey: a.UKey}
	switch {
	case a.IsTypeLayer():
		v.UnfinishedImage, v.FinishedImage, v.OptionalImage = layerImages(a.Images)
		tip := &api.ClickTip{
			TopColor:   a.TopColor,
			Title:      a.Title,
			TitleColor: a.TitleColor,
		}
		if tipBytes, err := json.Marshal(tip); err == nil {
			v.Tip = string(tipBytes)
		}
		clickExt.LayerImage = a.LayerImage
		clickExt.ButtonImage = a.ButtonImage
		clickExt.ShareImage = a.ShareImage
		clickExt.Style = a.Style
		clickExt.Images = restLayerImages(a.Images)
	case a.IsTypeAPP():
		v.UnfinishedImage = a.IosLink
		v.FinishedImage = a.AndroidLink
	case a.IsTypeProgress(), a.IsTypeStaticProgress():
		v.UnfinishedImage = a.UKey
		v.FinishedImage = strconv.FormatInt(a.StatDimension, 10)
		ext := &api.ClickTip{
			FontSize:    a.FontSize,
			FontColor:   a.FontColor,
			Num:         a.Num,
			FontType:    a.FontType,
			DisplayType: a.DisplayType,
			PSort:       a.PSort,
			NodeId:      a.NodeID,
			GroupId:     a.GroupID,
		}
		if a.PSort == api.ProcessLottery {
			ext.LotteryID = a.LotteryID
		} else if a.PSort == api.ProcessTaskStatics {
			ext.Activity = a.Activity
			ext.Counter = a.Counter
			ext.StatPc = a.StatPeriodicity
		}
		if extBytes, err := json.Marshal(ext); err == nil {
			v.Tip = string(extBytes)
		}
	case a.IsTypeInterface():
		clickExt.Style = a.Style
	case a.IsTypeOnlyImage():
		v.OptionalImage = a.ButtonImage
	case a.IsTypePublishBtn():
		clickExt.Sid = a.NewTid
		clickExt.UpType = a.UpType
	default:
		v.UnfinishedImage = a.UnfinishedImage
		v.FinishedImage = a.FinishedImage
		v.Tip = a.Tip
		v.OptionalImage = a.OptionalImage
	}
	setCustomClickExt(clickExt, a)
	func() {
		if clickExt == nil {
			return
		}
		extBytes, err := json.Marshal(clickExt)
		if err != nil {
			log.Error("Fail to marshal clickExt, clickExt=%+v error=%+v", clickExt, err)
			return
		}
		v.Ext = string(extBytes)
	}()
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
func (v *VideoExt) ToVideoExt(moduleID int64, sortType int64, order, category int, sortName string) {
	v.ModuleID = moduleID
	v.State = 1
	v.SortType = sortType
	v.Rank = order
	v.Category = category
	v.SortName = sortName
}

// ToParticipation .
func (v *ParticipationExt) ToParticipation(moduleID int64, order int) {
	v.ModuleID = moduleID
	v.State = 1
	v.Rank = order + 1
	func() {
		if v.NewTid <= 0 {
			return
		}
		ext, err := json.Marshal(&api.PartiExt{NewTid: v.NewTid})
		if err != nil {
			log.Error("Fail to marshal PartiExt, data=%+v error=%+v", v, err)
			return
		}
		v.Ext = string(ext)
	}()
}

type ConfEditor struct {
	Category    int            `json:"category"`
	Attribute   int64          `json:"attribute"` //bit第10位，是否展示三点操作
	IDs         []*ResourceIDs `json:"ids"`       //资源
	Positions   *Positions     `json:"positions"` //属性展示位置
	BgColor     string         `json:"bg_color"`  //卡片背景色
	RDBType     int32          `json:"rdb_type"`  // 外接数据源类型
	FID         int64          `json:"fid"`
	Bar         string         `json:"bar"`
	PointSid    int64          `json:"point_sid"`    //积分数据源id
	Counter     string         `json:"counter"`      //数据源统计规则counter
	MustseeType int32          `json:"mustsee_type"` //入站必刷类型：1 新晋宝藏；2 历史经典；0 全部
	Num         int            `json:"num"`
}

type Positions struct {
	Position1 string `json:"position1"` //位置1
	Position2 string `json:"position2"` //位置2
	Position3 string `json:"position3"` //位置3
	Position4 string `json:"position4"` //位置4
	Position5 string `json:"position5"` //位置5
}

type RcmdContent struct {
	TopContent      string `json:"top_content,omitempty"`       //顶部推荐语
	TopFontColor    string `json:"top_font_color,omitempty"`    //顶部字体颜色
	BottomContent   string `json:"bottom_content,omitempty"`    //底部推荐语
	BottomFontColor string `json:"bottom_font_color,omitempty"` //底部字体颜色
	MiddleIcon      string `json:"middle_icon,omitempty"`       //排行icon
}

type ConfProgress struct {
	ProgressStyle   int64  `json:"progress_style"`
	IndicatorStyle  int64  `json:"indicator_style"`
	BackgroundStyle int64  `json:"background_style"`
	BgColor         string `json:"bg_color"`
	FilledColor     string `json:"filled_color"`
	Texture         int64  `json:"texture"`
	Attribute       int64  `json:"attribute"`
	Fid             int64  `json:"fid"`
	GroupID         int64  `json:"group_id"`
	Bar             string `json:"bar"`
}

type ConfHoverButton struct {
	ButtonType      string `json:"button_type"`
	ForeignId       int64  `json:"foreign_id"`
	UnfinishedImage string `json:"unfinished_image"`
	FinishedImage   string `json:"finished_image"`
	SuccessHint     string `json:"success_hint"`
	Link            string `json:"link"`
	MutexUkeys      string `json:"mutex_ukeys"`
	Image           string `json:"image"`
}

type ConfMatchMedal struct {
	BgColor string `json:"bg_color"`
	Fid     int64  `json:"fid"`
}

type ConfMatchEvent struct {
	BgColor  string  `json:"bg_color"`
	Fid      int64   `json:"fid"`
	EventIds []int64 `json:"event_ids"`
}

type ProgressNode struct {
	Name string `json:"name"`
	Num  int64  `json:"num"`
}

type MixContent struct {
	Stime       int64      `json:"stime,omitempty"`        //时间轴组件-时间控件
	Title       string     `json:"title,omitempty"`        //时间轴组件-主标题
	SubTitle    string     `json:"sub_title,omitempty"`    //时间轴组件-副标题
	Desc        string     `json:"desc,omitempty"`         //时间轴组件-描述
	Image       string     `json:"image,omitempty"`        //时间轴组件-图片&&推荐用户icon
	Url         string     `json:"url,omitempty"`          //时间轴组件-跳转连接
	Name        string     `json:"name,omitempty"`         //时间轴组件-阶段名
	Width       int32      `json:"width,omitempty"`        //时间轴组件-图片宽
	Length      int32      `json:"length,omitempty"`       //时间轴组件-图片长
	Type        string     `json:"type,omitempty"`         //inline-tab&筛选组件-定位类型 每周必卡:week
	LocationKey string     `json:"location_key,omitempty"` //inline-tab&筛选组件-定位类型 type=week：每周必看期数id
	UnI         *ImageComm `json:"un_i,omitempty"`         //inline-tab组件未解锁态图片
	SI          *ImageComm `json:"si,omitempty"`           //inline-tab组件选中态图片
	UnSI        *ImageComm `json:"un_si,omitempty"`        //inline-tab组件未选中态图片
	DStime      int64      `json:"d_stime,omitempty"`      //inline-tab&筛选组件 默认tab定时生效开始时间
	DEtime      int64      `json:"d_etime,omitempty"`      //inline-tab&筛选组件  默认tab定时生效结束时间
	DefType     int        `json:"def_type,omitempty"`     //inline-tab&筛选组件 默认tab选择模式 0:无需处理 1:默认生效 2:定时生效
}

type DefTime struct {
	DStime int64 `json:"d_stime,omitempty"` //inline-tab&筛选组件 默认tab定时生效开始时间
	DEtime int64 `json:"d_etime,omitempty"` //inline-tab&筛选组件  默认tab定时生效结束时间
}

func DefCheckTime(in []*DefTime) bool {
	fullLen := len(in)
	if fullLen == 0 {
		return false
	}
	sort.Slice(in, func(i, j int) bool {
		return in[i].DStime < in[j].DStime
	})
	for i, v := range in {
		if v.DStime >= v.DEtime {
			return true
		}
		if i+1 < fullLen {
			if v.DEtime > in[i+1].DStime {
				return true
			}
		}
	}
	return false
}

type ImageComm struct {
	Image  string `json:"image,omitempty"`
	Width  int32  `json:"width,omitempty"`
	Height int32  `json:"height,omitempty"`
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
	ID           int64     `json:"id"`
	State        int       `json:"state"`
	Pid          int64     `json:"pid"`
	Ctime        time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime        time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	Title        string    `json:"title"`
	ForeignID    int64     `json:"foreign_id"`
	Msg          string    `json:"msg"`
	VideoDisplay string    `json:"video_display"`
	AuditType    string    `json:"audit_type"`
	AuditTime    int64     `json:"audit_time"`
	ShareImage   string    `json:"share_image"`
}

// NatModule .
type NatTsModule struct {
	ID        int64     `json:"id"             gorm:"column:id"`
	Category  int       `json:"category"       gorm:"column:category"`
	TsID      int64     `json:"ts_id"      gorm:"column:ts_id"`
	Rank      int       `json:"rank"           gorm:"column:rank"`
	Meta      string    `json:"meta"           gorm:"column:meta"`
	Width     int       `json:"width"          gorm:"column:width"`
	Length    int       `json:"length"         gorm:"column:length"`
	State     int       `json:"state"          gorm:"column:state"`
	Ukey      string    `json:"ukey"           gorm:"column:ukey"`
	Remark    string    `json:"remark" gorm:"column:remark"`
	PType     int       `json:"p_type" gorm:"column:p_type"`
	Ctime     time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	Num       int       `json:"num" gorm:"column:num"`
	Attribute int64     `json:"attribute" gorm:"column:attribute"`
}

func (m *NatTsModule) Trans2ConfResource() *ConfResource {
	return &ConfResource{
		Category:  m.Category,
		Num:       m.Num,
		Attribute: m.Attribute,
	}
}

func (m *NatTsModule) Trans2ConfArchive() *ConfArchive {
	return &ConfArchive{
		Category:  m.Category,
		Num:       m.Num,
		Attribute: m.Attribute,
	}
}

type NatTsModuleResource struct {
	ID           int64   `json:"id" gorm:"column:id"`
	ModuleID     int64   `json:"module_id" gorm:"column:module_id"`
	ResourceID   int64   `json:"resource_id" gorm:"column:resource_id"`
	ResourceType int64   `json:"resource_type" gorm:"column:resource_type"`
	Rank         int64   `json:"rank" gorm:"column:rank"`
	ResourceFrom string  `json:"resource_from" gorm:"column:resource_from"`
	State        int64   `json:"state" gorm:"column:state"`
	Ctime        StrTime `json:"ctime" gorm:"column:ctime"`
	Mtime        StrTime `json:"mtime" gorm:"column:mtime"`
	Ext          string  `json:"ext" gorm:"column:ext"`
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

func layerImages(images []*api.Image) (unfinishedImage, finishedImage, optionalImage string) {
	var (
		zero = 0
		one  = 1
		two  = 2
	)
	l := len(images)
	if l > zero {
		if image, err := json.Marshal(images[0]); err == nil {
			unfinishedImage = string(image)
		}
	}
	if l > one {
		if image, err := json.Marshal(images[1]); err == nil {
			finishedImage = string(image)
		}
	}
	if l > two {
		if image, err := json.Marshal(images[2]); err == nil {
			optionalImage = string(image)
		}
	}
	return
}

func restLayerImages(images []*api.Image) []*api.Image {
	var (
		three  = 3
		twelve = 12
	)
	l := len(images)
	if l <= three {
		return nil
	}
	if l > twelve {
		l = twelve
	}
	// 最多12张图片的限制
	return images[three:l]
}

func setCustomClickExt(clickExt *api.ClickExt, a *Areas) {
	if !a.IsTypeCustom() {
		return
	}
	clickExt.SynHover = a.SyncHoverButton
	clickExt.DisplayMode = a.DisplayMode
	clickExt.UnlockCondition = a.UnlockCondition
	clickExt.Stime = a.Stime
	clickExt.GroupId = a.GroupID
	clickExt.NodeId = a.NodeID
	clickExt.Sid = a.Sid
}

type ConfNewactHeader struct {
	Fid int64 `json:"fid"`
}

type ConfNewactAward struct {
	Fid int64 `json:"fid"`
}

type ConfNewactStatement struct {
	Fid  int64 `json:"fid"`
	Type int64 `json:"type"` //文本类型：1 任务玩法；2 规则说明；3 平台免责
}

type ConfBottomButton struct {
	Image            string   `json:"image"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	Areas            []*Areas `json:"areas"`
	Bar              string   `json:"bar"`
	Attribute        int64    `json:"attribute"`
	DisplayType      int      `json:"display_type"`      //0:无要求 1.解锁后展示
	DisplayCondition int      `json:"display_condition"` //0:无要求 1:时间 2.预约数据源
	Stime            int64    `json:"stime"`             //时间类型 开始时间
}
