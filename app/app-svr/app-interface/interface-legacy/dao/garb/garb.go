package garb

import (
	"context"

	"go-common/library/ecode"

	api "git.bilibili.co/bapis/bapis-go/garb/service"
	"github.com/pkg/errors"
)

const (
	_activeGarbGroup = "active"
	_spaceBGPart     = 7
)

// SpaceBG picks the garb detail with the ID
func (d *Dao) SpaceBG(c context.Context, garbID int64) (result *api.SpaceBG, err error) {
	var reply *api.SpaceBGReply
	if reply, err = d.garbClient.SpaceBG(c, &api.SpaceBGReq{ItemID: garbID}); err != nil {
		err = errors.Wrapf(err, "itemID %d", garbID)
		return
	}
	if reply == nil || reply.Item == nil {
		err = ecode.NothingFound
		err = errors.Wrapf(err, "itemID %d", garbID)
		return
	}
	result = reply.Item
	return
}

// UserFanInfo
func (d *Dao) UserFanInfo(c context.Context, mid, suitItemID int64) (number int64, err error) {
	var reply *api.UserFanInfoReply
	if reply, err = d.garbClient.UserFanInfo(c, &api.UserFanInfoReq{
		Mid:        mid,
		SuitItemID: suitItemID, // 是套装中的空间头图的id
	}); err != nil {
		err = errors.Wrapf(err, "suitItemID %d, mid %d", suitItemID, mid)
		return
	}
	if reply == nil || reply.Number == 0 {
		err = ecode.NothingFound
		err = errors.Wrapf(err, "suitItemID %d, mid %d", suitItemID, mid)
		return
	}
	number = reply.Number
	return
}

// SpaceBGEquip def
func (d *Dao) SpaceBGEquip(c context.Context, mid int64) (reply *api.SpaceBGUserEquipReply, err error) {
	if reply, err = d.garbClient.SpaceBGUserEquip(c, &api.SpaceBGUserEquipReq{
		Mid: mid,
	}); err != nil {
		err = errors.Wrapf(err, "mid %d", mid)
		return
	}
	if reply == nil || reply.Item == nil {
		err = ecode.NothingFound
		err = errors.Wrapf(err, "mid %d", mid)
	}
	return
}

// UserFanInfos def
func (d *Dao) UserFanInfos(c context.Context, mid int64, suitItemIDs []int64) (result map[int64]*api.UserFanInfoReply, err error) {
	var (
		reply *api.UserFanInfoListReply
		req   = &api.UserFanInfoListReq{
			Mid:         mid,
			SuitItemIDs: suitItemIDs,
		}
	)
	if reply, err = d.garbClient.UserFanInfoList(c, req); err != nil {
		err = errors.Wrapf(err, "mid %d, suitItemIDs %v", mid, suitItemIDs)
		return
	}
	if reply == nil {
		err = errors.Wrapf(ecode.NothingFound, "mid %d, suitItemIDs %v", mid, suitItemIDs)
		return
	}
	result = reply.List
	return
}

// SpaceBGUserAssetList def
func (d *Dao) SpaceBGUserAssetList(c context.Context, mid, pn, ps int64) (reply *api.SpaceBGUserAssetListReply, err error) {
	if reply, err = d.garbClient.SpaceBGUserAssetList(c, &api.SpaceBGUserAssetListReq{
		Mid:   mid,
		Group: _activeGarbGroup,
		Pn:    pn,
		Ps:    ps,
	}); err != nil {
		err = errors.Wrapf(err, "mid %d, pn %d ps %d", mid, pn, ps)
		return
	}
	if reply == nil {
		err = ecode.NothingFound
	}
	return
}

// SpaceBGUserAssetListWithFan def
func (d *Dao) SpaceBGUserAssetListWithFan(c context.Context, mid, pn, ps int64) (*api.SpaceBGUserAssetListReply, error) {
	reply, err := d.garbClient.SpaceBGUserAssetListWithFan(c, &api.SpaceBGUserAssetListReq{
		Mid:   mid,
		Group: _activeGarbGroup,
		Pn:    pn,
		Ps:    ps,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "mid %d, pn %d ps %d", mid, pn, ps)
	}
	if reply == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "mid %d, pn %d ps %d", mid, pn, ps)
	}
	return reply, nil
}

// SpaceBGUnload def
func (d *Dao) SpaceBGUnload(c context.Context, mid int64) (err error) {
	if _, err = d.garbClient.UserUnloadEquip(c, &api.UserUnloadEquipReq{
		Mid:    mid,
		PartID: _spaceBGPart,
	}); err != nil {
		err = errors.Wrapf(err, "mid %d", mid)
	}
	return
}

// SpaceBGLoad def.
func (d *Dao) SpaceBGLoad(c context.Context, mid, suitItemID, index int64) (err error) {
	if _, err = d.garbClient.UserLoadEquip(c, &api.UserLoadEquipReq{
		Mid:    mid,
		ItemID: suitItemID,
		Index:  index,
	}); err != nil {
		err = errors.Wrapf(err, "mid %d suitItemID %d, index %d", mid, suitItemID, index)
	}
	return
}

// UserAsset .
func (d *Dao) UserAsset(c context.Context, itemID, mid int64) (*api.UserAssetReply, error) {
	return d.garbClient.UserAsset(c, &api.UserAssetReq{PartID: 1, Mid: mid, ItemID: itemID})
}
