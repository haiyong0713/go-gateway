package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/region"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

const customRegionPs = 50

func customModulePage(pn int, aids []int64) []int64 {
	if pn < 1 {
		pn = 1
	}
	start := (pn - 1) * customRegionPs
	if start >= len(aids) {
		return nil
	}
	end := start + customRegionPs
	if end > len(aids) {
		end = len(aids)
	}
	return aids[start:end]
}

func (s *Service) RegionList(c context.Context, plat int8, param *region.RegionParam) ([]cardm.Handler, error) { //nolint: gocognit
	var list []*api.Arc
	var err error
	if param.Rid == model.CustomModuleRid51 || param.Rid == model.CustomModuleRid61Childhood || param.Rid == model.CustomModuleRid61Eden || param.Rid == model.CustomModuleRidDW {
		var aids []int64
		switch param.Rid {
		case model.CustomModuleRid51:
			aids = customModulePage(param.Pn, s.c.CustomModule51.ChannelAids[param.Channel])
		case model.CustomModuleRid61Childhood:
			aids = customModulePage(param.Pn, s.c.CustomModule61Childhood.ChannelAids[param.Channel])
		case model.CustomModuleRid61Eden:
			aids = customModulePage(param.Pn, s.c.CustomModule61Eden.ChannelAids[param.Channel])
		case model.CustomModuleRidDW:
			aids = customModulePage(param.Pn, s.c.CustomModuleDW.ChannelAids[param.Channel])
		default:
			// nop
		}
		if len(aids) == 0 {
			return []cardm.Handler{}, nil
		}
		arcMap, arcErr := s.arc.Archives(c, aids)
		if arcErr != nil {
			log.Error("RegionList s.arc.Archives err=%+v, aids=%+v", err, aids)
			return []cardm.Handler{}, nil
		}
		list = make([]*api.Arc, 0)
		for _, aid := range aids {
			v := arcMap[aid]
			if v != nil {
				list = append(list, v)
			}
		}
	} else {
		list, err = s.reg.RegionDynamic(c, param.Rid, param.Pn, param.Ps)
	}
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}

	var (
		topAids []int64
	)
	// 插入逻辑
	if param.Rid != model.CustomModuleRid51 && param.Rid != model.CustomModuleRid61Childhood &&
		param.Rid != model.CustomModuleRid61Eden && param.Rid != model.CustomModuleRidDW &&
		param.Pn == 1 && param.ParamStr != "" {
		_, id, _, ok := cardm.FromGtPrune(param.ParamStr)
		if ok {
			var isok bool
			for _, v := range list {
				if v.Aid == id {
					isok = true
					break
				}
			}
			if !isok {
				cards := []*api.Arc{{Aid: id}}
				cards = append(cards, list...)
				list = cards
				topAids = append(topAids, id)
			}
		}
	}
	var (
		aids     []int64
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	arcs := map[int64]*arcgrpc.Arc{}
	for _, v := range list {
		arcs[v.Aid] = v
		if v.Aid == 0 {
			continue
		}
		aids = append(aids, v.Aid)
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seamAids, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(topAids) > 0 {
		group.Go(func(ctx context.Context) error {
			// 只处理顶部插入卡片的数据，如果接口返回失败直接抛弃
			topArcs, err := s.arc.Archives(ctx, topAids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			for k, v := range topArcs {
				arcs[k] = v
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []cardm.Handler{}
	for _, v := range list {
		var (
			r        = &ai.Item{Goto: model.GotoAv, ID: v.Aid}
			op       = &operate.Card{Rid: param.Rid}
			main     interface{}
			cardType model.CardType
		)
		op.From(model.CardGt(r.Goto), model.EntranceRegion, r.ID, plat, param.Build, param.MobiApp)
		switch param.FromType {
		case model.FromList:
			cardType = model.SmallCoverV1
		default:
			cardType = model.SmallCoverV4
		}
		// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
		main = arcs
		if _, ok := seamAids[int32(r.ID)]; ok {
			main = seamAids
			r.Goto = model.GotoPGC
		}
		materials := &cardm.Materials{
			Prune: cardm.GtPrune(r.Goto, r.ID),
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 过滤互动视频
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			continue
		}
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}, nil
	}
	return is, nil
}
