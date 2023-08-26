package favorite

import (
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/model"

	favmodel "git.bilibili.co/bapis/bapis-go/community/model/favorite"
)

const (
	AttrBitPublic = uint32(0)
	AttrIsPublic  = int32(0) // 公开
	IsFav         = 1        // 已收藏
)

type Space struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Media *Media `json:"mediaListResponse"`
}

type Media struct {
	Count     int          `json:"count"`
	MediaList []*MediaList `json:"list"`
}

type MediaList struct {
	ID    int64 `json:"id"`
	Mid   int64 `json:"mid"`
	Upper struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
	Type       int    `json:"type"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	CoverType  int    `json:"cover_type"`
	Intro      string `json:"intro"`
	Attr       int    `json:"attr"`
	State      int    `json:"state"`
	FavState   int    `json:"fav_state"`
	MediaCount int    `json:"media_count"`
}

type MediaParam struct {
	model.DeviceInfo
}

type FavoriteParam struct {
	model.DeviceInfo
	Pn int `form:"pn" default:"1" validate:"min=1"`
}

type MediaListParam struct {
	model.DeviceInfo
	FollowType int    `form:"follow_type"`
	Pn         int    `form:"pn" default:"1" validate:"min=1"`
	Ps         int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FavID      int64  `form:"fav_id"`
	FromType   string `form:"from_type"`
	Vmid       int64  `form:"vmid"`
	ParamStr   string `form:"param"`
}

type ToViewParam struct {
	model.DeviceInfo
	FromType string `form:"from_type"`
	Pn       int    `form:"pn" default:"1" validate:"min=1"`
	Ps       int    `form:"ps" default:"20" validate:"min=1,max=20"`
	ParamStr string `form:"param"`
}

type FavAddOrDelFolders struct {
	model.DeviceInfo
	AddFids string `form:"add_fids"`
	DelFids string `form:"del_fids"`
	Aid     int64  `form:"aid"`
	Oid     int64  `form:"oid"`
}

type AddFolder struct {
	Name   string `form:"name" validate:"required"`
	Desc   string `form:"desc"`
	Public int    `form:"public"`
}

type UserFolderParam struct {
	Aid int64 `form:"aid"`
}

type UserFolder struct {
	Fid      int64  `json:"fid"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	FavState int32  `json:"fav_state"`
}

func (f *UserFolder) FromUserFolder(fav *favmodel.Folder) {
	f.Fid = fav.ID
	f.Title = fav.Name
	f.Desc = fmt.Sprintf("%d个内容", fav.Count)
	f.FavState = fav.Favored
	// 0 - 是否公开（1为私密）
	if model.AttrVal(int32(fav.Attr), AttrBitPublic) == AttrIsPublic {
		f.Desc = f.Desc + " · " + "公开"
	} else {
		f.Desc = f.Desc + " · " + "私密"
	}
}
