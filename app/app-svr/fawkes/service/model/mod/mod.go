package mod

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

type (
	ModuleState   string
	Env           string
	Priority      string
	Compress      string
	Condition     string
	Perm          string
	VersionState  string
	Role          string
	FawkesPerm    int64
	ApplyState    string
	TrafficCost   int64
	TrafficAdvice string
	OperateType   string
)

var (
	ModuleOnline  = ModuleState("online")
	ModuleOffline = ModuleState("offline")

	EnvTest = Env("test")
	EnvProd = Env("prod")

	PriorityHigh   = Priority("high")
	PriorityMiddle = Priority("middle")
	PriorityLow    = Priority("low")

	CompressUnzip    = Compress("unzip")
	CompressOriginal = Compress("original")

	ConditionLt = Condition("lt")
	ConditionGt = Condition("gt")
	ConditionLe = Condition("le")
	ConditionGe = Condition("ge")

	PermAdmin = Perm("admin")
	PermNone  = Perm("none")

	RoleUser           = Role("user")
	RoleModAdmin       = Role("modAdmin")
	RoleWhitelistAdmin = Role("whitelistAdmin")
	RoleAppAdmin       = Role("appAdmin")
	RoleSuperAdmin     = Role("superAdmin")

	VersionProcessing = VersionState("processing")
	VersionSucceeded  = VersionState("succeeded")
	VersionDisable    = VersionState("disable")

	// 权限点从高到低：fawkes超级管理员 > fawkes app管理员 = whilte admin > 资源池管理员 > 普通成员(研发，测试，运营)
	FawkesUserPerm     = FawkesPerm(1)
	FawkesModAdminPerm = FawkesPerm(2)
	// FawkesWhitelistAdminPerm = FawkesPerm(3)
	FawkesAppAdminPerm   = FawkesPerm(4)
	FawkesSuperAdminPerm = FawkesPerm(5)

	ApplyStateChecking = ApplyState("checking")
	ApplyStatePassed   = ApplyState("passed")
	ApplyStateRefused  = ApplyState("refused")

	CostLow      = TrafficCost(1)
	CostMiddle   = TrafficCost(2)
	CostHigh     = TrafficCost(3)
	CostVeryHigh = TrafficCost(4)

	Release      = OperateType("发布")
	ConfigChange = OperateType("配置变更")
)

func (v ModuleState) Valid() bool {
	return v == ModuleOnline || v == ModuleOffline
}

func (v Env) Valid() bool {
	return v == EnvTest || v == EnvProd
}

func (v Priority) Valid() bool {
	return v == PriorityHigh || v == PriorityMiddle || v == PriorityLow
}

func (v Compress) Valid() bool {
	return v == CompressUnzip || v == CompressOriginal
}

func (v Condition) Valid() bool {
	return v == ConditionLt || v == ConditionGt || v == ConditionLe || v == ConditionGe
}

func (v Perm) Valid() bool {
	return v == PermAdmin || v == PermNone
}

func (v VersionState) IsPolling() bool {
	return v == VersionProcessing
}

func (v ApplyState) Valid() bool {
	return v == ApplyStateChecking || v == ApplyStatePassed || v == ApplyStateRefused
}

type Page struct {
	Total int64 `json:"total"`
	Pn    int64 `json:"pn"`
	Ps    int64 `json:"ps"`
}

type Pool struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Remark           string `json:"remark"`
	ModuleCountLimit int64  `json:"module_count_limit"`
	ModuleSizeLimit  int64  `json:"module_size_limit"`
	ModuleCount      int64  `json:"module_count"`
	ModuleSize       int64  `json:"module_size"`
	AppKey           string `json:"app_key"`
}

type Module struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Remark   string      `json:"remark"`
	Compress Compress    `json:"compress"`
	IsWifi   bool        `json:"is_wifi"`
	State    ModuleState `json:"state"`
	Deleted  int64       `json:"deleted"`
	PoolID   int64       `json:"pool_id"`
	ZipCheck bool        `json:"zip_check"`
}

type Version struct {
	ID               int64         `json:"id"`
	Version          int64         `json:"version"`
	Remark           string        `json:"remark"`
	Released         bool          `json:"released"`
	ReleaseTime      xtime.Time    `json:"release_time"`
	File             *File         `json:"file"`
	Patch            *VersionPatch `json:"patch"`
	ModuleID         int64         `json:"module_id"`
	Env              Env           `json:"env"`
	FromVerID        int64         `json:"from_ver_id"`
	State            VersionState  `json:"state"`
	ApplyState       ApplyState    `json:"apply_state"`
	ConfigApplyState ApplyState    `json:"config_apply_state"`
	GrayApplyState   ApplyState    `json:"gray_apply_state"`
}

