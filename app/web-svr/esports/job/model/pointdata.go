package model

import "sync"

// DataMsg live data  message.
type DataMsg struct {
	Match struct {
		ID int64 `json:"id"`
	} `json:"match"`
}

// ContestData contests data.
type ContestData struct {
	CID     int64 `json:"cid"`
	MatchID int64 `json:"match_id"`
	Stime   int64 `json:"stime"`
	Etime   int64 `json:"etime"`
}

// SyncMatch store leida match list
type SyncMatch struct {
	Data map[int64]*ContestData
	sync.Mutex
}
