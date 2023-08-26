package record

import (
	"context"
	"database/sql"
	"fmt"

	"go-gateway/app/app-svr/steins-gate/service/api"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/pkg/errors"
)

const (
	_insertRecSQL            = "INSERT INTO %s(graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,is_preview,current_edge,current_cursor,cursor_choice) VALUE(?,?,?,?,?,?,?,0,?,?,?) ON DUPLICATE KEY UPDATE aid=values(aid),choices=values(choices),current_node=values(current_node),hidden_vars=values(hidden_vars),global_vars=values(global_vars),is_preview=0,current_edge=values(current_edge),current_cursor=values(current_cursor),cursor_choice=values(cursor_choice)"
	_insertRecPreviewSQL     = "INSERT INTO %s(graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,is_preview,current_edge,current_cursor,cursor_choice) VALUE(?,?,?,?,?,?,?,1,?,?,?) ON DUPLICATE KEY UPDATE aid=values(aid),choices=values(choices),current_node=values(current_node),hidden_vars=values(hidden_vars),global_vars=values(global_vars),is_preview=1,current_edge=values(current_edge),current_cursor=values(current_cursor),cursor_choice=values(cursor_choice)"
	_recordByGraphSQL        = "SELECT id,graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,current_edge,current_cursor,cursor_choice FROM %s WHERE mid=? AND graph_id=? AND is_preview=0"
	_recordByGraphPreviewSQL = "SELECT id,graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,current_edge,current_cursor,cursor_choice FROM %s WHERE mid=? AND graph_id=? AND is_preview=1"
	_recordByAIDSQL          = "SELECT id,graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,current_edge,current_cursor,cursor_choice FROM %s WHERE mid=? AND aid=? AND is_preview=0 ORDER BY graph_id DESC LIMIT 1"
	_recordsByGraphsSQL      = "SELECT id,graph_id,aid,mid,choices,current_node,hidden_vars,global_vars,current_edge,current_cursor,cursor_choice FROM %s WHERE mid=? AND graph_id IN (%s) AND is_preview=0"
	_recordByAIDsSQL         = "SELECT DISTINCT(aid) FROM %s WHERE mid=? AND aid IN (%s) AND is_preview=0"
)

func tableName(mid int64) string {
	return fmt.Sprintf("game_records_%02d", mid%100)
}

// RawRecord is
func (d *Dao) RawRecord(c context.Context, mid, graphID int64, preview bool) (res *api.GameRecords, err error) {
	var sqlStr string
	if preview {
		sqlStr = _recordByGraphPreviewSQL
	} else {
		sqlStr = _recordByGraphSQL
	}
	res = &api.GameRecords{}
	if err = d.db.QueryRow(c, fmt.Sprintf(sqlStr, tableName(mid)), mid, graphID).Scan(&res.Id, &res.GraphId, &res.Aid, &res.Mid, &res.Choices,
		&res.CurrentNode, &res.HiddenVars, &res.GlobalVars, &res.CurrentEdge, &res.CurrentCursor, &res.CursorChoice); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("recordByGraph mid %d, graphID %d empty record", mid, graphID)
		} else {
			err = errors.Wrapf(err, "record by gid %d mid %d", graphID, mid)
		}
		res = nil
		return
	}
	return
}

// RecordByAid checks an user's record by the Aid
func (d *Dao) RecordByAid(c context.Context, mid, aid int64) (res *api.GameRecords, err error) {
	res = &api.GameRecords{}
	if err = d.db.QueryRow(c, fmt.Sprintf(_recordByAIDSQL, tableName(mid)), mid, aid).Scan(&res.Id, &res.GraphId, &res.Aid, &res.Mid, &res.Choices, &res.CurrentNode,
		&res.HiddenVars, &res.GlobalVars, &res.CurrentEdge, &res.CurrentCursor, &res.CursorChoice); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("RecordByAid mid %d, aid %d empty record", mid, aid)
		} else {
			err = errors.Wrapf(err, "record by mid %d aid %d", mid, aid)
		}
		res = nil
		return
	}
	return
}

// AddRecord adds a new record
func (d *Dao) AddRecord(c context.Context, rec *api.GameRecords, preview bool) (err error) {
	var sqlStr string
	if preview {
		sqlStr = _insertRecPreviewSQL
	} else {
		sqlStr = _insertRecSQL
	}
	_, err = d.db.Exec(c, fmt.Sprintf(sqlStr, tableName(rec.Mid)), rec.GraphId, rec.Aid, rec.Mid, rec.Choices, rec.CurrentNode,
		rec.HiddenVars, rec.GlobalVars, rec.CurrentEdge, rec.CurrentCursor, rec.CursorChoice)
	if err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s) error(%v)", sqlStr, err)
		return
	}
	return
}

// RawRecords only for not preview， 以graphID为key, missAIDs是用graphID找不到的稿件ID的集合
func (d *Dao) RawRecords(c context.Context, mid int64, graphIDs []int64, buvid string) (records map[int64]*api.GameRecords, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_recordsByGraphsSQL, tableName(mid), xstr.JoinInts(graphIDs)), mid); err != nil {
		err = errors.Wrapf(err, "mid %d gids %v", mid, graphIDs)
		return
	}
	defer rows.Close()
	records = make(map[int64]*api.GameRecords)
	for rows.Next() {
		res := &api.GameRecords{
			Buvid: buvid,
		}
		if err = rows.Scan(&res.Id, &res.GraphId, &res.Aid, &res.Mid, &res.Choices, &res.CurrentNode, &res.HiddenVars, &res.GlobalVars,
			&res.CurrentEdge, &res.CurrentCursor, &res.CursorChoice); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		records[res.GraphId] = res
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// RecordByAids 以aid为key
func (d *Dao) RecordByAids(c context.Context, mid int64, aids []int64) (records map[int64]struct{}, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_recordByAIDsSQL, tableName(mid), xstr.JoinInts(aids)), mid); err != nil {
		err = errors.Wrapf(err, "mid %d aids %v", mid, aids)
		return
	}
	defer rows.Close()
	records = make(map[int64]struct{})
	for rows.Next() {
		var aid int64
		if err = rows.Scan(&aid); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		records[aid] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return

}
