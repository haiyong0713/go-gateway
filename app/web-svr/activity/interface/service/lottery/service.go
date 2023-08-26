package lottery

import (
	"context"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/dao/actplat"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	suitapi "go-main/app/account/usersuit/service/api"
	"time"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	coinapi "git.bilibili.co/bapis/bapis-go/community/service/coin"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/cards"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	lottery "go-gateway/app/web-svr/activity/interface/dao/lottery_v2"
	"go-gateway/app/web-svr/activity/interface/dao/pay"
	"go-gateway/app/web-svr/activity/interface/dao/wechat"
	modell "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	couponapi "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	figapi "git.bilibili.co/bapis/bapis-go/account/service/figure"
	spyapi "git.bilibili.co/bapis/bapis-go/account/service/spy"
	ogvapi "git.bilibili.co/bapis/bapis-go/cheese/service/coupon/v2"
	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/pay"
	locationapi "git.bilibili.co/bapis/bapis-go/community/service/location"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	vipresourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
	vipinfoapi "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
)

// Service ...
type Service struct {
	c                  *conf.Config
	lottery            lottery.Dao
	pay                pay.Dao
	accClient          api.AccountClient
	cache              *fanout.Fanout
	actplat            *actplat.Dao
	cardsdao           *cards.Dao
	vipInfoClient      vipinfoapi.VipInfoClient
	figureClient       figapi.FigureClient
	spyClient          spyapi.SpyClient
	passportClient     passportinfoapi.PassportUserClient
	cheeseClient       cheeseapi.PayClient
	silverbulletClient silverbulletapi.GaiaClient
	locationClient     locationapi.LocationClient
	coinClient         coinapi.CoinClient
	couponClient       couponapi.CouponClient
	vipResourceClient  vipresourceapi.ResourceClient
	suitClient         suitapi.UsersuitClient
	ogvClient          ogvapi.CouponPlatformClient
	likedao            *like.Dao
	wechatdao          *wechat.Dao

	newLotterySids map[string]struct{}
	actTaskMap     map[int64][]*cardsmdl.Task
	cluesSrcs      []*modell.Item
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		lottery:    lottery.New(c),
		pay:        pay.New(c),
		likedao:    like.New(c),
		wechatdao:  wechat.New(c),
		cache:      fanout.New("lottery_cache", fanout.Worker(1), fanout.Buffer(1024)),
		actplat:    actplat.New(c),
		cardsdao:   cards.New(c),
		actTaskMap: make(map[int64][]*cardsmdl.Task),
		cluesSrcs:  c.AprilFoolsAct.CluesSrcs,
	}
	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passportinfoapi.NewClient(c.PassportClient); err != nil {
		panic(err)
	}
	if s.silverbulletClient, err = silverbulletapi.NewClient(c.SilverGaiaClient); err != nil {
		panic(err)
	}
	if s.vipInfoClient, err = vipinfoapi.NewClient(c.VipClient); err != nil {
		panic(err)
	}
	if s.locationClient, err = locationapi.NewClient(c.LocationRPC); err != nil {
		panic(err)
	}
	if s.spyClient, err = spyapi.NewClient(c.SpyClient); err != nil {
		panic(err)
	}
	if s.figureClient, err = figapi.NewClient(c.Figure); err != nil {
		panic(err)
	}
	if s.coinClient, err = coinapi.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if s.couponClient, err = couponapi.NewClient(c.CouponClient); err != nil {
		panic(err)
	}
	if s.vipResourceClient, err = vipresourceapi.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	if s.suitClient, err = suitapi.NewClient(c.SuitClient); err != nil {
		panic(err)
	}
	if s.ogvClient, err = ogvapi.NewClient(c.OgvClient); err != nil {
		panic(err)
	}
	if s.cheeseClient, err = cheeseapi.NewClient(c.CheeseClient); err != nil {
		panic(err)
	}
	ctx := context.Background()
	if s.actTaskMap, err = s.updateActTask(ctx); err != nil {
		panic(err)
	}
	log.Infoc(ctx, "syncActTask in memory : CluesSrc %v", s.cluesSrcs)
	go s.syncActTask(ctx)
	return s
}

// Close ...
func (s *Service) Close() {
	s.lottery.Close()
}

func (s *Service) syncActTask(ctx context.Context) (err error) {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		TaskMap, err := s.updateActTask(ctx)
		log.Infoc(ctx, "syncActTask updateActTask %v err(%v)", TaskMap, err)
		if err == nil {
			s.actTaskMap = TaskMap
		}

		clues, err := s.lottery.GetCluesSrc(ctx, s.c.AprilFoolsAct.JsonData, time.Now().Unix())
		log.Infoc(ctx, "syncActTask GetCluesSrc %v err(%v)", clues, err)
		if err == nil {
			s.cluesSrcs = clues
		}
	}
	return
}

func (s *Service) updateActTask(ctx context.Context) (actMap map[int64][]*cardsmdl.Task, err error) {
	task, err := s.cardsdao.AllTaskList(ctx)
	if err == nil {
		actMap = make(map[int64][]*cardsmdl.Task)
		for _, v := range task {
			if list, ok := actMap[v.ActivityID]; ok {
				actMap[v.ActivityID] = append(list, v)
				continue
			}
			actMap[v.ActivityID] = []*cardsmdl.Task{v}
		}
	}
	return
}
