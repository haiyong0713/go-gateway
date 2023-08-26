package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type Text struct {
	text string
}

func NewText(text string) *Text {
	return &Text{text: text}
}

func (bu *Text) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeText.String(),
		CardDetail: &api.ModuleItem_TextCard{
			TextCard: &api.TextCard{
				Text: bu.text,
			},
		},
	}
}
