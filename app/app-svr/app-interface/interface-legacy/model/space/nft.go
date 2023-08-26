package space

import gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"

type NftShowModule struct {
	Total        int64  `json:"total,omitempty"`
	ArtsMoreJump string `json:"arts_more_jump,omitempty"`
	Nfts         []*Nft `json:"nfts,omitempty"`
	FloorTitle   string `json:"floor_title"`
}

type Nft struct {
	ItemName     string               `json:"item_name"`
	Issuer       string               `json:"issuer"`
	SerialNumber string               `json:"serial_number"`
	DetailUrl    string               `json:"detail_url"`
	NftStatus    int64                `json:"nft_status"`
	Display      *gallerygrpc.Display `json:"display"`
}

type NftDisplay struct {
	BgThemeLight string `json:"bg_theme_light"`
	BgThemeNight string `json:"bg_theme_night"`
	NftPoster    string `json:"nft_poster"`
}

type NftFaceIcon struct {
	RegionType int32  `json:"region_type"` // nft所属区域 0 默认 1 大陆 2 港澳台
	Icon       string `json:"icon"`        // 角标链接
	ShowStatus int32  `json:"show_status"` // 展示状态 0:默认 1:放大20% 2:原图大小
}

type NftFaceButton struct {
	FaceButtonChs string `json:"face_button_chs"` // 头像按钮文案简体
	FaceButtonCht string `json:"face_button_cht"` // 头像按钮文案繁体
}

func ConvertToNftShowModule(total int64, artsMoreJump, floorTitle string, artsList []*gallerygrpc.NFT) *NftShowModule {
	var nft []*Nft
	for _, v := range artsList {
		nft = append(nft, &Nft{
			ItemName:     v.ItemName,
			Issuer:       v.Issuer,
			SerialNumber: v.SerialNumber,
			DetailUrl:    v.DetailUrl,
			NftStatus:    int64(v.NftStatus),
			Display:      v.Display,
		})
	}
	return &NftShowModule{
		Total:        total,
		ArtsMoreJump: artsMoreJump,
		FloorTitle:   floorTitle,
		Nfts:         nft,
	}
}
