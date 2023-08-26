package show

type PopAIChannelResource struct {
	RID        int64  `json:"id" form:"id" validate:"id"`
	TagId      string `json:"tag" form:"tag" validate:"tag"`
	Goto       string `json:"goto"`
	FromType   string `json:"from_type"`
	Desc       string `json:"desc"`
	CornerMark int8   `json:"corner_mark"`
}

type PopularCard struct {
	Type       string `json:"type"`
	Value      int64  `json:"value"`
	FromType   string `json:"from_type"`
	TagId      string `json:"tag_id"`
	Reason     string `json:"reason"`
	CornerMark int8   `json:"corner_mark"`
}

type PopTopEntrance struct {
	ID   int   `json:"id"`
	Rank int64 `json:"rank"`
}

type PopularCardAI struct {
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
	Idx             int64                       `json:"-"`
	CornerMark      int8                        `json:"corner_mark"`
	CoverGif        string                      `json:"cover_gif"`
	HotwordID       int64                       `json:"hotword_id"`
	CanPlay         bool                        `json:"-"`
}

type PopularCardPlat struct {
	CardID    int64  `json:"card_id"`
	Plat      int8   `json:"plat"`
	Condition string `json:"condition"`
	Build     int    `json:"build"`
}

type CardListAI struct {
	ID         int64              `json:"id"`
	Goto       string             `json:"goto"`
	FromType   string             `json:"from_type"`
	Desc       string             `json:"desc"`
	CornerMark int8               `json:"corner_mark"`
	CoverGif   string             `json:"cover_gif"`
	Condition  []*CardConditionAI `json:"condition"`
	HotwordID  int64              `json:"hotword_id"`
	RcmdReason *RcmdReasonAI      `json:"rcmd_reason"`
}

type CardConditionAI struct {
	Plat      int8   `json:"plat"`
	Condition string `json:"conditions"`
	Build     int    `json:"build"`
}

type RcmdReasonAI struct {
	Content string `json:"content"`
	Style   int8   `json:"style"`
}

func (c *CardListAI) CardListChange() (p *PopularCardAI) {
	p = &PopularCardAI{
		Value:      c.ID,
		Type:       c.Goto,
		FromType:   c.FromType,
		Reason:     c.Desc,
		CornerMark: c.CornerMark,
		CoverGif:   c.CoverGif,
		HotwordID:  c.HotwordID,
	}
	if p.Reason != "" {
		p.ReasonType = 3
	}
	if len(c.Condition) > 0 {
		tmpcondition := map[int8][]*PopularCardPlat{}
		for _, condition := range c.Condition {
			tmpcondition[condition.Plat] = append(tmpcondition[condition.Plat], &PopularCardPlat{
				Plat:      condition.Plat,
				Condition: condition.Condition,
				Build:     condition.Build,
			})
		}
		p.PopularCardPlat = tmpcondition
	}
	return
}
