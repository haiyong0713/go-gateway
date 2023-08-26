package app

// fawkes-admin const
const (
	EffectNo  = 1
	EffectYes = 2
)

// HfPushEnvReq is hotfixPushEnv params
type HfPushEnvReq struct {
	AppKey  string `form:"app_key" validate:"required"`
	BuildID int64  `form:"build_id" validate:"required"`
	Env     string `form:"env" validate:"required"`
}

// HfConfSetReq is hotfixConfSet params
type HfConfSetReq struct {
	AppKey    string `form:"app_key" validate:"required"`
	Env       string `form:"env" validate:"required"`
	BuildID   int64  `form:"build_id" validate:"required,min=0"`
	Channel   string `form:"channel"`
	City      string `form:"city"`
	UpgradNum int    `form:"upgrad_num"`
	Device    string `form:"device"`
	Status    int    `form:"status" default:"2"`
}

// HfConfGetReq is hotfixConf
// Get params
type HfConfGetReq struct {
	AppKey  string `form:"app_key" validate:"required"`
	Env     string `form:"env" validate:"required"`
	BuildID int64  `form:"build_id" validate:"required,min=0"`
}

// HotfixConf is hotfix_config struct
type HotfixConf struct {
	ID        int    `json:"id"`
	AppKey    string `json:"app_key"`
	Env       string `json:"env"`
	BuildID   int64  `json:"build_id"`
	City      string `json:"city"`
	Channel   string `json:"channel"`
	UpgradNum int    `json:"upgrad_num"`
	Device    string `json:"device"`
	Status    int    `json:"status"`
	Effect    int    `json:"effect"`
}

// HotfixInfo is hotfix information
type HotfixInfo struct {
	AppID               string `json:"app_id"`
	AppKey              string `json:"app_key"`
	Env                 string `json:"env"`
	OrigVersion         string `json:"origin_version"`
	OrigVersionCode     int64  `json:"origin_version_code"`
	OrigBuildID         int64  `json:"origin_build_id"`
	BuildID             int64  `json:"build_id"`
	GlJobID             int64  `json:"gl_job_id"`
	InternalVersionCode int64  `json:"internal_version_code"`
	Commit              string `json:"commit"`
	Operator            string `json:"operator"`
}

// HfEffectReq is hotfix effect params
type HfEffectReq struct {
	AppKey  string `form:"app_key" validate:"required"`
	Env     string `form:"env" validate:"required"`
	BuildID int64  `form:"build_id" validate:"required,min=0"`
	Effect  int    `form:"effect" validate:"required,min=1,max=2"`
}

// GetPreEnv get previous env
func GetPreEnv(env string) (next string) {
	switch env {
	case "test":
		next = "prod"
	case "prod":
		next = "test"
	}
	return
}

// HfListReq is hotfix list request struct
type HfListReq struct {
	AppKey string `form:"app_key" validate:"required"`
	Env    string `form:"env" validate:"required"`
	Pn     int    `form:"pn" validate:"min=0" default:"1"`
	Ps     int    `form:"ps" validate:"min=0,max=20" default:"20"`
	Order  string `form:"order"`
	Sort   string `form:"sort"`
}

// HfListItem is hotfix list response struct
type HfListItem struct {
	AppID               string   `json:"app_id"`
	AppKey              string   `json:"app_key"`
	Version             string   `json:"version"`
	VersionCode         int      `json:"version_code"`
	OriginBuildID       int64    `json:"origin_build_id"`
	BuildID             int64    `json:"build_id"`
	Operator            string   `json:"operator"`
	Status              int      `json:"status"`
	Size                int      `json:"size"`
	Md5                 string   `json:"md5"`
	Mtime               int64    `json:"mtime"`
	PatchURL            string   `json:"patch_url"`
	CdnURL              string   `json:"cdn_url"`
	GitType             int      `json:"git_type"`
	GitName             string   `json:"git_name"`
	Commit              string   `json:"commit"`
	ShortCommit         string   `json:"short_commit"`
	GlJobID             int64    `json:"gl_job_id"`
	GlPrjID             string   `json:"gl_prj_id"`
	JobURL              string   `json:"job_url"`
	Origin              HfOrigin `json:"origin"`
	Config              HfConfig `json:"config"`
	InternalVersionCode int64    `json:"internal_version_code"`
	EnvVars             string   `json:"env_vars"`
}