type File struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	ContentType string     `json:"content_type"`
	Size        int64      `json:"size"`
	Md5         string     `json:"md5"`
	URL         string     `json:"url"`
	IsPatch     bool       `json:"is_patch"`
	FromVer     int64      `json:"from_ver"`
	Ctime       xtime.Time `json:"ctime"`
}

func (f *File) SetURL(modCDN map[string]string) bool {
	if f.URL == "" {
		return false
	}
	for prefix, host := range modCDN {
		if strings.HasPrefix(f.URL, prefix) {
			f.URL = host + f.URL
			return true
		}
	}
	return false
}

type VersionPatch struct {
	Count int `json:"count"`
}

type Patch struct {
	ID      int64      `json:"id"`
	Name    string     `json:"name"`
	Md5     string     `json:"md5"`
	Size    int64      `json:"size"`
	URL     string     `json:"url"`
	Ctime   xtime.Time `json:"ctime"`
	FromVer string     `json:"from_ver"`
	Declare string     `json:"declare,omitempty"`
}

func (p *Patch) SetURL(modCDN map[string]string) bool {
	if p.URL == "" {
		return false
	}
	for prefix, host := range modCDN {
		if strings.HasPrefix(p.URL, prefix) {
			p.URL = host + p.URL
			return true
		}
	}
	return false
}

type VersionConfig struct {
	ID       int64                 `json:"id"`
	Priority Priority              `json:"priority"`
	AppVer   []map[Condition]int64 `json:"app_ver"`
	SysVer   []map[Condition]int64 `json:"sys_ver"`
	Stime    xtime.Time            `json:"stime"`
	Etime    xtime.Time            `json:"etime"`
}

type Config struct {
	ID        int64      `json:"id"`
	VersionID int64      `json:"version_id"`
	Priority  Priority   `json:"priority"`
	AppVer    string     `json:"app_ver"`
	SysVer    string     `json:"sys_ver"`
	Stime     xtime.Time `json:"stime"`
	Etime     xtime.Time `json:"etime"`
}

type ConfigParam struct {
	VersionID int64      `form:"version_id" validate:"min=1"`
	Priority  Priority   `form:"priority" validate:"required"`
	AppVer    string     `form:"app_ver"`
	SysVer    string     `form:"sys_ver"`
	Stime     xtime.Time `form:"stime"`
	Etime     xtime.Time `form:"etime"`
}

type Gray struct {
	ID             int64  `json:"id"`
	VersionID      int64  `json:"version_id"`
	Strategy       int64  `json:"strategy"`
	Salt           string `json:"salt"`
	BucketStart    int64  `json:"bucket_start"`
	BucketEnd      int64  `json:"bucket_end"`
	Whitelist      string `json:"whitelist"`
	WhitelistURL   string `json:"whitelist_url"`
	ManualDownload bool   `json:"manual_download"`
}

type GrayParam struct {
	VersionID      int64   `form:"version_id" validate:"min=1"`
	Strategy       int64   `form:"strategy"`
	Salt           string  `form:"salt"`
	BucketStart    int64   `form:"bucket_start"`
	BucketEnd      int64   `form:"bucket_end"`
	Whitelist      []int64 `form:"whitelist,split" validate:"dive,gt=0"`
	WhitelistURL   string  `form:"whitelist_url"`
	ManualDownload bool    `form:"manual_download"`
}

type Permission struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	PoolID     int64  `json:"pool_id"`
	Permission Perm   `json:"permission"`
	Deleted    int64  `json:"deleted"`
}

type PermissionParam struct {
	Username   string `form:"username" validate:"required"`
	PoolID     int64  `form:"pool_id" validate:"min=1"`
	Permission Perm   `form:"permission" validate:"required"`
}

