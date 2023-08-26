package appstore

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/cache"
	xsql "go-common/library/database/sql"
	appstoremdl "go-gateway/app/web-svr/activity/interface/model/appstore"

	"github.com/pkg/errors"
)

// appstoreDao const.
const (
	_appstoreAllSQL = "SELECT `id`,`name`,`model_name`,`batch_token`,`appkey`,`operator`,`remark`,`ctime`,`mtime`,`start_time`,`end_time`,`state` FROM activity_appstore"
)

// RawAppstoreAll get activity_appstore all .
func (dao *Dao) RawAppstoreAll(ctx context.Context) (datas []*appstoremdl.ActivityAppstore, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(ctx, _appstoreAllSQL); err != nil {
		err = errors.Wrap(err, "RawActivityAppstoreAll:dao.db.Query()")
		return
	}
	defer rows.Close()
	datas = make([]*appstoremdl.ActivityAppstore, 0)
	for rows.Next() {
		data := &appstoremdl.ActivityAppstore{}
		if err = rows.Scan(&data.ID, &data.Name, &data.ModelName, &data.BatchToken, &data.Appkey, &data.Operator, &data.Remark, &data.Ctime, &data.Mtime, &data.StartTime, &data.EndTime, &data.State); err != nil {
			err = errors.Wrap(err, "RawActivityAppstoreAll:rows.Scan")
			return
		}
		datas = append(datas, data)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawActivityAppstoreAll: rows.Err()")
	}
	return
}

const (
	//`id`,`mid`,`tel_hash`,`batch_token`,`fingerprint`,`local_fingerprint`,`buvid`,`build`,`order_no`,`state`,`user_ip`,`ctime`,`mtime`
	_countAppstoreReceivedByMIDSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `batch_token`=? AND `mid`=? AND `state`=1"
)

// CountAppstoreReceivedByMID .
func (d *Dao) CountAppstoreReceivedByMID(c context.Context, batchToken string, mid int64) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByMIDSQL, batchToken, mid)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByMID:row.Scan()")
		}
	}
	return
}

const (
	_countAppstoreReceivedByTelHashSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `batch_token`=? AND `tel_hash`=? AND `state`=1"
)

// CountAppstoreReceivedByTel .
func (d *Dao) CountAppstoreReceivedByTel(c context.Context, batchToken string, telHash string) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByTelHashSQL, batchToken, telHash)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByTel:row.Scan()")
		}
	}
	return
}

const (
	_countAppstoreReceivedByMatchSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `batch_token`=? AND `match_label`=? AND `match_kind`=? AND `state`=1"
)

// CountAppstoreReceivedByMatch .
func (d *Dao) CountAppstoreReceivedByMatch(c context.Context, batchToken string, matchLabel string, matchKind int64) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByMatchSQL, batchToken, matchLabel, matchKind)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByMatch:row.Scan()")
		}
	}
	return
}

const (
	_countAppstoreReceivedByFingerprintSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `fingerprint`=?"
)

// CountAppstoreReceivedByFingerprint .
func (d *Dao) CountAppstoreReceivedByFingerprint(c context.Context, fingerprint string) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByFingerprintSQL, fingerprint)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByFingerprint:row.Scan()")
		}
	}
	return
}

const (
	_countAppstoreReceivedByLocalFingerprintSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `local_fingerprint`=?"
)

// CountAppstoreReceivedByLocalFingerprint .
func (d *Dao) CountAppstoreReceivedByLocalFingerprint(c context.Context, localFingerprint string) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByLocalFingerprintSQL, localFingerprint)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByLocalFingerprint:row.Scan()")
		}
	}
	return
}

const (
	_countAppstoreReceivedByBuvidSQL = "SELECT COUNT(1) FROM activity_appstore_received WHERE `buvid`=?"
)

