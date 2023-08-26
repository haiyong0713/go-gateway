package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type TimelineMore struct {
	ButtonText       string
	SupernatantTitle string
	Params           string
}

func NewTimelineMore(buttonText string, supernatantTitle string, params string) *TimelineMore {
	return &TimelineMore{ButtonText: buttonText, SupernatantTitle: supernatantTitle, Params: params}
}

func (bu TimelineMore) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeTimelineMore.String(),
		CardDetail: &api.ModuleItem_TimelineMoreCard{
			TimelineMoreCard: &api.TimelineMoreCard{
				ButtonText:       bu.ButtonText,
				SupernatantTitle: bu.SupernatantTitle,
				Params:           bu.Params,
			},
		},
	}
}
