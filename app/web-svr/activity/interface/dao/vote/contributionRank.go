package vote

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/tool"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	maxContribRankSize = 100
)

const (
	sql4IncrUserItemContribRank = `
INSERT INTO act_vote_user_item_summary_%v (main_id, source_group_id, source_item_id,mid,votes) VALUES(?,?,?,?,?) on duplicate key update votes=votes+?
`
	sql4IncrUserItemContribRankRisk = `
INSERT INTO act_vote_user_item_summary_%v (main_id, source_group_id, source_item_id,mid,risk_votes) VALUES(?,?,?,?,?) on duplicate key update risk_votes=risk_votes+?
`

	sql4DecrUserItemContribRank = `
UPDATE act_vote_user_item_summary_%v
SET votes = votes - ?
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    mid = ?
`

	sql4DecrUserItemContribRankRisk = `
UPDATE act_vote_user_item_summary_%v
SET risk_votes = risk_votes - ?
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    mid = ?
`

	sql4GetUserItemContribCount = `
select votes from act_vote_user_item_summary_%v
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    mid = ?
`
	sql4GetUserItemContribCountWithRisk = `
select votes+risk_votes from act_vote_user_item_summary_%v
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    mid = ?
`

	sql4GetItemContribRankWithoutRisk = `
select mid,votes,mtime from act_vote_user_item_summary_%v
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    votes > 0
ORDER BY votes DESC
LIMIT 100
`

	sql4GetItemContribRankWithRisk = `
select mid,(votes+risk_votes)as c, mtime from act_vote_user_item_summary_%v
WHERE  main_id = ?
AND    source_group_id = ?
AND	   source_item_id = ?
AND    votes+risk_votes > 0
ORDER BY c DESC
LIMIT 100
`
)

// redisItemContribRankCacheKey: 某个投票项的贡献榜, 贡献榜只保留前100
func redisItemContribRankCacheKey(sourceGroupId, sourceItemId int64) string {
	return fmt.Sprintf("vote_item_contrib_rank_v2_%v_%v", sourceGroupId, sourceItemId)
}

// redisItemContribRankCacheEmptyMarkKey: 排行榜的为空标识
func redisItemContribRankCacheEmptyMarkKey(sourceGroupId, sourceItemId int64) string {
	return fmt.Sprintf("vote_item_contrib_rank_empty_%v_%v", sourceGroupId, sourceItemId)
}

func (d *Dao) incrUserItemContribRank(tx *sql.Tx, activityId, sourceGroupId, sourceItemId, mid, voteCount int64, haveRisk, displayRisk bool) (current int64, err error) {
	sqlStr := sql4IncrUserItemContribRank
	if haveRisk {
		sqlStr = sql4IncrUserItemContribRankRisk
	}
	_, err = tx.Exec(fmt.Sprintf(sqlStr, tableIdx(sourceItemId)), activityId, sourceGroupId, sourceItemId, mid, voteCount, voteCount)
	if err == nil {
		current, err = d.getUserItemContribCount(nil, tx, activityId, sourceGroupId, sourceItemId, mid, displayRisk)
	}
	return
}

func (d *Dao) getUserItemContribCount(ctx context.Context, tx *sql.Tx, activityId, sourceGroupId, sourceItemId, mid int64, displayRisk bool) (current int64, err error) {
	sqlStr := fmt.Sprintf(sql4GetUserItemContribCount, tableIdx(sourceItemId))
	if displayRisk {
		sqlStr = fmt.Sprintf(sql4GetUserItemContribCountWithRisk, tableIdx(sourceItemId))
	}
	if tx == nil {
		err = d.db.QueryRow(ctx, sqlStr, activityId, sourceGroupId, sourceItemId, mid).Scan(&current)
	} else {
		err = tx.QueryRow(sqlStr, activityId, sourceGroupId, sourceItemId, mid).Scan(&current)
	}
	return
}

func (d *Dao) decrUserItemContribRank(tx *sql.Tx, activityId, sourceGroupId, sourceItemId, mid, voteCount, riskCount int64, displayRisk bool) (current int64, err error) {
	if voteCount != 0 {
		sqlStr := sql4DecrUserItemContribRank
		_, err = tx.Exec(fmt.Sprintf(sqlStr, tableIdx(sourceItemId)), voteCount, activityId, sourceGroupId, sourceItemId, mid)
		if err != nil {
			return
		}
	}

	if riskCount != 0 {
		sqlStr := sql4DecrUserItemContribRankRisk
		_, err = tx.Exec(fmt.Sprintf(sqlStr, tableIdx(sourceItemId)), riskCount, activityId, sourceGroupId, sourceItemId, mid)
		if err != nil {
			return
		}
	}

	current, err = d.getUserItemContribCount(nil, tx, activityId, sourceGroupId, sourceItemId, mid, displayRisk)
	return
}

// parseScoreToVoteAndMtime: score解析为时间和vote
// score: 1627464932, vote: 16, mtime: 27464932
// BenchmarkNum-8   	10358492	       112 ns/op
func (d *Dao) parseScoreToVoteAndMtime(score int64) (vote, mtime int64) {
	vote = score / 1e08
	mtime = score - (vote * 1e08) + 1600000000
	return
}

