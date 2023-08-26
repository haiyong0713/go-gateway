package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/admin/model"
	"go-gateway/app/web-svr/appstatic/admin/util"
)

const (
	_retry = 3
	_sleep = 100 * time.Millisecond
)

// GendDiff picks the already generated diff packages
func (s *Service) GendDiff(resID int) (generated map[int64]int64, err error) {
	generated = make(map[int64]int64)
	genVers := []*model.Ver{}
	if err = s.DB.Where("file_type IN (1,2)"). // 1=diff pkg, 2=diff pkg calculation in progress
							Where("is_deleted = 0").Where("resource_id = ?", resID).Select("id, from_ver").Find(&genVers).Error; err != nil {
		log.Error("generatedDiff Error %v", err)
		return
	}
	for _, v := range genVers {
		generated[v.FromVer] = v.ID
	}
	return
}

// Publish returns the second trigger result
func (s *Service) Publish(ctx context.Context, resID int) (data *model.PubResp, err error) {
	var (
		prodVers, testVers []int64 // the history versions that we should generate for
		currRes            *model.Resource
		generated          map[int64]int64
		prodMore, testMore []int64
	)
	// pick history versions to calculate diff
	if prodVers, testVers, currRes, err = s.pickDiff(resID); err != nil {
		return
	}
	// pick already generated diff packages
	if generated, err = s.GendDiff(resID); err != nil {
		return
	}
	// filter already generated
	for _, v := range prodVers {
		if _, ok := generated[v]; !ok {
			prodMore = append(prodMore, v)
		}
	}
	for _, v := range testVers {
		if _, ok := generated[v]; !ok {
			testMore = append(testMore, v)
		}
	}
	// put diff packages in our DB
	if err = s.putDiff(resID, mergeSlice(prodMore, testMore), currRes); err != nil {
		return
	}
	data = &model.PubResp{
		CurrVer:  currRes.Version,
		DiffProd: prodMore,
		DiffTest: testMore,
	}
	return
}

// Retry . retry one function until no error
func Retry(callback func() error, retry int, sleep time.Duration) (err error) {
	for i := 0; i < retry; i++ {
		if err = callback(); err == nil {
			return
		}
		time.Sleep(sleep)
	}
	return
}

// Push .
func (s *Service) Push(ctx context.Context, uname string) (err error) {
	var (
		timeValue int64
	)
	if timeValue, err = s.dao.PushTime(ctx); err != nil {
		return err
	}
	if timeValue != 0 {
		return fmt.Errorf("20分钟内仅能发起一次推送")
	}
	if err = s.dao.CallPush(ctx); err != nil {
		log.Error("Push_CallPush error(%v) user(%s)", err, uname)
		return
	}
	log.Info("Push Success user(%s) time(%d)", uname, time.Now().Unix())
	s.cache.Do(ctx, func(ctx context.Context) {
		if err = Retry(func() (err error) {
			return s.dao.AddPushTime(ctx, time.Now().Unix())
		}, _retry, _sleep); err != nil {
			log.Error("Push_AddPushTimeToRedis (%d) error(%v) user(%s)", time.Now().Unix(), err, uname)
			return
		}
		msg := fmt.Sprintf("Push time is %d User is %s", time.Now().Unix(), uname)
		_ = util.AddLogs(uname, msg)
	})
	return
}
