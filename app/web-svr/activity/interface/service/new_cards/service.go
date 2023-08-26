package cards

import (
	"sync/atomic"

	http "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/web-svr/activity/interface/conf"
	act "go-gateway/app/web-svr/activity/interface/dao/actplat"
	dao "go-gateway/app/web-svr/activity/interface/dao/cards"
	likedao "go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/wechat"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	"go-gateway/app/web-svr/activity/interface/service/account"
	likeApi "go-gateway/app/web-svr/activity/interface/service/like"
	lotteryApi "go-gateway/app/web-svr/activity/interface/service/lottery"
	taskApi "go-gateway/app/web-svr/activity/interface/service/task"
)

// Service ...
type Service struct {
	c          *conf.Config
	dao        *dao.Dao
	likeDao    *likedao.Dao
	actDao     *act.Dao
	config     *atomic.Value
	httpClient *http.Client
	lotterySvr *lotteryApi.Service
	likeSvr    *likeApi.Service
	taskSvr    *taskApi.Service
	allTask    []*cardsmdl.Task
	followMid  []*cardsmdl.FollowMid
	ogvLink    []*cardsmdl.OgvLink
	account    *account.Service
	wechatdao  *wechat.Dao
	cache      *fanout.Fanout
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		dao:        dao.New(c),
		likeDao:    likedao.New(c),
		config:     &atomic.Value{},
		httpClient: http.NewClient(c.HTTPClient),
		lotterySvr: lotteryApi.New(c),
		likeSvr:    likeApi.New(c),
		taskSvr:    taskApi.New(c),
		actDao:     act.New(c),
		wechatdao:  wechat.New(c),
		account:    account.New(c),
		cache:      fanout.New("cards", fanout.Worker(1), fanout.Buffer(1024)),
	}

	return s
}

// Close ...
func (s *Service) Close() {
}
