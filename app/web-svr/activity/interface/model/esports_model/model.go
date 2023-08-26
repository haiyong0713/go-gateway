package esports_model

type ContestCard struct {
	Contest Contest4Frontend `json:"contest"`
	More    []*ContestMore   `json:"more"`
}

type ContestMore struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	OnClick string `json:"on_click"`
}

type Contest4Frontend struct {
	ID        int64          `json:"id"`
	StartTime int64          `json:"start_time"`
	EndTime   int64          `json:"end_time"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Home      Team4Frontend  `json:"home"`
	Away      Team4Frontend  `json:"away"`
	Series    *ContestSeries `json:"series"`
	SeriesID  int64          `json:"series_id"`
}

type Team4Frontend struct {
	ID       int64  `json:"id"`
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Wins     int64  `json:"wins"`
	Region   string `json:"region"`
	RegionID int    `json:"region_id"`
}

type ContestSeries struct {
	ID          int64  `json:"id"`
	ParentTitle string `json:"parent_title" validate:"required"`
	ChildTitle  string `json:"child_title" validate:"required"`
	StartTime   int64  `json:"start_time" validate:"min=1"`
	EndTime     int64  `json:"end_time" validate:"min=1"`
	ScoreID     string `json:"score_id" validate:"required"`
}
