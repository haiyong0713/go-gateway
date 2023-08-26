package show

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// WebRcmdAuth  .
type WebRcmdAuth struct {
	ID int64 `json:"auth" form:"id"`
}

// WebRcmd  web
type WebRcmd struct {
	ID             int64       `json:"id" form:"id"`
	CardType       int         `json:"card_type" form:"card_type"`
	CardValue      string      `json:"card_value" form:"card_value"`
	Stime          xtime.Time  `json:"stime" form:"stime"`
	Etime          xtime.Time  `json:"etime" form:"etime"`
	Partition      string      `json:"partition" form:"partition"`
	PartitionNames interface{} `json:"partition_names,omitempty" gorm:"-"`
	Tag            string      `json:"tag" form:"tag"`
	TagNames       interface{} `json:"tag_names,omitempty" gorm:"-"`
	Avid           string      `json:"avid" form:"avid"`
	Check          int         `json:"check" form:"check"`
	Priority       int         `json:"priority" form:"priority"`
	Person         string      `json:"person" form:"person"`
	ApplyReason    string      `json:"apply_reason" form:"apply_reason"`
	Deleted        int         `json:"deleted" form:"deleted"`
	Card           interface{} `json:"card" gorm:"-"`
	BvID           string      `json:"bvid" gorm:"-"`
	Order          int         `json:"order"`
	Image          string      `json:"image" gorm:"-"`
	BvidRelate     string      `json:"bvid_relate" gorm:"-"`
	RoleId         int         `json:"group_id" gorm:"role_id"`
	GroupName      string      `json:"group_name" gorm:"-"`
}

// WebRcmdPager .
type WebRcmdPager struct {
	Item []*WebRcmd  `json:"item"`
	Page common.Page `json:"page"`
}

// TableName .
func (a WebRcmd) TableName() string {
	return "web_rcmd"
}

/*
---------------------------
 struct param
---------------------------
*/

// WebRcmdAP add param
type WebRcmdAP struct {
	ID          int64      `json:"id" form:"id"`
	CardType    int        `json:"card_type" form:"card_type" validate:"required"`
	CardValue   string     `json:"card_value" form:"card_value" validate:"required"`
	Stime       xtime.Time `json:"stime" form:"stime" validate:"required"`
	Etime       xtime.Time `json:"etime" form:"etime" validate:"required"`
	Priority    int        `json:"priority" form:"priority" validate:"required"`
	Partition   string     `json:"partition" form:"partition"`
	Tag         string     `json:"tag" form:"tag"`
	Avid        string     `json:"avid" form:"avid"`
	Check       int        `form:"check" default:"1"`
	Person      string     `json:"person" form:"person"`
	ApplyReason string     `json:"apply_reason" form:"apply_reason"`
	Order       int        `json:"order" form:"order" validate:"required"`
	//nolint:staticcheck
	RoleId int `json:"role_id" form:"role_id" form:"role_id"`
}

// WebRcmdUP update param
type WebRcmdUP struct {
	ID          int64      `form:"id" validate:"required"`
	CardType    int        `json:"card_type" form:"card_type"`
	CardValue   string     `json:"card_value" form:"card_value"`
	Stime       xtime.Time `json:"stime" form:"stime"`
	Etime       xtime.Time `json:"etime" form:"etime"`
	Check       int        `json:"check" form:"check"`
	Priority    int        `json:"priority" form:"priority"`
	Partition   string     `json:"partition" form:"partition"`
	Tag         string     `json:"tag" form:"tag"`
	Avid        string     `json:"avid" form:"avid"`
	Person      string     `json:"person" form:"person" gorm:"-"`
	ApplyReason string     `json:"apply_reason" form:"apply_reason"`
	Order       int        `json:"order" form:"order" validate:"required"`
}

// WebRcmdLP list param
type WebRcmdLP struct {
	ID        string `form:"id"`
	Check     int    `form:"check"`
	Person    string `form:"person"`
	STime     string `form:"stime"`
	ETime     string `form:"etime"`
	Ps        int    `form:"ps" default:"20"`
	Pn        int    `form:"pn" default:"1"`
	CardType  int    `form:"card_type"`
	Partition string `form:"partition"`
	Tag       string `form:"tag"`
	Avid      string `form:"avid"`
	GroupID   []int  `form:"group_id,split"`
}

// WebRcmdOption option web card (online,hidden,pass,reject)
type WebRcmdOption struct {
	ID    int64 `form:"id" validate:"required"`
	Check int   `form:"check" json:"check"`
}

// TableName .
func (a WebRcmdOption) TableName() string {
	return "web_rcmd"
}

// TableName .
func (a WebRcmdAP) TableName() string {
	return "web_rcmd"
}

// TableName .
func (a WebRcmdUP) TableName() string {
	return "web_rcmd"
}
