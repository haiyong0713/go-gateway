package match_component

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	egv2 "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	_componentContestCardListCacheKey       = "esport:component:contestList:v2:sid:%v"
	_componentContestsBySeasonCacheKey      = "esport:component:contests:v2:sid:%v"
	_componentContestBattleCardListCacheKey = "esport:component:battle:card:v2:sid:%v"
	_componentContestBattleBySeasonCacheKey = "esport:component:battle:v2:sid:%v"
	_componentGoingSeasonListCacheKey       = "esport:component:going:seasonList"
	_componentGoingBattleSeasonListCacheKey = "esport:component:going:battle:seasonList"
	_componentHomeAwayContestCacheKey       = "esport:component:home:%v:away:%v:contest"
	_limitKey2FetchComponentContest         = "component_contest_limit"
	_limitKey2FetchComponentContestSeries   = "component_contest_series_limit"
	_limitKey2FetchComponentContestBattle   = "component_contest_battle_limit"
	_esTopicVideoListCacheKey               = "es:video:list:game:%v:match:%v:year:%v:pn:%v:ps:%v"
)

func componentContestCardListCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestCardListCacheKey, sid)
}
func esTopicVideoListCacheKey(param *pb.EsTopicVideoListRequest) string {
	return fmt.Sprintf(_esTopicVideoListCacheKey, param.GameId, param.MatchId, param.YearId, param.Pn, param.Ps)
}

func componentContestBySeasonCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestsBySeasonCacheKey, sid)
}

func componentContestBattleCardListCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestBattleCardListCacheKey, sid)
}

func componentContestBattleBySeasonCacheKey(sid int64) string {
	return fmt.Sprintf(_componentContestBattleBySeasonCacheKey, sid)
}

func componentHomeAwayContestCacheKey(homeTeamID, awayTeamID int64) string {
	return fmt.Sprintf(_componentHomeAwayContestCacheKey, homeTeamID, awayTeamID)
}

const sql4FetchContestsOfAll = `
SELECT id, UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(stime), '%Y-%m-%d')) AS date_unix
	, stime, etime, game_stage, live_room, playback
	, collection_url, home_id, home_score, away_id
    , away_score, match_id, series_id, guess_type
    , data_type, sid, status, contest_status
    FROM es_contests
    WHERE sid = ? AND status = 0 and special = 0 and stime > 0
    ORDER BY stime DESC`

func RawComponentContestListBySeasonID(ctx context.Context, seasonID int64) (list []*model.Contest2TabComponent, err error) {
	var rows *xsql.Rows
	rows, err = component.GlobalDBOfMaster.Query(ctx, sql4FetchContestsOfAll, seasonID)
	if err != nil {
		return
	}
	list = make([]*model.Contest2TabComponent, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "RawComponentContestListBySeasonID rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		contest := new(model.Contest2TabComponent)
		err = rows.Scan(
			&contest.ID,
			&contest.StimeDate,
			&contest.Stime,
			&contest.Etime,
			&contest.GameStage,
			&contest.LiveRoom,
			&contest.PlayBack,
			&contest.CollectionUrl,
			&contest.HomeID,
			&contest.HomeScore,
			&contest.AwayID,
			&contest.AwayScore,
			&contest.MatchID,
			&contest.SeriesID,
			&contest.GuessType,
			&contest.DataType,
			&contest.SeasonID,
			&contest.Status,
			&contest.ContestStatus)
		if err == nil {
			list = append(list, contest)
		}
	}
	return
}

const sql4FetchContestSeriesComponent = `
SELECT id, parent_title, child_title, score_id, start_time, end_time
FROM contest_series
WHERE id IN (%v) and is_deleted = 0
`

