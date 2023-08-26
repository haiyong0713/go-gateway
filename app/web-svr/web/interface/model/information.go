package model

type Informations struct {
	Page  *Page          `json:"page"`
	Items []*Information `json:"items"`
}

type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Count int `json:"count"`
}

type Information struct {
	CardType string  `json:"card_type"`
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Cover    string  `json:"cover"`
	Author   *Author `json:"author"`
	Stat     *Stat   `json:"stat"`
	Duration int64   `json:"duration"`
	Position int     `json:"position"`
	BVID     string  `json:"bvid"`
	Cid      int64   `json:"cid"`
}

type Author struct {
	MID  int64  `json:"mid"`
	Face string `json:"face"`
	Name string `json:"name"`
}
