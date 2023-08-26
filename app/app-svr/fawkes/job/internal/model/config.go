package model

import "time"

// ClearNasJob config
type ClearNasJob struct {
	Corn     string
	PackType []int64
	Start    time.Time
	End      time.Time
	AppKey   string
}

// OutCfg config
type OutCfg struct {
	FAWKES string
	NAME   string
	METHOD string
	LIST   string
	DELETE string
}
