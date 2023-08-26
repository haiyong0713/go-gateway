package model

const (
	SeasonTypeNormal = 0
	SeasonTypeEscape = 1

	SeasonStatusFalse = 1
	SeasonStatusTrue  = 0
)

type SeasonModel struct {
	ID int64 `json:"id"`
	// 赛事id
	Mid int64 `json:"mid"`
	// 简称
	Title string `json:"title"`
	// 全称
	SubTitle string `json:"sub_title"`
	// 开始时间
	Stime int64 `json:"stime"`
	// 结束时间
	Etime int64 `json:"etime"`
	// 主办方
	Sponsor string `json:"sponsor"`
	// logo
	Logo string `json:"logo"`
	// 备注
	Dic string `json:"dic"`
	// 0 启用  1 冻结
	Status int64 `json:"status"`
	// 0 启用  1 冻结
	Rank int64 `json:"rank"`
	// 是否在移动端展示: 0否1是
	IsApp int64 `json:"is_app"`
	// 赛季URL
	URL string `json:"url"`
	// 比赛数据页焦点图
	DataFocus string `json:"data_focus"`
	//比赛数据页焦点图url
	FocusURL string `json:"focus_url"`
	// 禁止类型
	ForbidIndex int64 `json:"forbid_index"`
	// 三方赛季id
	LeidaSid int64 `json:"leida_sid"`
	// 赛季类型：0系列赛，1常规赛
	SerieType int64 `json:"serie_type"`
	// 搜索赛程卡标题底图
	SearchImage string `json:"search_image"`
	// 同步平台
	SyncPlatform int64 `json:"sync_platform"`
	// 竞猜版本
	GuessVersion int64 `json:"guess_version"`
	// 赛季对战类型：0常规对阵，1大逃杀类
	SeasonType int64 `json:"season_type"`
	// 私信通知卡片发送账号uid
	MessageSenduid int64 `json:"message_senduid"`
}