func FetchContestSeriesListFromDB(ctx context.Context, idList []int64) (list map[int64]*pb.ContestSeriesComponent, err error) {
	var rows *xsql.Rows
	list = make(map[int64]*pb.ContestSeriesComponent, 0)
	if len(idList) == 0 {
		return
	}
	idList = tool.Unique(idList)
	rows, err = component.GlobalDBOfMaster.Query(ctx, fmt.Sprintf(sql4FetchContestSeriesComponent, xstr.JoinInts(idList)))
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if e := rows.Err(); e != nil {
			log.Errorc(ctx, "FetchContestSeriesListFromDB rows.Err() error(%v)", e)
		}
	}()
	for rows.Next() {
		series := new(pb.ContestSeriesComponent)
		scanErr := rows.Scan(
			&series.ID,
			&series.ParentTitle,
			&series.ChildTitle,
			&series.ScoreID,
			&series.StartTime,
			&series.EndTime)
		if scanErr == nil {
			list[series.ID] = series
		} else {
			log.Errorc(ctx, "contest component FetchContestSeriesListFromDB scan error(%+v)", scanErr)
		}
	}
	return
}

const sql4FetchAllTeams = `
SELECT id, title, sub_title, logo, region_id, leida_tid
FROM es_teams 
WHERE is_deleted = 0 limit 10000 
`

func FetchAllTeams(ctx context.Context) (res map[int64]*model.Team2TabComponent, err error) {
	var rows *xsql.Rows
	if rows, err = component.GlobalDBOfMaster.Query(ctx, sql4FetchAllTeams); err != nil {
		log.Error("FetchAllTeams component.GlobalDBOfMaster.Query error(%v)", err)
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchAllTeams rows.Err() error(%v)", err)
		}
	}()
	res = make(map[int64]*model.Team2TabComponent, 0)
	for rows.Next() {
		team := new(model.Team2TabComponent)
		err = rows.Scan(&team.ID, &team.Title, &team.SubTitle, &team.Logo, &team.RegionID, &team.ScoreTeamID)
		if err == nil {
			res[team.ID] = team
		} else {
			log.Warnc(ctx, "contest component FetchAllTeams error(%+v)", err)
		}
	}
	return
}

func FetchContestCardListFromCache(ctx context.Context, seasonID int64) (res map[int64][]*pb.ContestCardComponent, err error) {
	cacheKey := componentContestCardListCacheKey(seasonID)
	err = component.GlobalMemcached.Get(ctx, cacheKey).Scan(&res)
	return
}

func FetchContestCardListDeleteCache(ctx context.Context, seasonID int64) (err error) {
	cacheKey := componentContestCardListCacheKey(seasonID)
	if err = retry.WithAttempts(ctx, "old_contest_card_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = component.GlobalMemcached.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component FetchContestCardListDeleteCache seasonID(%d) GlobalMemcached.Delete error(%+v)", seasonID, err)
		return err
	}
	return
}

func FetchContestCardListToCache(ctx context.Context, seasonID int64, contestListMap map[int64][]*pb.ContestCardComponent, expire int32) (err error) {
	item := &memcache.Item{
		Key:        componentContestCardListCacheKey(seasonID),
		Object:     contestListMap,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contest_list_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component  FetchContestListToCache component.GlobalMemcached.Set error(%+v)", err)
	}
	return
}

func FetchContestsBySeasonComponent(ctx context.Context, seasonID int64) (res []*model.Contest2TabComponent, err error) {
	if res, err = FetchContestsBySeasonFromCache(ctx, seasonID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component FetchContestsBySeasonComponent FetchContestsBySeasonFromCache() sid(%d) error(%+v)", seasonID, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		if tool.IsLimiterAllowedByUniqBizKey(_limitKey2FetchComponentContest, _limitKey2FetchComponentContest) {
			res, err = RawComponentContestListBySeasonID(context.Background(), seasonID)
			if err != nil {
				return
			}
			if e := FetchContestsBySeasonToCache(ctx, seasonID, res, int32(tool.CalculateExpiredSeconds(1))); e != nil {
				log.Errorc(ctx, "contest component FetchContestsBySeasonComponent FetchContestsBySeasonToCache() sid(%d) error(%+v)", seasonID, e)
			}
		} else {
			err = xecode.LimitExceed
		}
	}
	return
}

func FetchContestsBySeasonFromCache(ctx context.Context, seasonID int64) (res []*model.Contest2TabComponent, err error) {
	cacheKey := componentContestBySeasonCacheKey(seasonID)
	err = component.GlobalMemcached.Get(ctx, cacheKey).Scan(&res)
	return
}

