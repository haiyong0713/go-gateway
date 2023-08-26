package model

import (
	xtime "go-common/library/time"
)

const (
	// table state
	StateInvalid   = 0
	StateValid     = 1
	DynamicType    = 1
	PageOnLine     = 1
	PageOffLine    = 2
	PageWaitOnLine = 0
	PageFromUid    = 1
	TabTypeUpAct   = "up_act"
)

// NatPage .
type NatPage struct {
	ID        int64      `json:"id"`
	ForeignID int64      `json:"foreign_id"`
	Stime     xtime.Time `json:"stime"`
	Type      int64      `json:"type"`
}
