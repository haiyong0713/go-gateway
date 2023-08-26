package feed

import (
	"context"
	"time"

	"go-common/library/log"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

var (
	// av2314237 已经删除，后续看情况处理
	// 已删除失效id已更换新的稿件id
	_aids = []int64{308040, 2431658, 2432648, 2427553, 539600, 1968681, 850424, 887861, 1960912, 1935680, 1406019,
		1985297, 1977493, 2312184, 2316891, 864845, 1986932, 880857, 875624, 744299}
)

// Audit2 check audit plat and ip, then return audit data.
func (s *Service) Audit(c context.Context, mobiApp string, plat int8, build int, column cdm.ColumnStatus) (is []card.Handler, ok bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return s.auditData(c, plat, build, column), true
		}
	}
	return
}

// auditData2 some data for audit.
func (s *Service) auditData(c context.Context, plat int8, _ int, column cdm.ColumnStatus) (is []card.Handler) {
	am, err := s.arc.Archives(c, _aids)
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
			main = map[int64]*arcgrpc.ArcPlayer{a.Aid: {Arc: a}}
			if err = i.From(main, op); err != nil {
				log.Error("Failed to From: %+v", err)
			}
			if !i.Get().Right {
				continue
			}
			if model.IsIPad(plat) {
				// ipad卡片不展示标签
				i.Get().DescButton = nil
			}
			is = append(is, i)
		}
	}
	return
}

func (s *Service) loadAuditCache() {
	as, err := s.rsc.AppAudit(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.auditCache = as
}

// auditproc load audit cache.
func (s *Service) auditproc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.Tick))
		s.loadAuditCache()
	}
}