func FetchContestsBySeasonDeleteCache(ctx context.Context, seasonID int64) (err error) {
	cacheKey := componentContestBySeasonCacheKey(seasonID)
	if err = retry.WithAttempts(ctx, "old_contest_by_season_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = component.GlobalMemcached.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component FetchContestsBySeasonDeleteCache seasonID(%d) GlobalMemcached.Delete error(%+v)", seasonID, err)
		return err
	}
	return
}

func FetchContestsBySeasonToCache(ctx context.Context, seasonID int64, contests []*model.Contest2TabComponent, expire int32) (err error) {
	item := &memcache.Item{
		Key:        componentContestBySeasonCacheKey(seasonID),
		Object:     contests,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contests_by_season_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component  FetchContestsBySeasonToCache component.GlobalMemcached.Set() sid(%d) error(%+v)", seasonID, err)
	}
	return
}

func FetchGoingSeasonsFromCache(ctx context.Context) (res []*model.ComponentSeason, err error) {
	res = make([]*model.ComponentSeason, 0)
	err = component.GlobalMemcached.Get(ctx, _componentGoingSeasonListCacheKey).Scan(&res)
	return
}

func FetchGoingBattleSeasonsFromCache(ctx context.Context) (res []*model.ComponentSeason, err error) {
	res = make([]*model.ComponentSeason, 0)
	err = component.GlobalMemcached.Get(ctx, _componentGoingBattleSeasonListCacheKey).Scan(&res)
	return
}

const _keySeries = "contest_series_%d"

func keyContestSeriesID(id int64) string {
	return fmt.Sprintf(_keySeries, id)
}

// AddCacheContestSeries .
func AddCacheContestSeries(ctx context.Context, data map[int64]*pb.ContestSeriesComponent) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsCid = redis.Args{}
	)
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheContestSeries.json.Marshal error(%v)", err)
			continue
		}
		keyID = keyContestSeriesID(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsCid = argsCid.Add(keyID).Add(string(bs))
	}
	if _, err = component.GlobalAutoSubCache.Do(ctx, "MSET", argsCid...); err != nil {
		log.Error("AddCacheContestSeries conn.Send(MSET) error(%v)", err)
		return
	}
	for _, v := range keyIDs {
		if _, err = component.GlobalAutoSubCache.Do(ctx, "EXPIRE", v, 86400); err != nil {
			return err
		}
	}
	return
}

