package model

import (
	"encoding/json"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-job/job/model/space"
)

type ArcMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
}

type AccMsg struct {
	Mid    int64  `json:"mid"`
	Action string `json:"action"`
}

type StatMsg struct {
	Type         string `json:"type,omitempty"`
	ID           int64  `json:"id,omitempty"`
	Count        int32  `json:"count,omitempty"`
	DislikeCount int32  `json:"dislike_count,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	BusType      string
}

type ContributeMsg struct {
	Vmid          int64        `json:"vmid"`
	CTime         xtime.Time   `json:"ctime"`
	Attrs         *space.Attrs `json:"attrs"`
	IP            string       `json:"ip"`
	IsCooperation bool         `json:"is_cooperation"`
	IsComic       bool         `json:"is_comic"`
}
