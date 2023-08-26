package ci

import (
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/model"
)

const (
	GitTypeBranch = 0
	GitTypeTag    = 1
	GitTypeCommit = 2
)

// 是否推送CD
const (
	DidPush = 1
	NotPUsh = 0
)

// ci构建状态:-2 取消,-1失败,1等待,2打包中,3成功
const (
	CICancel    = -2
	CIFailed    = -1
	CIInWaiting = 1
	CIBuilding  = 2
	CISuccess   = 3
)

const (
	TrackLogID     = "018190" // ci编译上报Lancer的logId
	TrackSeparator = "|"      // 数据分隔符
)

type (
	PackType   int
	NotifyMail int8
)

// pack type
const (
	Debug       PackType = 1  // Debug包
	Release     PackType = 2  // Release包
	Enter       PackType = 3  // 企业包
	Publish     PackType = 4  // AppStore包
	FastDebug   PackType = 5  // Debug包 - 快速编译
	EnterDebug  PackType = 6  // 企业包 - DEBUG宏
	FastRelease PackType = 7  // Release包 - 快速编译
	CoverDebug  PackType = 8  // Debug包 - 代码覆盖率专用
	TestFlight  PackType = 9  // 内测包 - TestFlight
	AppBundle   PackType = 10 // App Bundle包
	Assets      PackType = 11 // 资源包
)

var packTypeDict = map[PackType]string{
	Debug:       "Debug包",
	Release:     "Release包",
	Enter:       "企业包",
	Publish:     "AppStore包",
	FastDebug:   "Debug包快速编译",
	EnterDebug:  "企业包Debug",
	FastRelease: "Release包快速编译",
	CoverDebug:  "Debug包代码覆盖率专用",
	TestFlight:  "内测包TestFlight",
	AppBundle:   "App_Bundle包",
	Assets:      "Assets资源",
}

func (p PackType) name(packType PackType) string {
	if packName, ok := packTypeDict[packType]; !ok {
		return "未知包类型"
	} else {
		return packName
	}
}

const (
	NotifyMailDisable   = NotifyMail(0) // mail不通知
	NotifyMailRecipient = NotifyMail(1) // mail只通知收件人
	NotifyMailCC        = NotifyMail(2) // mail通知cc
)

const (
	MetaKeyCompatible  = "compatibleVersions"
	MetaKeyFeatures    = "features"
	MetaKeyFeatureName = "featureName"
)

const SendJsonDefault = "{}" // send字段为空时，默认发送的格式

type Feature struct {
	Name              string
	CompatibleVersion int64
}

const Comma = ","

// BuildPack struct.
type BuildPack struct {
	BuildID             int64               `json:"id"`
	AppID               string              `json:"app_id"`
	AppKey              string              `json:"app_key"`
	GitlabProjectID     string              `json:"-"`
	GitPath             string              `json:"-"`
	GitlabJobID         int64               `json:"gl_job_id"`
	DepGitlabJobID      int64               `json:"dep_gl_job_id"`
	GitlabJobURL        string              `json:"job_url"`
	GitType             int8                `json:"git_type"`
	GitName             string              `json:"git_name"`
	Commit              string              `json:"commit"`
	ShortCommit         string              `json:"short_commit"`
	PkgType             int8                `json:"pkg_type"`
	Version             string              `json:"version"`
	VersionCode         int64               `json:"version_code"`
	InternalVersionCode int64               `json:"internal_version_code"`
	Operator            string              `json:"operator"`
	Size                int64               `json:"size"`
	Md5                 string              `json:"md5"`
	PkgPath             string              `json:"pkg_path"`
	PkgURL              string              `json:"pkg_url"`
	BbrURL              string              `json:"bbr_url"`
	MappingURL          string              `json:"mapping_url"`
	RURL                string              `json:"r_url"`
	RMappingURL         string              `json:"r_mapping_url"`
	Status              int8                `json:"status"`
	TestStatus          int64               `json:"test_status"`
	TaskIds             string              `json:"task_ids"`
	DidPush             int8                `json:"did_push"`
	ChangeLog           string              `json:"change_log"`
	EnvVars             string              `json:"env_var"`
	Description         string              `json:"description"`
	NotifyGroup         int8                `json:"notify_group"`
	SubRepos            []*BuildPackSubRepo `json:"sub_repos"`
	BuildStartTime      xtime.Time          `json:"build_start_time"`
	BuildEndTime        xtime.Time          `json:"build_end_time"`
	CTime               int64               `json:"ctime"`
	MTime               int64               `json:"mtime"`
	IsExpired           int                 `json:"is_expired"`
	IsCompatible        int                 `json:"is_compatible"` // Deprecated:tribe feature api之后这个字段已无法表示兼容关系
	WebhookURL          string              `json:"webhook_url"`
	Features            string              `json:"features"`
	CIEnvVars           string              `json:"ci_env_vars"`
	Send                string              `json:"send"`
}

