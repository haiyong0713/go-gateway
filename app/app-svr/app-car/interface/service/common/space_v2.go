package common

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	sch "go-gateway/app/app-svr/app-car/interface/model/search"
	"go-gateway/app/app-svr/app-car/interface/model/space"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

func (s *Service) SpaceV2(ctx context.Context, param *space.SpaceParamV2) (resp *space.SpaceRespV2, err error) {
	var (
		arcItems  []*common.Item
		pageReq   *sch.PageInfo
		upAccount *space.AccountInfo
		arcIdsRes *space.ArcIdsRes
		matReq    *common.Params
		matResp   *common.CarContext
	)
	pageReq, err = extractPageInfo(param.Ps, param.PageNext)
	if err != nil {
		log.Error("SpaceV2 extractSpacePage err:%+v, param:%+v", err, param)
		return nil, err
	}
	eg := errgroup.WithContext(ctx)
	// 1.1 获取up主账户信息（不含稿件总数）
	eg.Go(func(ctx context.Context) error {
		var localErr error
		upAccount, localErr = s.getUpAccount(ctx, param.UpMid, param.Mid)
		if localErr != nil {
			return localErr
		}
		return nil
	})
	// 1.2 获取up主稿件aid
	eg.Go(func(ctx context.Context) error {
		var localErr error
		arcIdsRes, localErr = s.getUpArcIds(ctx, param.UpMid, pageReq)
		if localErr != nil {
			return localErr
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("SpaceV2 get raw data error:%+v, param:%+v", err, param)
		return nil, err
	}
	// 稿件总数赋值
	if upAccount != nil && arcIdsRes != nil {
		upAccount.VideoCount = arcIdsRes.Total
	}
	// 2 依据id列表获取原始物料
	matReq = generateMatReq(arcIdsRes)
	matResp, err = s.material(ctx, matReq, param.DeviceInfo)
	if err != nil {
		log.Error("SpaceV2 s.material error:%+v, param:%+v", err, param)
		return nil, err
	}
	// 3 依据原始物料获取item
	arcItems = s.getSpaceArcItems(matResp, arcIdsRes, param.DeviceInfo)
	return &space.SpaceRespV2{
		Account:  upAccount,
		ArcItems: arcItems,
		PageNext: arcIdsRes.PageNext,
		HasNext:  arcIdsRes.HasNext,
	}, nil
}

func generateMatReq(arcIdsRes *space.ArcIdsRes) *common.Params {
	if arcIdsRes == nil || len(arcIdsRes.Aids) == 0 {
		return new(common.Params)
	}
	avs := make([]*archivegrpc.PlayAv, 0)
	for _, v := range arcIdsRes.Aids {
		avs = append(avs, &archivegrpc.PlayAv{Aid: v})
	}
	return &common.Params{
		ArchiveReq: &common.ArchiveReq{
			PlayAvs: avs,
		},
	}
}

func (s *Service) getUpAccount(ctx context.Context, upMid int64, mid int64) (*space.AccountInfo, error) {
	var (
		profile   *accountgrpc.Profile
		stat      *relationgrpc.StatReply
		relations map[int64]*relationgrpc.InterrelationReply
	)
	eg := errgroup.WithContext(ctx)
	// up主基本信息（包含大会员数据）
	eg.Go(func(ctx context.Context) (err error) {
		if profile, err = s.accountDao.Profile3(ctx, upMid); err != nil {
			return errors.Wrap(err, "s.accountDao.Profile3 error")
		}
		return nil
	})
	// up主粉丝数
	eg.Go(func(ctx context.Context) (err error) {
		if stat, err = s.relationDao.StatGRPC(ctx, upMid); err != nil {
			return errors.Wrap(err, "s.relationDao.StatGRPC error")
		}
		return nil
	})
	// 用户与up主的关系
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if relations, err = s.relationDao.RelationsInterrelations(ctx, mid, []int64{upMid}); err != nil {
				return errors.Wrap(err, "s.relationDao.RelationsInterrelations error")
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &space.AccountInfo{
		Mid:       upMid,
		Name:      profile.Name,
		Face:      profile.Face,
		FansCount: stat.GetFollower(),
		Relation:  model.RelationChange(upMid, relations),
		VipInfo:   &profile.Vip,
	}, nil

}

func (s *Service) getUpArcIds(ctx context.Context, upMid int64, page *sch.PageInfo) (res *space.ArcIdsRes, err error) {
	// up主稿件+稿件总数
	var (
		upArcs []*upgrpc.Arc
		aids   []int64
		total  int64
	)
	upArcs, total, err = s.upDao.UpArcs(ctx, upMid, int64(page.Pn), int64(page.Ps))
	if err != nil {
		// 兼容无投稿的up主
		if ecode.Cause(err) != ecode.NothingFound {
			return nil, err
		}
		upArcs = make([]*upgrpc.Arc, 0)
		total = 0
	}
	aids = make([]int64, 0)
	for _, v := range upArcs {
		aids = append(aids, v.Aid)
	}

	hasNext := int64(page.Ps*page.Pn) < total
	page.Pn = page.Pn + 1
	return &space.ArcIdsRes{
		Aids:     aids,
		Total:    total,
		PageNext: page,
		HasNext:  hasNext,
	}, nil
}

func (s *Service) getSpaceArcItems(carCtx *common.CarContext, arcs *space.ArcIdsRes, dev model.DeviceInfo) []*common.Item {
	items := make([]*common.Item, 0)
	if len(arcs.Aids) == 0 {
		return items
	}
	for _, aid := range arcs.Aids {
		carCtx.OriginData = &common.OriginData{
			MaterialType: common.MaterialTypeUGC,
			Oid:          aid,
		}
		items = append(items, s.formItem(carCtx, dev))
	}
	return items
}
