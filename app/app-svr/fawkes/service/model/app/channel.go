package app

// channel constant
const (
	ChannelStatic = 1 // 渠道类型：静态
	ChannelCustom = 2 // 渠道类型：自定义
	ChannelNormal = 0 // 渠道状态：正常
	ChannelDelete = 1 // 渠道状态：已删除
)

// Channel Struct
type Channel struct {
	AppKey   string            `json:"app_key,omitempty"`
	ID       int64             `json:"id"`
	AID      int64             `json:"aid"`
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Plate    string            `json:"plate"`
	Status   int8              `json:"status"`
	Operator string            `json:"operator"`
	Group    *ChannelGroupInfo `json:"group,omitempty"`
	Ctime    int64             `json:"ctime"`
	Mtime    int64             `json:"mtime"`
}

// ChannelGroupInfo Struct
type ChannelGroupInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AutoPushCdn int8   `json:"auto_push_cdn"` //是否自动推送CDN
	IsAutoGen   int8   `json:"is_auto_gen"`   //是否自动生成
	Priority    int8   `json:"priority"`      //优先级
	QaOwner     string `json:"qa_owner"`      //测试负责人
	MarketOwner string `json:"market_owner"`  //市场负责人
}

// ChannelGroup Struct
type ChannelGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
	AutoPushCdn int8   `json:"auto_push_cdn"` //是否自动推送CDN
	IsAutoGen   int8   `json:"is_auto_gen"`   //是否自动生成
	Priority    int8   `json:"priority"`      //优先级 (0-100)优先级递增
	QaOwner     string `json:"qa_owner"`      //测试负责人
	MarketOwner string `json:"market_owner"`  //市场负责人
	Ctime       int64  `json:"ctime"`
	Mtime       int64  `json:"mtime"`
}

// GroupChannels Struct
type GroupChannels struct {
	Group      *ChannelGroupInfo  `json:"group"`
	ChannelMap map[int64]*Channel `json:"channel_map"`
	Channels   []*Channel         `json:"channels"`
}
