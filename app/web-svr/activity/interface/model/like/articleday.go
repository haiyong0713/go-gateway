package like

import xtime "go-common/library/time"

// ArticleDay .
type ArticleDay struct {
	ID           int64      `json:"id"`
	Mid          int64      `json:"mid"`
	Publish      string     `json:"publish"`
	PublishCount int64      `json:"publish_count"`
	FinishTask   int64      `json:"finish_task"`
	FinishTime   int64      `json:"finish_time"`
	Ctime        xtime.Time `json:"ctime"`
	Status       int64      `json:"status"`
}

// ArticleDayInfo .
type ArticleDayInfo struct {
	IsJoin      bool         `json:"is_join"`
	HaveMoney   float64      `json:"have_money"`
	Ctime       xtime.Time   `json:"ctime"`
	ActivityDay *ActivityDay `json:"activity_day"`
	ClockIn     []string     `json:"clock_in"`
	RightInfo   *RightInfo   `json:"right_info"`
	Status      int64        `json:"status"`
}

// ActivityDay .
type ActivityDay struct {
	ApplyTime  int64 `json:"apply_time"`
	BeginTime  int64 `json:"begin_time"`
	EndTime    int64 `json:"end_time"`
	ResultTime int64 `json:"result_time"`
}

// RightInfo .
type RightInfo struct {
	DaysLater       int64 `json:"days_later"`
	YesterdayPeople int64 `json:"yesterday_people"`
	BeforePeople    int64 `json:"before_people"`
	BeforePublish   int64 `json:"before_publish"`
}

// ArticleDayAward .
type ArticleDayAward struct {
	ID           int64  `json:"id"`
	ActivityUID  string `json:"activity_uid"`
	ConditionMin int64  `json:"condition_min"`
	ConditionMax int64  `json:"condition_max"`
	SplitPeople  int64  `json:"split_people"`
	SplitMoney   int64  `json:"split_money"`
}
