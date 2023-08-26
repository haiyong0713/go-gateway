package bwsonline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"strings"
)

const (
	_userBindInfoCachePre    = "bwpark:inter:reserve:act:cache:pre"
	_userTicketsListCachePre = "bwpark:user:tickets:list:cache:pre"

	_addTicketBindRecordSQL = "INSERT INTO %s (user_name,mid,personal_id,personal_id_type,personal_id_sum) VALUES(?,?,?,?,?)"
	_ticketsByMidSQL        = "SELECT id , user_name,mid,personal_id,personal_id_type,personal_id_sum , ctime ,mtime FROM %s WHERE mid = ? and state = 1 "
	_bindRecordByIdsSQL     = "SELECT id , user_name,mid,personal_id,personal_id_type,personal_id_sum , ctime ,mtime FROM %s WHERE id >= ? and state = 1 order by  id  asc limit ?"
)

func getBwsTicketTableName(year int) string {
	return fmt.Sprintf("act_bws_online_bind_record_%v", year)
}

func (d *Dao) AddTicketBindRecord(ctx context.Context, uname string, mid int64, pId string, idType int, idSum string, year int) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_addTicketBindRecordSQL, getBwsTicketTableName(year)), uname, mid, pId, idType, idSum)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, ecode.BwsOnlineIdRepeateBind
		}
		return 0, errors.Wrap(err, "AddTicketBindRecord")
	}
	return row.RowsAffected()
}

func (d *Dao) RawTicketsByMid(ctx context.Context, mid int64, year int) (res []*bwsonline.TicketBindRecord, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_ticketsByMidSQL, getBwsTicketTableName(year)), mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageByIDs Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsonline.TicketBindRecord)
		if err = rows.Scan(&r.Id, &r.UserName, &r.Mid, &r.PersonalId, &r.PersonalIdType, &r.PersonalIdSum, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawTicketsByMid Scan")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawTicketsByMid rows")
	}
	return res, nil
}

func (d *Dao) RawTicketsListByIds(ctx context.Context, startId int64, limit int32, year int) (res []*bwsonline.TicketBindRecord, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_bindRecordByIdsSQL, getBwsTicketTableName(year)), startId, limit)
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageByIDs Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsonline.TicketBindRecord)
		if err = rows.Scan(&r.Id, &r.UserName, &r.Mid, &r.PersonalId, &r.PersonalIdType, &r.PersonalIdSum, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawTicketsByMid Scan")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawTicketsByMid rows")
	}
	return res, nil
}

func (d *Dao) AddBindInfoCache(ctx context.Context, mid int64, bindList []*bwsonline.TicketBindRecord, expireTime int32) (err error) {
	cacheKey := buildKey(_userBindInfoCachePre, mid)
	var (
		data []byte
	)
	if data, err = json.Marshal(bindList); err != nil {
		return err
	}
	if expireTime <= 0 {
		expireTime = d.userExpire
	}
	if _, err = d.redis.Do(ctx, "SETEX", cacheKey, expireTime, data); err != nil {
		log.Errorc(ctx, "CacheBindInfo conn.Do(SETEX) key(%s) error(%v)", cacheKey, err)
	}
	return
}

func (d *Dao) GetBindInfoFromCache(ctx context.Context, mid int64) (bindList []*bwsonline.TicketBindRecord, ok bool, err error) {
	var (
		bs       []byte
		cacheKey = buildKey(_userBindInfoCachePre, mid)
	)
	bs, err = redis.Bytes(d.redis.Do(ctx, "GET", cacheKey))
	if err != nil {
		log.Errorc(ctx, "GetBindInfoFromCache conn.Do(GET key(%v)) error(%v)", cacheKey, err)
		return
	}
	if len(bs) > 0 {
		ok = true
	}
	bindList = []*bwsonline.TicketBindRecord{}
	if err = json.Unmarshal(bs, &bindList); err != nil {
		log.Errorc(ctx, "GetBindInfoFromCache json.Unmarshal(%s) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddUserTicketInfosCache(ctx context.Context, pId string, idType int, tickets *bwsonline.TicketInfoFromHYG) (err error) {
	var (
		data     []byte
		cacheKey = buildKey(_userTicketsListCachePre, pId, idType)
	)
	if data, err = json.Marshal(tickets); err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", cacheKey, d.dataExpire, data); err != nil {
		log.Errorc(ctx, "AddUserTicketInfosCache conn.Do(SETEX) key(%s) error(%v)", cacheKey, err)
	}
	return
}

func (d *Dao) GetUserTicketInfosFromCache(ctx context.Context, pId string, idType int) (tickets *bwsonline.TicketInfoFromHYG, err error) {
	var (
		bs       []byte
		cacheKey = buildKey(_userTicketsListCachePre, pId, idType)
	)

	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", cacheKey)); err != nil {
		log.Errorc(ctx, "GetUserTicketInfosFromCache conn.Do(GET key(%v)) error(%v)", cacheKey, err)
		return
	}

	tickets = &bwsonline.TicketInfoFromHYG{}
	if err = json.Unmarshal(bs, tickets); err != nil {
		log.Errorc(ctx, "GetUserTicketInfosFromCache json.Unmarshal(%s) error(%v)", bs, err)
	}
	return
}
