package model

type TreeRoleApp struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        int64  `json:"type"`
	Role        int64  `json:"role"`
	DiscoveryID string `json:"discovery_id"`
	SRE         bool   `json:"sre"`
	Level       int64  `json:"level"`
	Leader      bool   `json:"leader"`
}
