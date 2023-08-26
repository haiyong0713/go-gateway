package rewards

type AddActivityParam struct {
	//活动名称
	Name string `form:"name" validate:"required"`
	//发奖通知发送人
	NotifySenderId int64 `form:"notify_sender_id" validate:"min=1,required"`
	//卡片通知码
	NotifyCode string `form:"notify_code" validate:"required"`
	//发奖通知内容
	NotifyMessage string `form:"notify_message"`
	//发奖跳转链接
	NotifyJumpUrl string `form:"notify_jump_url"`
	//发奖跳转链接2
	NotifyJumpUrl2 string `form:"notify_jump_url2"`
}

type UpdateActivityParam struct {
	//活动ID
	Id int64 `form:"id" validate:"min=1,required"`
	//活动名称
	Name string `form:"name" validate:"required"`
	//发奖通知发送人
	NotifySenderId int64 `form:"notify_sender_id" validate:"min=1,required"`
	//卡片通知码
	NotifyCode string `form:"notify_code" validate:"required"`
	//发奖通知内容
	NotifyMessage string `form:"notify_message"`
	//发奖跳转链接
	NotifyJumpUrl string `form:"notify_jump_url"`
	//发奖跳转链接
	NotifyJumpUrl2 string `form:"notify_jump_url2"`
}

// 活动整体配置
type ActivityConfig struct {
	//活动ID
	Id int64
	//活动名称
	Name string
	//发奖通知发送人
	NotifySenderId int64
	//卡片通知码
	NotifyCode string
	//发奖通知内容
	NotifyMessage string
	//发奖跳转链接
	NotifyJumpUrl string
	//发奖跳转链接
	NotifyJumpUrl2 string
	//奖品列表
	Awards []*GenericConfig
}

// 活动列表展示
type ActivityListInfo struct {
	//活动ID
	Id int64
	//活动名称
	Name string
	//发奖通知发送人
	NotifySenderId int64
	//卡片通知码
	NotifyCode string
	//发奖通知内容
	NotifyMessage string
	//发奖跳转链接1
	NotifyJumpUri1 string
	//发奖跳转链接2
	NotifyJumpUri2 string
	//奖品个数
	AwardsCount int64
}

type Activity struct {
	//活动ID
	Id int64
	//活动名称
	Name string
	//发奖通知发送人
	NotifySenderId int64
	//卡片通知码
	NotifyCode string
	//发奖通知内容
	NotifyMessage string
	//发奖跳转链接
	NotifyJumpUrl string
}
