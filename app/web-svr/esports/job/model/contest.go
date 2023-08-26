package model

import v1 "go-gateway/app/web-svr/esports/service/api/v1"

type ContestSchedule struct {
	ID int64 `json:"id"`
	//比赛阶段文本信息
	GameStage string `json:"game_stage"`
	//比赛开始时间
	Stime int64 `json:"stime"`
	//比赛结束时间
	Etime int64 `json:"etime"`
	//主场队伍id
	HomeID int64 `json:"home_id"`
	//客场队伍id
	AwayID int64 `json:"away_id"`
	//主场分数
	HomeScore int64 `json:"home_score"`
	//客场分数
	AwayScore int64 `json:"away_score"`
	// 游戏id
	GameId int64 `json:"game_id"`
	// 游戏详情
	Game *v1.GameDetail `json:"game"`
	//赛季id
	Sid int64 `json:"sid"`
	//赛季
	Season *v1.SeasonDetail `json:"season"`
	//赛事id
	Mid int64 `json:"mid"`
	// 赛程关联的阶段id
	SeriesId int64 `json:"series_id"`
	//直播房间号
	LiveRoom int64 `json:"live_room"`
	//回播房间号
	Aid int64 `json:"aid"`
	//集锦房间号
	Collection int64 `json:"collection"`
	// 赛程描述
	Dic string `json:"dic"`
	//赛程类型：0普通1特殊
	Special int64 `json:"special"`
	//特殊赛程的胜利战队, 无主客队仅有胜利队伍
	SuccessTeam int64 `json:"success_team"`
	//特殊赛程，赛程名称
	SpecialName string `json:"special_name"`
	//特殊赛程，胜利文案
	SpecialTips string `json:"special_tips"`
	//特殊赛程，赛程图片
	SpecialImage string `json:"special_image"`
	//回播房间号url
	Playback string `json:"playback"`
	//集锦房间号url
	CollectionURL string `json:"collection_url"`
	//集锦房间号url
	LiveURL string `json:"live_url"`
	//比赛数据页类型 0：无 1：LOL 2:DATA2
	DataType int64 `json:"data_type"`
	// 赛程的冻结状态，1冻结不展示，0未冻结 可展示
	ContestFrozen int64 `json:"contest_freeze"`
	// 比赛状态，枚举：1未开始，2进行中，3已结束
	ContestStatus int64 `json:"contest_status"`
}

type ContestModel struct {
	ID int64 `json:"id"`
	//比赛阶段文本信息
	GameStage string `json:"game_stage"`
	//比赛开始时间
	Stime int64 `json:"stime"`
	//比赛结束时间
	Etime int64 `json:"etime"`
	//主场队伍id
	HomeID int64 `json:"home_id"`
	//客场队伍id
	AwayID int64 `json:"away_id"`
	//主场分数
	HomeScore int64 `json:"home_score"`
	//客场分数
	AwayScore int64 `json:"away_score"`
	// 游戏详情
	//赛季id
	Sid int64 `json:"sid"`
	//赛事id
	Mid int64 `json:"mid"`
	// 赛程关联的阶段id
	SeriesId int64 `json:"series_id"`
	//直播房间号
	LiveRoom int64 `json:"live_room"`
	//回播房间号
	Aid int64 `json:"aid"`
	//集锦房间号
	Collection int64 `json:"collection"`
	// 赛程描述
	Dic string `json:"dic"`
	//赛程类型：0普通1特殊
	Special int64 `json:"special"`
	//特殊赛程的胜利战队, 无主客队仅有胜利队伍
	SuccessTeam int64 `json:"success_team"`
	//特殊赛程，赛程名称
	SpecialName string `json:"special_name"`
	//特殊赛程，胜利文案
	SpecialTips string `json:"special_tips"`
	//特殊赛程，赛程图片
	SpecialImage string `json:"special_image"`
	//回播房间号url
	Playback string `json:"playback"`
	//集锦房间号url
	CollectionURL string `json:"collection_url"`
	//集锦房间号url
	LiveURL string `json:"live_url"`
	//比赛数据页类型 0：无 1：LOL 2:DATA2
	DataType int64 `json:"data_type"`
	// 赛程的冻结状态，1冻结不展示，0未冻结 可展示
	Status int64 `json:"status"`
	// 比赛状态，枚举：1未开始，2进行中，3已结束
	ContestStatus int64 `json:"contest_status"`
}
