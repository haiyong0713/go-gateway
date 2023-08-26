package common

type Version struct {
	Plat       int    `json:"plat"`
	Build      int    `json:"build"`
	Conditions string `json:"conditions"`
}

const (
	//安卓:0  iPhone:1
	Android = "0"
	IPhone  = "1"
)

const (
	PlatAndroid = 0
	PlatIPhone  = 2
	PlatWeb     = 30
)
