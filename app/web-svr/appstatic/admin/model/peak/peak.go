package peak

import (
	"go-common/library/time"
)

var (
	// is_deleted 0-未删除 1-已删除
	NotDeleted uint8
	//deleted
	Deleted uint8 = 1
	// 未上线
	NotOnline uint8
	// 已上线
	Online uint8 = 1
)

type AppVer struct {
	PeakID     uint   `json:"peak_id" form:"peak_id" gorm:"column:peak_id"`          // 错峰任务ID
	Plat       uint8  `json:"plat" form:"plat" gorm:"column:plat"`                   //所属平台 0-安卓
	Conditions string `json:"conditions" form:"conditions" gorm:"column:conditions"` // 判断条件：gt大于、lt小于、eq等于、ne不等于
	Build      string `json:"build" form:"build" gorm:"column:build"`                //版本号
	Deleted    uint8  `json:"is_deleted" form:"is_deleted" gorm:"column:is_deleted"` // 是否已删除
}

type Index struct {
	ID           uint      `json:"id"`            // 错峰任务ID
	Priority     int       `json:"priority"`      // 下载优先级 默认0
	FileName     string    `json:"file_name"`     //业务资源文件名
	Type         int       `json:"type"`          // 1-mod资源
	Url          string    `json:"url"`           // 资源下载链接
	Md5          string    `json:"md5"`           // 资源Md5
	Size         int64     `json:"size"`          // 资源大小
	ExpectDw     int       `json:"expect_dw"`     // 资源下载方式 1 - cdn 2- pcdn
	EffectTime   time.Time `json:"effect_time"`   //资源配置生效时间
	ExpireTime   time.Time `json:"expire_time"`   // 资源配置过期时间
	OnlineStatus uint8     `json:"online_status"` // 上线状态 0-未上线 1-已上线
	Person       string    `json:"person"`
	AppVer       []AppVer  `json:"app_ver"` //应用版本控制
}

type IndexPager struct {
	Item []*Index `json:"item"`
	Page Page     `json:"page"`
}

// Page pager
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// IndexParam Index peak index param
type IndexParam struct {
	ID         string    `json:"id" form:"id"`                   // ID
	EffectTime time.Time `json:"effect_time" form:"effect_time"` // 开始时间
	ExpireTime time.Time `json:"expire_time" form:"expire_time"` // 结束时间
	FileName   string    `json:"file_name" form:"file_name"`
	Url        string    `json:"url" form:"url"`
	Type       int       `json:"type" form:"type" validate:"required"`
	PageSize   int       `json:"page_size" form:"page_size" default:"30"`    // 分页大小
	PageNumber int       `json:"page_number" form:"page_number" default:"1"` // 第几个分页
}

type AddPeakParam struct {
	FileName   string    `json:"file_name" form:"file_name"`           // 业务资源文件名
	Priority   int       `json:"priority" form:"priority" default:"0"` //下载优先级 // ，默认为0
	Type       int       `json:"type" form:"type" default:"1"`         //
	Url        string    `json:"url" form:"url"`                       // 错峰任务类型 1-mod
	Md5        string    `json:"md5" form:"md5"`                       //
	Size       int64     `json:"size" form:"size"`                     //
	ExpectDw   int       `json:"expect_dw" form:"expect_dw"`
	EffectTime time.Time `json:"effect_time" form:"effect_time"`
	ExpireTime time.Time `json:"expire_time" form:"expire_time"`
	AppVer     string    `json:"app_ver" form:"app_ver"` //
	Person     string    `json:"person" form:"person"`
	Uid        int64     `json:"uid" form:"uid"`
}

type UpdatePeakParam struct {
	ID           uint      `form:"id" validate:"required"`
	FileName     string    `json:"file_name" form:"file_name"`           // 业务资源文件名
	Priority     int       `json:"priority" form:"priority" default:"0"` //下载优先级 // ，默认为0
	Type         int       `json:"type" form:"type" default:"1"`         //
	Url          string    `json:"url" form:"url"`                       // 错峰任务类型 1-mod
	Md5          string    `json:"md5" form:"md5"`                       //
	Size         int64     `json:"size" form:"size"`                     //
	ExpectDw     int       `json:"expect_dw" form:"expect_dw"`
	EffectTime   time.Time `json:"effect_time" form:"effect_time"`
	ExpireTime   time.Time `json:"expire_time" form:"expire_time"`
	AppVer       string    `json:"app_ver" form:"app_ver"` //
	OnlineStatus uint8     `json:"online_status" form:"online_status"`
}

type Peak struct {
	ID           uint      `json:"id" form:"id" gorm:"column:id"`
	Priority     int       `json:"priority" form:"priority" gorm:"column:priority"`
	FileName     string    `json:"file_name" form:"file_name" gorm:"column:file_name"`
	Type         int       `json:"type" form:"type" gorm:"column:type"`
	Url          string    `json:"url" form:"url" gorm:"column:url"`
	Md5          string    `json:"md5" form:"md5" gorm:"column:md5"`     //
	Size         int64     `json:"size" form:"size"  gorm:"column:size"` //
	ExpectDw     int       `json:"expect_dw" form:"expect_dw"  gorm:"column:expect_dw"`
	EffectTime   time.Time `json:"effect_time" form:"effect_time"  gorm:"column:effect_time"`
	ExpireTime   time.Time `json:"expire_time" form:"expire_time"  gorm:"column:expire_time"`
	OnlineStatus uint8     `json:"online_status" form:"online_status"  gorm:"column:online_status"`
	IsDeleted    uint8     `json:"is_deleted" form:"is_deleted"  gorm:"column:is_deleted"`
	Person       string    `json:"person" form:"person" gorm:"column:person"`
	Uid          int64     `json:"uid" form:"uid" gorm:"column:uid"`
}

// TableName peak_app_ver 错峰任务版本限制
func (a AppVer) TableName() string {
	return "peak_appver"
}

// TableName peak 错峰任务表
func (a Index) TableName() string {
	return "peak"
}

func (a Peak) TableName() string {
	return "peak"
}
