package jsonwebcard

type WebDynamicWordCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicWordCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicWordCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicWord
}

func (c *WebDynamicWordCard) GetModules() *Modules {
	return c.Modules
}
