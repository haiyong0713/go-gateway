package pay

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-main/app/ep/merlin/ecode"
)

const (
	_selectPayOrderSQL               = "SELECT `mid`,`order_desc`,`money`,`order_id`,`pay_order_id`,`order_time`,`pay_time`,`order_status`,`pay_status`,`ctime`,`mtime` FROM wx_pay_order WHERE order_id=?"
	_insertPayOrderSQL               = "INSERT INTO wx_pay_order (`mid`,`order_desc`,`money`,`order_id`) VALUES(?,?,?,?)"
	_updatePayOrderSQL               = "UPDATE wx_pay_order SET pay_order_id=?,order_time=?,order_status=? WHERE mid=? AND order_id=?"
	_updatePayOrderPayStatusSQL      = "UPDATE wx_pay_order SET pay_status=? WHERE mid=? AND order_id=? AND pay_order_id=?"
	_updateWxLotteryLogPayOrderIDSQL = "UPDATE wx_lottery_log_%02d SET order_id=?,pay_order_id=?,order_time=?,order_status=? WHERE mid=? AND id=?"
	_updateWxLotteryLogPayStatusSQL  = "UPDATE wx_lottery_log_%02d SET pay_status=? WHERE mid=? AND id=?"
	_userWxLotteryLogPageSQL         = "SELECT `id`,`mid`,`lottery_id`,`gift_type`,`gift_id`,`gift_name`,`gift_money`,`order_id`,`pay_order_id`,`order_time`,`order_status`,`pay_status`,`ctime`,`mtime` FROM wx_lottery_log_%02d WHERE id > ? ORDER BY ID limit ?"
)

// PayOrderByOrderID .
func (d *Dao) PayOrderByOrderID(ctx context.Context, orderID string) (*like.PayOrder, error) {
	row := d.db.QueryRow(ctx, _selectPayOrderSQL, orderID)
	ir := &like.PayOrder{}
	if err := row.Scan(&ir.Mid, &ir.OrderDesc, &ir.Money, &ir.OrderID, &ir.PayOrderID, &ir.OrderTime, &ir.PayTime, &ir.OrderStatus, &ir.PayStatus, &ir.CTime, &ir.MTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		log.Error("d.PayOrderByOrderID orderID:%s error(%+v)", orderID, err)
		return nil, err
	}
	return ir, nil
}

// InsertPayOrder .
func (d *Dao) InsertPayOrder(ctx context.Context, mid int64, orderDesc string, money int64, orderID string) (int64, error) {
	res, err := d.db.Exec(ctx, _insertPayOrderSQL, mid, orderDesc, money, orderID)
	if err != nil {
		log.Error("d.InsertPayOrder(%d,%s,%d,%s) error(%+v)", mid, orderDesc, money, orderID, err)
		return 0, err
	}
	return res.LastInsertId()
}

// UpdatePayOrder .
func (d *Dao) UpdatePayOrder(ctx context.Context, payOrderID string, orderTime int64, orderStatus, mid int64, orderID string) (int64, error) {
	res, err := d.db.Exec(ctx, _updatePayOrderSQL, payOrderID, orderTime, orderStatus, mid, orderID)
	if err != nil {
		log.Error("d.UpdatePayOrder(%s,%d,%d,%d,%s) error(%v)", payOrderID, orderTime, orderStatus, mid, orderID, err)
		return 0, err
	}
	return res.RowsAffected()
}

// PayTransfer 活动红包转入 .
func (d *Dao) PayTransferInner(ctx context.Context, pt *PayTransferInner) (res *ResultInner, err error) {
	payConfig := d.c.WxLottery.PayCfg
	if payConfig == nil {
		log.Error("payConfig error struct:%+v", pt)
		return &ResultInner{}, ecode.NothingFound
	}
	cfg := &PayConfig{
		CustomerID:   payConfig.CustomerID,
		MerchantCode: payConfig.MerchantCode,
		CoinType:     payConfig.CoinType,
		PayHost:      payConfig.PayHost,
		Token:        payConfig.Token,
		ActivityID:   payConfig.ActivityID,
	}
	return d.pay.PayTransferInner(ctx, cfg, pt)
}

// ProfitCancel 活动红包撤回 .
func (d *Dao) ProfitCancelInner(ctx context.Context, pt *ProfitCancelInner) (res *ResultInner, err error) {
	payConfig := d.c.WxLottery.PayCfg
	if payConfig == nil {
		log.Error("payConfig error struct:%+v", pt)
		return &ResultInner{}, ecode.NothingFound
	}
	cfg := &PayConfig{
		CustomerID:   payConfig.CustomerID,
		MerchantCode: payConfig.MerchantCode,
		CoinType:     payConfig.CoinType,
		PayHost:      payConfig.PayHost,
		Token:        payConfig.Token,
		ActivityID:   payConfig.ActivityID,
	}
	return d.pay.ProfitCancelInner(ctx, cfg, pt)
}

// UpdateWxLotteryLogPayOrderID .
func (d *Dao) UpdateWxLotteryLogPayOrderID(ctx context.Context, orderID, payOrderID string, orderTime int64, orderStatus, mid int64, ID int64) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_updateWxLotteryLogPayOrderIDSQL, mid%100), orderID, payOrderID, orderTime, orderStatus, mid, ID)
	if err != nil {
		log.Error("d.UpdateWxLotteryLogPayOrderID(%s,%s,%d,%d,%d,%d) error(%v)", orderID, payOrderID, orderTime, orderStatus, mid, ID, err)
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateWxLotteryLogPayStatus .
func (d *Dao) UpdateWxLotteryLogPayStatus(ctx context.Context, payStatus int32, mid int64, id int64) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_updateWxLotteryLogPayStatusSQL, mid%100), payStatus, mid, id)
	if err != nil {
		log.Error("d.UpdateUserTaskPayStatus(%d,%d,%d) error(%+v)", payStatus, mid, id, err)
		return 0, err
	}
	return res.RowsAffected()
}

// UpdatePayOrderPayStatus .
func (d *Dao) UpdatePayOrderPayStatus(ctx context.Context, payStatus int32, mid int64, orderID string, payOrderID string) (int64, error) {
	res, err := d.db.Exec(ctx, _updatePayOrderPayStatusSQL, payStatus, mid, orderID, payOrderID)
	if err != nil {
		log.Error("d.UpdatePayOrderPayStatus(%d,%d,%s,%s) error(%v)", payStatus, mid, orderID, payOrderID, err)
		return 0, err
	}
	return res.RowsAffected()
}

// WxLotteryLogPage .
func (d *Dao) WxLotteryLogPage(ctx context.Context, region, index, limit int64) (res []*like.WxLotteryLog, lastID int64, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_userWxLotteryLogPageSQL, region), index, limit)
	if err != nil {
		log.Error("WxLotteryLogPage %d,%d error:%+v", index, limit, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := new(like.WxLotteryLog)
		if err = rows.Scan(&c.ID, &c.Mid, &c.LotteryID, &c.GiftType, &c.GiftID, &c.GiftName, &c.GiftMoney, &c.OrderID, &c.PayOrderID, &c.OrderTime, &c.OrderStatus, &c.PayStatus, &c.Ctime, &c.Mtime); err != nil {
			log.Error("d.WxLotteryLogPage(%d,%d) error(%v)", index, limit, err)
			return
		}
		res = append(res, c)
		if c.ID > lastID {
			lastID = c.ID
		}
	}
	err = rows.Err()
	return
}
