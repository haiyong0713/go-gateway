package show

import (
	"go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	WEB_RCMD_RE_TYPE_URL = 1
)

// WebRcmdCard web recommand card
type WebRcmdCard struct {
	ID      int64     `form:"id" gorm:"column:id" json:"id"`
	Type    int64     `form:"type" gorm:"column:type" json:"type"`
	Title   string    `form:"title" gorm:"column:title" json:"title"`
	Desc    string    `form:"desc" gorm:"column:desc" json:"desc"`
	Cover   string    `form:"cover" gorm:"column:cover" json:"cover"`
	ReType  int64     `form:"re_type" gorm:"column:re_type" json:"re_type"`
	ReValue string    `form:"re_value" gorm:"column:re_value" json:"re_value"`
	Person  string    `form:"person" gorm:"column:person" json:"person"`
	Deleted int64     `form:"deleted" gorm:"column:deleted" json:"deleted"`
	Ctime   time.Time `form:"string" gorm:"column:ctime" json:"ctime"`
	Mtime   time.Time `form:"string" gorm:"column:mtime" json:"mtime"`
	Image   string    `form:"image" gorm:"column:-" json:"image"`
}

// WebRcmdCardPager .
type WebRcmdCardPager struct {
	Item []*WebRcmdCard `json:"item"`
	Page common.Page    `json:"page"`
}

// TableName .
func (a WebRcmdCard) TableName() string {
	return "web_rcmd_card"
}

/*
---------------------------
 struct param
---------------------------
*/

// WebRcmdCardAP web card add param
type WebRcmdCardAP struct {
	Type    int64  `form:"type" gorm:"column:type" json:"type"`
	Title   string `form:"title" gorm:"column:title" json:"title"`
	Desc    string `form:"desc" gorm:"column:desc" json:"desc"`
	Cover   string `form:"cover" gorm:"column:cover" json:"cover"`
	ReType  int64  `form:"re_type" gorm:"column:re_type" json:"re_type"`
	ReValue string `form:"re_value" gorm:"column:re_value" json:"re_value"`
	Person  string `form:"person" gorm:"column:person" json:"person"`
}

// WebRcmdCardUP web card update param
type WebRcmdCardUP struct {
	ID      int64  `form:"id" gorm:"column:id" json:"id"`
	Type    int64  `form:"type" gorm:"column:type" json:"type"`
	Title   string `form:"title" gorm:"column:title" json:"title"`
	Desc    string `form:"desc" gorm:"column:desc" json:"desc"`
	Cover   string `form:"cover" gorm:"column:cover" json:"cover"`
	ReType  int64  `form:"re_type" gorm:"column:re_type" json:"re_type"`
	ReValue string `form:"re_value" gorm:"column:re_value" json:"re_value"`
}

// WebRcmdCardLP  web card list param
type WebRcmdCardLP struct {
	ID     int    `form:"id"`
	Person string `form:"person"`
	Title  string `form:"title"`
	Ps     int    `form:"ps" default:"20"` // 分页大小
	Pn     int    `form:"pn" default:"1"`  // 第几个分页
	STime  string `form:"stime"`
	ETime  string `form:"etime"`
}

// TableName .
func (a WebRcmdCardAP) TableName() string {
	return "web_rcmd_card"
}

// TableName .
func (a WebRcmdCardUP) TableName() string {
	return "web_rcmd_card"
}
