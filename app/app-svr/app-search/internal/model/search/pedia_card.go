package search

type PediaCardNavigation struct {
	Title          string                 `json:"title"`
	Children       []*PediaCardNavigation `json:"children"`
	InlineChildren []*PediaCardNavigation `json:"inline_children"`
	ReType         int64                  `json:"re_type"`
	ReValue        string                 `json:"re_value"`
	ModuleCount    int64                  `json:"module_count"`
	Button         *pediaCardButton       `json:"button"`
}

type pediaCardButton struct {
	Type    int64  `json:"1"`
	Text    string `json:"text"`
	ReType  int64  `json:"re_type"`
	ReValue string `json:"re_value"`
}

type PediaCard struct {
	ID             int64     `json:"id"`
	ExtraInfo      ExtraInfo `json:"extra_info"`
	NavigationCard struct {
		CardID         int64                `json:"card_id"`
		Title          string               `json:"title"`
		CornerType     int64                `json:"corner_type"`
		CornerText     string               `json:"corner_text"`
		CornerSunURL   string               `json:"corner_sun_url"`
		CornerNightURL string               `json:"corner_night_url"`
		CornerHeight   int64                `json:"corner_height"`
		CornerWidth    int64                `json:"corner_width"`
		BtnType        int64                `json:"btn_type"`
		BtnText        string               `json:"btn_text"`
		BtnReType      int64                `json:"btn_re_type"`
		BtnReValue     string               `json:"btn_re_value"`
		Avid           int64                `json:"avid"`
		Navigation     *PediaCardNavigation `json:"navigation"`
		CoverType      int64                `json:"cover_type"`
		CoverSunURL    string               `json:"cover_sun_url"`
		CoverNightURL  string               `json:"cover_night_url"`
		CoverWidth     int64                `json:"cover_width"`
		CoverHeight    int64                `json:"cover_height"`
	} `json:"navigation_card"`
}
