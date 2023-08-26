package handwrite

const (
	// HandWriteKey handwrite key
	HandWriteKey = "handWrite"
	// TaskTypeGod 神仙模式
	TaskTypeGod = 1
	// TaskTypeTiredLevel1 佛系模式
	TaskTypeTiredLevel1 = 2
	// TaskTypeTiredLevel2 爆肝模式1
	TaskTypeTiredLevel2 = 3
	// TaskTypeTiredLevel3 爆肝模式2
	TaskTypeTiredLevel3 = 4
)

// Mid ...
type Mid struct {
	Mid int64 `json:"mid"`
}

// MidAward 用户获奖情况
type MidAward struct {
	Mid   int64 `json:"mid"`
	God   int   `json:"god"`
	Tired int   `json:"tired"`
	New   int   `json:"new"`
	Score int64 `json:"score"`
	Rank  int   `json:"rank"`
}

// AwardCount 用户获奖人数统计
type AwardCount struct {
	God   int `json:"god"`
	Tired int `json:"tired"`
	New   int `json:"new"`
}

// Remark 备注信息
type Remark struct {
	Follower int64 `json:"follower"`
}

// AwardCountNew 任务完成情况统计
type AwardCountNew struct {
	God         int64 `json:"god"`
	TiredLevel1 int64 `json:"tired_level_1"`
	TiredLevel2 int64 `json:"tired_level_2"`
	TiredLevel3 int64 `json:"tired_level_3"`
}

// MidTask 用户任务情况
type MidTask struct {
	Mid         int64   `json:"mid"`
	TaskType    int     `json:"task_type"`
	FinishCount int     `json:"finish_count"`
	TaskDetail  []int64 `json:"task_detail"`
}

// MidTaskAll 用户任务完成情况
type MidTaskAll struct {
	Mid         int64 `json:"mid"`
	God         int   `json:"god"`
	TiredLevel1 int   `json:"tired_level_1"`
	TiredLevel2 int   `json:"tired_level_2"`
	TiredLevel3 int   `json:"tired_level_3"`
}

// MidTaskDB 用户任务情况db
type MidTaskDB struct {
	Mid              int64   `json:"mid"`
	TaskType         int     `json:"task_type"`
	FinishCount      int     `json:"finish_count"`
	TaskDetail       string  `json:"task_detail"`
	TaskDetailStruct []int64 `json:"task_detail_struct"`
	FinishTime       int64   `json:"finish_time"`
}

// MidTaskAllData 用户任务完成情况
type MidTaskAllData struct {
	Mid               int64  `json:"mid"`
	God               int    `json:"god"`
	GodTime           int64  `json:"god_time"`
	GodDetail         string `json:"god_detail"`
	TiredLevel1       int    `json:"tired_level_1"`
	TiredLevel1Detail string `json:"tired_level_1_detail"`
	TiredLevel1Time   int64  `json:"tired_level_1_time"`
	TiredLevel2       int    `json:"tired_level_2"`
	TiredLevel2Detail string `json:"tired_level_2_detail"`
	TiredLevel2Time   int64  `json:"tired_level_2_time"`
	TiredLevel3       int    `json:"tired_level_3"`
	TiredLevel3Time   int64  `json:"tired_level_3_time"`
	TiredLevel3Detail string `json:"tired_level_3_detail"`
	Money             int64  `json:"money"`
	NickName          string `json:"nickname"`
	Fans              int64  `json:"fans"`
}
