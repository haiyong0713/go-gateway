package like

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/xstr"
	"strconv"

	"go-common/library/log"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_addReserveSQL     = "INSERT INTO `%s` (`sid`,`mid`,`state`,`num`,`ipv6`,`from`,`typ`,`oid`,`platform`,`mobiapp`,`buvid`,`spmid`, `order`)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_upReserveSQL      = "UPDATE `%s` SET `state` = ?,`num` = ?,`ipv6` = ?, `score` = ?, `from` = ?, `typ` = ?, `oid` = ?, `platform` = ?, `mobiapp` = ?, `buvid` = ?, `spmid` = ? WHERE `sid` = ? AND `mid` = ?"
	_cancelReserveSQL  = "UPDATE `%s` SET `state` = ?,`ipv6` = ? WHERE `sid` = ? AND `mid` = ?"
	_querySQL          = "SELECT `id`,`state`,`num`,`mtime`,`ctime`, `order` FROM `%s` WHERE `sid` = ? AND `mid` = ?"
	_queryGroupIdSQL   = "SELECT `id` FROM `act_subject_counter_group` WHERE `sid` = ? and `state` = 1"
	_queryGroupSQL     = "SELECT `id`, `sid`, `group_name`, `dim1`, `dim2`, `threshold`, `counter_info`, `author`, `ctime`, `mtime` FROM `act_subject_counter_group` WHERE `id` in (%s)"
	_queryGroupNodeSQL = "SELECT `id`, `sid`, `group_id`, `node_name`, `node_val`, `ctime`, `mtime` FROM `act_subject_counter_node` WHERE `group_id` in (%s)  and `state` = 1"
	_tableName         = "act_reserve_%02d"
	_tableNum          = 100
)

// tableName [00,99].
func (d *Dao) tableName(sid int64) string {
	if _, ok := d.c.Rule.SpecReserveSids[strconv.FormatInt(sid, 10)]; ok {
		return fmt.Sprintf("act_reserve_%d", sid)
	}
	return fmt.Sprintf(_tableName, sid%_tableNum)
}

// AddReserve .
func (d *Dao) AddReserve(c context.Context, item *lmdl.ActReserve) (id int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_addReserveSQL, d.tableName(item.Sid)), item.Sid, item.Mid, item.State, item.Num, item.IPv6, item.Report.From, item.Report.Typ, item.Report.Oid, item.Report.Platform, item.Report.Mobiapp, item.Report.Buvid, item.Report.Spmid, item.Order); err != nil {
		log.Error("TxAddReserve:tx.Exec(%s) error(%v)", _addReserveSQL, err)
		return
	}
	id, err = res.LastInsertId()
	return
}

// UpReserve .
func (d *Dao) UpReserve(c context.Context, item *lmdl.ActReserve) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_upReserveSQL, d.tableName(item.Sid)), item.State, item.Num, item.IPv6,
		item.Score, item.Report.From, item.Report.Typ, item.Report.Oid, item.Report.Platform, item.Report.Mobiapp, item.Report.Buvid, item.Report.Spmid, item.Sid, item.Mid); err != nil {
		log.Error("UpReserve:tx.Exec(%s) error(%v)", _upReserveSQL, err)
	}
	return
}

// CancelReserve .
func (d *Dao) CancelReserve(c context.Context, item *lmdl.ActReserve) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_cancelReserveSQL, d.tableName(item.Sid)), item.State, item.IPv6, item.Sid, item.Mid); err != nil {
		log.Errorc(c, "CancelReserve:tx.Exec(%s) error(%v)", _cancelReserveSQL, err)
	}
	return
}

// RawReserveOnly .
func (d *Dao) RawReserveOnly(c context.Context, sid, mid int64) (res *lmdl.HasReserve, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_querySQL, d.tableName(sid)), sid, mid)
	res = &lmdl.HasReserve{}
	if err = row.Scan(&res.ID, &res.State, &res.Num, &res.Mtime, &res.Ctime, &res.Order); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Errorc(c, "RawReserveOnly error(%v)", err)
		}
	}
	return
}

func (d *Dao) RawGetReserveCounterGroupIDBySid(c context.Context, sid int64) ([]int64, error) {
	rows, err := d.db.Query(c, _queryGroupIdSQL, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]int64, 0, 10)
	for rows.Next() {
		var tmp int64
		if err = rows.Scan(&tmp); err != nil {
			log.Errorc(c, "RawGetReserveCounterGroupIDBySid rows.Scan() failed. error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, rows.Err()
}

func (d *Dao) RawGetReserveCounterGroupInfoByGid(c context.Context, gid []int64) (map[int64]*lmdl.ReserveCounterGroupItem, error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_queryGroupSQL, xstr.JoinInts(gid)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*lmdl.ReserveCounterGroupItem)
	for rows.Next() {
		tmp := new(lmdl.ReserveCounterGroupItem)
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.GroupName, &tmp.Dim1, &tmp.Dim2, &tmp.Threshold, &tmp.CounterInfo, &tmp.Author, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "RawGetReserveCounterGroupInfoByGid rows.Scan() failed. error(%v)", err)
			return nil, err
		}
		res[tmp.ID] = tmp
	}
	return res, rows.Err()
}

func (d *Dao) RawGetReserveCounterNodeByGid(c context.Context, gid []int64) (map[int64][]*lmdl.ReserveCounterNodeItem, error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_queryGroupNodeSQL, xstr.JoinInts(gid)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64][]*lmdl.ReserveCounterNodeItem)
	for rows.Next() {
		tmp := new(lmdl.ReserveCounterNodeItem)
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.GroupID, &tmp.NodeName, &tmp.NodeVal, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "RawGetReserveCounterNodeByGid rows.Scan() failed. error(%v)", err)
			return nil, err
		}
		if p, ok := res[tmp.GroupID]; ok {
			p = append(p, tmp)
			res[tmp.GroupID] = p
		} else {
			res[tmp.GroupID] = []*lmdl.ReserveCounterNodeItem{tmp}
		}
	}
	return res, rows.Err()
}
