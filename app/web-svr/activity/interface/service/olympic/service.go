package olympic

import (
	"context"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/olympic"
	"strconv"
	"strings"
	"time"
)

var localS *Service

type Service struct {
	conf *conf.Config
	dao  *olympic.Dao
}

func New(c *conf.Config) (s *Service) {
	if localS != nil {
		return localS
	}
	s = &Service{
		conf: c,
		dao:  olympic.New(c),
	}
	localS = s
	return
}

func (s *Service) GetOlympicContestDetail(ctx context.Context, id int64, skipCache bool) (contest *pb.GetOlympicContestDetailResp, err error) {
	contest = new(pb.GetOlympicContestDetailResp)
	resp, err := s.dao.GetOlympicContest(ctx, id, s.conf.OlympicConf.ContestSourceId, skipCache, skipCache)
	if err != nil {
		return
	}
	// 转时间
	stime := int64(0)
	if resp.Stime != "" {
		timeValue, errG := time.ParseInLocation("2006-01-02 15:04:05", resp.Stime, time.Local)
		if errG != nil {
			log.Errorc(ctx, "[GetOlympicContestDetail][ParseInLocation][Error], err:%+v", errG)
		} else {
			stime = timeValue.Unix()
		}
	}
	contest = &pb.GetOlympicContestDetailResp{
		Id:            resp.Id,
		GameStage:     resp.GameStage,
		Stime:         stime,
		HomeTeamName:  resp.HomeTeamName,
		AwayTeamName:  resp.AwayTeamName,
		HomeTeamUrl:   resp.HomeTeamUrl,
		AwayTeamUrl:   resp.AwayTeamUrl,
		HomeScore:     formatScore(ctx, resp.HomeScore),
		AwayScore:     formatScore(ctx, resp.AwayScore),
		ContestStatus: int64(formatContestStatus(resp.ContestStatus)),
		SeasonTitle:   resp.SeasonTitle,
		SeasonUrl:     resp.SeasonUrl,
		VideoUrl:      resp.VideoUrl,
		BottomUrl:     resp.BottomUrl,
		ShowRule:      s.formatShowRule(ctx, resp.ShowRule),
	}
	return
}

func (s *Service) formatShowRule(ctx context.Context, rule string) (showRule int32) {
	showRule = 0
	if rule == "" {
		return
	}
	rule = strings.Trim(rule, " ")
	value, err := strconv.ParseInt(rule, 10, 32)
	if err == nil {
		showRule = int32(value)
	} else {
		log.Errorc(ctx, "[Olympic][formatScore][Error], err:%+v", err)
	}
	return
}

func (s *Service) GetOlympicQueryConfigs(ctx context.Context, skipCache bool) (resp *pb.GetOlympicQueryConfigResp, err error) {
	resp = new(pb.GetOlympicQueryConfigResp)
	resp.QueryConfigs = make([]*pb.OlympicQueryConfig, 0)
	queryConfigs, err := s.dao.GetQueryConfigs(ctx, s.conf.OlympicConf.QuerySourceId, skipCache)
	if err != nil {
		return
	}
	for _, queryConfig := range queryConfigs {
		resp.QueryConfigs = append(resp.QueryConfigs, &pb.OlympicQueryConfig{
			ContestId: queryConfig.MatchId,
			QueryWord: queryConfig.QueryWord,
			State:     queryConfig.State,
		})
	}
	return
}

func formatScore(ctx context.Context, score string) (scoreInt int64) {
	scoreInt = 0
	if score == "" {
		return
	}
	score = strings.Trim(score, " ")
	value, err := strconv.ParseInt(score, 10, 64)
	if err == nil {
		scoreInt = value
	} else {
		log.Errorc(ctx, "[Olympic][formatScore][Error], err:%+v", err)
	}
	return
}

func formatContestStatus(contestStatus string) int {
	switch contestStatus {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	default:
		return 1
	}
}
