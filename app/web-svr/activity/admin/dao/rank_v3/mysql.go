package rank

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	xsql "go-common/library/database/sql"
	"go-common/library/xstr"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank_v3"

	"go-common/library/log"
)

const (
	baseDBName          = "act_rank_base"
	sourceDBName        = "act_rank_source"
	ruleDBName          = "act_rank_rule"
	scoreConfigDBName   = "act_rank_score_config"
	blackWhiteDBName    = "act_rank_black_white"
	adjustDBName        = "act_rank_adjust"
	archiveDBName       = "act_rank_archives"
	upDBName            = "act_rank_ups"
	tagDBName           = "act_rank_tags"
	ruleBatchDBName     = "act_rank_rule_batch"
	resultDBName        = "act_rank_result_%d"
	resultArchiveDBName = "act_rank_result_archive_%d"
)

const (
	addNewRank = `INSERT INTO %s 
	(name,is_show_score,rank_type,is_type,tids,archive_stime,archive_etime,author,authority) 
	VALUES(?,?,?,?,?,?,?,?,?)`
	updateRank             = "UPDATE %s set name=?,is_show_score=?,rank_type=?,is_type=?,tids=?,archive_stime=?,archive_etime=?,authority=?,state=? where id=?"
	updateRuleState        = "UPDATE %s set state=? where id in (%s)"
	updateSourceState      = "UPDATE %s set state=? where base_id = ?"
	addRankSource          = "INSERT INTO %s (base_id,source_id,source_type,state) VALUES %s ON DUPLICATE KEY UPDATE base_id=VALUES(base_id),source_id=VALUES(source_id),source_type=VALUES(source_type),state=VALUES(state)"
	addRule                = "INSERT INTO %s (base_id,name,statistics_type,nums,update_frequency,update_scope,state,`precision`,unit,description,stime,etime) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	updateRule             = "UPDATE %s SET name=?,nums=?,stime=?,etime=?,state=? where id=?"
	UpdateRankRuleShow     = "UPDATE %s SET `precision`=?,unit=?,description=? where id=?"
	updateOneRuleState     = "UPDATE %s SET state=? where id=?"
	baseList               = "SELECT id,name,rank_type,is_type,tids,archive_stime,archive_etime,author,authority,state FROM act_rank_base %s order by id desc LIMIT ? OFFSET ?"
	sourceList             = "SELECT id,base_id,source_id,source_type,state,ctime,mtime FROM act_rank_source where base_id=? and state=1  LIMIT ? OFFSET ?"
	createResult           = "CREATE TABLE IF NOT EXISTS act_rank_result_%d LIKE act_rank_result"
	createAidSnapshot      = "CREATE TABLE IF NOT EXISTS act_rank_result_archive_%d LIKE act_rank_result_archive"
	listTotal              = "SELECT count(1) FROM  act_rank_base %s"
	sourceListTotal        = "SELECT count(1) FROM  act_rank_source where base_id=? and state=1"
	addOrUpdateScoreConfig = `INSERT INTO %s (rank_id,action,base,cnt_per_day) 
	VALUES %s ON DUPLICATE KEY UPDATE 
	rank_id=VALUES(rank_id),
	action=VALUES(action),
	base=VALUES(base),
	cnt_per_day=VALUES(cnt_per_day)`
	addOrUpdateBlackWhite = `INSERT INTO %s (base_id,oid,score,state,intervention_type,object_type )
		VALUES %s ON DUPLICATE KEY UPDATE 
		base_id=VALUES(base_id),
		oid=VALUES(oid),
		score=VALUES(score),
		state=VALUES(state),
		intervention_type=VALUES(intervention_type),
		object_type=VALUES(object_type)`
	getRank = `SELECT 
	id,name,rank_type,is_type,tids,archive_stime,archive_etime,author,authority,state,ctime,mtime 
	FROM %s WHERE id=?`
	getRankByState = "SELECT id,name,base_id,name,statistics_type,nums,update_frequency,update_scope,state,`precision`,unit,description,stime,etime,ctime,mtime FROM %s WHERE state=? %s"
	getSource      = `SELECT
	id,base_id,source_id,source_type,state,ctime,mtime
	FROM %s WHERE base_id=? and source_type = ? and state=1
	`
	getSourceBatch = `SELECT
	id,base_id,source_id,source_type,state,ctime,mtime
	FROM %s WHERE base_id in(%s) and state=1
	`
	getRuleBatch = `SELECT
	rule_id,last_batch,last_batch_time
	FROM %s WHERE rule_id in(%s)
	`
	getScoreBatch = `SELECT
	rank_id,action,base,cnt_per_day
	FROM %s WHERE rank_id in(%s)
	`
	updateRuleBatch = "UPDATE %s set show_batch=?, show_batch_time =? where id = ?"
	getRule         = "SELECT id,base_id,name,statistics_type,nums,update_frequency,update_scope,show_batch,show_batch_time,state,`precision`,unit,description,stime,etime,ctime,mtime FROM %s WHERE base_id=? and state!=2 order by id desc"
	getRuleByID     = "SELECT id,base_id,name,statistics_type,nums,update_frequency,update_scope,show_batch,show_batch_time,state,`precision`,unit,description,stime,etime,ctime,mtime FROM %s WHERE id=?"
	getBlackOrWhite = `SELECT
	id,base_id,oid,score,state,intervention_type,object_type,ctime,mtime
	FROM %s WHERE base_id=? and intervention_type = ? and object_type = ? and state = 1
	`
	addOrUpdateAdjust = `INSERT INTO %s (base_id,rank_id,parent_id,oid,object_type,rank,is_show,state) 
	VALUES %s ON DUPLICATE KEY UPDATE 
	base_id=VALUES(base_id),
	parent_id=VALUES(parent_id),
	rank_id=VALUES(rank_id),
	oid=VALUES(oid),
	object_type=VALUES(object_type),
	rank=VALUES(rank),
	is_show=VALUES(is_show),
	state=VALUES(state)
	`
	rankArchive = `SELECT
	id,base_id,aid,rank_id,batch,source_id,log_id,score,mid,tag_id,rank_type,play_score,likes_score,coin_score,share_score,rank,count_score,white_score,ctime,mtime
	FROM %s WHERE base_id=? and rank_id = ? and batch = ? %s order by score desc limit ?,?`
	rankUp = `SELECT
	id,base_id,rank_id,batch,source_id,log_id,score,mid,rank_type,play_score,likes_score,coin_score,share_score,rank,count_score,white_score,fans_score,ctime,mtime
	FROM %s WHERE base_id=? and rank_id = ? and batch = ? %s order by score desc limit 1000`
	rankTag = `SELECT
	id,base_id,rank_id,batch,source_id,log_id,score,tag_id,rank_type,play_score,likes_score,coin_score,share_score,rank,count_score,white_score,ctime,mtime
	FROM %s WHERE base_id=? and rank_id = ? and batch = ? %s order by score desc limit 500`
	getAdjust = `SELECT id,base_id,parent_id,rank_id,oid,object_type,rank,is_show,state,ctime,mtime from %s where base_id=? and rank_id=? and object_type=? and state =1
	`
	updateRankArchive = `UPDATE %s set white_score=?,score=count_score+white_score where base_id=? and aid =? and batch=?
	`
	updateUpArchive = `UPDATE %s set white_score=?,score=count_score+white_score where base_id=? and mid =? and batch=?
	`
	updateTagArchive = `UPDATE %s set white_score=?,score=count_score+white_score where base_id=? and tag_id =? and batch=?
	`
	insertOidResultSQL  = "INSERT INTO %s (base_id,rank_id,oid,rank,score,batch,show_score) VALUES %s"
	insertOidArchiveSQL = "INSERT INTO %s (base_id,rank_id,aid,oid,rank,score,batch,show_score) VALUES %s"
	getRankArchive      = `SELECT id,base_id,rank_id,oid,aid,rank,score,batch,state from %s 
	where rank_id=? and batch=? and state=1  order by rank asc
	`
	getRankOid = `SELECT id,base_id,rank_id,oid,rank,score,batch,state from %s 
	where rank_id=? and batch=? order by rank asc
	`
)

