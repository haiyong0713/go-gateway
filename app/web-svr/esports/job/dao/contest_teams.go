package dao

import (
	"context"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/xstr"
	cacheModel "go-gateway/app/web-svr/esports/common/cache/model"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	_getTeamsByContestIdsOrderByScoreDescSql = "select contest_id, team_id, survival_rank, kill_number, score, rank_edit_status" +
		" from es_contest_teams" +
		" where is_deleted = 0 and contest_id in (?) " +
		" order by contest_id asc, score desc, kill_number desc, id desc"
)

func (d *Dao) GetTeamsByContestIds(ctx context.Context, contestIds []int64) (teamsDbInfo []*model.ContestTeamDbInfo, err error) {
	contestIdsStr := xstr.JoinInts(contestIds)
	teamsDbInfo = make([]*model.ContestTeamDbInfo, 0)
	rows, err := d.db.Query(ctx, _getTeamsByContestIdsOrderByScoreDescSql, contestIdsStr)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		teamInfo := new(model.ContestTeamDbInfo)
		err = rows.Scan(&teamInfo.ContestId, &teamInfo.TeamId, &teamInfo.SurvivalRank, &teamInfo.KillNumber, &teamInfo.Score, &teamInfo.RankEditStatus)
		if err != nil {
			log.Errorc(ctx, "[Dao][GetTeamsByContestIds][[Scan][Error], err(%+v)", err)
			return
		}
		teamsDbInfo = append(teamsDbInfo, teamInfo)
	}
	return
}

func (d *Dao) RebuildTeamsMcCacheByContestTeamInfos(ctx context.Context, contestId int64, contestTeams []*model.ContestTeamInfo) (err error) {
	teamsScoreInfo := make([]*model.ContestTeamScoreInfo, 0)
	for _, v := range contestTeams {
		teamsScoreInfo = append(teamsScoreInfo, &model.ContestTeamScoreInfo{
			TeamId:         v.TeamId,
			Score:          v.ScoreInfo.Score,
			KillNumber:     v.ScoreInfo.KillNumber,
			SurvivalRank:   v.ScoreInfo.SurvivalRank,
			SeasonTeamRank: v.ScoreInfo.SeasonTeamRank,
			Rank:           v.ScoreInfo.Rank,
		})
	}
	cacheKey := cacheModel.GetContestTeamsCacheKey(contestId)
	value := cacheModel.ContestTeamsMcCache{
		ContestId:   contestId,
		Teams:       teamsScoreInfo,
		BuildMethod: "job",
	}
	item := &memcache.Item{
		Key:        cacheKey,
		Object:     value,
		Expiration: cacheModel.GetContestTeamsCacheKeyTtlSecond(),
		Flags:      memcache.FlagJSON}
	err = d.memcache.Set(ctx, item)
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestTeams][[SetCache][ErrorGroup][Error], key:(%s), err(%+v)", cacheKey, err)
	}
	return
}
