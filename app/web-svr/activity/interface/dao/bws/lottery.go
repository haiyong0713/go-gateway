package bws

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/pkg/errors"
)

const _lotteryAwardSQL = "SELECT id,title,image,intro,cate,owner_mid,stage,ctime,mtime FROM act_bws_award WHERE id=? AND state=1"

func (d *Dao) RawAward(ctx context.Context, awardID int64) (*bwsmdl.Award, error) {
	data := new(bwsmdl.Award)
	row := d.db.QueryRow(ctx, _lotteryAwardSQL, awardID)
	if err := row.Scan(&data.ID, &data.Title, &data.Image, &data.Intro, &data.Cate, &data.Owner, &data.Stage, &data.Ctime, &data.Mtime); err != nil {
		return nil, errors.Wrap(err, "RawAward:QueryRow")
	}
	return data, nil
}

const _lotteryAwardStockSQL = "SELECT stock FROM act_bws_award WHERE id=? AND state=1"

func (d *Dao) RawAwardStock(ctx context.Context, awardID int64) (int64, error) {
	var stock int64
	if err := d.db.QueryRow(ctx, _lotteryAwardStockSQL, awardID).Scan(&stock); err != nil {
		return 0, errors.Wrap(err, "RawAwardStock:QueryRow")
	}
	return stock, nil
}

const _awardsListSQL = "SELECT id,title,image,intro,cate,is_online,owner_mid,stage,stock,ctime,mtime FROM act_bws_award WHERE state=1"

func (d *Dao) RawAwardList(ctx context.Context) (map[int64]*bwsmdl.Award, error) {
	rows, err := d.db.Query(ctx, _awardsListSQL)
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardList Query")
	}
	defer rows.Close()
	data := make(map[int64]*bwsmdl.Award)
	for rows.Next() {
		r := new(bwsmdl.Award)
		if err = rows.Scan(&r.ID, &r.Title, &r.Image, &r.Intro, &r.Cate, &r.IsOnline, &r.Owner, &r.Stage, &r.Stock, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawAwardList Scan")
		}
		data[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawAwardList rows")
	}
	return data, nil
}

const _onlineAwardIDsSQL = "SELECT id FROM act_bws_award WHERE state=1 AND is_online=0 AND stock!=0"

func (d *Dao) RawOnlineAwardIDs(ctx context.Context) ([]int64, error) {
	rows, err := d.db.Query(ctx, _onlineAwardIDsSQL)
	if err != nil {
		return nil, errors.Wrap(err, "RawOnlineAwardIDs Query")
	}
	defer rows.Close()
	var data []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "RawOnlineAwardIDs Scan")
		}
		data = append(data, id)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawOnlineAwardIDs rows")
	}
	return data, nil
}

const _decrAwardStock = "UPDATE act_bws_award SET stock = stock-1 WHERE id=? AND stock>0"

func (d *Dao) DecrAwardStock(ctx context.Context, awardID int64) (int64, error) {
	row, err := d.db.Exec(ctx, _decrAwardStock, awardID)
	if err != nil {
		return 0, errors.Wrap(err, "DecrAwardStock")
	}
	return row.RowsAffected()
}

const _userAwardSQL = "SELECT id,user_token,award_id,state,ctime,mtime FROM act_bws_user_award WHERE user_token=? ORDER BY id DESC"

func (d *Dao) RawUserAward(ctx context.Context, userToken string) ([]*bwsmdl.UserAward, error) {
	rows, err := d.db.Query(ctx, _userAwardSQL, userToken)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserAward Query")
	}
	defer rows.Close()
	var data []*bwsmdl.UserAward
	for rows.Next() {
		r := new(bwsmdl.UserAward)
		if err = rows.Scan(&r.ID, &r.UserToken, &r.AwardId, &r.State, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawUserAward Scan")
		}
		data = append(data, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserAward rows")
	}
	return data, nil
}

func userAwardKey(userToken string) string {
	return fmt.Sprintf("bws20_user_award_%s", userToken)
}

func (d *Dao) CacheUserAward(ctx context.Context, userToken string) ([]*bwsmdl.UserAward, error) {
	key := userAwardKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bytes, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "CacheUserAward:redis GET")
	}
	var data []*bwsmdl.UserAward
	if err = json.Unmarshal(bytes, &data); err != nil {
		return nil, errors.Wrap(err, "CacheUserAward json.Unmarshal")
	}
	return data, nil
}

