package fawkes

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"

	ffmdl "go-gateway/app/app-svr/fawkes/service/model/ff"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_ffWhitchList   = `SELECT app_key,env,mid,nick,operator,unix_timestamp(ctime) FROM ff_whith WHERE app_key=? AND env=?`
	_addFFWhithlist = `INSERT INTO ff_whith (app_key,env,mid,operator) VALUES %v`
	_delFFWhithlist = `DELETE FROM ff_whith WHERE app_key=? AND env=? AND mid=?`

	_ffcount     = `SELECT count(*) FROM ff_config WHERE app_key=? AND env=? %s`
	_addFFConfig = `INSERT INTO ff_config (app_key,env,name,description,status,salt,bucket,bucket_count,version,
un_version,rom_version,brand,un_brand,network,isp,channel,whith,black_mid,black_list,operator,state) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`
	_upFFConfig = "UPDATE ff_config SET description=?,status=?,salt=?,bucket=?,bucket_count=?,version=?,un_version=?," +
		"rom_version=?,brand=?,un_brand=?,network=?,isp=?,channel=?,whith=?,black_mid=?,black_list=?,operator=?,state=? WHERE app_key=? AND env=? AND name=?"
	_fflist = `SELECT id,app_key,env,name,description,status,salt,bucket,bucket_count,version,un_version,
rom_version,brand,un_brand,network,isp,channel,whith,black_mid,black_list,state,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM ff_config 
WHERE app_key=? AND env=? AND state>0 %s ORDER BY id DESC %s`
	_ffModiyCount  = `SELECT count(*) FROM ff_config WHERE app_key=? AND env=? AND state<>3`
	_fulshffConfig = `DELETE FROM ff_config WHERE app_key=? AND env=? AND state<0`
	_ffconfigs     = `SELECT id,app_key,env,name,description,status,salt,bucket,bucket_count,version,un_version,
rom_version,brand,un_brand,network,isp,channel,whith,black_mid,black_list,state,operator,unix_timestamp(mtime) FROM ff_config WHERE app_key=? AND env=?`
	_ffconfig = `SELECT id,app_key,env,name,description,status,salt,bucket,bucket_count,version,un_version,
rom_version,brand,un_brand,network,isp,channel,whith,black_mid,black_list,state,operator,unix_timestamp(mtime) FROM ff_config WHERE app_key=? AND env=? AND name=?`
	_delFFConfig     = `UPDATE ff_config SET state=-1 WHERE app_key=? AND env=? AND name=?`
	_delFFConfig2    = `DELETE FROM ff_config WHERE app_key=? AND env=? AND name=?`
	_upffConfigState = `UPDATE ff_config SET state=? WHERE app_key=? AND env=?`

	_setFFConfigPublish      = `INSERT INTO ff_publish (app_key,env,description,operator) VALUES(?,?,?,?)`
	_upFFConfigPublishURL    = `UPDATE ff_publish SET cdn_url=?,local_path=?,diffs=? WHERE id=?`
	_upFFConfigPublishStatus = `UPDATE ff_publish SET state=? WHERE app_key=? AND env=? AND id=?`
	_upffPublishTotal        = `UPDATE ff_publish SET total_path=?,total_url=? WHERE id=?`
	_appFFHistory            = `SELECT id,app_key,env,description,cdn_url,local_path,diffs,total_url,total_path,operator,
state,unix_timestamp(ctime),unix_timestamp(mtime) FROM ff_publish WHERE app_key=? AND env=? ORDER BY id DESC %v`
	_appFFHistoryCount = `SELECT count(*) FROM ff_publish WHERE app_key=? AND env=?`
	_appFFHistoryByID  = `SELECT id,app_key,env,description,cdn_url,local_path,diffs,total_url,total_path,operator,
state,unix_timestamp(ctime),unix_timestamp(mtime) FROM ff_publish WHERE app_key=? AND id=?`
	_ffconfigPublishLastID = `SELECT id FROM ff_publish WHERE app_key=? AND env=? %v ORDER BY id DESC LIMIT 1`

	_setFFConfigFile = `INSERT INTO ff_config_file (app_key,env,ff_id,name,description,status,salt,bucket,bucket_count,
version,un_version,rom_version,brand,un_brand,network,isp,channel,whith,black_mid,black_list,operator,state) VALUES %v`
	_appFFFile = `SELECT app_key,env,name,description,status,salt,bucket,bucket_count,version,un_version,rom_version,
brand,un_brand,network,isp,channel,whith,black_mid,black_list,operator,state,unix_timestamp(mtime) FROM ff_config_file WHERE app_key=? AND env=? AND ff_id=?`
)

