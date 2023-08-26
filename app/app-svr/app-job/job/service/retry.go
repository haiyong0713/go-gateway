package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-job/job/model"
	"go-gateway/app/app-svr/app-job/job/model/space"
)

func (s *Service) retryproc() {
	defer s.waiter.Done()
	var (
		bs  []byte
		err error
	)
	c := context.Background()
	retry := &model.Retry{}
	for {
		if s.closed {
			break
		}
		if bs, err = s.vdao.PopFail(c); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if len(bs) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		if err = json.Unmarshal(bs, retry); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", bs, err)
			continue
		}
		log.Info("retry action(%s) data(%s)", retry.Action, bs)
		switch retry.Action {
		case model.ActionUpContribute:
			if retry.Data.Mid != 0 {
				s.updateContribute(retry.Data.Mid, retry.Data.Attrs, retry.Data.Items, retry.Data.IsCooperation, retry.Data.IsComic)
			}
		case model.ActionUpContributeAid:
			if retry.Data.Mid != 0 {
				_ = s.contributeCache(retry.Data.Mid, retry.Data.Attrs, retry.Data.Time, retry.Data.IP, retry.Data.IsCooperation, retry.Data.IsComic, "retry")
			}
		}
	}
}

// nolint:unparam
func retry(callback func() error, retry int, sleep time.Duration) (err error) {
	for i := 0; i < retry; i++ {
		if err = callback(); err == nil {
			return
		}
		time.Sleep(sleep)
	}
	return
}

func (s *Service) syncRetry(c context.Context, action string, mid, aid int64, attrs *space.Attrs, items []*space.Item, time xtime.Time, ip string, isCooperation, isComic bool) (err error) {
	retry := &model.Retry{Action: action}
	retry.Data.Mid = mid
	retry.Data.Aid = aid
	retry.Data.Attrs = attrs
	retry.Data.Items = items
	retry.Data.Time = time
	retry.Data.IP = ip
	retry.Data.IsCooperation = isCooperation
	retry.Data.IsComic = isComic
	return s.vdao.PushFail(c, retry)
}
