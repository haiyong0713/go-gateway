package bws

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/pkg/errors"
)

const _unUserFinishVoteIDSQL = "SELECT id FROM act_bws_user_vote_log WHERE user_token=? AND point_id=? AND state=? LIMIT 1"

func (d *Dao) RawUserUnFinishVoteID(ctx context.Context, userToken string, pid int64) (int64, error) {
	var id int64
	if err := d.db.QueryRow(ctx, _unUserFinishVoteIDSQL, userToken, pid, bwsmdl.VoteStateInit).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "RawUserUnFinishVoteID:Scan")
	}
	return id, nil
}

func userVoteIDKey(userToken string, pid int64) string {
	return fmt.Sprintf("bws20_user_vote_%s_%d", userToken, pid)
}

func (d *Dao) CacheUserUnFinishVoteID(ctx context.Context, userToken string, pid int64) (int64, error) {
	key := userVoteIDKey(userToken, pid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	data, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, errors.Wrap(err, "CacheUserUnFinishVoteID:redis GET")
	}
	return data, nil
}

func (d *Dao) AddCacheUserUnFinishVoteID(ctx context.Context, userToken string, voteID, pid int64) error {
	key := userVoteIDKey(userToken, pid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("SETEX", key, d.bwsUserExpire, voteID); err != nil {
		return errors.Wrapf(err, "AddUserUnFinishVoteID SETEX key:%s", key)
	}
	return nil
}

func (d *Dao) DelCacheUserUnFinishVoteID(ctx context.Context, userToken string, pid int64) error {
	key := userVoteIDKey(userToken, pid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheUserTasks conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _addVoteSQL = "INSERT INTO act_bws_user_vote_log(user_token,point_id,result,state) VALUES(?,?,?,?)"

func (d *Dao) AddVoteLog(ctx context.Context, userToken string, pid, result int64) (int64, error) {
	row, err := d.db.Exec(ctx, _addVoteSQL, userToken, pid, result, bwsmdl.VoteStateInit)
	if err != nil {
		return 0, errors.Wrap(err, "AddVote")
	}
	return row.LastInsertId()
}

const _unFinishVoteLogSQL = "SELECT id,user_token,point_id,result FROM act_bws_user_vote_log WHERE point_id=? AND state=? ORDER BY id"

func (d *Dao) UnFinishVoteLog(ctx context.Context, pointID int64) ([]*bwsmdl.VoteLog, error) {
	rows, err := d.db.Query(ctx, _unFinishVoteLogSQL, pointID, bwsmdl.VoteStateInit)
	if err != nil {
		return nil, errors.Wrap(err, "UnFinishVoteLog Query")
	}
	defer rows.Close()
	var data []*bwsmdl.VoteLog
	for rows.Next() {
		r := new(bwsmdl.VoteLog)
		if err = rows.Scan(&r.ID, &r.UserToken, &r.PointID, &r.Result); err != nil {
			return nil, errors.Wrap(err, "UnFinishVoteLog Scan")
		}
		data = append(data, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "UnFinishVoteLog rows")
	}
	return data, nil
}

const _finishVoteLog = "UPDATE act_bws_user_vote_log SET state=? WHERE state=?"

func (d *Dao) FinishVoteLog(ctx context.Context) (int64, error) {
	row, err := d.db.Exec(ctx, _finishVoteLog, bwsmdl.VoteStateFinish, bwsmdl.VoteStateInit)
	if err != nil {
		return 0, errors.Wrap(err, "FinishVoteLog")
	}
	return row.RowsAffected()
}
