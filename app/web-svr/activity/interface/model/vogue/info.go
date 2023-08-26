package model

type ShareInfo struct {
	User  *ShareInfoItem `json:"user"`
	Prize *ShareInfoItem `json:"prize"`
}

type ShareInfoItem struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type PrizeInfo struct {
	User   *PrizeInfoItem `json:"user"`
	Prize  *PrizeInfoItem `json:"prize"`
	Source string         `json:"source"`
	Time   int64          `json:"time"`
}

type PrizeInfoItem struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type WinListItem struct {
	Mid        int64  `json:"mid"`
	Name       string `json:"name"`
	GiftId     int64  `json:"gift_id"`
	GiftName   string `json:"gift_name"`
	GiftImgUrl string `json:"gift_img_url"`
	CTime      int64  `json:"ctime"`
}

type RiskInfo struct {
	Level    int64
	Score    int64
	Message  string
	HitRules []string
}

type FavInfo struct {
	Bvid string `json:"bvid"`
	Type int64  `json:"type"`
}
