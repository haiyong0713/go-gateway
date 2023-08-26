package model

import (
	"database/sql"
	"encoding/json"
	"time"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

// fawkes-admin const
const (
	BFSBucket               = "fawkes"
	TestEnv                 = 1
	ProdEnv                 = 10
	LaserTypeTask           = "task"
	LaserTypeUser           = "user"
	LaserParseStatusSuccess = 3
	LaserParseStatusRunning = 2
	LaserParseStatusWaiting = 1
	LaserParseStatusDefault = 0
	LaserParseStatusFailed  = -1
	SystemAuto              = "system_auto"
)

// PageInfo struct.
type PageInfo struct {
	Total int `json:"total"`
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
}

// Version struct.
type Version struct {
	ID          int64  `json:"version_id"`
	BuildID     int64  `json:"build,omitempty"`
	AppID       string `json:"app_id,omitempty"`
	AppKey      string `json:"app_key,omitempty"`
	Env         string `json:"env"`
	CIEnvVars   string `json:"ci_env_vars"`
	Version     string `json:"version"`
	VersionCode int64  `json:"version_code"`
	IsUpgrade   int8   `json:"is_upgrade"`
	CTime       int64  `json:"ctime"`
	MTime       int64  `json:"mtime"`
}

// ChannelResult struct.
type ChannelResult struct {
	Page    *PageInfo         `json:"page"`
	Channel []*appmdl.Channel `json:"items"`
}

// LaserResult struct.
type LaserResult struct {
	PageInfo *PageInfo       `json:"page"`
	Items    []*appmdl.Laser `json:"items"`
}

// LaserPendingResult struct
type LaserPendingResult struct {
	LogUploadList []*appmdl.Laser `json:"log_upload_list"`
	CommandList   []*appmdl.Laser `json:"command_list"`
}

// LaserCmdResult struct.
type LaserCmdResult struct {
	PageInfo *PageInfo          `json:"page"`
	Items    []*appmdl.LaserCmd `json:"items"`
}

// FormEnv form env.
func FormEnv(env string) int {
	switch env {
	case "test":
		return TestEnv
	case "prod":
		return ProdEnv
	}
	return TestEnv
}

// EvolutionEnv from func Evolution.
func EvolutionEnv(env string) (nextEnv string) {
	switch env {
	case "test":
		nextEnv = "prod"
	}
	return
}

// HfList is hotfix list response struct
type HfList struct {
	PageInfo PageInfo             `json:"page"`
	Items    []*appmdl.HfListItem `json:"items"`
}

// TotalFile config && ff file struct.
type TotalFile struct {
	MCV      string                  `json:"v"`
	Platform string                  `json:"p"`
	Version  map[string]*FileVersion `json:"version"`
}

// FileVersion config && ff struct.
type FileVersion struct {
	Diffs   map[string]string `json:"d,omitempty"`
	Md5     string            `json:"m"`
	URL     string            `json:"u"`
	Version string            `json:"v"`
}

// Diff struct.
type Diff struct {
	URL  string `json:"url"`
	MD5  string `json:"md5"`
	Size int64  `json:"size"`
}

// SagaReq struct.
type SagaReq struct {
	ToUser  []string `json:"touser"`
	Content string   `json:"content"`
}

// SagaRes struct.
type SagaRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type Broadcast struct {
	Username string      `json:"username"`
	Bots     string      `json:"bots"`
	Hook     *Hook       `json:"hook"`
	Param    interface{} `json:"param"`
}

type Hook struct {
	URI    string `json:"uri"`
	Method string `json:"method"`
}

// FeedbackList struct.
type FeedbackList struct {
	PageInfo *PageInfo             `json:"page"`
	Items    []*appmdl.FeedbackRes `json:"items"`
}

type NullTime struct {
	sql.NullTime
}

func (v *NullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (v NullTime) UnmarshalJSON(data []byte) error {
	var s *time.Time
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		v.Valid = true
		v.Time = *s
	} else {
		v.Valid = false
	}
	return nil
}
