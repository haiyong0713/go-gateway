package rank

import (
	"context"
	"fmt"
	"go-common/library/log"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v3"
	"time"
)

const (
	weekDiff  = 7
	monthDiff = 31
)

// getOnlineRankBase 获取有效排行配置
func (s *Service) getOnlineRankBase(c context.Context) (rank []*rankmdl.Base, err error) {
	return s.dao.GetBaseOnline(c, time.Now())
}

// getOnlineRule 获取有效子榜
func (s *Service) getOnlineRule(c context.Context) (rank []*rankmdl.Rule, err error) {
	return s.dao.GetRuleOnline(c, time.Now())
}

func (s *Service) getDate(c context.Context, day time.Time) string {
	return day.Format("20060102")
}

func (s *Service) getTodayTime(c context.Context, day time.Time) (int64, error) {
	t, err := time.ParseInLocation("2006-01-02", day.Format("2006-01-02"), time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
func (s *Service) getDay(c context.Context, timestamp int64) int {
	return time.Unix(timestamp, 0).Day()
}

func (s *Service) getToday(c context.Context, timestamp int64) string {
	return time.Unix(timestamp, 0).Format("20060102")
}

// needRank ...
func (s *Service) needRank(c context.Context, startTimestamp int64, frequency int, scope int) (bool, string, error) {

	startTime, err := s.getTodayTime(c, time.Unix(startTimestamp, 0))

	if err != nil {
		return false, "", err
	}
	todayTime, err := s.getTodayTime(c, time.Unix(time.Now().Unix(), 0))
	if err != nil {
		return false, "", err
	}
	diffTime := todayTime - startTime
	stime := s.getDate(c, time.Unix(startTimestamp, 0))
	diff := diffTime / 86400
	if diff == 0 {
		return true, "", nil
	}
	switch frequency {
	case rankmdl.FrequencyTypeDay:
		if scope == rankmdl.UpdateScopeIncrement {
			return true, s.getDate(c, time.Now().AddDate(0, 0, -1)), nil
		}
		return true, stime, nil

	case rankmdl.FrequencyTypeWeek:
		if diff%weekDiff == 0 {
			if scope == rankmdl.UpdateScopeIncrement {
				return true, s.getDate(c, time.Now().AddDate(0, 0, -7)), nil
			}
			return true, stime, nil
		}
		return false, "", nil
	case rankmdl.FrequencyTypeMonth:
		if s.getDay(c, startTimestamp) == s.getDay(c, time.Now().Unix()) {
			if scope == rankmdl.UpdateScopeIncrement {
				return true, s.getDate(c, time.Now().AddDate(0, -1, 0)), nil
			}
			return true, stime, nil
		}
	case rankmdl.FrequencyTypeOnce:
		if s.getToday(c, startTimestamp) == s.getToday(c, time.Now().Unix()) {
			return true, stime, nil
		}
	}
	return false, "", nil
}

// getNeedRankLog 计算当日需要统计的排行榜
func (s *Service) getNeedRankLog(c context.Context, rule []*rankmdl.Rule) []*rankmdl.Log {
	thisDate := s.getDate(c, time.Now().AddDate(0, 0, -1))
	needRankLog := make([]*rankmdl.Log, 0)
	if rule != nil && len(rule) > 0 {
		for _, v := range rule {
			need, lastTime, err := s.needRank(c, int64(v.Stime), v.UpdateFrequency, v.UpdateScope)
			if err != nil {
				log.Errorc(c, "getNeedRankLog rule_id (%d) err(%v)", v.ID, err)
				continue
			}
			if need {
				needRankLog = append(needRankLog, &rankmdl.Log{
					BaseID:   v.BaseID,
					RankID:   v.ID,
					Batch:    v.LastBatch + 1,
					ThisDate: thisDate,
					LastDate: lastTime,
				})
			}
		}
	}
	return needRankLog
}

// SetRankLog 设置需要计算排行的log
func (s *Service) SetRankLog(ctx context.Context) (err error) {
	err = s.setRankLog(ctx)
	if err != nil {
		log.Errorc(ctx, "s.setRankLog err(%v)", err)
		err = s.sendWechat(ctx, "[排行榜 NEW SetRankLog]", fmt.Sprintf("%v", err), "zhangtinghua", s.c.WechatToken.LittleFlower)
		if err != nil {
			log.Errorc(ctx, "s.sendWechat err(%v)", err)
			return
		}
	}
	return
}

// setRankLog ...
func (s *Service) setRankLog(ctx context.Context) (err error) {
	rule, err := s.getOnlineRule(ctx)
	if err != nil {
		log.Errorc(ctx, "SetRankLog s.getOnlineRule err(%v)", err)
		return
	}
	needlog := s.getNeedRankLog(ctx, rule)
	if len(needlog) == 0 {
		return
	}

	rankIDs := make([]int64, 0)
	for _, v := range needlog {
		rankIDs = append(rankIDs, v.RankID)
	}
	historyLog, err := s.dao.GetRankLog(ctx, rankIDs, s.getDate(ctx, time.Now().AddDate(0, 0, -1)))
	if err != nil {
		log.Errorc(ctx, "SetRankLog s.dao.GetRankLog err(%v)", err)
		return
	}
	historyRankMap := make(map[int64]struct{})
	if historyLog != nil {
		for _, v := range historyLog {
			historyRankMap[v.RankID] = struct{}{}
		}
	}
	newLog := make([]*rankmdl.Log, 0)
	for _, v := range needlog {
		if _, ok := historyRankMap[v.RankID]; !ok {
			newLog = append(newLog, v)
			historyRankMap[v.RankID] = struct{}{}
		}
	}
	if len(newLog) == 0 {
		return
	}
	err = s.dao.InsertRankLog(ctx, newLog)
	if err != nil {
		log.Errorc(ctx, "SetRankLog s.dao.InsertRankLog")
		return
	}
	return nil
}