// AppFFWhithlist get overrall ff whith list.
func (d *Dao) AppFFWhithlist(c context.Context, appKey, env string) (res []*ffmdl.Whitch, err error) {
	rows, err := d.db.Query(c, _ffWhitchList, appKey, env)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &ffmdl.Whitch{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.MID, &re.Nick, &re.Operator, &re.CTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddFFWhithlist add ff overrall whith list.
func (d *Dao) TxAddFFWhithlist(tx *xsql.Tx, appKey, env, userName string, mids []int64) (r int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, mid := range mids {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, appKey, env, mid, userName)
	}
	res, err := tx.Exec(fmt.Sprintf(_addFFWhithlist, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxAddFFWhithlist %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelFFWhithlist del ff overrall whith list.
func (d *Dao) TxDelFFWhithlist(tx *xsql.Tx, appKey, env string, mid int64) (r int64, err error) {
	res, err := tx.Exec(_delFFWhithlist, appKey, env, mid)
	if err != nil {
		log.Error("TxDelFFWhithlist %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetFFConfig set ff config.
func (d *Dao) TxSetFFConfig(tx *xsql.Tx, appKey, env, userName, name, desc, status, salt, bucket, version, unVersion,
	romVersion, brand, unBrand, network, isp, channel, whith, blackMid, blackList string, bucketCount int64) (r int64, err error) {
	res, err := tx.Exec(_addFFConfig, appKey, env, name, desc, status, salt, bucket, bucketCount, version, unVersion,
		romVersion, brand, unBrand, network, isp, channel, whith, blackMid, blackList, userName)
	if err != nil {
		log.Error("TxSetFFConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpFFConfig set ff config.
func (d *Dao) TxUpFFConfig(tx *xsql.Tx, appKey, env, userName, name, desc, status, salt, bucket, version, unVersion,
	romVersion, brand, unBrand, network, isp, channel, whith, blackMid, blackList string, bucketCount int64, state int) (r int64, err error) {
	res, err := tx.Exec(_upFFConfig, desc, status, salt, bucket, bucketCount, version, unVersion, romVersion, brand,
		unBrand, network, isp, channel, whith, blackMid, blackList, userName, state, appKey, env, name)
	if err != nil {
		log.Error("TxUpFFConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// FFCount get ff count.
func (d *Dao) FFCount(c context.Context, appKey, env, filterKey string) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey)
		sqlAdd = "AND ((name LIKE ?) OR (description LIKE ?))"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_ffcount, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("FFCount %v", err)
		}
	}
	return
}

// FFList get ff list.
func (d *Dao) FFList(c context.Context, appKey, env, filterKey string, pn, ps int) (res []*ffmdl.FF, err error) {
	var (
		sqlAdd       string
		filterSqlAdd string
		args         []interface{}
	)
	args = append(args, appKey, env)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey)
		filterSqlAdd = "AND ((name LIKE ?) OR (description LIKE ?))"
	}
	if pn != 0 && ps != 0 {
		sqlAdd = "LIMIT ?,?"
		args = append(args, (pn-1)*ps, ps)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_fflist, filterSqlAdd, sqlAdd), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &ffmdl.FF{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Key, &re.Desc, &re.Status, &re.Salt, &re.Bucket,
			&re.BucketCount, &re.Version, &re.UnVersion, &re.RomVersion, &re.Brand, &re.UnBrand, &re.Network, &re.ISP,
			&re.Channel, &re.Whith, &re.BlackMid, &re.BlackList, &re.State, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// FFModiyCount get ff modify count.
func (d *Dao) FFModiyCount(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _ffModiyCount, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("FFModiyCount %v", err)
		}
	}
	return
}

// TxAddFFPublish add ff publish.
func (d *Dao) TxAddFFPublish(tx *xsql.Tx, appKey, env, userName, desc string) (id int64, err error) {
	res, err := tx.Exec(_setFFConfigPublish, appKey, env, desc, userName)
	if err != nil {
		log.Error("TxAddFFPublish %v", err)
		return
	}
	return res.LastInsertId()
}

// TxAddFFConfigFile add ff config file.
func (d *Dao) TxAddFFConfigFile(tx *xsql.Tx, sqls []string, args []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_setFFConfigFile, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxAddFFConfigFile %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpFFConfigPublishURL up ff config publish url.
func (d *Dao) TxUpFFConfigPublishURL(tx *xsql.Tx, url, localPath, md5, diff string, fvid int64) (r int64, err error) {
	res, err := tx.Exec(_upFFConfigPublishURL, url, localPath, diff, fvid)
	if err != nil {
		log.Error("TxUpFFConfigPublishURL %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpFFPublishTotal up ff config publish total.
func (d *Dao) TxUpFFPublishTotal(tx *xsql.Tx, totalPath, totalURL string, fvid int64) (r int64, err error) {
	res, err := tx.Exec(_upffPublishTotal, totalPath, totalURL, fvid)
	if err != nil {
		log.Error("TxUpFFPublishTotal %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpFFConfigPublishState up ff config published state to -1.
func (d *Dao) TxUpFFConfigPublishState(tx *xsql.Tx, appKey, env string, fvid int64, state int) (r int64, err error) {
	res, err := tx.Exec(_upFFConfigPublishStatus, state, appKey, env, fvid)
	if err != nil {
		log.Error("TxUpFFConfigPublishState %v", err)
		return
	}
	return res.RowsAffected()
}

// TxFlushFFConfig delete ff_config state<0
func (d *Dao) TxFlushFFConfig(tx *xsql.Tx, appKey, env string) (r int64, err error) {
	res, err := tx.Exec(_fulshffConfig, appKey, env)
	if err != nil {
		log.Error("TxFlushFFConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpFFConfigState up ff config state to publish 3.
func (d *Dao) TxUpFFConfigState(tx *xsql.Tx, appKey, env string, state int) (r int64, err error) {
	res, err := tx.Exec(_upffConfigState, state, appKey, env)
	if err != nil {
		log.Error("TxUpFFConfigState %v", err)
		return
	}
	return res.RowsAffected()
}

// AppFFHistoryCount get app ff publish count.
func (d *Dao) AppFFHistoryCount(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _appFFHistoryCount, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppFFHistoryCount %v", err)
		}
	}
	return
}

// AppFFHistory get ff config publish history.
func (d *Dao) AppFFHistory(c context.Context, appKey, env string, pn, ps int) (res []*ffmdl.ConfigPublish, err error) {
	var (
		sqlLimit string
		args     []interface{}
	)
	args = append(args, appKey, env)
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlLimit += "LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appFFHistory, sqlLimit), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &ffmdl.ConfigPublish{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Desc, &re.URL, &re.LocalPath, &re.Diffs, &re.TotalURL,
			&re.TotalLocalPath, &re.Operator, &re.State, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AppFFLastFvid get ff count.
func (d *Dao) AppFFLastFvid(c context.Context, appKey, env string, fvid int64) (id int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if fvid > 0 {
		sqlAdd = " AND id<?"
		args = append(args, fvid)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_ffconfigPublishLastID, sqlAdd), args...)
	if err = row.Scan(&id); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppFFLastFvid %v", err)
		}
	}
	return
}

// AppFFHistoryByID get ff history by id.
func (d *Dao) AppFFHistoryByID(c context.Context, appKey string, ffid int64) (re *ffmdl.ConfigPublish, err error) {
	row := d.db.QueryRow(c, _appFFHistoryByID, appKey, ffid)
	re = &ffmdl.ConfigPublish{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Env, &re.Desc, &re.URL, &re.LocalPath, &re.Diffs, &re.TotalURL,
		&re.TotalLocalPath, &re.Operator, &re.State, &re.CTime, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppFFHistoryByID %v", err)
		}
	}
	return
}

// AppFFFile get aa ff_file.
func (d *Dao) AppFFFile(c context.Context, appKey, env string, fvid int64) (res []*ffmdl.File, err error) {
	rows, err := d.db.Query(c, _appFFFile, appKey, env, fvid)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &ffmdl.File{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.Key, &re.Desc, &re.Status, &re.Salt, &re.Bucket, &re.BucketCount,
			&re.Version, &re.UnVersion, &re.RomVersion, &re.Brand, &re.UnBrand, &re.Network, &re.ISP, &re.Channel,
			&re.Whith, &re.BlackMid, &re.BlackList, &re.Operator, &re.State, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AppFFConfigs get ff config.
func (d *Dao) AppFFConfigs(c context.Context, appKey, env string) (res []*ffmdl.FF, err error) {
	rows, err := d.db.Query(c, _ffconfigs, appKey, env)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &ffmdl.FF{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Key, &re.Desc, &re.Status, &re.Salt, &re.Bucket,
			&re.BucketCount, &re.Version, &re.UnVersion, &re.RomVersion, &re.Brand, &re.UnBrand, &re.Network, &re.ISP, &re.Channel, &re.Whith, &re.BlackMid, &re.BlackList, &re.State, &re.Operator, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AppFFConfig get ff config.
func (d *Dao) AppFFConfig(c context.Context, appKey, env, key string) (re *ffmdl.FF, err error) {
	row := d.db.QueryRow(c, _ffconfig, appKey, env, key)
	re = &ffmdl.FF{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Env, &re.Key, &re.Desc, &re.Status, &re.Salt, &re.Bucket, &re.BucketCount,
		&re.Version, &re.UnVersion, &re.RomVersion, &re.Brand, &re.UnBrand, &re.Network, &re.ISP, &re.Channel, &re.Whith, &re.BlackMid,
		&re.BlackList, &re.State, &re.Operator, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("%v", err)
		}
	}
	return
}

// TxDelFFConfig del ff config.
func (d *Dao) TxDelFFConfig(tx *xsql.Tx, appKey, env, name string) (r int64, err error) {
	res, err := tx.Exec(_delFFConfig, appKey, env, name)
	if err != nil {
		log.Error("TxDelFFConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelFFConfig2 true·del ff config true.
func (d *Dao) TxDelFFConfig2(tx *xsql.Tx, appKey, env, name string) (r int64, err error) {
	res, err := tx.Exec(_delFFConfig2, appKey, env, name)
	if err != nil {
		log.Error("TxDelFFConfig2 %v", err)
		return
	}
	return res.RowsAffected()
}
