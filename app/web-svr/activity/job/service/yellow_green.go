package service

import (
	"context"
	"time"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model/like"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) YingYuanVote() {
	var (
		ctx                   = context.Background()
		yellowVote, greenVote int64
	)
	nowTime := time.Now()
	log.Infoc(ctx, "YingYuanVote begin %v", nowTime.Format("2006-01-02 15:04:05"))
	for _, period := range s.c.YellowAndGreen.Period {
		eg := errgroup.WithContext(ctx)
		eg.Go(func(ctx context.Context) (err error) {
			if yellowVote, err = s.YingYuanArchive(ctx, period.YellowYingYuanSid, period); err != nil {
				log.Errorc(ctx, "YingYuanVote yellowVote s.YingYuanArchive() sid(%d) error(%+v)", period.YellowYingYuanSid, err)
			}
			return err
		})
		eg.Go(func(ctx context.Context) (err error) {
			if greenVote, err = s.YingYuanArchive(ctx, period.GreenYingYuanSid, period); err != nil {
				log.Errorc(ctx, "YingYuanVote greenVote s.YingYuanArchive() sid(%d) error(%+v)", period.GreenYingYuanSid, err)
			}
			return err
		})
		if err := eg.Wait(); err != nil {
			log.Errorc(ctx, "YingYuanVote greenVote eg.Wait()() period(%+v) error(%+v)", period, err)
			continue
		}
		s.retry(ctx, func() error {
			vote := &like.YgVote{
				YellowVote: yellowVote,
				GreenVote:  greenVote,
			}
			e := s.dao.AddCacheYellowGreenVote(ctx, vote, period)
			if e != nil {
				log.Errorc(ctx, "YingYuanVote s.dao.AddCacheYellowGreenVote vote(%+v)", vote, e)
			}
			return e
		})
	}
	log.Infoc(ctx, "YingYuanVote end %v since(%v)", time.Now().Format("2006-01-02 15:04:05"), time.Since(nowTime).Seconds())
}

func (s *Service) YingYuanArchive(ctx context.Context, sid int64, period *like.YellowGreenPeriod) (res int64, err error) {
	likeArcs, err := s.loadLikeList(ctx, sid, _retryTimes)
	if err != nil {
		log.Error("YingYuanVote YingYuanArchive s.loadLikeList sid(%d) error(%v)", sid, err)
		return
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("actArchives len(aids) == 0")
		return
	}
	for i := 0; i < aidsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("actArchives s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				if v != nil && v.Stat.View >= period.YingYuanView {
					res += period.YingYuanVote
				}
			}
		}
	}
	return
}
