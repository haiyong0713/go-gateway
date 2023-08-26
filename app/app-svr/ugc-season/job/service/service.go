package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/ugc-season/job/conf"
	"go-gateway/app/app-svr/ugc-season/job/dao/archive"
	"go-gateway/app/app-svr/ugc-season/job/dao/result"
	statDao "go-gateway/app/app-svr/ugc-season/job/dao/stat"
	"go-gateway/app/app-svr/ugc-season/job/model/stat"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
)

const _sharding = 100

// Service service
type Service struct {
	c                    *conf.Config
	closeRetry           bool
	closeSub             bool
	archiveDao           *archive.Dao
	resultDao            *result.Dao
	statDao              *statDao.Dao
	redis                *redis.Pool
	waiter               sync.WaitGroup
	waiterSeason         sync.WaitGroup
	seasonSub            *databus.Databus
	seasonWithArchivePub *databus.Databus
	arcClient            arcApi.ArchiveClient
	seasonClient         seasonApi.UGCSeasonClient
	seasonIDs            chan int64
	// season stat
	subSnMap map[string]*databus.Databus
	statCh   chan interface{} // model.StatMsg or model.SeasonResult
	//rail_gun
	CoinSnSubV2 *railgun.Railgun
}

// nolint:biligowordcheck
// New is archive service implementation.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                    c,
		archiveDao:           archive.New(c),
		resultDao:            result.New(c),
		statDao:              statDao.New(c),
		seasonSub:            databus.New(c.SeasonSub),
		seasonWithArchivePub: databus.New(c.SeasonWithArchivePub),
		redis:                redis.NewPool(c.Redis),
		seasonIDs:            make(chan int64, 1024),
		subSnMap:             make(map[string]*databus.Databus),
		statCh:               make(chan interface{}, 10240),
	}
	var err error
	if s.arcClient, err = arcApi.NewClient(s.c.ArcClient); err != nil {
		panic(fmt.Sprintf("archive-service grpc new error(%v)", err))
	}
	if s.seasonClient, err = seasonApi.NewClient(s.c.SeasonClient); err != nil {
		panic(fmt.Sprintf("ugc-service grpc new error(%+v)", err))
	}
	s.waiter.Add(1)
	go s.seasonConsumer()
	s.waiter.Add(1)
	go s.retryproc()
	s.subSnMap[stat.TypeForView] = databus.New(c.ViewSnSub)
	s.subSnMap[stat.TypeForDm] = databus.New(c.DmSnSub)
	s.subSnMap[stat.TypeForReply] = databus.New(c.ReplySnSub)
	s.subSnMap[stat.TypeForFav] = databus.New(c.FavSnSub)
	//s.subSnMap[stat.TypeForCoin] = databus.New(c.CoinSnSub)
	s.subSnMap[stat.TypeForShare] = databus.New(c.ShareSnSub)
	s.subSnMap[stat.TypeForLike] = databus.New(c.LikeSnSub)
	for i := int64(0); i < _sharding; i++ {
		s.waiterSeason.Add(1)
		go s.statSnDealproc()
	}
	for k, d := range s.subSnMap {
		s.waiterSeason.Add(1)
		go s.consumerSnproc(k, d)
	}
	if s.c.Custom.Flush {
		s.waiter.Add(1)
		go s.flushSeason()
	}
	s.startDatabus(c)
	return s
}

func (s *Service) startDatabus(c *conf.Config) {
	//投币
	s.CoinSnSubV2 = railgun.NewRailGun("投币", s.c.CoinSnSubV2Config.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.CoinSnSub}),
		railgun.NewSingleProcessor(c.CoinSnSubV2Config.SingleConfig, s.SnUnpack, s.CoinSnUpDo))
	s.CoinSnSubV2.Start()
}

// Close kafaka consumer close.
func (s *Service) Close() (err error) {
	s.closeSub = true
	s.closeRetry = true
	time.Sleep(2 * time.Second)
	for k, d := range s.subSnMap { // close season stat data bus
		d.Close()
		log.Info("databusSn (%s) cloesed", k)
	}
	close(s.statCh) // close season stat channel
	s.waiterSeason.Wait()
	s.seasonSub.Close()
	s.seasonWithArchivePub.Close()
	s.waiter.Wait()
	//rail_gun
	s.CoinSnSubV2.Close()
	return
}

func (s *Service) flushSeason() {
	defer s.waiter.Done()
	maxSid, err := s.resultDao.MaxSeasonID(context.Background())
	if err != nil {
		log.Error("flushSeason s.resultDao.SeasonIDs err(%+v)", err)
		return
	}
	for sid := maxSid; sid > 0; sid-- {
		log.Info("flushSeason sid(%d)", sid)
		s.seasonUpdate(sid)
		time.Sleep(10 * time.Millisecond)
	}
	log.Info("flushSeason end")
}
