package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type DynamicMore struct {
	Text        string
	Uri         string
	SubpageData *api.SubpageData
}

func (bu DynamicMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeDynamicMore.String(),
		CardDetail: &api.ModuleItem_DynamicMoreCard{
			DynamicMoreCard: &api.DynamicMoreCard{
				Text:        bu.Text,
				Uri:         bu.Uri,
				SubpageData: bu.SubpageData,
			},
		},
	}
}

func NewDynamicMore(text, uri string, subpageData *api.SubpageData) *DynamicMore {
	return &DynamicMore{Text: text, Uri: uri, SubpageData: subpageData}
}
