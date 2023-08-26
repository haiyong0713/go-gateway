package bwsonline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"

	"github.com/pkg/errors"
)

const _pieceSQL = "SELECT id,title,token,ctime,mtime FROM bws_online_piece WHERE id=? AND state = 1"

func (d *Dao) RawPiece(ctx context.Context, id int64) (*bwsonline.Piece, error) {
	data := new(bwsonline.Piece)
	row := d.db.QueryRow(ctx, _pieceSQL, id)
	if err := row.Scan(&data.ID, &data.Title, &data.Token, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawPiece:QueryRow")
	}
	return data, nil
}

func pieceKey(id int64) string {
	return fmt.Sprintf("bws_piece_%d", id)
}

func (d *Dao) CachePiece(ctx context.Context, id int64) (*bwsonline.Piece, error) {
	key := pieceKey(id)
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		return nil, err
	}
	res := new(bwsonline.Piece)
	if err = res.Unmarshal(data); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCachePiece(ctx context.Context, id int64, data *bwsonline.Piece) error {
	key := pieceKey(id)
	bytes, err := data.Marshal()
	if err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.dataExpire, bytes); err != nil {
		return err
	}
	return nil
}

const _pieceSendSQL = "INSERT INTO bws_online_piece_add_log (mid,pid,num,bid) VALUES(?,?,?,?)"

func (d *Dao) PieceAddLog(ctx context.Context, mid, pid, num, bid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _pieceSendSQL, mid, pid, num, bid)
	if err != nil {
		return 0, errors.Wrap(err, "PieceAddLog")
	}
	return row.LastInsertId()
}

const _pieceAddUseLog = "INSERT INTO bws_online_piece_use_log (mid,piece_id,num,batch_id) VALUES %s"

func (d *Dao) PieceAddUseLog(ctx context.Context, mid int64, list []*bwsonline.UserPiece, batchID string) (int64, error) {
	var (
		rowStrs []string
		args    []interface{}
	)
	for _, v := range list {
		rowStrs = append(rowStrs, "(?,?,?,?)")
		args = append(args, mid, v.Pid, v.Num, batchID)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_pieceAddUseLog, strings.Join(rowStrs, ",")), args...)
	if err != nil {
		return 0, errors.Wrap(err, "PieceAddUseLog")
	}
	return row.RowsAffected()
}

const _addUserPieceSQL = "INSERT INTO bws_online_user_piece(mid,pid,num,bid) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE num=num+?"

func (d *Dao) AddUserPiece(ctx context.Context, mid, pid, num, bid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _addUserPieceSQL, mid, pid, num, bid, num)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserPiece")
	}
	return row.RowsAffected()
}

const _decrUserPieceSQL = "UPDATE bws_online_user_piece SET num=num - CASE %s END WHERE mid=? AND bid=? AND pid IN (%s)"

func (d *Dao) DecrUserPiece(ctx context.Context, mid int64, list []*bwsonline.UserPiece, bid int64) (int64, error) {
	var (
		caseStr string
		pids    []int64
		params  []interface{}
	)
	for _, v := range list {
		caseStr = fmt.Sprintf("%s WHEN pid=? THEN ?", caseStr)
		params = append(params, v.Pid, v.Num)
		pids = append(pids, v.Pid)
	}
	params = append(params, mid, bid)
	row, err := d.db.Exec(ctx, fmt.Sprintf(_decrUserPieceSQL, caseStr, xstr.JoinInts(pids)), params...)
	if err != nil {
		err = errors.Wrap(err, "DecrUserPiece:db.Exec error")
		return 0, err
	}
	return row.RowsAffected()
}

const _usePieceSQL = "SELECT pid,num FROM bws_online_user_piece WHERE mid=? AND bid=?"

func (d *Dao) RawUserPiece(ctx context.Context, mid, bid int64) ([]*bwsonline.UserPiece, error) {
	rows, err := d.db.Query(ctx, _usePieceSQL, mid, bid)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserPiece Query")
	}
	defer rows.Close()
	var list []*bwsonline.UserPiece
	for rows.Next() {
		r := new(bwsonline.UserPiece)
		if err = rows.Scan(&r.Pid, &r.Num); err != nil {
			return nil, errors.Wrap(err, "RawUserPiece Scan")
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserPiece rows")
	}
	return list, nil
}

