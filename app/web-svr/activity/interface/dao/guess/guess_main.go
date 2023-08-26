package guess

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/xstr"

	"go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	bizLimitKey4DBRestoreOfMainID = "restore_guess_main"

	bizNameOfMainID = "guess_main"

	cacheKey4GuessMainID = "guess:mainIDList:1016:%v:%v"

	secondsOfSevenDays   = 7 * 86400
	sql4HotGuessMainList = `
SELECT id, oid, is_deleted, business
FROM act_guess_main
WHERE (result_id > 0
		AND etime > ?
		AND is_deleted = 0)
	OR (result_id = 0
		AND is_deleted = 0)
`
	sql4AvailableMainList = `
SELECT m.id, m.business, m.oid, m.title, m.stake_type
	, m.max_stake, m.result_id, m.guess_count, m.stime, m.etime, m.template_type
	, d.id AS detail_id, d.option, d.odds, d.total_stake, m.is_deleted
FROM act_guess_main m
	INNER JOIN act_guess_detail d ON m.id = d.main_id
WHERE m.business = ? and m.is_deleted = 0
	AND m.id IN (%s)
ORDER BY m.id, d.id
`
)

func GenHotMainMapByMainIDList(list []*guess.MainID) (m map[string][]*guess.MainID) {
	m = make(map[string][]*guess.MainID, 0)
	for _, v := range list {
		key := guess.GenHotMapKeyByOIDAndBusiness(v.OID, v.Business)
		if _, ok := m[key]; !ok {
			m[key] = make([]*guess.MainID, 0)
		}

		m[key] = append(m[key], v.DeepCopy())
	}

	return
}

func (d *Dao) HotMainResMap(ctx context.Context) (list []*guess.MainID, m map[string]map[int64]*guess.MainRes, err error) {
	list = make([]*guess.MainID, 0)
	m = make(map[string]map[int64]*guess.MainRes, 0)

	list, err = d.HotMainIDList(ctx)
	if err != nil {
		return
	}

	if len(list) > 0 {
		tmpList := make([]int64, 0)
		for _, v := range list {
			tmpList = append(tmpList, v.ID)
		}

		m, err = d.HotMainDetailListByMainIDList(ctx, tmpList, 1)
	}

	return
}

func (d *Dao) AvailableHotMainDetailMap(ctx context.Context, ids []int64, business int64) (m map[int64]*guess.MainRes, err error) {
	var rows *sql.Rows
	m = make(map[int64]*guess.MainRes, 0)

	if len(ids) == 0 {
		return
	}

	if rows, err = d.db.Query(ctx, fmt.Sprintf(sql4AvailableMainList, xstr.JoinInts(ids)), business); err != nil {
		return
	}

	defer rows.Close()

	list := make([]*guess.MainDetail, 0)
	for rows.Next() {
		r := new(guess.MainDetail)
		err = rows.Scan(
			&r.ID,
			&r.Business,
			&r.Oid,
			&r.Title,
			&r.StakeType,
			&r.MaxStake,
			&r.ResultID,
			&r.GuessCount,
			&r.Stime,
			&r.Etime,
			&r.TemplateType,
			&r.DetailID,
			&r.Option,
			&r.Odds,
			&r.TotalStake,
			&r.IsDeleted)
		if err != nil {
			return
		}

		list = append(list, r)
	}

	err = rows.Err()
	if err != nil {
		return
	}

	m = genMainResByDetailList(list)

	return
}

func genMainResByDetailList(list []*guess.MainDetail) (m map[int64]*guess.MainRes) {
	m = make(map[int64]*guess.MainRes, 0)
	for _, v := range list {
		if _, ok := m[v.ID]; !ok {
			m[v.ID] = guess.GenMainResByDetail(nil, v)
		} else {
			m[v.ID] = guess.GenMainResByDetail(m[v.ID], v)
		}
	}

	return
}

func currentDateUnix() int64 {
	year, month, day := time.Now().Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
}

func (d *Dao) HotMainIDList(ctx context.Context) (list []*guess.MainID, err error) {
	editTime := currentDateUnix() - secondsOfSevenDays

	var rows *sql.Rows
	rows, err = d.db.Query(ctx, sql4HotGuessMainList, editTime)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		main := new(guess.MainID)
		err = rows.Scan(&main.ID, &main.OID, &main.IsDeleted, &main.Business)
		if err != nil {
			return
		}

		list = append(list, main)
	}

	err = rows.Err()

	return
}

func (d *Dao) MainIDListByOID(ctx context.Context, oID, business int64) (list []*guess.MainID, err error) {
	list = make([]*guess.MainID, 0)
	canRestore := false
	list, canRestore, err = d.FetchMainIDListFromCacheByOID(ctx, oID, business)
	if err != nil {
		return
	}

	if len(list) > 0 || !canRestore {
		return
	}

	if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizLimitKey4DBRestoreOfMainID) {
		tool.AddDBBackSourceMetrics(bizNameOfMainID)

		list, err = d.RawOidMIDs(ctx, oID, business)
		if err != nil {
			tool.AddDBErrMetrics(bizNameOfMainID)

			return
		}

		_ = d.ResetMainIDListInCache(ctx, oID, business, list)
	}

	return
}

func (d *Dao) DeleteMainIDCache(ctx context.Context, oID, business int64) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	for i := 0; i < 3; i++ {
		_, err = conn.Do("DEL", guessMainIDCacheKeyByOIDAndBusiness(oID, business))
		if err == nil {
			break
		}
	}

	return
}

func (d *Dao) ResetMainIDListInCacheByOID(ctx context.Context, oID, business int64) (list []*guess.MainID, err error) {
	list = make([]*guess.MainID, 0)
	list, err = d.RawOidMIDs(ctx, oID, business)
	if err != nil {
		return
	}

	for i := 0; i < 3; i++ {
		err = d.ResetMainIDListInCache(ctx, oID, business, list)
		if err == nil {
			break
		}
	}

	return
}

func (d *Dao) ResetMainIDListInCache(ctx context.Context, oID, business int64, list []*guess.MainID) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	bs, _ := json.Marshal(list)
	status := tool.StatusOfSucceed
	_, err = conn.Do("SETEX", guessMainIDCacheKeyByOIDAndBusiness(oID, business), tool.CalculateExpiredSeconds(0), bs)
	if err != nil {
		status = tool.StatusOfFailed
	}

	tool.IncrCacheResetMetric(bizLimitKey4DBRestoreOfMainID, status)

	return
}

func guessMainIDCacheKeyByOIDAndBusiness(oID, business int64) string {
	return fmt.Sprintf(cacheKey4GuessMainID, oID, business)
}

func (d *Dao) FetchMainIDListFromCacheByOID(ctx context.Context, oID, business int64) (list []*guess.MainID, canRestore bool, err error) {
	list = make([]*guess.MainID, 0)
	bs := make([]byte, 0)

	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	bs, err = redis.Bytes(conn.Do("GET", guessMainIDCacheKeyByOIDAndBusiness(oID, business)))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			canRestore = true
		}

		return
	}

	_ = json.Unmarshal(bs, &list)

	return
}