// CacheContestSeries .
func CacheContestSeries(ctx context.Context, ids []int64) (res map[int64]*pb.ContestSeriesComponent, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, csid := range ids {
		key = keyContestSeriesID(csid)
		args = args.Add(key)
	}
	if bss, err = redis.ByteSlices(component.GlobalAutoSubCache.Do(ctx, "MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheEpTeams conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*pb.ContestSeriesComponent, len(ids))
	for _, bs := range bss {
		contestSeries := new(pb.ContestSeriesComponent)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, contestSeries); err != nil {
			log.Error("CacheContestSeries json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[contestSeries.ID] = contestSeries
	}
	return
}

// ContestSeriesComponent get data from cache if miss will call source method, then add to cache.
func ContestSeriesComponent(c context.Context, ids []int64) (res map[int64]*pb.ContestSeriesComponent, err error) {
	if len(ids) == 0 {
		return
	}
	addCache := true
	if res, err = CacheContestSeries(c, ids); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range ids {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	cache.MetricHits.Add(float64(len(ids)-len(miss)), "bts:EpContestSeries")
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*pb.ContestSeriesComponent
	cache.MetricMisses.Add(float64(len(miss)), "bts:EpContestSeries")
	if tool.IsLimiterAllowedByUniqBizKey(_limitKey2FetchComponentContestSeries, _limitKey2FetchComponentContestSeries) {
		missData, err = FetchContestSeriesListFromDB(c, miss)
		if err != nil {
			return
		}
	} else {
		err = xecode.LimitExceed
		return
	}
	if res == nil {
		res = make(map[int64]*pb.ContestSeriesComponent, len(ids))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	AddCacheContestSeries(c, missData)
	return
}

func DelContestSeriesCacheKey(ctx context.Context, id int64) (err error) {
	if err = retry.WithAttempts(ctx, "old_contest_series_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = component.GlobalAutoSubCache.Do(ctx, "DEL", keyContestSeriesID(id))
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component delContestSeriesCacheKey id(%d) error(%+v)", id, err)
		return err
	}
	return
}

const sql4FetchContestBattle = `
SELECT id, UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(stime), '%Y-%m-%d')) AS date_unix
	, stime, etime, game_stage, live_room, playback
	, collection_url, match_id, series_id, guess_type
    , sid, status, contest_status
    FROM es_contests
    WHERE sid = ? AND status = 0 and special = 1 and stime > 0
    ORDER BY stime ASC`

func RawComponentContestBattleBySeasonID(ctx context.Context, seasonID int64) (list []*model.ContestBattle2DBComponent, err error) {
	var rows *xsql.Rows
	rows, err = component.GlobalDBOfMaster.Query(ctx, sql4FetchContestBattle, seasonID)
	if err != nil {
		return
	}
	list = make([]*model.ContestBattle2DBComponent, 0)
	defer func() {
		_ = rows.Close()
		if e := rows.Err(); e != nil {
			log.Errorc(ctx, "RawComponentContestBattleBySeasonID rows.Err() error(%v)", e)
		}
	}()
	for rows.Next() {
		contest := new(model.ContestBattle2DBComponent)
		err = rows.Scan(
			&contest.ID,
			&contest.StimeDate,
			&contest.Stime,
			&contest.Etime,
			&contest.GameStage,
			&contest.LiveRoom,
			&contest.PlayBack,
			&contest.CollectionUrl,
			&contest.MatchID,
			&contest.SeriesID,
			&contest.GuessType,
			&contest.SeasonID,
			&contest.Status,
			&contest.ContestStatus)
		if err == nil {
			list = append(list, contest)
		}
	}
	return
}

func FetchContestBattleBySeasonComponent(ctx context.Context, seasonID int64) (res []*model.ContestBattle2DBComponent, err error) {
	if res, err = FetchContestBattleBySeasonFromCache(ctx, seasonID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "contest component FetchContestBattleBySeasonComponent FetchContestsBySeasonFromCache() sid(%d) error(%+v)", seasonID, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		if tool.IsLimiterAllowedByUniqBizKey(_limitKey2FetchComponentContestBattle, _limitKey2FetchComponentContestBattle) {
			res, err = RawComponentContestBattleBySeasonID(context.Background(), seasonID)
			if err != nil {
				return
			}
			if e := FetchContestBattleBySeasonToCache(ctx, seasonID, res, int32(tool.CalculateExpiredSeconds(1))); e != nil {
				log.Errorc(ctx, "contest component FetchContestBattleBySeasonComponent FetchContestBattleBySeasonToCache() sid(%d) error(%+v)", seasonID, e)
			}
		} else {
			err = xecode.LimitExceed
		}
	}
	return
}

func FetchContestBattleBySeasonFromCache(ctx context.Context, seasonID int64) (res []*model.ContestBattle2DBComponent, err error) {
	cacheKey := componentContestBattleBySeasonCacheKey(seasonID)
	err = component.GlobalMemcached.Get(ctx, cacheKey).Scan(&res)
	return
}

func FetchContestBattleBySeasonToCache(ctx context.Context, seasonID int64, contests []*model.ContestBattle2DBComponent, expire int32) (err error) {
	item := &memcache.Item{
		Key:        componentContestBattleBySeasonCacheKey(seasonID),
		Object:     contests,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contest_battle_by_season_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component  FetchContestBattleBySeasonToCache component.GlobalMemcached.Set() sid(%d) error(%+v)", seasonID, err)
	}
	return
}

func FetchContestBattleBySeasonDeleteCache(ctx context.Context, seasonID int64) (err error) {
	cacheKey := componentContestBattleBySeasonCacheKey(seasonID)
	if err = retry.WithAttempts(ctx, "old_contest_battle_by_season_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = component.GlobalMemcached.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component FetchContestBattleBySeasonDeleteCache seasonID(%d) GlobalMemcached.Delete error(%+v)", seasonID, err)
		return err
	}
	return
}

// FetchContestBattleCardListFromCache get api card battle list cache.
func FetchContestBattleCardListFromCache(ctx context.Context, seasonID int64) (res map[int64][]*pb.ContestBattleCardComponent, err error) {
	cacheKey := componentContestBattleCardListCacheKey(seasonID)
	err = component.GlobalMemcached.Get(ctx, cacheKey).Scan(&res)
	return
}

// FetchContestBattleCardListToCache set api card list cache.
func FetchContestBattleCardListToCache(ctx context.Context, seasonID int64, contestBattleMap map[int64][]*pb.ContestBattleCardComponent, expire int32) (err error) {
	item := &memcache.Item{
		Key:        componentContestBattleCardListCacheKey(seasonID),
		Object:     contestBattleMap,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "component_contest_battle_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component  FetchContestBattleCardListToCache component.GlobalMemcached.Set error(%+v)", err)
	}
	return
}

// FetchContestBattleCardListDeleteCache delete battle card cache.
func FetchContestBattleCardListDeleteCache(ctx context.Context, seasonID int64) (err error) {
	cacheKey := componentContestBattleCardListCacheKey(seasonID)
	if err = retry.WithAttempts(ctx, "old_contest_battle_card_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = component.GlobalMemcached.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component FetchContestCardListDeleteCache seasonID(%d) GlobalMemcached.Delete error(%+v)", seasonID, err)
		return err
	}
	return
}

const _howeAwayContest = `
SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,
       live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,
       special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,guess_type,
       match_id,game_stage1,game_stage2,push_switch,series_id,contest_status
       FROM es_contests FORCE INDEX (idx_stime_home_id)
       WHERE status = 0 AND special = 0 AND etime < ? AND ((home_id=? AND away_id=?) OR (home_id=? AND away_id=?)) 
       ORDER BY stime DESC LIMIT 11`

func RawHoweAwayContest(ctx context.Context, param *model.ParamEsGuess) (list []*model.Contest, err error) {
	var rows *xsql.Rows
	list = make([]*model.Contest, 0)
	rows, err = component.GlobalDBOfMaster.Query(ctx, _howeAwayContest, time.Now().Unix(), param.HomeID, param.AwayID, param.AwayID, param.HomeID)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "RawHoweAwayContest rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		r := new(model.Contest)
		scanErr := rows.Scan(
			&r.ID, &r.GameStage, &r.Stime, &r.Etime, &r.HomeID, &r.AwayID, &r.HomeScore, &r.AwayScore,
			&r.LiveRoom, &r.Aid, &r.Collection, &r.GameState, &r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Sid, &r.Mid,
			&r.Special, &r.SpecialName, &r.SpecialTips, &r.SuccessTeam, &r.SpecialImage, &r.Playback, &r.CollectionURL,
			&r.LiveURL, &r.DataType, &r.GuessType, &r.MatchID, &r.GameStage1, &r.GameStage2, &r.PushSwitch,
			&r.SeriesID, &r.ContestStatus,
		)
		if scanErr == nil {
			list = append(list, r)
		} else {
			log.Errorc(ctx, "RawHoweAwayContest scan error(%+v)", scanErr)
		}
	}
	return
}

func FetchHomeAwayContestsFromCache(ctx context.Context, homeID, awayID int64) (res []*model.Contest, err error) {
	res = make([]*model.Contest, 0)
	cacheKey := componentHomeAwayContestCacheKey(homeID, awayID)
	if err == nil {
		err = component.GlobalMemcached4UserGuess.Get(ctx, cacheKey).Scan(&res)
	}
	return
}

func FetchHomeAwayContestsToCache(ctx context.Context, homeID, awayID int64, data []*model.Contest) (err error) {
	cacheKey := componentHomeAwayContestCacheKey(homeID, awayID)
	item := &memcache.Item{
		Key:        cacheKey,
		Object:     data,
		Expiration: int32(tool.CalculateExpiredSeconds(1)),
		Flags:      memcache.FlagJSON,
	}
	if err = component.GlobalMemcached4UserGuess.Set(ctx, item); err != nil {
		log.Errorc(ctx, "FetchHomeAwayContestsToCache component.GlobalMemcached4UserGuess.Set home(%d) away(%d) error(%+v)", homeID, awayID, err)
	}
	return
}

func FetchHomeAwayContestsDeleteCache(ctx context.Context, homeID, awayID int64) (err error) {
	if homeID == 0 || awayID == 0 {
		return
	}
	errGroupV2 := egv2.WithContext(ctx)
	errGroupV2.Go(func(ctx context.Context) (err error) {
		return DeleteHomeAwayContests(ctx, homeID, awayID)
	})
	errGroupV2.Go(func(ctx context.Context) (err error) {
		return DeleteHomeAwayContests(ctx, awayID, homeID)
	})
	if err = errGroupV2.Wait(); err != nil {
		log.Errorc(ctx, "FetchHomeAwayContestsDeleteCache homeID(%d) awayID(%d) error(%+v)", homeID, awayID, err)
	}
	return
}

func DeleteHomeAwayContests(ctx context.Context, homeID, awayID int64) (err error) {
	cacheKey := componentHomeAwayContestCacheKey(homeID, awayID)
	if err = retry.WithAttempts(ctx, "home_away_contest_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = component.GlobalMemcached4UserGuess.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "contest component DeleteHomeAwayContests homeID(%d) awayID(%d) GlobalMemcached4UserGuess.Delete error(%+v)", homeID, awayID, err)
		return err
	}
	return
}

const sql4FetchAllSeasons = `
SELECT id, title, sub_title, logo, leida_sid
       FROM es_seasons WHERE status = 0 limit 10000 
`

func FetchAllSeasons(ctx context.Context) (res map[int64]*model.SeasonComponent, err error) {
	var rows *xsql.Rows
	if rows, err = component.GlobalDBOfMaster.Query(ctx, sql4FetchAllSeasons); err != nil {
		log.Error("FetchAllSeasons component.GlobalDBOfMaster.Query error(%v)", err)
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchAllSeasons rows.Err() error(%v)", err)
		}
	}()
	res = make(map[int64]*model.SeasonComponent)
	for rows.Next() {
		season := new(model.SeasonComponent)
		err = rows.Scan(&season.ID, &season.Title, &season.SubTitle, &season.Logo, &season.LeidaSid)
		if err == nil {
			res[season.ID] = season
		} else {
			log.Warnc(ctx, "contest component FetchAllSeasons error(%+v)", err)
		}
	}
	return
}

