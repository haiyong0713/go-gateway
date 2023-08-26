package space

import (
	"fmt"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	api "git.bilibili.co/bapis/bapis-go/garb/service"
)

const (
	_garbTitlePostfix = "" // 先去掉这个值，后面可能还要加回来
	_fansNumberPrefix = "NO."
	_visitorPurchase  = "购买同款"
	_purchaseURI      = "/h5/mall/suit/detail?navhide=1&id="
)

// GarbDetailReq def.
type GarbDetailReq struct {
	GarbID  int64 `form:"garb_id" validate:"required"`
	ImageID int   `form:"image_id"` // 因为image_id实际上为index，从0开始所以不能校验
	UserReq
}

type UserReq struct {
	Vmid int64 `form:"vmid" validate:"required"`
	Mid  int64 `form:"-"`
}

// VisitMySpace the visited space's owner is the user
func (v UserReq) VisitMySpace() bool {
	if v.Vmid > 0 && v.Mid > 0 && v.Mid == v.Vmid {
		return true
	}
	return false
}

// GarbDetailReply def
type GarbDetailReply struct {
	GarbTitle  string             `json:"garb_title"`
	Mid        int64              `json:"mid"`
	Face       string             `json:"face"`
	Name       string             `json:"name"`
	SuitItemID int64              `json:"suit_item_id"`
	FansNumber string             `json:"fans_number,omitempty"`
	Images     []*GarbDetailImage `json:"images"`
}

// FromGarb combines account info & fan's number
func (v *GarbDetailReply) FromGarb(ownerInfo *accgrpc.Card, fansNumber int64) {
	v.Mid = ownerInfo.Mid // 主人mid
	v.Name = ownerInfo.Name
	v.Face = ownerInfo.Face
	if fansNumber > 0 {
		v.FansNumber = fansNbrLabel(fansNumber)
	}
}

func garbTitle(name string) string {
	return name + _garbTitlePostfix
}

func fansNbrLabel(number int64) string {
	return fmt.Sprintf("%s%06d", _fansNumberPrefix, number)
}

// GarbState 决定当前图和当前装扮的图
func (v *GarbDetailReply) GarbState(garbInfo *api.SpaceBG, req *GarbDetailReq, ownerEquipInfo *api.SpaceBGUserEquipReply) (legalImageID bool) {
	v.GarbTitle = garbTitle(garbInfo.Name)
	v.SuitItemID = garbInfo.SuitItemID
	for k, img := range garbInfo.Images {
		detailImg := new(GarbDetailImage)
		detailImg.FromSpaceBgImage(img, k)
		if k == req.ImageID {
			detailImg.IsCurrent = true
			legalImageID = true
		}
		if ownerEquipInfo != nil && // 买了并未装扮的情况下为nil
			ownerEquipInfo.Item.Id == garbInfo.Id { // 用户装扮的就是当前请求的粉丝头图
			if int64(k) == ownerEquipInfo.Index {
				detailImg.IsDressed = true
			}
		}
		v.Images = append(v.Images, detailImg)
	}
	return
}

// GarbDetailImage def.
type GarbDetailImage struct {
	ID         int    `json:"id"` // image_id 实质上是数组的index
	SmallImage string `json:"small_image"`
	LargeImage string `json:"large_image"`
	IsCurrent  bool   `json:"is_current,omitempty"`
	IsDressed  bool   `json:"is_dressed,omitempty"`
}

func (v *GarbDetailImage) FromSpaceBgImage(img *api.SpaceBGImage, idx int) {
	v.ID = idx
	v.SmallImage = img.Landscape
	v.LargeImage = img.Portrait
}

type GarbListReply struct {
	List  []*GarbListItem `json:"list"`
	Count int64           `json:"count"`
}

type GarbListReq struct {
	UserReq
	Pn int64 `form:"pn" default:"1"`
	PS int64 `form:"ps" default:"20"`
}

