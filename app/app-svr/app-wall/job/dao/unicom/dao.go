package unicom

import (
	"context"
	"database/sql"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-wall/job/conf"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"
)

const (
	// unicom integral change
	_orderUserSyncSQL = "SELECT usermob,spid,`type`,ordertime,endtime FROM unicom_order WHERE usermob=? AND `type`=0 ORDER BY `type` DESC"
	_bindAllSQL       = "SELECT usermob,phone,mid,state,integral,flow,monthlytime FROM unicom_user_bind WHERE state=1 ORDER BY id ASC LIMIT ?,?"
	_userBindSQL      = "SELECT usermob,phone,mid,state,integral,flow,monthlytime FROM unicom_user_bind WHERE state=1 AND mid=?"
	// update unicom ip
	_inUnicomIPSyncSQL = "INSERT IGNORE INTO unicom_ip (ipbegion,ipend,isopen,ctime,mtime) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE ipbegion=VALUES(ipbegion),ipend=VALUES(ipend),isopen=VALUES(isopen),mtime=VALUES(mtime)"
	_upUnicomIPSQL     = "UPDATE unicom_ip SET isopen=?,mtime=? WHERE ipbegion=? AND ipend=?"
	_ipSyncSQL         = "SELECT ipbegion,ipend FROM unicom_ip WHERE isopen=1"
	_inUserPackLogSQL  = "INSERT INTO unicom_user_packs_log (phone,usermob,mid,request_no,ptype,integral,pdesc) VALUES (?,?,?,?,?,?,?)"
	// update score
	_addUserIntegralSQL        = "UPDATE unicom_user_bind SET integral=integral+? WHERE mid=? AND phone=? AND state=1"
	_addUserMonthlyIntegralSQL = "UPDATE unicom_user_bind SET integral=integral+?,monthlytime=? WHERE mid=? AND phone=? AND state=1"
	_addUserFlowSQL            = "UPDATE unicom_user_bind SET flow=flow+? WHERE mid=? AND phone=? AND state=1"
	_addUserScoreSQL           = "UPDATE unicom_user_bind SET integral=integral+?,flow=flow+? WHERE mid=? AND phone=? AND state=1"
	_userPacksByIDSQL          = "SELECT id,ptype,pdesc,amount,capped,integral,param,state,original,kind,cover,new_param FROM unicom_user_packs WHERE id=?"
	_upUserPacksCappedSQL      = "UPDATE unicom_user_packs SET capped=? WHERE id=?"
	// unicom statistic
	_returnPackStatisticSQL = `UPDATE unicom_integral_statistic SET reduce_pack_integral=reduce_pack_integral-? WHERE phone=? AND period_time=? AND period=?`
	_returnFlowStatisticSQL = `UPDATE unicom_integral_statistic SET reduce_flow_integral=reduce_flow_integral-? WHERE phone=? AND period_time=? AND period=?`
)

type Dao struct {
	c       *conf.Config
	db      *xsql.DB
	uclient *httpx.Client
	// memcache
	mc             *memcache.Pool
	flowKeyExpired int32
	expire         int32
	// unicom url
	unicomFlowExchangeURL string
	unicomIPURL           string
	// redis
	redis *redis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:       c,
		db:      xsql.NewMySQL(c.MySQL.Show),
		uclient: httpx.NewClient(c.HTTPUnicom),
		// memcache
		mc:             memcache.NewPool(c.Memcache.Operator.Config),
		expire:         int32(time.Duration(c.Unicom.PackKeyExpired) / time.Second),
		flowKeyExpired: int32(time.Duration(c.Unicom.KeyExpired) / time.Second),
		// unicom url
		unicomFlowExchangeURL: c.Host.UnicomFlow + _unicomFlowExchangeURL,
		unicomIPURL:           c.Host.Unicom + _unicomIPURL,
		// redis
		redis: redis.NewRedis(c.Redis.Wall.Config),
	}
	return
}