const sql4GoingVideoList = `SELECT id,ugc_aids,game_id,match_id,year_id FROM es_video_lists WHERE is_deleted=0 AND stime<=? AND etime>=? limit 1000`

func GoingVideoList(ctx context.Context, before, after int64) (list []*model.VideoListInfo, err error) {
	var rows *xsql.Rows
	rows, err = component.GlobalDBOfMaster.Query(ctx, sql4GoingVideoList, before, after)
	if err != nil {
		return
	}
	list = make([]*model.VideoListInfo, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "GoingVideoList rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		videoList := new(model.VideoListInfo)
		scanErr := rows.Scan(&videoList.ID, &videoList.UgcAids, &videoList.GameID, &videoList.MatchID, &videoList.YearID)
		if scanErr == nil {
			list = append(list, videoList)
		} else {
			log.Errorc(ctx, "GoingVideoList scan error(%+v)", scanErr)
		}
	}
	return
}

func FetchEsTopicVideoListFromCache(ctx context.Context, param *pb.EsTopicVideoListRequest) (res *pb.EsTopicVideoListReply, err error) {
	cacheKey := esTopicVideoListCacheKey(param)
	err = component.GlobalMemcached4UserGuess.Get(ctx, cacheKey).Scan(&res)
	return
}

func FetchEsTopicVideoListToCache(ctx context.Context, param *pb.EsTopicVideoListRequest, data *pb.EsTopicVideoListReply, expire int32) (err error) {
	item := &memcache.Item{
		Key:        esTopicVideoListCacheKey(param),
		Object:     data,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "es_video_list_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return component.GlobalMemcached4UserGuess.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component FetchEsTopicVideoListToCache component.GlobalMemcached4UserGuess.Set error(%+v)", err)
	}
	return
}
