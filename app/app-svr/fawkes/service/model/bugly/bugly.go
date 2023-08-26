package bugly

import (
	xtime "time"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

// PageInfo struct.
type PageInfo struct {
	Total int `json:"total"`
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
}

type CrashIndex struct {
	AppKey                       string `json:"app_key"`
	Count                        int64  `json:"count"`
	DistinctBuvidCount           int64  `json:"distinct_buvid_count"`
	ErrorStackHashWithoutUseless string `json:"error_stack_hash_without_useless"`
	ErrorStackBeforeHash         string `json:"error_stack_before_hash"`
	AnalyseErrorStack            string `json:"analyse_error_stack"`
	ErrorType                    string `json:"error_type"`
	ErrorMsg                     string `json:"error_msg"`
	AssignOperator               string `json:"assign_operator"`
	SolveStatus                  int64  `json:"solve_status"`
	SolveOperator                string `json:"solve_operator"`
	SolveVersionCode             int64  `json:"solve_version_code"`
	SolveDescription             string `json:"solve_description"`
	HappenNewestVersionCode      int64  `json:"happen_newest_version_code"`
	HappenOldestVersionCode      int64  `json:"happen_oldest_version_code"`
	HappenTime                   int64  `json:"happen_time"`
	CTime                        int64  `json:"ctime"`
	MTime                        int64  `json:"mtime"`
}

type CrashIndexRes struct {
	Items    []*CrashIndex `json:"items"`
	PageInfo *PageInfo     `json:"page"`
}

type JankIndex struct {
	Command                 string  `json:"command"`
	Count                   int64   `json:"count"`
	MidDistinctCount        int64   `json:"distinct_mid_count"`
	AppKey                  string  `json:"app_key"`
	BuvidDistinctCount      int64   `json:"distinct_buvid_count"`
	AnalyseJankStack        string  `json:"analyse_jank_stack"`
	AnalyseJankStackHash    string  `json:"analyse_jank_stack_hash"`
	DurationQuantile80      float64 `json:"quantile_80_duration"`
	SolveStatus             int64   `json:"solve_status"`
	SolveOperator           string  `json:"solve_operator"`
	SolveVersionCode        int64   `json:"solve_version_code"`
	SolveDescription        string  `json:"solve_description"`
	HappenNewestVersionCode int64   `json:"happen_newest_version_code"`
	HappenOldestVersionCode int64   `json:"happen_oldest_version_code"`
	HappenTime              int64   `json:"happen_time"`
	CTime                   int64   `json:"ctime"`
	MTime                   int64   `json:"mtime"`
}

type JankIndexRes struct {
	Items    []*JankIndex `json:"items"`
	PageInfo *PageInfo    `json:"page"`
}

type OOMIndex struct {
	Command                 string  `json:"command"`
	Count                   int64   `json:"count"`
	MidDistinctCount        int64   `json:"distinct_mid_count"`
	AppKey                  string  `json:"app_key"`
	BuvidDistinctCount      int64   `json:"distinct_buvid_count"`
	AnalyseStack            string  `json:"analyse_stack"`
	Hash                    string  `json:"hash"`
	RetainedSizeQuantile80  float64 `json:"quantile_80_retained_size"`
	LeakReason              string  `json:"leak_reason"`
	GcRoot                  string  `json:"gc_root"`
	SolveStatus             int64   `json:"solve_status"`
	SolveOperator           string  `json:"solve_operator"`
	SolveVersionCode        int64   `json:"solve_version_code"`
	SolveDescription        string  `json:"solve_description"`
	HappenNewestVersionCode int64   `json:"happen_newest_version_code"`
	HappenOldestVersionCode int64   `json:"happen_oldest_version_code"`
	HappenTime              int64   `json:"happen_time"`
	CTime                   int64   `json:"ctime"`
	MTime                   int64   `json:"mtime"`
}

type OOMIndexRes struct {
	Items    []*OOMIndex `json:"items"`
	PageInfo *PageInfo   `json:"page"`
}

type CrashInfo struct {
	EventId                      string        `json:"event_id"`                         // 埋点唯一标识符
	TimeStamp                    int64         `json:"timestamp"`                        // 埋点生成时间
	TimeISO                      int64         `json:"time_iso"`                         // 埋点上报时间
	IP                           string        `json:"ip"`                               // 用户设备ip
	Mid                          string        `json:"mid"`                              // 用户UID
	BuildId                      string        `json:"build_id"`                         // 构建号
	Buvid                        string        `json:"buvid"`                            // 设备号
	Brand                        string        `json:"brand"`                            // 品牌
	DeviceId                     string        `json:"device_id"`                        // 设备硬件ID
	Uid                          string        `json:"uid"`                              // uid
	Chid                         string        `json:"chid"`                             // 发布渠道ID
	Model                        string        `json:"model"`                            // 手机型号
	Fts                          int64         `json:"fts"`                              // App首次运行时间
	Network                      int64         `json:"network"`                          // 网络状况 未授权:-1. UNKNOWN:0, WIFI:1, 移动网络:2, 未连接:3, 其他网络:4, 以太网:5
	Oid                          string        `json:"oid"`                              // 运营商ID
	AppId                        int32         `json:"app_id"`                           // App类型
	Version                      string        `json:"version"`                          // App版本
	VersionCode                  int64         `json:"version_code"`                     // 版本号
	CrashVersion                 string        `json:"crash_version"`                    // 崩溃版本
	Platform                     int64         `json:"platform"`                         // 设备类型 iPhone:1, iPad:2 Android:3 WP:4
	Osvr                         string        `json:"osver"`                            // 设备系统版本
	FfVersion                    string        `json:"ff_version"`                       // ff版本
	ConfigVersion                string        `json:"config_version"`                   // config版本
	Abi                          string        `json:"abi"`                              // abi
	AppKey                       string        `json:"app_key"`                          // mobi_app 应用唯一标识
	Rate                         float64       `json:"rate"`                             // 采样率[0,1]
	Country                      string        `json:"country"`                          // 国家
	Province                     string        `json:"province"`                         // 省份
	City                         string        `json:"city"`                             // 城市
	Isp                          string        `json:"isp"`                              // 运营商
	InternalVersion              int64         `json:"internal_version"`                 // 内部版本号
	Process                      string        `json:"process"`                          // 进程
	Thread                       string        `json:"thread"`                           // 线程
	CrashType                    int32         `json:"crash_type"`                       // 崩溃类型：android java:0, native:2, ANR:4, iOS objective_c:10, c:11, cpp:12, Hybrid Fultter:1000
	ErrorType                    string        `json:"error_type"`                       // 错误类型
	ErrorMsg                     string        `json:"error_msg"`                        // 错误信息
	ErrorStack                   string        `json:"error_stack"`                      // 错误堆栈
	LastActivity                 string        `json:"last_activity"`                    // 最后记录界面
	TopActivity                  string        `json:"top_activity"`                     // 顶部记录界面
	CallStack                    string        `json:"call_stack"`                       // 调用堆栈
	Macho                        string        `json:"macho"`                            // macho文件信息
	AllMacho                     string        `json:"all_macho"`                        // macho系统文件信息
	AnalyseErrorCode             int32         `json:"analyse_error_code"`               // 解析后的错误码
	AnalyseErrorStack            string        `json:"analyse_error_stack"`              // 解析后的错误堆栈
	ErrorStackHashWithoutUseless uint64        `json:"error_stack_hash_without_useless"` // hash前的堆栈
	IsHarmony                    int64         `json:"is_harmony"`                       // 是否是鸿蒙标识
	MemFree                      string        `json:"mem_free"`                         // 可用内存大小
	StorageFree                  string        `json:"storage_free"`                     // 可用存储空间
	SdcardFree                   string        `json:"sdcard_free"`                      // 可用sd卡大小
	CrashTime                    int64         `json:"crash_time"`                       // app 崩溃发生时间
	LifeTime                     int64         `json:"lifetime"`                         // app 奔溃前的运行时间
	Laser                        *appmdl.Laser `json:"laser"`                            // laser
	Manufacturer                 string        `json:"manufacturer"`                     // 设备厂商
	DomesticRomVer               string        `json:"domestic_rom_ver"`                 // 国产rom版本号
}

type CrashRes struct {
	Items    []*CrashInfo `json:"items"`
	PageInfo *PageInfo    `json:"page"`
}

type JankInfo struct {
	EventId              string  `json:"event_id"`
	TimeStamp            int64   `json:"timestamp"`
	TimeISO              int64   `json:"time_iso"`
	IP                   string  `json:"ip"`
	Mid                  string  `json:"mid"`
	BuildId              string  `json:"build_id"`
	Buvid                string  `json:"buvid"`
	Brand                string  `json:"brand"`
	DeviceId             string  `json:"device_id"`
	Uid                  string  `json:"uid"`
	Chid                 string  `json:"chid"`
	Model                string  `json:"model"`
	Fts                  int64   `json:"fts"`
	Network              int64   `json:"network"`
	Oid                  string  `json:"oid"`
	AppId                int32   `json:"app_id"`
	Version              string  `json:"version"`
	VersionCode          int64   `json:"version_code"`
	Platform             int64   `json:"platform"`
	Osvr                 string  `json:"osver"`
	FfVersion            string  `json:"ff_version"`
	ConfigVersion        string  `json:"config_version"`
	Abi                  string  `json:"abi"`
	AppKey               string  `json:"app_key"`
	Rate                 float64 `json:"rate"`
	Country              string  `json:"country"`
	Province             string  `json:"province"`
	City                 string  `json:"city"`
	Isp                  string  `json:"isp"`
	InternalVersion      int64   `json:"internal_version"`
	Process              string  `json:"process"`
	Thread               string  `json:"thread"`
	StacktraceCount      string  `json:"stacktrace_count"`
	Duration             int32   `json:"duration"`
	JankStack            string  `json:"jank_stack"`
	JankStackCountJson   string  `json:"jank_stack_count_json"`
	JankStackMaxCount    int     `json:"jank_stack_max_count"`
	AnalyseJankCode      string  `json:"analyse_jank_code"`
	AnalyseJankStack     string  `json:"analyse_jank_stack"`
	AnalyseJankStackHash string  `json:"analyse_jank_stack_hash"`
	Route                string  `json:"route"`
	IsHarmony            int64   `json:"is_harmony"`
	LifeTime             int64   `json:"lifetime"`
}

type JankInfoRes struct {
	Items    []*JankInfo `json:"items"`
	PageInfo *PageInfo   `json:"page"`
}

type CrashLaserRel struct {
	Id                           int64      `json:"id"`
	ErrorStackHashWithoutUseless uint64     `json:"error_stack_hash_without_useless"`
	LaserId                      int64      `json:"laser_id"`
	Operator                     string     `json:"operator"`
	Ctime                        xtime.Time `json:"ctime"`
	Mtime                        xtime.Time `json:"mtime"`
}

type OOMInfo struct {
	EventId         string  `json:"event_id"`
	TimeStamp       int64   `json:"timestamp"`
	TimeISO         int64   `json:"time_iso"`
	IP              string  `json:"ip"`
	Mid             string  `json:"mid"`
	BuildId         string  `json:"build_id"`
	Buvid           string  `json:"buvid"`
	Brand           string  `json:"brand"`
	DeviceId        string  `json:"device_id"`
	Uid             string  `json:"uid"`
	Chid            string  `json:"chid"`
	Model           string  `json:"model"`
	Fts             int64   `json:"fts"`
	Network         int64   `json:"network"`
	Oid             string  `json:"oid"`
	AppId           int32   `json:"app_id"`
	Version         string  `json:"version"`
	VersionCode     int64   `json:"version_code"`
	Platform        int64   `json:"platform"`
	Osvr            string  `json:"osver"`
	FfVersion       string  `json:"ff_version"`
	ConfigVersion   string  `json:"config_version"`
	Abi             string  `json:"abi"`
	AppKey          string  `json:"app_key"`
	Rate            float64 `json:"rate"`
	Country         string  `json:"country"`
	Province        string  `json:"province"`
	City            string  `json:"city"`
	Isp             string  `json:"isp"`
	InternalVersion int64   `json:"internal_version"`
	Process         string  `json:"process"`
	Thread          string  `json:"thread"`
	LastActivity    string  `json:"last_activity"`
	TopActivity     string  `json:"top_activity"`
	AppMemory       int64   `json:"app_memory"`
	AppMemoryRate   string  `json:"app_memory_rate"`
	DeviceRam       int64   `json:"device_ram"`
	DumpTime        int64   `json:"dump_time"`
	FileSize        int64   `json:"file_size"`
	FileUrl         string  `json:"file_url"`
	SessionId       string  `json:"session_id"`
	Hash            string  `json:"hash"`
	Stack           string  `json:"stack"`
	AnalyseStack    string  `json:"analyse_stack"`
	InstanceCount   int32   `json:"instance_count"`
	LeakReason      string  `json:"leak_reason"`
	GcRoot          string  `json:"gc_root"`
	Signature       string  `json:"signature"`
	RetainedSize    int32   `json:"retained_size"`
	Path            string  `json:"path"`
}

type OOMInfoRes struct {
	Items    []*OOMInfo `json:"items"`
	PageInfo *PageInfo  `json:"page"`
}

type IndexStatus struct {
	AppKey           string `json:"app_key"`
	Hash             string `json:"hash"`
	SolveStatus      int64  `json:"solve_status"`
	AssignOperator   string `json:"assign_operator"`
	SolveOperator    string `json:"solve_operator"`
	SolveVersionCode int64  `json:"solve_version_code"`
	SolveDescription string `json:"solve_description"`
}

type LogText struct {
	AppKey   string `json:"app_key"`
	Hash     string `json:"hash"`
	Operator string `json:"operator"`
	LogText  string `json:"log_text"`
	Ctime    int64  `json:"ctime"`
	Mtime    int64  `json:"mtime"`
}

const (
	ErrorStackHashWithoutUseless = "error_stack_hash_without_useless_v2"
	AnalyseJankStackHash         = "analyse_jank_stack_hash"
)
