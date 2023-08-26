package feed

import (
	"context"

	"go-common/library/log"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

var (
	_auditBanners = []*banner.Banner{
		{
			Title: "充电",
			Image: "https://i0.hdslb.com/bfs/archive/9ce8f6cdf76e6cbd50ce7db76262d5a35e594c79.png",
			Hash:  "3c4990d06c46de0080e3821fca6bedca",
			URI:   "bilibili://video/813060",
		},
	}
	_auditIPhoneBanners = []*banner.Banner{
		{
			Title: "充电",
			Image: "https://i0.hdslb.com/bfs/archive/9ce8f6cdf76e6cbd50ce7db76262d5a35e594c79.png",
			Hash:  "3c4990d06c46de0080e3821fca6bedca",
			URI:   "bilibili://video/813060",
		},
		{
			Title: "零基础小白到大神，全实战案例教学",
			Image: "https://i0.hdslb.com/bfs/archive/0654df00324d503e1f1e5017256987aa8f3ca38d.png",
			Hash:  "510d6cd94e6a503015f7e34158239e8b",
			URI:   "https://m.bilibili.com/cheese/play/ss21",
		},
	}
	// av2314237 已经删除，后续看情况处理
	// 已删除失效id已更换新的稿件id
	_aids = []int64{308040, 2431658, 2432648, 2427553, 539600, 1968681, 850424, 887861, 1960912, 1935680, 1406019,
		1985297, 1977493, 2312184, 2316891, 864845, 1986932, 880857, 875624, 744299}
)

// Audit check audit plat then return audit data.
func (s *Service) Audit(c context.Context, mid int64, mobiApp, device string, plat int8, build int) (is []*feed.Item, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditData(c, mid, mobiApp, device), true
		}
	}
	return
}

// Audit2 check audit plat and ip, then return audit data.
func (s *Service) Audit2(c context.Context, mid int64, mobiApp, device string, plat int8, build int, column cdm.ColumnStatus) (is []card.Handler, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditData2(c, mid, mobiApp, device, plat, build, column), true
		}
	}
	return
}

// auditData some data for audit.
func (s *Service) auditData(c context.Context, mid int64, mobiApp, device string) (is []*feed.Item) {
	i := &feed.Item{}
	i.FromBanner(_auditBanners, "")
	is = append(is, i)
	am, err := s.arc.Archives(c, _aids, mid, mobiApp, device)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	for _, aid := range _aids {
		if a, ok := am[aid]; ok {
			i := &feed.Item{}
			i.FromAv(a)
			is = append(is, i)
		}
	}
	return
}

// auditData2 some data for audit.
func (s *Service) auditData2(c context.Context, mid int64, mobiApp, device string, plat int8, build int, column cdm.ColumnStatus) (is []card.Handler) {
	i := card.Handle(plat, model.GotoBanner, "", column, nil, nil, nil, nil, nil, nil, nil)
	if i != nil {
		op := &operate.Card{}
		if plat == model.PlatIPhone && build == 8860 {
			op.FromBanner(_auditIPhoneBanners, "")
		} else {
			op.FromBanner(_auditBanners, "")
		}
		if err := i.From(nil, op); err != nil {
			log.Error("Failed to From: %+v", err)
		}
		is = append(is, i)
	}
	am, err := s.arc.Archives(c, _aids, mid, mobiApp, device)
	if err != nil {
		log.Error("%+v", err)
	}
	var main interface{}
	for _, aid := range _aids {
		if a, ok := am[aid]; ok {
			i := card.Handle(plat, model.GotoAv, "", column, nil, nil, nil, nil, nil, nil, nil)
			if i == nil {
				continue
			}
			op := &operate.Card{}
			op.From(cdm.CardGotoAv, aid, 0, 0, 0, "")
			main = map[int64]*arcgrpc.ArcPlayer{a.Aid: {Arc: a, DefaultPlayerCid: a.FirstCid}}
			if err := i.From(main, op); err != nil {
				log.Error("Failed to From: %+v", err)
			}
			if !i.Get().Right {
				continue
			}
			if model.IsPad(plat) {
				// ipad卡片不展示标签
				i.Get().DescButton = nil
			}
			is = append(is, i)
		}
	}
	return
}

func (s *Service) loadAuditCache() {
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	as, err := s.rsc.AppAudit(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.auditCache = as
}
