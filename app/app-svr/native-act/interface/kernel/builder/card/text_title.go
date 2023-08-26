package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type TextTitle struct {
	Title string
}

func NewTextTitle(text string) *TextTitle {
	return &TextTitle{Title: text}
}

func (bu *TextTitle) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeTextTitle.String(),
		CardDetail: &api.ModuleItem_TextTitleCard{
			TextTitleCard: &api.TextTitleCard{
				Title: bu.Title,
			},
		},
	}
}