func (d *Dao) AddUserToItemContribRankCache(ctx context.Context, sourceGroupId, sourceItemId, mid, voteCount, mtime int64) (err error) {
	//时间戳1627464932的前两位16可以舍去.后八位即可唯一标记.
	score := tool.Int64Append(voteCount, mtime-1600000000)
	rankCacheKey := redisItemContribRankCacheKey(sourceGroupId, sourceItemId)
	//KEYS[1]: sorted set key name
	//ARGV[1]: sorted set max size
	//ARGV[2]: new key value
	//ARGV[3]: new key name
	script := `
	redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
	if redis.call('ZCARD', KEYS[1]) > tonumber(ARGV[1])
	then
		redis.call('ZPOPMIN',KEYS[1])
	end
	return 1`
	err = retry.WithAttempts(ctx, "AddUserToItemContribRankCache", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = redis.Bool(d.redis.Do(ctx, "EVAL", script, 1, rankCacheKey, maxContribRankSize, score, mid))
		return
	})

	if err == nil {
		_, _ = d.redis.Do(ctx, "DEL", redisItemContribRankCacheEmptyMarkKey(sourceGroupId, sourceItemId))
		_, _ = d.redis.Do(ctx, "EXPIRE", rankCacheKey, 324000 /*90天*/)
	}
	return
}

func (d *Dao) GetItemContributionRank(ctx context.Context, input *api.VoteGetItemContributionRankReq) (res *api.VoteGetItemContributionRankResp, err error) {
	res = &api.VoteGetItemContributionRankResp{
		UserAvailVoteCount:    0,
		UserAvailTmpVoteCount: 0,
		DataSourceType:        "",
		SourceGroupId:         input.SourceGroupId,
		Rank:                  make([]*api.VoteItemContributionRankItem, 0),
	}
	DSG, activity, err := d.existCheckDSGAndActivity(ctx, input.SourceGroupId, input.ActivityId)
	if err != nil {
		return
	}
	res.DataSourceType = DSG.SourceType
	res.UserAvailVoteCount, res.UserAvailTmpVoteCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, input.ActivityId, input.Mid)
	if err != nil {
		return
	}
	res.Rank, err = d.CacheGetItemContribRank(ctx, input.ActivityId, input.SourceGroupId, input.SourceItemId, input.Limit)
	return
}

func (d *Dao) CacheGetItemContribRank(ctx context.Context, activityId, sourceGroupId, sourceItemId, limit int64) (res []*api.VoteItemContributionRankItem, err error) {
	res = make([]*api.VoteItemContributionRankItem, 0)
	rankCacheKey := redisItemContribRankCacheKey(sourceGroupId, sourceItemId)
	empty, err := redis.Bool(d.redis.Do(ctx, "EXISTS", redisItemContribRankCacheEmptyMarkKey(sourceGroupId, sourceItemId)))
	if err != nil {
		return
	}
	if empty {
		return
	}

	scoreSlice, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", rankCacheKey, 0, limit, "WITHSCORES"))
	if err != nil {
		return
	}
	if len(scoreSlice) == 0 {
		empty, err = d.rebuildItemContribRankCache(ctx, activityId, sourceGroupId, sourceItemId)
		if err != nil {
			return
		}
		if empty {
			return
		}
		scoreSlice, err = redis.Values(d.redis.Do(ctx, "ZREVRANGE", rankCacheKey, 0, limit, "WITHSCORES"))
	}
	tmpResMap := make(map[int64]*api.VoteItemContributionRankItem, 0)
	mids := make([]int64, 0)
	for len(scoreSlice) > 0 {
		var mid, score int64
		scoreSlice, err = redis.Scan(scoreSlice, &mid, &score)
		if err != nil {
			return
		}
		vote, mtime := d.parseScoreToVoteAndMtime(score)
		mids = append(mids, mid)
		tmpResMap[mid] = &api.VoteItemContributionRankItem{
			UserMid:    mid,
			UserFace:   "",
			UserName:   "",
			Times:      vote,
			LastVoteAt: mtime,
		}
	}
	midsReply, err := client.AccountClient.Infos3(ctx, &accapi.MidsReq{
		Mids: mids,
	})
	if err != nil {
		return
	}
	for _, mid := range mids {
		tmpRes := tmpResMap[mid]
		info, ok := midsReply.Infos[mid]
		if ok {
			tmpRes.UserName = info.Name
			tmpRes.UserFace = info.Face
			res = append(res, tmpRes)
		}
	}
	return

}

func (d *Dao) rebuildItemContribRankCache(ctx context.Context, activityId, sourceGroupId, sourceItemId int64) (empty bool, err error) {
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		return
	}
	var sqlStr string
	if activity.Rule.DisplayRiskVote {
		sqlStr = sql4GetItemContribRankWithRisk
	} else {
		sqlStr = sql4GetItemContribRankWithoutRisk
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(sqlStr, tableIdx(sourceItemId)), activityId, sourceGroupId, sourceItemId)
	if err == sql.ErrNoRows {
		empty = true
		err = nil
		_, _ = d.redis.Do(ctx, "SET", redisItemContribRankCacheEmptyMarkKey(sourceGroupId, sourceItemId))
		return
	}
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var mid, count int64
		var mtime xtime.Time
		err = rows.Scan(&mid, &count, &mtime)
		if err != nil {
			return
		}
		err = d.AddUserToItemContribRankCache(ctx, sourceGroupId, sourceItemId, mid, count, mtime.Time().Unix())
		if err != nil {
			return
		}
	}
	err = rows.Err()
	return
}
