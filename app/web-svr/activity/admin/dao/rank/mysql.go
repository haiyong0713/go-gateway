package rank

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	xsql "go-common/library/database/sql"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank"

	"go-common/library/log"
	xtime "go-common/library/time"

	"github.com/pkg/errors"
)

const (
	interventionDBName = "act_rank_intervention_%d"
	oidDBName          = "act_rank_oid_%d"
	snapshotDBName     = "act_aid_snapshot_%d"
	rankDBName         = "act_rank_config"
	rankLogDBName      = "act_rank_log"
)

const (
	addNewRank               = "INSERT INTO act_rank_config(sid,sid_source,stime,etime) VALUES(?,?,?,?)"
	createResult             = "CREATE TABLE IF NOT EXISTS act_rank_oid_%d LIKE act_rank_oid"
	createAidSnapshot        = "CREATE TABLE IF NOT EXISTS act_aid_snapshot_%d LIKE act_aid_snapshot"
	createIntervention       = "CREATE TABLE IF NOT EXISTS act_rank_intervention_%d LIKE act_rank_intervention"
	updateConfigSQL          = "UPDATE act_rank_config SET  ratio = ?,rank_type=?,rank_attribute = ?,rank_top=?,is_auto=?,is_show_score=?,state=?,statistics_time=?,stime=?,etime=? where `id` = ?"
	getRankConfigBySIDSQL    = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime from act_rank_config where sid=? and sid_source=? and state = ?"
	getRankConfigByIDSQL     = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime from act_rank_config where id=? and state = ?"
	getRankConfigByIDAllSQL  = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime from act_rank_config where id=?"
	getAllIntervention       = "SELECT id,oid,score,state,intervention_type,object_type,ctime,mtime FROM %s where object_type = ? and intervention_type=? and state = 1 limit ?,?"
	oidResultUpdate          = "INSERT INTO %s (id,rank,score,state) VALUES %s ON DUPLICATE KEY UPDATE id=VALUES(id),score=VALUES(score),rank=VALUES(rank),state=VALUES(state)"
	snapshotUpdate           = "INSERT INTO %s (id,rank,score,state) VALUES %s ON DUPLICATE KEY UPDATE id=VALUES(id),score=VALUES(score),rank=VALUES(rank),state=VALUES(state)"
	interventionUpdate       = "INSERT INTO %s (oid,score,state,intervention_type,object_type) VALUES %s ON DUPLICATE KEY UPDATE oid=VALUES(oid),score=VALUES(score),intervention_type=VALUES(intervention_type),object_type=VALUES(object_type),state=VALUES(state)"
	interventionTotal        = "SELECT count(id) FROM %v WHERE object_type = ? and intervention_type=? and state = 1"
	allOidRankTotal          = "SELECT count(distinct oid) FROM %v WHERE batch = ? and rank_attribute=? and rank !=0 and score>0"
	getOidRankByLastBatch    = "SELECT distinct id,oid,rank,score,rank_attribute,state,batch,remark,ctime,mtime FROM %s where batch = ? and rank_attribute = ? and rank !=0 and score>0 group by oid order by rank asc limit ?,?"
	getAllOidRankByLastBatch = "SELECT distinct id,oid,rank,score,rank_attribute,state,batch,remark,ctime,mtime FROM %s where batch = ? and rank_attribute = ? and score>0 order by rank asc limit ?,?"
	getAllSnapShotByMids     = "SELECT distinct id,aid,mid,tid,views,danmaku,reply,fav,coin,likes,shares,videos,rank,rank_attribute,score,batch,state,remark,ctime,mtime FROM %s where batch=? and rank_attribute = ? and score>0 and rank>0 and state = 1 and mid IN (%s)"
	getAllSnapShotByAids     = "SELECT distinct id,aid,mid,tid,views,danmaku,reply,fav,coin,likes,shares,videos,rank,rank_attribute,score,batch,state,remark,ctime,mtime FROM %s where batch=? and rank_attribute = ? and score>0 and rank>0 and aid IN (%s)"
	getNewBatch              = "SELECT id,rank_id,batch,state,rank_attribute,ctime,mtime FROM %s where state=1 and rank_id=? and rank_attribute =? order by ctime desc limit 0,1"
)

