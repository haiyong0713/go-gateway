package service

import (
	"context"
	"encoding/json"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/esports/admin/client"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

func (s *Service) contestSeriesList(ctx context.Context, seasonID, limit, offset int64) (res []*model.ContestSeries, err error) {
	res, err = model.ContestSeriesList(seasonID, limit, offset)
	if err == nil {
		for _, c := range res {
			tc := c
			var g bool
			switch c.Type {
			case 1:
				r, grpcErr := client.EsportsGrpcClient.IsSeriesPointMatchInfoGenerated(ctx, &v1.IsSeriesPointMatchInfoGeneratedReq{
					SeriesId: tc.ID,
				})
				if grpcErr == nil {
					g = r.ViewGenerated
				}
			case 2:
				r, grpcErr := client.EsportsGrpcClient.IsSeriesKnockoutMatchInfoGenerated(ctx, &v1.IsSeriesKnockoutMatchInfoGeneratedReq{
					SeriesId: tc.ID,
				})
				if grpcErr == nil {
					g = r.ViewGenerated
				}
			}

			tc.ViewGenerated = g
		}
	}
	return
}

func (s *Service) DeleteContestSeriesByID(ctx context.Context, id int64) (series *model.ContestSeries, err error) {
	series, err = model.FindContestSeriesByID(id)
	if err != nil {
		return
	}

	err = series.Delete()
	if err != nil {
		return
	}
	s.cache.Do(ctx, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: series.SeasonID, SeriesID: series.ID}); e != nil {
			log.Error("contest component ClearComponentContestCacheGRPC SeasonID(%+v) SeriesID(%d) error(%+v)", series.SeasonID, series.ID, err)
		}
	})
	return
}

func (s *Service) FetchContestSeriesByID(ctx context.Context, id int64) (series *model.ContestSeries, err error) {
	series, err = model.FindContestSeriesByID(id)
	if err == nil {
		switch series.Type {
		case 1:
			g, grpcErr := client.EsportsGrpcClient.IsSeriesPointMatchInfoGenerated(ctx, &v1.IsSeriesPointMatchInfoGeneratedReq{
				SeriesId: series.ID,
			})
			if grpcErr != nil {
				err = grpcErr
				return
			}
			series.ViewGenerated = g.ViewGenerated

		case 2:
			g, grpcErr := client.EsportsGrpcClient.IsSeriesKnockoutMatchInfoGenerated(ctx, &v1.IsSeriesKnockoutMatchInfoGeneratedReq{
				SeriesId: series.ID,
			})
			if grpcErr != nil {
				err = grpcErr
				return
			}
			series.ViewGenerated = g.ViewGenerated
		}

	}
	return
}

func (s *Service) AddContestSeries(ctx context.Context, series *model.ContestSeries) error {
	err := series.Insert()
	if err != nil {
		return err
	}
	// 删除阶段缓存.
	s.cache.Do(ctx, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: series.SeasonID, SeriesID: series.ID}); e != nil {
			log.Error("contest component ClearComponentContestCacheGRPC SeasonID(%+v) SeriesID(%d) error(%+v)", series.SeasonID, series.ID, err)
		}
	})
	return err
}

func (s *Service) UpdateContestSeries(ctx context.Context, series *model.ContestSeries) error {
	err := series.Update()
	if err != nil {
		return err
	}
	// 删除阶段缓存.
	s.cache.Do(ctx, func(c context.Context) {
		if e := s.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: series.SeasonID, SeriesID: series.ID}); e != nil {
			log.Error("contest component ClearComponentContestCacheGRPC SeasonID(%+v) SeriesID(%d) error(%+v)", series.SeasonID, series.ID, err)
		}
	})
	return err
}

func (s *Service) ContestSeriesPaging(ctx context.Context, seasonID, pageSize, pageNum int64) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)

	limit := pageSize
	if limit > 100 {
		limit = 10
	}
	offset := (pageNum - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	list, fetchErr := s.contestSeriesList(ctx, seasonID, limit, offset)
	if fetchErr != nil {
		err = fetchErr

		return
	}

	page := map[string]int64{
		"num":   pageNum,
		"size":  pageSize,
		"total": model.ContestSeriesCount(seasonID),
	}
	m["page"] = page
	m["list"] = list

	return
}

func (s *Service) GetScoreRules(ctx context.Context, seriesId int64) (scoreRules *model.PUBGContestSeriesScoreRule, err error) {
	contestSeries, err := s.dao.GetScoreRuleConfigBySeriesId(ctx, seriesId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段信息失败, 请重试")
		return
	}
	scoreRules = &model.PUBGContestSeriesScoreRule{
		KillScore:  0,
		RankScores: make([]int64, 0),
	}
	if contestSeries.ScoreRuleConfig == "" || contestSeries.ScoreRuleConfig == "null" {
		log.Errorc(ctx, "[Service][ContestSeries][GetScoreRules][ExtraConfig][Empty],contestSeriesId:%d", seriesId)
		return
	}
	if err = json.Unmarshal([]byte(contestSeries.ScoreRuleConfig), &scoreRules); err != nil {
		log.Errorc(ctx, "[Service][ContestSeries][GetScoreRules][ExtraConfig][Unmarshal][Error],err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段积分规则失败，请重试")
	}
	return
}

func (s *Service) SaveScoreRules(ctx context.Context, seriesId int64, rule *model.PUBGContestSeriesScoreRule) (err error) {
	contestSeries, err := s.dao.GetScoreRuleConfigBySeriesId(ctx, seriesId)
	if err != nil || contestSeries == nil {
		log.Errorc(ctx, "[Service][SaveScoreRules][FindContestSeriesByID][Error], err:(%+v), info:(%+v)", err, contestSeries)
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段信息失败, 请重试")
		return
	}
	err = s.dao.ScoreRuleConfigUpdate(ctx, seriesId, rule)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "设置积分规则失败，请重试")
	}
	return
}
