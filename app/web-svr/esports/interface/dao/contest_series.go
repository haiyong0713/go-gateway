package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

// 阶段类型
const (
	//无-- 普通阶段
	SeriesTypEmpty = 0
	//积分赛
	SeriesTypPoint = 1
	//淘汰赛
	SeriesTypKnockout = 2
)

// SQL
const (
	sql4CountContestSeriesByIdAndTyp = `
SELECT /*master*/ count(1)
FROM contest_series
WHERE id = ? and type = ? and is_deleted = 0
`

	sql4CountEmptyConfigContestSeriesByIdAndTyp = `
SELECT /*master*/ count(1)
FROM contest_series
WHERE id = ? and type = ? and extra_config='' and is_deleted = 0
`

	sql4SetContestSeriesExtraConfig = `
update contest_series set extra_config = ?
WHERE id = ? and is_deleted = 0
`

	sql4GetContestSeriesExtraConfig = `
select /*master*/  extra_config from contest_series
WHERE id = ? and is_deleted = 0
`
	sql4GetContestIdsByContestSeriesId = `
select id from es_contests
where series_id = ?
`
	sql4GetDeletedContestSeriesIdAndType = `
select type from contest_series where id = ?
`
	sql4GetAllSeriesInSeason = `
SELECT id, parent_title, child_title
FROM contest_series
WHERE season_id = ? and is_deleted = 0
`
	sql4GetContestIdsByContestSeasonId = `
select id from es_contests
where sid = ?
`
)

// Cache Key
const (
	cache4SeriesPointMatchInfo    = "point_m_info_%v"
	cache4SeriesKnockoutMatchInfo = "knockout_contest_points_info_%v"
	cache4SeriesIdExist           = "series_exist_%v_%v"
	cache4SeriesIdRefreshing      = "series_refreshing_%v_%v"
	cache4SeriesDeleting          = "series_deleting_%v"
)

func getCacheKey4SeriesPointMatchInfo(seriesId int64) string {
	return fmt.Sprintf(cache4SeriesPointMatchInfo, seriesId)
}

func getCacheKey4SeriesKnockoutMatchInfo(seriesId int64) string {
	return fmt.Sprintf(cache4SeriesKnockoutMatchInfo, seriesId)
}

func getCacheKey4SeriesIdExist(seriesId, typ int64) string {
	return fmt.Sprintf(cache4SeriesIdExist, typ, seriesId)
}

func getCacheKey4SeriesIdRefreshing(seriesId, typ int64) string {
	return fmt.Sprintf(cache4SeriesIdRefreshing, typ, seriesId)
}

func getCacheKey4SeriesDeleting(seriesId int64) string {
	return fmt.Sprintf(cache4SeriesDeleting, seriesId)
}

func (d *Dao) SetSeriesExtraConfig(ctx context.Context, id int64, configContent string) (err error) {
	_, err = d.db.Exec(ctx, sql4SetContestSeriesExtraConfig, configContent, id)
	if err != nil {
		log.Errorc(ctx, "SetSeriesExtraConfig error: %v", err)
	}
	return err
}

func (d *Dao) GetSeriesExtraConfig(ctx context.Context, id int64) (configContent string, err error) {
	err = d.db.QueryRow(ctx, sql4GetContestSeriesExtraConfig, id).Scan(&configContent)
	if err != nil {
		log.Errorc(ctx, "GetSeriesExtraConfig error: %v", err)
	}
	return
}

func (d *Dao) IsSeriesExistsDB(ctx context.Context, id, typ int64) (exists bool, err error) {
	count := 0
	err = d.db.QueryRow(ctx, sql4CountContestSeriesByIdAndTyp, id, typ).Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "IsSeriesExistsDB error: %v", err)
	}
	exists = count != 0
	return
}

func (d *Dao) MarkSeriesRefreshing(ctx context.Context, id, typ int64) (ok bool, err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := getCacheKey4SeriesIdRefreshing(id, typ)
	var res string
	res, err = redis.String(conn.Do("SET", key, 1, "EX", 1, "NX"))
	ok = res == "OK"
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Error("conn.Send(SETNX) key(%s) error(%v)", key, err)
		return
	}
	return
}

func (d *Dao) IsSeriesExists(ctx context.Context, id, typ int64) (exists bool, err error) {
	conn := d.redis.Get(ctx)
	shouldUpdateCache := false
	var existInt int64
	defer func() {
		if shouldUpdateCache {
			conn.Do("SETEX", getCacheKey4SeriesIdExist(id, typ), 180, existInt)
		}
		conn.Close()
	}()
	existInt, err = redis.Int64(conn.Do("GET", getCacheKey4SeriesIdExist(id, typ)))
	if err == nil {
		exists = existInt == 1
		return
	}
	exists, err = d.IsSeriesExistsDB(ctx, id, typ)
	if exists {
		existInt = 1
	}
	shouldUpdateCache = err == nil
	return
}

