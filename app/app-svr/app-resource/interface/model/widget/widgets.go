package widget

type WidgetsMetaReq struct {
	Buvid    string `form:"-"`
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Platform string `form:"platform"`
	Build    int64  `form:"build"`
}

type WidgetsMeta struct {
	WidgetButtons []*Button `json:"widget_button"`
	HotWord       string    `json:"hot_word,omitempty"`
}

type WidgetsAndroidMeta struct {
	UserInfo      *UserInfo `json:"user_info"`
	WidgetButtons []*Button `json:"widget_button"`
	HotWord       string    `json:"hot_word,omitempty"`
}

type UserInfo struct {
	Mid  int64  `json:"mid"`
	Face string `json:"face"`
	Name string `json:"name"`
}

type Button struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
	URL  string `json:"url"`
}

type Hot struct {
	Code      int    `json:"code,omitempty"`
	SeID      string `json:"seid,omitempty"`
	Tips      string `json:"recommend_tips,omitempty"`
	NumResult int    `json:"numResult,omitempty"`
	ShowFront int    `json:"show_front,omitempty"`
	Result    []struct {
		ID        int64  `json:"id,omitempty"`
		Name      string `json:"name,omitempty"`
		ShowName  string `json:"show_name,omitempty"`
		Type      string `json:"type,omitempty"`
		GotoType  int    `json:"goto_type,omitempty"`
		GotoValue string `json:"goto_value,omitempty"`
		ModuleID  int64  `json:"module_id,omitempty"`
	} `json:"result,omitempty"`
	Trackid string `json:"trackid,omitempty"`
	ExpStr  string `json:"exp_str,omitempty"`
}
