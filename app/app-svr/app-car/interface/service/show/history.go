package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/history"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

const (
	_tpArchive = 3
	_tpPGC     = 4
	_arcStr    = "archive"
	_pgcStr    = "pgc"
)

var (
	businessMap = map[int32]string{
		_tpArchive: _arcStr,
		_tpPGC:     _pgcStr,
	}
	businesToGotoMap = map[string]string{
		_arcStr: model.GotoAv,
		_pgcStr: model.GotoPGC,
	}
)

// nolint: gocognit
func (s *Service) Cursor(c context.Context, plat int8, mid int64, buvid string, param *history.HisParam) (res []cardm.Handler, page *cardm.Page, err error) {
	var (
		paramMaxBus string
		isAudio     bool
	)
	if _, ok := businessMap[param.MaxTP]; ok {
		paramMaxBus = businessMap[param.MaxTP]
	}
	businesses := []string{_arcStr, _pgcStr}
	// 过滤
	if param.FromType == model.FromAudio {
		isAudio = true
	}
	hiss, err := s.his.HistoryCursorAll(c, mid, param.Max, _max, paramMaxBus, buvid, businesses, isAudio, param.Build)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil
	}
	// 插入逻辑
	if param.Max == 0 && param.ParamStr != "" {
		his, ok := cardm.FromCursor(param.ParamStr)
		if ok {
			var isok bool
			for _, hi := range hiss {
				if hi.Oid == his.Oid && hi.Business == his.Business {
					isok = true
					break
				}
			}
			// 不相同直接插入第一位
			if !isok {
				cards := []*hisApi.ModelResource{his}
				cards = append(cards, hiss...)
				hiss = cards
			}
		}
	}
	var (
		aids  []int64
		epids []int32
		arcs  map[int64]*arcgrpc.ViewReply
		seams map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, h := range hiss {
		switch businesToGotoMap[h.Business] {
		case model.GotoAv:
			aids = append(aids, h.Oid)
		case model.GotoPGC:
			aids = append(aids, h.Oid)
			epids = append(epids, int32(h.Epid))
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Views(ctx, aids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			seams, err = s.bgm.EpCards(ctx, epids)
			if err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	items := []cardm.Handler{}
	// 兜底卡片
	backupCard := []cardm.Handler{}
	page = &cardm.Page{}
	for _, hi := range hiss {
		var (
			r        = &ai.Item{Goto: businesToGotoMap[hi.Business], ID: hi.Oid}
			op       = &operate.Card{ViewAt: hi.Unix}
			main     interface{}
			cardType model.CardType
			arc      *arcgrpc.ViewReply
		)
		op.From(model.CardGt(r.Goto), model.EntranceHistoryRecord, r.ID, plat, param.Build, param.MobiApp)
		switch param.FromType {
		case model.FromList:
			main = hi
			cardType = model.SmallCoverV3
		default:
			cardType = model.SmallCoverV4
			switch r.Goto {
			case model.GotoAv:
				// 失效过滤: aid失效;cid失效
				var ok bool
				if arc, ok = arcs[hi.Oid]; !ok {
					continue
				}
				if arc.State < 0 { // 大于等于0 前台可见
					continue
				}
				var isCidExist bool
				for _, p := range arc.Pages {
					if p.Cid == hi.Cid {
						isCidExist = true
					}
				}
				if !isCidExist {
					continue
				}
				main = arcs
				op.Cid = hi.Cid
			case model.GotoPGC:
				op.ID = hi.Epid
				op.Cid = hi.Epid
				main = seams
			}
			if arc != nil {
				for _, p := range arc.Pages {
					if p.Cid != hi.Cid {
						continue
					}
					op.Duration = p.Duration
					op.Progress = p.Duration
					break
				}
			}
			if hi.Pro > -1 {
				op.Progress = hi.Pro
			}
		}
		// 分页
		page.Max = hi.Unix
		page.MaxTP = hi.Tp
		materials := &cardm.Materials{
			ViewReplym:         arcs,
			EpisodeCardsProtom: seams,
			Prune:              cardm.CursorPrune(hi),
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
		items = append(items, h)
	}
	if len(items) == 0 {
		items = backupCard
	}
	return items, page, nil
}
