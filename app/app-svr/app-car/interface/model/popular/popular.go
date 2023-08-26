package popular

import (
	"encoding/json"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
)

type PopularParam struct {
	model.DeviceInfo
	Pos      int    `form:"position"`
	FromType string `form:"from_type"`
	ParamStr string `form:"param"`
}

type PopularCard struct {
	ID              int64                       `json:"id"`
	Title           string                      `json:"title"`
	ChannelID       int64                       `json:"channel_id"`
	Type            string                      `json:"type"`
	Value           int64                       `json:"value"`
	Reason          string                      `json:"reason"`
	ReasonType      int8                        `json:"reason_type"`
	Pos             int                         `json:"pos"`
	FromType        string                      `json:"from_type"`
	PopularCardPlat map[int8][]*PopularCardPlat `json:"popularcardplat"`
	Idx             int                         `json:"-"`
	CornerMark      int8                        `json:"corner_mark"`
	CoverGif        string                      `json:"cover_gif"`
	HotwordID       int64                       `json:"hotword_id"`
	CanPlay         bool                        `json:"-"`
	// infoc
	TrackID   string          `json:"trackid,omitempty"`
	Source    string          `json:"source,omitempty"`
	AvFeature json.RawMessage `json:"av_feature,omitempty"`
}

type PopularCardPlat struct {
	CardID    int64  `json:"card_id"`
	Plat      int8   `json:"plat"`
	Condition string `json:"condition"`
	Build     int    `json:"build"`
}

func (c *PopularCard) PopularCardToAiChange() (a *ai.Item) {
	a = &ai.Item{
		Goto:    c.Type,
		ID:      c.Value,
		TrackID: c.TrackID,
	}
	return
}

type MediaPopularParam struct {
	Pn int `form:"pn" validate:"min=1"`
	Ps int `form:"ps" validate:"min=1,max=50"`
}
