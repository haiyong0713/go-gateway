package mobile

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/model/mobile"
)

const (
	_inOrderSyncSQL     = "INSERT INTO mobile_order (orderid,userpseudocode,channelseqid,price,actiontime,actionid,effectivetime,expiretime,channelid,productid,ordertype,threshold) VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE orderid=VALUES(orderid),channelseqid=VALUES(channelseqid),price=VALUES(price),actiontime=VALUES(actiontime),actionid=VALUES(actionid),effectivetime=VALUES(effectivetime),expiretime=VALUES(expiretime),channelid=VALUES(channelid),productid=VALUES(productid),ordertype=VALUES(ordertype),threshold=VALUES(threshold)"
	_upOrderFlowSQL     = "UPDATE mobile_order SET threshold=?,resulttime=? WHERE userpseudocode=? AND productid=?"
	_orderSyncByUserSQL = "SELECT orderid,userpseudocode,channelseqid,price,actionid,effectivetime,expiretime,channelid,productid,ordertype,threshold FROM mobile_order WHERE userpseudocode=?"
)

type Dao struct {
	c  *conf.Config
	db *xsql.DB
	// memcache
	mc     *memcache.Memcache
	expire int32
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: xsql.NewMySQL(c.MySQL.Show),
		// memcache
		mc:     memcache.New(c.Memcache.Operator.Config),
		expire: int32(time.Duration(c.Memcache.Operator.Expire) / time.Second),
	}
	return
}

func (d *Dao) InOrdersSync(ctx context.Context, u *mobile.MobileXML) (int64, error) {
	if u.Effectivetime == "" {
		log.Error("日志告警 移动订单同步 effectivetime 为空,order:%+v", u)
	}
	if d.productType(u.Productid) == 0 {
		log.Error("日志告警 移动订单同步未知的productid,order:%+v", u)
	}
	res, err := d.db.Exec(ctx, _inOrderSyncSQL, u.Orderid, u.Userpseudocode, u.Channelseqid, u.Price, u.Actiontime, u.Actionid, u.Effectivetime, u.Expiretime, u.Channelid, u.Productid, u.Ordertype, u.Threshold)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// FlowSync update OrdersSync
func (d *Dao) FlowSync(ctx context.Context, u *mobile.MobileXML) (int64, error) {
	res, err := d.db.Exec(ctx, _upOrderFlowSQL, u.Threshold, u.Resulttime, u.Userpseudocode, u.Productid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) OrdersUserFlow(ctx context.Context, usermob string) ([]*mobile.Mobile, error) {
	rows, err := d.db.Query(ctx, _orderSyncByUserSQL, usermob)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mobile.Mobile
	for rows.Next() {
		u := &mobile.Mobile{}
		if err = rows.Scan(&u.Orderid, &u.Userpseudocode, &u.Channelseqid, &u.Price, &u.Actionid, &u.Effectivetime, &u.Expiretime,
			&u.Channelid, &u.Productid, &u.Ordertype, &u.Threshold); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// nolint:gomnd
func (d *Dao) productType(id string) int {
	for _, product := range d.c.Mobile.FlowProduct {
		if id == product.ID {
			return 1
		}
	}
	for _, product := range d.c.Mobile.CardProduct {
		if id == product.ID {
			return 2
		}
	}
	return 0
}
