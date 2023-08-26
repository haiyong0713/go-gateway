package reserve

import (
	xtime "go-common/library/time"
)

const (
	_ = iota
	CounterNodeStateNormal
	CounterNodeStateDelete
)

const (
	_ = iota
	CounterGroupStateNormal
	CounterGroupStateDelete
)

const (
	_ = iota
	CounterGroupDim1All
	CounterGroupDim1Member
	CounterGroupDim1Personal
)

const (
	_ = iota
	NotifyTypeEmail
	NotifyTypeDatabus
)

// ParamList .
type ParamList struct {
	Sid int64 `form:"sid" validate:"min=1"`
	Mid int64 `form:"mid"`
	Pn  int   `form:"pn" default:"1" validate:"min=1"`
	Ps  int   `form:"ps" default:"15" validate:"min=1"`
}

type ParamAddReserve struct {
	Num int   `form:"num" validate:"min=1"`
	Sid int64 `form:"sid" validate:"min=1"`
	Mid int64 `form:"mid" validate:"min=1"`
}

type ParamImportReserve struct {
	Num      int      `form:"num" validate:"min=1"`
	Sid      int64    `form:"sid" validate:"min=1"`
	Username []string `form:"username,split" validate:"required"`
}

type ParamReserveScoreUpdate struct {
	Score int32 `form:"score"`
	Sid   int64 `form:"sid" validate:"min=1"`
	Mid   int64 `form:"mid" validate:"min=1"`
}

type ParamReserveNotifyDelete struct {
	Sid      int64  `form:"sid" validate:"min=1"`
	Author   string `form:"author" validate:"required,lt=16"`
	NotifyID string `form:"notify_id" validate:"required"`
}

type ParamReserveNotifyUpdate struct {
	Sid    int64  `form:"sid" validate:"min=1"`
	Author string `form:"author" validate:"required,lt=16"`
	Notify string `form:"notify" validate:"required"`
}

type ReplyAddReserve struct {
	ID int64 `json:"id"`
}
type ActReserve struct {
	ID          int64      `json:"id" gorm:"column:id"`
	Sid         int64      `json:"sid" gorm:"column:sid"`
	Mid         int64      `json:"mid" gorm:"column:mid"`
	Num         int32      `json:"num" gorm:"column:num"`
	Score       int32      `json:"score" gorm:"column:score"`
	AdjustScore int32      `json:"adjust_score" gorm:"column:adjust_score"`
	State       int        `json:"state" gorm:"column:state"`
	Ctime       xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime       xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
	From        string     `json:"from" gorm:"column:from"`
	Oid         string     `json:"oid" gorm:"column:oid"`
	Typ         string     `json:"typ" gorm:"column:typ"`
	Platform    string     `json:"platform" gorm:"column:platform"`
	Mobiapp     string     `json:"mobiapp" gorm:"column:mobiapp"`
	Buvid       string     `json:"buvid" gorm:"column:buvid"`
	Spmid       string     `json:"spmid" gorm:"column:spmid"`
}

// ListReply .
type ListReply struct {
	List  []*ActReserve `json:"list"`
	Count int64         `json:"count"`
}

type NodeListReply struct {
	Count int64       `json:"count"`
	Pn    int         `json:"pn"`
	Ps    int         `json:"ps"`
	List  []*NodeItem `json:"list"`
}

type GroupListReply struct {
	Count int64        `json:"count"`
	Pn    int          `json:"pn"`
	Ps    int          `json:"ps"`
	List  []*GroupItem `json:"list"`
}

type GroupItem struct {
	ID          int64      `json:"id" form:"id" gorm:"column:id"`
	Sid         int64      `json:"sid" form:"sid" gorm:"column:sid" validate:"min=1"`
	GroupName   string     `json:"group_name" form:"group_name" gorm:"column:group_name" validate:"required"`
	Dim1        int32      `json:"dim1" form:"dim1" gorm:"column:dim1" validate:"min=1"`
	Dim2        int32      `json:"dim2" form:"dim2" gorm:"column:dim2" validate:"min=1"`
	Threshold   int64      `json:"threshold" form:"threshold" gorm:"column:threshold"`
	CounterInfo string     `json:"counter_info" form:"counter_info" gorm:"column:counter_info" validate:"required"`
	Author      string     `json:"author" form:"author" gorm:"column:author" validate:"required"`
	State       uint8      `json:"state" form:"state" gorm:"column:state"`
	Ext         string     `json:"ext" form:"ext" gorm:"column:ext"`
	Ctime       xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime       xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
}

func (GroupItem) TableName() string {
	return "act_subject_counter_group"
}

type NodeItem struct {
	ID       int64      `json:"id" form:"id" gorm:"column:id"`
	Sid      int64      `json:"sid" gorm:"column:sid"`
	GroupID  int64      `json:"group_id" gorm:"column:group_id"`
	NodeName string     `json:"node_name" form:"node_name" gorm:"column:node_name" validate:"required"`
	NodeVal  int64      `json:"node_val" form:"node_val" gorm:"column:node_val"`
	State    uint8      `json:"state" form:"state" gorm:"column:state" validate:"required"`
	Ctime    xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime    xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
}

func (NodeItem) TableName() string {
	return "act_subject_counter_node"
}

// SearchReply .
type SearchReply struct {
	Count      int64   `json:"count"`
	Pn         int     `json:"pn"`
	Ps         int     `json:"ps"`
	ReserveNum int64   `json:"reserve_num"`
	List       []*Item `json:"list"`
}

type Item struct {
	*ActReserve
	Account    *AccInfo `json:"account"`
	TotalScore int32    `json:"total_score"`
}

type AccInfo struct {
	Name string `json:"name"`
	Mid  int64  `json:"mid"`
	Face string `json:"face"`
}

type ParamCounterGroupUpdate struct {
	GroupItem
	NodeStr string      `form:"nodes"`
	Nodes   []*NodeItem `json:"nodes"`
}

type Ext struct {
	DownStream `json:"down_stream"`
}

type DownStream struct {
	Switch bool   `json:"switch"`
	Type   int64  `json:"type"`
	Value  string `json:"value"`
}

const (
	NotifyActionTypePointsUnlockGroupNotifyChannelTypeDatabus     = 2
	NotifyActionTypePointsUnlockGroupNotifyTypeTotalDiffAndKeyMid = 3
	NotifyActionTypePointsUnlockGroupNotifyTypeEachLargerThan     = 4
)
