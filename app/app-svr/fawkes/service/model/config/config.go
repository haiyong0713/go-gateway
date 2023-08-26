package config

import (
	"go-gateway/app/app-svr/fawkes/service/model"
)

// const for Config.
const (
	ConfigPublishStateHistory = -1
	ConfigPublishStateNow     = 1

	ConfigStatDel      = -1
	ConfigStatAdd      = 1
	ConfigStatModify   = 2
	ConfigStatePublish = 3
)

const (
	BusinessShareChannel      = "business_share_channel"
	BusinessActiveListChannel = "business_active_list_channel"
)

// VersionResult struct for cd list.
type VersionResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Version      `json:"items"`
}

// HistoryResult struct for cd list.
type HistoryResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Publish      `json:"items"`
}

// Version config struct.
type Version struct {
	ID          int64  `json:"cvid"`
	AppKey      string `json:"app_key"`
	Env         string `json:"env"`
	Version     string `json:"version"`
	VersionCode int64  `json:"version_code"`
	CV          int64  `json:"config_version"`
	Operator    string `json:"operator,omitempty"`
	Desc        string `json:"description,omitempty"`
	ModifyNum   int    `json:"mnum"`
	PTime       int64  `json:"ptime,omitempty"`
}

// Config config struct.
type Config struct {
	AppKey   string `json:"app_key,omitempty"`
	Env      string `json:"env,omitempty"`
	CVID     int64  `json:"cvid,omitempty"`
	CV       int64  `json:"cv,omitempty"`
	Group    string `json:"group,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value"`
	State    int8   `json:"type,omitempty"`
	Operator string `json:"operator,omitempty"`
	Desc     string `json:"description,omitempty"`
	MTime    int64  `json:"mtime,omitempty"`
}

// Diff config diff struct.
type Diff struct {
	AppKey   string  `json:"app_key,omitempty"`
	Env      string  `json:"env,omitempty"`
	CVID     int64   `json:"cvid,omitempty"`
	Group    string  `json:"group,omitempty"`
	Key      string  `json:"key,omitempty"`
	State    int8    `json:"type,omitempty"`
	Origin   *Config `json:"origin,omitempty"`
	New      *Config `json:"new,omitempty"`
	Operator string  `json:"operator,omitempty"`
	Desc     string  `json:"description,omitempty"`
	MTime    int64   `json:"mtime,omitempty"`
}

// Publish struct.
type Publish struct {
	ID             string `json:"id"`
	AppKey         string `json:"app_key"`
	Env            string `json:"env"`
	CVID           int64  `json:"cvid"`
	Version        string `json:"version,omitempty"`
	VersionCode    int64  `json:"version_code,omitempty"`
	CV             int64  `json:"config_version"`
	State          int64  `json:"state"`
	MD5            string `json:"md5,omitempty"`
	URL            string `json:"cdn_url,omitempty"`
	LocalPath      string `json:"local_path,omitempty"`
	Diffs          string `json:"diffs,omitempty"`
	TotalURL       string `json:"total_url,omitempty"`
	TotalLocalPath string `json:"total_path,omitempty"`
	Operator       string `json:"operator,omitempty"`
	Desc           string `json:"description,omitempty"`
	CTime          int64  `json:"ctime,omitempty"`
	PTime          int64  `json:"ptime,omitempty"`
}

// ParamsConfig struct.
type ParamsConfig struct {
	AppKey    string    `json:"app_key"`
	Env       string    `json:"env"`
	GroupName string    `json:"group_name"`
	CVID      int64     `json:"cvid"`
	Desc      string    `json:"description"`
	Items     []*Config `json:"items"`
	Operator  string    `json:"operator"`
	Business  string    `json:"business"`
}

// Config modify count
type ModifyCount struct {
	Test int `json:"test"`
	Prod int `json:"prod"`
}

type PubConfig struct {
	Group string `json:"group"`
	Key   string `json:"key"`
}