func (d *Dao) IsSeriesDeleting(ctx context.Context, id int64) (deleting bool, err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	c, err := redis.Int64(conn.Do("GET", getCacheKey4SeriesDeleting(id)))
	if err == nil {
		deleting = c == 1
		return
	}
	return
}

func (d *Dao) IsSeriesExtraConfigEmpty(ctx context.Context, id, typ int64) (empty bool, err error) {
	count := 0
	err = d.db.QueryRow(ctx, sql4CountEmptyConfigContestSeriesByIdAndTyp, id, typ).Scan(&count)
	if err != nil {
		log.Errorc(ctx, "IsSeriesExtraConfigEmpty error: %v", err)
	}
	empty = count != 0
	return
}

func (d *Dao) IsSeriesExistsAndExtraConfigEmpty(ctx context.Context, id, typ int64) (exists, empty bool, err error) {
	exists, err = d.IsSeriesExistsDB(ctx, id, typ)
	if err != nil {
		return
	}
	empty, err = d.IsSeriesExtraConfigEmpty(ctx, id, typ)
	if err != nil {
		return
	}
	return
}

func (d *Dao) GetSeriesPointsMatchConfig(ctx context.Context, id int64) (config *v1.SeriesPointMatchConfig, err error) {
	config = &v1.SeriesPointMatchConfig{}
	exists, empty, err := d.IsSeriesExistsAndExtraConfigEmpty(ctx, id, SeriesTypPoint)
	if err != nil {
		return
	}
	if !exists {
		err = ecode.EsportsContestSeriesNotFound
		return
	}
	if empty {
		err = ecode.EsportsContestSeriesExtraConfigNotFound
		return
	}
	configContent, err := d.GetSeriesExtraConfig(ctx, id)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(configContent), &config)
	if err != nil {
		log.Errorc(ctx, "GetSeriesPointsMatchConfig error: %v", err)
		err = ecode.EsportsContestSeriesExtraConfigErr
	}
	return
}

func (d *Dao) GetContestsBySeriesId(ctx context.Context, id int64) (res map[int64]*model.Contest, err error) {
	var (
		rows *sql.Rows
	)
	res = make(map[int64]*model.Contest, 0)
	if rows, err = d.db.Query(ctx, sql4GetContestIdsByContestSeriesId, id); err != nil {
		log.Errorc(ctx, "GetContestsBySeriesId db.Query error: %v", err)
		return
	}
	defer func() {
		_ = rows.Close()
		if e := rows.Err(); e != nil {
			log.Errorc(ctx, "GetContestsBySeriesId rows.Err() error(%v)", e)
		}
	}()
	contestIds := make([]int64, 0)
	for rows.Next() {
		contestId := int64(0)
		err = rows.Scan(&contestId)
		if err != nil {
			log.Errorc(ctx, "GetContestsBySeriesId rows.Scan error: %v", err)
			return
		}
		contestIds = append(contestIds, contestId)
	}
	if len(contestIds) == 0 {
		return
	}
	if res, err = d.RawEpContests(ctx, contestIds); err != nil {
		log.Errorc(ctx, "GetContestsBySeriesId d.RawEpContests id(%d) error(%+v)", id, err)
	}
	return
}

func (d *Dao) SetSeriesPointMatchInfo(ctx context.Context, info *v1.SeriesPointMatchInfo) (err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(info)
	if err != nil {
		log.Errorc(ctx, "SetSeriesPointMatchInfo json.Marshal error: %v", err)
		return
	}
	for i := 0; i < 3; i++ {
		_, err = conn.Do("SETEX", getCacheKey4SeriesPointMatchInfo(info.SeriesId), tool.CalculateExpiredSeconds(90), bs)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "SetSeriesPointMatchInfo conn.Do error: %v", err)
	}
	return
}

func (d *Dao) SetSeriesKnockoutMatchInfo(ctx context.Context, info *v1.SeriesKnockoutMatchInfo) (err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(info)
	if err != nil {
		log.Errorc(ctx, "SetSeriesKnockoutMatchInfo json.Marshal error: %v", err)
		return
	}
	for i := 0; i < 3; i++ {
		_, err = conn.Do("SETEX", getCacheKey4SeriesKnockoutMatchInfo(info.SeriesId), tool.CalculateExpiredSeconds(90), bs)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "SetSeriesKnockoutMatchInfo conn.Do error: %v", err)
	}
	return
}

