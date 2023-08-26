package native

import (
	"go-common/library/time"
)

// PageParam .
type PageParam struct {
	ID           int64     `json:"id"`
	Title        string    `form:"title" json:"title"`
	Type         int       `json:"type"`
	State        int       `json:"state"`
	Creator      string    `form:"creator" json:"creator"`
	Operator     string    `json:"operator"`
	Stime        time.Time `json:"stime" gorm:"column:stime" time_format:"2006-01-02 15:04:05"`
	Etime        time.Time `json:"etime" gorm:"column:etime" time_format:"2006-01-02 15:04:05"`
	ShareTitle   string    `json:"share_title"`
	ShareImage   string    `json:"share_image"`
	ShareUrl     string    `json:"share_url"`
	ForeignID    int64     `json:"foreign_id"`
	SkipUrl      string    `json:"skip_url"`
	Spmid        string    `json:"spmid"`
	RelatedUid   int64     `json:"related_uid"`
	ActType      int       `json:"act_type"`
	Hot          int64     `json:"hot"`
	DynamicID    string    `json:"dynamic_id"`
	Attribute    int64     `json:"attribute"`
	PcURL        string    `json:"pc_url"`
	AnotherTitle string    `json:"another_title"`
	ShareCaption string    `json:"share_caption"`
	BgColor      string    `json:"bg_color"`
	FromType     int       `form:"from_type" json:"from_type"`
	ConfSet      string    `json:"conf_set"  gorm:"column:conf_set"`
	FirstPid     int64     `json:"first_pid" gorm:"column:first_pid"`
}

// ModifyParam .
type ModifyParam struct {
	ID          int64   `form:"id" validate:"required,min=1"`
	RelatedUid  int64   `form:"related_uid" default:"0" validate:"min=0"`
	ActType     int     `form:"act_type"`
	Hot         int64   `form:"hot"`
	DynamicID   int64   `form:"dynamic_id"`
	Attribute   int64   `form:"attribute"`
	UserName    string  `form:"user_name" validate:"required"`
	Validity    int32   `form:"validity"`     //上榜有效期
	ValidStime  int64   `form:"valid_stime"`  //开始上榜时间
	SquareTitle string  `form:"square_title"` //广场标题
	SmallCard   string  `form:"small_card"`   //广场小卡
	BigCard     string  `form:"big_card"`     //广场大卡
	Tids        []int64 `form:"tids,split"`   //话题活动ids
	BgColor     string  `form:"bg_color"`     //背景色
	//首页tab相关配置
	BgType         int    `form:"bg_type"`          //背景配置模式 1:颜色 2:图片
	TabTopColor    string `form:"tab_top_color"`    //顶栏头部色值
	TabMiddleColor string `form:"tab_middle_color"` //中间色值
	TabBottomColor string `form:"tab_bottom_color"` //tab栏底部色值
	FontColor      string `form:"font_color"`       //tab文本高亮色值
	BarType        int    `form:"bar_type"`         //系统状态栏色值 0:默认黑色 1:白色
	BgImage1       string `form:"bg_image_1"`       //背景图1
	BgImage2       string `form:"bg_image_2"`       //背景图2
}

type AddPageParam struct {
	Title      string `form:"title" validate:"required"`
	UserName   string `form:"user_name" validate:"required"`
	Type       int    `form:"type" validate:"min=1,max=11"`
	RelatedUid int64  `form:"related_uid" default:"0" validate:"min=0"`
	ActType    int    `form:"act_type" default:"0" validate:"min=0"`
	Validity   int32  `form:"validity"`    //上榜有效期
	ValidStime int64  `form:"valid_stime"` //开始上榜时间
	FirstPid   int64  `form:"first_pid"`   //父id
}

// OnlineParam .
type OnlineParam struct {
	ID    int64 `form:"id" validate:"required,min=1"`
	Stime int64 `form:"stime"`
	Etime int64 `form:"etime"`
}

