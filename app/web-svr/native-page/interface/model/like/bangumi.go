package like

type EpPlayer struct {
	AID       int64  `json:"aid"`
	CID       int64  `json:"cid"`
	EpID      int64  `json:"episode_id"`
	Uri       string `json:"url"`
	Cover     string `json:"cover"`
	NewDesc   string `json:"new_desc"`
	Duration  int64  `json:"duration"`
	ShowTitle string `json:"show_title"`
	Stat      struct {
		Play    int64 `json:"play"`
		Danmaku int64 `json:"danmaku"`
		Follow  int64 `json:"follow"`
	} `json:"stat"`
	Season struct {
		Title    string `json:"title"`
		TypeName string `json:"type_name"`
		Type     int64  `json:"type"`
	}
	Dimension struct {
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
		Rotate int64 `json:"rotate"`
	}
}
