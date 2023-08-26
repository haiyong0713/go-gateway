package space

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/digital"

	digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"
	live2dgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"
)

func (s *Service) DigitalEntry(ctx context.Context, mid int64) (*digitalgrpc.GetGarbSpaceEntryResp, error) {
	return s.digitalDao.DigitalEntry(ctx, mid)
}

func (s *Service) DigitalInfo(ctx context.Context, mid, vmid int64, nftID string) (*digital.SpaceDigitalInfoResp, error) {
	reply, err := s.digitalDao.DigitalInfo(ctx, mid, vmid, nftID)
	if err != nil {
		log.Error("s.digitalDao.DigitalInfo vmid:%d, mid:%d, nftID:%s, err:%+v", vmid, mid, nftID, err)
		return nil, err
	}
	res := &digital.SpaceDigitalInfoResp{}
	res.FromGetGarbSpaceInfoResp(reply)
	return res, nil
}

func (s *Service) DigitalBind(ctx context.Context, mid, itemID int64, nftID string) error {
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) (err error) {
		return s.spcDao.TopPhotoArcCancel(c, mid)
	})
	eg.Go(func(c context.Context) (err error) {
		_, err = s.garbDao.RemoveUserSpaceCharacter(c, &live2dgrpc.RemoveUserSpaceCharacterReq{Mid: mid})
		return err
	})
	eg.Go(func(c context.Context) (err error) {
		return s.garbDao.SpaceBGUnload(c, mid)
	})
	if err := eg.Wait(); err != nil {
		log.Error("DigitalBind failed mid=%+v, error=%+v", mid, err)
		return err
	}
	return s.digitalDao.DigitalBind(ctx, mid, itemID, nftID)
}

func (s *Service) DigitalUnbind(ctx context.Context, mid int64) error {
	return s.digitalDao.DigitalUnbind(ctx, mid)
}

func (s *Service) DigitalExtraInfo(ctx context.Context, mid int64, nftID string) (*digital.SpaceDigitalExtraInfoResp, error) {
	reply, err := s.digitalDao.DigitalExtraInfo(ctx, mid, nftID)
	if err != nil {
		log.Error("s.digitalDao.DigitalExtraInfo mid:%d, nftID:%s, err:%+v", mid, nftID, err)
		return nil, err
	}
	res := &digital.SpaceDigitalExtraInfoResp{}
	res.SpaceExtraInfoResp(reply)
	return res, nil
}
