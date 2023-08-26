package show

import (
	"go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// PopRecommend popular recommend
type PopRecommend struct {
	ID        int64     `json:"id" form:"id"`
	CardValue string    `json:"card_value" form:"card_value"`
	Title     string    `json:"title" gorm:"-"`
	Label     string    `json:"label" form:"label"`
	Person    string    `json:"person" form:"person"`
	Deleted   int       `json:"deleted" form:"deleted"`
	Ctime     time.Time `json:"ctime" gorm:"-"`
	CoverGif  string    `json:"cover_gif" form:"cover_gif"`
	Bvid      string    `json:"bvid,omitempty" form:"-"`
}

// PopRecommendPager .
type PopRecommendPager struct {
	Item []*PopRecommend `json:"item"`
	Page common.Page     `json:"page"`
}

// TableName .
func (a PopRecommend) TableName() string {
	return "popular_recommend"
}

/*
---------------------------
 struct param
---------------------------
*/

// PopRecommendAP popular recommend add param
type PopRecommendAP struct {
	CardValue string `json:"card_value" form:"card_value" validate:"required"`
	Label     string `json:"label" form:"label"`
	Person    string `json:"person" form:"person"`
	CoverGif  string `json:"cover_gif" form:"cover_gif"`
}

// PopRecommendUP popular recommend update param
type PopRecommendUP struct {
	ID        int64  `form:"id" validate:"required"`
	CardValue string `json:"card_value" form:"card_value" validate:"required"`
	Label     string `json:"label" form:"label"`
	CoverGif  string `json:"cover_gif" form:"cover_gif"`
}

// PopRecommendLP event topic list param
type PopRecommendLP struct {
	ID     string `form:"id"`
	Person string `form:"person"`
	Ps     int    `form:"ps" default:"20"` // 分页大小
	Pn     int    `form:"pn" default:"1"`  // 第几个分页
	AID    int64
}

// TableName .
func (a PopRecommendAP) TableName() string {
	return "popular_recommend"
}

// TableName .
func (a PopRecommendUP) TableName() string {
	return "popular_recommend"
}
