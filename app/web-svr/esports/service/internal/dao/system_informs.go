package dao

import (
	"context"
	"go-common/library/log"
	egV2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

// fetchSeasonTeamsByContest  获取赛程的赛季信息，主队，客队信息
func (d *dao) fetchSeasonTeamsByContest(ctx context.Context, contest *model.ContestModel) (season *model.SeasonModel, teamsMap map[int64]*model.TeamModel, err error) {
	var (
		teamIDs []int64
	)
	if contest.HomeID > 0 {
		teamIDs = append(teamIDs, contest.HomeID)
	}
	if contest.AwayID > 0 {
		teamIDs = append(teamIDs, contest.AwayID)
	}
	group := egV2.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		if len(teamIDs) == 0 {
			return nil
		}
		if teamsMap, err = d.getTeamsMapByID(ctx, teamIDs); err != nil || len(teamsMap) == 0 {
			log.Errorc(ctx, "fetchSeasonTeamsByContest d.getTeamsMapByID teamIDs(%d) error(%v)", teamIDs, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if season, err = d.GetSeasonByID(ctx, contest.Sid); err != nil {
			log.Errorc(ctx, "fetchSeasonTeamsByContest d.getSeasonByID seasonID(%d) error(%v)", contest.Sid, err)
			return err
		}
		return nil
	})
	err = group.Wait()
	return
}
