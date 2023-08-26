package account

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/queue/databus"
	ottClient "go-gateway/app/app-svr/app-interface/api-dependence/ott-service"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	asdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/answer"
	audiodao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/audio"
	bplusdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bplus"
	favdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/favorite"
	gallerydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/gallery"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/game"
	hisdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/history"
	livedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/live"
	locdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/location"
	memberdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/member"
	paydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/pay"
	reldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/relation"
	relsh1dao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/relation-sh1"
	resdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/resource"
	sidedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/sidebar"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/toview"
	usersuitdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/usersuit"
	vodao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/videoup-open"
	vipdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/vip"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	resmodel "go-gateway/app/app-svr/resource/service/model"

	locmdl "git.bilibili.co/bapis/bapis-go/community/service/location"

	"github.com/robfig/cron"
)

// Service is space service
type Service struct {
	c            *conf.Config
	accDao       *accdao.Dao
	audioDao     *audiodao.Dao
	relDao       *reldao.Dao
	relsh1Dao    *relsh1dao.Dao
	bplusDao     *bplusdao.Dao
	payDao       *paydao.Dao
	memberDao    *memberdao.Dao
	sideDao      *sidedao.Dao
	vipDao       *vipdao.Dao
	usersuitDao  *usersuitdao.Dao
	voDao        *vodao.Dao
	asDao        *asdao.Dao
	liveDao      *livedao.Dao
	resDao       *resdao.Dao
	loc          *locdao.Dao
	favDao       *favdao.Dao
	toViewDao    *toview.Dao
	galleryDao   *gallerydao.Dao
	gameDao      *game.Dao
	hisDao       *hisdao.Dao
	sectionCache map[string][]*SectionItem
	white        map[int8][]*SectionURL
	redDot       map[int8][]*SectionURL
	// authPids         map[int8][]string
	authZoneLimitIDs map[int8]map[int64]*locmdl.ZoneLimitAuth
	// databus
	configSetPub *databus.Databus
	// cron
	cron      *cron.Cron
	ottclient ottClient.OTTServiceClient
}

// SectionItem is
type SectionItem struct {
	Item  *resmodel.SideBar
	Limit []*resmodel.SideBarLimit
}

// SectionURL is
type SectionURL struct {
	ID  int64
	URL string
}

// CheckLimit is
func (item *SectionItem) CheckLimit(build int) bool {
	for _, l := range item.Limit {
		if model.InvalidBuild(build, l.Build, l.Condition) {
			return false
		}
	}
	return true
}

// New new space
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		accDao:       accdao.New(c),
		audioDao:     audiodao.New(c),
		relDao:       reldao.New(c),
		relsh1Dao:    relsh1dao.New(c),
		loc:          locdao.New(c),
		bplusDao:     bplusdao.New(c),
		payDao:       paydao.New(c),
		memberDao:    memberdao.New(c),
		sideDao:      sidedao.New(c),
		vipDao:       vipdao.New(c),
		usersuitDao:  usersuitdao.New(c),
		voDao:        vodao.New(c),
		asDao:        asdao.New(c),
		liveDao:      livedao.New(c),
		resDao:       resdao.New(c),
		favDao:       favdao.New(c),
		toViewDao:    toview.New(c),
		galleryDao:   gallerydao.New(c),
		gameDao:      game.New(c),
		hisDao:       hisdao.New(c),
		sectionCache: map[string][]*SectionItem{},
		white:        map[int8][]*SectionURL{},
		redDot:       map[int8][]*SectionURL{},
		// authPids:     map[int8][]string{},
		authZoneLimitIDs: map[int8]map[int64]*locmdl.ZoneLimitAuth{},
		// databus
		configSetPub: databus.New(c.ConfigSetPub),
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	var err error
	s.ottclient, err = ottClient.NewClient(c.OTTGRPC)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Service) initCron() {
	s.loadSidebar()
	if err := s.cron.AddFunc(s.c.Cron.LoadSidebar, s.loadSidebar); err != nil {
		panic(err)
	}
}