func (d *Dao) AddUserBindScore(ctx context.Context, mid int64, phone string, integral, flow int, usermob string, now time.Time) (rows int64, err error) {
	res, err := d.db.Exec(ctx, _addUserScoreSQL, integral, flow, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// OrdersUserFlow select user OrdersSync
func (d *Dao) OrdersUserFlow(ctx context.Context, usermob string) ([]*unicom.Unicom, error) {
	rows, err := d.db.Query(ctx, _orderUserSyncSQL, usermob)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*unicom.Unicom
	for rows.Next() {
		u := &unicom.Unicom{}
		if err = rows.Scan(&u.Usermob, &u.Spid, &u.TypeInt, &u.Ordertime, &u.Endtime); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// BindAll select bind all mid state 1
func (d *Dao) BindAll(ctx context.Context, offset, limit int) ([]*unicom.UserBind, error) {
	rows, err := d.db.Query(ctx, _bindAllSQL, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*unicom.UserBind
	for rows.Next() {
		u := &unicom.UserBind{}
		if err = rows.Scan(&u.Usermob, &u.Phone, &u.Mid, &u.State, &u.Integral, &u.Flow, &u.Monthly); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// UserBind unicom select user bind
func (d *Dao) UserBind(ctx context.Context, mid int64) (*unicom.UserBind, error) {
	row := d.db.QueryRow(ctx, _userBindSQL, mid)
	res := &unicom.UserBind{}
	if err := row.Scan(&res.Usermob, &res.Phone, &res.Mid, &res.State, &res.Integral, &res.Flow, &res.Monthly); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

// InUnicomIPSync insert or update unicom_ip
func (d *Dao) InUnicomIPSync(tx *xsql.Tx, u *unicom.UnicomIP, now time.Time) error {
	_, err := tx.Exec(_inUnicomIPSyncSQL, u.Ipbegin, u.Ipend, 1, now, now)
	return err
}

// UpUnicomIP update unicom_ip state
func (d *Dao) UpUnicomIP(tx *xsql.Tx, ipStart, ipEnd, state int, now time.Time) (int64, error) {
	res, err := tx.Exec(_upUnicomIPSQL, state, now, ipStart, ipEnd)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// IPSync select all ipSync
func (d *Dao) IPSync(ctx context.Context) ([]*unicom.UnicomIP, error) {
	rows, err := d.db.Query(ctx, _ipSyncSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*unicom.UnicomIP
	for rows.Next() {
		u := &unicom.UnicomIP{}
		if err = rows.Scan(&u.Ipbegin, &u.Ipend); err != nil {
			return nil, err
		}
		u.UnicomIPChange()
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// InUserPackLog insert unicom user pack log
func (d *Dao) InUserPackLog(ctx context.Context, u *unicom.UserPackLog) (int64, error) {
	res, err := d.db.Exec(ctx, _inUserPackLogSQL, u.Phone, u.Usermob, u.Mid, u.RequestNo, u.Type, u.Integral, u.Desc)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BeginTran begin a transacition
func (d *Dao) BeginTran(ctx context.Context) (*xsql.Tx, error) {
	return d.db.Begin(ctx)
}

func monthStatistic(now time.Time) (periodTime time.Time, period string) {
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local), "month"
}

func (d *Dao) AddUserBindIntegral(ctx context.Context, mid int64, phone string, integral int) (rows int64, err error) {
	res, err := d.db.Exec(ctx, _addUserIntegralSQL, integral, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) BackUserBindIntegral(ctx context.Context, mid int64, phone string, integral int, now time.Time) (rows int64, err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	periodTime, period := monthStatistic(now)
	if _, err := tx.Exec(_returnPackStatisticSQL, integral, phone, periodTime, period); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_addUserIntegralSQL, integral, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) AddUserBindFlow(ctx context.Context, mid int64, phone string, flow int) (rows int64, err error) {
	res, err := d.db.Exec(ctx, _addUserFlowSQL, flow, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) BackUserBindFlow(ctx context.Context, mid int64, phone string, flow int, now time.Time) (rows int64, err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	periodTime, period := monthStatistic(now)
	if _, err := tx.Exec(_returnFlowStatisticSQL, flow, phone, periodTime, period); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_addUserFlowSQL, flow, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) AddUserBindMonthlyIntegral(ctx context.Context, mid int64, phone string, integral int, monthly time.Time, usermob string) (rows int64, err error) {
	res, err := d.db.Exec(ctx, _addUserMonthlyIntegralSQL, integral, monthly, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UserPackByID user pack by id
func (d *Dao) UserPackByID(ctx context.Context, id int64) (*unicom.UserPack, error) {
	row := d.db.QueryRow(ctx, _userPacksByIDSQL, id)
	u := &unicom.UserPack{}
	if err := row.Scan(&u.ID, &u.Type, &u.Desc, &u.Amount, &u.Capped, &u.Integral, &u.Param, &u.State, &u.Original, &u.Kind, &u.Cover, &u.NewParam); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (d *Dao) SetUserPackFlow(ctx context.Context, id int64, capped int8) (int64, error) {
	res, err := d.db.Exec(ctx, _upUserPacksCappedSQL, capped, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
