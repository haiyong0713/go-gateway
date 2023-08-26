package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type DynamicActMore struct {
	Text        string
	Uri         string
	SubpageData *api.SubpageData
}

func NewDynamicActMore(text string, uri string, subpageData *api.SubpageData) *DynamicActMore {
	return &DynamicActMore{Text: text, Uri: uri, SubpageData: subpageData}
}

func (bu DynamicActMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeDynamicActMore.String(),
		CardDetail: &api.ModuleItem_DynamicActMoreCard{
			DynamicActMoreCard: &api.DynamicActMoreCard{
				Text:        bu.Text,
				Uri:         bu.Uri,
				SubpageData: bu.SubpageData,
			},
		},
	}
}