func (d *Dao) GetSeriesPointMatchInfo(ctx context.Context, seriesId int64) (info *v1.SeriesPointMatchInfo, found bool, err error) {
	info = &v1.SeriesPointMatchInfo{}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	var bs []byte
	found = true
	for i := 0; i < 3; i++ {
		bs, err = redis.Bytes(conn.Do("GET", getCacheKey4SeriesPointMatchInfo(seriesId)))
		if err == redis.ErrNil {
			err = nil
			found = false
			return
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "GetSeriesPointMatchInfo conn.Do error: %v", err)
		return
	}
	err = json.Unmarshal(bs, info)
	if err != nil {
		log.Errorc(ctx, "GetSeriesPointMatchInfo json.Unmarshal error: %v", err)
		return
	}
	return
}

// DelSeriesExtraInfo: 删除阶段的积分表或树状图
func (d *Dao) DelSeriesExtraInfo(ctx context.Context, seriesId int64) (err error) {
	if seriesId == 0 {
		return
	}
	var typ int64
	conn := d.redis.Get(ctx)
	defer conn.Close()

	retry.WithAttempts(ctx, "DelSeriesExtraInfo_QueryRow", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		return d.db.QueryRow(ctx, sql4GetDeletedContestSeriesIdAndType, seriesId).Scan(&typ)
	})

	retry.WithAttempts(ctx, "DelSeriesExtraInfo_Exist", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err := conn.Do("DEL", getCacheKey4SeriesIdExist(seriesId, typ))
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})

	retry.WithAttempts(ctx, "DelSeriesExtraInfo_Extra", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		switch typ {
		case SeriesTypPoint:
			_, err = conn.Do("DEL", getCacheKey4SeriesPointMatchInfo(seriesId))
		case SeriesTypKnockout:
			_, err = conn.Do("DEL", getCacheKey4SeriesKnockoutMatchInfo(seriesId))
		}
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})

	return
}

func (d *Dao) GetSeriesKnockoutMatchConfig(ctx context.Context, id int64) (config *v1.SeriesKnockoutMatchConfig, err error) {
	config = &v1.SeriesKnockoutMatchConfig{}
	exists, empty, err := d.IsSeriesExistsAndExtraConfigEmpty(ctx, id, SeriesTypKnockout)
	if err != nil {
		return
	}
	if !exists {
		err = ecode.EsportsContestSeriesNotFound
		return
	}
	if empty {
		err = ecode.EsportsContestSeriesExtraConfigNotFound
		return
	}
	configContent, err := d.GetSeriesExtraConfig(ctx, id)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(configContent), &config)
	if err != nil {
		log.Errorc(ctx, "GetSeriesKnockoutMatchConfig error: %v", err)
		err = ecode.EsportsContestSeriesExtraConfigErr
	}
	return
}

func (d *Dao) GetSeriesKnockoutMatchInfo(ctx context.Context, seriesId int64) (info *v1.SeriesKnockoutMatchInfo, found bool, err error) {
	info = &v1.SeriesKnockoutMatchInfo{}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	var bs []byte
	found = true
	for i := 0; i < 3; i++ {
		bs, err = redis.Bytes(conn.Do("GET", getCacheKey4SeriesKnockoutMatchInfo(seriesId)))
		if err == redis.ErrNil {
			err = nil
			found = false
			return
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "GetSeriesKnockoutMatchInfo conn.Do error: %v", err)
		return
	}
	err = json.Unmarshal(bs, info)
	if err != nil {
		log.Errorc(ctx, "GetSeriesKnockoutMatchInfo json.Unmarshal error: %v", err)
		return
	}
	return
}

type SeriesNames struct {
	Id          int64
	ParentTitle string
	ChildTitle  string
}

func (d *Dao) GetAllSeriesInSeason(ctx context.Context, seasonId int64) (res map[int64]*SeriesNames, err error) {
	res = make(map[int64]*SeriesNames, 0)
	rows, err := d.db.Query(ctx, sql4GetAllSeriesInSeason, seasonId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &SeriesNames{}
		err = rows.Scan(&r.Id, &r.ParentTitle, &r.ChildTitle)
		if err != nil {
			return
		}
		res[r.Id] = r
	}
	err = rows.Err()
	return
}

func (d *Dao) GetContestsBySeasonId(ctx context.Context, id int64) (res map[int64]*model.Contest, err error) {
	res = make(map[int64]*model.Contest, 0)
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, sql4GetContestIdsByContestSeasonId, id)
	if err != nil {
		log.Errorc(ctx, "GetContestsBySeasonId db.Query error: %v", err)
		return
	}
	defer rows.Close()
	contestIds := make([]int64, 0)
	for rows.Next() {
		contestId := int64(0)
		err = rows.Scan(&contestId)
		if err != nil {
			log.Errorc(ctx, "GetContestsBySeasonId rows.Scan error: %v", err)
			return
		}
		contestIds = append(contestIds, contestId)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	if len(contestIds) == 0 {
		return
	}
	res, err = d.RawEpContests(ctx, contestIds)
	return
}
