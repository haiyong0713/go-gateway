package model

type RankReply struct {
	Page   *Page           `json:"page"`
	Ranks  []*PlayerRankV2 `json:"list"`
	Player *PlayerRankV2   `json:"player,omitempty"`
}

type Page struct {
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
	Total int `json:"total"`
}

type PlayerRankV2 struct {
	*Player
	Score int `json:"score"`
	Rank  int `json:"rank"`
}
