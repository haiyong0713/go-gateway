package telecom

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/cache/memcache"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/model/telecom"
)

const (
	_inOrderSyncSQL = `INSERT IGNORE INTO telecom_order (request_no,result_type,flowpackageid,flowpackagesize,flowpackagetype,trafficattribution,begintime,endtime,
		ismultiplyorder,settlementtype,operator,order_status,remainedrebindnum,maxbindnum,orderid,sign_no,accesstoken,phoneid,isrepeatorder,paystatus,
		paytime,paychannel,sign_status,refund_status) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE
		request_no=?,result_type=?,flowpackagesize=?,flowpackagetype=?,trafficattribution=?,begintime=?,endtime=?,
		ismultiplyorder=?,settlementtype=?,operator=?,order_status=?,remainedrebindnum=?,maxbindnum=?,orderid=?,sign_no=?,accesstoken=?,
		isrepeatorder=?,paystatus=?,paytime=?,paychannel=?,sign_status=?,refund_status=?`
	_inRechargeSyncSQL = `INSERT INTO telecom_recharge (request_no,fcrecharge_no,recharge_status,ordertotalsize,flowbalance) VALUES (?,?,?,?,?)`
	_orderByPhoneSQL   = `SELECT phoneid,orderid,order_status,sign_no,isrepeatorder,begintime,endtime FROM telecom_order WHERE phoneid=?`
	_orderByOrderIDSQL = `SELECT phoneid,orderid,order_status,sign_no,isrepeatorder,begintime,endtime FROM telecom_order WHERE orderid=?`
	// card
	_inCardOrderSyncSQL = `INSERT IGNORE INTO telecom_card_order (phone,nbr,starttime,endtime,action,createdate,appkey) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE
	nbr=?,starttime=?,endtime=?,action=?,createdate=?,appkey=?`
	_cardOrderSyncSQL = `SELECT phone,nbr,starttime,endtime,action,appkey FROM telecom_card_order WHERE phone=?`
	_vipPackLogSQL    = `SELECT phone,state,request_no,ptype,ctime FROM telecom_card_pack_log WHERE mtime>=? AND mtime<?`
)

type Dao struct {
	c                    *conf.Config
	client               *httpx.Client
	payInfoURL           string
	cancelRepeatOrderURL string
	sucOrderListURL      string
	telecomReturnURL     string
	telecomCancelPayURL  string
	phoneAreaURL         string
	orderStateURL        string
	smsSendURL           string
	activeStateURL       string
	// card
	phoneAuthURL string
	// card
	phoneKeyExpired   int32
	payKeyExpired     int32
	db                *xsql.DB
	inOrderSyncSQL    *xsql.Stmt
	inRechargeSyncSQL *xsql.Stmt
	orderByPhoneSQL   *xsql.Stmt
	orderByOrderIDSQL *xsql.Stmt
	phoneRds          credis.Redis
	// memcache
	mc     *memcache.Memcache
	expire int32
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                    c,
		client:               httpx.NewClient(c.HTTPTelecom),
		payInfoURL:           c.Host.Telecom + _payInfo,
		cancelRepeatOrderURL: c.Host.Telecom + _cancelRepeatOrder,
		sucOrderListURL:      c.Host.Telecom + _sucOrderList,
		phoneAreaURL:         c.Host.Telecom + _phoneArea,
		orderStateURL:        c.Host.Telecom + _orderState,
		telecomReturnURL:     c.Host.TelecomReturnURL,
		telecomCancelPayURL:  c.Host.TelecomCancelPayURL,
		activeStateURL:       c.Host.TelecomActive + _acviteState,
		// card
		phoneAuthURL: c.Host.TelecomCard + _phoneAuth,
		// card
		smsSendURL: c.Host.Sms + _smsSendURL,
		db:         xsql.NewMySQL(c.MySQL.Show),
		phoneRds:   credis.NewRedis(c.Redis.Recommend.Config),
		//reids
		phoneKeyExpired: int32(time.Duration(c.Telecom.KeyExpired) / time.Second),
		payKeyExpired:   int32(time.Duration(c.Telecom.PayKeyExpired) / time.Second),
		// memcache
		mc:     memcache.New(c.Memcache.Operator.Config),
		expire: int32(time.Duration(c.Memcache.Operator.Expire) / time.Second),
	}
	d.inOrderSyncSQL = d.db.Prepared(_inOrderSyncSQL)
	d.inRechargeSyncSQL = d.db.Prepared(_inRechargeSyncSQL)
	d.orderByPhoneSQL = d.db.Prepared(_orderByPhoneSQL)
	d.orderByOrderIDSQL = d.db.Prepared(_orderByOrderIDSQL)
	return
}

