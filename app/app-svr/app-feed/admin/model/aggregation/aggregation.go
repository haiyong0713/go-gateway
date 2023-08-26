package aggregation

import (
	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/time"
)

const (
	AggreAuditNum = 0
	AggPassNum    = 1
	AggRefuseNum  = 2
	AggOfflineNum = 3
	AggDelNum     = 4
	AggAdd        = "add"
	AggEdit       = "up" // 更新
	AggDel        = "del"
	AggRefuse     = "refuse"   // 拒绝
	AggOffline    = "offline"  // 下线
	AggPass       = "pass"     // 通过审核
	AggreAudit    = "re_audit" // 重新审核
	Ctime         = "ctime"
	Asc           = "asc"
	Desc          = "desc"
	DefaultState  = 1
	BlockState    = 2 // 禁用
)

// AggPub  aggregation pub system .
type AggPub struct {
	ID       int64  `form:"id"        json:"id"        gorm:"column:id" `
	HotTitle string `form:"hot_title" json:"hot_title" validate:"required"  gorm:"column:hot_title"`
	Title    string `form:"title"     json:"title"     validate:"required"  gorm:"column:title"`
	SubTitle string `form:"sub_title"  json:"sub_title" validate:"required"  gorm:"column:subtitle"`
	Image    string `form:"image"     json:"image"     validate:"required"  gorm:"column:image" `
	State    int    `json:"state" gorm:"column:state"`
}

// AggSaveReq aggregation request param .
type AggSaveReq struct {
	AggPub
	TagID []int64 `form:"tag_id,split"`
}

// AggListReq aggregation request param .
type AggListReq struct {
	ID       int64  `form:"id"`
	HotTitle string `form:"hot_title"`
	State    int    `form:"state"`
	TagID    int64  `form:"tag_id"`
	TagName  string `form:"tag_name"`
	Sort     string `form:"sort" default:"desc"`
	Order    string `form:"order" default:"ctime"`
	Ps       int    `form:"ps" default:"20"`
	Pn       int    `form:"pn" default:"1"`
}

// AggListReply aggregation reply data .
type AggListReply struct {
	Items []*AggList `json:"items"`
	Pager PagerCfg   `json:"pager"`
}

// AggList .
type AggList struct {
	AggPub
	Ctime time.Time `json:"ctime"`
	Tag   []*Tag    `json:"tag,omitempty"`
}

type Tag struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

// AggTag .
type AggTag struct {
	HotwordID int64 `gorm:"column:hotword_id"`
	TagID     int64 `gorm:"column:tag_id"`
	State     int   `gorm:"column:state"`
}

// PagerCfg .
type PagerCfg struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// Views .
type Views struct {
	TagID   int64   `json:"tag_id"`
	TagIDs  []int64 `json:"-"`
	TagName string  `json:"tag_name"`
	AvID    int64   `json:"avid"`
	Title   string  `json:"title"`
	MID     int64   `json:"mid"`
	UpName  string  `json:"up_name"`
	BvID    string  `json:"bvid,omitempty"`
	State   int     `json:"state"`
}

// ViewsReply .
type ViewsReply struct {
	Items []*Views `json:"items"`
	Total int      `json:"total"`
	Tags  []*Tag   `json:"tags"`
}

// Aggregation def.
type Aggregation struct {
	AggPub
	H5Url string `json:"h5_url"`
}

// CardList AI return .
type CardList struct {
	ID         int64            `json:"id"`
	Goto       string           `json:"goto"`
	FromType   string           `json:"from_type"`
	Desc       string           `json:"desc"`
	CornerMark int8             `json:"corner_mark"`
	CoverGif   string           `json:"cover_gif"`
	Condition  []*CardCondition `json:"condition"`
	Tag        string           `json:"tag"`
	TagIDs     []int64          `json:"tag_id"`
}

// CardCondition .
type CardCondition struct {
	Plat      int8   `json:"plat"`
	Condition string `json:"conditions"`
	Build     int    `json:"build"`
}

// PopChannelResourceAD
type HotwordAggResource struct {
	Oid       int64 `json:"oid" form:"oid" gorm:"column:oid"`
	TagID     int64 `json:"tag_id" form:"tag_id"`
	HotwordID int64 `json:"hotword_id" form:"hotword_id"`
	Deleted   int   `json:"deleted" form:"deleted"`
	State     int   `json:"state" form:"state"`
}

// ToView .
func (v *Views) ToView(arc *api.Arc) {
	v.Title = arc.Title
	v.MID = arc.Author.Mid
	v.AvID = arc.Aid
	v.UpName = arc.Author.Name
}

func (*AggPub) TableName() string {
	return "hotword_aggregation"
}

func (*AggTag) TableName() string {
	return "hotword_aggregation_tag"
}

func (*HotwordAggResource) TableName() string {
	return "hotword_aggregation_resource"
}

func FilterDupIDs(val []int64) []int64 {
	var (
		exist = make(map[int64]struct{})
		res   = make([]int64, 0)
	)
	for _, v := range val {
		if _, ok := exist[v]; !ok {
			exist[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