// BatchOidResult batch add college
func (d *dao) BatchOidResult(c context.Context, tx *xsql.Tx, baseID int64, result []*rankmdl.ResultOid) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range result {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?)")

		rows = append(rows, r.BaseID, r.RankID, r.OID, r.Rank, r.Score, r.Batch, r.ShowScore)
	}
	sql := fmt.Sprintf(insertOidResultSQL, fmt.Sprintf(resultDBName, baseID), strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchOidResult: tx.Exec")
	}
	return
}

// BatchOidResult batch add college
func (d *dao) BatchOidArchiveResult(c context.Context, tx *xsql.Tx, baseID int64, result []*rankmdl.ResultOidArchive) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range result {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?)")

		rows = append(rows, r.BaseID, r.RankID, r.AID, r.OID, r.Rank, r.Score, r.Batch, r.ShowScore)
	}
	sql := fmt.Sprintf(insertOidArchiveSQL, fmt.Sprintf(resultArchiveDBName, baseID), strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchOidArchiveResult: tx.Exec")
	}
	return
}

// UpdateTagArchive 更新榜单结果
func (d *dao) UpdateRankTag(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(updateTagArchive, tagDBName), result.WhiteScore, result.BaseID, result.TagID, result.Batch); err != nil {
		log.Errorc(c, "rankv2@UpdateRankTag d.db.Exec() update failed. error(%v)", err)
		return
	}
	return
}

