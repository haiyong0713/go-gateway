package model

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-job/job/model/space"
)

const (
	ActionUpView           = "upView"
	ActionUpStat           = "upStat"
	ActionUpContribute     = "upContribute"
	ActionUpContributeAid  = "upContributeAid"
	ActionUpViewContribute = "upViewContribute"
	ActionUpAccount        = "upAccount"
)

type Retry struct {
	Action string `json:"action,omitempty"`
	Data   struct {
		Mid           int64         `json:"mid,omitempty"`
		Aid           int64         `json:"aid,omitempty"`
		Attrs         *space.Attrs  `json:"attrs,omitempty"`
		Items         []*space.Item `json:"item,omitempty"`
		Time          xtime.Time    `json:"time,omitempty"`
		IP            string        `json:"ip,omitempty"`
		Action        string        `json:"action,omitempty"`
		IsCooperation bool          `json:"is_cooperation,omitempty"`
		Aids          []int64       `json:"aids,omitempty"`
		IsComic       bool          `json:"is_comic,omitempty"`
	} `json:"data,omitempty"`
}