type GarbListItem struct {
	GarbTitle    string             `json:"garb_title"`
	GarbID       int64              `json:"garb_id"`
	FansNumber   string             `json:"fans_number,omitempty"`
	TitleColor   string             `json:"title_color"`
	TitleBgimage string             `json:"title_bg_image"`
	IsDressed    bool               `json:"is_dressed,omitempty"`
	Button       *BuyGarbButton     `json:"button,omitempty"`
	Images       []*GarbDetailImage `json:"images"`
}

func (v *GarbListItem) FromAsset(asset *api.SpaceBG, ownerFanNbrs map[int64]*api.UserFanInfoReply, ownerEquip *api.SpaceBGUserEquipReply, isOwner bool) {
	v.GarbTitle = garbTitle(asset.Name)
	v.GarbID = asset.Id
	v.TitleBgimage = asset.FanNOImage
	v.TitleColor = asset.FanNOColor
	if fansNbr, ok := ownerFanNbrs[asset.SuitItemID]; ok && fansNbr != nil && fansNbr.Number > 0 {
		v.FansNumber = fansNbrLabel(fansNbr.Number)
	}
	for k, assetImg := range asset.Images {
		img := new(GarbDetailImage)
		img.FromSpaceBgImage(assetImg, k)
		v.Images = append(v.Images, img)
	}
	if ownerEquip != nil && ownerEquip.Item != nil && ownerEquip.Item.Id == v.GarbID { // 装扮状态
		if isOwner { // 主人态才下发
			v.IsDressed = true
		}
		if int(ownerEquip.Index) < len(v.Images) { // image无论主人态客人态都下发
			v.Images[ownerEquip.Index].IsDressed = true
		}
	}
	//nolint:gosimple
	return
}

func (v *GarbListItem) PurchaseButton(visitorFanIDs map[int64]*api.UserFanInfoReply, suitItemID int64, hostCfg string) {
	if fansNbr, ok := visitorFanIDs[suitItemID]; ok && fansNbr != nil && fansNbr.Number > 0 { // 客人已购买不出现购买
		return
	}
	v.Button = &BuyGarbButton{
		Title: _visitorPurchase,
		URI:   fmt.Sprintf("%s%s%d", hostCfg, _purchaseURI, suitItemID),
	}
}

type BuyGarbButton struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}

type GarbDressReq struct {
	GarbID  int64 `form:"garb_id" validate:"required"`
	ImageID int64 `form:"image_id"`
	Mid     int64 `form:"-"`
}

type GarbDressReply struct {
	Count int64            `json:"count"`
	Total int64            `json:"total"`
	Items []*GarbDressItem `json:"item"`
}

type GarbDressItem struct {
	GarbTitle    string `json:"garb_title"`
	GarbID       int64  `json:"garb_id"`
	Image        string `json:"image"`
	ImageID      int64  `json:"image_id"`
	FansNumber   string `json:"fans_number,omitempty"`
	TitleColor   string `json:"title_color"`
	TitleBgimage string `json:"title_bg_image"`
	SuitBgColor  string `json:"suit_bg_color"`
	FanBgColor   string `json:"fan_bg_color"`
}

func (v *GarbDressItem) FromGarb(asset *api.SpaceBG, ownerFanNbrs map[int64]*api.UserFanInfoReply) {
	v.GarbTitle = garbTitle(asset.Name)
	v.GarbID = asset.Id
	v.TitleBgimage = asset.FanNOImage
	v.TitleColor = asset.FanNOColor
	v.SuitBgColor = asset.SuitBgColor
	if len(asset.Images) > 0 {
		v.Image = asset.Images[0].Portrait
	}
	if fansNbr, ok := ownerFanNbrs[asset.SuitItemID]; ok && fansNbr != nil && fansNbr.Number > 0 {
		v.FansNumber = fansNbrLabel(fansNbr.Number)
		v.FanBgColor = fansNbr.FanBgColor
	}
	//nolint:gosimple
	return
}

type CharacterListReq struct {
	Mid int64 `form:"-"`
	Pn  int64 `form:"pn" default:"1"`
	PS  int64 `form:"ps" default:"20"`
}

type CharacterSetReq struct {
	Mid       int64  `form:"-"`
	CostumeId string `form:"costume_id" validate:"required"`
	VersionId string `form:"version_id"`
}
