package channel_v2

import (
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
)

type Home2 struct {
	EntranceButton *EntranceButton `json:"entrance_button,omitempty"`
	SquareItems    []*SquareItem   `json:"square_items"`
}

type EntranceButton struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	Link string `json:"link,omitempty"`
}

type SquareItem struct {
	ModelType  string        `json:"model_type"`
	ModelTitle string        `json:"model_title"`
	HasMore    int           `json:"has_more"`
	Label      string        `json:"label,omitempty"`
	Offset     string        `json:"offset"`
	DescButton *cardm.Button `json:"desc_button,omitempty"`
	Items      interface{}   `json:"items,omitempty"`
}
