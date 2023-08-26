package anti_addiction

type SleepRemind struct {
	ID     int64  `json:"ID"`
	Mid    int64  `json:"mid"`
	Switch int64  `json:"switch"`
	Stime  string `json:"stime"`
	Etime  string `json:"etime"`
}
