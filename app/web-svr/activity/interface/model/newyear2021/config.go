package newyear2021

import (
	"go-common/library/time"
)

// Config 2020拜年祭
type Config struct {
	//直播间旧UP主祝福
	LiveOldGreetingAwardIds []int64
	//直播间新UP主祝福
	LiveNewUpGreetingAwardIds []int64
	//拜年祭活动平台活动名
	ActPlatActId              string
	ActPlatGameCounterName    string
	ActPlatMallCounterName    string
	ActPlatLotteryCounterName string
	//扭蛋
	//抽奖平台ID
	NiuDanLotterySid string //扭蛋抽奖ID
	//发奖平台ID
	NiuDanRewardsActivityId int64
	//Up主祝福奖励的ID
	NiuDanUpGreetingAwardIds []int64

	//直播间抽奖(拓展1)
	LiveLotterySid1        string
	LiveRewardsActivityId1 int64
	UpGreetingAwardIds1    []int64

	//直播间抽奖(拓展2)
	LiveLotterySid2        string
	LiveRewardsActivityId2 int64
	UpGreetingAwardIds2    []int64

	//直播间抽奖(拓展3)
	LiveLotterySid3        string
	LiveRewardsActivityId3 int64
	UpGreetingAwardIds3    []int64

	TimePeriod *Period
	TaskConfig *TaskConfig
}

type TaskConfig struct {
	//每日任务中的Ogv跳转片单地址, 按天分配
	DailyTaskOgvSeasons map[string]string
	//Ogv兜底片单
	DailyTaskOgvDefaultSeason string
	//每日任务列表
	DailyTasks *DailyTask

	//战令系统
	LevelTask *StageTask
}

type Task struct {
	//任务名: 会添加到API的返回值中.无其他作用
	DisplayName string

	//任务图标: 后端不使用, 前端展示使用
	DisplayIcon string

	//任务ID: 用于唯一区分一个任务
	Id int64

	//是否是VIP隐藏任务
	VipHidden bool

	//关联的VIP装扮ID
	VipSuitID int64

	//任务对应活动ID
	ActPlatId string

	//任务对应活动Counter
	ActPlatCounterId string

	//任务条件: Counter大于RequireCount即视为任务完成
	RequireCount int64

	//任务奖励类型: 需要预先在rewards服务中配置
	AwardId int64

	//完成后不再显示此任务
	HideOnFinish bool
	//PC跳转地址
	PcUrl string
	//H5跳转地址
	H5Url string
}

type StageTask struct {
	//任务名: 会添加到API的返回值中.无其他作用
	DisplayName string

	//任务ID: 用于唯一区分一个任务
	Id int64

	//任务对应活动ID
	ActPlatId string

	//任务对应活动Counter
	ActPlatCounterId string

	//阶段配置
	Stages []*Task
}

type DailyTask struct {
	//每日任务列表
	Tasks []*Task
}

type Period struct {
	Start time.Time
	End   time.Time
}

func (task *Task) DeepCopy() (newOne *Task) {
	newOne = new(Task)
	{
		newOne.DisplayName = task.DisplayName
		newOne.DisplayIcon = task.DisplayIcon
		newOne.Id = task.Id
		newOne.VipHidden = task.VipHidden
		newOne.VipSuitID = task.VipSuitID
		newOne.ActPlatId = task.ActPlatId
		newOne.ActPlatCounterId = task.ActPlatCounterId
		newOne.RequireCount = task.RequireCount
		newOne.AwardId = task.AwardId
		newOne.HideOnFinish = task.HideOnFinish
		newOne.PcUrl = task.PcUrl
		newOne.H5Url = task.H5Url
	}

	return
}

func DeepCopTaskList(oldOne []*Task) (newOne []*Task) {
	newOne = make([]*Task, 0)
	if oldOne == nil || len(oldOne) == 0 {
		return
	}

	for _, v := range oldOne {
		newOne = append(newOne, v.DeepCopy())
	}

	return
}
