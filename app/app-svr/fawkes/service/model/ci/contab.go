package ci

import "go-gateway/app/app-svr/fawkes/service/model"

const (
	CronStop = -1
	CronWait = 0
	CronRun  = 1
)

// ContabResult struct log list.
type ContabResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Contab       `json:"items"`
}

type Contab struct {
	ID        int64  `json:"id"`
	AppKey    string `json:"app_key"`
	STime     int64  `json:"stime"`
	Tick      string `json:"tick"`
	GitType   int    `json:"git_type"`
	GitName   string `json:"git_name"`
	PkgType   int    `json:"pkg_type"`
	BuildID   int64  `json:"build_id"`
	CIEnvVars string `json:"ci_env_vars"`
	Send      string `json:"send"`
	State     int8   `json:"state"`
	Operator  string `json:"operator"`
}