// Page struct.
type Page struct {
	Total    int `json:"total"`
	PageNum  int `json:"pn"`
	PageSize int `json:"ps"`
}

// ResultBuildPacks struct.
type ResultBuildPacks struct {
	PageInfo *Page        `json:"page,omitempty"`
	Items    []*BuildPack `json:"items,omitempty"`
}

// BuildEnvs struct.
type BuildEnvs struct {
	ID         int64  `json:"id"`
	EnvKey     string `json:"env_key"`
	EnvVal     string `json:"env_val"`
	EnvType    int    `json:"type"`
	Descrition string `json:"description"`
	IsDefault  int    `json:"is_default"`
	IsGlobal   int    `json:"is_global"`
	PushCDAble int    `json:"push_cd_able"`
	Platform   string `json:"platform"`
	AppKeys    string `json:"app_keys"`
	Operator   string `json:"operator"`
	Mtime      int64  `json:"mtime"`
	Ctime      int64  `json:"ctime"`
}

// EnvValue struct.
type EnvValue struct {
	EnvVal      string `json:"env_val"`
	Description string `json:"description"`
	IsDefault   int    `json:"is_default"`
	IsGlobal    int    `json:"is_global"`
	PushCDAble  int    `json:"push_cd_able"`
	Platform    string `json:"platform"`
	AppKeys     string `json:"app_keys"`
	ID          int64  `json:"id"`
}

// EPMonkey struct
type EPMonkey struct {
	ID           int64  `json:"id"`
	AppKey       string `json:"app_key"`
	BuildID      int64  `json:"build_id"`
	ExecDuration int    `json:"exec_duration"`
	OSVer        string `json:"osver"`
	LogUrl       string `json:"log_url"`
	PlayUrl      string `json:"play_url"`
	Status       string `json:"status"`
	SchemeUrl    string `json:"scheme_url"`
	MessageTo    string `json:"message_to"`
	Operator     string `json:"operator"`
	Mtime        int64  `json:"mtime"`
	Ctime        int64  `json:"ctime"`
}

// Jekins 请求体
type EPMonkeyRequstBody struct {
	Ref          string `json:"ref"`
	MobEnv       string `json:"MOB_ENV"`
	MobAndroidOS string `json:"MOB_ANDROID_OS"`
	BundleID     string `json:"BUNDLE_ID"`
	AppKey       string `json:"APP_KEY"`
	ApkURL       string `json:"APK_URL"`
	MappingURL   string `json:"MAPPING_URL"`
	ExecDuration string `json:"EXEC_DURATION"`
	MonoDuration string `json:"MONO_DURATION"`
	Schemes      string `json:"SCHEMES"`
	CC           string `json:"CC"`
	CallbackURL  string `json:"CALLBACK_URL"`
	HookID       string `json:"FAWKES_HOOK_ID"`
}

// Jekins 执行回调
type EPMonkeyCallbackBody struct {
	AppKey      string                 `json:"app_key"`
	ExternalID  string                 `json:"external_id"`
	Status      int                    `json:"status"`
	Emulator    map[string][]string    `json:"emulator"`
	PipelineUrl string                 `json:"pipeline_url"`
	ReportUrl   string                 `json:"report_url"`
	Advanced    map[string]interface{} `json:"advanced"`
}

// ci job 记录参数
type JobRecordParam struct {
	BuildID      int64      `form:"build_id"`
	JobID        int64      `form:"job_id"`
	PipelineID   int64      `form:"pipeline_id"`
	JobStartTime xtime.Time `form:"job_start_time"`
	JobEndTime   xtime.Time `form:"job_end_time"`
	PkgVersion   int64      `form:"pkg_version"`
	JobStatus    int        `form:"job_status"`
	JobName      string     `form:"job_name"`
	AppKey       string     `form:"app_key"`
	Stage        string     `form:"stage"`
	TagList      string     `form:"tag_list"`
	RunnerInfo   string     `form:"runner_info"`
}

// CICompileRecordParam ci编辑上报接口参数模型
type CICompileRecordParam struct {
	AppKey            string     `form:"app_key"`
	PkgType           int        `form:"pkg_type"`
	BuildEnv          int        `form:"build_env"`
	BuildLogURL       string     `form:"build_log_url"`
	JobID             int64      `form:"job_id"`
	Status            int        `form:"status"`
	StepsCount        int        `form:"steps_count"`
	UptodateCount     int        `form:"uptodate_count"`
	CacheCount        int        `form:"cache_count"`
	ExecutedCount     int        `form:"executed_count"`
	FastTotal         int        `form:"fast_total"`
	FastRemote        int        `form:"fast_remote"`
	FastLocal         int        `form:"fast_local"`
	AfterSyncTask     int        `form:"after_sync_task"`
	BuildSourceLocal  int        `form:"build_source_local"`
	BuildSourceRemote int        `form:"build_source_remote"`
	OptimizeLevel     int        `form:"optimize_level"`
	Operator          string     `form:"operator"`
	StartTime         xtime.Time `form:"start_time"`
	EndTime           xtime.Time `form:"end_time"`
}