// HfOrigin of HfListItem
type HfOrigin struct {
	Size                int    `json:"size"`
	Md5                 string `json:"md5"`
	Mtime               int64  `json:"mtime"`
	PkgURL              string `json:"pkg_url"`
	BuildID             int64  `json:"build_id"`
	GitType             int    `json:"git_type"`
	GitName             string `json:"git_name"`
	Commit              string `json:"commit"`
	InternalVersionCode int64  `json:"internal_version_code"`
}

// HfConfig of HfListItem
type HfConfig struct {
	AppKey    string `json:"app_key"`
	Env       string `json:"env"`
	BuildID   int64  `json:"build_id"`
	City      string `json:"city"`
	Channel   string `json:"channel"`
	UpgradNum int    `json:"upgrad_num"`
	Device    string `json:"device"`
	Status    int    `json:"status"`
	Effect    int    `json:"effect"`
}

// HfBuildReq is hotfixbuild params
type HfBuildReq struct {
	AppKey              string `form:"app_key" validate:"required"`
	BuildID             int64  `form:"build_id" validate:"required"`
	GitType             int    `form:"git_type" default:"0"`
	GitName             string `form:"git_name" validate:"required"`
	InternalVersionCode int64  `form:"internal_version_code" validate:"required"`
}

// HfUpdateReq is hotfixupdate params
type HfUpdateReq struct {
	PatchBuildID int64  `form:"patch_build_id" validate:"required"`
	GlJobID      int64  `form:"gl_job_id" validate:"required"`
	Commit       string `form:"commit" validate:"required"`
}

// HfUploadReq is hotfixupload params
type HfUploadReq struct {
	PatchBuildID int64  `form:"patch_build_id" validate:"required"`
	PatchName    string `form:"patch_name" validate:"required"`
}

// HfOriginVersion struct
type HfOriginVersion struct {
	VersionID   int64
	Version     string
	VersionCode int64
}

// HfCancelReq is hotfixcancel params
type HfCancelReq struct {
	AppKey       string `form:"app_key" validate:"required"`
	PatchBuildID int64  `form:"patch_build_id" validate:"required"`
}

// HfOriginInfoReq is hotfixOrigGet params
type HfOriginInfoReq struct {
	AppKey  string `form:"app_key" validate:"required"`
	PatchID int64  `form:"patch_id" validate:"required"`
}

// HfOrigURLInfo is hotfix origin package's URL information
type HfOrigURLInfo struct {
	PackPath    string `json:"pack_path"`
	PackURL     string `json:"pack_url"`
	MappingURL  string `json:"mapping_url"`
	RURL        string `json:"r_url"`
	RMappingURL string `json:"r_mapping_url"`
}

// HfDelReq is hotfixdel params
type HfDelReq struct {
	AppKey       string `form:"app_key" validate:"required"`
	PatchBuildID int64  `form:"patch_build_id" validate:"required"`
}

// HfUpgrade struct
type HfUpgrade struct {
	AppID           string        `json:"app_id"`
	AppKey          string        `json:"app_key"`
	Env             string        `json:"env"`
	GlPrjID         string        `json:"gl_prj_id"`
	GlJobID         string        `json:"gl_job_id"`
	OrigVersion     string        `json:"origin_version"`
	OrigVersionCode int64         `json:"origin_version_code"`
	OrigBuildID     int64         `json:"origin_build_id"`
	BuildID         int64         `json:"build_id"`
	GitType         int           `json:"git_type"`
	GitName         string        `json:"git_name"`
	Commit          string        `json:"commit"`
	Size            string        `json:"size"`
	Md5             string        `json:"md5"`
	HotfixPath      string        `json:"hotfix_path"`
	HotfixURL       string        `json:"hotfix_url"`
	CDNURL          string        `json:"cdn_url"`
	Description     string        `json:"description"`
	Status          int           `json:"status"`
	State           int           `json:"state"`
	Config          *HotfixConfig `json:"config"`
}

// HotfixConfig stuct
type HotfixConfig struct {
	AppKey    string `json:"app_key"`
	Env       string `json:"env"`
	BuildID   int64  `json:"build_id"`
	City      string `json:"city"`
	Channel   string `json:"channel"`
	UpgradNum int    `json:"upgrad_num"`
	Gray      int    `json:"gray"`
	Device    string `json:"device"`
	Status    int    `json:"status"`
	Effect    int    `json:"effect"`
}

// HotfixJobInfo struct
type HotfixJobInfo struct {
	BuildPatchID    int64
	GitlabProjectID string
	GitlabJobID     int64
	Status          int
}
