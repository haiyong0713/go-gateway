package apm

import (
	"encoding/json"
	"errors"
)

const (
	CLICKHOUSE_WEB_TRACK_PV    = "fawkes.pv"
	CLICKHOUSE_WEB_TRACK_ERROR = "fawkes.error"
)

type MatchOption struct {
	EventID int64 `form:"event_id"`

	// 查询算子关键字
	QueryKeys string `form:"query_keys"`
	FilterStr string `form:"filters"`
	Filters   []*Filter

	// 基础字段
	Mid           string `form:"mid"`
	Buvid         string `form:"buvid"`
	AppKey        string `form:"app_key"`
	IntervalTime  string `form:"interval_time"`
	Country       string `form:"country"`
	Province      string `form:"province"`
	City          string `form:"city"`
	Isp           string `form:"isp"`
	Model         string `form:"upper(model)"`
	Brand         string `form:"upper(brand)"`
	Network       string `form:"network"`
	VersionCode   string `form:"version_code"`
	OSVer         string `form:"osver"`
	TunnelSDK     string `form:"tunnel_sdk"`
	Platform      string `form:"platform"`
	Oid           string `form:"oid"`
	StartTime     int64  `form:"start_time"`
	EndTime       int64  `form:"end_time"`
	In            string `form:"in"`
	OrderBy       string `form:"order_by"`
	Limit         int    `form:"limit"`
	Pn            int    `form:"pn"`
	Ps            int    `form:"ps"`
	FFVersion     string `form:"ff_version"`
	ConfigVersion string `form:"config_version"`

	// 业务扩展字段
	Command                      string `form:"command"`
	Domain                       string `form:"domain"`
	HTTPCode                     string `form:"http_code"`
	BizCode                      string `form:"biz_code"`
	RealRequestUrl               string `form:"real_request_url"`
	NegotiatedProtocol           string `form:"negotiated_protocol"`
	InternetProtocolVersion      string `form:"internet_protocol_version"`
	ResponseCode                 string `form:"response_code"`
	ErrorType                    string `form:"error_type"`
	ErrorMessage                 string `form:"error_msg"`
	Process                      string `form:"process"`
	Thread                       string `form:"thread"`
	LastActivity                 string `form:"last_activity"`
	TopActivity                  string `form:"top_activity"`
	CrashType                    string `form:"crash_type"`
	ErrorStack                   string `form:"error_stack"`
	AnalyseErrorStack            string `form:"analyse_error_stack"`
	AnalyseJankStack             string `form:"analyse_jank_stack"`
	ErrorStackHashWithoutUseless string `form:"error_stack_hash_without_useless"`
	AnalyseJankStackHash         string `form:"analyse_jank_stack_hash"`
	JankStackMaxCount            string `form:"jank_stack_max_count"`
	Hash                         string `form:"hash"`
	Chid                         string `form:"chid"`

	// 自定义埋点
	StatusCode      string `form:"status_code"`
	ExternalNumber1 string `form:"external_num1"`
	ExternalNumber2 string `form:"external_num2"`
	ExternalNumber3 string `form:"external_num3"`
	ExternalNumber4 string `form:"external_num4"`
	GroupKey        string `form:"group_key"`

	NameFrom     string `form:"name_from"`
	NameTo       string `form:"name_to"`
	RealNameFrom string `form:"real_name_from"`

	// 前端独有字段
	Href string `form:"href"`

	// 独立计算函数专用
	CalculateArgs string `form:"calculate_args"`
}

