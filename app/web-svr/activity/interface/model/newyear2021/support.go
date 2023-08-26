package newyear2021

type UserSupportInfo struct {
	Mid            int64       `json:"mid"`
	IsVip          bool        `json:"is_vip"`
	LevelTasks     interface{} `json:"level_tasks"`
	DailyTasks     interface{} `json:"daily_tasks"`
	ReceiveRecords interface{} `json:"receive_records"`
}
