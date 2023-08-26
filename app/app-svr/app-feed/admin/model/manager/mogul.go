package manager

import (
	"time"

	xtime "go-common/library/time"
)

type AppMogulLogParam struct {
	Mid   int64      `form:"mid" validate:"required"`
	Path  string     `form:"path"`
	Stime xtime.Time `form:"stime"`
	Etime xtime.Time `form:"etime"`
	Pn    int        `form:"pn" default:"1"`  // 第几个分页
	Ps    int        `form:"ps" default:"20"` // 分页大小
}

type AppMogulLogReply struct {
	Items []*AppMogulLog `json:"items"`
	Page  *Page          `json:"page"`
}

// Comment: app网关接口大佬行为日志
type AppMogulLog struct {
	// Comment: 自增ID
	ID int64 `json:"id" gorm:"column:id"`
	// Comment: 用户mid
	Mid int64 `json:"mid" gorm:"column:mid"`
	// Comment: Buvid
	Buvid string `json:"buvid" gorm:"column:buvid"`
	// Comment: 接口请求路径
	Path string `json:"path" gorm:"column:path"`
	// Comment: 接口请求方法
	Method string `json:"method" gorm:"column:method"`
	// Comment: 接口请求header
	Header string `json:"header" gorm:"column:header"`
	// Comment: 接口请求参数
	Param string `json:"param" gorm:"column:param"`
	// Comment: 接口请求body
	Body string `json:"body" gorm:"column:body"`
	// Comment: 接口响应header
	ResponseHeader string `json:"response_header" gorm:"column:response_header"`
	// Comment: 接口响应
	Response string `json:"response" gorm:"column:response"`
	// Comment: 接口HTTP响应状态码
	StatusCode string `json:"status_code" gorm:"column:status_code"`
	// Comment: 接口响应错误码
	ErrCode string `json:"err_code" gorm:"column:err_code"`
	// Comment: 请求时间
	// Default: 0000-00-00 00:00:00
	RequestTime xtime.Time `json:"request_time" gorm:"column:request_time"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime xtime.Time `json:"ctime" gorm:"column:ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime xtime.Time `json:"mtime" gorm:"column:mtime"`
	// Comment: 接口响应时长
	Duration time.Duration `json:"-" gorm:"column:duration"`
	// extra:
	// 可读的接口响应时长
	DurationHuman string `json:"duration_human"`
}

type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
