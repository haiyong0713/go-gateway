package cd

import (
	"time"

	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

// cd const.
const (
	GenerateGitName  = "keep/build_apk_channels"
	GenerateUpload   = 1
	GenerateTest     = 2
	GeneratePublish  = 3
	GenerateQueue    = -2
	GenerateRunning  = -3
	GenerateFailed   = -4
	GenerateSuccess  = 0
	GenerateUnhandle = -1

	Upgrade    = 0
	NotUpgrade = -1

	PatchGitName        = "keep/build_patch_pipeline"
	PatchStatusFailed   = 0
	PatchStatusNotStart = 1
	PatchStatusProccess = 2
	PatchStatusSuccess  = 3
	PatchStatusWaiting  = 4
	PatchStatusInQueue  = 5

	// 更新策略
	UpdateByDefault    = 0
	UpdateByGooglePlay = 1
	UpdateByWeb        = 2
	UpdateByCustom     = 3

	// 流量配比
	PackFlowFull = "0,99"
	PackFlowZero = "0,0"
)

type Env string

const (
	EnvTest = Env("test")
	EnvProd = Env("prod")
)

// PackResult struct for cd list.
type PackResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*PackItem     `json:"items"`
}

// PatchResult struct patch list.
type PatchResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Patch        `json:"items"`
}

// GenerateResult struct generate list.
type GenerateResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Generate     `json:"items"`
}

// PackItem struct.
type PackItem struct {
	*model.Version
	Items []*Pack `json:"items,omitempty"`
}

// Pack struct.
type Pack struct {
	ID                  int64               `json:"id"`
	AppID               string              `json:"app_id"`
	AppKey              string              `json:"app_key"`
	Env                 string              `json:"env"`
	VersionID           int64               `json:"version_id"`
	Version             string              `json:"version"`
	VersionCode         int64               `json:"version_code"`
	InternalVersionCode int64               `json:"internal_version_code"`
	BuildID             int64               `json:"build_id"`
	GlJobURL            string              `json:"gl_job_url"`
	GitType             int8                `json:"git_type"`
	GitName             string              `json:"git_name"`
	Commit              string              `json:"commit"`
	PackType            int8                `json:"pack_type"`
	SteadyState         int8                `json:"steady_state"`
	Operator            string              `json:"operator"`
	Size                int64               `json:"size"`
	MD5                 string              `json:"md5"`
	PackPath            string              `json:"pack_path"`
	PackURL             string              `json:"pack_url"`
	MappingURL          string              `json:"mapping_url"`
	RURL                string              `json:"r_url"`
	RMappingURL         string              `json:"r_mapping_url"`
	CDNURL              string              `json:"cdn_url"`
	Desc                string              `json:"description"`
	Sender              string              `json:"sender"`
	ChangeLog           string              `json:"change_log"`
	CTime               int64               `json:"ctime"`
	PTime               xtime.Time          `json:"ptime"`
	MTime               int64               `json:"mtime"`
	Flow                string              `json:"flow,omitempty"`
	Config              *FilterConfig       `json:"config,omitempty"`
	TestFlightInfo      *TestFlightPackInfo `json:"tf_info,omitempty"`
	DepGitJobId         int64               `json:"dep_gl_job_id"`
	IsCompatible        int                 `json:"is_compatible"` // Deprecated:tribe feature api之后这个字段已无法表示兼容关系
	BbrUrl              string              `json:"bbr_url"`
	Features            string              `json:"features"`
}

