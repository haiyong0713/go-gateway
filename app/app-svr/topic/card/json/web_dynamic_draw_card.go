package jsonwebcard

type WebDynamicDrawCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicDrawCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicDrawCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicDraw
}

func (c *WebDynamicDrawCard) GetModules() *Modules {
	return c.Modules
}
