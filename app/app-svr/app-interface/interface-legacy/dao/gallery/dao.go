package gallery

import (
	"context"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"go-common/library/ecode"

	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

type Dao struct {
	galleryClient gallerygrpc.GalleryServiceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.galleryClient, err = gallerygrpc.NewClient(c.GalleryGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) SpaceHasNFT(ctx context.Context, req *gallerygrpc.SpaceGetNFTReq) (*gallerygrpc.SpaceGetNFTReply, error) {
	return d.galleryClient.SpaceGetNFT(ctx, req)
}

func (d *Dao) IsNFTFaceOwner(ctx context.Context, req *gallerygrpc.MidReq) (*gallerygrpc.OwnerReply, error) {
	return d.galleryClient.IsNFTFaceOwner(ctx, req)
}

func (d *Dao) GetNFTRegionBatch(ctx context.Context, nftID []string) (*gallerygrpc.GetNFTRegionReply, error) {
	return d.galleryClient.GetNFTRegion(ctx, &gallerygrpc.GetNFTRegionReq{NftId: nftID})
}

func (d *Dao) GetNFTRegion(ctx context.Context, nftID string) (*gallerygrpc.NFTRegion, error) {
	reply, err := d.galleryClient.GetNFTRegion(ctx, &gallerygrpc.GetNFTRegionReq{
		NftId: []string{nftID},
	})
	if err != nil {
		return nil, err
	}
	region, ok := reply.Region[nftID]
	if !ok {
		return nil, ecode.NothingFound
	}
	return region, nil
}

func (d *Dao) OwnNFTStatus(ctx context.Context, mid int64, mobiApp string) (*gallerygrpc.OwnNFTFaceStatusReply, error) {
	return d.galleryClient.OwnNFTFaceStatus(ctx, &gallerygrpc.OwnNFTFaceStatusReq{
		Mid:     mid,
		MobiApp: mobiApp,
	})
}
