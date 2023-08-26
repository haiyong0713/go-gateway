package model

const (
	// 赛程类型
	DefaultContest = 0
	ContestSpecial = 1

	// 比赛状态
	ContestStatusInit    = int64(0)
	ContestStatusWaiting = int64(1)
	ContestStatusIng     = int64(2)
	ContestStatusOver    = int64(3)

	// 人群包
	BGroupExits = 145202

	// 小卡事件
	TunnelV2EventAlready  = 108009
	TunnelV2NotExists     = 108019
	TunnelV2CardStatusErr = 108014

	ContestGuessHasTrue  = int64(1)
	ContestGuessHasFalse = int64(0)
)

type ContestModel struct {
	// 赛程id
	ID int64 `json:"id"`
	// 比赛阶段
	GameStage string `json:"game_stage"`
	// 比赛开始时间
	Stime int64 `json:"stime"`
	// 比赛结束时间
	Etime int64 `json:"etime"`
	// 主场队伍id
	HomeID int64 `json:"home_id"`
	// 客场队伍id
	AwayID int64 `json:"away_id"`
	// 主场分数
	HomeScore int64 `json:"home_score"`
	// 客场分数
	AwayScore int64 `json:"away_score"`
	// 直播房间号
	LiveRoom int64 `json:"live_room"`
	// 回播房间号
	Aid int64 `json:"aid"`
	// 集锦房间号
	Collection int64 `json:"collection"`
	// 赛程描述， 但是不清楚为啥之前描述用dic命名？
	Dic string `json:"dic"`
	// 0 启用 1 冻结
	Status int64 `json:"status"`
	// 季度id
	Sid int64 `json:"sid"`
	// 赛事id
	Mid int64 `json:"mid"`
	// 赛程类型：0普通1特殊
	Special int64 `json:"special"`
	// 胜利战队
	SuccessTeam int64 `json:"success_team"`
	// 赛程名称
	SpecialName string `json:"special_name"`
	// 胜利文案
	SpecialTips string `json:"special_tips"`
	// 赛程图片
	SpecialImage string `json:"special_image"`
	// 回播房间号url
	Playback string `json:"playback"`
	// 集锦房间号url
	CollectionURL string `json:"collection_url"`
	// 集锦房间号url
	LiveURL string `json:"live_url"`
	// 比赛数据页类型 0：无 1：LOL 2:DATA2
	DataType int64 `json:"data_type"`
	// 雷达数据match_id
	MatchID int64 `json:"match_id"`
	// 是否有竞猜
	GuessType int64 `json:"guess_type"`
	// 比赛阶段1
	GameStage1 string `json:"game_stage1"`
	// 比赛阶段2
	GameStage2 string `json:"game_stage2"`
	// 阶段id
	SeriesId   int64 `json:"series_id"`
	PushSwitch int64 `json:"push_switch"`
	ActivePush int64 `json:"active_push"`
	// 比赛状态，枚举：1未开始，2进行中，3已结束
	ContestStatus int64 `json:"contest_status"`
	// 三方赛程id
	ExternalID int64 `json:"external_id"`
	GameId     int64 `json:"game_id" gorm:"-"`
}

type ContestUpdateModel struct {
	// 赛程id
	ID int64 `json:"id"`
	// 比赛阶段
	GameStage string `json:"game_stage"`
	// 比赛开始时间
	Stime int64 `json:"stime"`
	// 比赛结束时间
	Etime int64 `json:"etime"`
	// 主场队伍id
	HomeID int64 `json:"home_id"`
	// 客场队伍id
	AwayID int64 `json:"away_id"`
	// 主场分数
	HomeScore int64 `json:"home_score"`
	// 客场分数
	AwayScore int64 `json:"away_score"`
	// 直播房间号
	LiveRoom int64 `json:"live_room"`
	// 回播房间号
	Aid int64 `json:"aid"`
	// 集锦房间号
	Collection int64 `json:"collection"`
	// 赛程描述， 但是不清楚为啥之前描述用dic命名？
	Dic string `json:"dic"`
	// 0 启用 1 冻结
	Status int64 `json:"status"`
	// 季度id
	Sid int64 `json:"sid"`
	// 赛事id
	Mid int64 `json:"mid"`
	// 赛程类型：0普通1特殊
	Special int64 `json:"special"`
	// 胜利战队
	SuccessTeam int64 `json:"success_team"`
	// 赛程名称
	SpecialName string `json:"special_name"`
	// 胜利文案
	SpecialTips string `json:"special_tips"`
	// 赛程图片
	SpecialImage string `json:"special_image"`
	// 回播房间号url
	Playback string `json:"playback"`
	// 集锦房间号url
	CollectionURL string `json:"collection_url"`
	// 集锦房间号url
	LiveURL string `json:"live_url"`
	// 比赛数据页类型 0：无 1：LOL 2:DATA2
	DataType int64 `json:"data_type"`
	// 雷达数据match_id
	MatchID int64 `json:"match_id"`
	// 比赛阶段1
	GameStage1 string `json:"game_stage1"`
	// 比赛阶段2
	GameStage2 string `json:"game_stage2"`
	// 阶段id
	SeriesId   int64 `json:"series_id"`
	PushSwitch int64 `json:"push_switch"`
	// 比赛状态，枚举：1未开始，2进行中，3已结束
	ContestStatus int64 `json:"contest_status"`
}

type AutoSubscribeDetail struct {
	SeasonID  int64 `json:"season_id"`
	TeamId    int64 `json:"team_id"`
	ContestID int64 `json:"contest_id"`
}
