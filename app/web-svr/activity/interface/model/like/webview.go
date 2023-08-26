package like

import (
	xtime "go-common/library/time"
)

type WebDataItem struct {
	ID    int64      `json:"id"`
	VID   int64      `json:"vid"`
	State int64      `json:"state"`
	Name  string     `json:"name"`
	Raw   []byte     `json:"data"`
	STime xtime.Time `json:"stime"`
	ETime xtime.Time `json:"etime"`
}