// UpdateRankAdjust 更新榜单结果
func (d *dao) UpdateRankUp(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(updateUpArchive, upDBName), result.WhiteScore, result.BaseID, result.MID, result.Batch); err != nil {
		log.Errorc(c, "rankv2@UpdateRuleBatch d.db.Exec() update failed. error(%v)", err)
		return
	}
	return
}

// UpdateRankAdjust 更新榜单结果
func (d *dao) UpdateRankArchive(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(updateRankArchive, archiveDBName), result.WhiteScore, result.BaseID, result.AID, result.Batch); err != nil {
		log.Errorc(c, "rankv2@UpdateRuleBatch d.db.Exec() update failed. error(%v)", err)
		return
	}
	return
}

// GetAdjust get all rank by id
func (d *dao) GetAdjust(c context.Context, baseID int64, rankID int64, objectType int) (list []*rankmdl.Adjust, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getAdjust, adjustDBName), baseID, rankID, objectType)
	if err != nil {
		err = errors.Wrapf(err, "GetAdjust:d.db.Query(%d)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Adjust)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.ParentID, &n.RankID, &n.OID, &n.ObjectType, &n.Rank, &n.IsShow, &n.State, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetAdjust:row.Scan row (%d)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetAdjust:rowsErr(%d)", baseID)
	}
	return
}

// RankUp get all rank by id
func (d *dao) RankUp(c context.Context, baseID, rankID int64, batch int, mid []int64) (list []*rankmdl.Result, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, baseID, rankID, batch)
	if len(mid) > 0 {
		if len(mid) > 0 {
			sqlAdd += fmt.Sprintf(" AND mid in (%s) ", xstr.JoinInts(mid))
		}
	}
	rows, err := d.db.Query(c, fmt.Sprintf(rankUp, upDBName, sqlAdd), args...)
	if err != nil {
		err = errors.Wrapf(err, "RankUp:d.db.Query(%d)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Result)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.Batch, &n.SourceID, &n.LogID, &n.Score, &n.MID, &n.RankType, &n.PlayScore, &n.LikesScore, &n.CoinScore, &n.ShareScore, &n.Rank, &n.CountScore, &n.WhiteScore, &n.FansScore, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "RankUp:row.Scan row (%d)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RankUp:rowsErr(%d)", baseID)
	}
	return
}

// RankUp get all rank by id
func (d *dao) RankTag(c context.Context, baseID, rankID int64, batch int, tagID []int64) (list []*rankmdl.Result, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, baseID, rankID, batch)

	if len(tagID) > 0 {
		if len(tagID) > 0 {
			sqlAdd += fmt.Sprintf(" AND tag_id in (%s) ", xstr.JoinInts(tagID))
		}
	}
	rows, err := d.db.Query(c, fmt.Sprintf(rankTag, tagDBName, sqlAdd), args...)
	if err != nil {
		err = errors.Wrapf(err, "RankTag:d.db.Query(%d)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Result)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.Batch, &n.SourceID, &n.LogID, &n.Score, &n.TagID, &n.RankType, &n.PlayScore, &n.LikesScore, &n.CoinScore, &n.ShareScore, &n.Rank, &n.CountScore, &n.WhiteScore, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "RankTag:row.Scan row (%d)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RankTag:rowsErr(%d)", baseID)
	}
	return
}

// RankDayArchive get all rank by id
func (d *dao) RankArchive(c context.Context, baseID, rankID int64, batch int, tagID []int64, mid []int64, offset, limit int) (list []*rankmdl.Result, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, baseID, rankID, batch)
	if len(tagID) > 0 || len(mid) > 0 {
		if len(tagID) > 0 {
			sqlAdd += fmt.Sprintf(" AND tag_id in (%s) ", xstr.JoinInts(tagID))
		}
		if len(mid) > 0 {
			sqlAdd += fmt.Sprintf(" AND mid in (%s) ", xstr.JoinInts(mid))
		}

	}
	args = append(args, offset, limit)
	rows, err := d.db.Query(c, fmt.Sprintf(rankArchive, archiveDBName, sqlAdd), args...)
	if err != nil {
		err = errors.Wrapf(err, "RankArchive:d.db.Query(%d)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Result)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.AID, &n.RankID, &n.Batch, &n.SourceID, &n.LogID, &n.Score, &n.MID, &n.TagID, &n.RankType, &n.PlayScore, &n.LikesScore, &n.CoinScore, &n.ShareScore, &n.Rank, &n.CountScore, &n.WhiteScore, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "RankArchive:row.Scan row (%d)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RankArchive:rowsErr(%d)", baseID)
	}
	return
}

