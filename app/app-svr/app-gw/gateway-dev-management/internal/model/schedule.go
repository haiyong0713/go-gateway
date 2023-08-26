package model

import "go-common/library/time"

type Stuff struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type GatewaySchedule struct {
	Id    int
	Key   string
	Value string
	Ctime time.Time
	Mtime time.Time
}

type SchedulerReq struct {
	StartDay    string   `json:"start_day"`
	EndDay      string   `json:"end_day"`
	Level       int      `json:"level"`
	Users       []string `json:"users"`
	SchedulerId int      `json:"scheduler_id"`
}
