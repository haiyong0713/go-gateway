package service

import (
	"context"
	"go-common/library/sync/pipeline/fanout"

	"go-common/library/conf/paladin"
	pb "go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/dao"
	ottDao "go-gateway/app/web-svr/dance-taiko/interface/internal/dao/ott"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.DanceTaikoServer), new(*Service)))

type CommentType string

const (
	_commentOk      = CommentType("OK")
	_commentGood    = CommentType("Good")
	_commentPerfect = CommentType("Perfect")
	_commentMiss    = CommentType("Miss")
)

// Service service.
type Service struct {
	ac     *paladin.Map
	dao    dao.Dao
	ottDao ottDao.Dao
	cron   *cron.Cron

	// 评分标准
	arcExamples  map[int64]map[int64]float64
	arcKeyFrames map[int64][]int64
	accClient    accClient.AccountClient

	// stats
	standardAcc map[CommentType]float64

	conf   *model.Conf
	fanout *fanout.Fanout
}

// New new a service and return.
func New(d dao.Dao, ottDao ottDao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:           &paladin.TOML{},
		dao:          d,
		ottDao:       ottDao,
		cron:         cron.New(),
		arcExamples:  make(map[int64]map[int64]float64),
		arcKeyFrames: make(map[int64][]int64),
		standardAcc:  make(map[CommentType]float64),
		fanout:       fanout.New("cache"),
	}

	if err = paladin.Get("application.toml").UnmarshalTOML(&s.conf); err != nil {
		panic(err)
	}
	if err = paladin.Watch("application.toml", s.conf); err != nil {
		panic(err)
	}

	if err = s.loadExamples(context.Background()); err != nil {
		panic(err)
	}
	func() {
		// 从配置中加载关键帧
		for _, v := range s.conf.Frames {
			s.arcKeyFrames[v.Aid] = v.Frames
		}

		// 从配置中加载评分标准
		s.standardAcc[_commentPerfect] = s.conf.Perfect
		s.standardAcc[_commentGood] = s.conf.Good
		s.standardAcc[_commentOk] = s.conf.OK

	}()

	if s.accClient, err = accClient.NewClient(nil); err != nil {
		panic(err)
	}
	return s, s.Close, nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *Service) Close() {
	s.fanout.Close()
}
