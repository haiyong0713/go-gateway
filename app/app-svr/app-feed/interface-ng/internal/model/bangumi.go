package model

type EpPlayerReq struct {
	EpIDs    []int64 `json:"ep_i_ds"`
	MobiApp  string  `json:"mobi_app"`
	Platform string  `json:"platform"`
	Device   string  `json:"device"`
	Build    int     `json:"build"`
	Fnver    int     `json:"fnver"`
	Fnval    int     `json:"fnval"`
}
