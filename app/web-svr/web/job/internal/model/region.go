package model

type Region struct {
	ID       int64           `json:"id"`
	Rid      int64           `json:"rid"`
	Reid     int64           `json:"reid"`
	Name     string          `json:"name"`
	Logo     string          `json:"logo"`
	URI      string          `json:"uri,omitempty"`
	Children []*Region       `json:"children,omitempty"`
	Config   []*RegionConfig `json:"config,omitempty"`
	Plat     int64           `json:"plat"`
	Area     string          `json:"area"`
	Language string          `json:"language"`
}

type RegionConfig struct {
	ID         int64  `json:"id"`
	Rid        int64  `json:"rid"`
	ScenesID   int64  `json:"scenes_id"`
	ScenesName string `json:"scenes_name,omitempty"`
	ScenesType string `json:"scenes_type,omitempty"`
}

// nolint:gomnd
func (c *RegionConfig) ConfigChange() {
	switch c.ScenesID {
	case 0:
		c.ScenesName = "region"
		c.ScenesType = "bottom"
	case 1:
		c.ScenesName = "region"
		c.ScenesType = "top"
	case 2:
		c.ScenesName = "rank"
	case 3:
		c.ScenesName = "search"
	case 4:
		c.ScenesName = "tag"
	case 5:
		c.ScenesName = "attention"
	}
}
