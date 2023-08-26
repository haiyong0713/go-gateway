package service

import (
	"context"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/model/college"
	"math/rand"
	"time"
)

const (
	// BonusKey ...
	BonusKey               = "bonus"
	concurrencySetMidScore = 100
)

var collegeBonusCtx context.Context

func collegeBonusCtxInit() {
	collegeBonusCtx = trace.SimpleServerTrace(context.Background(), "collegeBonus")
}

// CollegeBonus 额外加分
func (s *Service) CollegeBonus() {
	s.collegeBonusRunning.Lock()
	defer s.collegeBonusRunning.Unlock()
	start := time.Now()
	collegeBonusCtxInit()
	log.Errorc(collegeBonusCtx, "college bonus start (%d)", start.Unix())
	collegeList, err := s.getAllCollege(collegeBonusCtx)
	if err != nil {
		log.Errorc(collegeBonusCtx, "s.getAllCollege(c) err(%v)", err)
		return
	}
	err = s.collegeAllTagArchive(collegeBonusCtx, collegeList)
	if err != nil {
		log.Errorc(collegeBonusCtx, "s.collegeAllArchive(c) err(%v)", err)
		return
	}
	log.Errorc(collegeBonusCtx, "CollegeBonus success()")

}

func (s *Service) collegeAllTagArchive(c context.Context, collegeList []*college.College) (err error) {
	for _, v := range collegeList {
		if v.TagID > 0 {
			midInfo, err := s.getCollegeAllMid(c, v.ID)
			if err != nil {
				log.Errorc(c, "s.getCollegeAllMid(%d)", v.ID)
				return err
			}
			mids := make(map[int64]struct{})
			if midInfo != nil && len(midInfo) > 0 {
				for _, v := range midInfo {
					mids[v.MID] = struct{}{}
				}
			}
			if len(mids) > 0 {
				err = s.collegeTagAllArchive(c, v, mids)
				if err != nil {
					log.Errorc(c, "s.collegeTagArchive collegeID(%d) err(%v)", v.ID, err)
					return err
				}
			}
		}

	}
	return nil
}

// collegeTagArchive tag下的稿件
func (s *Service) collegeTagAllArchive(c context.Context, collegeInfo *college.College, mids map[int64]struct{}) (err error) {
	aidCh := make(chan []int64, aidChannelLength)
	arcCh := make(chan *api.ArcsReply, archiveChannelLength)
	eg := errgroup.WithContext(c)
	var (
		midCtime []*college.MIDCtime
	)
	eg.Go(func(ctx context.Context) (err error) {
		err = s.collegeArchiveIntoChannel(c, collegeInfo.TagID, aidCh)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		midCtime, err = s.collegeTagArchiveDetail(c, collegeInfo, aidCh, arcCh, mids)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	if midCtime != nil && len(midCtime) > 0 {
		err = s.concurrencySendCollegeMidScore(c, midCtime)
	}
	return err
}

// sendPoint
func (s *Service) sendPoint(c context.Context, points int64, timesstamp int64, mid int64, source int64, business string) error {
	data := &college.ActPlatActivityPoints{
		Points:    points,
		Mid:       mid,
		Source:    source,
		Activity:  s.c.College.MidActivity,
		Business:  business,
		Timestamp: timesstamp,
	}
	err := s.college.SendPoint(c, mid, data)
	if err != nil {
		log.Errorc(c, " s.college.SendPoint(%d,%v)", mid, *data)
		return err
	}
	return nil
}

// concurrencySendCollegeMidScore 并发给mid加分
func (s *Service) concurrencySendCollegeMidScore(c context.Context, midCtime []*college.MIDCtime) (err error) {
	var times int
	concurrency := concurrencySetMidScore
	times = len(midCtime) / concurrency
	for index := 0; index <= times; index++ {
		// 这个轮次的开始时
		startTime := time.Now().UnixNano() / 1e6
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*concurrency + b
				if start >= len(midCtime) {
					return nil
				}
				reqMid := midCtime[start].MID
				rand.Seed(time.Now().UnixNano())
				timeStamp := time.Now().Unix() + rand.Int63n(1000)
				aid := midCtime[start].AID
				err := s.sendPoint(c, s.c.College.VideoBonusPoint, timeStamp, reqMid, aid, BonusKey)
				if err != nil {
					log.Errorc(c, "s.sendPoint(%d,%d,%d)", timeStamp, reqMid, aid)
				}
				return err
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			// return nil, err
		}
		endTime := time.Now().UnixNano() / 1e6
		waitTime := s.getWaitTime(startTime, endTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
	}
	return err
}

func (s *Service) collegeTagArchiveDetail(c context.Context, collegeInfo *college.College, aidCh chan []int64, arcCh chan *api.ArcsReply, mids map[int64]struct{}) ([]*college.MIDCtime, error) {
	midCtime := make([]*college.MIDCtime, 0)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		defer close(arcCh)
		for v := range aidCh {
			aids := v
			if len(aids) > 0 {
				err = s.archiveInfo(c, aids, arcCh)
			}
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		for v := range arcCh {
			if v == nil || v.Arcs == nil {
				err = ecode.ActivityWriteHandArchiveErr
			}
			for _, arc := range v.Arcs {
				if arc == nil {
					err = ecode.ActivityWriteHandArchiveErr
				}
				// 防止aid重复
				if arc.IsNormal() {
					ctime := arc.Ctime.Time().Unix()
					if ctime < s.c.College.ArchiveCtime {
						continue
					}
					if arc.Stat.View >= s.c.College.ArchiveVideoState {
						if _, midOk := mids[arc.Author.Mid]; midOk {
							midCtime = append(midCtime, &college.MIDCtime{MID: arc.Author.Mid, Ctime: arc.Ctime.Time().Unix(), AID: arc.Aid})
						}
					}
				}
			}
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	return midCtime, nil
}