// InOrderSync
func (d *Dao) InOrderSync(ctx context.Context, requestNo, resultType int, phone string, t *telecom.TelecomJSON) (row int64, err error) {
	res, err := d.inOrderSyncSQL.Exec(ctx, requestNo, resultType, t.FlowpackageID, t.FlowPackageSize, t.FlowPackageType, t.TrafficAttribution, t.BeginTime, t.EndTime,
		t.IsMultiplyOrder, t.SettlementType, t.Operator, t.OrderStatus, t.RemainedRebindNum, t.MaxbindNum, t.OrderID, t.SignNo, t.AccessToken,
		phone, t.IsRepeatOrder, t.PayStatus, t.PayTime, t.PayChannel, t.SignStatus, t.RefundStatus,
		requestNo, resultType, t.FlowPackageSize, t.FlowPackageType, t.TrafficAttribution, t.BeginTime, t.EndTime,
		t.IsMultiplyOrder, t.SettlementType, t.Operator, t.OrderStatus, t.RemainedRebindNum, t.MaxbindNum, t.OrderID,
		t.SignNo, t.AccessToken, t.IsRepeatOrder, t.PayStatus, t.PayTime, t.PayChannel, t.SignStatus, t.RefundStatus)
	if err != nil {
		log.Error("d.inOrderSyncSQL.Exec error(%v)", err)
		return
	}
	tmp := &telecom.OrderInfo{}
	tmp.OrderInfoJSONChange(t)
	phoneInt, _ := strconv.Atoi(t.PhoneID)
	if err = d.AddTelecomCache(ctx, phoneInt, tmp); err != nil {
		log.Error("s.AddTelecomCache error(%v)", err)
	}
	orderID, _ := strconv.ParseInt(t.OrderID, 10, 64)
	if err = d.AddTelecomOrderIDCache(ctx, orderID, tmp); err != nil {
		log.Error("s.AddTelecomOrderIDCache error(%v)", err)
	}
	return res.RowsAffected()
}

// InRechargeSync
func (d *Dao) InRechargeSync(ctx context.Context, r *telecom.RechargeJSON) (row int64, err error) {
	res, err := d.inRechargeSyncSQL.Exec(ctx, r.RequestNo, r.FcRechargeNo, r.RechargeStatus, r.OrderTotalSize, r.FlowBalance)
	if err != nil {
		log.Error("d.inRechargeSyncSQL.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) OrdersUserFlow(ctx context.Context, phoneID int) (res map[int]*telecom.OrderInfo, err error) {
	res = map[int]*telecom.OrderInfo{}
	var (
		PhoneIDStr string
		OrderIDStr string
	)
	t := &telecom.OrderInfo{}
	row := d.orderByPhoneSQL.QueryRow(ctx, phoneID)
	if err = row.Scan(&PhoneIDStr, &OrderIDStr, &t.OrderState, &t.SignNo, &t.IsRepeatorder, &t.Begintime, &t.Endtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("OrdersUserFlow row.Scan err (%v)", err)
		}
		return
	}
	t.TelecomChange()
	t.PhoneID, _ = strconv.Atoi(PhoneIDStr)
	t.OrderID, _ = strconv.ParseInt(OrderIDStr, 10, 64)
	if t.PhoneID > 0 {
		res[t.PhoneID] = t
	}
	return
}

func (d *Dao) OrdersUserByOrderID(ctx context.Context, orderID int64) (res map[int64]*telecom.OrderInfo, err error) {
	res = map[int64]*telecom.OrderInfo{}
	var (
		PhoneIDStr string
		OrderIDStr string
	)
	t := &telecom.OrderInfo{}
	row := d.orderByOrderIDSQL.QueryRow(ctx, orderID)
	if err = row.Scan(&PhoneIDStr, &OrderIDStr, &t.OrderState, &t.SignNo, &t.IsRepeatorder, &t.Begintime, &t.Endtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("OrdersUserFlow row.Scan err (%v)", err)
		}
		return
	}
	t.TelecomChange()
	t.PhoneID, _ = strconv.Atoi(PhoneIDStr)
	t.OrderID, _ = strconv.ParseInt(OrderIDStr, 10, 64)
	if t.OrderID > 0 {
		res[t.OrderID] = t
	}
	return
}

// InCardOrderSync inster telecom card order
func (d *Dao) InCardOrderSync(ctx context.Context, t *telecom.CardOrderBizJson) (err error) {
	row, err := d.db.Exec(ctx, _inCardOrderSyncSQL, t.Phone, t.Nbr, t.StartTime, t.EndTime, t.Action, t.CreateTime, t.AppKey,
		t.Nbr, t.StartTime, t.EndTime, t.Action, t.CreateTime, t.AppKey)
	if err != nil {
		log.Error("InCardOrderSync d.db.Exec error(%v)", err)
		return
	}
	result, err := row.RowsAffected()
	if err != nil || result == 0 {
		log.Error("d.dao.InCardOrderSync error(%v) or result==0", err)
		if result == 0 {
			err = ecode.ServerErr
		}
		return
	}
	return
}

// OrderUserByPhone select user order
func (d *Dao) OrderUserByPhone(ctx context.Context, phone int) (res []*telecom.CardOrder, err error) {
	rows, err := d.db.Query(ctx, _cardOrderSyncSQL, phone)
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &telecom.CardOrder{}
		if err = rows.Scan(&t.Phone, &t.Nbr, &t.StartTime, &t.EndTime, &t.Action, &t.AppKey); err != nil {
			log.Error("OrdersUserFlow row.Scan err (%v)", err)
			return
		}
		t.CardOrderChange()
		res = append(res, t)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// VipPackLog vip pack log
func (d *Dao) VipPackLog(ctx context.Context, start, end time.Time) (res []*telecom.CardVipLog, err error) {
	rows, err := d.db.Query(ctx, _vipPackLogSQL, start, end)
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &telecom.CardVipLog{}
		if err = rows.Scan(&t.Phone, &t.State, &t.RequestNo, &t.Ptype, &t.Ctime); err != nil {
			log.Error("vipPackLog packs log sql error(%v)", err)
			return
		}
		res = append(res, t)
	}
	err = rows.Err()
	return
}
