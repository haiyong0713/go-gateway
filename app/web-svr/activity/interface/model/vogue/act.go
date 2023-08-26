package model

type State struct {
	UserInfo       *UserInfo       `json:"user_info,omitempty"`
	User           *User           `json:"user,omitempty"`
	Duration       *Duration       `json:"duration,omitempty"`
	DoubleDuration *Duration       `json:"double_duration,omitempty"`
	ScoreList      *ScoreList      `json:"score_list,omitempty"`
	PlayList       []*PlayListItem `json:"play_list,omitempty"`
}

type UserInfo struct {
	Name    string `json:"name"`    // 用户名称
	Picture string `json:"picture"` // 用户头像
}

type User struct {
	Score      int64      `json:"score"`               // 用户积分
	Today      int64      `json:"today"`               // 当天获得分数
	TodayLimit int64      `json:"today_limit"`         // 当天获得分数上限
	Token      string     `json:"token"`               // 用户token
	Selection  *Selection `json:"selection,omitempty"` // 选择奖品
}

type Selection struct {
	Name       string `json:"name"`        // 名称
	Type       string `json:"type"`        // real virtual
	Cost       int64  `json:"cost"`        // 总花费
	Picture    string `json:"picture"`     // 商品图片
	Status     int64  `json:"status"`      // 状态 1进行中 2填写地址 3已兑换 4补货中 5缺货
	Stock      int64  `json:"stock"`       // 库存
	TotalStock int64  `json:"total_stock"` // 总库存
}

type Duration struct {
	Start int64 `json:"start"` // 活动开始时间
	End   int64 `json:"end"`   // 活动结束时间
}
type ScoreList struct {
	MaxScore       int64   `json:"max_score"`
	FirstBonusTime []int64 `json:"first_bonus_time"`
	InviteScore    int64   `json:"invite_score"`
	TodayLimit     int64   `json:"today_limit"`
}

type ScoreItem struct {
	Num  int64 `json:"num"`
	Min  int64 `json:"min"`
	Max  int64 `json:"max"`
	Show bool  `json:"show"`
}

type PlayListItem struct {
	Id   int64  `json:"id"`   // 播单id
	Name string `json:"name"` // 播单名称
}
