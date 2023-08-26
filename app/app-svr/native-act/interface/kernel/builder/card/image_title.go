package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type ImageTitle struct {
	Image string
}

func NewImageTitle(image string) *ImageTitle {
	return &ImageTitle{Image: image}
}

func (bu *ImageTitle) Build() *api.ModuleItem {
	return &api.ModuleItem{
		CardType: model.CardTypeImageTitle.String(),
		CardDetail: &api.ModuleItem_ImageTitleCard{
			ImageTitleCard: &api.ImageTitleCard{
				Image: bu.Image,
			},
		},
	}
}
