package tab

type Tab struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	URI          string `json:"uri,omitempty"`
	TabID        string `json:"tab_id,omitempty"`
	Pos          int    `json:"pos,omitempty"`
	IsDefault    bool   `json:"is_default,omitempty"`
	Icon         string `json:"icon,omitempty"`
	IconSelected string `json:"icon_selected,omitempty"`
}

type TabWeb struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Goto         string `json:"goto,omitempty"`
	TabID        string `json:"tab_id,omitempty"`
	Pos          int    `json:"pos,omitempty"`
	IsDefault    bool   `json:"is_default,omitempty"`
	Icon         string `json:"icon,omitempty"`
	IconSelected string `json:"icon_selected,omitempty"`
}
