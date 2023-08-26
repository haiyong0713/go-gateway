package fawkes

import (
	xtime "go-common/library/time"
)

// fawkes consts.
const (
	FilterStatusCustom = 0
	FilterStatusAll    = 1
	FilterStatusInner  = 2

	UpgradeNormal = int8(1)
	UpgradeForce  = int8(2)

	StatusUpFaild   = -2
	StatusSendFaild = -1
	StatusQueuing   = 1
	StatusWaitSend  = 2
	StatusUpSuccess = 3
)

// Item struct.
type Item struct {
	Titel       string     `json:"title"`
	Content     string     `json:"content"`
	Version     string     `json:"version"`
	VersionCode int64      `json:"version_code"`
	URL         string     `json:"url"`
	Size        int64      `json:"size"`
	MD5         string     `json:"md5"`
	Patch       *ItemPatch `json:"patch,omitempty"`
	Silent      int8       `json:"silent"`
	UpType      int8       `json:"upgrade_type"`
	Cycle       int8       `json:"cycle"`
	Policy      int        `json:"policy"`
	PolicyURL   string     `json:"policy_url"`
	PTime       xtime.Time `json:"ptime"`
}

type IOSItem struct {
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Version        string     `json:"version"`
	VersionCode    int64      `json:"version_code"`
	PolicyURL      string     `json:"policy_url"`
	Cycle          int8       `json:"cycle"`
	Ptime          xtime.Time `json:"ptime"`
	IconURL        string     `json:"icon_url"`
	ConfirmBtnText string     `json:"confirm_btn_text"`
	CancelBtnText  string     `json:"cancel_btn_text"`
}

// ItemPatch for patch.
type ItemPatch struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
	MD5  string `json:"md5"`
}

// UpgradeVersion struct.
type UpgradeVersion struct {
	VersionID int64   `json:"version_id"`
	BuildIDs  []int64 `json:"build_id"`
}

// HfUpgradeInfo struct
type HfUpgradeInfo struct {
	Version     string `json:"version"`
	VersionCode int64  `json:"version_code"`
	PatchURL    string `json:"patch_url"`
	PatchMd5    string `json:"patch_md5"`
}

// LaserMsg struct.
type LaserMsg struct {
	Date   string `json:"date"`
	TaskID string `json:"taskid"`
}

type ApkListParam struct {
	Buvid    string `form:"-"`
	AppKey   string `form:"-"`
	Env      string `form:"-"`
	Sn       int64  `form:"sn" validate:"required"`
	Vn       string `form:"vn" validate:"required"`
	Build    int    `form:"build" validate:"required"`
	Channel  string `form:"channel" validate:"required"`
	Nt       string `form:"nt" validate:"required"`
	Bundle   string `form:"bundle"`
	Priority int    `form:"priority" validate:"min=0"`
	Ov       string `form:"ov"`
}

type Apk struct {
	Name      string `json:"name"`
	BundleVer int64  `json:"bundle_ver"`
	MD5       string `json:"md5"`
	ApkCdnURL string `json:"apk_cdn_url"`
	Priority  int    `json:"priority"`
}

type TribeListParam struct {
	Buvid    string `form:"-"`
	AppKey   string `form:"-"`
	Env      string `form:"-"`
	HostVer  int64  `form:"host_ver" validate:"required"`
	Vn       string `form:"vn" validate:"required"`
	Build    int    `form:"build" validate:"required"`
	Channel  string `form:"channel" validate:"required"`
	Nt       string `form:"nt" validate:"required"`
	Bundle   string `form:"bundle"`
	Priority int    `form:"priority" validate:"min=0"`
	Ov       string `form:"ov"`
}

type TribeApk struct {
	Name      string `json:"name"`
	BundleVer int64  `json:"bundle_ver"`
	MD5       string `json:"md5"`
	ApkCdnURL string `json:"apk_cdn_url"`
	Priority  int    `json:"priority"`
}

type TestFlightParam struct {
	MobiApp      string `form:"mobi_app"`
	Build        int64  `form:"build"`
	IsTestflight int    `form:"is_testflight"`
	Buvid        string
}

type TestFlight struct {
	AppKey         string            `json:"app_key"`
	MobiApp        string            `json:"mobi_app"`
	OnlinePack     *TestFlightPack   `json:"latest_online"`
	TestFlightPack *TestFlightPack   `json:"latest_tf"`
	Packs          []*TestFlightPack `json:"tf_packs"`
	BlackList      []int64           `json:"black_list"`
	WhiteList      []int64           `json:"white_list"`
}

type TestFlightPack struct {
	Version     string     `json:"version,omitempty"`
	VersionCode int64      `json:"version_code,omitempty"`
	DisPermil   uint32     `json:"dis_permil,omitempty"`
	UpdateURL   string     `json:"update_url,omitempty"`
	GuideText   string     `json:"guide_tf_txt,omitempty"`
	RemindText  string     `json:"remind_upd_txt,omitempty"`
	RemindTime  xtime.Time `json:"remind_upd_time"`
	ForceText   string     `json:"force_upd_txt,omitempty"`
	ForceTime   xtime.Time `json:"force_upd_time"`
	PackageType string     `json:"package_type"`
}

type TestFlightResult struct {
	IsForce     bool   `json:"is_force"`
	URL         string `json:"url"`
	Desc        string `json:"desc"`
	Version     int64  `json:"version"`
	PackageType string `json:"package_type"`
	IsWhite     bool   `json:"is_white"`
}

type LaserActive struct {
	TaskID int64 `json:"task_id"`
}

type UpgradeIOSParam struct {
	Sn           int64  `form:"sn" validate:"required"`
	Vn           string `form:"vn" validate:"required"`
	Build        int64  `form:"build" validate:"required"`
	Nt           string `form:"nt" validate:"required"`
	Ov           string `form:"ov" validate:"required"`
	Model        string `form:"model" validate:"required"`
	IsTestflight bool   `form:"is_testflight"`
	Is32bit      bool   `form:"is_32bit"`
	Buvid        string `form:"-"`
	FawkesAppKey string `form:"-"`
	FawkesEnv    string `form:"-"`
	IP           string `form:"-"`
}
