package show

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	operate "go-gateway/app/app-svr/app-card/interface/model/card/operate"
	cardapi "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	cardmv2 "go-gateway/app/app-svr/app-card/interface/model/card/v2"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/banner"
	"go-gateway/app/app-svr/app-show/interface/model/feed"
	"go-gateway/app/app-svr/app-show/interface/model/show"
)

var (
	_auditBanner = &banner.Banner{
		Title: "充电",
		Image: "http://i0.hdslb.com/bfs/archive/9ce8f6cdf76e6cbd50ce7db76262d5a35e594c79.png",
		Hash:  "3c4990d06c46de0080e3821fca6bedca",
		URI:   "bilibili://video/813060",
	}
	_auditRids = map[int8]map[string]struct{}{
		model.PlatIPhone: {
			"13":  {},
			"167": {},
			"177": {},
			"23":  {},
			"11":  {},
		},
	}
)

// GetAudit check audit plat and ip, then return audit data.
func (s *Service) Audit(c context.Context, mobiApp string, plat int8, build int, mid int64, device string) (ss []*show.Show, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditData(c, plat, mid, mobiApp, device), true
		}
	}
	return nil, false
}

func (s *Service) AuditChild(c context.Context, mobiApp string, plat int8, build int, mid int64, device string) (res []*show.Item, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			res = s.auditList(c, mid, mobiApp, device)
			return res, true
		}
	}
	return nil, false
}

// AuditFeed check audit plat and ip, then return audit data.
func (s *Service) AuditFeed(c context.Context, mobiApp string, plat int8, build int, mid int64, device string) (res []*feed.Item, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditFeed(c, mid, mobiApp, device), true
		}
	}
	return nil, false
}

// AuditFeed2 check audit plat and ip, then return audit data.
func (s *Service) AuditFeed2(c context.Context, mobiApp string, plat int8, build int, mid int64, device string) (res []cardm.Handler, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditFeed2(c, plat, mid, mobiApp, device), true
		}
	}
	return nil, false
}

// AuditFeed3 check audit plat and ip, then return audit data.
func (s *Service) AuditFeed3(c context.Context, mobiApp string, plat int8, build int, mid int64, device string) (res []*cardapi.Card, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditFeed3(c, plat, mid, mobiApp, device), true
		}
	}
	return nil, false
}

// Audit region data list.
func (s *Service) auditRegion(mobiApp string, plat int8, build int, rid string) (isAudit bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			if params, ok := _auditRids[plat]; ok {
				if _, ok = params[rid]; ok {
					return true
				}
			}
		}
	}
	return false
}

func (s *Service) loadAuditCache() {
	as, err := s.adt.Audits(context.TODO())
	if err != nil {
		log.Error("s.adt.Audits error(%v)", err)
		return
	}
	s.auditCache = as
}

