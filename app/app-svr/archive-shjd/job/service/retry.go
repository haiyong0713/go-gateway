package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-shjd/job/model"

	"github.com/pkg/errors"
)

const (
	_retryList = "retry_list"
)

func (s *Service) retryconsumer() {
	defer s.waiter.Done()
	for {
		if s.close {
			log.Info("retryconsumer closed")
			return
		}
		var (
			c    = context.TODO()
			err  error
			item *model.RetryItem
		)
		if item, err = s.PopItem(c); err != nil {
			log.Error("%+v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if item == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		log.Info("get retry item(%+v)", item)
		switch item.Tp {
		case model.TypeForDelVideo:
			s.DelVideoCache(context.Background(), item.AID, item.CID)
		case model.TypeForUpdateVideo:
			s.UpdateVideoCache(context.Background(), item.AID, item.CID)
		case model.TypeForUpdateArchive:
			s.UpdateCache(item.Old, item.Nw, item.Action)
		case model.TypeForVideoShot:
			s.addVideoShotCache(context.Background(), item.CID, item.Count, item.HdCnt, item.SdCnt, item.HdImg, item.SdImg)
		case model.TypeForInternal:
			s.internalCacheHandler(context.Background(), item.AID)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// PushItem is
func (s *Service) PushItem(c context.Context, item *model.RetryItem) (err error) {
	conn := s.rds.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(item)
	if err != nil {
		err = errors.Wrap(err, "json.Marshal")
		return
	}
	if _, err = conn.Do("RPUSH", _retryList, bs); err != nil {
		err = errors.Wrap(err, "conn.Send(RPUSH)")
		return
	}
	return
}

// PopItem is
func (s *Service) PopItem(c context.Context) (item *model.RetryItem, err error) {
	conn := s.rds.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("LPOP", _retryList))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		err = errors.WithStack(err)
		return
	}
	item = &model.RetryItem{}
	if err = json.Unmarshal(bs, item); err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
