package rank

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/xstr"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v3"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	baseDBName      = "act_rank_base"
	ruleDBName      = "act_rank_rule"
	resultLogDBName = "act_rank_result_log"
	ruleBatchDBName = "act_rank_rule_batch"
)

const (
	getBaseOnline = "SELECT id,name,state from %s where state=1"
	getRuleOnline = "SELECT id,base_id,name,statistics_type,update_frequency,update_scope,state,stime,etime from %s where  etime>=? and stime<=?"
	insertNewLog  = `INSERT INTO %s (base_id,rank_id,batch,this_date,last_date) VALUES %s
	ON DUPLICATE KEY UPDATE 
	base_id=VALUES(base_id),
	rank_id=VALUES(rank_id),
	batch=VALUES(batch),
	this_date=VALUES(this_date),
	last_date=VALUES(last_date)
	`
	getRuleBatch = `SELECT
	rule_id,last_batch,last_batch_time,show_batch,show_batch_time
	FROM %s WHERE rule_id in(%s)
	`
	getRankLog = "SELECT base_id,rank_id,batch,state,this_date,last_date from %s where rank_id in (%s) and this_date=?"
)

// GetBaseOnline get rank config online
func (d *dao) GetBaseOnline(c context.Context, now time.Time) (list []*rankmdl.Base, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getBaseOnline, baseDBName))
	if err != nil {
		err = errors.Wrapf(err, "GetBaseOnline:d.db.Query(%v)", now)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rank := new(rankmdl.Base)
		if err = rows.Scan(&rank.ID, &rank.Name, &rank.State, &rank.Stime, &rank.Etime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				list = nil
			} else {
				err = errors.Wrapf(err, "GetBaseOnline:row.Scan row (%v)", now)
			}
			return
		}
		list = append(list, rank)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetBaseOnline:rowsErr(%v)", now)
	}
	return
}

// GetSource get all rank by id
func (d *dao) GetRuleOnline(c context.Context, now time.Time) (list []*rankmdl.Rule, err error) {
	list, err = d.getRule(c, now)
	if err != nil {
		return
	}
	ruleID := make([]int64, 0)

	if len(list) > 0 {
		for _, v := range list {
			ruleID = append(ruleID, v.ID)
		}
		batch, err := d.getRuleBatch(c, ruleID)
		if err != nil {
			return nil, err
		}
		batchMap := make(map[int64]*rankmdl.RuleBatchTime)
		if batch != nil {
			for _, v := range batch {
				batchMap[v.RuleID] = v
			}
		}
		for i, v := range list {
			if b, ok := batchMap[v.ID]; ok {
				list[i].LastBatch = b.LastBatch
				list[i].LastBatchTime = b.LastBatchTime
				list[i].ShowBatchTime = b.ShowBatchTime
				list[i].ShowBatch = b.ShowBatch
			}
		}
	}
	return
}

// GetSourceBatch get all rank by id
func (d *dao) getRuleBatch(c context.Context, ruleIDs []int64) (list []*rankmdl.RuleBatchTime, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRuleBatch, ruleBatchDBName, xstr.JoinInts(ruleIDs)))
	if err != nil {
		err = errors.Wrapf(err, "getRuleBatch:d.db.Query(%+v)", ruleIDs)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.RuleBatchTime)
		if err = rows.Scan(&n.RuleID, &n.LastBatch, &n.LastBatchTime, &n.ShowBatch, &n.ShowBatchTime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "getRuleBatch:row.Scan row (%+v)", ruleIDs)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "getRuleBatch:rowsErr(%+v)", ruleIDs)
	}
	return
}

// GetRuleOnline ...
func (d *dao) getRule(c context.Context, now time.Time) (list []*rankmdl.Rule, err error) {
	nowTime := now.Format("2006-01-02 15:04:05")
	rows, err := d.db.Query(c, fmt.Sprintf(getRuleOnline, ruleDBName), nowTime, nowTime)
	if err != nil {
		err = errors.Wrapf(err, "GetRuleOnline:d.db.Query(%v)", now)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rank := new(rankmdl.Rule)
		if err = rows.Scan(&rank.ID, &rank.BaseID, &rank.Name, &rank.StatisticsType, &rank.UpdateFrequency, &rank.UpdateScope, &rank.State, &rank.Stime, &rank.Etime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				list = nil
			} else {
				err = errors.Wrapf(err, "GetRuleOnline:row.Scan row (%v)", now)
			}
			return
		}
		list = append(list, rank)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetRuleOnline:rowsErr(%v)", now)
	}
	return
}

// GetRankLog 获取日志
func (d *dao) GetRankLog(c context.Context, rankID []int64, thisDate string) (list []*rankmdl.Log, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRankLog, resultLogDBName, xstr.JoinInts(rankID)), thisDate)
	if err != nil {
		err = errors.Wrapf(err, "GetRankLog:d.db.Query(%+v)", rankID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rank := new(rankmdl.Log)
		if err = rows.Scan(&rank.BaseID, &rank.RankID, &rank.Batch, &rank.State, &rank.ThisDate, &rank.LastDate); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				list = nil
			} else {
				err = errors.Wrapf(err, "GetRankLog:row.Scan row (%+v)", rankID)
			}
			return
		}
		list = append(list, rank)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetRankLog:rowsErr(%+v)", rankID)
	}
	return
}

// InsertRankLog batch add rank
func (d *dao) InsertRankLog(c context.Context, rank []*rankmdl.Log) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range rank {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?)")
		rows = append(rows, r.BaseID, r.RankID, r.Batch, r.ThisDate, r.LastDate)
	}
	sql := fmt.Sprintf(insertNewLog, resultLogDBName, strings.Join(rowsTmp, ","))
	if _, err = d.db.Exec(c, sql, rows...); err != nil {
		err = errors.Wrap(err, "InsertRankLog: tx.Exec")
	}
	return
}