func userPieceKey(mid, bid int64) string {
	return fmt.Sprintf("bws_user_piece_all_%d_%d", mid, bid)
}

func (d *Dao) CacheUserPiece(ctx context.Context, mid, bid int64) ([]*bwsonline.UserPiece, error) {
	key := userPieceKey(mid, bid)
	values, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", key, 0, -1))
	if err != nil {
		log.Errorc(ctx, "CacheUserPiece conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	var res []*bwsonline.UserPiece
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Errorc(ctx, "CacheUserAward redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		item := &bwsonline.UserPiece{}
		if err = json.Unmarshal(bs, item); err != nil {
			log.Errorc(ctx, "CacheUserPiece json.Unmarshal(%v) error(%v)", bs, err)
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (d *Dao) AddCacheUserPiece(ctx context.Context, mid int64, data []*bwsonline.UserPiece, bid int64) error {
	key := userPieceKey(mid, bid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	err := conn.Send("DEL", key)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserPiece conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for _, v := range data {
		var bs []byte
		bs, err = json.Marshal(v)
		if err != nil {
			log.Errorc(ctx, "AddCacheUserPiece json.Marshal() error(%v)", err)
			return err
		}
		args = args.Add(v.Pid).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(ctx, "AddCacheUserPiece conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Errorc(ctx, "AddCacheUserPrint conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheUserPiece conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheUserPiece conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) DelCacheUserPiece(ctx context.Context, mid, bid int64) error {
	key := userPieceKey(mid, bid)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserPiece conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _usedTimesSQL = "SELECT type_id FROM bws_online_used_times WHERE mid=? AND use_day=?"

func (d *Dao) UsedTimes(ctx context.Context, mid, useDay int64) (map[int64]int64, error) {
	rows, err := d.db.Query(ctx, _usedTimesSQL, mid, useDay)
	if err != nil {
		return nil, errors.Wrap(err, "RawUsedTimes Query")
	}
	defer rows.Close()
	res := make(map[int64]int64)
	for rows.Next() {
		var typeID int64
		if err = rows.Scan(&typeID); err != nil {
			return nil, errors.Wrap(err, "RawUsedTimes Scan")
		}
		res[typeID]++
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUsedTimes rows")
	}
	return res, nil
}

const _usedTimesAddSQL = "INSERT INTO bws_online_used_times (mid,type_id,use_day) VALUES (?,?,?)"

func (d *Dao) AddUsedTimes(ctx context.Context, mid, typ, useDay int64) (int64, error) {
	row, err := d.db.Exec(ctx, _usedTimesAddSQL, mid, typ, useDay)
	if err != nil {
		return 0, errors.Wrap(err, "AddUsedTimes:db.Exec error")
	}
	return row.RowsAffected()
}

const _pieceUsedLogSQL = "SELECT batch_id,piece_id,num FROM bws_online_piece_use_log WHERE mid=? AND batch_id IN(%s)"

func (d *Dao) RawPieceUsedLog(ctx context.Context, mid int64, batchIDs []string) (map[string]map[int64]int64, error) {
	if len(batchIDs) == 0 {
		return nil, nil
	}
	var (
		sqlPlaces []string
		sqlParam  []interface{}
	)
	sqlParam = append(sqlParam, mid)
	for _, v := range batchIDs {
		sqlPlaces = append(sqlPlaces, "?")
		sqlParam = append(sqlParam, v)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_pieceUsedLogSQL, strings.Join(sqlPlaces, ",")), sqlParam...)
	if err != nil {
		return nil, errors.Wrap(err, "RawPieceUsedLog Query")
	}
	defer rows.Close()
	res := make(map[string]map[int64]int64)
	for rows.Next() {
		var (
			pieceID, num int64
			batchID      string
		)
		if err = rows.Scan(&batchID, &pieceID, &num); err != nil {
			return nil, errors.Wrap(err, "RawPieceUsedLog Scan")
		}
		if _, ok := res[batchID]; !ok {
			res[batchID] = make(map[int64]int64)
		}
		res[batchID][pieceID] = num
	}
	return res, rows.Err()
}

func (d *Dao) CachePieceUsedLog(ctx context.Context, mid int64, batchIDs []string) (map[string]map[int64]int64, error) {
	return nil, nil
}

func (d *Dao) AddCachePieceUsedLog(ctx context.Context, mid int64, data map[string]map[int64]int64, batchIDs []string) error {
	return nil
}
