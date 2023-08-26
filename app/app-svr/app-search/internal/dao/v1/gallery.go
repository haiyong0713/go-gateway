package v1

import (
	"context"

	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

func (d *dao) GetNFTRegionBatch(ctx context.Context, nftID []string) (*gallerygrpc.GetNFTRegionReply, error) {
	return d.galleryClient.GetNFTRegion(ctx, &gallerygrpc.GetNFTRegionReq{NftId: nftID})
}
