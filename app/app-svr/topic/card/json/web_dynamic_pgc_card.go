package jsonwebcard

type WebDynamicPGCCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicPGCCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicPGCCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicPGC
}

func (c *WebDynamicPGCCard) GetModules() *Modules {
	return c.Modules
}
