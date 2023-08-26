package ranklist

// Meta is
type Meta struct {
	ID         int64 `json:"id"`
	RankConfig struct {
		Title        string         `json:"title"` // 标题
		Cover        string         `json:"cover"` // 封面图
		Description  []*Description `json:"description"`
		HelpTips     []string       `json:"help_tips"` // 助力提示文案集
		STime        int64          `json:"stime"`
		ETime        int64          `json:"etime"`
		Cycle        int64          `json:"cycle"`
		PerUpdate    int64          `json:"per_update"`
		Tids         []int64        `json:"tids"`
		ActIDs       []int64        `json:"act_ids"` // 活动 ID
		ArchiveSTime int64          `json:"archive_stime"`
		ArchiveETime int64          `json:"archive_etime"`
	} `json:"rank_config"`
	RankVideos []int64          `json:"rank_videos"`
	RankState  int64            `json:"rank_state"`
	FinalRank  []*FinalRankItem `json:"final_rank"`
}

// FinalRankItem is
type FinalRankItem struct {
	Position int64   `json:"position"`
	Mode     int64   `json:"mode"`
	Title    string  `json:"title"`
	List     []int64 `json:"list"`
}

// MetaPagination is
type MetaPagination struct {
	List []*Meta `json:"list"`
	Page struct {
		Total int64 `json:"total"`
		Size  int64 `json:"size"`
		Page  int64 `json:"page"`
	} `json:"page"`
}
