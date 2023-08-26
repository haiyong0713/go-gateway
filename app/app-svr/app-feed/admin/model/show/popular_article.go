package show

import "time"

type ArticleCardRes struct {
	Items []*ArticleCard `json:"items"`
	Pager PagerCfg       `json:"pager"`
}

type ArticleCard struct {
	ID           int64     `json:"id" form:"id"`
	ArticleID    int64     `json:"article_id" form:"article_id" gorm:"column:article_id" validate:"required"`
	Cover        string    `json:"cover" form:"cover" gorm:"column:cover"`
	CreateBy     string    `json:"create_by" form:"-" gorm:"column:create_by"`
	State        int       `json:"state" form:"state"`
	Mtime        time.Time `json:"-" form:"-"`
	MtimeStr     string    `json:"mtime" form:"-"`
	ArticleTitle string    `json:"article_title" form:"-"`
}

type ArticleCardAD struct {
	ID        int64  `json:"id" form:"id"`
	ArticleID int64  `json:"article_id" form:"article_id" gorm:"column:article_id" validate:"required"`
	Cover     string `json:"cover" form:"cover" gorm:"column:cover"`
	CreateBy  string `json:"create_by" form:"-" gorm:"column:create_by"`
	State     int    `json:"state" form:"state"`
}

type ArticleCardUP struct {
	ID        int64  `form:"id" validate:"required"`
	ArticleID int64  `json:"article_id" form:"article_id" gorm:"column:article_id"`
	Cover     string `json:"cover" form:"cover" gorm:"column:cover"`
	CreateBy  string `json:"create_by" form:"-" gorm:"column:create_by"`
	State     int    `json:"state" form:"state"`
}

// TableName .
func (a ArticleCard) TableName() string {
	return "popular_article_card"
}

// TableName .
func (a ArticleCardAD) TableName() string {
	return "popular_article_card"
}

// TableName .
func (a ArticleCardUP) TableName() string {
	return "popular_article_card"
}
