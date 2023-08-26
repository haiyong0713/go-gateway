package resource

type EntryMsg struct {
	// 时间设定id
	ID int64 `json:"id"`
	// 状态id
	StateID int64 `json:"state_id"`
	// 入口名称
	EntryName string `json:"entry_name"`
	// 状态名称
	StateName string `json:"state_name"`
	// 静态icon
	StaticIcon string `json:"static_icon"`
	// 动态icon
	DynamicIcon string `json:"dynamic_icon"`
	// 跳转url
	Url string `json:"url"`
	// 动画循环次数
	LoopCnt int32 `json:"loop_count"`
	// 状态起效时间
	STime int64 `json:"stime"`
	// 入口配置结束时间
	ETime int64 `json:"etime"`
	// 平台信息
	Plat []*PlatLimit `json:"platforms"`
}

type PlatLimit struct {
	Plat       int32  `json:"plat"`
	Conditions string `json:"conditions"`
	Build      int32  `json:"build"`
}
