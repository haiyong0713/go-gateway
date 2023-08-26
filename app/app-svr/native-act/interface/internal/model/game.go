package model

type GaItem struct {
	GameBaseId   int64    `json:"game_base_id"`  // 游戏唯一标识
	GameName     string   `json:"game_name"`     // 游戏名称
	GameIcon     string   `json:"game_icon"`     // 游戏icon
	GameTags     []string `json:"game_tags"`     // 游戏标签
	GameSubtitle string   `json:"game_subtitle"` // 小标题
	GameLink     string   `json:"game_link"`     // 游戏跳转链接
	GameButton   string   `json:"game_button"`   // 按钮文案 2种：预约，进入
	GameStatus   int      `json:"game_status"`   // 游戏状态：0 下载，1 预约（跳过详情），2 预约，3 测试，4 测试+预约，5 跳过详情页，6 仅展示， 7 社区，只有动态小卡行动点游戏信息接口会返回
}