// EditParam .
type EditParam struct {
	ID           int64   `form:"id"          validate:"required"`
	Stime        int64   `form:"stime"       validate:"required"`
	Etime        int64   `form:"etime"       validate:"required"`
	ShareTitle   string  `form:"share_title" validate:"required"`
	ShareImage   string  `form:"share_image" validate:"required"`
	ShareUrl     string  `form:"share_url"`
	UserName     string  `form:"user_name"   validate:"required"`
	SkipUrl      string  `form:"skip_url"`
	Spmid        string  `form:"spmid"`
	PcUrl        string  `form:"pc_url" validate:"max=255"`
	AnotherTitle string  `form:"another_title" validate:"max=100"`
	ShareCaption string  `form:"share_caption"`
	Attribute    int64   `form:"attribute"`
	Validity     int32   `form:"validity"`     //上榜有效期
	ValidStime   int64   `form:"valid_stime"`  //开始上榜时间
	SquareTitle  string  `form:"square_title"` //广场标题
	SmallCard    string  `form:"small_card"`   //广场小卡
	BigCard      string  `form:"big_card"`     //广场大卡
	Tids         []int64 `form:"tids,split"`   //话题活动ids
	BgColor      string  `form:"bg_color"`     //背景色
	//白名单数据源
	WhiteValue string `form:"white_value"`
}

// SearchParam .
type SearchParam struct {
	PageParam
	ActOrigin string `form:"act_origin"`
	BeginTime string `form:"begin_time"`
	EndTime   string `form:"end_time"`
	Pn        int    `form:"pn" default:"1"`
	Ps        int    `form:"ps" default:"20" validate:"min=1,max=50"`
	Ptypes    []int  `form:"ptypes,split" default:"1" validate:"min=1,max=50,dive,min=0"`
	States    []int  `form:"states,split" default:"0,1,2"`
	FromTypes []int  `form:"from_types,split"`
}

type UpParam struct {
	Title     string `form:"title"`
	Uid       int64  `form:"uid"`
	Pn        int    `form:"pn" default:"1"`
	Ps        int    `form:"ps" default:"20" validate:"min=1,max=50"`
	ActOrigin string `form:"act_origin"`
}

// SearchModule .
type SearchModule struct {
	ID       int64 `form:"id"`
	ModuleID int64 `form:"module_id"`
}

// ModuleRes .
type ModuleRes struct {
	Item *ModuleAll `json:"item"`
}

// ModuleAll .
type ModuleAll struct {
	Click            []*ModuleCli           `json:"click"`
	DynamicExt       []*ModuleDy            `json:"dynamic"`
	VideoExt         []*ModuleVideo         `json:"video"`
	Act              []*ModuleAct           `json:"act"`
	ParticipationExt []*ModuleParticipation `json:"participation"`
}

// ModuleCil .
type ModuleCli struct {
	*ModuleData
	Cli []*Click `json:"click_ext"`
}

// ModuleDy .
type ModuleDy struct {
	*ModuleData
	Dy []*DynamicExt `json:"dynamic_ext"`
}

// ModuleVideo .
type ModuleVideo struct {
	*ModuleData
	Video []*VideoExt `json:"video_ext"`
}

// ModuleAct .
type ModuleAct struct {
	*ModuleData
	Act []*Act `json:"act_ext"`
}

// ModuleParticipation .
type ModuleParticipation struct {
	*ModuleData
	Part []*ParticipationExt `json:"participation_ext"`
}

// ModuleRecommend
type ModuleRecommend struct {
	*ModuleData
}

// SearchRes .
type SearchRes struct {
	Item []*NatPage `json:"item"`
	Page Cfg        `json:"page"`
}

type UpReply struct {
	Item []*UpItem `json:"item"`
	Page Cfg       `json:"page"`
}

type UpItem struct {
	UID       int64  `json:"uid"`
	Name      string `json:"name"`
	TagName   string `json:"tag_name"`
	TagID     int64  `json:"tag_id"`
	PageID    int64  `json:"page_id"`
	TagType   int32  `json:"tag_type"`
	ActOrigin string `json:"act_origin"`
}

// ModuleParam .
type ModuleParam struct {
	NativeID int64  `form:"native_id" validate:"required"`
	Data     string `form:"data"`
	UserName string `form:"user_name"   validate:"required"`
}

// Cfg .
type Cfg struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// FindRes .
type FindRes struct {
	Stime     time.Time `json:"stime"`
	ForeignID int64     `json:"foreign_id"`
	State     int       `json:"state"`
	Type      int       `json:"type"`
}

// AddReID .
type AddReID struct {
	ID int64 `json:"id"`
}

// TableName native_page .
func (PageParam) TableName() string {
	return "native_page"
}

