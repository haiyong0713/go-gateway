package operate

import (
	"strconv"

	"go-gateway/app/app-svr/app-car/interface/model"
)

type Card struct {
	Plat     int8         `json:"plat,omitempty"`
	Build    int          `json:"build,omitempty"`
	Network  string       `json:"network,omitempty"`
	ID       int64        `json:"id,omitempty"`
	Cid      int64        `json:"cid,omitempty"`
	Param    string       `json:"param,omitempty"`
	CardGoto model.CardGt `json:"card_goto,omitempty"`
	Goto     string       `json:"goto,omitempty"`
	URI      string       `json:"uri,omitempty"`
	Title    string       `json:"title,omitempty"`
	Desc     string       `json:"desc,omitempty"`
	Cover    string       `json:"cover,omitempty"`
	Score    int32        `json:"score,omitempty"`
	TrackID  string       `json:"trackid,omitempty"`
	FromType string       `json:"from_type,omitempty"`
	MobiApp  string       `json:"mobi_app,omitempty"`
	Epid     int32        `json:"-"`
	ViewAt   int64        `json:"-"`
	Progress int64        `json:"-"`
	Duration int64        `json:"-"`
	// 入口类型
	Entrance string `json:"-"`
	Business string `json:"-"`
	// ======
	KeyWord    string `json:"-"`
	FollowType string `json:"-"`
	Rid        int64  `json:"-"`
	FavID      int64  `json:"-"`
	Vmid       int64  `json:"-"`
	DynCtime   int64  `json:"-"`
}

func (c *Card) From(cardGoto model.CardGt, entrance string, id int64, plat int8, build int, mobiApp string) {
	c.CardGoto = cardGoto
	c.Entrance = entrance
	c.ID = id
	c.Goto = string(cardGoto)
	c.Param = strconv.FormatInt(id, 10)
	c.URI = strconv.FormatInt(id, 10)
	c.Plat = plat
	c.Build = build
	c.MobiApp = mobiApp
}
