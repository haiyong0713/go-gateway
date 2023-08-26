package model

import "fmt"

type AutoSubscribeSeason struct {
	SeasonID int64 `json:"season_id"`
}

type AutoSubscribeDetail struct {
	SeasonID  int64 `json:"season_id"`
	TeamId    int64 `json:"team_id"`
	ContestID int64 `json:"contest_id"`
}

const (
	sql2CreateAutoSubSeasonDetail = `CREATE TABLE auto_subscribe_season_detail_%v LIKE auto_subscribe_season_detail`

	tableName4AutoSubSeasons = "auto_subscribe_seasons"
)

func GenAutoSubSeasonDetailSql(seasonID int64) string {
	return fmt.Sprintf(sql2CreateAutoSubSeasonDetail, seasonID)
}

func (*AutoSubscribeSeason) TableName() string {
	return tableName4AutoSubSeasons
}
