package jsonwebcard

type WebDynamicForwardCard struct {
	*Base
	Basic   *Basic    `json:"basic,omitempty"`
	Fold    *Fold     `json:"fold,omitempty"`
	Modules *Modules  `json:"modules,omitempty"`
	Orig    TopicCard `json:"orig,omitempty"`
}

func (c *WebDynamicForwardCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicForwardCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicForward
}

func (c *WebDynamicForwardCard) GetModules() *Modules {
	return c.Modules
}
