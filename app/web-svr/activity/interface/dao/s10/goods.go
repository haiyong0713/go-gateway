package s10

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/s10"
)

const _allGoodsSQL = "select id,robin,gname,figture,score,rank,send,stock,category,exchange_times,round_stock,round_exchange_times,start_time,end_time,extra,`desc` from act_s10_goods where id>0;"

func (d *Dao) AllGoods(ctx context.Context) (map[int32][]*s10.Bonus, error) {
	rows, err := component.S10GlobalDB.Query(ctx, _allGoodsSQL)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AllGoods() error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int32][]*s10.Bonus, 5)
	for rows.Next() {
		tmp := new(s10.Bonus)
		if err = rows.Scan(&tmp.ID, &tmp.Robin, &tmp.Name, &tmp.Figure, &tmp.Score, &tmp.Rank, &tmp.Send, &tmp.Stock, &tmp.Type,
			&tmp.ExchangeTimes, &tmp.RoundStock, &tmp.RoundExchangeTimes, &tmp.Start, &tmp.End, &tmp.Extra, &tmp.Desc); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res[tmp.Robin] = append(res[tmp.Robin], tmp)
	}
	return res, rows.Err()
}

const _updateGoodsSendCountSQL = "update act_s10_goods set send=send+1 where id=? and send<stock;"

func (d *Dao) UpdateGoodsSendCount(ctx context.Context, gid int32) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _updateGoodsSendCountSQL, gid)
	if err != nil {
		log.Errorc(ctx, "s10 d.da.UpdateGoodsSendCount(gid:%d) error(%v)", gid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _correctGoodsSendCountSQL = "update act_s10_goods set send=send-1 where id=?;"

func (d *Dao) CorrectGoodsSendCount(ctx context.Context, gid int32) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _correctGoodsSendCountSQL, gid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.CorrectGoodsSendCount(gid:%d) error(%v)", gid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _addGoodsRoundSendCountSQL = "insert into act_s10_round_robin_goods(gid,stock,rtime) values (?,?,?);"

func (d *Dao) AddGoodsRoundSendCount(ctx context.Context, gid, stock int32, rtime xtime.Time) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _addGoodsRoundSendCountSQL, gid, stock, rtime)
	if err != nil {
		log.Errorc(ctx, "s10 d.da.UpdateGoodsRoundSendCount(gid:%d) error(%v)", gid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _updateGoodsRoundSendCountSQL = "update act_s10_round_robin_goods set send=send+1 where rtime=? and gid=? and send<stock;"

func (d *Dao) UpdateGoodsRoundSendCount(ctx context.Context, gid int32, rtime xtime.Time) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _updateGoodsRoundSendCountSQL, rtime, gid)
	if err != nil {
		log.Errorc(ctx, "s10d.dao.UpdateGoodsRoundSendCount(gid:%d) error(%v)", gid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _correctGoodsRoundSendCountSQL = "update act_s10_round_robin_goods set send=send-1 where rtime=? and gid=?;"

func (d *Dao) CorrectGoodsRoundSendCount(ctx context.Context, gid int32, rtime xtime.Time) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _correctGoodsRoundSendCountSQL, rtime, gid)
	if err != nil {
		log.Errorc(ctx, "s10 d.da.CorrectGoodsRoundSendCount(gid:%d) error(%v)", gid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _allRobinGoodsSQL = "select gid,send from act_s10_round_robin_goods where rtime=?;"

func (d *Dao) AllRobinGoods(ctx context.Context, rtime xtime.Time) (map[int32]int32, error) {
	rows, err := component.S10GlobalDB.Query(ctx, _allRobinGoodsSQL, rtime)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AllRobinGoods() error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int32]int32, 10)
	for rows.Next() {
		var gid, send int32
		if err = rows.Scan(&gid, &send); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res[gid] = send
	}
	return res, rows.Err()
}

const _goodsSQL = "select id,gname,type from act_s10_goods where id in(%s);"

func (d *Dao) Goods(ctx context.Context, gids []int64) (map[int32]*s10.Good, error) {
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_goodsSQL, xstr.JoinInts(gids)))
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.Goods(gids:%v) error(%v)", gids, err)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int32]*s10.Good, len(gids))
	for rows.Next() {
		tmp := new(s10.Good)
		if err = rows.Scan(&tmp.Gid, &tmp.Name, &tmp.Type); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res[tmp.Gid] = tmp
	}
	return res, rows.Err()
}

const _goodsByIDSQL = "select stock,send,round_stock from act_s10_goods where id=?;"

func (d *Dao) GoodsByID(ctx context.Context, gid int32) (stock, send, roundStock int32, err error) {
	row := component.S10GlobalDB.Master().QueryRow(ctx, _goodsByIDSQL, gid)
	if err = row.Scan(&stock, &send, &roundStock); err != nil {
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
	}
	return
}

const _roundgoodsByIDSQL = "select send,stock from act_s10_round_robin_goods where  rtime=? and gid=?;"

func (d *Dao) RoundGoodsByID(ctx context.Context, gid int32, rtime xtime.Time) (send, stock int32, exist bool, err error) {
	row := component.S10GlobalDB.Master().QueryRow(ctx, _roundgoodsByIDSQL, rtime, gid)
	if err = row.Scan(&send, &stock); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
		return 0, 0, false, err
	}
	return send, stock, true, nil
}
