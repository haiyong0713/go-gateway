package mogul

import (
	"time"

	xtime "go-common/library/time"
)

// Comment: app网关接口大佬行为日志
type AppMogulLog struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 用户mid
	Mid int64 `json:"mid"`
	// Comment: Buvid
	Buvid string `json:"buvid"`
	// Comment: 接口请求路径
	Path string `json:"path"`
	// Comment: 接口请求路径
	Method string `json:"method"`
	// Comment: 接口请求header
	Header string `json:"header"`
	// Comment: 接口请求参数
	Param string `json:"param"`
	// Comment: 接口请求body
	Body string `json:"body"`
	// Comment: 接口响应header
	ResponseHeader string `json:"response_header"`
	// Comment: 接口响应
	Response string `json:"response"`
	// Comment: 接口HTTP响应状态码
	StatusCode string `json:"status_code"`
	// Comment: 接口响应错误码
	ErrCode string `json:"err_code"`
	// Comment: 请求时间
	// Default: 0000-00-00 00:00:00
	RequestTime xtime.Time `json:"request_time"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime xtime.Time `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime xtime.Time `json:"mtime"`
	// Comment: 接口响应时长
	Duration time.Duration `json:"duration"`
}
