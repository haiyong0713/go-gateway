package mod

import (
	"encoding/json"
)

type BinlogMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	Schema string          `json:"schema"`
	Old    json.RawMessage `json:"old"`
	New    json.RawMessage `json:"new"`
}

// BinlogModVersion Comment: mod 资源版本
type BinlogModVersion struct {
	// Comment: 主键ID
	ID int64 `json:"id"`
	// Comment: module_id
	// Default: 0
	ModuleID int64 `json:"module_id"`
	// Comment: 环境
	Env string `json:"env"`
	// Comment: 版本号
	// Default: 0
	Version int64 `json:"version"`
	// Comment: 备注
	Remark string `json:"remark"`
	// Comment: env是prod时有效,标明是从哪个测试vesion_id推送到线上
	// Default: 0
	FromVerID int64 `json:"from_ver_id"`
	// Comment: 已发布,0-false,1-true
	// Default: 0
	Released int64 `json:"released"`
	// Comment: 发布时间
	// Default: 0000-00-00 00:00:00
	ReleaseTime string `json:"release_time"`
	// Comment: 状态
	State string `json:"state"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime string `json:"ctime"`
	// Comment: 更新时间
	// Default: CURRENT_TIMESTAMP
	Mtime string `json:"mtime"`
}
