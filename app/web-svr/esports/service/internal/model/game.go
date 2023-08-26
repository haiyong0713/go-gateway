package model

type GameModel struct {
	ID int64 `json:"id"`
	// 赛事简称
	Title string `json:"title"`
	// 赛事全称
	SubTitle string `json:"sub_title"`
	// 英文全名
	ETitle string `json:"e_title"`
	// 平台
	Plat int64 `json:"plat"`
	// 游戏类型
	Type int64 `json:"type"`
	// logo
	Logo string `json:"logo"`
	// 发行商
	Publisher string `json:"publisher"`
	// 运行商
	Operations string `json:"operations"`
	// 运行商
	PbTime int64 `json:"pb_time"`
	// 备注
	Dic string `json:"dic"`
	// 状态
	Status int64 `json:"status"`
	// 排序
	Rank int64 `json:"rank"`
}
