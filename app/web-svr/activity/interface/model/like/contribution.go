package like

import xtime "go-common/library/time"

type ActContributions struct {
	ID          int64      `json:"id"`
	Mid         int64      `json:"mid"`
	UpArchives  int64      `json:"up_archives"`
	Likes       int64      `json:"likes"`
	Views       int64      `json:"views"`
	LightVideos int64      `json:"light_videos"`
	Bcuts       int64      `json:"bcuts"`
	Ctime       xtime.Time `json:"ctime"`
}

type ContributionUser struct {
	Mid          int64 `json:"mid"`
	UpArchives   int64 `json:"up_archives"`
	Likes        int64 `json:"likes"`
	Views        int64 `json:"views"`
	LightVideos  int64 `json:"light_videos"`
	Bcuts        int64 `json:"bcuts"`
	SnUpArchives int64 `json:"sn_up_archives"`
	SnLikes      int64 `json:"sn_likes"`
}

type ArchiveInfo struct {
	IsJoin     bool          `json:"is_join"`
	HaveMoney  float64       `json:"have_money"`
	ViewCounts *ViewCounts   `json:"view_counts"`
	Awards     *AwardsFinish `json:"awards"`
}

type ViewCounts struct {
	TargetViews  int64   `json:"target_views"`
	CurrentViews int64   `json:"current_views"`
	ViewPercent  float64 `json:"view_percent"`
	Money        int64   `json:"money"`
}

type AwardsFinish struct {
	BaseFinished bool `json:"base_finished"`
	OneFinished  bool `json:"one_finished"`
	TwoFinished  bool `json:"two_finished"`
	SnFinished   bool `json:"sn_finished"`
}

type LightBcut struct {
	IsJoin bool             `json:"is_join"`
	Lights *LightBcutFinish `json:"lights"`
	Bcuts  *LightBcutFinish `json:"bcuts"`
}

type LightBcutFinish struct {
	MyFinish int64 `json:"my_finish"`
}

type ContriAwards struct {
	ID           int64 `json:"id"`
	AwardType    int64 `json:"award_type"`
	CurrentViews int64 `json:"current_views"`
	UpArchives   int64 `json:"up_archives"`
	Likes        int64 `json:"likes"`
	Views        int64 `json:"views"`
	LightVideos  int64 `json:"light_videos"`
	Bcuts        int64 `json:"bcuts"`
	SplitPeople  int64 `json:"split_people"`
	SplitMoney   int64 `json:"split_money"`
	SnUpArchives int64 `json:"sn_up_archives"`
	SnLikes      int64 `json:"sn_likes"`
}

type TotalRank struct {
	Date string `json:"date"`
	Top  string `json:"top"`
}
