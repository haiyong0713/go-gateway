package push

const (
	// msg_source
	MsgSourceUnderage    = "app_underage"
	MsgSourceSleepRemind = "app_sleep_remind"
	MsgSourceTimelock    = "app_timelock"
)

type Message struct {
	Title     string `json:"title,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Position  int32  `json:"position,omitempty"`
	Duration  int32  `json:"duration,omitempty"`
	Expire    int64  `json:"expire,omitempty"`
	MsgSource string `json:"msg_source,omitempty"`
	HideArrow bool   `json:"hide_arrow,omitempty"`
	Link      string `json:"link,omitempty"`
}
