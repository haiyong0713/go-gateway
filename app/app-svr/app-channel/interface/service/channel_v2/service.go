package channel_v2

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/log/infoc"

	"go-gateway/app/app-svr/app-channel/interface/conf"
	accdao "go-gateway/app/app-svr/app-channel/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-channel/interface/dao/archive"
	pgcdao "go-gateway/app/app-svr/app-channel/interface/dao/bangumi"
	chdao "go-gateway/app/app-svr/app-channel/interface/dao/channel"
	coindao "go-gateway/app/app-svr/app-channel/interface/dao/coin"
	dyndao "go-gateway/app/app-svr/app-channel/interface/dao/dynamic"
	favdao "go-gateway/app/app-svr/app-channel/interface/dao/favorite"
	natdao "go-gateway/app/app-svr/app-channel/interface/dao/nat-page"
	pediadao "go-gateway/app/app-svr/app-channel/interface/dao/pedia"
	tabdao "go-gateway/app/app-svr/app-channel/interface/dao/tab"
	thumbupdao "go-gateway/app/app-svr/app-channel/interface/dao/thumbup"
	topicdao "go-gateway/app/app-svr/app-channel/interface/dao/topic"
	chmdl "go-gateway/app/app-svr/app-channel/interface/model/channel_v2"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/robfig/cron"
)

type Service struct {
	c       *conf.Config
	chDao   *chdao.Dao
	arcDao  *arcdao.Dao
	tabDao  *tabdao.Dao
	favDao  *favdao.Dao
	coinDao *coindao.Dao
	pgcDao  *pgcdao.Dao
	// default tab cache
	menuCache map[int64][]*tab.Menu
	// infoc
	logCh chan interface{}
	// cron
	cron       *cron.Cron
	natDao     *natdao.Dao
	topicDao   *topicdao.Dao
	dynDao     *dyndao.Dao
	pediaDao   *pediadao.Dao
	accDao     *accdao.Dao
	thumbupDao *thumbupdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		chDao:     chdao.New(c),
		arcDao:    arcdao.New(c),
		tabDao:    tabdao.New(c),
		favDao:    favdao.New(c),
		coinDao:   coindao.New(c),
		pgcDao:    pgcdao.New(c),
		menuCache: make(map[int64][]*tab.Menu),
		// infoc
		logCh: make(chan interface{}, 1024),
		// cron
		cron: cron.New(),
		// nat dao
		natDao: natdao.New(c),
		// dynamic topic
		dynDao:     dyndao.New(c),
		topicDao:   topicdao.New(c),
		pediaDao:   pediadao.New(c),
		accDao:     accdao.New(c),
		thumbupDao: thumbupdao.New(c),
	}
	s.loadMenusCache()
	s.initCron()
	s.cron.Start()
	// nolint:biligowordcheck
	go s.infocproc()
	return
}

func (s *Service) initCron() {
	if err := s.cron.AddFunc(s.c.Cron.LoadMenusCacheV2, s.loadMenusCache); err != nil {
		panic(err)
	}
}

func (s *Service) loadMenusCache() {
	var (
		menus map[int64][]*tab.Menu
		err   error
	)
	if menus, err = s.tabDao.MenusNew(context.TODO(), time.Now()); err != nil {
		log.Error("%v", err)
		return
	}
	s.menuCache = menus
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

func (s *Service) infocproc() {
	var cardShowInfoc = infoc.New(s.c.NewChannelCardShowInfoc)
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch i := i.(type) {
		case *chmdl.ChannelInfoc:
			b, _ := json.Marshal(&i.Items)
			_ = cardShowInfoc.Info(i.EventId, i.Page, i.Sort, i.Filt, string(b), strconv.Itoa(i.CardNum), i.RequestUrl,
				strconv.FormatInt(i.TimeIso, 10), i.Ip, strconv.Itoa(int(i.AppId)), strconv.Itoa(int(i.Platform)),
				i.Buvid, i.Version, i.VersionCode, i.Mid, i.Ctime, i.Abtest, i.AutoRefresh, i.From, i.Pos, i.CurRefresh)
		}
	}
}

// Archives 存在aids中某个稿件不需要秒开的可能性 因此aids聚合做在外层
func (s *Service) Archives(c context.Context, aidsp []*arcgrpc.PlayAv, isPlayurl bool) (map[int64]*arcgrpc.ArcPlayer, error) {
	if isPlayurl {
		res, err := s.arcDao.ArcsPlayer(c, aidsp, false)
		if err != nil {
			log.Error("%v", err)
			return nil, err
		}
		return res, nil
	} else {
		var aids []int64
		for _, aidp := range aidsp {
			if aidp != nil && aidp.Aid != 0 {
				aids = append(aids, aidp.Aid)
			}
		}
		tmps, err := s.arcDao.Arcs(c, aids)
		if err != nil {
			log.Error("%v", err)
			return nil, err
		}
		var res = make(map[int64]*arcgrpc.ArcPlayer)
		for aid, tmp := range tmps {
			if tmp != nil {
				var re = new(arcgrpc.Arc)
				*re = *tmp
				res[aid] = &arcgrpc.ArcPlayer{Arc: re}
			}
		}
		return res, nil
	}
}
