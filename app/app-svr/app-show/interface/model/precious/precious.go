package precious

import (
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
)

// Precious .
type Precious struct {
	H5Title     string          `json:"h5_title,omitempty"`
	Explain     string          `json:"explain,omitempty"`
	MediaID     int64           `json:"media_id,omitempty"`
	Card        []cardm.Handler `json:"card,omitempty"`
	LatestCard  []cardm.Handler `json:"latest_card,omitempty"`
	OriginCard  []cardm.Handler `json:"origin_card,omitempty"`
	LatestCount int64           `json:"latest_count,omitempty"`
	Subscribed  bool            `json:"subscribed"`

	PageSubTitle   string `json:"page_sub_title,omitempty"`
	ShareMainTitle string `json:"share_main_title,omitempty"`
	ShareSubTitle  string `json:"share_sub_title,omitempty"`
}