// UpgradConfig struct.
type UpgradConfig struct {
	AppKey         string `json:"app_key"`
	Env            string `json:"env"`
	VersionID      int64  `json:"version_id"`
	Version        string `json:"-"`
	VersionCode    int64  `json:"-"`
	Normal         string `json:"normal"`
	Force          string `json:"force"`
	ExNormal       string `json:"exclude_normal"`
	ExForce        string `json:"exclude_force"`
	System         string `json:"system"`
	ExcludeSystem  string `json:"exclude_system"`
	Cycle          int8   `json:"cycle"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	IsSilent       int8   `json:"is_silent"`
	Policy         int    `json:"policy"`
	PolicyURL      string `json:"policy_url"`
	IconURL        string `json:"icon_url"`
	ConfirmBtnText string `json:"confirm_btn_text"`
	CancelBtnText  string `json:"cancel_btn_text"`
}

// FilterConfig struct.
type FilterConfig struct {
	AppKey     string `json:"app_key,omitempty"`
	Env        string `json:"env,omitempty"`
	BuildID    int64  `json:"-"`
	Network    string `json:"network,omitempty"`
	ISP        string `json:"isp,omitempty"`
	City       string `json:"city,omitempty"`
	Channel    string `json:"channel,omitempty"`
	Percent    int8   `json:"percent,omitempty"`
	Salt       string `json:"salt,omitempty"`
	Device     string `json:"device,omitempty"`
	Status     int8   `json:"status,omitempty"`
	PhoneModel string `json:"phone_model,omitempty"`
	Brand      string `json:"brand,omitempty"`
}

// FlowConfig struct
type FlowConfig struct {
	AppKey  string `json:"app_key,omitempty"`
	Env     string `json:"env,omitempty"`
	BuildID int64  `json:"-"`
	Flow    string `json:"flow"`
	CTime   int64  `json:"ctime"`
	MTime   int64  `json:"mtime"`
}

// Patch struct
type Patch struct {
	ID                int64     `json:"id"`
	AppKey            string    `json:"app_key"`
	BuildID           int64     `json:"build_id"`
	TargetBuildID     int64     `json:"target_build_id"`
	TargetVersionID   int64     `json:"target_version_id"`
	TargetVersionCode int64     `json:"target_version_code"`
	TargetVersion     string    `json:"target_version"`
	OriginBuildID     int64     `json:"origin_build_id"`
	OriginVersionID   int64     `json:"origin_version_id"`
	OriginVersionCode int64     `json:"origin_version_code"`
	OriginVersion     string    `json:"origin_version"`
	Size              int64     `json:"size"`
	Status            int       `json:"status"`
	GlJobID           int64     `json:"gl_job_id"`
	GlJobURL          string    `json:"gl_job_url"`
	MD5               string    `json:"md5"`
	PackURL           string    `json:"pack_url"`
	PatchPath         string    `json:"patch_path"`
	PatchURL          string    `json:"patch_url"`
	CDNURL            string    `json:"cdn_url"`
	PatchState        int64     `json:"patch_state"`
	CTime             time.Time `json:"ctime"`
	MTime             time.Time `json:"mtime"`
}

// UpgradeVersion struct.
type UpgradeVersion struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// ExcludeUpgradeVersion struct.
type ExcludeUpgradeVersion struct {
	VersionID int64   `json:"version_id"`
	BuildIDs  []int64 `json:"build_id"`
}

// Generate struct
type Generate struct {
	ID               int64  `json:"id"`
	AppKey           string `json:"app_key"`
	BuildID          int64  `json:"build_id"`
	ChannelID        int64  `json:"channel_id"`
	Name             string `json:"name"`
	Folder           string `json:"-"`
	GeneratePath     string `json:"generate_path"`
	GenerateURL      string `json:"generate_url"`
	CDNURL           string `json:"cdn_url"`
	Status           int8   `json:"status"`
	MD5              string `json:"md5"`
	Size             int64  `json:"size"`
	Operator         string `json:"operator"`
	ChannelTestState int8   `json:"channel_test_state"`
	PackState        int8   `json:"pack_state"`
	CTime            int64  `json:"ctime"`
	PTime            int64  `json:"ptime"`
	Mtime            int64  `json:"mtime"`
	ChannelCode      string `json:"channel_code,omitempty"`
	PackCDNURL       string `json:"pack_cdn_url,omitempty"`
	GlJobURL         string `json:"gl_job_url"`
}

// GeneratePublishLastest struct
type GeneratePublishLastest struct {
	AppKey      string `json:"app_key"`
	BuildID     int64  `json:"build_id"`
	CDNURL      string `json:"cdn_url"`      // cdn上固定地址的渠道Url
	SoleCDNURL  string `json:"sole_cdn_url"` // cdn上唯一的url
	Version     string `json:"version"`
	VersionCode string `json:"version_code"`
	Size        int64  `json:"size"`
	MD5         string `json:"md5"`
	MTime       int64  `json:"mtime"`
}

// ChannelGenerate Struct
type ChannelGenerate struct {
	*appmdl.Channel
	State    int8      `json:"state"`
	Generate *Generate `json:"generate,omitempty"`
}

// ChannelResult struct
type ChannelResult struct {
	PageInfo *model.PageInfo    `json:"page"`
	Items    []*ChannelGenerate `json:"items"`
}

type UploadResult struct {
	FilePath string `json:"file_path,omitempty"`
}

type ChannelGitPipe struct {
	Channel string `json:"channel"`
	ID      int64  `json:"id"`
}

type ChannelGeneParam struct {
	Channel   string `json:"channel"`
	ChannelID int64  `json:"channel_id"`
}

type ChannelFileInfo struct {
	ID   int64  `json:"id"`
	MD5  string `json:"md5"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type ManagerVersion struct {
	ID          int64  `json:"id"`
	Platform    int8   `json:"plat"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Build       int64  `json:"build"`
	State       int8   `json:"state"`
	CTime       int64  `json:"ctime"`
	PTime       int64  `json:"ptime"`
	MTime       int64  `json:"mtime"`
}

type ManagerVersionUpdate struct {
	ID         int64  `json:"id"`
	VID        int64  `json:"vid"`
	Channel    string `json:"channnel"`
	Coverage   int8   `json:"coverage"`
	Size       int8   `json:"size"`
	URL        string `json:"url"`
	MD5        string `json:"md5"`
	State      int8   `json:"state"`
	CTime      int64  `json:"ctime"`
	MTime      int64  `json:"mtime"`
	SDKInt     int64  `json:"sdkint"`
	Model      string `json:"model"`
	Policy     int8   `json:"policy"`
	IsForce    int8   `json:"is_force"`
	PolicyName int64  `json:"policy_name"`
	IsPush     int8   `json:"is_push"`
	PolicyURL  int64  `json:"policy_url"`
	SDKIntList string `json:"sdkint_list"`
	BuvidStart int64  `json:"buvid_start"`
	BuvidEnd   int64  `json:"buvid_end"`
}

type ManagerVersionUpdateLimit struct {
	ID    int64  `json:"id"`
	UPID  int8   `json:"up_id"`
	CONDI string `json:"condi"`
	VALUE int64  `json:"value"`
	CTime int64  `json:"ctime"`
	MTime int64  `json:"mtime"`
}

type CDNRefreshReq struct {
	Action    string   `json:"action"`
	AccountID int64    `json:"account_id"`
	Urls      []string `json:"urls"`
}

type CDNRefreshRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Custom Channel Pack struct.
type CustomChannelPack struct {
	ID       int64  `json:"id"`
	AppKey   string `json:"app_key"`
	BuildID  int64  `json:"build_id"`
	Operator string `json:"operator"`
	Size     int64  `json:"size"`
	MD5      string `json:"md5"`
	PackName string `json:"pack_name"`
	PackPath string `json:"pack_path"`
	PackURL  string `json:"pack_url"`
	CDNURL   string `json:"cdn_url"`
	Sender   string `json:"sender"`
	State    int    `json:"state"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

// DownloadURLRes struct.
type DownloadURLRes struct {
	URL string `json:"url"`
}

// TestFlightAppInfo struct.
type TestFlightAppInfo struct {
	AppKey            string `json:"app_key,omitempty"`
	StoreAppID        string `json:"store_app_id,omitempty"`
	BetaGroupID       string `json:"beta_group_id,omitempty"`
	PublicLink        string `json:"public_link,omitempty"`
	BetaGroupIDTest   string `json:"beta_group_id_test,omitempty"`
	PublicLinkTest    string `json:"public_link_test,omitempty"`
	TagPrefix         string `json:"tag_prefix,omitempty"`
	BuglyAppID        string `json:"bugly_app_id,omitempty"`
	BuglyAppKey       string `json:"bugly_app_key,omitempty"`
	OnlineVersion     string `json:"online_version,omitempty"`
	OnlineVersionCode int64  `json:"online_version_code,omitempty"`
	OnlineBuildID     int64  `json:"online_build_id,omitempty"`
	IssuerID          string `json:"-"`
	KeyID             string `json:"-"`
}

// TestFlightPackInfo struct.
type TestFlightPackInfo struct {
	ID              int64      `json:"pack_tf_id"`
	AppKey          string     `json:"app_key"`
	PackID          int64      `json:"pack_id"`
	Version         string     `json:"version"`
	VersionCode     int64      `json:"version_code"`
	Env             string     `json:"env"`
	PackPath        string     `json:"-"`
	BuildID         int64      `json:"-"`
	StoreAppID      string     `json:"-"`
	BetaGroupID     string     `json:"-"`
	PublicLink      string     `json:"-"`
	BetaGroupIDTest string     `json:"-"`
	PublicLinkTest  string     `json:"-"`
	BetaBuildID     string     `json:"beta_build_id,omitempty"`
	ExpireTime      xtime.Time `json:"expire_time,omitempty"`
	PackState       string     `json:"pack_state,omitempty"`
	ReviewState     string     `json:"review_state,omitempty"`
	BetaState       int        `json:"beta_state,omitempty"`
	DisPermil       int        `json:"dis_permil,omitempty"`
	DisNum          int64      `json:"dis_num,omitempty"`
	DisLimit        int64      `json:"dis_limit,omitempty"`
	RemindUpdTime   xtime.Time `json:"remind_upd_time,omitempty"`
	ForceupdTime    xtime.Time `json:"force_upd_time,omitempty"`
	GuideTFTxt      string     `json:"guide_tf_txt,omitempty"`
	RemindUpdTxt    string     `json:"remind_upd_txt,omitempty"`
	ForceUpdTxt     string     `json:"force_upd_txt,omitempty"`
	CTime           xtime.Time `json:"ctime,omitempty"`
}

// TFAppBaseInfo TestFlight App base info
type TFAppBaseInfo struct {
	AppKey         string
	MobiApp        string
	StoreAppID     string
	PublicLink     string
	PublicLinkTest string
}

// TestFlightAttribute struct.
type TestFlightAttribute struct {
	Version      string `json:"version"`
	VersionCode  int64  `json:"version_code"`
	UpdateURL    string `json:"update_url"`
	DisPermil    int    `json:"dis_permil,omitempty"`
	GuideTFTxt   string `json:"guide_tf_txt,omitempty"`
	RemindUpdTxt string `json:"remind_upd_txt"`
	ForceUpdTxt  string `json:"force_upd_txt"`
	PackageType  string `json:"package_type"`
}

// TestFlightUpdTimeInfo struct.
type TestFlightUpdTimeInfo struct {
	Version       string     `json:"version"`
	VersionCode   int64      `json:"version_code"`
	RemindUpdTime xtime.Time `json:"remind_upd_time"`
	ForceUpdTime  xtime.Time `json:"force_upd_time"`
}

// TestFlightUpdInfo struct.
type TestFlightUpdInfo struct {
	AppKey       string                   `json:"app_key"`
	MobiApp      string                   `json:"mobi_app"`
	LatestOnline *TestFlightAttribute     `json:"latest_online,omitempty"`
	LatestTF     *TestFlightAttribute     `json:"latest_tf,omitempty"`
	TFPacks      []*TestFlightUpdTimeInfo `json:"tf_packs,omitempty"`
	BlackList    []int64                  `json:"black_list,omitempty"`
	WhiteList    []int64                  `json:"white_list,omitempty"`
}

// TestFlightBWList struct.
type TestFlightBWList struct {
	ID       int64      `json:"id"`
	MID      int64      `json:"mid"`
	Nick     string     `json:"nick"`
	Operator string     `json:"operator"`
	CTime    xtime.Time `json:"ctime"`
}

// PatchPipeline struct.
type PatchPipeline struct {
	ID      int64  `json:"id"`
	PackURL string `json:"pack_url"`
}

// PatchPipelineData struct.
type PatchPipelineData struct {
	Data []*PatchPipeline `json:"data"`
}

type PackGreyHistory struct {
	ID             int64     `json:"id"`
	AppKey         string    `json:"app_key"`
	Version        string    `json:"version"`
	VersionCode    int64     `json:"version_code"`
	GlJobID        int64     `json:"gl_job_id"`
	IsUpgrade      int8      `json:"is_upgrade"`
	Flow           string    `json:"flow"`
	GreyStartTime  time.Time `json:"grey_start_time"`
	GreyFinishTime time.Time `json:"grey_finish_time"`
	GreyCloseTime  time.Time `json:"grey_close_time"`
	Operator       string    `json:"operator"`
	CTime          time.Time `json:"ctime"`
	MTime          time.Time `json:"mtime"`
}

// PackGreyData struct
type PackGreyData struct {
	Id              int64         `json:"id"`
	AppKey          string        `json:"app_key"`
	DatacenterAppId int64         `json:"datacenter_app_id"`
	Platform        string        `json:"platform"`
	MobiApp         string        `json:"mobi_app"`
	Version         string        `json:"version"`
	VersionCode     int64         `json:"version_code"`
	GlJobID         int64         `json:"gl_job_id"`
	IsUpgrade       int8          `json:"is_upgrade"`
	Flow            string        `json:"flow,omitempty"`
	Config          *FilterConfig `json:"config,omitempty"`
	GreyStartTime   time.Time     `json:"grey_start_time"`
	GreyFinishTime  time.Time     `json:"grey_finish_time"`
	GreyCloseTime   time.Time     `json:"grey_close_time"`
	OperateTime     time.Time     `json:"operate_time"`
	CTime           time.Time     `json:"ctime"`
	MTime           time.Time     `json:"mtime"`
}

type PackGreyDataResp struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*PackGreyData `json:"items"`
}
