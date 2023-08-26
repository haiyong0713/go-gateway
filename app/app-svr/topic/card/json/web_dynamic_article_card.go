package jsonwebcard

type WebDynamicArticleCard struct {
	*Base
	Basic   *Basic   `json:"basic,omitempty"`
	Fold    *Fold    `json:"fold,omitempty"`
	Modules *Modules `json:"modules,omitempty"`
}

func (c *WebDynamicArticleCard) GetBase() *Base {
	return c.Base
}

func (c *WebDynamicArticleCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeDynamicArticle
}

func (c *WebDynamicArticleCard) GetModules() *Modules {
	return c.Modules
}
