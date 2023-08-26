package model

import xtime "go-common/library/time"

// Kv baidu kv struct.
type Kv struct {
	ID    int32      `json:"id"`
	Name  string     `json:"name"`
	Pic   string     `json:"pic"`
	URL   string     `json:"url"`
	ResID int32      `json:"resource_id"`
	STime xtime.Time `json:"stime"`
	ETime xtime.Time `json:"etime"`
}