// CountAppstoreReceivedByLocalBuvid .
func (d *Dao) CountAppstoreReceivedByBuvid(c context.Context, buvid string) (count int64, err error) {
	row := d.db.QueryRow(c, _countAppstoreReceivedByBuvidSQL, buvid)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountAppstoreReceivedByLocalBuvid:row.Scan()")
		}
	}
	return
}

const (
	_InsertAppstoreReceivedSQL = "INSERT INTO activity_appstore_received(`mid`,`tel_hash`,`batch_token`,`fingerprint`,`local_fingerprint`,`buvid`,`match_label`,`match_kind`,`build`,`order_no`,`state`,`user_ip`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
)

// AddAppstoreReceived .
func (d *Dao) AddAppstoreReceived(c context.Context, arg *appstoremdl.ActivityAppstoreReceived) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _InsertAppstoreReceivedSQL, arg.Mid, arg.TelHash, arg.BatchToken, arg.Fingerprint, arg.LocalFingerprint, arg.Buvid, arg.MatchLabel, arg.MatchKind, arg.Build, arg.OrderNo, arg.State, arg.UserIP); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("AddAppstoreReceived error d.db.Exec(%+v)", arg))
		return
	}
	return res.LastInsertId()
}

const (
	_UpdateAppstoreStateSQL = "UPDATE activity_appstore SET `state`=2 WHERE `batch_token`=?"
)

// UpdateAppstoreState .
func (d *Dao) UpdateAppstoreState(c context.Context, batchToken string) (res int64, err error) {
	var (
		sqlRes sql.Result
	)
	if sqlRes, err = d.db.Exec(c, _UpdateAppstoreStateSQL, batchToken); err != nil {
		err = errors.Wrap(err, "d.db.Exec()")
		return
	}
	return sqlRes.RowsAffected()
}

// AppstoreMIDIsRecieved  get data from cache if miss will call source method, then add to cache.
func (dao *Dao) AppstoreMIDIsRecieved(ctx context.Context, batchToken string, mid int64) (res int64, err error) {
	addCache := true
	res, err = dao.CacheAppstoreMIDIsRecieved(ctx, batchToken, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("AppstoreMIDIsRecieved")
		return
	}
	cache.MetricMisses.Inc("AppstoreMIDIsRecieved")
	res, err = dao.CountAppstoreReceivedByMID(ctx, batchToken, mid)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	dao.cache.Do(ctx, func(ctx context.Context) {
		dao.AddCacheAppstoreMIDIsRecieved(ctx, batchToken, mid, miss)
	})
	return
}

// AppstoreTelIsRecieved  get data from cache if miss will call source method, then add to cache.
func (dao *Dao) AppstoreTelIsRecieved(ctx context.Context, batchToken string, tel string) (res int64, err error) {
	addCache := true
	res, err = dao.CacheAppstoreTelIsRecieved(ctx, batchToken, tel)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("AppstoreTelIsRecieved")
		return
	}
	cache.MetricMisses.Inc("AppstoreTelIsRecieved")
	res, err = dao.CountAppstoreReceivedByTel(ctx, batchToken, tel)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	dao.cache.Do(ctx, func(ctx context.Context) {
		dao.AddCacheAppstoreTelIsRecieved(ctx, batchToken, tel, miss)
	})
	return
}

// AppstoreIsRecieved  get data from cache if miss will call source method, then add to cache.
func (dao *Dao) AppstoreIsRecieved(ctx context.Context, batchToken string, matchLabel string, matchKind int64) (res int64, err error) {
	addCache := true
	res, err = dao.CacheAppstoreIsRecieved(ctx, batchToken, matchLabel, matchKind)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("AppstoreFingerprintIsRecieved")
		return
	}
	cache.MetricMisses.Inc("AppstoreFingerprintIsRecieved")
	res, err = dao.CountAppstoreReceivedByMatch(ctx, batchToken, matchLabel, matchKind)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	dao.cache.Do(ctx, func(ctx context.Context) {
		dao.AddCacheAppstoreIsRecieved(ctx, batchToken, matchLabel, matchKind, miss)
	})
	return
}
