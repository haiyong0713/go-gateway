package digital

import digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"

type SpaceDigitalInfoResp struct {
	// nft id
	NftId string `json:"nft_id"`
	// NFT系列名称，例如"干杯2022！"
	Name string `json:"name"`
	// NFT编号，例如 "#2101/2233"
	SerialNumber string `json:"serial_number"`
	// NFT链上地址，例如 "up14hndnqq2xzkh4jvnjjsyj9kdjjrpsa9h7s2ke2#2101"
	NftAddress string `json:"nft_address"`
	// NFT图像地址
	Image string `json:"image"`
	// NFT视频地址
	AnimationUrl string `json:"animation_url"`
	// NFT详情页跳转链接
	DetailJump string `json:"detail_jump"`
	// 拥有者NFT列表页跳转链接
	OwnerListJump string `json:"owner_list_jump"`
	// NFT系列合集页跳转链接
	ItemGalleryJump string `json:"item_gallery_jump"`
	// NFT系列号
	ItemId int64 `json:"item_id"`
	// 概览页背景图地址
	BackgroundImage string `json:"background_image"`
	// NFT附属品信息
	Appendage *Appendage `json:"appendage"`
	// NFT点赞信息
	LikeInfo *LikeInfo `json:"like_info"`
	// NFT类型
	NftType int32 `json:"nft_type"`
	// NFT所属区域 0 默认 1 大陆 2 港澳台
	RegionType int32 `json:"region_type,omitempty"`
	// icon
	Icon string `json:"icon,omitempty"`
	// 视频头图地址列表
	AnimationUrlList    []string                `json:"animation_url_list,omitempty"`
	BackgroundHandle    int32                   `json:"background_handle"`
	AnimationFirstFrame string                  `json:"animation_first_frame"`
	MusicAlbum          *digitalgrpc.MusicAlbum `json:"music_album"`
	Animation           *digitalgrpc.Animation  `json:"animation"`
	// NFT所属区域的展示文案
	NftRegionTitle string `json:"nft_region_title"`
	// NFT图片相关元数据
	NFTImage *digitalgrpc.NFTImage `json:"nft_image,omitempty"`
}

type Appendage struct {
	// 附属头像图片地址
	Avatar string `json:"avatar"`
	// 附属头像边框图片地址
	AvatarFrame string `json:"avatar_frame"`
	// 附属视频缩略图地址
	VideoThumbnail string `json:"video_thumbnail,omitempty"`
	// 附属视频边框图片地址
	VideoFrame string `json:"video_frame,omitempty"`
	// 附属nft id
	NftId string `json:"nft_id,omitempty"`
	// 附属图片处理方式，0/1 圆形、2 圆角矩形
	ImageHandleType int32 `json:"image_handle_type,omitempty"`
	// 附属nft类型
	NftType int32 `json:"nft_type,omitempty"`
}

type LikeInfo struct {
	// nft id
	NftId string `json:"nft_id"`
	// nft的点赞总数
	LikeNumber int64 `json:"like_number"`
	// 发起请求用户的点赞状态
	LikeState int32 `json:"like_state"`
}

type SpaceDigitalExtraInfoResp struct {
	// NFT类型 2 视频 4 音乐
	NftType int32 `json:"nft_type"`
	// 头像地址
	Image string `json:"image"`
	// 音乐专辑
	MusicAlbum *digitalgrpc.MusicAlbum `json:"music_album,omitempty"`
	// 视频内容
	Animation *digitalgrpc.Animation `json:"animation,omitempty"`
	// NFT图片相关元数据
	NFTImage *digitalgrpc.NFTImage `json:"nft_image,omitempty"`
}

func (s *SpaceDigitalInfoResp) FromGetGarbSpaceInfoResp(r *digitalgrpc.GetGarbSpaceInfoResp) {
	if r == nil {
		return
	}
	s.NftId = r.NftId
	s.Name = r.Name
	s.SerialNumber = r.SerialNumber
	s.NftAddress = r.NftAddress
	s.Image = r.Image
	s.AnimationUrl = r.AnimationUrl
	s.DetailJump = r.DetailJump
	s.OwnerListJump = r.OwnerListJump
	s.ItemGalleryJump = r.ItemGalleryJump
	s.ItemId = r.ItemId
	s.BackgroundImage = r.BackgroundImage
	s.Appendage = &Appendage{
		Avatar:          r.GetAppendage().GetAvatar(),
		AvatarFrame:     r.GetAppendage().GetAvatarFrame(),
		VideoThumbnail:  r.GetAppendage().GetVideoThumbnail(),
		VideoFrame:      r.GetAppendage().GetVideoFrame(),
		NftId:           r.GetAppendage().GetNftId(),
		ImageHandleType: int32(r.GetAppendage().GetImageHandleType()),
		NftType:         int32(r.GetAppendage().GetNftType()),
	}
	s.LikeInfo = &LikeInfo{
		NftId:      r.GetLikeInfo().GetNftId(),
		LikeNumber: r.GetLikeInfo().GetLikeNumber(),
		LikeState:  int32(r.GetLikeInfo().GetLikeState()),
	}
	s.NftType = int32(r.NftType)
	s.RegionType = int32(r.NftRegion)
	s.Icon = r.NftRegionIcon
	s.AnimationUrlList = r.AnimationUrlList
	s.BackgroundHandle = int32(r.BackgroundHandle)
	s.AnimationFirstFrame = r.AnimationFirstFrame
	s.MusicAlbum = r.MusicAlbum
	s.Animation = r.Animation
	s.NftRegionTitle = r.NftRegionTitle
	s.NFTImage = r.GetNftImage()
}

func (s *SpaceDigitalExtraInfoResp) SpaceExtraInfoResp(r *digitalgrpc.SpaceExtraInfoResp) {
	if r == nil {
		return
	}
	s.NftType = int32(r.NftType)
	s.Image = r.Image
	s.MusicAlbum = r.MusicAlbum
	s.Animation = r.Animation
	s.NFTImage = r.GetNftImage()
}