// GetBlackOrWhite get all rank by id
func (d *dao) GetBlackOrWhite(c context.Context, baseID int64, interventionType, objectType int) (list []*rankmdl.BlackWhite, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getBlackOrWhite, blackWhiteDBName), baseID, interventionType, objectType)
	if err != nil {
		err = errors.Wrapf(err, "GetBlackOrWhite:d.db.Query(%d)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.BlackWhite)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.Oid, &n.Score, &n.State, &n.InterventionType, &n.ObjectType, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetBlackOrWhite:row.Scan row (%d)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetBlackOrWhite:rowsErr(%d)", baseID)
	}
	return
}

// ListTotal get list information total
func (d *dao) ListTotal(c context.Context, state int, keyword string, rankType int, validateTime int64) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if state != 0 || keyword != "" || rankType != 0 || validateTime != 0 {
		sqlAdd = "WHERE "
		flag := false
		if state != 0 {
			args = append(args, state)
			sqlAdd += "state=? "
			flag = true
		}
		if rankType != 0 {
			args = append(args, rankType)

			if flag {
				sqlAdd += " AND "
			}
			sqlAdd += "rank_type=? "
			flag = true
		}
		if int64(validateTime) != 0 {
			args = append(args, validateTime, validateTime)

			if flag {
				sqlAdd += " AND "
			}
			sqlAdd += "archive_stime <= ? and archive_etime>=? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(name LIKE ? OR id LIKE ? or author LIKE ?)"
		}
	}
	result := d.db.QueryRow(c, fmt.Sprintf(listTotal, sqlAdd), args...)
	if err = result.Scan(&total); err != nil {
		log.Error("rank@ListTotal result.Scan() failed. error(%v)", err)
	}
	return
}

// SourceListTotal get list information total
func (d *dao) SourceListTotal(c context.Context, baseID int64) (total int, err error) {

	result := d.db.QueryRow(c, sourceListTotal, baseID)
	if err = result.Scan(&total); err != nil {
		log.Error("rank@SourceListTotal result.Scan() failed. error(%v)", err)
	}
	return
}

// GetSourceList get lottery base information list
func (d *dao) GetSourceList(c context.Context, pn, ps int, baseID int64) (list []*rankmdl.Source, err error) {
	var (
		args []interface{}
		rows *xsql.Rows
	)
	args = append(args, baseID)
	args = append(args, ps)
	args = append(args, (pn-1)*ps)
	if rows, err = d.db.Query(c, sourceList, args...); err != nil {
		log.Errorc(c, "lottery@GetSourceList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &rankmdl.Source{}
		if err = rows.Scan(&tmp.ID, &tmp.BaseID, &tmp.SourceID, &tmp.SourceType, &tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "rank@GetSourceList rows.Scan() failed. error(%v)", err)
			return
		}
		list = append(list, tmp)
	}
	err = rows.Err()
	return
}

// GetRankList get lottery base information list
func (d *dao) GetRankList(c context.Context, pn, ps, state int, keyword string, rankType int, validateTime int64) (list []*rankmdl.Base, err error) {
	var (
		sqlAdd string
		args   []interface{}
		rows   *xsql.Rows
	)
	if state != 0 || keyword != "" || rankType != 0 || validateTime != 0 {
		sqlAdd = "WHERE "
		flag := false
		if state != 0 {
			args = append(args, state)
			sqlAdd += "state=? "
			flag = true
		}
		if rankType != 0 {
			if flag {
				sqlAdd += " AND "
			}
			args = append(args, rankType)
			sqlAdd += "rank_type=? "
			flag = true
		}
		if int64(validateTime) != 0 {
			args = append(args, validateTime, validateTime)

			if flag {
				sqlAdd += " AND "
			}
			sqlAdd += "archive_stime < ? and archive_etime >? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(name LIKE ? OR id LIKE ? or author LIKE ?)"
		}
	}
	args = append(args, ps)
	args = append(args, (pn-1)*ps)
	if rows, err = d.db.Query(c, fmt.Sprintf(baseList, sqlAdd), args...); err != nil {
		log.Error("lottery@BaseList d.db.Query() failed. error(%v)", err)
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &rankmdl.Base{}
		if err = rows.Scan(&tmp.ID, &tmp.Name, &tmp.RankType, &tmp.IsType, &tmp.Tids, &tmp.ArchiveStime, &tmp.ArchiveEtime, &tmp.Author, &tmp.Authority, &tmp.State); err != nil {
			log.Errorc(c, "rank@BaseList rows.Scan() failed. error(%v)", err)
			return
		}
		if tmp.Tids != "" {
			tmp.TidsStruct, _ = xstr.SplitInts(tmp.Tids)
		}
		list = append(list, tmp)
	}
	err = rows.Err()
	return
}

