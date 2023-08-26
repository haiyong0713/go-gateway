package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type VideoMore struct {
	Text        string
	Uri         string
	SubpageData *api.SubpageData
}

func (bu VideoMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeVideoMore.String(),
		CardDetail: &api.ModuleItem_VideoMoreCard{
			VideoMoreCard: &api.VideoMoreCard{
				Text:        bu.Text,
				Uri:         bu.Uri,
				SubpageData: bu.SubpageData,
			},
		},
	}
}

func NewVideoMore(text string, uri string, subpageData *api.SubpageData) *VideoMore {
	return &VideoMore{Text: text, Uri: uri, SubpageData: subpageData}
}
