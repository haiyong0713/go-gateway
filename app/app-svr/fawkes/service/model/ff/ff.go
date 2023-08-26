package ff

import (
	"go-gateway/app/app-svr/fawkes/service/model"
)

// ff const.
const (
	FFStatDel      = -1
	FFStatAdd      = 1
	FFStatModify   = 2
	FFStatePublish = 3

	FFPublishHistoryState = -1
	FFPublishNowState     = 1
)

// Version for ff config version.
type Version struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// RomVersion for ff config rom version.
type RomVersion struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

// ResultFF struct.
type ResultFF struct {
	PageInfo  *model.PageInfo `json:"page"`
	ModifyNum int             `json:"mnum"`
	Items     []*FF           `json:"items"`
}

// Whitch struct.
type Whitch struct {
	AppKey   string `json:"app_key"`
	Env      string `json:"env"`
	MID      int64  `json:"mid"`
	Nick     string `json:"nick,omitempty"`
	Operator string `json:"operator"`
	CTime    int64  `json:"ctime,omitempty"`
}

// FF struct.
type FF struct {
	ID          int64  `json:"id"`
	AppKey      string `json:"app_key"`
	Env         string `json:"env"`
	Key         string `json:"key"`
	Operator    string `json:"operator"`
	Status      string `json:"status"`
	Salt        string `json:"salt"`
	Bucket      string `json:"bucket"`
	BucketCount int64  `json:"bucket_count"`
	Whith       string `json:"whith"`
	BlackMid    string `json:"black_mid"`
	Version     string `json:"version"`
	UnVersion   string `json:"un_version"`
	RomVersion  string `json:"rom_version"`
	Brand       string `json:"brand"`
	UnBrand     string `json:"un_brand"`
	Network     string `json:"network"`
	ISP         string `json:"isp"`
	Channel     string `json:"channel"`
	BlackList   string `json:"black_list,omitempty"`
	Desc        string `json:"description"`
	State       int8   `json:"state"`
	CTime       int64  `json:"ctime"`
	MTime       int64  `json:"mtime"`
}

// Publish struct.
type Publish struct {
	WhitchList  string         `json:"wl"`
	PlatForm    string         `json:"p"`
	VID         string         `json:"v"`
	PublishTree []*PublishItem `json:"ab_list"`
}

// PublishItem struct.
type PublishItem struct {
	Name        string         `json:"n"`
	WhitchList  string         `json:"wl,omitempty"`
	BlackList   string         `json:"bl,omitempty"`
	PublishTree *PublishTree   `json:"tree,omitempty"`
	BlackTree   []*PublishTree `json:"black_tree,omitempty"`
}

// PublishTree struct.
type PublishTree struct {
	OP          string       `json:"op,omitempty"`
	Prop        string       `json:"prop,omitempty"`
	Value       string       `json:"val,omitempty"`
	Salt        string       `json:"s,omitempty"`
	Logic       string       `json:"l,omitempty"`
	Bucket      string       `json:"b,omitempty"`
	BucketCount int64        `json:"bc,omitempty"`
	Son         *PublishTree `json:"son,omitempty"`
}

// ConfigPublish struct.
type ConfigPublish struct {
	ID             int64  `json:"id"`
	AppKey         string `json:"app_key"`
	Env            string `json:"env"`
	Desc           string `json:"description"`
	URL            string `json:"cdn_url"`
	LocalPath      string `json:"local_path,omitempty"`
	Diffs          string `json:"diffs,omitempty"`
	TotalURL       string `json:"total_url,omitempty"`
	TotalLocalPath string `json:"total_path,omitempty"`
	Operator       string `json:"operator"`
	State          int8   `json:"state"`
	CTime          int64  `json:"ctime"`
	MTime          int64  `json:"mtime"`
}

// File struct.
type File struct {
	ID          int64  `json:"id,omitempty"`
	AppKey      string `json:"app_key,omitempty"`
	Env         string `json:"env,omitempty"`
	FVID        int64  `json:"fvid,omitempty"`
	Key         string `json:"key,omitempty"`
	Desc        string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Salt        string `json:"salt,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	BucketCount int64  `json:"bucket_count,omitempty"`
	Whith       string `json:"whith,omitempty"`
	BlackMid    string `json:"black_mid,omitempty"`
	Version     string `json:"version,omitempty"`
	UnVersion   string `json:"un_version,omitempty"`
	RomVersion  string `json:"rom_version,omitempty"`
	Brand       string `json:"brand"`
	UnBrand     string `json:"un_brand"`
	Network     string `json:"network,omitempty"`
	ISP         string `json:"isp,omitempty"`
	Channel     string `json:"channel,omitempty"`
	BlackList   string `json:"black_list,omitempty"`
	State       int8   `json:"state,omitempty"`
	Operator    string `json:"operator,omitempty"`
	MTime       int64  `json:"mtime,omitempty"`
}

// Diff config diff struct.
type Diff struct {
	AppKey   string    `json:"app_key,omitempty"`
	Env      string    `json:"env,omitempty"`
	CVID     int64     `json:"cvid,omitempty"`
	Key      string    `json:"key,omitempty"`
	State    int8      `json:"type,omitempty"`
	Origin   *DiffItem `json:"origin,omitempty"`
	New      *DiffItem `json:"new,omitempty"`
	Operator string    `json:"operator,omitempty"`
	Desc     string    `json:"description,omitempty"`
	MTime    int64     `json:"mtime,omitempty"`
}

// DiffItem struct.
type DiffItem struct {
	Status      string      `json:"status,omitempty"`
	Salt        string      `json:"salt,omitempty"`
	Bucket      string      `json:"bucket,omitempty"`
	BucketCount int64       `json:"bucket_count,omitempty"`
	Whith       string      `json:"whith,omitempty"`
	BlackMid    string      `json:"black_mid,omitempty"`
	Version     *Version    `json:"version,omitempty"`
	UnVersion   string      `json:"un_version,omitempty"`
	RomVersion  string      `json:"rom_version,omitempty"`
	Brand       string      `json:"brand,omitempty"`
	UnBrand     string      `json:"un_brand,omitempty"`
	Network     string      `json:"network,omitempty"`
	ISP         string      `json:"isp,omitempty"`
	Channel     string      `json:"channel,omitempty"`
	BlackList   []*DiffItem `json:"black_list,omitempty"`
}

// FF modify count
type ModifyCount struct {
	Test int `json:"test"`
	Prod int `json:"prod"`
}

// HistoryResult struct for cd list.
type HistoryResult struct {
	PageInfo *model.PageInfo  `json:"page"`
	Items    []*ConfigPublish `json:"items"`
}