// GetSource get all rank by id
func (d *dao) GetRule(c context.Context, baseID int64) (list []*rankmdl.Rule, err error) {
	list, err = d.getRule(c, baseID)
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
			log.Errorc(c, "d.getRuleBatch (%v)", err)
			return nil, err
		}
		batchMap := make(map[int64]*rankmdl.RuleBatchTime)
		if batch != nil {
			for _, v := range batch {
				batchMap[v.RuleID] = v
			}
		}
		score, err := d.getRankScoreConfig(c, ruleID)
		if err != nil {
			log.Errorc(c, "d.getRuleBatch (%v)", err)
			return nil, err
		}
		for i, v := range list {
			if b, ok := batchMap[v.ID]; ok {
				list[i].LastBatch = b.LastBatch
				list[i].LastBatchTime = b.LastBatchTime
			}
			if s, ok := score[v.ID]; ok {
				list[i].Score = s
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
		if err = rows.Scan(&n.RuleID, &n.LastBatch, &n.LastBatchTime); err != nil {
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
func (d *dao) getRule(c context.Context, baseID int64) (list []*rankmdl.Rule, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRule, ruleDBName), baseID)
	if err != nil {
		err = errors.Wrapf(err, "GetRule:d.db.Query(%v)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Rule)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.Name, &n.StatisticsType, &n.Nums, &n.UpdateFrequency, &n.UpdateScope, &n.ShowBatch, &n.ShowBatchTime, &n.State, &n.Precision, &n.Unit, &n.Description, &n.Stime, &n.Etime, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetRule:row.Scan row (%v)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetRule:rowsErr(%v)", baseID)
	}
	return
}

func (d *dao) GetRuleByID(c context.Context, id int64) (n *rankmdl.Rule, err error) {
	n, err = d.getRuleByID(c, id)
	if err != nil {
		log.Errorc(c, "d.getRuleByID err(%v)", err)
		return
	}
	if n != nil {
		ruleID := make([]int64, 0)
		ruleID = append(ruleID, n.ID)
		batch, err := d.getRuleBatch(c, ruleID)
		if err != nil {
			log.Errorc(c, "d.getRuleBatch (%v)", err)
			return nil, err
		}
		batchMap := make(map[int64]*rankmdl.RuleBatchTime)
		if batch != nil {
			for _, v := range batch {
				batchMap[v.RuleID] = v
			}
		}
		if b, ok := batchMap[n.ID]; ok {
			n.LastBatch = b.LastBatch
			n.LastBatchTime = b.LastBatchTime
		}
		score, err := d.getRankScoreConfig(c, ruleID)
		if err != nil {
			log.Errorc(c, "d.getRuleBatch (%v)", err)
			return nil, err
		}
		if s, ok := score[n.ID]; ok {
			n.Score = s
		}
	}
	return
}

// GetSource get all rank by id
func (d *dao) getRuleByID(c context.Context, id int64) (n *rankmdl.Rule, err error) {

	var (
		arg []interface{}
	)
	n = &rankmdl.Rule{}
	arg = append(arg, id)
	row := d.db.QueryRow(c, fmt.Sprintf(getRuleByID, ruleDBName), arg...)
	if err = row.Scan(&n.ID, &n.BaseID, &n.Name, &n.StatisticsType, &n.Nums, &n.UpdateFrequency, &n.UpdateScope, &n.ShowBatch, &n.ShowBatchTime, &n.State, &n.Precision, &n.Unit, &n.Description, &n.Stime, &n.Etime, &n.Ctime, &n.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankConfigByID d.db.QueryRow() SELECT failed. error(%v)", err)
		return
	}

	return
}

// GetSourceBatch get all rank by id
func (d *dao) GetSourceBatch(c context.Context, baseIDs []int64) (list []*rankmdl.Source, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getSourceBatch, sourceDBName, xstr.JoinInts(baseIDs)))
	if err != nil {
		err = errors.Wrapf(err, "GetSource:d.db.Query(%+v)", baseIDs)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Source)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.SourceID, &n.SourceType, &n.State, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetSource:row.Scan row (%+v)", baseIDs)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetSource:rowsErr(%+v)", baseIDs)
	}
	return
}

// GetSource get all rank by id
func (d *dao) GetSource(c context.Context, baseID int64, sourceType int) (list []*rankmdl.Source, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getSource, sourceDBName), baseID, sourceType)
	if err != nil {
		err = errors.Wrapf(err, "GetSource:d.db.Query(%v)", baseID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Source)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.SourceID, &n.SourceType, &n.State, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetSource:row.Scan row (%v)", baseID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetSource:rowsErr(%v)", baseID)
	}
	return
}

