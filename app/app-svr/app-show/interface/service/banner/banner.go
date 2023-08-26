package banner

import (
	"context"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/app-show/interface/conf"
	resdao "go-gateway/app/app-svr/app-show/interface/dao/resource"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/banner"
	resource "go-gateway/app/app-svr/resource/service/model"
)

var (
	_banners = map[string]map[string]map[int8]int{
		"discover": {
			"bottom": {
				model.PlatIPhone:  452,
				model.PlatIPad:    800,
				model.PlatIPhoneI: 1085,
				model.PlatIPadI:   1255,
			},
		},
		"mine": {
			"top": {
				model.PlatIPhone:  449,
				model.PlatIPad:    801,
				model.PlatIPhoneI: 1089,
				model.PlatIPadI:   1259,
			},
			"center": {
				model.PlatIPhone:  450,
				model.PlatIPad:    802,
				model.PlatIPhoneI: 1093,
				model.PlatIPadI:   1263,
			},
			"bottom": {
				model.PlatIPhone:  451,
				model.PlatIPad:    803,
				model.PlatIPhoneI: 1097,
				model.PlatIPadI:   1267,
			},
		},
	}
	_bannersPlat = map[int8]string{
		model.PlatIPhone:  "452,499,450,451",
		model.PlatIPad:    "800,801,802,803",
		model.PlatIPhoneI: "1085,1089,1093,1097",
		model.PlatIPadI:   "1255,1259,1263,1267",
	}
)

// Service is banner service.
type Service struct {
	c *conf.Config
	// dao              *bndao.Dao
	res         *resdao.Dao
	bannerCache map[int8]map[int][]*resource.Banner
	// prom
	prmobi *prom.Prom
}

// New new a banner service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		// dao:              bndao.New(c),
		res:         resdao.New(c),
		bannerCache: map[int8]map[int][]*resource.Banner{},
		// prom
		prmobi: prom.BusinessInfoCount,
	}
	s.load()
	// nolint:biligowordcheck
	go s.loadproc()
	return
}

// Display get banner.
func (s *Service) Display(c context.Context, plat int8, build int, channel, module, position, mobiApp string) (res map[string][]*banner.Banner) {
	ip := metadata.String(c, metadata.RemoteIP)
	res = s.getCache(c, plat, build, channel, module, position, ip)
	s.prmobi.Incr("banner_plat_" + mobiApp)
	return
}

// getCahce get banner from cache.
func (s *Service) getCache(c context.Context, plat int8, build int, channel, module, position, ip string) (res map[string][]*banner.Banner) {
	res = map[string][]*banner.Banner{}
	var (
		resIDs = _bannersPlat[plat]
		err    error
		resbs  map[int][]*resource.Banner
		plm    = s.bannerCache[plat]
		resID  int
	)
	if resbs, err = s.res.ResBanner(c, plat, build, 0, resIDs, channel, ip, "", "", "", "", "", false); err != nil || len(resbs) == 0 {
		log.Error("s.res.ResBanner is null or err(%v)", err)
		resbs = plm
	}
	mds := strings.Split(module, ",")
	poss := strings.Split(position, ",")
	for _, md := range mds {
		for _, pos := range poss {
			resID = _banners[md][pos][plat]
			res[md+"_"+pos] = s.resBanners(resbs[resID])
		}
	}
	return
}

// resBannersplat
func (s *Service) resBanners(rbs []*resource.Banner) (res []*banner.Banner) {
	for _, rb := range rbs {
		b := &banner.Banner{}
		b.ResChangeBanner(rb)
		res = append(res, b)
	}
	return
}

// load load all banner.
func (s *Service) load() {
	var (
		resbs = map[int8]map[int][]*resource.Banner{}
	)
	for plat, resIDStr := range _bannersPlat {
		mobiApp := model.MobiApp(plat)
		res, err := s.res.ResBanner(context.TODO(), plat, 515007, 0, resIDStr, "master", "", "", "", mobiApp, "", "", false)
		if err != nil || len(res) == 0 {
			log.Error("s.res.ResBanner is null or err(%v)", err)
			return
		}
		resbs[plat] = res
	}
	if len(resbs) > 0 {
		s.bannerCache = resbs
	}
	log.Info("banner cacheproc success")
}

// cacheproc load cache.
func (s *Service) loadproc() {
	for {
		time.Sleep(time.Duration(s.c.CustomTick.Tick))
		s.load()
	}
}
