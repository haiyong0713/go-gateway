package dao

import (
	"context"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-common/library/xstr"
	cacheModel "go-gateway/app/web-svr/esports/common/cache/model"
	"go-gateway/app/web-svr/esports/interface/component"
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
		if err = rows.Scan(&teamInfo.ContestId, &teamInfo.TeamId, &teamInfo.SurvivalRank, &teamInfo.KillNumber, &teamInfo.Score, &teamInfo.RankEditStatus); err != nil {
			log.Errorc(ctx, "[Dao][ContestTeams][GetTeamsByContestIds][Error], err(%+v)", err)
			return
		}
		teamsDbInfo = append(teamsDbInfo, teamInfo)
	}
	return
}

func (d *Dao) FetchTeamsMcCacheByContestId(ctx context.Context, contestId int64) (res map[int64][]*model.ContestTeamScoreInfo, err error) {
	cacheKey := cacheModel.GetContestTeamsCacheKey(contestId)
	replies := component.GlobalMemcached.Get(ctx, cacheKey)
	if nil != replies {
		log.Errorc(ctx, "[Dao][ContestTeams][FetchTeamsByCache][Error], err(%+v)", err)
		return
	}
	res = make(map[int64][]*model.ContestTeamScoreInfo)
	v := &cacheModel.ContestTeamsMcCache{}
	err = replies.Scan(v)
	if err != nil {
		if err != memcache.ErrNotFound {
			log.Errorc(ctx, "[Dao][ContestTeams][FetchTeamsByCache][Error], key:(%s), err(%+v)", cacheKey, err)
			return
		} else {
			err = nil
		}
	}
	if v.ContestId != 0 {
		res[v.ContestId] = v.Teams
	}
	return
}

func (d *Dao) FetchTeamsMcCacheByContestIds(ctx context.Context, contestIds []int64) (res map[int64][]*model.ContestTeamScoreInfo, err error) {
	cacheKeys := make([]string, 0)
	for _, v := range contestIds {
		cacheKeys = append(cacheKeys, cacheModel.GetContestTeamsCacheKey(v))
	}
	replies, err := component.GlobalMemcached.GetMulti(ctx, cacheKeys)
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestTeams][FetchTeamsByCache][Error], err(%+v)", err)
		return
	}
	res = make(map[int64][]*model.ContestTeamScoreInfo)
	for _, key := range replies.Keys() {
		v := &cacheModel.ContestTeamsMcCache{}
		err = replies.Scan(key, v)
		if err != nil {
			if err != memcache.ErrNotFound {
				log.Errorc(ctx, "[Dao][ContestTeams][FetchTeamsByCache][Error], key:(%s), err(%+v)", key, err)
				return
			} else {
				err = nil
			}
		}
		if v.ContestId != 0 {
			res[v.ContestId] = v.Teams
		}
	}
	return
}

func (d *Dao) RebuildTeamsMcCacheByContestTeamInfos(ctx context.Context, contestTeamsMap map[int64][]*model.ContestTeamInfo) (err error) {
	group, errCtx := errgroup.WithContext(ctx)
	for loopContestId, loopTeams := range contestTeamsMap {
		contestId := loopContestId
		teamsScoreInfo := make([]*model.ContestTeamScoreInfo, 0)
		for _, v := range loopTeams {
			teamsScoreInfo = append(teamsScoreInfo, &model.ContestTeamScoreInfo{
				TeamId:         v.TeamId,
				Score:          v.ScoreInfo.Score,
				KillNumber:     v.ScoreInfo.KillNumber,
				SurvivalRank:   v.ScoreInfo.SurvivalRank,
				SeasonTeamRank: v.ScoreInfo.SeasonTeamRank,
				Rank:           v.ScoreInfo.Rank,
			})
		}
		group.Go(func() error {
			cacheKey := cacheModel.GetContestTeamsCacheKey(contestId)
			item := &memcache.Item{
				Key: cacheKey,
				Object: cacheModel.ContestTeamsMcCache{
					ContestId:   contestId,
					Teams:       teamsScoreInfo,
					BuildMethod: "interface",
				},
				Expiration: cacheModel.GetContestTeamsCacheKeyTtlSecond(),
				Flags:      memcache.FlagJSON}
			groupErr := component.GlobalMemcached.Set(ctx, item)
			if groupErr != nil {
				log.Errorc(errCtx, "[Dao][ContestTeams][[SetCache][ErrorGroup][Error], key:(%s), err(%+v)", cacheKey, err)
			}
			return groupErr
		})
	}
	err = group.Wait()
	if err != nil {
		return
	}
	return
}