// GetRankByID ...
func (d *dao) GetRankByID(c context.Context, id int64) (rank *rankmdl.Base, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Base{}
	arg = append(arg, id)
	row := d.db.QueryRow(c, fmt.Sprintf(getRank, baseDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.Name, &rank.RankType, &rank.IsType, &rank.Tids, &rank.ArchiveStime, &rank.ArchiveEtime, &rank.Author, &rank.Authority, &rank.State, &rank.Ctime, &rank.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankConfigByID d.db.QueryRow() SELECT failed. error(%v)", err)
		return
	}
	rank.TidsStruct, _ = xstr.SplitInts(rank.Tids)
	return
}

// GetRankByStateAndTime get all rank by id
func (d *dao) GetRankByStateAndTime(c context.Context, state int, sqlCondition string) (list []*rankmdl.Rule, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRankByState, ruleDBName, sqlCondition), state)
	if err != nil {
		err = errors.Wrapf(err, "GetRankByStateAndTime:d.db.Query(%v)", state)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rank := new(rankmdl.Rule)

		if err = rows.Scan(&rank.ID, &rank.Name, &rank.BaseID, &rank.Name, &rank.StatisticsType, &rank.Nums, &rank.UpdateFrequency, &rank.UpdateScope, &rank.State, &rank.Precision, &rank.Unit, &rank.Description, &rank.Stime, &rank.Etime, &rank.Ctime, &rank.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "GetRankByStateAndTime:row.Scan row (%v)", state)
			return
		}

		list = append(list, rank)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetRankByStateAndTime:rowsErr(%v)", state)
	}
	return
}

