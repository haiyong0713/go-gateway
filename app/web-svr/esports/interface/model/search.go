package model

// Page es page
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// SearchVideo search video.
type SearchVideo struct {
	AID int64 `json:"aid"`
}

// SearchEsp big search esports.
type SearchEsp struct {
	Code       int                `json:"code,omitempty"`
	Seid       string             `json:"seid"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pagesize"`
	NumResults int                `json:"numResults"`
	NumPages   int                `json:"numPages"`
	Result     []*SearchEspResult `json:"result"`
}

type SearchEspResult struct {
	Play       int      `json:"play"`
	Pubdate    int      `json:"pubdate"`
	Title      string   `json:"title"`
	Cover      string   `json:"cover"`
	Bvid       string   `json:"bvid"`
	RankOffset int      `json:"rank_offset"`
	HitColumns []string `json:"hit_columns"`
	Mid        int      `json:"mid"`
	DmCount    int      `json:"dm_count"`
	Uname      string   `json:"uname"`
	Duration   int      `json:"duration"`
	RankIndex  int      `json:"rank_index"`
	Type       string   `json:"type"`
	ID         int64    `json:"id"`
	RankScore  int      `json:"rank_score"`
}

// FilterES  filter ES video and match
type FilterES struct {
	GroupByGid []struct {
		DocCount int    `json:"doc_count"`
		Key      string `json:"key"`
	} `json:"group_by_gid"`
	GroupByMatch []struct {
		DocCount int    `json:"doc_count"`
		Key      string `json:"key"`
	} `json:"group_by_match"`
	GroupByTag []struct {
		DocCount int    `json:"doc_count"`
		Key      string `json:"key"`
	} `json:"group_by_tag"`
	GroupByTeam []struct {
		DocCount int    `json:"doc_count"`
		Key      string `json:"key"`
	} `json:"group_by_team"`
	GroupByYear []struct {
		DocCount int    `json:"doc_count"`
		Key      string `json:"key"`
	} `json:"group_by_year"`
}

// EsResult .
type EsResult struct {
	Code int `json:"code"`
	Data *struct {
		Order   string     `json:"order"`
		Sort    string     `json:"sort"`
		Page    Page       `json:"page"`
		Result  []*Contest `json:"result"`
		Message string     `json:"message"`
	}
}
