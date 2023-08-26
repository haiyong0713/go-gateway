package rank

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"

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
	getRankConfigByID        = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime FROM %s where state=1 and id=?"
	getRankLog               = "SELECT id,rank_id,batch,rank_attribute,state,ctime,mtime FROM %s where state=1 and rank_id=? and batch=? and rank_attribute=?"
	getRankLogByTime         = "SELECT id,rank_id,batch,rank_attribute,state,ctime,mtime FROM %s where state=1 and rank_id=? and rank_attribute=? order by ctime desc limit 1"
	getRankLogByTimeAll      = "SELECT id,rank_id,batch,rank_attribute,state,ctime,mtime FROM %s where rank_id=? and rank_attribute=? order by ctime desc limit 1"
	getRankConfigOnline      = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime FROM %s where state=1 and etime>=? and stime<=?"
	getAllIntervention       = "SELECT id,oid,score,state,intervention_type,object_type,ctime,mtime FROM %s where state=1 and object_type = ? limit ?,?"
	getAllOidRankByLastBatch = "SELECT id,oid,rank,score,rank_attribute,state,batch,remark,ctime,mtime FROM %s where state=1 and batch = ? and rank_attribute = ? order by rank asc limit ?,?"
	insertOidRankSQL         = "INSERT INTO %s (oid,rank,score,state,batch,rank_attribute,remark) VALUES %s"
	insertSnapshotRankSQL    = "INSERT INTO %s (aid,mid,tid,views,danmaku,reply,fav,coin,likes,shares,videos,rank,rank_attribute,score,batch,state,remark,arc_ctime,pub_time) VALUES %s"
	insertConfigLastBatch    = "INSERT INTO %s (rank_id,batch,rank_attribute,state) VALUES %s "
	updateConfigLastBatch    = "UPDATE %s SET state = ? WHERE id = ? "
	getAllSnapShotByAids     = "SELECT distinct id,aid,mid,tid,views,danmaku,reply,fav,coin,likes,shares,videos,rank,rank_attribute,score,batch,state,remark,ctime,mtime,arc_ctime,pub_time FROM %s where batch=? and rank_attribute = ? and score>0 and rank>0 and aid IN (%s)"
	allSnapShotByAids        = "SELECT distinct id,aid,mid,tid,views,danmaku,reply,fav,coin,likes,shares,videos,rank,rank_attribute,score,batch,state,remark,ctime,mtime,arc_ctime,pub_time FROM %s where batch=? and rank_attribute = ? and aid IN (%s)"
)

// AllIntervention get all intervention by id
func (d *dao) AllIntervention(c context.Context, id int64, objectType, offset, limit int) (list []*rankmdl.Intervention, err error) {
	var (
		sqls []interface{}
	)
	tableName := fmt.Sprintf(interventionDBName, id)
	sqls = append(sqls, offset, limit)
	rows, err := d.db.Query(c, fmt.Sprintf(getAllIntervention, tableName), objectType, offset, limit)
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

// GetRankConfigOnline get rank config online
func (d *dao) GetRankConfigOnline(c context.Context, now time.Time) (list []*rankmdl.Rank, err error) {
	nowTime := now.Format("2006-01-02 15:04:05")
	rows, err := d.db.Query(c, fmt.Sprintf(getRankConfigOnline, rankDBName), nowTime, nowTime)
	if err != nil {
		err = errors.Wrapf(err, "GetRankConfigOnline:d.db.Query(%v)", now)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rank := new(rankmdl.Rank)
		if err = rows.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				list = nil
			} else {
				err = errors.Wrapf(err, "GetRankConfigOnline:row.Scan row (%v)", now)
			}
			return
		}
		if rank.StatisticsTime != "" {
			if err = json.Unmarshal([]byte(rank.StatisticsTime), &rank.StatisticsTimeStruct); err != nil {
				log.Errorc(c, "json.Unmarshal(%s) error(%v)", rank.StatisticsTime, err)
				return nil, err
			}
		}
		if rank.Ratio != "" {
			if err = json.Unmarshal([]byte(rank.Ratio), &rank.RatioStruct); err != nil {
				log.Errorc(c, "json.Unmarshal(%s) error(%v)", rank.Ratio, err)
				return nil, err
			}
		}
		list = append(list, rank)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GetRankConfigOnline:rowsErr(%v)", now)
	}
	return
}

// AllOidRank get all rank by id
func (d *dao) AllOidRank(c context.Context, id int64, lastBatch, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error) {
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

// BatchAddOidRank batch add rank
func (d *dao) BatchAddOidRank(c context.Context, tx *xsql.Tx, id int64, rank []*rankmdl.OidResult) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	tableName := fmt.Sprintf(oidDBName, id)
	for _, r := range rank {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?)")
		rows = append(rows, r.OID, r.Rank, r.Score, r.State, r.Batch, r.RankAttribute, r.Remark)
	}
	sql := fmt.Sprintf(insertOidRankSQL, tableName, strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddOidRank: tx.Exec")
	}
	return
}

