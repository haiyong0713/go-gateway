package model

type EpPlayer struct {
	AID       int64  `json:"aid"`
	CID       int64  `json:"cid"`
	EpID      int64  `json:"episode_id"`
	Uri       string `json:"url"`
	Cover     string `json:"cover"`
	NewDesc   string `json:"new_desc"`
	ShowTitle string `json:"show_title"`
	Duration  int64  `json:"duration"`
	IsPreview int    `json:"is_preview"`
	Stat      struct {
		Play    int64 `json:"play"`
		Danmaku int64 `json:"danmaku"`
		Follow  int64 `json:"follow"`
	} `json:"stat"`
	Season struct {
		Type     int64  `json:"type"`
		Title    string `json:"title"`
		TypeName string `json:"type_name"`
		SeasonID int64  `json:"season_id"`
	} `json:"season"`
}
