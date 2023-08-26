package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"strconv"
	"time"
)

// SQL
const (
	sql4AddActivity = `
INSERT INTO act_vote_main(name, start_time, end_time, creator, editor, vote_rule)
VALUES(?,?,?,?,?,?)
`

	sql4RawActivitysById = `
SELECT id,
       name,
       start_time,
       end_time,
       last_refresh_time,
       creator,
       editor,
       ctime,
       mtime,
       vote_rule
FROM act_vote_main
WHERE id = ?
  AND is_deleted = 0
ORDER BY id DESC
`

	sql4UpdateActivitysById = `
UPDATE act_vote_main
SET
       name=?,
       start_time=?,
       end_time=?,
       editor=?
WHERE id = ?
  AND is_deleted = 0
`

	sql4DeleteActivitysById = `
UPDATE act_vote_main
SET is_deleted=1
WHERE id = ?
  AND is_deleted = 0
`

	sql4DeleteDSGByActivityId = `
UPDATE act_vote_data_sources_group
SET is_deleted=1
WHERE main_id = ?
  AND is_deleted = 0
`

	sql4RawActivitys = `
SELECT id,
       name,
       start_time,
       end_time,
       last_refresh_time,
       creator,
       editor,
       ctime,
       mtime,
       vote_rule
FROM act_vote_main
WHERE is_deleted = 0
`

	sql4CountActivitys = `
SELECT count(*)
FROM act_vote_main
WHERE is_deleted = 0
`

	sql4RawNotEndActivitysAll = `
SELECT id,
       name,
       start_time,
       end_time,
       last_refresh_time,
       creator,
       editor,
       ctime,
       mtime,
       vote_rule
FROM act_vote_main
WHERE is_deleted = 0
AND end_time >= %v
ORDER BY id DESC
`

	sql4RawEndWithinNActivitysAll = `
SELECT id,
       name,
       start_time,
       end_time,
       last_refresh_time,
       creator,
       editor,
       ctime,
       mtime,
       vote_rule
FROM act_vote_main
WHERE is_deleted = 0
AND end_time <= %v
AND end_time >= %v
ORDER BY id DESC
`

	sql4UpdateActivitysRuleById = `
update act_vote_main set vote_rule = ?
WHERE id = ? AND is_deleted = 0
`

	sql4AddActivitysDSGroup = `
INSERT INTO act_vote_data_sources_group(main_id, source_type, source_id)
VALUES(?,?,?)
`

	sql4DelActivitysDSGroup = `
UPDATE act_vote_data_sources_group
SET is_deleted=1
WHERE main_id=?
  AND id=?
  AND is_deleted = 0
`

	sql4UpdateActivitysDSGroup = `
UPDATE act_vote_data_sources_group
SET source_type=?,
    source_id=?
WHERE main_id=?
  AND id=?
  AND is_deleted = 0
`

	sql4GetDSGroupByActivityId = `
SELECT id,
       main_id,
       source_type,
       source_id
FROM act_vote_data_sources_group
WHERE main_id=?
  AND is_deleted = 0
`
	sql4GetDSGroupById = `
SELECT id,
       main_id,
       source_type,
       source_id
FROM act_vote_data_sources_group
WHERE id=?
  AND is_deleted = 0
`
	sql4UpdateActivityRankRefreshTime = `
UPDATE act_vote_main
SET last_refresh_time=?,
    mtime=mtime
WHERE id=?
`
)

var defaultActivityRule = &api.VoteActivityRule{
	BaseTimes:            1,
	SingleDayLimit:       1,
	TotalLimit:           10,
	SingleOptionBehavior: int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorDayOnce),
	RiskControlRule:      model.RiskControlRuleGeneric,
	DisplayRiskVote:      false,
	DisplayVoteCount:     true,
	VoteUpdateRule:       int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime),
	VoteUpdateCron:       0,
}

/**********************************缓存Key控制**********************************/
func redisActivityConfigCacheKey(activityId int64) string {
	return fmt.Sprintf("vote_act_c_%v", activityId)
}

func redisActivityDSGCacheKey(dataSourceGroupId int64) string {
	return fmt.Sprintf("vote_DSG_Config_%v", dataSourceGroupId)
}

/**********************************活动相关CURD**********************************/

func (d *Dao) RawActivity(ctx context.Context, id int64) (res *api.VoteActivity, err error) {
	res = &api.VoteActivity{}
	tmpVoteRule := ""
	row := d.db.QueryRow(ctx, sql4RawActivitysById, id)
	err = row.Scan(&res.Id, &res.Name, &res.StartTime, &res.EndTime, &res.LastRankRefreshTime,
		&res.Creator, &res.Editor, &res.Ctime,
		&res.Mtime, &tmpVoteRule)
	if tmpVoteRule != "" {
		tmpV := &api.VoteActivityRule{}
		err = json.Unmarshal([]byte(tmpVoteRule), tmpV)
		if err != nil {
			return
		}
		res.Rule = tmpV
	}
	if err == sql.ErrNoRows {
		err = nil
		res.Id = -1
	}
	if err != nil {
		return
	}
	return
}

