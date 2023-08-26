package delay

import (
	"context"
	xsql "database/sql"

	"go-common/library/database/sql"
	"go-common/library/log"
)

const (
	_insertWF       = "INSERT INTO workflow_list(api_name,version) VALUES (?,?)"
	_upWFLog        = "UPDATE workflow_list SET display_name=?,display_state=?,log=? WHERE id=?"
	_upWFBoss       = "UPDATE workflow_list SET boss=?,display_name=?,display_state=? WHERE id=?"
	_selectLatestWF = "SELECT id,api_name,boss,version,wf_name,image,discovery_id,display_name,display_state,state,log FROM workflow_list WHERE api_name=? ORDER BY ctime DESC LIMIT 1"
	_selectWFByApi  = "SELECT id,api_name,boss,version,wf_name,image,discovery_id,display_name,display_state,state,log,mtime,ctime FROM workflow_list WHERE api_name=? LIMIT 1000"
	_selectAllWF    = "SELECT id,api_name,boss,version,wf_name,image,discovery_id,display_name,display_state,state,log,mtime,ctime FROM workflow_list WHERE state=0 LIMIT 10000"
	_upWFName       = "UPDATE workflow_list SET wf_name=? WHERE id=?"
	_upWFDiscovery  = "UPDATE workflow_list SET discovery_id=? WHERE id=?"
	_upWFImage      = "UPDATE workflow_list SET image=? WHERE id=?"
	_upWFState      = "UPDATE workflow_list SET state=? WHERE id=?"
	_upWFDisplay    = "UPDATE workflow_list SET display_name=?,display_state=? WHERE id=?"
	_upFailedWF     = "UPDATE workflow_list SET display_name=?,display_state=?,state=?,log=? WHERE id=?"
)

func (d *dao) AddRowDB(ctx context.Context, apiName, version string) (id int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(ctx, _insertWF, apiName, version); err != nil {
		log.Errorc(ctx, "d.AddRowDB error:%+v", err)
		return
	}
	id, _ = res.LastInsertId()
	return
}

func (d *dao) UpdateLog(ctx context.Context, id int64, dName string, dState int8, logs string) (err error) {
	if _, err = d.db.Exec(ctx, _upWFLog, dName, dState, logs, id); err != nil {
		log.Errorc(ctx, "d.UpdateLog error:%+v", err)
	}
	return
}

func (d *dao) UpdateBoss(ctx context.Context, id int64, boss string, dName string, dState int8) (err error) {
	if _, err = d.db.Exec(ctx, _upWFBoss, boss, dName, dState, id); err != nil {
		log.Errorc(ctx, "d.UpdateBoss error:%+v", err)
	}
	return
}

// 获取最新的一条发布记录
func (d *dao) GetLatestWF(ctx context.Context, apiName string) (res *WFDetail, err error) {
	res = &WFDetail{}
	if err = d.db.QueryRow(ctx, _selectLatestWF, apiName).Scan(&res.ID, &res.ApiName,
		&res.Boss, &res.Version, &res.WFName, &res.Image, &res.DiscoveryID, &res.DisplayName, &res.DisplayState, &res.State, &res.Log); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorc(ctx, "d.GetLatestWF error:%+v", err)
	}
	return
}

// 更新发布单的workflow名字
func (d *dao) UpdateWFName(ctx context.Context, id int64, wfName string) (err error) {
	if _, err = d.db.Exec(ctx, _upWFName, wfName, id); err != nil {
		log.Errorc(ctx, "d.UpdateWFName error:%+v", err)
	}
	return
}

// 更新发布单的discoveryID
func (d *dao) UpdateWFDis(ctx context.Context, id int64, discoveryID string) (err error) {
	if _, err = d.db.Exec(ctx, _upWFDiscovery, discoveryID, id); err != nil {
		log.Errorc(ctx, "d.UpdateWFDis error:%+v", err)
	}
	return
}

// 更新发布单的workflow镜像
func (d *dao) UpdateWFImage(ctx context.Context, id int64, image string) (err error) {
	if _, err = d.db.Exec(ctx, _upWFImage, image, id); err != nil {
		log.Errorc(ctx, "d.UpdateWFImage error:%+v", err)
	}
	return
}

// 更新发布单的workflow状态 比如完结或者发布失败时更新
func (d *dao) UpdateWFState(ctx context.Context, id int64, state int) (err error) {
	if _, err = d.db.Exec(ctx, _upWFState, state, id); err != nil {
		log.Errorc(ctx, "d.UpdateWFState error:%+v", err)
	}
	return
}

// 更新发布单的workflow的displayName
func (d *dao) UpdateWFDisplay(ctx context.Context, id int64, dName string, dState int8) (err error) {
	if _, err = d.db.Exec(ctx, _upWFDisplay, dName, dState, id); err != nil {
		log.Errorc(ctx, "d.UpdateWFDisplay error:%+v", err)
	}
	return
}

// 更新失败的发布单
func (d *dao) UpdateFailedWF(ctx context.Context, id int64, dName string, dState, state int8, logs string) (err error) {
	if _, err = d.db.Exec(ctx, _upFailedWF, dName, dState, state, logs, id); err != nil {
		log.Errorc(ctx, "d.UpdateFailedWF error:%+v", err)
	}
	return
}

// 获取某个应用所有的发布单
func (d *dao) GetWFByApi(ctx context.Context, apiName string) (res []*WFDetail, err error) {
	res = make([]*WFDetail, 0)
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectWFByApi, apiName); err != nil {
		log.Errorc(ctx, "d.GetWFByApi error:%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r = &WFDetail{}
		if err = rows.Scan(&r.ID, &r.ApiName, &r.Boss, &r.Version, &r.WFName, &r.Image, &r.DiscoveryID, &r.DisplayName, &r.DisplayState, &r.State, &r.Log, &r.Mtime, &r.Ctime); err != nil {
			log.Error("d.GetWFByApi error(%+v)", err)
			r = nil
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("d.GetWFByApi error:%+v", err)
	}
	return
}

// 获取所有运行状态的发布单
func (d *dao) GetAllWF(ctx context.Context) (res []*WFDetail, err error) {
	res = make([]*WFDetail, 0)
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectAllWF); err != nil {
		log.Errorc(ctx, "d.GetAllWF error:%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r = &WFDetail{}
		if err = rows.Scan(&r.ID, &r.ApiName, &r.Boss, &r.Version, &r.WFName, &r.Image, &r.DiscoveryID, &r.DisplayName, &r.DisplayState, &r.State, &r.Log, &r.Mtime, &r.Ctime); err != nil {
			log.Error("d.GetAllWF error(%+v)", err)
			r = nil
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("d.GetWFByApi error:%+v", err)
	}
	return
}
