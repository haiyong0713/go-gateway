package show

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/conf"
	accdao "go-gateway/app/app-svr/app-car/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-car/interface/dao/archive"
	bgmdao "go-gateway/app/app-svr/app-car/interface/dao/bangumi"
	channeldao "go-gateway/app/app-svr/app-car/interface/dao/channel"
	dyndao "go-gateway/app/app-svr/app-car/interface/dao/dynamic"
	favdao "go-gateway/app/app-svr/app-car/interface/dao/favorite"
	historydao "go-gateway/app/app-svr/app-car/interface/dao/history"
	intervenedao "go-gateway/app/app-svr/app-car/interface/dao/intervene"
	rcmddao "go-gateway/app/app-svr/app-car/interface/dao/recommend"
	regdao "go-gateway/app/app-svr/app-car/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-car/interface/dao/relation"
	resdao "go-gateway/app/app-svr/app-car/interface/dao/resource"
	srchdao "go-gateway/app/app-svr/app-car/interface/dao/search"
	showdao "go-gateway/app/app-svr/app-car/interface/dao/show"
	silverdao "go-gateway/app/app-svr/app-car/interface/dao/silverbullet"
	updao "go-gateway/app/app-svr/app-car/interface/dao/up"
	intervene "go-gateway/app/app-svr/app-car/interface/model/intervene"

	"github.com/robfig/cron"
)

type Service struct {
	c            *conf.Config
	dao          *showdao.Dao
	arc          *arcdao.Dao
	reldao       *reldao.Dao
	acc          *accdao.Dao
	srch         *srchdao.Dao
	bgm          *bgmdao.Dao
	his          *historydao.Dao
	dyn          *dyndao.Dao
	rcmd         *rcmddao.Dao
	up           *updao.Dao
	fav          *favdao.Dao
	reg          *regdao.Dao
	channelDao   *channeldao.Dao
	silverDao    *silverdao.Dao
	resDao       *resdao.Dao
	rnd          *rand.Rand
	interveneDao *intervenedao.Dao
	cron         *cron.Cron
	xiaoPengRecs map[string]*intervene.XiaoPengRecShowList
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:            c,
		dao:          showdao.New(c),
		arc:          arcdao.New(c),
		reldao:       reldao.New(c),
		acc:          accdao.New(c),
		srch:         srchdao.New(c),
		bgm:          bgmdao.New(c),
		his:          historydao.New(c),
		dyn:          dyndao.New(c),
		rcmd:         rcmddao.New(c),
		up:           updao.New(c),
		fav:          favdao.New(c),
		reg:          regdao.New(c),
		channelDao:   channeldao.New(c),
		silverDao:    silverdao.New(c),
		resDao:       resdao.New(c),
		rnd:          rand.New(rand.NewSource(time.Now().Unix())),
		interveneDao: intervenedao.New(c, conf.GetDB(c)),
		cron:         cron.New(),
		xiaoPengRecs: make(map[string]*intervene.XiaoPengRecShowList),
	}
	// 预定一个小时一次 定时加载小鹏美妆空间的干预卡片
	//todo 没有配置化,测试时候5分钟一次
	if s.c.CustomModule.XiaoPengInterveneCron == "" {
		s.c.CustomModule.XiaoPengInterveneCron = "0 0 * * * *"
	}
	if err := s.cron.AddFunc(s.c.CustomModule.XiaoPengInterveneCron, s.loadInterveneData); err != nil {
		panic(err)
	}
	s.cron.Start()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Warn("LoadInterveneData err. err: %+v", err)
				return
			}
		}()
		//系统初始化 加载干预卡片
		s.loadInterveneData()
	}()
	return s
}
func (s *Service) loadInterveneData() {
	items, err := s.interveneDao.LoadInterveneData(context.Background())
	if err != nil || items == nil {
		log.Warn("LoadInterveneData err. err: %+v", err)
		return
	}
	if len(items) < 1 {
		log.Info("LoadInterveneData empty.")
		return
	}
	log.Info("LoadInterveneData begin.")
	defer func() {
		if err := recover(); err != nil {
			// 打印异常，关闭资源，退出此函数
			log.Warn("LoadInterveneData err %+v", err)
		}
	}()
	_xiaoPengRecs := make(map[string]*intervene.XiaoPengRecShowList)
	for _, item := range items {
		key := fmt.Sprintf("%d_%s", item.Type, item.KeyWord)
		interv, ok := _xiaoPengRecs[key]
		if !ok && interv == nil {
			interv = &intervene.XiaoPengRecShowList{
				Items: make([]int64, 0),
			}
		}
		interv.Items = append(interv.Items, item.Aid)
		_xiaoPengRecs[key] = interv
	}
	s.xiaoPengRecs = _xiaoPengRecs
	log.Info("LoadInterveneData end.len:%d", len(s.xiaoPengRecs))
}
