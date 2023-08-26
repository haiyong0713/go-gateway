package model

type Region struct {
	Rid      int64           `json:"rid"`
	Reid     int64           `json:"reid"`
	Name     string          `json:"name"`
	Logo     string          `json:"logo"`
	URI      string          `json:"uri"`
	Children []*Region       `json:"children,omitempty"`
	Config   []*RegionConfig `json:"config,omitempty"`
	Area     string          `json:"area,omitempty"`
}

type RegionConfig struct {
	ScenesName string `json:"scenes_name"`
}
