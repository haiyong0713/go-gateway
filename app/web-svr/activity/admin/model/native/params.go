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
	Validity    int32   `form:"validity"`                                    //上榜有效期
	ValidStime  int64   `form:"valid_stime"`                                 //开始上榜时间
	SquareTitle string  `form:"square_title"`                                //广场标题
	SmallCard   string  `form:"small_card"`                                  //广场小卡
	BigCard     string  `form:"big_card"`                                    //广场大卡
	Tids        []int64 `form:"tids,split" validate:"min=1,max=5,dive,min=0` //话题活动ids
}

type AddPageParam struct {
	Title      string `form:"title" validate:"required"`
	UserName   string `form:"user_name" validate:"required"`
	Type       int    `form:"type" validate:"min=1,max=8"`
	RelatedUid int64  `form:"related_uid" default:"0" validate:"min=0"`
	ActType    int    `form:"act_type" default:"0" validate:"min=0"`
	Validity   int32  `form:"validity"`    //上榜有效期
	ValidStime int64  `form:"valid_stime"` //开始上榜时间
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
	Validity     int32   `form:"validity"`                                    //上榜有效期
	ValidStime   int64   `form:"valid_stime"`                                 //开始上榜时间
	SquareTitle  string  `form:"square_title"`                                //广场标题
	SmallCard    string  `form:"small_card"`                                  //广场小卡
	BigCard      string  `form:"big_card"`                                    //广场大卡
	Tids         []int64 `form:"tids,split" validate:"min=1,max=5,dive,min=0` //话题活动ids
}

// SearchParam .
type SearchParam struct {
	PageParam
	BeginTime string `form:"begin_time"`
	EndTime   string `form:"end_time"`
	Pn        int    `form:"pn" default:"1"`
	Ps        int    `form:"ps" default:"20" validate:"min=1,max=50"`
	Ptypes    []int  `form:"ptypes,split" default:"1" validate:"min=1,max=50,dive,min=0"`
	States    []int  `form:"states,split" default:"0,1,2"`
}

type UpParam struct {
	Title string `form:"title"`
	Uid   int64  `form:"uid"`
	Pn    int    `form:"pn" default:"1"`
	Ps    int    `form:"ps" default:"20" validate:"min=1,max=50"`
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
	UID     int64  `json:"uid"`
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	TagID   int64  `json:"tag_id"`
	PageID  int64  `json:"page_id"`
	TagType int32  `json:"tag_type"`
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
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime       time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (PageDyn) TableName() string {
	return "native_page_dyn"
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
	Oid    string `form:"oid" validate:"required"` //ts_id
	Pid    int64  `form:"pid" validate:"required"`
	State  int    `form:"state"`
	Reason string `form:"reason"`
}

type SaveModuleReply struct {
	Ver string `json:"ver"`
}
