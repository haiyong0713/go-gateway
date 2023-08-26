package hidden_vars

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	xecode "go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_edgesWithAttrSQL = "SELECT id,from_node,to_node,attribute FROM graph_edge WHERE attribute <>'' AND graph_id=?"
	_insertRecSQL     = "INSERT INTO %s(mid,graph_id,current_id,cursor_id,value) VALUE(?,?,?,?,?)"
	_updateRecSQL     = "UPDATE %s SET value=? WHERE id=?"
	_recordExistSQL   = "SELECT id FROM %s WHERE mid=? AND graph_id=? AND current_id=? AND cursor_id=?"
	_rawRecordSQL     = "SELECT value FROM %s WHERE mid=? AND graph_id=? AND current_id=? AND cursor_id=?"
)

func tableName(mid int64) string {
	return fmt.Sprintf("hidden_vars_rec_%02d", mid%100)
}

// 隐藏变量回源逻辑
func (d *Dao) hvarsInfo(c context.Context, mid int64, buvid string, graphInfo *api.GraphInfo, hvarReq *model.HvarReq) (a *model.HiddenVarsRecord, err error) {
	var (
		passedIDs, cursorList, choiceList []int64
		//nolint:ineffassign
		variables = make(map[string]*model.RegionalVal)
		edas      *model.EdgeAttrsCache
		//nolint:ineffassign
		recs      = make(map[int64]*model.HiddenVarsRecord)
		cursorMap = make(map[int64]int64)
		isMatch   bool
	)
	a = &model.HiddenVarsRecord{
		Vars: make(map[string]*model.HiddenVar),
	}
	if variables, err = model.GetVarsMap(graphInfo); err != nil { // service层应有判断，如果当前图并无隐藏变量，无需回源，这里兜底
		log.Error("GraphID %d GetVarsMap Err %v", graphInfo.Id, err)
		return
	}
	for k, v := range variables {
		if v.Type != model.RegionalVarTypeRandom { // 随机变量无存档
			initialV := new(model.HiddenVar)
			initialV.FromRegionalVar(v)
			a.Vars[k] = initialV
		}
	}
	if cursorList, err = xstr.SplitInts(hvarReq.CursorChoices); err != nil {
		log.Error("Mid %d GraphID %d, Record %s, FromNode %d is Illegal!!! Buvid %s", mid, graphInfo.Id, hvarReq.Choices, hvarReq.CurrentID, buvid)
		return
	}
	if choiceList, err = xstr.SplitInts(hvarReq.Choices); err != nil {
		log.Error("Mid %d GraphID %d, Record %s, FromNode %d is Illegal!!! Buvid %s", mid, graphInfo.Id, hvarReq.Choices, hvarReq.CurrentID, buvid)
		return
	}
	if len(cursorList) != len(choiceList) {
		log.Error("Mid %d GraphID %d, Record %s, FromNode %d is Illegal!!! Buvid %s", mid, graphInfo.Id, hvarReq.Choices, hvarReq.CurrentID, buvid)
		err = ecode.RequestErr
		return
	}
	for i := 0; i < len(cursorList); i++ {
		if cursorList[i] <= hvarReq.CurrentCursorID {
			passedIDs = append(passedIDs, choiceList[i])
		}
		if cursorList[i] == hvarReq.CurrentCursorID { // 验证当前游标在不在列表
			isMatch = true
		}
		cursorMap[choiceList[i]] = cursorList[i]
	}
	if !isMatch {
		log.Error("Mid %d GraphID %d, Record %s, FromNode %d is Illegal!!! Buvid %s", mid, graphInfo.Id, hvarReq.Choices, hvarReq.CurrentID, buvid)
		err = xecode.GraphLoopRecordErr
		return
	}
	if edas, err = d.edgeAttrsByGraph(c, graphInfo.Id); err != nil {
		log.Error("Mid %d GraphID %d passedIDs %v, Err %v", mid, graphInfo.Id, passedIDs, err)
		return
	}
	rec := new(model.HiddenVarsRecord) // 根节点存档
	rec.DeepCopy(a)
	if model.IsEdgeGraph(graphInfo) || model.IsInterventionGraph(graphInfo) { // edge图直接通过edgeID获取attribute
		recs = edas.GenerateEdgeRecs(passedIDs, a)
		recs[model.RootEdge] = rec
	} else { // node图通过分解nodeIDs拼接出edge获取attribute
		recs = edas.GenerateNodeRecs(passedIDs, a)
		recs[graphInfo.FirstNid] = rec
	}
	d.addHiddenVarsRecDBCache(c, recs, mid, graphInfo.Id, buvid, cursorMap) // 需要异步去补偿数据库和缓存
	return
}

// RawRecord is
func (d *Dao) rawHiddenVarsRec(c context.Context, mid, graphID, currentID, cursor int64) (a *model.HiddenVarsRecord, err error) {
	var (
		value string
		list  []*model.HiddenVar
	)
	if err = d.db.QueryRow(c, fmt.Sprintf(_rawRecordSQL, tableName(mid)), mid, graphID, currentID, cursor).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("RawRecord mid %d, graphID %d, currentID %d, cursor %d empty record", mid, graphID, currentID, cursor)
		} else {
			err = errors.Wrapf(err, "record by gid %d mid %d currentID %d cursor %d", graphID, mid, currentID, cursor)
		}
		return
	}
	if value == "" {
		return
	}
	if err = json.Unmarshal([]byte(value), &list); err != nil || len(list) == 0 {
		return
	}
	if len(list) == 0 {
		return
	}
	a = new(model.HiddenVarsRecord)
	a.Vars = make(map[string]*model.HiddenVar)
	for _, item := range list {
		a.Vars[item.ID] = item
	}
	return
}

// AddRecord adds a new record
func (d *Dao) addHiddenVarsRec(c context.Context, rec *model.HiddenVarRec) (err error) {
	if err = d.db.QueryRow(c, fmt.Sprintf(_recordExistSQL, tableName(rec.MID)), rec.MID, rec.GraphID, rec.CurrentID, rec.CursorID).Scan(&rec.ID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("RawRecord mid %d, graphID %d, currentID %d, cursor %d empty record", rec.MID, rec.GraphID, rec.CurrentID, rec.CursorID)
		} else {
			err = errors.Wrapf(err, "record by gid %d mid %d currentID %d cursor %d", rec.GraphID, rec.MID, rec.CurrentID, rec.CursorID)
			return
		}
	}
	if rec.ID > 0 {
		_, err = d.db.Exec(c, fmt.Sprintf(_updateRecSQL, tableName(rec.MID)), rec.Value, rec.ID)
		if err != nil {
			err = errors.Wrapf(err, "d.db.Exec(%s) error(%v)", _updateRecSQL, err)
			return
		}
		return
	}
	_, err = d.db.Exec(c, fmt.Sprintf(_insertRecSQL, tableName(rec.MID)), rec.MID, rec.GraphID, rec.CurrentID, rec.CursorID, rec.Value)
	if err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s) error(%v)", _insertRecSQL, err)
	}
	return

}
