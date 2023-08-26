package aggregation

import (
	"fmt"

	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
)

const (
	_h5URL = "https://www.bilibili.com/h5/hot-gather?hotword_id=%d&navhide=1"
)

// Aggregation def.
type Aggregation struct {
	ID       int64  `json:"id"`
	HotTitle string `json:"hot_title"`
	State    int    `json:"state"`
	Image    string `json:"image"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

// AggRes .
type AggRes struct {
	H5Title string          `json:"h5_title,omitempty"`
	Desc    string          `json:"desc,omitempty"`
	Image   string          `json:"image,omitempty"`
	Card    []cardm.Handler `json:"card,omitempty"`
}

// ToResH5URl .
func ToResH5URl(hotID int64) string {
	return fmt.Sprintf(_h5URL, hotID)
}
