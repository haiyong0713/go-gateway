package space

import (
	"context"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	api "git.bilibili.co/bapis/bapis-go/garb/service"
	live2dgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"
)

func (s *Service) GarbDetail(c context.Context, req *space.GarbDetailReq) (result *space.GarbDetailReply, err error) {
	var (
		garbInfo   *api.SpaceBG
		ownerEquip *api.SpaceBGUserEquipReply
		ownerInfo  *accgrpc.Card
		fansNbr    int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		garbInfo, err = s.garbDao.SpaceBG(c, req.GarbID)
		return
	})
	if req.VisitMySpace() { // 主人态 是否装扮
		eg.Go(func(c context.Context) (err error) {
			currentEquip, equipErr := s.garbDao.SpaceBGEquip(c, req.Mid)
			if equipErr != nil {
				log.Error("%+v", err)
				return
			}
			ownerEquip = currentEquip
			return
		})
	}
	eg.Go(func(c context.Context) (err error) {
		ownerInfo, err = s.accDao.Card(c, req.Vmid)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if garbInfo.SuitItemID != 0 {
		if fansNbr, err = s.garbDao.UserFanInfo(c, req.Vmid, garbInfo.SuitItemID); err != nil {
			log.Warn("UserFanInfo fail err(%+v) mid(%d) suitItemId(%d)", err, req.Vmid, garbInfo.SuitItemID)
			err = nil
		}
		if fansNbr <= 0 {
			log.Warn("vmid %d, suitItemID %d, fansNumber 0", req.Vmid, garbInfo.SuitItemID)
		}
	}
	result = new(space.GarbDetailReply)
	result.FromGarb(ownerInfo, fansNbr)
	if legalImageID := result.GarbState(garbInfo, req, ownerEquip); !legalImageID {
		err = ecode.GarbImageIDIllegal
		result = nil
	}
	return
}

func (s *Service) UserGarbList(c context.Context, req *space.GarbListReq) (reply *space.GarbListReply, err error) {
	var (
		ownerEquip                 *api.SpaceBGUserEquipReply
		listReply                  *api.SpaceBGUserAssetListReply
		suitItemIDs                []int64
		ownerFanIDs, visitorFanIDs map[int64]*api.UserFanInfoReply
		//nolint:ineffassign
		isGuest = false
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		currentEquip, equipErr := s.garbDao.SpaceBGEquip(c, req.Vmid)
		if equipErr != nil {
			log.Error("%+v", err)
			return
		}
		ownerEquip = currentEquip
		return
	})
	eg.Go(func(c context.Context) (err error) {
		listReply, err = s.garbDao.SpaceBGUserAssetList(c, req.Vmid, req.Pn, req.PS)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range listReply.List {
		if v != nil && v.Item != nil {
			suitItemIDs = append(suitItemIDs, v.Item.SuitItemID)
		}
	}
	eg = errgroup.WithContext(c)
	if req.Vmid != 0 { // 主人粉丝号码信息
		eg.Go(func(c context.Context) (err error) {
			ownerFanIDs, err = s.garbDao.UserFanInfos(c, req.Vmid, suitItemIDs)
			return
		})
	}
	if isGuest = req.Mid > 0 && req.Vmid > 0 && req.Vmid != req.Mid; isGuest { // 客人购买状态
		eg.Go(func(c context.Context) (err error) {
			visitorFanIDs, err = s.garbDao.UserFanInfos(c, req.Mid, suitItemIDs)
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	reply = &space.GarbListReply{
		Count: listReply.Total,
		List:  make([]*space.GarbListItem, 0),
	}
	log.Info("UserGarbList Mid %d, VMid %d, OwnerFanIDs %v, VisitorFanIDs %v", req.Mid, req.Vmid, ownerFanIDs, visitorFanIDs)
	for _, v := range listReply.List {
		if v == nil || v.Item == nil {
			continue
		}
		item := new(space.GarbListItem)
		item.FromAsset(v.Item, ownerFanIDs, ownerEquip, req.VisitMySpace())
		if isGuest || req.Mid == 0 { // 客态判断是否购买过，出购买同款按钮；主态不出；未登录态直接出
			item.PurchaseButton(visitorFanIDs, v.Item.SuitItemID, s.c.Host.WWW)
		}
		reply.List = append(reply.List, item)
	}
	return
}

func (s *Service) GarbDress(c context.Context, req *space.GarbDressReq) (err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		return s.spcDao.TopPhotoArcCancel(c, req.Mid)
	})
	eg.Go(func(c context.Context) (err error) {
		_, err = s.garbDao.RemoveUserSpaceCharacter(c, &live2dgrpc.RemoveUserSpaceCharacterReq{Mid: req.Mid})
		return err
	})
	eg.Go(func(c context.Context) (err error) {
		err = s.digitalDao.DigitalUnbind(c, req.Mid)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("GarbDress failed req=%+v, error=%+v", req, err)
		return err
	}
	return s.garbDao.SpaceBGLoad(c, req.Mid, req.GarbID, req.ImageID)
}

func (s *Service) GarbTakeOff(c context.Context, mid int64) (err error) {
	return s.garbDao.SpaceBGUnload(c, mid)
}

func (s *Service) TopphotoReset(c context.Context, mid int64, accesskey, platform, device string, typ int64) (err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		return s.garbDao.SpaceBGUnload(c, mid)
	})
	eg.Go(func(c context.Context) (err error) {
		_, err = s.garbDao.RemoveUserSpaceCharacter(c, &live2dgrpc.RemoveUserSpaceCharacterReq{Mid: mid})
		return err
	})
	eg.Go(func(c context.Context) (err error) {
		err = s.digitalDao.DigitalUnbind(c, mid)
		return err
	})
	// 视频模式不重置头图
	//nolint:gomnd
	if typ != 2 {
		eg.Go(func(c context.Context) (err error) {
			return s.spcDao.TopphotoReset(c, accesskey, platform, device)
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("TopphotoReset Mid %d, Accesskey %s, Platform %s, Device %s, Err %v", mid, accesskey, platform, device, err)
	}
	return
}

func (s *Service) UserCharacterList(ctx context.Context, req *space.CharacterListReq) (*live2dgrpc.GetUserSpaceCharacterListResp, error) {
	return s.garbDao.GetUserSpaceCharacterList(ctx, &live2dgrpc.GetUserSpaceCharacterListReq{Mid: req.Mid, Pn: req.Pn, Ps: req.PS})
}

func (s *Service) CharacterSet(ctx context.Context, req *space.CharacterSetReq) (*live2dgrpc.SetUserSpaceCharacterResp, error) {
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) (err error) {
		return s.spcDao.TopPhotoArcCancel(c, req.Mid)
	})
	eg.Go(func(c context.Context) (err error) {
		return s.garbDao.SpaceBGUnload(c, req.Mid)
	})
	eg.Go(func(c context.Context) (err error) {
		err = s.digitalDao.DigitalUnbind(c, req.Mid)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("CharacterSet failed req=%+v, error=%+v", req, err)
		return nil, err
	}
	return s.garbDao.SetUserSpaceCharacter(ctx, &live2dgrpc.SetUserSpaceCharacterReq{Mid: req.Mid, CostumeId: req.CostumeId, VersionId: req.VersionId})
}

func (s *Service) CharacterRemove(ctx context.Context, mid int64) (*live2dgrpc.RemoveUserSpaceCharacterResp, error) {
	return s.garbDao.RemoveUserSpaceCharacter(ctx, &live2dgrpc.RemoveUserSpaceCharacterReq{Mid: mid})
}
