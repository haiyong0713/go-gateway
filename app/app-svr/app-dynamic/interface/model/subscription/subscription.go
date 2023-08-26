package subscription

type Subscription struct {
	OID           int64  `json:"oid"`
	Icon          string `json:"icon"`
	Title         string `json:"title"`
	Desc          string `json:"desc"`
	JumpURL       string `json:"jump_url"`
	TagName       string `json:"tag_name"`
	TagColor      string `json:"tag_color"`
	TagColorNight string `json:"tag_color_night"`
	Tips          string `json:"tips"`
	MenuText      string `json:"menu_text"`
}