// auditData some data for audit.
func (s *Service) auditData(c context.Context, p int8, mid int64, mobiApp, device string) (ss []*show.Show) {
	ss = []*show.Show{
		{
			Head: &show.Head{
				Param: "",
				Type:  "recommend",
				Style: "medium",
				Title: "热门推荐",
			},
		},
		{
			Head: &show.Head{
				Param: "3",
				Type:  "region",
				Style: "medium",
				Title: "音乐区",
			},
		},
		{
			Head: &show.Head{
				Param: "129",
				Type:  "region",
				Style: "medium",
				Title: "舞蹈区",
			},
		},
		{
			Head: &show.Head{
				Param: "4",
				Type:  "region",
				Style: "medium",
				Title: "游戏区",
			},
		},
		{
			Head: &show.Head{
				Param: "36",
				Type:  "region",
				Style: "medium",
				Title: "游戏区",
			},
		},
	}
	aids := []int64{308040, 2431658, 2432648, 2427553, 539600, 1968681, 850424, 887861, 1960912, 1935680, 1406019, 1985297, 1977493, 2312184, 2316891, 864845, 1986932, 2314237, 880857, 875624}
	n := 4
	if p == model.PlatIPad {
		aids = []int64{2455179, 2473608, 1711253, 2476389, 308040, 360940, 482844, 221107, 539600, 1968681, 850424, 887861, 936016, 1773160, 886841, 1958897, 1960912, 1935680,
			1406019, 1985297, 1635344, 572952, 2316655, 2317928, 1977493, 2312184, 2316891, 864845, 2313588, 875076, 2312249, 842756, 1986932, 2314237, 880857, 875624}
		n = 8
		// ss[0].Head.Type = ""
		// banner
		ss[0].Banner = map[string][]*banner.Banner{
			"top": {_auditBanner, _auditBanner},
		}
	} else if p == model.PlatIPhone {
		aids = []int64{308040, 2431658, 2432648, 2427553, 2455179, 2473608, 539600, 1968681, 850424, 887861, 1960912, 1935680, 1406019, 1985297, 1977493, 2312184, 2316891, 864845,
			1986932, 2314237, 880857, 875624}
		n = 6
		// banner
		ss[0].Banner = map[string][]*banner.Banner{
			"top": {_auditBanner},
		}
	}
	as, err := s.arc.ArchivesPB(c, aids, mid, mobiApp, device)
	if err != nil {
		log.Error("s.arc.ArchivesPB error(%v)", err)
		as = map[int64]*api.Arc{}
	}
	for i, aid := range aids {
		if aid == 0 {
			continue
		}
		item := &show.Item{}
		item.Goto = model.GotoAv
		item.Param = strconv.FormatInt(aid, 10)
		item.URI = model.FillURI(item.Goto, item.Param, nil)
		if a, ok := as[aid]; ok {
			item.Title = a.Title
			item.Cover = a.Pic
			item.Play = int(a.Stat.View)
			item.Danmaku = int(a.Stat.Danmaku)
			item.URI = model.AvHandler(a)(item.URI)
		}
		ss[i/n].Body = append(ss[i/n].Body, item)
	}
	return
}

func (s *Service) auditList(c context.Context, mid int64, mobiApp, device string) (ss []*show.Item) {
	aids := []int64{308040, 2431658, 2432648, 2427553, 2455179, 2473608, 539600, 1968681, 850424, 887861, 1960912, 1935680, 1406019, 1985297, 1977493, 2312184, 2316891, 864845,
		1986932, 2314237, 880857, 875624}
	as, err := s.arc.ArchivesPB(c, aids, mid, mobiApp, device)
	if err != nil {
		log.Error("s.arc.ArchivesPB error(%v)", err)
		as = map[int64]*api.Arc{}
	}
	for _, aid := range aids {
		if aid == 0 {
			continue
		}
		item := &show.Item{}
		item.Goto = model.GotoAv
		item.Param = strconv.FormatInt(aid, 10)
		item.URI = model.FillURI(item.Goto, item.Param, nil)
		if a, ok := as[aid]; ok {
			item.Title = a.Title
			item.Cover = a.Pic
			item.Play = int(a.Stat.View)
			item.Danmaku = int(a.Stat.Danmaku)
			item.URI = model.AvHandler(a)(item.URI)
		}
		ss = append(ss, item)
	}
	return
}

func (s *Service) auditFeed(c context.Context, mid int64, mobiApp, device string) (res []*feed.Item) {
	var (
		aids = []int64{2455179, 2473608, 1711253, 2476389, 308040, 360940, 482844, 221107, 539600, 1968681, 850424, 887861, 936016, 1773160, 886841, 1958897, 1960912, 1935680,
			1406019, 1985297, 1635344, 572952, 2316655, 2317928, 1977493, 2312184, 2316891, 864845, 2313588, 875076, 2312249, 842756, 1986932, 2314237, 880857, 875624}
		as  map[int64]*api.Arc
		err error
	)
	if as, err = s.arc.ArchivesPB(c, aids, mid, mobiApp, device); err != nil {
		log.Error("hottab s.arc.ArchivesPB aids(%v) error(%v)", aids, err)
		return
	}
	if len(as) == 0 {
		log.Warn("hottab s.arc.ArchivesPB(%v) length is 0", aids)
		return
	}
	for i, aid := range aids {
		item := &feed.Item{}
		item.Idx = int64(i + 1)
		item.Pos = i + 1
		if aid == 0 {
			continue
		}
		if a, ok := as[aid]; ok {
			item.FromPlayerAv(a, "")
			// if tag, ok := s.hotArcTag[a.Aid]; ok {
			// 	item.Tag = &feed.Tag{TagID: tag.ID, TagName: tag.Name}
			// }
			item.Goto = model.GotoAv
			res = append(res, item)
		}
	}
	if len(res) == 0 {
		res = _emptyList
		return
	}
	return
}