// BatchAddSnapshotRank batch add rank
func (d *dao) BatchAddSnapshotRank(c context.Context, tx *xsql.Tx, id int64, rank []*rankmdl.Snapshot) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	tableName := fmt.Sprintf(snapshotDBName, id)
	for _, r := range rank {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		rows = append(rows, r.AID, r.MID, r.TID, r.View, r.Danmaku, r.Reply, r.Fav, r.Coin, r.Like, r.Share, r.Videos, r.Rank, r.RankAttribute, r.Score, r.Batch, r.State, r.Remark, r.ArcCtime, r.PubTime)
	}
	sql := fmt.Sprintf(insertSnapshotRankSQL, tableName, strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddSnapshotRank: tx.Exec")
	}
	return
}

// UpdateRankConfigLastBatch update rank config last batch
func (d *dao) InsertRankLog(c context.Context, id int64, lastBatch, lastAttribute, state int) (int64, error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)

	rowsTmp = append(rowsTmp, "(?,?,?,?)")
	rows = append(rows, id, lastBatch, lastAttribute, state)
	sql := fmt.Sprintf(insertConfigLastBatch, rankLogDBName, strings.Join(rowsTmp, ","))
	res, err := d.db.Exec(c, sql, rows...)
	if err != nil {
		err = errors.Wrap(err, "BatchAddSnapshotRank: tx.Exec")
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateRankConfigLastBatch update rank config last batch
func (d *dao) UpdateRankLog(c context.Context, tx *xsql.Tx, id int64, state int) (err error) {
	var (
		rows []interface{}
	)

	rows = append(rows, state, id)
	sql := fmt.Sprintf(updateConfigLastBatch, rankLogDBName)
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddSnapshotRank: tx.Exec")
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

	row := d.db.QueryRow(c, fmt.Sprintf(getRankConfigByID, rankDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankConfigByID d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	if rank.StatisticsTime != "" {
		if err = json.Unmarshal([]byte(rank.StatisticsTime), &rank.StatisticsTimeStruct); err != nil {
			log.Errorc(c, "json.Unmarshal(%s) error(%v)", rank.StatisticsTime, err)
			return nil, err
		}
	}
	if rank.Ratio != "" {
		if err = json.Unmarshal([]byte(rank.Ratio), &rank.RatioStruct); err != nil {
			log.Errorc(c, "json.Unmarshal(%s) error(%v)", rank.Ratio, err)
			return nil, err
		}
	}

	return
}

// GetRankLog ...
func (d *dao) GetRankLog(c context.Context, rankID int64, batch, attributeType int) (rank *rankmdl.Log, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Log{}
	arg = append(arg, rankID, batch, attributeType)

	row := d.db.QueryRow(c, fmt.Sprintf(getRankLog, rankLogDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.RankID, &rank.Batch, &rank.RankAttribute, &rank.State, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			rank = nil
		} else {
			log.Errorc(c, "Rank@GetRankLog d.db.QueryRow() SELECT failed. error(%v)", err)
		}
	}
	return
}

// GetRankLog ...
func (d *dao) GetRankLogOrderByTime(c context.Context, rankID int64, attributeType int) (rank *rankmdl.Log, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Log{}
	arg = append(arg, rankID, attributeType)

	row := d.db.QueryRow(c, fmt.Sprintf(getRankLogByTime, rankLogDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.RankID, &rank.Batch, &rank.RankAttribute, &rank.State, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			rank = nil
		} else {
			log.Errorc(c, "Rank@GetRankLogOrderByTime d.db.QueryRow() SELECT failed. error(%v)", err)
		}
	}
	return
}

// GetRankLog ...
func (d *dao) GetRankLogOrderByTimeAll(c context.Context, rankID int64, attributeType int) (rank *rankmdl.Log, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Log{}
	arg = append(arg, rankID, attributeType)
	row := d.db.QueryRow(c, fmt.Sprintf(getRankLogByTimeAll, rankLogDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.RankID, &rank.Batch, &rank.RankAttribute, &rank.State, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			rank = nil
		} else {
			log.Errorc(c, "Rank@getRankLogByTimeAll d.db.QueryRow() SELECT failed. error(%v)", err)
		}
	}
	return
}

// AllSnapshotByAids get all rank by id
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
		if err = rows.Scan(&n.ID, &n.AID, &n.MID, &n.TID, &n.View, &n.Danmaku, &n.Reply, &n.Fav, &n.Coin, &n.Like, &n.Share, &n.Videos, &n.Rank, &n.RankAttribute, &n.Score, &n.Batch, &n.State, &n.Remark, &n.Ctime, &n.Mtime, &n.ArcCtime, &n.PubTime); err != nil {
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

// SnapshotByAllAids get all rank by id
func (d *dao) SnapshotByAllAids(c context.Context, id int64, aids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error) {
	tableName := fmt.Sprintf(snapshotDBName, id)
	aidStrs := make([]string, 0)
	for _, v := range aids {
		aidStrs = append(aidStrs, strconv.FormatInt(v, 10))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(allSnapShotByAids, tableName, strings.Join(aidStrs, ",")), lastBatch, rankAttribute)
	if err != nil {
		err = errors.Wrapf(err, "AllSnapshot:d.db.Query(%v)", id)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.Snapshot)
		if err = rows.Scan(&n.ID, &n.AID, &n.MID, &n.TID, &n.View, &n.Danmaku, &n.Reply, &n.Fav, &n.Coin, &n.Like, &n.Share, &n.Videos, &n.Rank, &n.RankAttribute, &n.Score, &n.Batch, &n.State, &n.Remark, &n.Ctime, &n.Mtime, &n.ArcCtime, &n.PubTime); err != nil {
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
