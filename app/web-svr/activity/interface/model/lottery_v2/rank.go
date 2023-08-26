package lottery

import "fmt"

type ProcessRate struct {
	Rate  float64 `json:"rate"`
	Clues []Clue  `json:"clues"`
}

type Clue struct {
	*Item
	Status bool `json:"status"`
}

type Item struct {
	Title  string `json:"title" toml:"Title"`
	SrcPc  string `json:"src_pc" toml:"SrcPc"`
	SrcH5  string `json:"src_h5" toml:"SrcH5"`
	Avatar string `json:"avatar" toml:"Avatar"`
}

func (item *Item) String() string {
	return fmt.Sprintf("title:%v , src_pc:%v , src_h5:%v , avatar:%v",
		item.Title, item.SrcPc, item.SrcH5, item.Avatar)
}