// Create new rank
func (d *dao) Create(c context.Context, tx *xsql.Tx, rank *rankmdl.Base, sources []*rankmdl.Source, blackwhite []*rankmdl.BlackWhite) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = tx.Exec(fmt.Sprintf(addNewRank, baseDBName), rank.Name, rank.IsShowScore, rank.RankType, rank.IsType, rank.Tids, rank.ArchiveStime, rank.ArchiveEtime, rank.Author, rank.Authority); err != nil {
		log.Errorc(c, "rankv2@Add d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Errorc(c, "rankv2@Add result.LastInsertId() failed. error(%v)", err)
		return
	}
	if err = d.addBlackWhite(c, tx, id, blackwhite); err != nil {
		log.Errorc(c, "rank@Add d.addBlackWhite(%d) failed. error(%v)", id, err)
		return
	}

	if err = d.addRankSource(c, tx, id, sources); err != nil {
		log.Errorc(c, "rank@Add d.createAddress(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createResult(c, tx, id); err != nil {
		log.Errorc(c, "rank@Add d.createResult(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createAidSnapshot(c, tx, id); err != nil {
		log.Errorc(c, "rank@Add d.createAddTimes(%d) failed. error(%v)", id, err)
		return
	}

	return
}

func (d *dao) AddBlackWhite(c context.Context, id int64, whiteBlack []*rankmdl.BlackWhite) (err error) {
	var (
		sqls = make([]string, 0, len(whiteBlack))
		args = make([]interface{}, 0)
	)
	if len(whiteBlack) == 0 {
		return
	}
	for _, v := range whiteBlack {
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, id, v.Oid, v.Score, v.State, v.InterventionType, v.ObjectType)
	}
	_, err = d.db.Exec(c, fmt.Sprintf(addOrUpdateBlackWhite, blackWhiteDBName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "addRankSource:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

func (d *dao) addBlackWhite(c context.Context, tx *xsql.Tx, id int64, whiteBlack []*rankmdl.BlackWhite) (err error) {
	var (
		sqls = make([]string, 0, len(whiteBlack))
		args = make([]interface{}, 0)
	)
	if len(whiteBlack) == 0 {
		return
	}
	for _, v := range whiteBlack {
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, id, v.Oid, v.Score, v.State, v.InterventionType, v.ObjectType)
	}
	_, err = tx.Exec(fmt.Sprintf(addOrUpdateBlackWhite, blackWhiteDBName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "addRankSource:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// AddRankSourceEliminateOld ...
func (d *dao) AddRankSourceEliminateOld(ctx context.Context, tx *xsql.Tx, baseID int64, sources []*rankmdl.Source) (err error) {
	err = d.updateSourceState(ctx, tx, baseID, rankmdl.SourceStateOffline)
	if err != nil {
		log.Errorc(ctx, "d.updateSourceState err(%v)", err)
		return
	}
	err = d.addRankSource(ctx, tx, baseID, sources)
	if err != nil {
		log.Errorc(ctx, "d.addRankSource err(%v)", err)
		return
	}
	return
}

func (d *dao) addRankSource(c context.Context, tx *xsql.Tx, id int64, sources []*rankmdl.Source) (err error) {
	var (
		sqls = make([]string, 0, len(sources))
		args = make([]interface{}, 0)
	)
	if len(sources) == 0 {
		return
	}
	for _, v := range sources {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, id, v.SourceID, v.SourceType, v.State)
	}
	_, err = tx.Exec(fmt.Sprintf(addRankSource, sourceDBName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "addRankSource:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

func (d *dao) updateSourceState(ctx context.Context, tx *xsql.Tx, baseID int64, state int) (err error) {
	if _, err = d.db.Exec(ctx, fmt.Sprintf(updateSourceState, sourceDBName), state, baseID); err != nil {
		log.Errorc(ctx, "rankv2@updateSourceState d.db.Exec() update failed. error(%v)", err)
		return
	}
	return
}

func (d *dao) createResult(c context.Context, tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(createResult, id)); err != nil {
		log.Errorc(c, "rankv2@createResult CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *dao) createAidSnapshot(c context.Context, tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(createAidSnapshot, id)); err != nil {
		log.Errorc(c, "rankv2@createAidSnapshot CREATE TABLE failed. error(%v)", err)
	}
	return
}

// UpdateRuleState new rank
func (d *dao) UpdateRuleState(c context.Context, ruleIDs []int64, state int) (err error) {

	if _, err = d.db.Exec(c, fmt.Sprintf(updateRuleState, ruleDBName, xstr.JoinInts(ruleIDs)), state); err != nil {
		log.Errorc(c, "rankv2@UpdateRuleState d.db.Exec() update failed. error(%v)", err)
		return
	}

	return
}

// UpdateRuleBatch new rank
func (d *dao) UpdateRuleBatch(c context.Context, tx *xsql.Tx, ruleID int64, showBatch int, showBatchTime int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(updateRuleBatch, ruleDBName), showBatch, showBatchTime, ruleID); err != nil {
		log.Errorc(c, "rankv2@UpdateRuleBatch d.db.Exec() update failed. error(%v)", err)
		return
	}

	return
}

// Update new rank
func (d *dao) Update(c context.Context, tx *xsql.Tx, rank *rankmdl.Base, sources []*rankmdl.Source, blackWhite []*rankmdl.BlackWhite) (err error) {

	if _, err = tx.Exec(fmt.Sprintf(updateRank, baseDBName), rank.Name, rank.IsShowScore, rank.RankType, rank.IsType, rank.Tids, rank.ArchiveStime, rank.ArchiveEtime, rank.Authority, rank.State, rank.ID); err != nil {
		log.Errorc(c, "rankv2@Update d.db.Exec() update failed. error(%v)", err)
		return
	}
	if err = d.addRankSource(c, tx, rank.ID, sources); err != nil {
		log.Errorc(c, "rankv2@Update d.db.Exec() d.addRankSource failed. error(%v)", err)
		return
	}
	if err = d.addBlackWhite(c, tx, rank.ID, blackWhite); err != nil {
		log.Errorc(c, "rank@Add d.addBlackWhite(%d) failed. error(%v)", rank.ID, err)
		return
	}
	return
}

// AddRankRule 增加子榜
func (d *dao) AddRankRule(c context.Context, tx *xsql.Tx, rule *rankmdl.Rule, score []*rankmdl.ScoreConfig) (err error) {
	var (
		result sql.Result
		id     int64
	)
	if result, err = tx.Exec(fmt.Sprintf(addRule, ruleDBName), rule.BaseID, rule.Name, rule.StatisticsType, rule.Nums, rule.UpdateFrequency, rule.UpdateScope, rule.State, rule.Precision, rule.Unit, rule.Description, rule.Stime, rule.Etime); err != nil {
		log.Errorc(c, "rankv2@AddRankRule d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Errorc(c, "rankv2@Add result.LastInsertId() failed. error(%v)", err)
		return
	}
	if err = d.addRankScoreConfig(c, tx, id, score); err != nil {
		log.Errorc(c, "rankv2@addRankScoreConfig d.db.Exec() INSERT failed. error(%v)", err)
		return
	}

	return
}

func (d *dao) addRankScoreConfig(c context.Context, tx *xsql.Tx, rankID int64, scoreConfig []*rankmdl.ScoreConfig) (err error) {
	var (
		sqls = make([]string, 0, len(scoreConfig))
		args = make([]interface{}, 0)
	)
	if len(scoreConfig) == 0 {
		return
	}
	for _, v := range scoreConfig {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, rankID, v.Action, v.Base, v.CntPerDay)
	}
	_, err = tx.Exec(fmt.Sprintf(addOrUpdateScoreConfig, scoreConfigDBName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "addRankScoreConfig:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// getRankScoreConfig ...
func (d *dao) getRankScoreConfig(c context.Context, ruleIDs []int64) (scoreConfig map[int64][]*rankmdl.ScoreConfig, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getScoreBatch, scoreConfigDBName, xstr.JoinInts(ruleIDs)))
	if err != nil {
		err = errors.Wrapf(err, "getRankScoreConfig:d.db.Query(%+v)", ruleIDs)
		return
	}
	defer rows.Close()
	list := make([]*rankmdl.ScoreConfig, 0)
	scoreConfig = make(map[int64][]*rankmdl.ScoreConfig, 0)
	for rows.Next() {
		n := new(rankmdl.ScoreConfig)
		if err = rows.Scan(&n.RankID, &n.Action, &n.Base, &n.CntPerDay); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "getRankScoreConfig:row.Scan row (%+v)", ruleIDs)
			return
		}
		list = append(list, n)
	}

	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "getRankScoreConfig:rowsErr(%+v)", ruleIDs)
	}
	for _, v := range list {
		if _, ok := scoreConfig[v.RankID]; ok {
			scoreConfig[v.RankID] = append(scoreConfig[v.RankID], v)
			continue
		}
		scoreConfig[v.RankID] = make([]*rankmdl.ScoreConfig, 0)
		scoreConfig[v.RankID] = append(scoreConfig[v.RankID], v)
	}
	return
}

// UpdateRankRule 更新子榜
func (d *dao) UpdateRankRule(c context.Context, tx *xsql.Tx, rule *rankmdl.Rule, score []*rankmdl.ScoreConfig) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(updateRule, ruleDBName), rule.Name, rule.Nums, rule.Stime, rule.Etime, rule.State, rule.ID); err != nil {
		log.Errorc(c, "rankv2@UpdateRankRule d.db.Exec() Update failed. error(%v)", err)
		return
	}
	if err = d.addRankScoreConfig(c, tx, rule.ID, score); err != nil {
		log.Errorc(c, "rankv2@addRankScoreConfig d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	return
}

// UpdateRankRule 更新子榜
func (d *dao) UpdateRankRuleShow(c context.Context, ruleID int64, unit int, precision int, description string) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(UpdateRankRuleShow, ruleDBName), precision, unit, description, ruleID); err != nil {
		log.Errorc(c, "rankv2@UpdateRankRuleShow d.db.Exec() Update failed. error(%v)", err)
		return
	}

	return
}

// UpdateRankRule 更新子榜
func (d *dao) UpdateRankRuleState(c context.Context, rule *rankmdl.Rule) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(updateOneRuleState, ruleDBName), rule.State, rule.ID); err != nil {
		log.Errorc(c, "rankv2@UpdateRankRuleState d.db.Exec() Update failed. error(%v)", err)
		return
	}
	return
}

