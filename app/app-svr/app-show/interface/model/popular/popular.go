package popular

type PopularArchiveRequest struct{}

type PopularArchiveReply struct {
	List []*PopularArchiveItem `json:"list"`
}

type PopularArchiveItem struct {
	Aid      int64  `json:"aid"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	URI      string `json:"uri"`
	Play     int64  `json:"play"`
	Danmaku  int64  `json:"danmaku"`
	Duration string `json:"duration"`
}
