package databus

// const is
const (
	//season
	RouteSeasonShow = "season_show"
	// season with archive
	SeasonRouteForUpdate = "season_update"
	SeasonRouteForRemove = "season_remove"
)

// SeasonMsg is
type SeasonMsg struct {
	Route    string `json:"route"`
	SeasonID int64  `json:"season_id"`
}

// SeasonWithArchive is
type SeasonWithArchive struct {
	Route    string  `json:"route"`
	SeasonID int64   `json:"season_id"`
	Aids     []int64 `json:"aids"`
}
