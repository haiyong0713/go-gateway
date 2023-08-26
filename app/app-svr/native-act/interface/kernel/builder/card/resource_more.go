package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type ResourceMore struct {
	Text        string
	Uri         string
	SubpageData *api.SubpageData
}

func (bu ResourceMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeResourceMore.String(),
		CardDetail: &api.ModuleItem_ResourceMoreCard{
			ResourceMoreCard: &api.ResourceMoreCard{
				Text:        bu.Text,
				Uri:         bu.Uri,
				SubpageData: bu.SubpageData,
			},
		},
	}
}

func NewResourceMore(text string, uri string, subpageData *api.SubpageData) *ResourceMore {
	return &ResourceMore{Text: text, Uri: uri, SubpageData: subpageData}
}
