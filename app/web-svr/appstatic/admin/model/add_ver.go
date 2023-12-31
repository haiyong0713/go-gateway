package model

import (
	"go-common/library/time"
)

const (
	BfsMaxSize  = 20 * 1024 * 1024  // 20M BFS
	BossMaxSize = 100 * 1024 * 1024 // 100M BFS
	BossBucket  = "appstaticboss"
	UploadBfs   = 1
	UploadBoss  = 2
)

// Limit def
type Limit struct {
	MobiApp   []string   // white list
	Device    []string   // black list
	Plat      []string   // white list
	Build     *Build     // build range
	TimeRange *TimeRange // time range
	Sysver    *Build     // system version
	Scale     []string
	Arch      []string
	Level     []string
	IsWifi    int // only wifi download
}

// Build def
type Build struct {
	LT int `json:"lt"` // less than
	GT int `json:"gt"` // great than
	LE int `json:"le"` // less than or equal
	GE int `json:"ge"` // great than or equal
}

// CheckRange  whether the build range is valid
func (build *Build) CheckRange() (res bool) {
	if (build.GE != 0 && build.GT != 0) || (build.LE != 0 && build.LT != 0) { // two values by one side
		return false
	}
	var (
		gt = build.GT
		lt = build.LT
	)
	// transform E to T
	if build.GE != 0 {
		gt = build.GE - 1
	}
	if build.LE != 0 {
		lt = build.LE + 1
	}
	// range check
	if lt != 0 && gt != 0 && lt-gt <= 1 {
		return false
	}
	return true
}

// TimeRange def
type TimeRange struct {
	Stime time.Time `json:"stime"`
	Etime time.Time `json:"etime"`
}

// ResourceLimit def
type ResourceLimit struct {
	ID        int64
	ConfigID  int64
	Column    string
	Condition string
	Value     string
	IsDeleted int8
	Mtime     time.Time
	Ctime     time.Time
}

// ResourceConfig def
type ResourceConfig struct {
	ID             int64
	ResourceID     int64 `gorm:"column:resource_id"`
	Stime          time.Time
	Etime          time.Time
	Valid          int8
	IsDeleted      int8 `gorm:"column:is_deleted"`
	Mtime          time.Time
	Ctime          time.Time
	DefaultPackage int8 `gorm:"column:default_package"`
	IsWifi         int  `gorm:"column:is_wifi"`
}

// FileBig .
type FileBig struct {
	URL string `json:"url"`
}

// FileInfo : the uploaded file information
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
	Md5  string `json:"md5"`
	URL  string `json:"url"`
}

// ResourceFile represents the table structure
type ResourceFile struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Md5        string    `json:"md5"`
	Size       int       `json:"size"`
	URL        string    `json:"url"`
	ResourceID int       `json:"resource_id"`
	Ctime      time.Time `json:"ctime"`
	Mtime      time.Time `json:"mtime"`
	FileType   int8      `json:"file_type"`
	FromVer    int64     `json:"from_ver"`
	IsDeleted  int8      `json:"is_deleted"`
}

// Resource reprensents the resource table
type Resource struct {
	ID      int64     `json:"id" params:"id"`
	Name    string    `json:"name" params:"name"`
	Version int64     `json:"version" params:"version"`
	PoolID  int64     `json:"pool_id" params:"pool_id"`
	Ctime   time.Time `json:"ctime" params:"ctime"`
	Mtime   time.Time `json:"mtime" params:"mtime"`
}

// ResourcePool reprensents the resource_pool table
type ResourcePool struct {
	ID       int64     `json:"id" params:"id"`
	Name     string    `json:"name" params:"name"`
	Platform string    `json:"platform" params:"platform"`
	Ctime    time.Time `json:"ctime" params:"ctime"`
	Mtime    time.Time `json:"mtime" params:"mtime"`
}

// Department reprensents the resource_department table
type Department struct {
	ID        int64     `json:"id" params:"id"`
	Name      string    `json:"name" params:"name"`
	Ctime     time.Time `json:"ctime" params:"ctime"`
	Mtime     time.Time `json:"mtime" params:"mtime"`
	Desc      string    `json:"desc" params:"desc"`
	IsDeleted uint8     `json:"is_deleted" params:"is_deleted"`
}

// ResponseNas represents the NAS response struct
type ResponseNas struct {
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

// RequestVer is the struct of the request to upload an new version's package
type RequestVer struct {
	Department     string   `form:"department" validate:"required"`
	DefaultPackage int      `form:"default_package" validate:"min=0,max=1"`
	ResName        string   `form:"res_name" validate:"required"`
	ModName        string   `form:"mod_name" validate:"required"`
	MobiAPP        []string `form:"mobi_app,split"`
	Plat           []string `form:"plat,split"`
	Device         []string `form:"device,split"`
	BuildRange     string   `form:"build_range"`
	TimeRange      string   `form:"time_range"`
	Sysver         string   `form:"sysver"`
	Arch           []int    `form:"arch,split" validate:"dive,min=1,max=3"`
	Level          int      `form:"level" validate:"min=0,max=3"`
	Scale          []int    `form:"scale,split" validate:"dive,min=1,max=3"`
	IsWifi         int      `form:"is_wifi" validate:"max=1"`
}

// ResourceGray .
type ResourceGray struct {
	ID              int64     `form:"id" json:"id"`
	ResourceId      int64     `form:"resource_id" json:"resource_id" validate:"required"`
	Strategy        int64     `form:"strategy" json:"strategy" validate:"required"`
	Salt            string    `form:"salt" json:"salt" validate:"required,max=50"`
	BucketStart     int64     `form:"bucket_start" json:"bucket_start"`
	BucketEnd       int64     `form:"bucket_end" json:"bucket_end"`
	WhitelistInput  string    `form:"whitelist_input" json:"whitelist_input"`
	WhitelistUpload string    `form:"whitelist_upload" json:"whitelist_upload"`
	ManualUpdate    int64     `form:"manual_update" json:"manual_update"`
	Ctime           time.Time `form:"ctime" json:"ctime"`
	Mtime           time.Time `form:"mtime" json:"mtime"`
}

// TableName ResourceGray
func (*ResourceGray) TableName() string {
	return "resource_gray"
}

// RespAdd is the structure for add ver return
type RespAdd struct {
	ResID   int `json:"res_id"`
	Version int `json:"version"`
}

// TableName gives the table name of the model
func (*Resource) TableName() string {
	return "resource"
}

// TableName gives the table name of the model
func (*ResourcePool) TableName() string {
	return "resource_pool"
}

// TableName gives the table name of the model
func (*ResourceFile) TableName() string {
	return "resource_file"
}

// TableName gives the table name of the model
func (*ResourceLimit) TableName() string {
	return "resource_limit"
}

// TableName gives the table name of the model
func (*ResourceConfig) TableName() string {
	return "resource_config"
}

// TableName gives the table name of the model
func (*Department) TableName() string {
	return "resource_department"
}
