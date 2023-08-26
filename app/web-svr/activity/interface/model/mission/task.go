package mission

import v1 "go-gateway/app/web-svr/activity/interface/api"

const (
	TaskRewardStatusInit    = 0 // 初始化记录状态，非法
	TaskRewardStatusIdle    = 1 // 待领取状态
	TaskRewardStatusIn      = 2 // 领取中
	TaskRewardStatusSuccess = 3 //领取成功
	TaskRewardStatusFailed  = 4 // 领取失败

	TaskPeriodNone    = 1 // 无周期，即活动内一次
	TaskPeriodDaily   = 2 // 每天一次
	TaskPeriodWeekly  = 3 // 每周一次
	TaskPeriodMonthly = 4 // 每月一次
)

type TaskGroupsMapping struct {
	ID            int64 `json:"id"`
	ActId         int64 `json:"act_id"`
	TaskId        int64 `json:"task_id"`
	GroupId       int64 `json:"group_id"`
	CompleteScore int64 `json:"complete_score"`
}

type UserCompleteRecord struct {
	ID                int64  `json:"id"`
	ActId             int64  `json:"act_id"`
	TaskId            int64  `json:"task_id"`
	Mid               int64  `json:"mid"`
	CompletePeriod    int64  `json:"complete_period"`
	CompleteTime      int64  `json:"complete_time"`
	FailureTime       int64  `json:"failure_time"`
	TaskRewardsStatus int64  `json:"task_rewards_status"`
	SerialNum         string `json:"serial_num"`
	Reason            string `json:"reason"`
	ReceivePeriod     int64  `json:"receive_period"`
}

type ActivityTaskStatInfo struct {
	TaskDetail *v1.MissionTaskDetail
	StockStat  *TaskStockStat
	PeriodStat *TaskPeriodStat
}

type TaskStockStat struct {
	CycleType      int64
	LimitType      int64
	Total          int64
	Consumed       int64
	StockBeginTime int64
	StockEndTime   int64
	StockPeriod    int64
}

type TaskPeriodStat struct {
	Period          int64
	PeriodBeginTime int64
	PeriodEndTime   int64
}

type UserTaskDetail struct {
	ID                      int64            `json:"id"`
	ActId                   int64            `json:"act_id"`
	GroupList               []*GroupSchedule `json:"group_list"`
	TaskPeriod              int64            `json:"task_period"`
	RewardId                int64            `json:"reward_id"`
	RewardType              int64            `json:"reward_type"`
	RewardInfo              *RewardInfo      `json:"reward_info"`
	RewardPeriodPoolNum     int64            `json:"reward_period_pool_num"`
	RewardPeriodReceivedNum int64            `json:"reward_period_received_num"`
	RewardPeriodStockNum    int64            `json:"reward_period_stock_num"`
	RewardReceiveBeginTime  int64            `json:"reward_receive_begin_time"`
	RewardReceiveEndTime    int64            `json:"reward_receive_end_time"`
	ReceiveId               int64            `json:"receive_id"`
	ReceiveStatus           int64            `json:"receive_status"`
	ReceivePeriod           int64            `json:"receive_period"`
}

type GroupSchedule struct {
	TaskId           int64 `json:"task_id"`
	ActId            int64 `json:"act_id"`
	GroupId          int64 `json:"group_id"`
	GroupBaseNum     int64 `json:"group_base_num"`
	GroupCompleteNum int64 `json:"group_complete_num"`
}

type RewardInfo struct {
	RewardId    int64        `json:"reward_id"`
	RewardName  string       `json:"reward_name"`
	RewardIcon  string       `json:"reward_icon"`
	RewardActId int64        `json:"reward_act_id"`
	Type        string       `json:"type"`
	Awards      []*AwardInfo `json:"awards"`
}

type AwardInfo struct {
	AwardId   int64  `json:"award_id"`
	AwardName string `json:"award_name"`
	AwardIcon string `json:"award_icon"`
	AwardType string `json:"award_type"`
}