type PermissionRole struct {
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

type BusPool struct {
	ID     int64  `json:"id"`
	AppKey string `json:"app_key"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

type BusModule struct {
	ID       int64    `json:"id"`
	PoolID   int64    `json:"pool_id"`
	Name     string   `json:"name"`
	Remark   string   `json:"remark"`
	Compress Compress `json:"compress"`
	IsWifi   bool     `json:"is_wifi"`
	Deleted  int64    `json:"deleted"`
	ZipCheck bool     `json:"zip_check"`
}

type BusVersion struct {
	ID          int64        `json:"id"`
	ModuleID    int64        `json:"module_id"`
	Env         Env          `json:"env"`
	Version     int64        `json:"version"`
	Remark      string       `json:"remark"`
	FromVerID   int64        `json:"from_ver_id"`
	ReleaseTime xtime.Time   `json:"release_time"`
	State       VersionState `json:"state"`
	PoolID      int64        `json:"pool_id"`
	PoolName    string       `json:"pool_name"`
	ModuleName  string       `json:"module_name"`
	Compress    Compress     `json:"compress"`
	IsWifi      bool         `json:"is_wifi"`
	ZipCheck    bool         `json:"zip_check"`
	Mtime       xtime.Time   `json:"mtime"`
}

type BusFile struct {
	ID          int64             `json:"id"`
	VersionID   int64             `json:"version_id"`
	Name        string            `json:"name"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Md5         string            `json:"md5"`
	URL         string            `json:"url"`
	IsPatch     bool              `json:"is_patch"`
	FromVer     int64             `json:"from_ver"`
	Version     *BusVersion       `json:"version"`
	Config      *BusVersionConfig `json:"config"`
	Gray        *BusVersionGray   `json:"gray"`
}

func (f *BusFile) SetURL(mod *conf.Mod) bool {
	if f.URL == "" {
		return false
	}
	for prefix, cdn := range mod.CDN {
		if strings.HasPrefix(f.URL, prefix) {
			h := md5.New()
			_, _ = h.Write([]byte(f.URL))
			b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
			if err != nil {
				log.Error("日志告警 分组错误,error:%+v", err)
			}
			if b%100 < cdn.Bucket {
				f.URL = cdn.NewDomain + f.URL
				return true
			}
			f.URL = cdn.OldDomain + f.URL
			return true
		}
	}
	for prefix, host := range mod.ModCDN {
		if strings.HasPrefix(f.URL, prefix) {
			f.URL = host + f.URL
			return true
		}
	}
	return false
}

type BusVersionConfig struct {
	ID             int64      `json:"id"`
	VersionID      int64      `json:"version_id"`
	Priority       Priority   `json:"priority"`
	AppVer         string     `json:"app_ver"`
	SysVer         string     `json:"sys_ver"`
	Stime          xtime.Time `json:"stime"`
	Etime          xtime.Time `json:"etime"`
	Scale          string     `json:"scale"`
	ForbidenDevice string     `json:"forbiden_device"`
	Arch           string     `json:"arch"`
	Mtime          xtime.Time `json:"mtime"`
}

type BusVersionGray struct {
	ID             int64      `json:"id"`
	VersionID      int64      `json:"version_id"`
	Strategy       int64      `json:"strategy"`
	Salt           string     `json:"salt"`
	BucketStart    int64      `json:"bucket_start"`
	BucketEnd      int64      `json:"bucket_end"`
	Whitelist      string     `json:"whitelist"`
	WhitelistURL   string     `json:"whitelist_url"`
	ManualDownload bool       `json:"manual_download"`
	Mtime          xtime.Time `json:"mtime"`
}

func (g *BusVersionGray) SetWhitelistURL(modCDN map[string]string) bool {
	if g.WhitelistURL == "" {
		return true
	}
	for prefix, host := range modCDN {
		if strings.HasPrefix(g.WhitelistURL, prefix) {
			g.WhitelistURL = host + g.WhitelistURL
			return true
		}
	}
	return false
}

type RoleApply struct {
	ID         int64      `json:"id"`
	AppKey     string     `json:"app_key"`
	PoolID     int64      `json:"pool_id"`
	Username   string     `json:"username"`
	Permission Perm       `json:"permission"`
	Operator   string     `json:"operator"`
	State      ApplyState `json:"state"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
	Pool       *Pool      `json:"pool"`
}

type VersionApply struct {
	ID        int64      `json:"id"`
	AppKey    string     `json:"app_key"`
	Username  string     `json:"username"`
	VersionID int64      `json:"version_id"`
	Operator  string     `json:"operator"`
	Remark    string     `json:"remark"`
	State     ApplyState `json:"state"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
	Pool      *Pool      `json:"pool"`
	Module    *Module    `json:"module"`
	Version   *Version   `json:"version"`
}

type VersionOverView struct {
	ID             int64           `json:"id"`
	AppKey         string          `json:"app_key"`
	Username       string          `json:"username"`
	VersionID      int64           `json:"version_id"`
	Operator       string          `json:"operator"`
	Remark         string          `json:"remark"`
	Ctime          xtime.Time      `json:"ctime"`
	Mtime          xtime.Time      `json:"mtime"`
	Pool           *Pool           `json:"pool"`
	Module         *Module         `json:"module"`
	Version        *Version        `json:"version"`
	OnlineConfig   *OnlineConfig   `json:"online_config"`
	ToOnlineConfig *ToOnlineConfig `json:"to_online_config"`
	OnlineHash     string          `json:"online_hash"`
}

type OnlineConfig struct {
	Config *Config `json:"config"`
	Gray   *Gray   `json:"gray"`
}

type ToOnlineConfig struct {
	Config *ConfigApply `json:"config"`
	Gray   *GrayApply   `json:"gray"`
}

type ConfigApply struct {
	Config
	State ApplyState `json:"state"`
}

type GrayApply struct {
	Gray
	State ApplyState `json:"state"`
}

type SyncPool struct {
	ID     int64       `json:"id"`
	Name   string      `json:"name"`
	Module *SyncModule `json:"module"`
}

type SyncModule struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SyncVersion struct {
	ID      int64 `json:"id"`
	Version int64 `json:"version"`
}

type SyncParam struct {
	VersionID          int64      `form:"version_id" validate:"min=1"`
	ToModuleID         int64      `form:"to_module_id" validate:"min=1"`
	ToVersionID        int64      `form:"to_version_id"`
	ConfigPriority     Priority   `form:"config_priority"`
	ConfigAppVer       string     `form:"config_app_ver"`
	ConfigSysVer       string     `form:"config_sys_ver"`
	ConfigStime        xtime.Time `form:"config_stime"`
	ConfigEtime        xtime.Time `form:"config_etime"`
	GrayStrategy       int64      `form:"gray_strategy"`
	GraySalt           string     `form:"gray_salt"`
	GrayBucketStart    int64      `form:"gray_bucket_start"`
	GrayBucketEnd      int64      `form:"gray_bucket_end"`
	GrayWhitelist      []int64    `form:"gray_whitelist,split" validate:"dive,gt=0"`
	GrayWhitelistURL   string     `form:"gray_whitelist_url"`
	GrayManualDownload bool       `form:"gray_manual_download"`
}

type SyncVersionInfo struct {
	Version *Version `json:"version"`
	File    *File    `json:"file"`
	Config  *Config  `json:"config"`
	Gray    *Gray    `json:"gray"`
}

type Mod struct {
	Pool    *Pool
	Module  *Module
	Version *Version
	File    *File
	Patches []*Patch
	Config  *Config
	Gray    *Gray
}

type Md5DataPair struct {
	MD5  string
	Data map[string]map[string][]*BusFile
}

type Traffic struct {
	Pool                    *Pool
	Module                  *Module
	Version                 *Version
	File                    *File
	Patches                 []*Patch
	Config                  *Config
	Gray                    *Gray
	SetUpUserCount          float64 // 5min启动用户数
	DownloadSizeOnlineBytes float64 // 线上下载量 单位byte
	Operator                string  // 操作人
}

type ReleaseCheckResponse struct {
	CostLevel            int64    `json:"cost_level"`
	Percentage           string   `json:"percentage"`
	Advice               []string `json:"advice"`
	DocUrl               string   `json:"doc_url" json:"doc_url,omitempty"`
	DownloadCount        int64    `json:"download_count,omitempty"`
	OriginFileSize       string   `json:"origin_file_size,omitempty"`
	PatchFileSize        string   `json:"patch_file_size,omitempty"`
	AvgFileSize          string   `json:"avg_file_size,omitempty"`
	CDNBandwidthOnline   string   `json:"cdn_bandwidth_online,omitempty"`
	CDNBandwidthEstimate string   `json:"cdn_bandwidth_estimate,omitempty"`
	CDNBandwidthTotal    string   `json:"cdn_bandwidth_total,omitempty"`
	IsManual             bool     `json:"is_manual"`
}

type TrafficDetail struct {
	AppKey                       string
	PoolName                     string
	ModName                      string
	Operator                     string
	VerNum                       int64
	OriginFileSize               string
	PatchFileSize                string
	AvgFileSize                  string
	DownloadCount                int64
	DownloadSizeEstimate         string // 预估下载量
	DownloadSizeOnline           string // 线上下载量
	DownloadSizeOnlineTotal      string // 发布后的总下载量
	DownloadCDNBandwidthEstimate string // 预估带宽
	DownloadCDNBandwidthOnline   string // 线上带宽
	DownloadCDNBandwidthTotal    string // 发布后的总带宽
	ModUrl                       string
	Percentage                   string      // 上升百分比
	IsManual                     bool        // 是否手动下载
	Cost                         TrafficCost // 发布成本
	Advice                       []string    // 发布意见
	ErrorMsg                     string
	Doc                          string
}

type TrafficNotify struct {
	*TrafficDetail
	OperateType OperateType
}
