package show

import (
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// EventTopic event topic
type EventTopic struct {
	ID        int64  `json:"id" form:"id"`
	Title     string `json:"title" form:"title"`
	Desc      string `json:"desc" form:"desc"`
	ShowTitle int    `json:"show_title" form:"show_title"`
	Cover     string `json:"cover" form:"cover"`
	CoverType int    `json:"cover_type" form:"cover_type"`
	Retype    int    `json:"re_type" gorm:"column:re_type" form:"re_type"`
	Revalue   string `json:"re_value" gorm:"column:re_value" form:"string"`
	Corner    string `json:"corner" form:"corner"`
	Person    string `json:"person" form:"person"`
	Deleted   int    `json:"deleted" form:"deleted"`
}

// EventTopicPager .
type EventTopicPager struct {
	Item []*EventTopic `json:"item"`
	Page common.Page   `json:"page"`
}

// TableName .
func (a EventTopic) TableName() string {
	return "event_topic"
}

/*
---------------------------
 struct param
---------------------------
*/

// EventTopicAP event topic add param
type EventTopicAP struct {
	Title     string `json:"title" form:"title" validate:"required"`
	Desc      string `json:"desc" form:"desc"`
	ShowTitle int    `json:"show_title" form:"show_title"`
	Cover     string `json:"cover" form:"cover" validate:"required"`
	CoverType int    `json:"cover_type" form:"cover_type"`
	Retype    int    `json:"re_type" form:"re_type" gorm:"column:re_type" validate:"required"`
	Revalue   string `json:"re_value" form:"re_value" gorm:"column:re_value" validate:"required"`
	Corner    string `json:"corner" form:"corner"`
	Person    string `json:"person" form:"person"`
}

// EventTopicUP event topic update param
type EventTopicUP struct {
	ID        int64  `form:"id" validate:"required"`
	Title     string `json:"title" form:"title" validate:"required"`
	Desc      string `json:"desc" form:"desc"`
	ShowTitle int    `json:"show_title" form:"show_title" gorm:"column:show_title"`
	Cover     string `json:"cover" form:"cover" validate:"required"`
	CoverType int    `json:"cover_type" form:"cover_type" gorm:"column:cover_type"`
	Retype    int    `json:"re_type" form:"re_type" gorm:"column:re_type" validate:"required"`
	Revalue   string `json:"re_value" form:"re_value" gorm:"column:re_value" validate:"required"`
	Corner    string `json:"corner" form:"corner" gorm:"column:corner"`
}

// EventTopicLP event topic list param
type EventTopicLP struct {
	ID     int    `form:"id"`
	Person string `form:"person"`
	Title  string `form:"title"`
	Ps     int    `form:"ps" default:"20"` // 分页大小
	Pn     int    `form:"pn" default:"1"`  // 第几个分页
}

// TableName .
func (a EventTopicAP) TableName() string {
	return "event_topic"
}

// TableName .
func (a EventTopicUP) TableName() string {
	return "event_topic"
}