func (matchOption *MatchOption) Check() (err error) {
	//if matchOption.AppKey == "" {
	//	return ecode.Error(ecode.RequestErr, "app_key 不能为空")
	//}
	if matchOption.FilterStr != "" {
		if err = json.Unmarshal([]byte(matchOption.FilterStr), &matchOption.Filters); err != nil {
			return err
		}
	}
	// event_id 为必传字段
	if matchOption.EventID == 0 {
		err = errors.New("event_id 不能为空")
		return
	}
	// 开始时间不能为空
	if matchOption.StartTime == 0 {
		err = errors.New("start_time异常")
		return
	}
	// 时间间隔默认5min
	if matchOption.IntervalTime == "" {
		matchOption.IntervalTime = "5 minute"
	}
	// 默认 count() 倒序
	if matchOption.OrderBy == "" {
		matchOption.OrderBy = "count() DESC"
	}
	if matchOption.Pn == 0 {
		matchOption.Pn = 1
	}
	if matchOption.Ps == 0 || matchOption.Ps > 100 {
		matchOption.Ps = 100
	}
	// Limit 默认为100 （ 旧逻辑. 待废弃 ）
	if matchOption.Limit == 0 || matchOption.Limit > 100 {
		matchOption.Limit = 100
	}
	return
}

type Filter struct {
	AndType   string `json:"and_type"` // AND OR
	Column    string `json:"column"`
	EqualType string `json:"equal_type"` // = != < <= > >= NULL NOT NULL LIKE
	Values    string `json:"values"`
	ValueType string `json:"value_type"`
}

// fawkes 前端埋点基础字段
type WebTrackParams struct {
	Models []*WebTrackModel `json:"models"`
}

// WebTrackModel
type WebTrackModel struct {
	// 基础字段
	EventId           string `json:"event_id"`
	AppKey            string `json:"app_key"`
	Username          string `json:"username"`
	Timestamp         int64  `json:"timestamp"`
	BowerName         string `json:"browser_name"`
	BowerCode         string `json:"browser_code"`
	BowerVersion      string `json:"browser_version"`
	NavigatorPlatform string `json:"navigator_platform"`

	// pv字段
	RoutePath string `json:"route_path"`
	RouteName string `json:"route_name"`
	RouteFrom string `json:"route_from"`

	// 异常埋点
	ErrorMessage string `json:"error_msg"`
}

// 通用模型
type Moni struct {
	Title interface{} `json:"title"`
	Value interface{} `json:"value"`
}

// 网络基础数据
type NetInfo struct {
	Command                 string  `json:"command"`
	Count                   int64   `json:"count"`
	TotalTimeQuantile80     float64 `json:"total_time_quantile_80"`
	TotalTimeQuantile95     float64 `json:"total_time_quantile_95"`
	NetSuccessRate          float64 `json:"http_success_rate"`
	NetSuccessRateDowngrade float64 `json:"http_success_rate_downgrade"`
	NetBizSuccessRate       float64 `json:"http_biz_success_rate"`
	ReqSizeAvg              float64 `json:"req_size_avg"`
	RecvSizeAvg             float64 `json:"recv_size_avg"`
}

// 通用数量模型
type CountInfo struct {
	Command            string `json:"command"`
	Count              int64  `json:"count"`
	MidDistinctCount   int64  `json:"distinct_mid_count"`
	BuvidDistinctCount int64  `json:"distinct_buvid_count"`
}

// 通用数量模型
type MetricInfo struct {
	Command string      `json:"command"`
	Value   interface{} `json:"value"`
}

// 自定义事件
type StatisticsInfo struct {
	Command interface{} `json:"command"`
	Count   interface{} `json:"count"`
	Num1    interface{} `json:"num1"`
	Num2    interface{} `json:"num2"`
	Num3    interface{} `json:"num3"`
	Num4    interface{} `json:"num4"`
	Num5    interface{} `json:"num5"`
}

// 流量表
type FlowmapRoute struct {
	NameFrom string  `json:"name_from"`
	NameTo   string  `json:"name_to"`
	Count    int     `json:"count"`
	Memory   float64 `json:"memory"`
}

// 流量表 - 别名配置
type FlowmapRouteAlias struct {
	ID         int    `json:"id"`
	AppKey     string `json:"app_key"`
	BusID      int    `json:"bus_id"`
	RouteName  string `json:"route_name"`
	RouteAlias string `json:"route_alias"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
	Operator   string `json:"operator"`
}

const SecToMilliUnit = 1000
