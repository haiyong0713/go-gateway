package like

import xtime "go-common/library/time"

type Cpc100Info struct {
	List []*Cpc100Egg `json:"list"`
}
type Cpc100Egg struct {
	Name     string            `json:"name"`
	Data     map[string]string `json:"data"`
	STime    xtime.Time        `json:"stime"`
	ETime    xtime.Time        `json:"etime"`
	Key      string            `json:"key"`
	Unlocked bool              `json:"unlocked"`
}
