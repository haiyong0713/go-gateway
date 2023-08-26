package card

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
)

type Card struct {
	ID         int64  `json:"-"`
	Title      string `json:"-"`
	ChannelID  int64  `json:"-"`
	Type       string `json:"-"`
	Value      int64  `json:"-"`
	Reason     string `json:"-"`
	ReasonType int8   `json:"-"`
	Pos        int    `json:"-"`
	FromType   string `json:"-"`
}

type CardPlat struct {
	CardID    int64  `json:"-"`
	Plat      int8   `json:"-"`
	Condition string `json:"-"`
	Build     int    `json:"-"`
}

func (c *Card) CardToAiChange() (a *ai.Item) {
	a = &ai.Item{
		Goto:       c.Type,
		ID:         c.Value,
		RcmdReason: c.fromRcmdReason(),
	}
	return
}

func (c *Card) fromRcmdReason() (a *ai.RcmdReason) {
	var content string
	// nolint:gomnd
	switch c.ReasonType {
	case 0:
		content = ""
	case 1:
		content = "编辑精选"
	case 2:
		content = "热门推荐"
	case 3:
		content = c.Reason
	}
	if content != "" {
		a = &ai.RcmdReason{ID: 1, Content: content, BgColor: "yellow", IconLocation: "left_top"}
	}
	return
}
