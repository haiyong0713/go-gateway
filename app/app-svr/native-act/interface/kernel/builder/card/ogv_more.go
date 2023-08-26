package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type OgvMore struct {
	ButtonText       string
	SupernatantTitle string
	Params           string
}

func NewOgvMore(buttonText string, supernatantTitle string, params string) *OgvMore {
	return &OgvMore{ButtonText: buttonText, SupernatantTitle: supernatantTitle, Params: params}
}

func (bu OgvMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeOgvMore.String(),
		CardDetail: &api.ModuleItem_OgvMoreCard{
			OgvMoreCard: &api.OgvMoreCard{
				ButtonText:       bu.ButtonText,
				SupernatantTitle: bu.SupernatantTitle,
				Params:           bu.Params,
			},
		},
	}
}
