package fawkes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/database/sql"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_txChannelAdd = `INSERT INTO channel (code,name,plate,status,operator,channel_status) VALUES(?,?,?,?,?,?) 
ON DUPLICATE KEY UPDATE code=?,name=?,plate=?,operator=?,channel_status=0`
	_checkChannelByCode = `SELECT count(1) FROM channel WHERE code=? AND name=? AND channel_status=0`
	_checkChannelByID   = `SELECT count(1) FROM channel WHERE id=? AND channel_status=0`
	_channelList        = `SELECT id,code,name,plate,status,operator,unix_timestamp(ctime),unix_timestamp(mtime) 
FROM channel WHERE status=1 AND channel_status=0 %s ORDER BY mtime DESC,id DESC LIMIT ? OFFSET ?`
	_channelAllList = `SELECT id,code,name,plate,status,operator,unix_timestamp(ctime),unix_timestamp(mtime) 
FROM channel WHERE status=1 AND channel_status=0`
	_txChannelDelete     = `UPDATE channel SET channel_status=1,operator=? WHERE id=?`
	_appChannelListCount = `SELECT count(*) FROM channel AS c,app_channel AS ac WHERE ac.app_key=? AND c.id=ac.channel_id AND c.channel_status=0 %s`
	_appChannelList      = `SELECT c.id,ac.id as aid,c.code,c.name,c.plate,c.status,c.operator, IFNULL(acg.id, 0) AS acg_id, IFNULL(acg.gname,"") AS acg_name,IFNULL(acg.description,"") AS description,IFNULL(acg.auto_push_cdn, 0) AS auto_push_cdn,IFNULL(acg.is_auto_gen, 0) AS is_auto_gen,IFNULL(acg.qa_owner, '') AS qa_owner,IFNULL(acg.market_owner, '') AS market_owner,unix_timestamp(ac.ctime),unix_timestamp(ac.mtime) 
FROM channel AS c ,app_channel AS ac LEFT JOIN app_channel_group AS acg on ac.group_id=acg.id WHERE ac.app_key=? AND c.id=ac.channel_id AND c.channel_status=0 %s ORDER BY %s %s`
	_getChannelByID = `SELECT id,code,name,plate,status,operator,unix_timestamp(ctime),unix_timestamp(mtime) 