func (d *Dao) AddActivity(ctx context.Context, req *api.AddVoteActivityReq) (err error) {
	if req.StartTime > req.EndTime {
		err = ecode.ActivityVoteRuleConfigError
		return
	}
	bs, _ := json.Marshal(defaultActivityRule)
	_, err = d.db.Exec(ctx, sql4AddActivity, req.Name, req.StartTime, req.EndTime, req.Creator, req.Creator, string(bs))
	return
}

func (d *Dao) DelActivity(ctx context.Context, req *api.DelVoteActivityReq) (err error) {
	tx, err := d.db.Begin(ctx)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	res, err := tx.Exec(sql4DeleteActivitysById, req.Id)
	if err != nil {
		return
	}
	rf, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rf == 0 {
		err = ecode.ActivityNotExist
		return
	}
	_, err = tx.Exec(sql4DeleteDSGByActivityId, req.Id)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (d *Dao) UpdateActivity(ctx context.Context, req *api.UpdateVoteActivityReq) (err error) {
	if req.StartTime > req.EndTime {
		err = ecode.ActivityVoteRuleConfigError
		return
	}
	res, err := d.db.Exec(ctx, sql4UpdateActivitysById, req.Name, req.StartTime, req.EndTime, req.Editor, req.Id)
	if err != nil {
		return
	}
	rf, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rf == 0 {
		err = ecode.ActivityNotExist
		return
	}
	return
}

func (d *Dao) ListVoteActivityForRefresh(ctx context.Context, req *api.ListVoteActivityForRefreshReq) (res *api.ListVoteActivityForRefreshResp, err error) {
	res = &api.ListVoteActivityForRefreshResp{
		Activitys: make([]*api.VoteActivity, 0),
	}
	var (
		rows   *sql.Rows
		sqlStr string
	)
	switch req.Type {
	case api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeNotEnded:
		sqlStr = fmt.Sprintf(sql4RawNotEndActivitysAll, time.Now().Unix())
	case api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeEndWithin90:
		sqlStr = fmt.Sprintf(sql4RawEndWithinNActivitysAll, time.Now().Unix(), time.Now().AddDate(0, 0, -90).Unix())
	default:
		sqlStr = fmt.Sprintf(sql4RawNotEndActivitysAll, time.Now().Unix())
	}

	rows, err = d.db.Query(ctx, sqlStr)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}

	}()
	for rows.Next() {
		tmp := &api.VoteActivity{}
		tmpVoteRule := ""
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.StartTime, &tmp.EndTime, &tmp.LastRankRefreshTime,
			&tmp.Creator, &tmp.Editor, &tmp.Ctime,
			&tmp.Mtime, &tmpVoteRule)
		if err != nil {
			return
		}
		if tmpVoteRule != "" {
			tmpV := &api.VoteActivityRule{}
			err = json.Unmarshal([]byte(tmpVoteRule), tmpV)
			if err != nil {
				return
			}
			tmp.Rule = tmpV
		}
		res.Activitys = append(res.Activitys, tmp)
	}
	return
}

func (d *Dao) ListActivity(ctx context.Context, req *api.ListVoteActivityReq) (res *api.ListVoteActivityResp, err error) {
	res = &api.ListVoteActivityResp{
		Page: &api.VotePage{
			Num:   0,
			Ps:    0,
			Total: 0,
		},
		Activitys: make([]*api.VoteActivity, 0),
	}
	var (
		listSql  string
		countSql string
		rows     *sql.Rows
	)

	listSql = sql4RawActivitys
	countSql = sql4CountActivitys

	start := req.Ps * (req.Pn - 1)
	end := req.Ps*req.Pn - 1
	switch req.Ongoing {
	case 1:
		cond := fmt.Sprintf(" AND start_time <= %v AND end_time >= %v", time.Now().Unix(), time.Now().Unix())
		listSql = listSql + cond
		countSql = countSql + cond
	case 2:
		cond := fmt.Sprintf(" AND (end_time < %v or start_time > %v)", time.Now().Unix(), time.Now().Unix())
		listSql = listSql + cond
		countSql = countSql + cond
	}
	if req.Keyword != "" {
		cond := fmt.Sprintf(" AND ( name like '%%%v%%'", req.Keyword)
		if kInt, err := strconv.Atoi(req.Keyword); err == nil {
			cond = cond + fmt.Sprintf(" OR id like %v", kInt)
		}
		cond = cond + " ) "
		listSql = listSql + cond
		countSql = countSql + cond
	}
	listSql = listSql + " ORDER BY id DESC"
	listSql = fmt.Sprintf("%v LIMIT %v, %v", listSql, start, end)
	count := int64(0)
	err = d.db.QueryRow(ctx, countSql).Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return
	}
	res.Page.Ps = req.Ps
	res.Page.Num = req.Pn
	res.Page.Total = count

	rows, err = d.db.Query(ctx, listSql)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}
	}()
	for rows.Next() {
		tmp := &api.VoteActivity{}
		tmpVoteRule := ""
		err = rows.Scan(&tmp.Id, &tmp.Name, &tmp.StartTime,
			&tmp.EndTime, &tmp.LastRankRefreshTime,
			&tmp.Creator, &tmp.Editor, &tmp.Ctime,
			&tmp.Mtime, &tmpVoteRule)
		if err != nil {
			return
		}
		if tmpVoteRule != "" {
			tmpV := &api.VoteActivityRule{}
			err = json.Unmarshal([]byte(tmpVoteRule), tmpV)
			if err != nil {
				return
			}
			tmp.Rule = tmpV
		}
		res.Activitys = append(res.Activitys, tmp)
	}
	return
}

