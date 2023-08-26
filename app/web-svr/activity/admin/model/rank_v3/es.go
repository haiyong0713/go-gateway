package rank

// ResultFromES ...
type ResultFromES struct {
	BaseID     int64  `json:"base_id"`
	AID        int64  `json:"aid"`
	RankID     int64  `json:"rank_id"`
	MID        int64  `json:"mid"`
	TagID      int64  `json:"tag_id"`
	LogID      int64  `json:"log_id"`
	LogDate    int64  `json:"log_date"`
	Batch      int64  `json:"batch"`
	WhiteScore int64  `json:"white_score"`
	CountScore int64  `json:"count_score"`
	Score      int64  `json:"score"`
	LastScore  int64  `json:"last_score"`
	TodayScore int64  `json:"today_score"`
	RankType   int64  `json:"rank_type"`
	ID         string `json:"id"`
	LikesScore int64  `json:"likes_score"`
	PlayScore  int64  `json:"play_score"`
	CoinScore  int64  `json:"coin_score"`
	ShareScore int64  `json:"share_score"`
	SourceID   int64  `json:"source_id"`
	Rank       int64  `json:"rank"`
}
