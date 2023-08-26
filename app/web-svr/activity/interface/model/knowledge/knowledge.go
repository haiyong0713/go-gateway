package knowledge

// KnowConfig .
type KnowConfig struct {
	ID            int64             `json:"id"`
	ConfigDetails *KnowConfigDetail `json:"config_details"`
}

type KnowTask struct {
	TaskName   string  `json:"task_name"`
	TaskColumn string  `json:"task_column"`
	TaskFinish int64   `json:"task_finish"`
	Priority   float64 `json:"priority"`
	Image      string  `json:"image"`
	ImageGrey  string  `json:"image_grey"`
	IsFinish   bool    `json:"is_finish"`
}

type LevelTask struct {
	Priority          float64     `json:"priority"`
	ParentName        string      `json:"parent_name"`
	ParentDescription string      `json:"parent_description"`
	Tasks             []*KnowTask `json:"tasks"`
}

type KnowConfigDetail struct {
	Name          string                `json:"name"`
	Table         string                `json:"table"`
	TableCount    int64                 `json:"table_count"`
	TotalBadge    float64               `json:"total_badge"`
	ShareFields   string                `json:"share_fields"`
	TotalProgress float64               `json:"total_progress"`
	LevelTask     map[string]*LevelTask `json:"level_task"`
}

// UserInfo 用户看板
type UserInfo struct {
	MID          int64 `json:"mid"`
	ArchiveCount int64 `json:"archive_count"`
	SingleView   int64 `json:"single_view"`
	AllView      int64 `json:"all_view"`
}

// UserInfoRes 用户看板
type UserInfoRes struct {
	ArchiveCount int64    `json:"archive_count"`
	SingleView   int64    `json:"single_view"`
	AllView      int64    `json:"all_view"`
	Live         int64    `json:"live_days"`
	ActivityEnd  int64    `json:"activity_end"`
	Account      *Account `json:"account"`
}

// Account ...
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// UserKnowTask .
type UserKnowTask struct {
	ID             int64 `json:"id"`
	Mid            int64 `json:"mid"`
	HadArc         int64 `json:"had_arc"`
	Coin           int64 `json:"coin"`
	Favorite       int64 `json:"favorite"`
	Share          int64 `json:"share"`
	Year2020Share  int64 `json:"year_2020_share"`
	Year2021Share  int64 `json:"year_2021_share"`
	See2020Share   int64 `json:"see_2020_share"`
	See2021Share   int64 `json:"see_2021_share"`
	Super2020Share int64 `json:"super_2020_share"`
	Super2021Share int64 `json:"super_2021_share"`
	Gold2020Share  int64 `json:"gold_2020_share"`
	Gold2021Share  int64 `json:"gold_2021_share"`
	Dark2020Share  int64 `json:"dark_2020_share"`
	Dark2021Share  int64 `json:"dark_2021_share"`
}

// Period ...
type Period struct {
	PeriodList []*PeriodList `json:"periodList"`
}

// PeriodList ...
type PeriodList struct {
	WeekList  []*WeekList `json:"weekList"`
	SuperList []string    `json:"superList"`
}

type WeekList struct {
	GoldList  []string `json:"goldList"`
	HorseList []string `json:"horseList"`
}