type PageDyn struct {
	ID          int64     `json:"id"`
	Pid         int64     `json:"pid"`
	Validity    int32     `json:"validity"`                                                    //上榜有效期
	Stime       time.Time `json:"stime" gorm:"column:stime" time_format:"2006-01-02 15:04:05"` //开始上榜时间
	SquareTitle string    `json:"square_title"`                                                //广场标题
	SmallCard   string    `json:"small_card"`                                                  //广场小卡
	BigCard     string    `json:"big_card"`                                                    //广场大卡
	Tids        string    `json:"tids"`                                                        //话题活动ids
	Dynamic     string    `json:"dynamic"`
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime       time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (PageDyn) TableName() string {
	return "native_page_dyn"
}

type PageExt struct {
	ID         int64     `json:"id"`
	Pid        int64     `json:"pid"`
	WhiteValue string    `json:"white_value"`
	Ctime      time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime      time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (PageExt) TableName() string {
	return "native_page_ext"
}

type SaveTabReq struct {
	ID            int32  `json:"id"`
	Title         string `json:"title" validate:"required"`
	Stime         int64  `json:"stime"`
	Etime         int64  `json:"etime"`
	BgType        int8   `json:"bg_type" validate:"required"`
	BgImg         string `json:"bg_img"`
	BgColor       string `json:"bg_color"`
	IconType      int8   `json:"icon_type"  validate:"required"`
	ActiveColor   string `json:"active_color"`
	InactiveColor string `json:"inactive_color"`
	TabModules    []*struct {
		ID          int32  `json:"id"`
		Title       string `json:"title" validate:"required"`
		ActiveImg   string `json:"active_img" validate:"required"`
		InactiveImg string `json:"inactive_img" validate:"required"`
		Category    int8   `json:"category" validate:"required"`
		Pid         int32  `json:"pid"`
		Url         string `json:"url"`
		Rank        int8   `json:"rank" validate:"required"`
	} `json:"tab_modules" validate:"required"`
}

type SaveTabRly struct {
	ID int32 `json:"id"`
}

type SearchTabReq struct {
	ID         int32  `form:"id"`
	Title      string `form:"title"`
	Creator    string `form:"creator"`
	CtimeStart int64  `form:"ctime_start"`
	CtimeEnd   int64  `form:"ctime_end"`
	State      int8   `form:"state"`
	Pn         int32  `form:"pn" validate:"required"`
	Ps         int32  `form:"ps" validate:"required"`
}

type SearchTabModuleItem struct {
	TabModule
	TopicName string `json:"topic_name"`
}

type SearchTabItem struct {
	Tab
	TabModules []*SearchTabModuleItem `json:"tab_modules"`
}

type SearchTabRly struct {
	Total int32            `json:"total"`
	List  []*SearchTabItem `json:"list"`
}

// MixtureExt .
type MixtureExt struct {
	ForeignID int64 `json:"foreign_id"`
}

// native_mixture_ext.reason: 播单json
type MixFolder struct {
	Fid         int64        `json:"fid,omitempty"`
	RcmdContent *RcmdContent `json:"rcmd_content,omitempty"` //编辑推荐内容
}

// TsOnlineReq .
type TsOnlineReq struct {
	Oid          string       `form:"oid" validate:"required"` //ts_id
	Pid          int64        `form:"pid" validate:"required"`
	State        int          `form:"state"`
	Reason       string       `form:"reason"`
	AuditTime    string       `form:"audit_time"`
	AuditContent AuditContent `form:"audit_content"`
}

type SaveModuleReply struct {
	Ver string `json:"ver"`
}

type SpaceOfflineReq struct {
	Mid     int64  `form:"mid" validate:"required"`
	PageID  int64  `form:"page_id" validate:"required"`
	TabType string `form:"tab_type" validate:"required"`
}

type TopicUpgradeReq struct {
	Topic      string `form:"topic" validate:"required"`
	Source     string `form:"source" validate:"required"`
	ListAction string `form:"list_action" default:"off"`
}

type TopicUpgradeRly struct {
	ID int64 `json:"id"`
}

type ReserveRly struct {
	SID   int64  `json:"sid,omitempty"`
	Title string `json:"title,omitempty"`
	Type  int32  `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Mid   int64  `json:"mid,omitempty"`
}

type ResourceExt struct {
	ImgUrl string `json:"img_url"`
	Length int64  `json:"length"`
	Width  int64  `json:"width"`
}

type TsPageReq struct {
	TsID   int64 `form:"ts_id" validate:"min=1"`
	PageID int64 `form:"page_id" validate:"min=1"`
}

type TsPageRly struct {
	Dynamic string `json:"dynamic"`
}

type HmtChannelRly struct {
	IDs []*HmtChannel
}
type HmtChannel struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type AddNewactReq struct {
	Sid int64 `form:"sid" validate:"required"`
}

type AddNewactRly struct {
	ID int64 `json:"id"`
}