// Create new rank
func (d *dao) Create(c context.Context, tx *xsql.Tx, sid int64, sidSource int, stime, etime xtime.Time) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = tx.Exec(addNewRank, sid, sidSource, stime, etime); err != nil {
		log.Error("rank@Add d.db.Exec() INSERT failed. error(%v)", err)
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Error("rank@Add result.LastInsertId() failed. error(%v)", err)
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
	if err = d.createIntervention(c, tx, id); err != nil {
		log.Errorc(c, "rank@Add d.createAddress(%d) failed. error(%v)", id, err)
		return
	}
	return
}

func (d *dao) createResult(c context.Context, tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(createResult, id)); err != nil {
		log.Errorc(c, "rank@createResult CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *dao) createAidSnapshot(c context.Context, tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(createAidSnapshot, id)); err != nil {
		log.Errorc(c, "rank@createAidSnapshot CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *dao) createIntervention(c context.Context, tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(createIntervention, id)); err != nil {
		log.Errorc(c, "rank@createIntervention CREATE TABLE failed. error(%v)", err)
	}
	return
}

// GetRankConfigBySid
func (d *dao) GetRankConfigBySid(c context.Context, sid int64, sidSource int) (rank *rankmdl.Rank, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Rank{}
	arg = append(arg, sid)
	arg = append(arg, sidSource)
	arg = append(arg, rankmdl.RankStateOnline)
	row := d.db.QueryRow(c, getRankConfigBySIDSQL, arg...)
	if err = row.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(c, "Rank@GetRankConfigBySid d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// GetRankConfigByID
func (d *dao) GetRankConfigByID(c context.Context, id int64) (rank *rankmdl.Rank, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Rank{}
	arg = append(arg, id)
	arg = append(arg, rankmdl.RankStateOnline)
	row := d.db.QueryRow(c, getRankConfigByIDSQL, arg...)
	if err = row.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankConfigByID d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// GetRankConfigByID ...
func (d *dao) GetRankConfigByIDAll(c context.Context, id int64) (rank *rankmdl.Rank, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Rank{}
	arg = append(arg, id)
	row := d.db.QueryRow(c, getRankConfigByIDAllSQL, arg...)
	if err = row.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankConfigByID d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// UpdateRankConfig update rank config information
func (d *dao) UpdateRankConfig(c context.Context, id int64, rank *rankmdl.Rank) (err error) {
	if rank != nil {
		if _, err = d.db.Exec(c, updateConfigSQL, rank.Ratio, rank.RankType, rank.RankAttribute, rank.Top, rank.IsAuto, rank.IsShowScore, rank.State, rank.StatisticsTime, rank.Stime, rank.Etime, id); err != nil {
			log.Errorc(c, "Rank@UpdateRankConfig() UPDATE act_rank_config failed. error(%v)", err)
		}
	}
	return
}

// AllIntervention get all intervention by id
func (d *dao) AllIntervention(c context.Context, id int64, objectType, interventionType, offset, limit int) (list []*rankmdl.Intervention, err error) {
	var (
		sqls []interface{}
	)
	tableName := fmt.Sprintf(interventionDBName, id)
	sqls = append(sqls, offset, limit)
	rows, err := d.db.Query(c, fmt.Sprintf(getAllIntervention, tableName), objectType, interventionType, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "AllIntervention:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Intervention)
		if err = rows.Scan(&n.ID, &n.OID, &n.Score, &n.State, &n.InterventionType, &n.ObjectType, &n.Ctime, &n.Mtime); err != nil {
			err = errors.Wrapf(err, "AllIntervention:row.Scan row (%v)", id)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "AllIntervention:rowsErr(%v)", id)
	}
	return
}

// AllInterventionTotal get gift total
func (d *dao) AllInterventionTotal(c context.Context, id int64, objectType, interventionType int) (total int, err error) {
	var (
		arg []interface{}
	)
	tableName := fmt.Sprintf(interventionDBName, id)
	arg = append(arg, objectType, interventionType)
	row := d.db.QueryRow(c, fmt.Sprintf(interventionTotal, tableName), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "rank@GiftDraftTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// BacthInsertOrUpdateBlackOrWhiteTx batch insert or update  BlackOrWhite
func (d *dao) BacthInsertOrUpdateBlackOrWhiteTx(c context.Context, tx *xsql.Tx, id int64, intervention []*rankmdl.Intervention) (err error) {
	var (
		sqls = make([]string, 0, len(intervention))
		args = make([]interface{}, 0)
	)
	if len(intervention) == 0 {
		return
	}
	for _, v := range intervention {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, v.OID, v.Score, v.State, v.InterventionType, v.ObjectType)
	}
	tableName := fmt.Sprintf(interventionDBName, id)
	_, err = tx.Exec(fmt.Sprintf(interventionUpdate, tableName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateBlackOrWhite:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BacthInsertOrUpdateBlackOrWhiteTx batch insert or update  BlackOrWhite
func (d *dao) BacthInsertOrUpdateOidResultTx(c context.Context, tx *xsql.Tx, id int64, oidResult []*rankmdl.OidResult) (err error) {
	var (
		sqls = make([]string, 0, len(oidResult))
		args = make([]interface{}, 0)
	)
	if len(oidResult) == 0 {
		return
	}
	tableName := fmt.Sprintf(oidDBName, id)
	for _, v := range oidResult {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, v.ID, v.Rank, v.Score, v.State)
	}
	_, err = tx.Exec(fmt.Sprintf(oidResultUpdate, tableName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateOidResultTx:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BacthInsertOrUpdateSnapshotTx batch insert or update  BlackOrWhite
func (d *dao) BacthInsertOrUpdateSnapshotTx(c context.Context, tx *xsql.Tx, id int64, snapShot []*rankmdl.Snapshot) (err error) {
	var (
		sqls = make([]string, 0, len(snapShot))
		args = make([]interface{}, 0)
	)
	if len(snapShot) == 0 {
		return
	}
	tableName := fmt.Sprintf(snapshotDBName, id)
	for _, v := range snapShot {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, v.ID, v.Rank, v.Score, v.State)
	}
	_, err = tx.Exec(fmt.Sprintf(snapshotUpdate, tableName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateSnapshotTx:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BacthInsertOrUpdateBlackOrWhite batch insert or update  BlackOrWhite
func (d *dao) BacthInsertOrUpdateBlackOrWhite(c context.Context, id int64, intervention []*rankmdl.Intervention) (err error) {
	var (
		sqls = make([]string, 0, len(intervention))
		args = make([]interface{}, 0)
	)
	if len(intervention) == 0 {
		return
	}
	for _, v := range intervention {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, v.OID, v.Score, v.State, v.InterventionType, v.ObjectType)
	}
	tableName := fmt.Sprintf(interventionDBName, id)
	_, err = d.db.Exec(c, fmt.Sprintf(interventionUpdate, tableName, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateBlackOrWhite:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// OidRankInRankTotal get gift total
func (d *dao) OidRankInRankTotal(c context.Context, id int64, lastBatch int64, rankAttribute int) (total int, err error) {
	var (
		arg []interface{}
	)
	tableName := fmt.Sprintf(oidDBName, id)
	arg = append(arg, lastBatch, rankAttribute)
	row := d.db.QueryRow(c, fmt.Sprintf(allOidRankTotal, tableName), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "rank@AllOidRankTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// OidRankInRank get all rank by id
func (d *dao) OidRankInRank(c context.Context, id int64, lastBatch int64, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error) {
	tableName := fmt.Sprintf(oidDBName, id)
	rows, err := d.db.Query(c, fmt.Sprintf(getOidRankByLastBatch, tableName), lastBatch, rankAttribute, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "OidRankInRank:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.OidResult)
		if err = rows.Scan(&n.ID, &n.OID, &n.Rank, &n.Score, &n.RankAttribute, &n.State, &n.Batch, &n.Remark, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "OidRankInRank:row.Scan row (%v)", id)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OidRankInRank:rowsErr(%v)", id)
	}
	return
}

// AllOidRank get all rank by id
func (d *dao) AllOidRank(c context.Context, id int64, lastBatch int64, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error) {
	tableName := fmt.Sprintf(oidDBName, id)
	rows, err := d.db.Query(c, fmt.Sprintf(getAllOidRankByLastBatch, tableName), lastBatch, rankAttribute, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "AllOidRank:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.OidResult)
		if err = rows.Scan(&n.ID, &n.OID, &n.Rank, &n.Score, &n.RankAttribute, &n.State, &n.Batch, &n.Remark, &n.Ctime, &n.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				return
			}
			err = errors.Wrapf(err, "AllOidRank:row.Scan row (%v)", id)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "AllOidRank:rowsErr(%v)", id)
	}
	return
}

// AllSnapshotByMids get all rank by id
func (d *dao) AllSnapshotByMids(c context.Context, id int64, mids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error) {
	tableName := fmt.Sprintf(snapshotDBName, id)
	midStrs := make([]string, 0)
	for _, v := range mids {
		midStrs = append(midStrs, strconv.FormatInt(v, 10))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(getAllSnapShotByMids, tableName, strings.Join(midStrs, ",")), lastBatch, rankAttribute)
	if err != nil {
		err = errors.Wrapf(err, "AllSnapshot:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Snapshot)
		if err = rows.Scan(&n.ID, &n.AID, &n.MID, &n.TID, &n.View, &n.Danmaku, &n.Reply, &n.Fav, &n.Coin, &n.Like, &n.Share, &n.Videos, &n.Rank, &n.RankAttribute, &n.Score, &n.Batch, &n.State, &n.Remark, &n.Ctime, &n.Mtime); err != nil {
			err = errors.Wrapf(err, "AllSnapshot:row.Scan row (%v)", id)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "AllSnapshot:rowsErr(%v)", id)
	}
	return
}

// AllSnapshotByMids get all rank by id
func (d *dao) AllSnapshotByAids(c context.Context, id int64, aids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error) {
	tableName := fmt.Sprintf(snapshotDBName, id)
	aidStrs := make([]string, 0)
	for _, v := range aids {
		aidStrs = append(aidStrs, strconv.FormatInt(v, 10))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(getAllSnapShotByAids, tableName, strings.Join(aidStrs, ",")), lastBatch, rankAttribute)
	if err != nil {
		err = errors.Wrapf(err, "AllSnapshot:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Snapshot)
		if err = rows.Scan(&n.ID, &n.AID, &n.MID, &n.TID, &n.View, &n.Danmaku, &n.Reply, &n.Fav, &n.Coin, &n.Like, &n.Share, &n.Videos, &n.Rank, &n.RankAttribute, &n.Score, &n.Batch, &n.State, &n.Remark, &n.Ctime, &n.Mtime); err != nil {
			err = errors.Wrapf(err, "AllSnapshot:row.Scan row (%v)", id)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "AllSnapshot:rowsErr(%v)", id)
	}
	return
}

// GetLastBatch
func (d *dao) GetLastBatch(c context.Context, id int64, rankAttribute int) (rank *rankmdl.Log, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Log{}
	arg = append(arg, id)
	arg = append(arg, rankAttribute)
	row := d.db.QueryRow(c, fmt.Sprintf(getNewBatch, rankLogDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.RankID, &rank.Batch, &rank.State, &rank.RankAttribute, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(c, "Rank@GetRankConfigBySid d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}
