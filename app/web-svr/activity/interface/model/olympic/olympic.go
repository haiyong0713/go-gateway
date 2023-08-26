package olympic

type OlympicContest struct {
	Id            int64  `json:"id"`
	GameStage     string `json:"game_stage"`
	Stime         string `json:"stime"`
	HomeTeamName  string `json:"home_team_name"`
	AwayTeamName  string `json:"away_team_name"`
	HomeTeamUrl   string `json:"home_team_url"`
	AwayTeamUrl   string `json:"away_team_url"`
	HomeScore     string `json:"home_score"`
	AwayScore     string `json:"away_score"`
	ContestStatus string `json:"contest_status"`
	SeasonTitle   string `json:"season_title"`
	SeasonUrl     string `json:"season_url"`
	VideoUrl      string `json:"video_url"`
	BottomUrl     string `json:"bottom_url"`
	ShowRule      string `json:"show_rule"`
}

type OlympicQueryConfig struct {
	QueryWord string `json:"query_word"`
	MatchId   int64  `json:"match_id"`
	State     int64  `json:"state"`
}

type OlympicDBData struct {
	Id    int64  `json:"id"`
	Data  string `json:"data"`
	State int64  `json:"state"`
}