// ci job 记录参数
type BuildPackSubRepo struct {
	SubRepoID  int64      `json:"sub_repo_id"`
	AppKey     string     `json:"app_key"`
	BuildID    int64      `json:"build_id"`
	PipelineID int64      `json:"pipeline_id"`
	RepoName   string     `json:"repo_name"`
	Commit     string     `json:"commit"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
}

// CISpecifyTimeReq 获取某段时间的包 请求参数
type CISpecifyTimeReq struct {
	StartTime xtime.Time `form:"start_time" json:"start_time"`
	EndTime   xtime.Time `form:"end_time" json:"end_time"`
	AppKey    string     `form:"app_key" json:"app_key"`
	PkgType   []int8     `form:"pkg_type" json:"pkg_type" validate:"required"`
}

// CISpecifyTimeDeleteReq 删除文件封装 请求参数封装
type CISpecifyTimeDeleteReq struct {
	DeleteKeys []*CISpecifyTimeDelete `json:"delete_keys"`
}

// CISpecifyTimeDelete 删除文件字段 请求参数
type CISpecifyTimeDelete struct {
	AppKey  string `form:"app_key" json:"app_key"`
	BuildId int64  `form:"build_id" json:"build_id"`
}

// CISpecifyTimeDeleteResp 删除文件  返回参数
type CISpecifyTimeDeleteResp struct {
	NeedDelete   []int64 `json:"need_delete"`
	BuildIdFail  []int64 `json:"failed_id_list"`
	AffectedRows int64   `json:"affected_rows"`
}

type GetAppBuildPackVersionInfoReq struct {
	AppKey      string  `json:"app_key" form:"app_key" validate:"required"`
	State       int64   `json:"state" form:"state"`
	GitlabJobId []int64 `json:"gitlab_job_id" form:"gitlab_job_id"`
}

type GetAppBuildPackVersionInfoResp struct {
	VersionInfo []*VersionInfo `json:"items" form:"items"`
}

type VersionInfo struct {
	Version     string   `json:"version" form:"version"`
	VersionCode int64    `json:"version_code" form:"version_code"`
	GitlabJobId int64    `json:"gitlab_job_id" form:"gitlab_job_id"`
	IsPushCD    int8     `json:"is_push_cd" form:"is_push_cd"`
	Env         []string `json:"env" form:"env"`
	SteadyState int8     `json:"steady_state" form:"steady_state"`
}

type BBRItem struct {
	Name        string `json:"name,omitempty"`
	Id          string `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// HookParam hook参数
type HookParam struct {
	AppKey      string `json:"app_key"`
	AppName     string `json:"app_name"`
	BuildID     int64  `json:"build_id"`
	GitlabJobID int64  `json:"gl_job_id"`
	CTime       int64  `json:"ctime"`
	ConfigName  string `json:"config_name,omitempty"`
	BizType     string `json:"biz_type,omitempty"`
	PackURL     string `json:"pack_url"`
}

// NotifyCI CI通知
type NotifyCI struct {
	BuildId      int64
	Receiver     *NotifyCIReceiver // 接收对象
	NotifyMail   NotifyMail        // mail通知开关
	IsNotifyBot  bool              // bot通知开关
	IsNotifyHook bool              // hook通知开关
	IsNotifyUser bool              // user通知开关
}

// NotifyCIReceiver CI通知接收者
type NotifyCIReceiver struct {
	Users   string      `json:"username"`
	Bots    string      `json:"bots"`
	Webhook *model.Hook `json:"hook"`
}

type TrackMessage struct {
	BuildId      int64  `json:"build_id"`      // 构建唯一标识
	HostName     string `json:"host_name"`     // 主机名
	Arch         string `json:"arch"`          // x86_64,arm64电脑架构
	Platform     string `json:"platform" `     // iOS/Android 手机平台
	OSName       string `json:"os_name"`       // 系统名字（Linux, Mac）
	OSVersion    string `json:"os_version"`    // 系统版本
	HardwareInfo string `json:"hardware_info"` // 硬件信息
	EventId      string `json:"event_id"`      // 事件名
	ExtendFields string `json:"extend_fields"` // 扩展字段
	Operator     string `json:"operator"`      // 操作用户
}
