package model

type WFStatus struct {
	ID           int64  `json:"id"`
	WFName       string `json:"wf_name"`
	DiscoveryID  string `json:"discovery_id"`
	DisplayName  string `json:"display_name"`
	DisplayState int8   `json:"display_state"`
	CodeAddress  string `json:"code_address"`
	CodeVersion  string `json:"code_version"`
	State        int8   `json:"state"`
	Logs         string `json:"logs"`
}
