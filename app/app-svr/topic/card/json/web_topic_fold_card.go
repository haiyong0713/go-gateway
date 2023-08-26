package jsonwebcard

import "fmt"

type WebTopicFoldCard struct {
	FoldCount    int64  `json:"fold_count"`
	CardShowDesc string `json:"card_show_desc"`
	FoldDesc     string `json:"fold_desc"`
}

func (c *WebTopicFoldCard) GetModules() *Modules {
	return nil
}

func (c *WebTopicFoldCard) GetTopicCardType() TopicCardType {
	return TopicCardTypeFold
}

func ConstructTopicFoldCard(foldCount int64, foldDesc string) *WebTopicFoldCard {
	return &WebTopicFoldCard{
		FoldCount:    foldCount,
		CardShowDesc: fmt.Sprintf("有%d条内容被折叠", foldCount),
		FoldDesc:     foldDesc,
	}
}