func (d *Dao) AddCacheUserAward(ctx context.Context, userToken string, data []*bwsmdl.UserAward) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "AddCacheUserAward json.Marshal")
	}
	key := userAwardKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, d.bwsUserExpire, bytes); err != nil {
		return errors.Wrapf(err, "AddCacheUserAward SETEX key:%s", key)
	}
	return nil
}

func (d *Dao) DelCacheUserAward(ctx context.Context, userToken string) error {
	key := userAwardKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheUserAward conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _addUserAwardSQL = "INSERT INTO act_bws_user_award(user_token,award_id,state) VALUES (?,?,?)"

func (d *Dao) AddUserAward(ctx context.Context, userToken string, awardID int64, state string) (int64, error) {
	row, err := d.db.Exec(ctx, _addUserAwardSQL, userToken, awardID, state)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserAward")
	}
	return row.LastInsertId()
}

const _upUserAwardSQL = "UPDATE act_bws_user_award SET state=? WHERE user_token=? AND id=? AND state=?"

func (d *Dao) UpUserAward(ctx context.Context, userAwardID int64, userToken, state, preState string) (int64, error) {
	row, err := d.db.Exec(ctx, _upUserAwardSQL, state, userToken, userAwardID, preState)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserAward")
	}
	return row.RowsAffected()
}

const _lotteryTimesSQL = "SELECT count(1) FROM act_bws_user_lottery_times WHERE user_token=? AND state=?"

func (d *Dao) RawUserLotteryTimes(ctx context.Context, userToken string) (int64, error) {
	var count int64
	if err := d.db.QueryRow(ctx, _lotteryTimesSQL, userToken, bwsmdl.LotteryTimesStateInit).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "RawUserLotteryTimes:Scan")
	}
	return count, nil
}

func userLotteryKey(userToken string) string {
	return fmt.Sprintf("bws20_user_lott_%s", userToken)
}

func (d *Dao) CacheUserLotteryTimes(ctx context.Context, userToken string) (int64, error) {
	key := userLotteryKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	data, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, errors.Wrap(err, "CacheUserLotteryTimes:redis GET")
	}
	return data, nil
}

func (d *Dao) AddCacheUserLotteryTimes(ctx context.Context, userToken string, lotteryTimes int64) error {
	key := userLotteryKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("SETEX", key, d.bwsUserExpire, lotteryTimes); err != nil {
		return errors.Wrapf(err, "AddUserLotteryTimes SETEX key:%s", key)
	}
	return nil
}

func (d *Dao) DelCacheUserLotteryTimes(ctx context.Context, userToken string) error {
	key := userLotteryKey(userToken)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheUserLotteryTimes conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

// const _useLotteryTimes = "UPDATE act_bws_user_lottery_times SET state=? WHERE user_token=? AND state=? LIMIT 1"

// func (d *Dao) UseLotteryTimes(ctx context.Context, userToken string) (int64, error) {
// 	row, err := d.db.Exec(ctx, _useLotteryTimes, bwsmdl.LotteryTimesStateFinish, userToken, bwsmdl.LotteryTimesStateInit)
// 	if err != nil {
// 		return 0, errors.Wrap(err, "UseLotteryTimes")
// 	}
// 	return row.RowsAffected()
// }

const _addLotteryTimesSQL = "INSERT INTO act_bws_user_lottery_times(user_token,task_id,state) VALUES (?,?,?)"

func (d *Dao) AddLotteryTimes(ctx context.Context, userToken string, taskID int64) (int64, error) {
	row, err := d.db.Exec(ctx, _addLotteryTimesSQL, userToken, taskID, bwsmdl.LotteryTimesStateInit)
	if err != nil {
		return 0, errors.Wrap(err, "AddLotteryTimes")
	}
	return row.LastInsertId()
}

const _useLotteryTimes = "INSERT INTO act_bws_user_lottery_times(user_token,task_id,state) VALUES (?,?,?)"

// UseLotteryTimes 使用抽奖次数
func (d *Dao) UseLotteryTimes(ctx context.Context, userToken string, taskID int64) (int64, error) {
	row, err := d.db.Exec(ctx, _useLotteryTimes, userToken, taskID, bwsmdl.LotteryTimesStateFinish)
	if err != nil {
		return 0, errors.Wrap(err, "AddLotteryTimes")
	}
	return row.LastInsertId()
}