FROM channel WHERE id=? AND channel_status=0`
	_txCustomChannelDeleteByID = `DELETE FROM channel WHERE id=?`
	_getChannelCount           = `SELECT count(1) FROM channel WHERE status=1 AND channel_status=0 %s`
	_getAppChannelCount        = `SELECT count(1) FROM channel AS c,app_channel AS ac WHERE ac.app_key=? AND c.id=ac.channel_id AND c.channel_status=0 %s`
	_appChannelAdd             = `INSERT INTO app_channel (app_key,channel_id,group_id,operator) VALUES(?,?,?,?)`
	_appChannelAdds            = `INSERT INTO app_channel (app_key,channel_id,operator) VALUES %s`
	_checkAppChannel           = `SELECT count(1) FROM app_channel WHERE app_key=? AND channel_id=?`
	_appChannelDelete          = `DELETE FROM app_channel WHERE app_key=? AND channel_id=?`
	_getChannelIDByCode        = `SELECT id FROM channel WHERE code=? AND name=? AND plate=? AND channel_status=0`
	_getAppCountByID           = `SELECT count(1) FROM app_channel WHERE channel_id=?`
	_txChannelToStatic         = `UPDATE channel SET status=1,operator=? WHERE id=?`
	_channleByCode             = `SELECT id,code,name,plate FROM channel WHERE code=?`
	_appChannelGroupRelate     = `UPDATE app_channel SET group_id = CASE id %s END WHERE id IN (%s)`
	_resetAppChannelGroup      = `UPDATE app_channel SET group_id=0, operator=? WHERE group_id=?`
	_appChannelGroupList       = `SELECT id,gname,description,operator,auto_push_cdn,is_auto_gen,qa_owner,market_owner,priority,unix_timestamp(mtime),unix_timestamp(ctime) FROM app_channel_group WHERE app_key=? AND state=1 %s`
	_appChanelGroupAdd         = `INSERT INTO app_channel_group (app_key, gname, description,operator,auto_push_cdn,is_auto_gen,qa_owner,market_owner,priority) VALUES (?,?,?,?,?,?,?,?,?)`
	_appChannelGroupUpdate     = `UPDATE app_channel_group SET gname=?,description=?,operator=?,auto_push_cdn=?,is_auto_gen=?,qa_owner=?,market_owner=?,priority=? WHERE id=?`
	_AppChannelGroupDel        = `UPDATE app_channel_group SET state=0, operator=? WHERE id=?`
)

// ChannelList get static channel list
func (d *Dao) ChannelList(c context.Context, page, size int, filterKey string) (chLists []*appmdl.Channel, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey, filterKey)
		sqlAdd = "AND (code LIKE ? OR name LIKE ? OR plate LIKE ?)"
	}
	args = append(args, size, (page-1)*size)
	rows, err := d.db.Query(c, fmt.Sprintf(_channelList, sqlAdd), args...)
	if err != nil {
		log.Error("ChannelList query failed. %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		stcChannel := &appmdl.Channel{}
		if err = rows.Scan(&stcChannel.ID, &stcChannel.Code, &stcChannel.Name, &stcChannel.Plate, &stcChannel.Status,
			&stcChannel.Operator, &stcChannel.Ctime, &stcChannel.Mtime); err != nil {
			log.Error("ChannelList rows.Scan failed. %v", err)
			return
		}
		chLists = append(chLists, stcChannel)
	}
	err = rows.Err()
	return
}

// ChannelAllList get all static channel list
func (d *Dao) ChannelAllList(c context.Context) (chLists []*appmdl.Channel, err error) {
	rows, err := d.db.Query(c, _channelAllList)
	if err != nil {
		log.Error("ChannelList query failed. %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		stcChannel := &appmdl.Channel{}
		if err = rows.Scan(&stcChannel.ID, &stcChannel.Code, &stcChannel.Name, &stcChannel.Plate, &stcChannel.Status,
			&stcChannel.Operator, &stcChannel.Ctime, &stcChannel.Mtime); err != nil {
			log.Error("ChannelList rows.Scan failed. %v", err)
			return
		}
		chLists = append(chLists, stcChannel)
	}
	err = rows.Err()
	return
}

// TxChannelAdd add static channel
func (d *Dao) TxChannelAdd(tx *sql.Tx, code, name, plate, operator string, status, channelStatus int8) (id int64, err error) {
	res, err := tx.Exec(_txChannelAdd, code, name, plate, status, operator, channelStatus, code, name, plate, operator)
	if err != nil {
		log.Error("TxChannelAdd INSERT failed. %v", err)
		return
	}
	return res.LastInsertId()
}

// TxChannelToStatic change channel status to static
func (d *Dao) TxChannelToStatic(tx *sql.Tx, channelID int64, operator string) (r int64, err error) {
	res, err := tx.Exec(_txChannelToStatic, operator, channelID)
	if err != nil {
		log.Error("TxChannelToStatic UPDATE failed. %v", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxChannelDelete delete static channel
func (d *Dao) TxChannelDelete(tx *sql.Tx, channelID int64, operator string) (r int64, err error) {
	res, err := tx.Exec(_txChannelDelete, operator, channelID)
	if err != nil {
		log.Error("TxChannelDelete UPDATE failed. %v", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// CheckChannelByCode check for channel by code and name
func (d *Dao) CheckChannelByCode(c context.Context, code, name string) (count int, err error) {
	rows := d.db.QueryRow(c, _checkChannelByCode, code, name)
	if err = rows.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("CheckChannelByCode %v", err)
		}
	}
	return
}

// GetChannelIDByCode get channelID by code and name
func (d *Dao) GetChannelIDByCode(c context.Context, code, name, plate string) (channelID int64, err error) {
	res := d.db.QueryRow(c, _getChannelIDByCode, code, name, plate)
	if err = res.Scan(&channelID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetChannelIDByCode %v", err)
		}
	}
	return
}

// CheckChannelByID check for channel by id
func (d *Dao) CheckChannelByID(c context.Context, channelID int64) (count int64, err error) {
	rows := d.db.QueryRow(c, _checkChannelByID, channelID)
	if err = rows.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("CheckChannelByID %v", err)
		}
	}
	return
}

// AppChannelListCount get app channel list count numeber
func (d *Dao) AppChannelListCount(c context.Context, appKey, filterKey string, groupID int64) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey, filterKey)
		sqlAdd = "AND (c.code LIKE ? OR c.name LIKE ? OR c.plate LIKE ?)"
	}
	if groupID != -1 {
		args = append(args, groupID)
		sqlAdd += "AND ac.group_id=?"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_appChannelListCount, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppChannelListCount %v", err)
		}
	}
	return
}

// AppChannelList get app channel list
func (d *Dao) AppChannelList(c context.Context, appKey, filterKey, order, sort string, pn, ps int, groupID int64) (chList []*appmdl.Channel, err error) {
	var (
		sqlAdd   string
		sqlLimit string
		sqlOrder string
		args     []interface{}
	)
	args = append(args, appKey)
	sqlOrder = "ac.mtime DESC"
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey, filterKey)
		sqlAdd = "AND (c.code LIKE ? OR c.name LIKE ? OR c.plate LIKE ?)"
	}
	if groupID != -1 {
		args = append(args, groupID)
		sqlAdd += "AND ac.group_id=?"
	}
	if order != "" && sort != "" {
		lowerOrder := strings.ToLower(order)
		sqlOrder = "c.id"
		switch lowerOrder {
		case "mtime":
			sqlOrder = "c.mtime"
		case "name":
			sqlOrder = "c.name"
		case "code":
			sqlOrder = "c.code"
		}
		if strings.ToLower(sort) == "asc" {
			sqlOrder += " ASC"
		} else {
			sqlOrder += " DESC"
		}
	}
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlLimit += "LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appChannelList, sqlAdd, sqlOrder, sqlLimit), args...)
	if err != nil {
		log.Error("AppChannelList SELECT failed. %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		channelTemp := &appmdl.Channel{}
		channelGroupInfo := &appmdl.ChannelGroupInfo{}
		if err = rows.Scan(&channelTemp.ID, &channelTemp.AID, &channelTemp.Code, &channelTemp.Name, &channelTemp.Plate, &channelTemp.Status,
			&channelTemp.Operator, &channelGroupInfo.ID, &channelGroupInfo.Name, &channelGroupInfo.Description, &channelGroupInfo.AutoPushCdn, &channelGroupInfo.IsAutoGen, &channelGroupInfo.QaOwner, &channelGroupInfo.MarketOwner,
			&channelTemp.Ctime, &channelTemp.Mtime); err != nil {
			log.Error("AppChannelList rows.Scan() failed. %v", err)
			return
		}
		if channelGroupInfo.ID != 0 {
			channelTemp.Group = channelGroupInfo
		}
		chList = append(chList, channelTemp)
	}
	err = rows.Err()
	return
}

// GetAppChannelCount get app channel total
func (d *Dao) GetAppChannelCount(c context.Context, appKey, filterKey string) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey, filterKey)
		sqlAdd = "AND (c.code LIKE ? OR c.name LIKE ? OR c.plate LIKE ?)"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_getAppChannelCount, sqlAdd), args...)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetAppChannelCount row.Scan failed. %v", err)
		}
	}
	return
}

// AppChannelAdd add app channel
func (d *Dao) AppChannelAdd(tx *sql.Tx, channelID, groupID int64, appKey, operator string) (r int64, err error) {
	rows, err := tx.Exec(_appChannelAdd, appKey, channelID, groupID, operator)
	if err != nil {
		log.Error("AppChannelAdd INSERT failed. %v", err)
		return
	}
	r, err = rows.LastInsertId()
	return
}

// AppChannelAdds add app channel
func (d *Dao) AppChannelAdds(tx *sql.Tx, sqls []string, sqlAdd []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_appChannelAdds, strings.Join(sqls, ",")), sqlAdd...)
	if err != nil {
		log.Error("TxSetGenerates %v", err)
		return
	}
	return res.RowsAffected()
}

// CheckAppChannel check for app channel by app_key and channel_id
func (d *Dao) CheckAppChannel(c context.Context, appKey string, channelID int64) (count int, err error) {
	rows := d.db.QueryRow(c, _checkAppChannel, appKey, channelID)
	if err = rows.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("CheckAppChannel rows.Scan() failed. %v", err)
		}
	}
	return
}

// GetAppCountByID get app count by channelID
func (d *Dao) GetAppCountByID(c context.Context, channelID int64) (count int, err error) {
	res := d.db.QueryRow(c, _getAppCountByID, channelID)
	if err = res.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetAppCountByID res.Scan() failed. %v", err)
		}
	}
	return
}

// AppChannelDelete delete app channel
func (d *Dao) AppChannelDelete(tx *sql.Tx, appKey string, channelID int64) (r int64, err error) {
	rows, err := tx.Exec(_appChannelDelete, appKey, channelID)
	if err != nil {
		log.Error("AppChannelDelete DELETE failed. %v", err)
		return
	}
	r, err = rows.RowsAffected()
	return
}

// GetChannelByID get Channel status By id
func (d *Dao) GetChannelByID(c context.Context, channelID int64) (chel *appmdl.Channel, err error) {
	res := d.db.QueryRow(c, _getChannelByID, channelID)
	chel = &appmdl.Channel{}
	if err = res.Scan(&chel.ID, &chel.Code, &chel.Name, &chel.Plate, &chel.Status, &chel.Operator, &chel.Ctime, &chel.Mtime); err != nil {
		if sql.ErrNoRows == err {
			chel = nil
			err = nil
		} else {
			log.Error("GetChannelByID failed. %v", err)
		}
	}
	return
}

// TxCustomChannelDeleteByID delete custom channel
func (d *Dao) TxCustomChannelDeleteByID(tx *sql.Tx, channelID int64) (r int64, err error) {
	rows, err := tx.Exec(_txCustomChannelDeleteByID, channelID)
	if err != nil {
		log.Error("TxCustomChannelDeleteByID DELETE failed. %v", err)
		return
	}
	r, err = rows.RowsAffected()
	return
}

// GetChannelCount get static channel total
func (d *Dao) GetChannelCount(c context.Context, filterKey string) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey, filterKey)
		sqlAdd = "AND (code LIKE ? OR name LIKE ? OR plate LIKE ?)"
	}
	rows := d.db.QueryRow(c, fmt.Sprintf(_getChannelCount, sqlAdd), args...)
	if err = rows.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetChannelCount %v", err)
		}
	}
	return
}

// ChannelByCode get channle by code.
func (d *Dao) ChannelByCode(c context.Context, code string) (re *appmdl.Channel, err error) {
	row := d.db.QueryRow(c, _channleByCode, code)
	re = &appmdl.Channel{}
	if err = row.Scan(&re.ID, &re.Code, &re.Name, &re.Plate); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("ChannelByCode %v", err)
		}
	}
	return
}

// TxAppChannelGroupRelate relate app channel to group
func (d *Dao) TxAppChannelGroupRelate(tx *sql.Tx, appChannelIDs []int64, groudID int64, userName string) (err error) {
	var (
		sql  string
		args []interface{}
		ids  []string
	)
	for _, appChannelID := range appChannelIDs {
		sql = sql + " WHEN ? THEN ?"
		args = append(args, appChannelID, groudID)
		ids = append(ids, strconv.FormatInt(appChannelID, 10))
	}
	res, err := tx.Exec(fmt.Sprintf(_appChannelGroupRelate, sql, strings.Join(ids, ",")), args...)
	if err != nil {
		log.Error("TxAppChannelGroupRelate error %v", err)
		return
	}
	_, err = res.RowsAffected()
	return
}

// AppChannelGroupList get channelgroup list
func (d *Dao) AppChannelGroupList(c context.Context, appKey, filterKey string) (res []*appmdl.ChannelGroup, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		sqlAdd += "AND gname LIKE ?"
		args = append(args, filterKey)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appChannelGroupList, sqlAdd), args...)
	if err != nil {
		log.Error("_appChannelGroupList query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		channelGroup := &appmdl.ChannelGroup{}
		if err = rows.Scan(&channelGroup.ID, &channelGroup.Name, &channelGroup.Description, &channelGroup.Operator,
			&channelGroup.AutoPushCdn, &channelGroup.IsAutoGen, &channelGroup.QaOwner, &channelGroup.MarketOwner, &channelGroup.Priority,
			&channelGroup.Mtime, &channelGroup.Ctime); err != nil {
			log.Error("_appChannelGroupList rows.Scan eror")
			return
		}
		res = append(res, channelGroup)
	}
	err = rows.Err()
	return
}

// TxAppChannelGroupAdd add app channel group
func (d *Dao) TxAppChannelGroupAdd(tx *sql.Tx, appKey, name, description, username, qaOwner, marketOwner string, autoPushCdn, autoGen, priority int64) (err error) {
	res, err := tx.Exec(_appChanelGroupAdd, appKey, name, description, username, autoPushCdn, autoGen, qaOwner, marketOwner, priority)
	if err != nil {
		log.Error("TxAppChannelGroupAdd error %v", err)
		return
	}
	_, err = res.RowsAffected()
	return
}

// TxAppChannelGroupUpdate update channel group
func (d *Dao) TxAppChannelGroupUpdate(tx *sql.Tx, id int64, name, description, userName string, autoPushCdn, autoGen int64, qaOwner, marketOwner string, priority int64) (err error) {
	res, err := tx.Exec(_appChannelGroupUpdate, name, description, userName, autoPushCdn, autoGen, qaOwner, marketOwner, priority, id)
	if err != nil {
		log.Error("TxAppChannelGroupUpdate error %v", err)
		return
	}
	_, err = res.RowsAffected()
	return
}

// TxResetAppChannelGroup reset app channel group
func (d *Dao) TxResetAppChannelGroup(tx *sql.Tx, id int64, userName string) (err error) {
	res, err := tx.Exec(_resetAppChannelGroup, userName, id)
	if err != nil {
		log.Error("TxResetAppChannelGroup error %v", err)
		return
	}
	_, err = res.RowsAffected()
	return
}

// TxAppChannelGroupDel del app channel group
func (d *Dao) TxAppChannelGroupDel(tx *sql.Tx, id int64, userName string) (err error) {
	res, err := tx.Exec(_AppChannelGroupDel, userName, id)
	if err != nil {
		log.Error("TxAppChannelGroupDel error %v", err)
		return
	}
	_, err = res.RowsAffected()
	return
}
