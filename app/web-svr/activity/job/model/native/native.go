package native

import xtime "go-common/library/time"

// DynamicType .
const (
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
	Etime     xtime.Time `json:"etime"`
	Type      int64      `json:"type"`
}
