package model

type AutoSubscribeDetail struct {
	SeasonID  int64 `json:"season_id"`
	TeamId    int64 `json:"team_id"`
	ContestID int64 `json:"contest_id"`
}
