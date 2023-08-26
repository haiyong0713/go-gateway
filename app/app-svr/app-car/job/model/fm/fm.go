package fm

const (
	TypeUpdate = ModifyType("update")
	TypeInsert = ModifyType("insert")

	AudioSeason   = FmType("audio_season")
	AudioSeasonUp = FmType("audio_season_up")
)

type FmType string

type ModifyType string

type CommonSeason struct {
	Scene Scene       `json:"scene"`
	Fm    FmSeason    `json:"fm"`
	Video VideoSeason `json:"video"`
}

// FmSeason FM合集
type FmSeason struct {
	Scene     Scene   `json:"scene"`
	FmType    FmType  `json:"fm_type"`
	FmId      int64   `json:"fm_id"`
	Title     string  `json:"title"`
	Cover     string  `json:"cover"`
	FmListStr string  `json:"fm_list"`
	FmList    []*Item `json:"-"`
}

// VideoSeason 视频合集
type VideoSeason struct {
	Scene         Scene   `json:"scene"`
	SeasonId      int64   `json:"season_id"`
	Title         string  `json:"title"`
	Cover         string  `json:"cover"`
	SeasonListStr string  `json:"season_list"`
	SeasonList    []*Item `json:"-"`
}

type Item struct {
	Aid int64 `json:"aid"`
}
