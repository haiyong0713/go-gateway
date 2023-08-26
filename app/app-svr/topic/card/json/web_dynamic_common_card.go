package jsonwebcard

type WebDynamicCommonCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicCommonCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicCommonCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicCommon
}

func (c *WebDynamicCommonCard) GetModules() *Modules {
	return c.Modules
}
