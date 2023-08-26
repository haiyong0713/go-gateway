package space

type SeasonArchiveParam struct {
	SeasonId int64  `form:"season_id" validate:"required"`
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Clocale  string `form:"clocale"`
	Slocale  string `form:"slocale"`
	SLocaleP string `form:"s_locale"`
	CLocaleP string `form:"c_locale"`
	Sort     string `form:"sort"`
}

type SeasonArchiveResp struct {
	Item           []*ArcItem      `json:"item"`
	EpisodicButton *EpisodicButton `json:"episodic_button,omitempty"`
	Order          []*ArcOrder     `json:"order,omitempty"`
}
