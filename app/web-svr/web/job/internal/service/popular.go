package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/web-svr/web/job/internal/model"
)

func (s *Service) setPopularSeries() {
	ctx := context.Background()
	data, err := s.dao.PopularSeries(ctx)
	if err != nil {
		log.Error("日志告警 获取每周必看数据错误,error:%+v", err)
		return
	}
	var configs []*model.MgrSeriesConfig
	listm := map[int64][]*model.MgrSeriesList{}
	for _, val := range data {
		if val == nil || val.Config == nil || val.Config.ID <= 0 || val.Config.Number <= 0 || len(val.List) == 0 {
			bs, _ := json.Marshal(val)
			log.Error("日志告警 发现每周必看数据错误,data:%s", bs)
			continue
		}
		configs = append(configs, val.Config)
		listm[val.Config.ID] = val.List
	}
	if err := retry(func() error {
		return s.dao.AddCacheSeriesDetail(ctx, listm)
	}); err != nil {
		log.Error("日志告警 设置每周必看缓存错误,error:%+v", err)
		return
	}
	if err := retry(func() error {
		return s.dao.AddCacheSeries(ctx, "weekly_selected", configs)
	}); err != nil {
		log.Error("日志告警 设置每周必看缓存错误,data:%+v,error:%+v", configs, err)
		return
	}
	log.Info("设置每周必看缓存成功")
}

func (s *Service) initActPlatRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.actPlatRailGunUnpack, s.actPlatRailGunDo)
	g := railgun.NewRailGun("活动平台历史记录databus", nil, inputer, processor)
	s.actPlatRailGun = g
	g.Start()
}

func (s *Service) actPlatRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	actMsg := &model.ActPlatHistory{}
	if err := json.Unmarshal(msg.Payload(), &actMsg); err != nil {
		log.Error("Failed to Unmarshal ActPlatHistory: %+v", err)
		return nil, err
	}
	switch actMsg.Activity {
	case s.ac.PopularAct.Activity:
		return &railgun.SingleUnpackMsg{
			Group: actMsg.Mid,
			Item:  actMsg,
		}, nil
	default:
		return nil, nil
	}
}

const (
	_popularActStep0 = 1
	_popularActStep1 = 9
	_popularActStep4 = 85
)

func (s *Service) actPlatRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	actMsg, ok := item.(*model.ActPlatHistory)
	if !ok || actMsg.Counter != s.ac.PopularAct.Counter {
		return railgun.MsgPolicyIgnore
	}
	if actMsg.Mid <= 0 || actMsg.Diff <= 0 {
		return railgun.MsgPolicyIgnore
	}
	switch actMsg.Total {
	case _popularActStep0:
		if _, err := s.dao.AddPopularWatchTime(ctx, actMsg.Mid, 0, timestampToTime(actMsg.TimeStamp)); err != nil {
			log.Error("Failed to AddPopularWatchTime: %d, %d, %+v", actMsg.Mid, 0, err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DelCachePopularWatchTime(ctx, actMsg.Mid, 0); err != nil {
			log.Error("Failed to DelCachePopularWatchTime: %d, %d, %+v", actMsg.Mid, 0, err)
			return railgun.MsgPolicyIgnore
		}
	case _popularActStep1:
		if _, err := s.dao.AddPopularWatchTime(ctx, actMsg.Mid, 1, timestampToTime(actMsg.TimeStamp)); err != nil {
			log.Error("Failed to AddPopularWatchTime: %d, %d, %+v", actMsg.Mid, 1, err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DelCachePopularWatchTime(ctx, actMsg.Mid, 1); err != nil {
			log.Error("Failed to DelCachePopularWatchTime: %d, %d, %+v", actMsg.Mid, 1, err)
			return railgun.MsgPolicyIgnore
		}
	case _popularActStep4:
		// 数据导入后开放排名表写入
		if s.ac.PopularAct.RankSwitch {
			if _, err := s.dao.AddPopularRank(ctx, actMsg.Mid); err != nil {
				log.Error("Failed to AddPopularRank: %d, %+v", actMsg.Mid, err)
			}
			if err := s.dao.DelCachePopularRank(ctx, actMsg.Mid); err != nil {
				log.Error("Failed to DelCachePopularRank: %d, %+v", actMsg.Mid, err)
			}
		}
		if _, err := s.dao.AddPopularWatchTime(ctx, actMsg.Mid, 4, time.Now()); err != nil {
			log.Error("Failed to AddPopularWatchTime: %d, %d, %+v", actMsg.Mid, 4, err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DelCachePopularWatchTime(ctx, actMsg.Mid, 4); err != nil {
			log.Error("Failed to DelCachePopularWatchTime: %d, %d, %+v", actMsg.Mid, 4, err)
			return railgun.MsgPolicyIgnore
		}
	}
	return railgun.MsgPolicyNormal
}

func timestampToTime(in int64) time.Time {
	return time.Unix(in, 0)
}
