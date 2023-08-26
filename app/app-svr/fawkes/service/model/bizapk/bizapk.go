package bizapk

const (
	NEED_UPLOAD = 1
	NOT_UPLOAD  = 0
)

// BuildListResp struct.
type BuildListResp struct {
	Name        string   `json:"name"`
	Cname       string   `json:"cname"`
	Description string   `json:"description"`
	BizApkID    int64    `json:"apk_id"`
	SettingsID  int64    `json:"setttings_id"`
	Active      int8     `json:"active"`
	Priority    int8     `json:"priority"`
	Builds      []*Build `json:"builds"`
}

// ApkPackSettings struct.
type ApkPackSettings struct {
	BizApkID    int64  `json:"apk_id"`
	Name        string `json:"name"`
	Cname       string `json:"cname"`
	Description string `json:"description"`
	SettingsID  int64  `json:"settings_id"`
	PackBuildID int64  `json:"-"`
	Env         string `json:"-"`
	Active      int8   `json:"active"`
	Priority    int8   `json:"priority"`
	Operator    string `json:"operator"`
	MTime       int64  `json:"mtime"`
}

// Build struct.
type Build struct {
	ID               int64         `json:"id"`
	AppKey           string        `json:"-"`
	Name             string        `json:"-"`
	Cname            string        `json:"-"`
	Description      string        `json:"-"`
	BizApkID         int64         `json:"-"`
	Env              string        `json:"-"`
	PackBuildID      int64         `json:"pack_build_id"`
	BundleVer        int64         `json:"bundle_ver"`
	MD5              string        `json:"md5"`
	Size             int64         `json:"size"`
	GitlabPipelineID int64         `json:"gl_ppl_id"`
	GitlabJobID      int64         `json:"gl_job_id"`
	GitlabPplURL     string        `json:"gl_ppl_url"`
	GitlabJobURL     string        `json:"gl_job_url"`
	GitType          int8          `json:"git_type"`
	GitName          string        `json:"git_name"`
	Commit           string        `json:"commit"`
	ApkPath          string        `json:"-"`
	MapPath          string        `json:"-"`
	MetaPath         string        `json:"-"`
	ApkURL           string        `json:"apk_url"`
	MapURL           string        `json:"map_url"`
	MetaURL          string        `json:"meta_url"`
	ApkCdnURL        string        `json:"apk_cdn_url"`
	Status           int8          `json:"status"`
	Operator         string        `json:"operator"`
	DidPush          int8          `json:"did_push"`
	BuiltIn          int           `json:"built_In"`
	Flow             string        `json:"flow,omitempty"`
	Config           *FilterConfig `json:"config,omitempty"`
	SettingsID       int64         `json:"-"`
	Active           int8          `json:"-"`
	Priority         int8          `json:"-"`
	CTime            int64         `json:"-"`
	MTime            int64         `json:"mtime"`
}

// UploadResp struct.
type UploadResp struct {
	ApkURL   string `json:"apk_url"`
	MapURL   string `json:"map_url"`
	MetaURL  string `json:"meta_url"`
	ApkPath  string `json:"-"`
	MapPath  string `json:"-"`
	MetaPath string `json:"-"`
}

// JobInfo struct.
type JobInfo struct {
	ID               int64
	GitlabProjectID  string
	GitlabPipelineID int64
	GitlabJobID      int64
	Status           int
}

// OrgPackURLResp struct.
type OrgPackURLResp struct {
	PkgURL     string
	MappingURL string
}

// FilterConfig struct.
type FilterConfig struct {
	Env            string `json:"env,omitempty"`
	BuildID        int64  `json:"-"`
	Network        string `json:"network,omitempty"`
	ISP            string `json:"isp,omitempty"`
	City           string `json:"city,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Percent        int8   `json:"percent,omitempty"`
	Salt           string `json:"salt,omitempty"`
	Device         string `json:"device,omitempty"`
	Status         int8   `json:"status,omitempty"`
	ExcludesSystem string `json:"excludes_system,omitempty"`
}

// FlowConfig struct
type FlowConfig struct {
	Env     string `json:"env,omitempty"`
	BuildID int64  `json:"-"`
	Flow    string `json:"flow"`
}

type Apk struct {
	Name         string        `json:"name"`
	BuildID      int64         `json:"-"`
	PackBuildID  int64         `json:"pack_build_id"`
	BundleVer    int64         `json:"bundle_ver"`
	Env          string        `json:"env"`
	MD5          string        `json:"md5"`
	ApkCdnURL    string        `json:"apk_cdn_url"`
	Priority     int           `json:"priority"`
	FilterConfig *FilterConfig `json:"filter_config"`
	FlowConfig   *FlowConfig   `json:"flow_config"`
}
