package rank

import (
	"context"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup"
	rankmod "go-gateway/app/app-svr/app-show/interface/api/rank"
	"go-gateway/app/app-svr/app-show/interface/conf"
	accdao "go-gateway/app/app-svr/app-show/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-show/interface/dao/archive"
	adtdao "go-gateway/app/app-svr/app-show/interface/dao/audit"
	condao "go-gateway/app/app-svr/app-show/interface/dao/control"
	rankdao "go-gateway/app/app-svr/app-show/interface/dao/rank"
	rcmmndao "go-gateway/app/app-svr/app-show/interface/dao/recommend"
	rgdao "go-gateway/app/app-svr/app-show/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-show/interface/dao/relation"
	"go-gateway/app/app-svr/app-show/interface/model/rank"
	"go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/robfig/cron"
)

const (
	_initRank = "rank_key_%s_%d"
)

type Service struct {
	c    *conf.Config
	cron *cron.Cron
	// region
	rdao *rgdao.Dao
	// rcmmnd
	rcmmnd *rcmmndao.Dao
	// archive
	arc *arcdao.Dao
	// audit
	adt *adtdao.Dao
	// account
	accd *accdao.Dao
	// relation
	reldao *reldao.Dao
	// rank
	rankd *rankdao.Dao
	// controldao
	controld *condao.Dao
	// ranking
	rankCache     map[string][]*rankmod.Item
	rankOseaCache map[string][]*rankmod.Item
	// audit cache
	auditCache map[string]map[int]struct{} // audit mobi_app builds
	// rank job railgun
	rankJobRailGun *railgun.Railgun
}

// New new a region service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		cron:   cron.New(),
		rdao:   rgdao.New(c),
		rcmmnd: rcmmndao.New(c),
		// archive
		arc: arcdao.New(c),
		// audit
		adt: adtdao.New(c),
		// account
		accd: accdao.New(c),
		// relation
		reldao: reldao.New(c),
		// rank
		rankd:    rankdao.New(c),
		controld: condao.New(c),
		// ranking
		rankCache:     map[string][]*rankmod.Item{},
		rankOseaCache: map[string][]*rankmod.Item{},
		// audit cache
		auditCache: map[string]map[int]struct{}{},
	}
	s.loadAuditCache()
	s.initAuditRailGun() // 间隔3分钟
	return
}

func (s *Service) initAuditRailGun() {
	r := railgun.NewRailGun("rankServiceRailGun", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 */3 * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadAuditCache()
		return railgun.MsgPolicyNormal
	}))
	s.rankJobRailGun = r
	r.Start()
}

func (s *Service) RankShow(c context.Context, plat int8, rid, pn, ps int, mid int64, order string, mobiApp, device string) (res []*rankmod.Item, err error) {
	res = []*rankmod.Item{}
	var (
		as              map[int64]*api.Arc
		authorMids      []int64
		list            []*rank.List
		aids            []int64
		authorRelations map[int64]*relationgrpc.InterrelationReply
		authorStats     map[int64]*relationgrpc.StatReply
		authorCards     map[int64]*accountgrpc.Card
	)
	start := (pn - 1) * ps
	end := start + (ps - 1)
	if list, err = s.rankd.RankCache(c, order, rid, start, end); err != nil {
		log.Error("%+v", err)
		err = ecode.New(http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		return
	}
	for _, l := range list {
		if l.Aid > 0 {
			aids = append(aids, l.Aid)
		}
		for _, other := range l.Others {
			if other.Aid > 0 {
				aids = append(aids, other.Aid)
			}
		}
	}
	if as, err = s.arc.ArchivesPB(c, aids, mid, mobiApp, device); err != nil || len(as) == 0 {
		log.Error("%+v", err)
		return
	}
	for _, l := range list {
		// up mid
		if a, ok := as[l.Aid]; ok && a.Author.Mid > 0 {
			authorMids = append(authorMids, a.Author.Mid)
		}
	}
	var innerAttr = make(map[int64]*rank.InnerAttr)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() error {
		innerAttr = s.controld.CircleReqInternalAttr(ctx, aids)
		return nil

	})
	if len(authorMids) > 0 {
		g.Go(func() error {
			if authorCards, err = s.accd.Cards3GRPC(ctx, authorMids); err != nil {
				log.Error("s.accd.Cards3 error(%v)", err)
			}
			return nil
		})
		if mid > 0 {
			g.Go(func() error {
				if authorRelations, err = s.reldao.RelationsInterrelations(ctx, mid, authorMids); err != nil {
					log.Error("s.accd.Relations2 error(%v)", err)
				}
				return nil
			})
		}
		g.Go(func() error {
			if authorStats, err = s.reldao.StatsGRPC(ctx, authorMids); err != nil {
				log.Error("s.reldao.Stats error(%v)", err)
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("RankUser errgroup.WithContext error(%v)", err)
	}
	for _, l := range list {
		if tmp := l.FromRanks(as, authorRelations, authorStats, authorCards, plat, innerAttr); tmp != nil {
			res = append(res, tmp)
		}
	}
	return
}

func (s *Service) Close() {
	s.rankJobRailGun.Close()
	s.cron.Stop()
}
