package cheese

type Cheese struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	SubTitle string  `json:"subtitle"`
	CardInfo string  `json:"card_info"`
	Cover    string  `json:"cover"`
	Button   *Button `json:"button"`
}

type Button struct {
	Type      int           `json:"type"`
	JumpURL   string        `json:"jump_url"`
	JumpStyle *ButtonCommon `json:"jump_style"`
	Check     *ButtonCommon `json:"check"`
	UnCheck   *ButtonCommon `json:"uncheck"`
}

type ButtonCommon struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}
