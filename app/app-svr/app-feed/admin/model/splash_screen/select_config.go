package splash_screen

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/api"
)

// Category 自选闪屏分类
type Category struct {
	ID        int64      `json:"id" form:"id" gorm:"column:id"`
	Name      string     `json:"name" form:"name" gorm:"column:category_name" validate:"required"`
	Sort      int64      `json:"sort" form:"sort" gorm:"column:sort"`
	IsDeleted int        `json:"-" form:"-" gorm:"column:is_deleted"`
	CUser     string     `json:"c_user" form:"-" gorm:"column:cuser"`
	MUser     string     `json:"m_user" form:"-" gorm:"column:muser"`
	CTime     xtime.Time `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime     xtime.Time `json:"mtime" form:"-" gorm:"column:mtime"`
}

func (*Category) TableName() string {
	return "splash_screen_select_category"
}

type CategoryWithConfigCount struct {
	*Category
	ConfigCount int32 `json:"config_count" form:"-" gorm:"column:config_count"`
}

type CategoryForSaving struct {
	*Category
	CTime xtime.Time `json:"-" form:"-" gorm:"-"`
	CUser string     `json:"-" form:"-" gorm:"-"`
	MTime xtime.Time `json:"-" form:"-" gorm:"-"`
	MUser string     `json:"-" form:"-" gorm:"-"`
}

// SelectConfig 自选闪屏配置
type SelectConfig struct {
	ID          int64      `json:"id" form:"id" gorm:"column:id"`
	CategoryIDs []int64    `json:"category_ids" form:"category_ids" gorm:"-"`
	ImageID     int64      `json:"image_id" form:"image_id" gorm:"column:image_id"`
	STime       xtime.Time `json:"stime" form:"stime" gorm:"column:stime"`
	ETime       xtime.Time `json:"etime" form:"etime" gorm:"column:etime"`
	Sort        int64      `json:"sort" form:"sort" gorm:"column:sort"`
	ShowSort    int64      `json:"show_sort" form:"-" gorm:"-"`
	// state 表示状态：
	// 0、待通过
	// 1、待生效
	// 2、已失效
	// 3、生效中
	State      api.SplashScreenConfigState_Enum       `json:"state" form:"-" gorm:"-"`
	AuditState api.SplashScreenConfigAuditStatus_Enum `json:"audit_state" form:"audit_state" gorm:"column:audit_state"`
	IsDeleted  int                                    `json:"-" form:"-" gorm:"column:is_deleted"`
	CUser      string                                 `json:"c_user" form:"-" gorm:"column:cuser"`
	MUser      string                                 `json:"m_user" form:"-" gorm:"column:muser"`
	CTime      xtime.Time                             `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime      xtime.Time                             `json:"mtime" form:"-" gorm:"column:mtime"`
}

type SelectConfigForSaving struct {
	*SelectConfig
	State      api.SplashScreenConfigState_Enum       `json:"-" form:"-" gorm:"-"`
	AuditState api.SplashScreenConfigAuditStatus_Enum `json:"-" form:"-" gorm:"-"`
	CTime      xtime.Time                             `json:"-" form:"-" gorm:"-"`
	CUser      string                                 `json:"-" form:"-" gorm:"-"`
	MTime      xtime.Time                             `json:"-" form:"-" gorm:"-"`
	MUser      string                                 `json:"-" form:"-" gorm:"-"`
}

func (*SelectConfig) TableName() string {
	return "splash_screen_select_config"
}

// SelectConfigsWithPager 自选配置展示结构
type SelectConfigsWithPager struct {
	Items []*SelectConfig `json:"items"`
	Pager *Pager          `json:"pager"`
}

// SelectConfigSortSeq 自选闪屏配置sort字段seq
type SelectConfigSortSeq struct {
	ID    int64      `json:"id" form:"id" gorm:"column:id"`
	CTime xtime.Time `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime xtime.Time `json:"mtime" form:"-" gorm:"column:mtime"`
}

func (*SelectConfigSortSeq) TableName() string {
	return "splash_screen_select_config_sort_seq"
}

// SelectConfigCategoryRel 自选闪屏配置与分类关系
type SelectConfigCategoryRel struct {
	ID         int64      `json:"id" form:"id" gorm:"column:id"`
	ConfigID   int64      `json:"config_id" form:"config_id" gorm:"column:config_id"`
	CategoryID int64      `json:"category_id" form:"category_id" gorm:"column:category_id"`
	IsDeleted  int        `json:"-" form:"-" gorm:"column:is_deleted"`
	CUser      string     `json:"c_user" form:"-" gorm:"column:cuser"`
	MUser      string     `json:"m_user" form:"-" gorm:"column:muser"`
	CTime      xtime.Time `json:"ctime" form:"-" gorm:"column:ctime"`
	MTime      xtime.Time `json:"mtime" form:"-" gorm:"column:mtime"`
}

func (*SelectConfigCategoryRel) TableName() string {
	return "splash_screen_select_config_category_rel"
}

type SelectConfigCategoryRelWithImageID struct {
	*SelectConfigCategoryRel
	ImageID int64 `json:"image_id" form:"-" gorm:"column:image_id"`
}
