package rank

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/component"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"

	"github.com/pkg/errors"
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
	getRuleByID = `SELECT
	id,base_id,name,statistics_type,nums,update_frequency,update_scope,show_batch,show_batch_time,state,stime,etime,ctime,mtime
	FROM %s WHERE id=?`
	getRuleBatch = `SELECT
	rule_id,last_batch,last_batch_time
	FROM %s WHERE rule_id in(%s)
	`
	getRankArchive = `SELECT id,base_id,rank_id,oid,aid,rank,score,show_score,batch,state from %s 
	where rank_id=? and batch=? and state=1  order by rank asc
	`
	getRankOid = `SELECT id,base_id,rank_id,oid,rank,score,show_score,batch,state from %s 
	where rank_id=? and batch=? order by rank asc
	`
	getRank = `SELECT 
	id,name,rank_type,is_type,tids,archive_stime,archive_etime,author,authority,state,ctime,mtime 
	FROM %s WHERE id=?`
)

// GetSource get all rank by id
func (d *Dao) getRuleByID(c context.Context, id int64) (n *rankmdl.Rule, err error) {

	var (
		arg []interface{}
	)
	n = &rankmdl.Rule{}
	arg = append(arg, id)
	row := component.GlobalDB.QueryRow(c, fmt.Sprintf(getRuleByID, ruleDBName), arg...)
	if err = row.Scan(&n.ID, &n.BaseID, &n.Name, &n.StatisticsType, &n.Nums, &n.UpdateFrequency, &n.UpdateScope, &n.ShowBatch, &n.ShowBatchTime, &n.State, &n.Stime, &n.Etime, &n.Ctime, &n.Mtime); err != nil {
		log.Errorc(c, "Rank@getRuleByID d.db.QueryRow() SELECT failed. error(%v)", err)
		return
	}

	return
}

// GetRuleByID ...
func (d *Dao) GetRuleByID(c context.Context, id int64) (n *rankmdl.Rule, err error) {
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
	}
	return
}

// GetRankByID ...
func (d *Dao) GetRankByID(c context.Context, id int64) (rank *rankmdl.Base, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Base{}
	arg = append(arg, id)
	row := component.GlobalDB.QueryRow(c, fmt.Sprintf(getRank, baseDBName), arg...)
	if err = row.Scan(&rank.ID, &rank.Name, &rank.RankType, &rank.IsType, &rank.Tids, &rank.ArchiveStime, &rank.ArchiveEtime, &rank.Author, &rank.Authority, &rank.State, &rank.Ctime, &rank.Mtime); err != nil {
		log.Errorc(c, "Rank@GetRankByID d.db.QueryRow() SELECT failed. error(%v)", err)
		return
	}
	rank.TidsStruct, _ = xstr.SplitInts(rank.Tids)
	return
}

// GetSourceBatch get all rank by id
func (d *Dao) getRuleBatch(c context.Context, ruleIDs []int64) (list []*rankmdl.RuleBatchTime, err error) {
	rows, err := component.GlobalDB.Query(c, fmt.Sprintf(getRuleBatch, ruleBatchDBName, xstr.JoinInts(ruleIDs)))
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

// GetRankArchive get all rank by id
func (d *Dao) GetRankArchive(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOidArchive, err error) {
	rows, err := component.GlobalDB.Query(c, fmt.Sprintf(getRankArchive, fmt.Sprintf(resultArchiveDBName, baseID)), ruleID, batch)
	if err != nil {
		err = errors.Wrapf(err, "getRankArchive:d.db.Query(%d,%d)", baseID, ruleID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.ResultOidArchive)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.OID, &n.AID, &n.Rank, &n.Score, &n.ShowScore, &n.Batch, &n.State); err != nil {
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
func (d *Dao) GetRankOid(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOid, err error) {
	rows, err := component.GlobalDB.Query(c, fmt.Sprintf(getRankOid, fmt.Sprintf(resultDBName, baseID)), ruleID, batch)
	if err != nil {
		err = errors.Wrapf(err, "getRankOid:d.db.Query(%d,%d)", baseID, ruleID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(rankmdl.ResultOid)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.RankID, &n.OID, &n.Rank, &n.Score, &n.ShowScore, &n.Batch, &n.State); err != nil {
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