// GetRankArchive get all rank by id
func (d *dao) GetRankArchive(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOidArchive, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRankArchive, fmt.Sprintf(resultArchiveDBName, baseID)), ruleID, batch)
	if err != nil {
		err = errors.Wrapf(err, "getRankArchive:d.db.Query(%d,%d)", baseID, ruleID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.ResultOidArchive)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.OID, &n.AID, &n.Rank, &n.Score, &n.Batch, &n.State); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "getRankArchive:d.db.Query(%d,%d)", baseID, ruleID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "getRankArchive:d.db.Query(%d,%d)", baseID, ruleID)
	}
	return
}

// GetRankOid get all rank by id
func (d *dao) GetRankOid(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOid, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(getRankOid, fmt.Sprintf(resultDBName, baseID)), ruleID, batch)
	if err != nil {
		err = errors.Wrapf(err, "getRankOid:d.db.Query(%d,%d)", baseID, ruleID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.ResultOid)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.OID, &n.Rank, &n.Score, &n.Batch, &n.State); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "getRankOid:d.db.Query(%d,%d)", baseID, ruleID)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "getRankOid:d.db.Query(%d,%d)", baseID, ruleID)
	}
	return
}

// UpdateRankAdjust 更新子榜
func (d *dao) UpdateRankAdjust(c context.Context, baseID int64, adjust *rankmdl.Adjust) (err error) {
	var (
		sqls = make([]string, 0, 1)
		args = make([]interface{}, 0)
	)

	sqls = append(sqls, "(?,?,?,?,?,?,?,?)")
	args = append(args, baseID, adjust.RankID, adjust.ParentID, adjust.OID, adjust.ObjectType, adjust.Rank, adjust.IsShow, adjust.State)
	_, err = d.db.Exec(c, fmt.Sprintf(addOrUpdateAdjust, adjustDBName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "UpdateRankAdjust:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}
