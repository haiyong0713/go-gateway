package module

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"

	xtime "go-common/library/time"
)

const (
	Total       = 0
	Incremental = 1

	EnvRelease = "1"
	EnvTest    = "2"
	EnvDefault = "3"

	NotValid = int8(0)
	Valid    = int8(1)

	md5Format = "mobi_app=%s&device=%s&build=%d&ts=%d&buvid=%s"
)

type ResourcePool struct {
	ID        int         `json:"-"`
	Name      string      `json:"name"`
	Resources []*Resource `json:"resources,omitempty"`
}

type Resource struct {
	ID           int           `json:"-"`
	ResID        int           `json:"-"`
	Name         string        `json:"name"`
	Compresstype int           `json:"compresstype"`
	Type         string        `json:"type"`
	URL          string        `json:"url"`
	MD5          string        `json:"md5"`
	TotalMD5     string        `json:"total_md5"`
	Size         int           `json:"size"`
	Version      int           `json:"ver"`
	Increment    int           `json:"increment"`
	FromVer      int           `json:"-"`
	Filename     string        `json:"-"`
	Condition    *Condition    `json:"-"`
	Level        int           `json:"level,omitempty"`
	IsWifi       int8          `json:"is_wifi"`
	Gray         *ResourceGray `json:"-"`
	PoolName     string        `json:"-"`
	// grpc 使用
	PoolID    int64      `json:"pool_id"`
	ModuleID  int64      `json:"module_id"`
	VersionID int64      `json:"version_id"`
	FileID    int64      `json:"file_id"`
	Mtime     xtime.Time `json:"mtime"`
}

type Condition struct {
	ID        int                          `json:"-"`
	ResID     int                          `json:"-"`
	STime     xtime.Time                   `json:"stime"`
	ETime     xtime.Time                   `json:"etime"`
	Valid     int8                         `json:"valid"`
	ValidTest int8                         `json:"valid_test"`
	Default   int                          `json:"-"`
	Columns   map[string]map[int][]*Column `json:"columns"`
	IsWifi    int8                         `json:"-"`
	Mtime     xtime.Time                   `json:"-"`
}

type Column struct {
	Condition string     `json:"condition"`
	Value     string     `json:"value"`
	Mtime     xtime.Time `json:"mtime"`
}

type Versions struct {
	PoolName string `json:"name"`
	Resource []struct {
		ResourceName string      `json:"name"`
		Version      interface{} `json:"ver"`
	} `json:"resources"`
}

func ModuleMd5(mobiApp, device string, build, ts int, buvid string, ps url.Values) (res string) {
	str := fmt.Sprintf(md5Format, mobiApp, device, build, ts, buvid)
	params := md5.Sum([]byte(str))
	res = hex.EncodeToString(params[:])
	return
}

// ResourceGray 静态资源灰度策略
type ResourceGray struct {
	// Comment: 主键id
	ID int64 `json:"id"`
	// Comment: 资源id
	// Default: 0
	ResourceID int `json:"resource_id"`
	// Comment: 策略 1-(UID MD5) 2-(DEVICE MD5) 3-(UID)
	// Default: 0
	Strategy int `json:"strategy"`
	// Comment: 盐值
	Salt string `json:"salt"`
	// Comment: 桶开始
	// Default: 0
	BucketStart int `json:"bucket_start"`
	// Comment: 桶结束
	// Default: 0
	BucketEnd int `json:"bucket_end"`
	// Comment: 手动输入白名单
	WhitelistInput string `json:"whitelist_input"`
	// Comment: 上传白名单url
	WhitelistUpload string `json:"whitelist_upload"`
	// Comment: 1-允许手动更新
	// Default: 0
	ManualUpdate int `json:"manual_update"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime xtime.Time `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime xtime.Time `json:"mtime"`
	// extra
	Whitelist map[int64]struct{} `json:"-"`
}
