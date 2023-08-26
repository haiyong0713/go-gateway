package like

// StatLikeMsg .
type StatLikeMsg struct {
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	Count     int64  `json:"count"`
	Timestamp int64  `json:"timestamp"`
	Mid       int64  `json:"mid"`
	UpMid     int64  `json:"up_mid"`
	Action    int64  `json:"action"`
}

// StatCoinMsg .
type StatCoinMsg struct {
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	Count     int64  `json:"count"`
	Timestamp int64  `json:"timestamp"`
	Mid       int64  `json:"mid"`
	UpMid     int64  `json:"up_mid"`
}

// ArticleMsg .
type ArticleMsg struct {
	ID  int64 `json:"article_id"`
	Mid int64 `json:"mid"`
}
