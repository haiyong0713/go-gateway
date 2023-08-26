package search

const (
	ChannelHide = 0
	ChannelOK   = 2
)

type SearchChannel struct {
	CID   int64 `json:"cid"`
	State int   `json:"state"`
}
