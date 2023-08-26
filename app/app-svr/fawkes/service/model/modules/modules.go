package modules

import (
	ggl "github.com/xanzy/go-gitlab"
)

// Module struct.
type Module struct {
	MID    int64  `json:"id"`
	MName  string `json:"name"`
	MCName string `json:"cname"`
	GID    int64  `json:"-"`
	GName  string `json:"-"`
	GCName string `json:"-"`
}

// Group struct.
type Group struct {
	GID     int64     `json:"gid"`
	GName   string    `json:"gname"`
	GCName  string    `json:"gcname"`
	Modules []*Module `json:"modules,omitempty"`
}

// ModuleSize struct.
type ModuleSize struct {
	ID          int64  `json:"-"`
	Name        string `json:"-"`
	CName       string `json:"-"`
	SizeType    string `json:"-"`
	LibVer      string `json:"lib_ver"`
	Size        int64  `json:"size"`
	BuildID     int64  `json:"build_id"`
	PackVersion string `json:"version"`
	VersionCode int64  `json:"version_code"`
}

// ModuleSizeRes struct.
type ModuleSizeRes struct {
	MID      int64         `json:"id"`
	MName    string        `json:"name"`
	MCName   string        `json:"cname"`
	SizeType string        `json:"size_type"`
	Meta     []*ModuleSize `json:"meta"`
}

// ModuleSizeAdd struct.
type ModuleSizeAdd struct {
	Name     string `json:"name"`
	SizeType string `json:"size_type"`
	LibVer   string `json:"version"`
	Size     int64  `json:"size"`
	Group    string `json:"group"`
}

// ModuleSizeReq struct.
type ModuleSizeReq struct {
	AppKey      string           `json:"app_key"`
	BuildID     int64            `json:"build_id"`
	IsAutoGroup int64            `json:"is_auto_group"`
	Meta        []*ModuleSizeAdd `json:"meta"`
}

// ModuleGroupSize struct.
type ModuleGroupSize struct {
	MID    int64  `json:"id"`
	MName  string `json:"name"`
	MCName string `json:"cname"`
	LibVer string `json:"lib_ver"`
	Size   int64  `json:"size"`
}

// GroupSize struct.
type GroupSize struct {
	Size        int64  `json:"size"`
	PackVersion string `json:"version"`
	VersionCode int64  `json:"version_code"`
	BuildID     int64  `json:"build_id"`
	SizeType    string `json:"-"`
	GID         int64  `json:"-"`
	GName       string `json:"-"`
	GCName      string `json:"-"`
}

// GroupSizeRes struct.
type GroupSizeRes struct {
	GID      int64        `json:"id"`
	GName    string       `json:"name"`
	GCName   string       `json:"cname"`
	SizeType string       `json:"size_type"`
	Meta     []*GroupSize `json:"meta"`
}

// GroupSizeInBuildRes struct.
type GroupSizeInBuildRes struct {
	Size   int64  `json:"size"`
	GID    int64  `json:"id"`
	GName  string `json:"name"`
	GCName string `json:"cname"`
}

// Pipeline struct.
type Pipeline struct {
	ID     int    `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// Job struct.
type Job struct {
	*ggl.Job
	Pipeline *Pipeline `json:"pipeline"`
}

type ModuleConfig struct {
	AppKey          string  `json:"app_key"`
	Version         string  `json:"version"`
	Percentage      float64 `json:"percentage"`
	ModuleGroupID   int64   `json:"module_group_id"`
	TotalSize       int64   `json:"total_size"`
	FixedSize       int64   `json:"fixed_size"`
	ApplyNormalSize int64   `json:"apply_normal_size"`
	ApplyForceSize  int64   `json:"apply_force_size"`
	ExternalSize    int64   `json:"external_size"`
	Description     string  `json:"description"`
	OPERATOR        string  `json:"operator"`
}
