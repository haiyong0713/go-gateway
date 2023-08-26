package model

type DynamicDetailReq struct {
	Platfrom string `json:"platfrom"`
	MobiApp  string `json:"mobi_app"`
	Device   string `json:"device"`
	Build    string `json:"build"`
}
