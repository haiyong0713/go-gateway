package show

type PopLargeCardRes struct {
	Items []*PopLargeCard `json:"items"`
	Pager PagerCfg        `json:"pager"`
}

type PopLargeCard struct {
	ID         int64  `json:"id" form:"id"`
	Title      string `json:"desc" form:"desc" gorm:"column:title"`
	CardType   string `json:"card_type" form:"-"`
	RID        int64  `json:"rid" form:"-" gorm:"column:rid"`
	Bvid       string `json:"bvid" form:"-"`
	WhiteList  string `json:"white_list" form:"white_list" gorm:"column:white_list"`
	CreateBy   string `json:"create_by" form:"create_by" gorm:"column:create_by"`
	Auto       int64  `json:"auto" form:"auto" gorm:"column:auto"`
	Deleted    int    `json:"deleted" form:"deleted"`
	VideoTitle string `json:"title" form:"-"`
	Author     string `json:"author" form:"-"`
}

// PopLargeCardAD
type PopLargeCardAD struct {
	Title     string `json:"title" form:"title" validate:"title" gorm:"column:title"`
	CardType  string `json:"card_type" form:"card_type" validate:"card_type"`
	RID       int64  `json:"rid" form:"rid" validate:"rid" gorm:"column:rid"`
	WhiteList string `json:"white_list" form:"white_list" validate:"white_list" gorm:"column:white_list"`
	CreateBy  string `json:"create_by" form:"create_by" validate:"create_by" gorm:"column:create_by"`
	Auto      int64  `json:"auto" form:"auto" validate:"auto" gorm:"column:auto"`
	Deleted   int    `json:"deleted" form:"deleted" validate:"required"`
}

// PopLargeCardUP
type PopLargeCardUP struct {
	ID        int64  `form:"id" validate:"required"`
	Title     string `form:"title" validate:"required" gorm:"column:title"`
	CardType  string `form:"card_type" validate:"required"`
	RID       int64  `form:"rid" validate:"required" gorm:"column:rid"`
	WhiteList string `form:"white_list" validate:"required" gorm:"column:white_list"`
	CreateBy  string `form:"create_by" validate:"required" gorm:"column:create_by"`
	Auto      int64  `form:"auto" validate:"required" gorm:"column:auto"`
}

// TableName .
func (a PopLargeCardAD) TableName() string {
	return "popular_large_card"
}

// TableName .
func (a PopLargeCard) TableName() string {
	return "popular_large_card"
}

// TableName .
func (a PopLargeCardUP) TableName() string {
	return "popular_large_card"
}
