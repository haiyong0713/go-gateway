package service

import (
	"context"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/esports/admin/model"
	espclient "go-gateway/app/web-svr/esports/interface/api/v1"
)

func (s *Service) RenewPosterFromTeam(ctx context.Context) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	var succeedNum, failedNum int64
	failedIDList := make([]int64, 0)
	list := make([]*model.Team, 0)
	err = s.dao.DB.Find(&list).Error
	if err != nil {
		return
	}

	succeedNum, failedNum, failedIDList = s.RenewPosterByTeamList(ctx, list)
	{
		m["succeedNum"] = succeedNum
		m["failedNum"] = failedNum
		m["failedIDList"] = failedIDList
	}

	return
}

func (s *Service) RenewPosterByTeamList(ctx context.Context, list []*model.Team) (succeedNum, failedNum int64, failedIDList []int64) {
	failedIDList = make([]int64, 0)
	for _, v := range list {
		var tmpErr error
		for i := 0; i < 3; i++ {
			tmpErr = s.RenewPosterByTeam(ctx, v)
			if tmpErr == nil {
				break
			}
		}

		if tmpErr != nil {
			failedNum++
			failedIDList = append(failedIDList, v.ID)
		} else {
			succeedNum++
		}
	}

	return
}

func (s *Service) RenewPosterByTeamID(ctx context.Context, list []int64) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	var succeedNum, failedNum int64
	failedIDList := make([]int64, 0)
	teamList := make([]*model.Team, 0)
	err = s.dao.DB.Where("id in (?)", list).Find(&teamList).Error
	if err != nil {
		return
	}

	succeedNum, failedNum, failedIDList = s.RenewPosterByTeamList(ctx, teamList)
	{
		m["succeedNum"] = succeedNum
		m["failedNum"] = failedNum
		m["failedIDList"] = failedIDList
	}

	return
}

func (s *Service) RenewPosterByTeam(ctx context.Context, teamInfo *model.Team) error {
	contests, err := s.listContest(teamInfo.ID)
	if err != nil {
		return err
	}

	if len(contests) == 0 {
		return nil
	}

	var cids []int64
	var contestMap = make(map[int64]*model.Contest)
	for _, c := range contests {
		cids = append(cids, c.ID)
		contestMap[c.ID] = c
	}

	rep, err := s.espClient.LiveContests(context.Background(), &espclient.LiveContestsRequest{
		Cids: cids,
	})
	if err != nil {
		return err
	}

	if rep != nil && len(rep.Contests) > 0 {
		eg := errgroup.WithContext(ctx)
		for _, c := range rep.Contests {
			if c.GameState <= 2 {
				continue
			}
			contest := contestMap[c.ID]
			eg.Go(func(c context.Context) error {
				return s.DrawPost(c, teamInfo, contest)
			})
		}

		err = eg.Wait()
	}

	return err
}
