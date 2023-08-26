package like

import "go-gateway/app/web-svr/activity/interface/model/lottery"

// Star .
type Star struct {
	JoinState    int   `json:"join_state"`
	DownState    int   `json:"down_state"`
	ArchiveCount int64 `json:"sub_count"`
	ArchiveStat  int64 `json:"archive_stat"`
}

type StarSpring struct {
	ArcCount   int64 `json:"arc_count"`
	OtherCount int64 `json:"other_count"`
}

type ReadDay struct {
	TotalCount int64 `json:"total_count"`
	Max        int64 `json:"max"`
	EndTime    int64 `json:"end_time"`
	MyCount    int64 `json:"my_count"`
}

type ImgUserRank struct {
	Self *ImageSelf `json:"self"`
	List []*ImgUser `json:"list"`
}

type ImageSelf struct {
	*SimpleUser
	DayRank    int64   `json:"day_rank"`
	DayScore   float64 `json:"day_score"`
	TotalRank  int64   `json:"total_rank"`
	TotalScore float64 `json:"total_score"`
}

type ImgUser struct {
	*SimpleUser
	ImageRank  int64   `json:"image_rank"`
	ImageScore float64 `json:"image_score"`
	Follower   int64   `json:"follower"`
}

type SimpleUser struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type StupidListReply struct {
	Global     *StupidGolbal     `json:"global"`
	Individual *StupidIndividual `json:"individual"`
}

type StupidGolbal struct {
	Total   int64 `json:"total"`
	Target1 int64 `json:"target_1"`
	Target2 int64 `json:"target_2"`
	Target3 int64 `json:"target_3"`
}

type StupidIndividual struct {
	Target1 bool `json:"target_1"`
	Target2 bool `json:"target_2"`
	Target3 bool `json:"target_3"`
}

type StupidVv struct {
	Mid int64 `json:"mid"`
	Vv  int64 `json:"vv"`
}

type StupidStatus struct {
	IsAfrican        bool                          `json:"is_african"`
	Lottery          *lottery.Lottery              `json:"lottery"`
	LotteryTimesConf []*lottery.LotteryTimesConfig `json:"lottery_times_conf"`
}
