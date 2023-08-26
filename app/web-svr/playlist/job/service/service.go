package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/stat/prom"
	plclient "go-gateway/app/web-svr/playlist/interface/api/v1"
	plmdl "go-gateway/app/web-svr/playlist/interface/model"
	"go-gateway/app/web-svr/playlist/job/conf"
	"go-gateway/app/web-svr/playlist/job/dao"
	"go-gateway/app/web-svr/playlist/job/model"
)

const (
	_sharding = 10 // goroutines for dealing the stat
	_chanSize = 10240
)

// Service .
type Service struct {
	c                *conf.Config
	dao              *dao.Dao
	waiter           *sync.WaitGroup
	closed           bool
	playlistViewSub  *databus.Databus
	playlistFavSub   *databus.Databus
	playlistReplySub *databus.Databus
	playlistShareSub *databus.Databus
	updateDbInterval int64
	statCh           [_sharding]chan *model.StatM
	plClient         plclient.PlaylistClient
}

// New creates a Service instance.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                c,
		dao:              dao.New(c),
		waiter:           new(sync.WaitGroup),
		playlistViewSub:  databus.New(c.PlaylistViewSub),
		playlistFavSub:   databus.New(c.PlaylistFavSub),
		playlistReplySub: databus.New(c.PlaylistReplySub),
		playlistShareSub: databus.New(c.PlaylistShareSub),
		updateDbInterval: int64(time.Duration(c.Job.UpdateDbInterval) / time.Second),
	}
	var err error
	if s.plClient, err = plclient.NewClient(c.PlClient); err != nil {
		panic(err)
	}
	for i := int64(0); i < _sharding; i++ {
		// for stat
		s.statCh[i] = make(chan *model.StatM, _chanSize)
		s.waiter.Add(1)
		go s.viewproc(i)
	}
	s.waiter.Add(1)
	go s.consumeView()
	s.waiter.Add(1)
	go s.consumeFav()
	s.waiter.Add(1)
	go s.consumeReply()
	s.waiter.Add(1)
	go s.consumeShare()
	return
}

// consumeView consumes playlist's view.
func (s *Service) consumeView() {
	defer s.waiter.Done()
	for {
		if s.closed {
			for i := 0; i < _sharding; i++ {
				close(s.statCh[i])
			}
			return
		}
		msg, ok := <-s.playlistViewSub.Messages()
		if !ok {
			log.Info("databus: playlist-job view consumer exit!")
			time.Sleep(10 * time.Millisecond)
			continue
		}
		msg.Commit()
		viewSM := &model.StatM{}
		if err := json.Unmarshal(msg.Value, viewSM); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if viewSM.Type != plmdl.PlDBusType || viewSM.ID <= 0 {
			continue
		}
		key := viewSM.ID % _sharding
		s.statCh[key] <- viewSM
		prom.BusinessInfoCount.State(fmt.Sprintf("statChan-%v", key), int64(len(s.statCh[key])))
		log.Info("consumeView key:%s partition:%d offset:%d msg: %v)", msg.Key, msg.Partition, msg.Offset, viewSM.String(model.ViewCountType))
	}
}

// consumeFav  consumes playlist's favorite.
func (s *Service) consumeFav() {
	defer s.waiter.Done()
	var c = context.TODO()
	for {
		msg, ok := <-s.playlistFavSub.Messages()
		if !ok {
			log.Info("databus: playlist-job favorite consumer exit!")
			return
		}
		msg.Commit()
		favSM := &model.StatM{}
		if err := json.Unmarshal(msg.Value, favSM); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if favSM.Type != plmdl.PlDBusType || favSM.ID <= 0 {
			continue
		}
		s.upStat(c, favSM, model.FavCountType)
		log.Info("consumeFav key:%s partition:%d offset:%d msg: %v)", msg.Key, msg.Partition, msg.Offset, favSM.String(model.FavCountType))
	}
}

// consumeReply  consumes playlist's reply.
func (s *Service) consumeReply() {
	defer s.waiter.Done()
	var c = context.TODO()
	for {
		msg, ok := <-s.playlistReplySub.Messages()
		if !ok {
			log.Info("databus: playlist-job reply consumer exit!")
			return
		}
		msg.Commit()
		replySM := &model.StatM{}
		if err := json.Unmarshal(msg.Value, replySM); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if replySM.Type != plmdl.PlDBusType || replySM.ID <= 0 {
			continue
		}
		s.upStat(c, replySM, model.ReplyCountType)
		log.Info("consumeReply key:%s partition:%d offset:%d msg: %v)", msg.Key, msg.Partition, msg.Offset, replySM.String(model.ReplyCountType))
	}
}

// consumeShare  consumes playlist's share.
func (s *Service) consumeShare() {
	defer s.waiter.Done()
	var c = context.TODO()
	for {
		msg, ok := <-s.playlistShareSub.Messages()
		if !ok {
			log.Info("databus: playlist-job share consumer exit!")
			return
		}
		msg.Commit()
		shareSM := &model.StatM{}
		if err := json.Unmarshal(msg.Value, shareSM); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if shareSM.Type != plmdl.PlDBusType || shareSM.ID <= 0 {
			continue
		}
		s.upStat(c, shareSM, model.ShareCountType)
		log.Info("consumeShare key:%s partition:%d offset:%d msg: %v)", msg.Key, msg.Partition, msg.Offset, shareSM.String(model.ShareCountType))
	}
}

// Ping reports the heath of services.
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close releases resources which owned by the Service instance.
func (s *Service) Close() (err error) {
	defer s.waiter.Wait()
	s.closed = true
	s.playlistViewSub.Close()
	s.playlistFavSub.Close()
	s.playlistReplySub.Close()
	s.playlistShareSub.Close()
	log.Info("playlist-job has been closed.")
	return
}
