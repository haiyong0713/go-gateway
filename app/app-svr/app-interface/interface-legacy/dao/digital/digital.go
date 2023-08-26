package digital

import (
	"context"

	digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"
)

func (d *Dao) DigitalEntry(ctx context.Context, mid int64) (*digitalgrpc.GetGarbSpaceEntryResp, error) {
	req := &digitalgrpc.GetGarbSpaceEntryReq{Mid: mid}
	return d.digitalClient.GetGarbSpaceEntry(ctx, req)
}

func (d *Dao) DigitalInfo(ctx context.Context, mid, vmid int64, nftID string) (*digitalgrpc.GetGarbSpaceInfoResp, error) {
	req := &digitalgrpc.GetGarbSpaceInfoReq{Mid: mid, Vmid: vmid, NftId: nftID}
	return d.digitalClient.GetGarbSpaceInfo(ctx, req)
}

func (d *Dao) DigitalBind(ctx context.Context, mid, itemID int64, nftID string) error {
	req := &digitalgrpc.SetGarbNFTSpaceReq{Mid: mid, ItemId: itemID, NftId: nftID}
	_, err := d.digitalClient.SetGarbNFTSpace(ctx, req)
	return err
}

func (d *Dao) DigitalUnbind(ctx context.Context, mid int64) error {
	req := &digitalgrpc.RemoveGarbNFTSpaceReq{Mid: mid}
	_, err := d.digitalClient.RemoveGarbNFTSpace(ctx, req)
	return err
}

func (d *Dao) DigitalExtraInfo(ctx context.Context, mid int64, nftID string) (*digitalgrpc.SpaceExtraInfoResp, error) {
	req := &digitalgrpc.SpaceExtraInfoReq{Mid: mid, NftId: nftID}
	return d.digitalClient.SpaceExtraInfo(ctx, req)
}