func (s *Service) auditFeed2(c context.Context, plat int8, mid int64, mobiApp, device string) (res []cardm.Handler) {
	var (
		aids = []int64{2455179, 2473608, 1711253, 2476389, 308040, 360940, 482844, 221107, 539600, 1968681, 850424, 887861, 936016, 1773160, 886841, 1958897, 1960912, 1935680,
			1406019, 1985297, 1635344, 572952, 2316655, 2317928, 1977493, 2312184, 2316891, 864845, 2313588, 875076, 2312249, 842756, 1986932, 2314237, 880857, 875624}
		as  map[int64]*api.Arc
		err error
	)
	if as, err = s.arc.ArchivesPB(c, aids, mid, mobiApp, device); err != nil {
		log.Error("hottab s.arc.ArchivesPB aids(%v) error(%v)", aids, err)
		return
	}
	if len(as) == 0 {
		log.Warn("hottab s.arc.ArchivesPB(%v) length is 0", aids)
		return
	}
	innerArc := s.controld.CircleReqInternalAttr(c, aids)
	for i, aid := range aids {
		if aid == 0 {
			continue
		}
		var (
			r    = &ai.Item{Goto: model.GotoAv, ID: aid}
			h    = cardm.Handle(plat, cdm.CardGt(r.Goto), "", 1, r, nil, nil, nil, nil, nil, nil)
			main interface{}
		)
		if h == nil {
			continue
		}
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, model.PlatIPhone, 0, "")
		var isOverseas bool
		if innerArc != nil && innerArc[aid] != nil {
			isOverseas = innerArc[aid].OverSeaBlock
		}
		if a, ok := as[aid]; ok && !isOverseas {
			main = map[int64]*arcgrpc.ArcPlayer{a.Aid: {Arc: a}}
		}
		if main != nil {
			_ = h.From(main, op)
		}
		h.Get().Idx = int64(i + 1)
		if h.Get().Right {
			res = append(res, h)
		}
	}
	if len(res) == 0 {
		res = _emptyList2
		return
	}
	return
}

func (s *Service) auditFeed3(c context.Context, plat int8, mid int64, mobiApp, device string) (res []*cardapi.Card) {
	var (
		aids = []int64{2455179, 2473608, 1711253, 2476389, 308040, 360940, 482844, 221107, 539600, 1968681, 850424, 887861, 936016, 1773160, 886841, 1958897, 1960912, 1935680,
			1406019, 1985297, 1635344, 572952, 2316655, 2317928, 1977493, 2312184, 2316891, 864845, 2313588, 875076, 2312249, 842756, 1986932, 2314237, 880857, 875624}
		as  map[int64]*api.Arc
		err error
	)
	if as, err = s.arc.ArchivesPB(c, aids, mid, mobiApp, device); err != nil {
		log.Error("hottab s.arc.ArchivesPB aids(%v) error(%v)", aids, err)
		return
	}
	if len(as) == 0 {
		log.Warn("hottab s.arc.ArchivesPB(%v) length is 0", aids)
		return
	}
	innerArc := s.controld.CircleReqInternalAttr(c, aids)
	for i, aid := range aids {
		if aid == 0 {
			continue
		}
		var (
			r    = &ai.Item{Goto: model.GotoAv, ID: aid}
			h    = cardmv2.Handle(plat, cdm.CardGt(r.Goto), "", 1, r, nil, nil, nil, nil, nil, nil, nil)
			main interface{}
		)
		if h == nil {
			continue
		}
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, model.PlatIPhone, 0, "")
		var isOverseas bool
		if innerArc != nil && innerArc[aid] != nil {
			isOverseas = innerArc[aid].OverSeaBlock
		}
		if a, ok := as[aid]; ok && !isOverseas {
			main = map[int64]*arcgrpc.ArcPlayer{a.Aid: {Arc: a}}
		}
		if main != nil {
			h.From(main, op)
		}
		h.Get().Idx = int64(i + 1)
		if h.Get().Right {
			res = append(res, cardmv2.AddCard(h))
		}
	}
	if len(res) == 0 {
		res = _emptyList3
		return
	}
	return
}
