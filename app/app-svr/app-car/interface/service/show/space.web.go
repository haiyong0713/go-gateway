package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/mine"
	"go-gateway/app/app-svr/app-car/interface/model/space"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

func (s *Service) SpaceWeb(c context.Context, mid int64, plat int8, buvid string, param *space.SpaceParam) (*space.Info, error) {
	var (
		aids            []int64
		mine            *mine.Mine
		arcm            map[int64]*arcgrpc.Arc
		stat            *relationgrpc.StatReply
		cardItem        []*ai.Item
		authorRelations map[int64]*relationgrpc.InterrelationReply
		uparcs          []*upgrpc.Arc
		count           int64
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		uparcs, count, err = s.up.UpArcs(ctx, param.Vmid, int64(param.Pn), int64(param.Ps))
		if err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) (err error) {
		if mine, err = s.userInfo(ctx, param.Vmid); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) (err error) {
		if stat, err = s.reldao.StatGRPC(ctx, param.Vmid); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if mid > 0 && mid != param.Vmid {
		group.Go(func(ctx context.Context) (err error) {
			if authorRelations, err = s.reldao.RelationsInterrelations(ctx, mid, []int64{param.Vmid}); err != nil {
				log.Error("s.accd.Relations2 error(%v)", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return &space.Info{Items: []cardm.Handler{}}, nil
	}
	info := &space.Info{Items: []cardm.Handler{}}
	info.FromSpace(mine, count, stat.GetFollower(), mid, authorRelations)
	for _, v := range uparcs {
		if v.Aid == 0 {
			continue
		}
		aids = append(aids, v.Aid)
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: v.Aid})
	}
	// 插入逻辑
	cardItem = s.listInsert(cardItem, param.Pn, param.ParamStr)
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		}
	}
	group = errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcm, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return info, nil
	}
	materials := &cardm.Materials{
		Arcs: arcm,
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: param.FromType,
	}
	op := &operate.Card{}
	cardType := model.SmallCoverV1
	if param.FromType != model.FromList {
		cardType = model.SmallCoverV4
	}
	list := s.cardDealWebItem(cardParam, cardItem, model.EntranceSpace, cardType, materials, op)
	info.Items = list
	return info, nil
}
