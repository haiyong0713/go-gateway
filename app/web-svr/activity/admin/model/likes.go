package model

import (
	xtime "go-common/library/time"
)

// LikesParam def.
type LikesParam struct {
	Sid          int64  `form:"sid" validate:"min=1"`
	Mid          int64  `form:"mid"`
	Wid          int64  `form:"wid"`
	Page         int    `form:"page" default:"1" validate:"min=1"`
	PageSize     int    `form:"pagesize" default:"15" validate:"min=1"`
	IDGt         int    `form:"id_gt"`
	IDLt         int    `form:"id_lt"`
	Order        string `form:"order" default:"id"`
	Sort         string `form:"sort" default:"asc"`
	States       []int  `form:"state,split"`
	IgnoreObject bool   `form:"ignore_object"`
}

// AddLikes .
type AddLikes struct {
	DealType string `form:"deal_type" validate:"required"`
	Wid      int64  `form:"wid"`
	Sid      int64  `form:"sid" validate:"min=1"`
	Type     int    `form:"type"`
	Mid      int64  `form:"mid"`
	State    int    `form:"state"`
	Plat     int    `form:"plat"`
	Device   int    `form:"device"`
}

// UpReply .
type UpReply struct {
	State int     `form:"state"`
	Reply string  `form:"reply"`
	IDs   []int64 `form:"ids,split" validate:"min=1,max=50"`
}

// UpWid .
type UpWid struct {
	Sid   int64 `form:"sid" validate:"min=1"`
	Wid   int64 `form:"wid" validate:"min=1"`
	State int   `form:"state"`
}

// BatchLike .
type BatchLike struct {
	Sid  int64   `form:"sid" validate:"min=1"`
	Wid  []int64 `form:"wid,split" validate:"min=1,max=200,dive,gt=0"`
	Mid  int64   `form:"mid"`
	Type int     `form:"type"`
}

// AddPic .
type AddPic struct {
	Sid     int64  `form:"sid" validate:"min=1"`
	Type    int    `form:"type"`
	Mid     int64  `form:"mid"`
	Wid     int64  `form:"wid"`
	Plat    int    `form:"plat"`
	Device  int    `form:"device"`
	Image   string `form:"image"`
	Message string `form:"message" validate:"required,max=450,min=1"`
	Link    string `form:"link"`
}

// UpLike .
type UpLike struct {
	Type     int    `form:"type"`
	Mid      int64  `form:"mid"`
	Wid      int64  `form:"wid"`
	State    int    `form:"state"`
	StickTop int    `form:"stick_top"`
	Lid      int64  `form:"lid" validate:"min=1"`
	Message  string `form:"message"`
	Reply    string `form:"reply"`
	Link     string `form:"link"`
	Image    string `form:"image"`
}

// ActivityAVInfo active_id -> avid
type ActivityAVInfo struct {
	ActivityID int64 `json:"mission_id"`
	AVID       int64 `json:"id"`
	MID        int64 `json:"mid"`
	Category   int   `json:"typeid"`
	TagID      int64 `json:"-"`
	Ratio      int   `json:"-"`
}

// LikeContent def
type LikeContent struct {
	ID      int64      `json:"id" form:"id" gorm:"column:id"`
	Message string     `json:"message" form:"message"`
	IP      int64      `json:"ip" form:"ip" gorm:"column:ip"`
	Plat    int        `json:"plat" form:"plat"`
	Device  int        `json:"device" form:"device"`
	Ctime   xtime.Time `json:"ctime" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime   xtime.Time `json:"mtime" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Image   string     `json:"image" form:"image"`
	Reply   string     `json:"reply" form:"reply"`
	Link    string     `json:"link" form:"link"`
	ExName  string     `json:"ex_name" form:"ex_name"`
	IPv6    []byte     `json:"ipv6" gorm:"column:ipv6"`
}

// ActLikeLog def
type ActLikeLog struct {
	ID    int64      `json:"id" form:"id" gorm:"column:id"`
	Lid   int64      `json:"lid" form:"lid" gorm:"column:lid"`
	User  string     `json:"user" form:"user" gorm:"column:user"`
	State int64      `json:"state" form:"state" gorm:"column:state"`
	Ctime xtime.Time `json:"ctime" form:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime xtime.Time `json:"mtime" form:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// LikesRes .
type LikesRes struct {
	Likes map[int64]*Like `json:"likes"`
	PageRes
}

// MusicRes .
type MusicRes struct {
	Code int `json:"code"`
	Data map[int64]struct {
		CoverURL  string   `json:"coverUrl"`
		Duration  string   `json:"duration"`
		Categorie string   `json:"categorie"`
		Intro     string   `json:"intro"`
		Mid       int64    `json:"mid"`
		Title     string   `json:"title"`
		SongID    int64    `json:"songId"`
		PlayURL   []string `json:"playUrl"`
	}
}

// LetterParam for private msg param
type LetterParam struct {
	RecverIDs  []uint64 `json:"recver_ids"`       //多人消息，列表型，限定每次客户端发送<=100
	SenderUID  uint64   `json:"sender_uid"`       //官号uid：发送方uid
	MsgKey     uint64   `json:"msg_key"`          //消息唯一标识
	MsgType    int32    `json:"msg_type"`         //文本类型 type = 1
	Content    string   `json:"content"`          //{"content":"test" //文本内容}
	NotifyCode string   `json:"notify_code"`      //通知码
	Params     string   `json:"params,omitempty"` //逗号分隔，通知卡片内容的可配置参数
	JumpUri    string   `json:"jump_uri"`         //通知卡片跳转链接
	Title      string   `json:"title"`
	Text       string   `json:"text"`
	JumpText   string   `json:"jump_text"`
}

// LikeContentNew
type LikeContentNew struct {
	ID      int64      `json:"id" form:"id" gorm:"column:id"`
	Message string     `json:"message" form:"message"`
	IP      int64      `json:"ip" form:"ip" gorm:"column:ip"`
	Plat    int        `json:"plat" form:"plat"`
	Device  int        `json:"device" form:"device"`
	Ctime   xtime.Time `json:"ctime" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime   xtime.Time `json:"mtime" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Image   string     `json:"image" form:"image"`
	Reply   string     `json:"reply" form:"reply"`
	Link    string     `json:"link" form:"link"`
	ExName  string     `json:"ex_name" form:"ex_name"`
	IPv6    []byte     `json:"ipv6" gorm:"column:ipv6"`
}

// TableName ActLikeLog def
func (ActLikeLog) TableName() string {
	return "act_like_log"
}

// TableName LikeContent def
func (LikeContent) TableName() string {
	return "like_content"
}

// TableName LikeContentNew
func (LikeContentNew) TableName() string {
	return "like_content_new"
}