func (s *Service) loadSidebar() {
	log.Info("cronLog start loadSidebar")
	var (
		sidebar *resmodel.SideBars
		err     error
	)
	if sidebar, err = s.sideDao.Sidebars(context.TODO()); err != nil {
		log.Error("s.sideDao.SideBars error(%v)", err)
		return
	}
	ss := make(map[string][]*SectionItem)
	white := make(map[int8][]*SectionURL)
	redDot := make(map[int8][]*SectionURL)
	// pids := make(map[int8][]string)
	authZoneLimitIDs := make(map[int8]map[int64]*locmdl.ZoneLimitAuth)
	moduleMap := map[int]struct{}{

		model.SelfCenter:           {},
		model.MyService:            {},
		model.Creative:             {},
		model.IPadSelfCenter:       {},
		model.IPadCreative:         {},
		model.AndroidSelfCenter:    {},
		model.AndroidCreative:      {},
		model.AndroidMyService:     {},
		model.OpModule:             {},
		model.AndroidBSelfCenter:   {},
		model.AndroidBMyService:    {},
		model.AndroidISelfCenter:   {},
		model.AndroidIMyService:    {},
		model.IPhoneBselfCenter:    {},
		model.IPhoneBmyService:     {},
		model.IPadHDSelfCenter:     {},
		model.IPadHDCreative:       {},
		model.AndroidPadSelfCenter: {},
	}
	for _, item := range sidebar.SideBar {
		item.Plat = s.convertMgrPlatToInterface(item.Plat)
		key := fmt.Sprintf(_initSidebarKey, item.Plat, item.Module, item.Language)
		ss[key] = append(ss[key], &SectionItem{
			Item:  item,
			Limit: sidebar.Limit[item.ID],
		})
		if _, ok := moduleMap[item.Module]; ok {
			if item.WhiteURL != "" {
				white[item.Plat] = append(white[item.Plat], &SectionURL{ID: item.ID, URL: item.WhiteURL})
			}
			if item.Red != "" {
				redDot[item.Plat] = append(redDot[item.Plat], &SectionURL{ID: item.ID, URL: item.Red})
			}
			// if item.Area != "" {
			// 	pids[item.Plat] = append(pids[item.Plat], item.Area)
			// }
			if item.AreaPolicy > 0 {
				if _, ok := authZoneLimitIDs[item.Plat]; !ok {
					authZoneLimitIDs[item.Plat] = map[int64]*locmdl.ZoneLimitAuth{}
				}
				authZoneLimitIDs[item.Plat][item.AreaPolicy] = &locmdl.ZoneLimitAuth{}
				switch item.ShowPurposed {
				case 0:
					authZoneLimitIDs[item.Plat][item.AreaPolicy].Play = locmdl.Status_Allow
				case 1:
					authZoneLimitIDs[item.Plat][item.AreaPolicy].Play = locmdl.Status_Forbidden
				}
			}
		}
	}
	s.sectionCache = ss
	s.white = white
	s.redDot = redDot
	// s.authPids = pids
	s.authZoneLimitIDs = authZoneLimitIDs
	//nolint:gosimple
	return
}

func (s *Service) convertMgrPlatToInterface(mgrPlat int8) int8 {
	switch mgrPlat {
	case model.PlatMgrIphoneB:
		return model.PlatIPhoneB
	case model.PlatMgrAndroidB:
		return model.PlatAndroidB
	case model.PlatMgrIPadHD:
		return model.PlatIpadHD
	default:
		return mgrPlat
	}
}

func (s *Service) convertInterfacePlatToMgr(interfacePlat int8) int8 {
	switch interfacePlat {
	case model.PlatIpadHD:
		return model.PlatMgrIPadHD
	case model.PlatAndroidB:
		return model.PlatMgrAndroidB
	case model.PlatIPhoneB:
		return model.PlatMgrIphoneB
	default:
		return interfacePlat
	}
}
