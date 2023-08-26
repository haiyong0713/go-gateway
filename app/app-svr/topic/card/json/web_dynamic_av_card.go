package jsonwebcard

type WebDynamicAvCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicAvCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicAvCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicAv
}

func (c *WebDynamicAvCard) GetModules() *Modules {
	return c.Modules
}
