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

func (s *Service) Space(c context.Context, mid int64, plat int8, buvid string, param *space.SpaceParam) (*space.Info, error) {
	var (
		aids            []int64
		mine            *mine.Mine
		arcm            map[int64]*arcgrpc.ArcPlayer
		stat            *relationgrpc.StatReply
		rs              []*ai.Item
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
		rs = append(rs, &ai.Item{Goto: model.GotoAv, ID: v.Aid})
	}
	// 插入逻辑
	rs = s.listInsert(rs, param.Pn, param.ParamStr)
	for _, v := range rs {
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
			if arcm, err = s.arc.ArcsPlayerAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return info, nil
	}
	is := []cardm.Handler{}
	// 兜底卡片
	backupCard := []cardm.Handler{}
	for _, r := range rs {
		var (
			op       = &operate.Card{}
			main     interface{}
			cardType model.CardType
		)
		op.From(model.CardGt(r.Goto), model.EntranceSpace, r.ID, plat, param.Build, param.MobiApp)
		main = arcm
		materials := &cardm.Materials{
			Prune: cardm.GtPrune(r.Goto, r.ID),
		}
		if param.FromType != model.FromList {
			cardType = model.SmallCoverV4
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 找出互动视频卡片，并把第一张放入兜底卡片里面，并且长度为0的时候放入
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			if len(backupCard) == 0 {
				backupCard = append(backupCard, h)
			}
			// 互动视频不放入列表里面
			continue
		}
		is = append(is, h)
	}
	// 如果过滤完当前列表一个都没有，放入兜底卡片数据
	if len(is) == 0 {
		is = backupCard
	}
	info.Items = is
	return info, nil
}