func (d *Dao) UpdateActivityRule(ctx context.Context, req *api.UpdateVoteActivityRuleReq) (err error) {
	if req.ActivityId == 0 {
		err = ecode.SystemActivityConfigErr
		return
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return
	}
	_, err = d.db.Exec(ctx, sql4UpdateActivitysRuleById, string(bs), req.ActivityId)
	if err != nil {
		return
	}
	return
}

/**********************************数据组相关CURD**********************************/

func (d *Dao) validateDataSourceGroup(ctx context.Context, sourceType string, sourceId int64) (err error) {
	dsI, ok := d.datasourceMap[sourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	_, err = dsI.ListAllItems(ctx, sourceId)
	return
}
func (d *Dao) AddActivityDataSourceGroup(ctx context.Context, req *api.AddVoteActivityDataSourceGroupReq) (err error) {
	if req.ActivityId == 0 {
		err = ecode.SystemActivityConfigErr
		return
	}
	err = d.validateDataSourceGroup(ctx, req.SourceType, req.SourceId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "数据组配置校验失败: %v", err.Error())
		return
	}
	_, err = d.db.Exec(ctx, sql4AddActivitysDSGroup, req.ActivityId, req.SourceType, req.SourceId)
	return
}

func (d *Dao) DelActivityDataSourceGroup(ctx context.Context, req *api.DelVoteActivityDataSourceGroupReq) (err error) {
	if req.GroupId == 0 {
		err = ecode.ActivityVoteDSGNotFound
		return
	}
	res, err := d.db.Exec(ctx, sql4DelActivitysDSGroup, req.ActivityId, req.GroupId)
	if err != nil {
		return
	}
	rf, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rf == 0 {
		err = ecode.ActivityVoteDSGNotFound
		return
	}
	return
}

func (d *Dao) ListActivityDataSourceGroups(ctx context.Context, req *api.ListVoteActivityDataSourceGroupsReq) (res *api.ListVoteActivityDataSourceGroupsResp, err error) {
	res = &api.ListVoteActivityDataSourceGroupsResp{Groups: make([]*api.VoteDataSourceGroupItem, 0)}
	if req.ActivityId == 0 {
		err = ecode.SystemActivityConfigErr
		return
	}
	rows, err := d.db.Query(ctx, sql4GetDSGroupByActivityId, req.ActivityId)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}
	}()
	for rows.Next() {
		ds := &api.VoteDataSourceGroupItem{}
		err = rows.Scan(&ds.GroupId, &ds.ActivityId, &ds.SourceType, &ds.SourceId)
		if err != nil {
			return
		}
		res.Groups = append(res.Groups, ds)
	}
	err = rows.Err()
	return
}

func (d *Dao) rawActivityDataSourceGroup(ctx context.Context, dataSourceGroupId int64) (res *api.VoteDataSourceGroupItem, err error) {
	res = &api.VoteDataSourceGroupItem{}
	row := d.db.QueryRow(ctx, sql4GetDSGroupById, dataSourceGroupId)
	err = row.Scan(&res.GroupId, &res.ActivityId, &res.SourceType, &res.SourceId)
	if err == sql.ErrNoRows {
		err = nil
		res.GroupId = -1
	}
	return
}

func (d *Dao) UpdateActivityDataSourceGroup(ctx context.Context, req *api.UpdateVoteActivityDataSourceGroupReq) (err error) {
	if req.GroupId == 0 {
		err = ecode.ActivityVoteDSGNotFound
		return
	}
	err = d.validateDataSourceGroup(ctx, req.SourceType, req.SourceId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "数据组配置校验失败: %v", err.Error())
		return
	}
	tmp, err := d.rawActivityDataSourceGroup(ctx, req.GroupId)
	if err != nil {
		return
	}
	if tmp.GroupId == -1 {
		err = ecode.ActivityVoteDSGNotFound
		return
	}
	_, err = d.db.Exec(ctx, sql4UpdateActivitysDSGroup, req.SourceType, req.SourceId, req.ActivityId, req.GroupId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "数据组修改: %v", err.Error())
		return
	}
	return
}

func (d *Dao) updateActivityRankRefreshTime(ctx context.Context, activityId int64) (err error) {
	err = retry.WithAttempts(ctx, "updateActivityRankRefreshTime", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.db.Exec(ctx, sql4UpdateActivityRankRefreshTime, time.Now().Unix(), activityId)
		return err
	})
	return
}
