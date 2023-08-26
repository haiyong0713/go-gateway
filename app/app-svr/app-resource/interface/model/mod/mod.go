package mod

import (
	xtime "go-common/library/time"
)

type Env string
type Priority string
type Compress string
type Condition string
type Scale string
type Device string
type Arch string
type Lite int32

var (
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

	Scale1x = Scale("1x")
	Scale2x = Scale("2x")
	Scale3x = Scale("3x")

	DevicePhone = Device("phone")
	DevicePad   = Device("pad")

	ArchArmeabiV7a = Arch("armeabi-v7a")
	ArchArm64V8a   = Arch("arm64-v8a")
	ArchX86        = Arch("x86")

	LiteV1 int32 = 1
	LiteV2 int32 = 2
)

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

func (v Scale) Valid() bool {
	return v == Scale1x || v == Scale2x || v == Scale3x
}

func (v Device) Valid() bool {
	return v == DevicePhone || v == DevicePad
}

func (v Arch) Valid() bool {
	return v == ArchArmeabiV7a || v == ArchArm64V8a || v == ArchX86
}

type Pool struct {
	ID     int64  `json:"id,omitempty"`
	AppKey string `json:"app_key,omitempty"`
	Name   string `json:"name,omitempty"`
	Remark string `json:"remark,omitempty"`
}

type Module struct {
	ID       int64    `json:"id,omitempty"`
	PoolID   int64    `json:"pool_id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Remark   string   `json:"remark,omitempty"`
	Compress Compress `json:"compress,omitempty"`
	IsWifi   bool     `json:"is_wifi,omitempty"`
	Deleted  int64    `json:"deleted,omitempty"`
}

type File struct {
	ID          int64          `json:"id,omitempty"`
	VersionID   int64          `json:"version_id,omitempty"`
	Name        string         `json:"name,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
	Size        int64          `json:"size,omitempty"`
	Md5         string         `json:"md5,omitempty"`
	URL         string         `json:"url,omitempty"`
	IsPatch     bool           `json:"is_patch,omitempty"`
	FromVer     int64          `json:"from_ver,omitempty"`
	TotalMd5    string         `json:"total_md5,omitempty"`
	Version     *Version       `json:"version,omitempty"`
	Config      *VersionConfig `json:"config,omitempty"`
	Gray        *VersionGray   `json:"gray,omitempty"`
}

type Version struct {
	ID          int64      `json:"id,omitempty"`
	ModuleID    int64      `json:"module_id,omitempty"`
	Env         Env        `json:"env,omitempty"`
	Version     int64      `json:"version,omitempty"`
	Remark      string     `json:"remark,omitempty"`
	FromVerID   int64      `json:"from_ver_id,omitempty"`
	ReleaseTime xtime.Time `json:"release_time,omitempty"`
	State       string     `json:"state"`
	PoolID      int64      `json:"pool_id,omitempty"`
	PoolName    string     `json:"pool_name,omitempty"`
	ModuleName  string     `json:"module_name,omitempty"`
	Compress    Compress   `json:"compress,omitempty"`
	IsWifi      bool       `json:"is_wifi,omitempty"`
	ZipCheck    bool       `json:"zip_check,omitempty"`
	Mtime       xtime.Time `json:"mtime,omitempty"`
}

type VersionConfig struct {
	ID              int64                 `json:"id,omitempty"`
	VersionID       int64                 `json:"version_id,omitempty"`
	Priority        Priority              `json:"priority,omitempty"`
	AppVer          string                `json:"app_ver,omitempty"`
	SysVer          string                `json:"sys_ver,omitempty"`
	Stime           xtime.Time            `json:"stime,omitempty"`
	Etime           xtime.Time            `json:"etime,omitempty"`
	Scale           string                `json:"scale,omitempty"`
	ForbidenDevice  string                `json:"forbiden_device,omitempty"`
	Arch            string                `json:"arch,omitempty"`
	AppVers         []map[Condition]int64 `json:"app_vers,omitempty"`
	SysVers         []map[Condition]int64 `json:"sys_vers,omitempty"`
	Scales          map[Scale]struct{}    `json:"scales,omitempty"`
	ForbidenDevices map[Device]struct{}   `json:"forbiden_devices,omitempty"`
	Archs           map[Arch]struct{}     `json:"archs,omitempty"`
	Mtime           xtime.Time            `json:"mtime,omitempty"`
}

type VersionGray struct {
	ID             int64              `json:"id,omitempty"`
	VersionID      int64              `json:"version_id,omitempty"`
	Strategy       int64              `json:"strategy,omitempty"`
	Salt           string             `json:"salt,omitempty"`
	BucketStart    int64              `json:"bucket_start,omitempty"`
	BucketEnd      int64              `json:"bucket_end,omitempty"`
	Whitelist      string             `json:"whitelist,omitempty"`
	WhitelistURL   string             `json:"whitelist_url,omitempty"`
	ManualDownload bool               `json:"manual_download,omitempty"`
	Whitelistm     map[int64]struct{} `json:"whitelistm,omitempty"`
	Mtime          xtime.Time         `json:"mtime,omitempty"`
}
