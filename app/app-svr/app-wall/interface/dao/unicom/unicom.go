package unicom

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/cache/memcache"
	"go-common/library/database/elastic"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

const (
	//unicom
	_inOrderSyncSQL = `INSERT INTO unicom_order (usermob,cpid,spid,type,ordertime,canceltime,endtime,channelcode,province,area,ordertype,videoid,ctime,mtime) 
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE cpid=VALUES(cpid),spid=VALUES(spid),type=VALUES(type),ordertime=VALUES(ordertime),canceltime=VALUES(canceltime),endtime=VALUES(endtime),channelcode=VALUES(channelcode),province=VALUES(province),area=VALUES(area),ordertype=VALUES(ordertype),videoid=VALUES(videoid)`
	_inAdvanceSyncSQL = `INSERT IGNORE INTO unicom_order_advance (usermob,userphone,cpid,spid,ordertime,channelcode,province,area,ctime,mtime) 
	VALUES (?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE cpid=?,spid=?,ordertime=?,channelcode=?,province=?,area=?,mtime=?`
	_upOrderFlowSQL   = `UPDATE unicom_order SET time=?,flowbyte=?,mtime=? WHERE usermob=?`
	_orderUserSyncSQL = `SELECT usermob,cpid,spid,type,ordertime,canceltime,endtime,channelcode,province,area,ordertype,videoid,time,flowbyte FROM unicom_order WHERE usermob=? 
	ORDER BY type DESC`
	_inIPSyncSQL = `INSERT IGNORE INTO unicom_ip (ipbegion,ipend,provinces,isopen,opertime,sign,ctime,mtime) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE 
		ipbegion=?,ipend=?,provinces=?,isopen=?,opertime=?,sign=?,mtime=?`
	_ipSyncSQL = `SELECT ipbegion,ipend FROM unicom_ip WHERE isopen=1`
	//pack
	_inPackSQL = `INSERT IGNORE INTO unicom_pack (usermob,mid) VALUES (?,?)`
	// unicom integral change
	_userBindSQL         = `SELECT usermob,phone,mid,state,integral,flow,monthlytime FROM unicom_user_bind WHERE state=1 AND mid=?`
	_userBindByMidsSQL   = `SELECT usermob,phone,mid,state,integral,flow,monthlytime FROM unicom_user_bind WHERE state=1 AND mid in (%s)`
	_userBindPhoneMidSQL = `SELECT mid FROM unicom_user_bind WHERE phone=? AND state=1`
	_userPacksSQL        = `SELECT id,ptype,pdesc,amount,capped,integral,param,original,kind,cover FROM unicom_user_packs WHERE state IN (%s)`
	_userPacksByIDSQL    = `SELECT id,ptype,pdesc,amount,capped,integral,param,state,original,kind,cover,new_param FROM unicom_user_packs WHERE id=?`
	_inUserPackLogSQL    = `INSERT INTO unicom_user_packs_log (phone,usermob,mid,request_no,ptype,integral,pdesc) VALUES (?,?,?,?,?,?,?)`
	_userPacksLogSQL     = `SELECT phone,integral,ptype,pdesc FROM unicom_user_packs_log WHERE mtime>=? AND mtime<?`
	// unicom select
	_userBindByPhoneSQL = `SELECT mid,usermob,integral,flow,ctime,mtime FROM unicom_user_bind WHERE phone=? AND state=1 LIMIT 1`
	// unicom pack
	_consumeUserIntegralSQL = "UPDATE unicom_user_bind SET integral=integral-? WHERE mid=? AND phone=? AND integral>=? AND state=1"
	_consumeUserFlowSQL     = "UPDATE unicom_user_bind SET flow=flow-? WHERE mid=? AND phone=? AND flow>=? AND state=1"
	_addUserIntegralSQL     = "UPDATE unicom_user_bind SET integral=integral+? WHERE mid=? AND phone=? AND state=1"
	_addUserFlowSQL         = "UPDATE unicom_user_bind SET flow=flow+? WHERE mid=? AND phone=? AND state=1"
	_upUserPacksCappedSQL   = "UPDATE unicom_user_packs SET capped=? WHERE id=?"
	// unicom statistic
	_consumePackStatisticSQL = `INSERT INTO unicom_integral_statistic (phone,period_time,period,usermob,reduce_pack_integral) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE usermob=VALUES(usermob),reduce_pack_integral=reduce_pack_integral+?`
	_consumeFlowStatisticSQL = `INSERT INTO unicom_integral_statistic (phone,period_time,period,usermob,reduce_flow_integral) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE usermob=VALUES(usermob),reduce_flow_integral=reduce_flow_integral+?`
	_addPackStatisticSQL     = `INSERT INTO unicom_integral_statistic (phone,period_time,period,usermob,add_pack_integral) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE usermob=VALUES(usermob),add_pack_integral=add_pack_integral+?`
	_addFlowStatisticSQL     = `INSERT INTO unicom_integral_statistic (phone,period_time,period,usermob,add_flow_integral) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE usermob=VALUES(usermob),add_flow_integral=add_flow_integral+?`
	_reducePackIntegralSQL   = "SELECT reduce_pack_integral FROM unicom_integral_statistic WHERE phone=? AND period_time=? AND period=?"
	_selectUsermobInfoSQL    = "SELECT usermob,fake_id,period,fake_id_month FROM unicom_usermob_info WHERE fake_id = ? AND period = ?"
	_syncUsermobInfoSQL      = "INSERT INTO unicom_usermob_info (usermob,period,fake_id,fake_id_month) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE fake_id = VALUES(fake_id),fake_id_month = VALUES(fake_id_month)"
)

type Dao struct {
	c              *conf.Config
	db             *xsql.DB
	logDB          *xsql.DB
	client         *httpx.Client
	uclient        *httpx.Client
	activateClient *httpx.Client
	//unicom
	inAdvanceSyncSQL *xsql.Stmt
	upOrderFlowSQL   *xsql.Stmt
	orderUserSyncSQL *xsql.Stmt
	inIPSyncSQL      *xsql.Stmt
	ipSyncSQL        *xsql.Stmt
	// unicom integral change
	userBindSQL         *xsql.Stmt
	userBindPhoneMidSQL *xsql.Stmt
	inUserPackLogSQL    *xsql.Stmt
	// memcache
	mc             *memcache.Memcache
	expire         int32
	flowKeyExpired int32
	emptyExpire    int32
	usermobExpire  int32
	// unicom url
	unicomFlowExchangeURL string
	// order url
	orderURL       string
	ordercancelURL string
	sendsmscodeURL string
	smsNumberURL   string
	// elastic
	es *elastic.Elastic
	// activate
	activateURL string
	// 取伪码地址
	usermobURL string
	// 联通免流订购验证
	unicomVerifyURL string
	// 免流试看订购关系生成
	unicomFlowTryout string
	// redis
	redis credis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		db:             xsql.NewMySQL(c.MySQL.Show),
		logDB:          xsql.NewMySQL(c.MySQL.ShowLog),
		client:         httpx.NewClient(c.HTTPBroadband),
		uclient:        httpx.NewClient(c.HTTPUnicom),
		activateClient: httpx.NewClient(c.HTTPActivate),
		redis:          credis.NewRedis(c.Redis.Wall.Config),
		// unicom url
		unicomFlowExchangeURL: c.Host.UnicomFlow + _unicomFlowExchangeURL,
		// memcache
		mc:             memcache.New(c.Memcache.Operator.Config),
		expire:         int32(time.Duration(c.Memcache.Operator.Expire) / time.Second),
		flowKeyExpired: int32(time.Duration(c.Unicom.KeyExpired) / time.Second),
		emptyExpire:    int32(time.Duration(c.Memcache.Operator.EmptyExpire) / time.Second),
		usermobExpire:  int32(time.Duration(c.Memcache.Operator.UsermobExpire) / time.Second),

		// order url
		orderURL:       c.Host.Broadband + _orderURL,
		ordercancelURL: c.Host.Broadband + _ordercancelURL,
		sendsmscodeURL: c.Host.Broadband + _sendsmscodeURL,
		smsNumberURL:   c.Host.Broadband + _smsNumberURL,
		// elastic
		es: elastic.NewElastic(nil),
		// activate
		activateURL: c.Host.UnicomActivate + _activateURL,
		// 联通取伪码
		usermobURL: c.Host.UnicomUsermob + _usermobURL,
		// 联通免流订购验证
		unicomVerifyURL: c.Host.UnicomVerify + _unicomVerifyURL,
		// 免流试看订购关系生成
		unicomFlowTryout: c.Host.UnicomFlowTryout + _unicomFlowTryoutURL,
	}
	d.inAdvanceSyncSQL = d.db.Prepared(_inAdvanceSyncSQL)
	d.upOrderFlowSQL = d.db.Prepared(_upOrderFlowSQL)
	d.orderUserSyncSQL = d.db.Prepared(_orderUserSyncSQL)
	d.inIPSyncSQL = d.db.Prepared(_inIPSyncSQL)
	d.ipSyncSQL = d.db.Prepared(_ipSyncSQL)
	// unicom integral change
	d.userBindSQL = d.db.Prepared(_userBindSQL)
	d.userBindPhoneMidSQL = d.db.Prepared(_userBindPhoneMidSQL)
	d.inUserPackLogSQL = d.db.Prepared(_inUserPackLogSQL)
	return
}

// InOrdersSync insert OrdersSync
func (d *Dao) InOrdersSync(ctx context.Context, usermob string, u *unicom.UnicomJson, now time.Time) error {
	if !d.cardType(u.Spid) {
		log.Error("日志告警 联通订单同步未知的spid,order:%+v", u)
	}
	_, err := d.db.Exec(ctx, _inOrderSyncSQL, usermob, u.Cpid, u.Spid, u.TypeInt, u.Ordertime, u.Canceltime, u.Endtime, u.Channelcode, u.Province, u.Area, u.Ordertypes, u.Videoid, now, now)
	return err
}

// InAdvanceSync insert AdvanceSync
func (d *Dao) InAdvanceSync(ctx context.Context, usermob string, u *unicom.UnicomJson, now time.Time) (row int64, err error) {
	res, err := d.inAdvanceSyncSQL.Exec(ctx, usermob, u.Userphone,
		u.Cpid, u.Spid, u.Ordertime, u.Channelcode, u.Province, u.Area, now, now,
		u.Cpid, u.Spid, u.Ordertime, u.Channelcode, u.Province, u.Area, now)
	if err != nil {
		log.Error("d.inAdvanceSyncSQL.Exec error(%v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

// FlowSync update OrdersSync
func (d *Dao) FlowSync(ctx context.Context, flowbyte int, usermob, time string, now time.Time) (row int64, err error) {
	res, err := d.upOrderFlowSQL.Exec(ctx, time, flowbyte, now, usermob)
	if err != nil {
		log.Error("d.upOrderFlowSQL.Exec error(%v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

// OrdersUserFlow select user OrdersSync
func (d *Dao) OrdersUserFlow(ctx context.Context, usermob string) (res []*unicom.Unicom, err error) {
	rows, err := d.orderUserSyncSQL.Query(ctx, usermob)
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		u := &unicom.Unicom{}
		if err = rows.Scan(&u.Usermob, &u.Cpid, &u.Spid, &u.TypeInt, &u.Ordertime, &u.Canceltime, &u.Endtime, &u.Channelcode, &u.Province,
			&u.Area, &u.Ordertypes, &u.Videoid, &u.Time, &u.Flowbyte); err != nil {
			log.Error("OrdersUserFlow row.Scan err (%v)", err)
			return
		}
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// InIPSync insert IpSync
func (d *Dao) InIPSync(ctx context.Context, u *unicom.UnicomIpJson, now time.Time) (row int64, err error) {
	res, err := d.inIPSyncSQL.Exec(ctx, u.Ipbegin, u.Ipend, u.Provinces, u.Isopen, u.Opertime, u.Sign, now, now,
		u.Ipbegin, u.Ipend, u.Provinces, u.Isopen, u.Opertime, u.Sign, now)
	if err != nil {
		log.Error("d.inIPSyncSQL.Exec error(%v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) InPack(tx *xsql.Tx, usermob string, mid int64) (row int64, err error) {
	res, err := tx.Exec(_inPackSQL, usermob, mid)
	if err != nil {
		return
	}
	return res.RowsAffected()
}

// IPSync select all ipSync
func (d *Dao) IPSync(ctx context.Context) (res []*unicom.UnicomIP, err error) {
	rows, err := d.ipSyncSQL.Query(ctx)
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	res = []*unicom.UnicomIP{}
	for rows.Next() {
		u := &unicom.UnicomIP{}
		if err = rows.Scan(&u.Ipbegin, &u.Ipend); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		u.UnicomIPChange()
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

func (d *Dao) BindUser(ctx context.Context, mid int64, phone, usermob string) (int64, error) {
	const (
		lastSQL = "SELECT integral,flow,state FROM unicom_user_bind WHERE phone=? ORDER BY ctime DESC LIMIT 1"
		addSQL  = "INSERT INTO unicom_user_bind (mid,phone,usermob,integral,flow,state,ctime) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE usermob=VALUES(usermob),integral=VALUES(integral),flow=VALUES(flow),state=VALUES(state),ctime=VALUES(ctime)"
	)
	row := d.db.QueryRow(ctx, lastSQL, phone)
	var (
		integral, flow int64
		state          int
	)
	if err := row.Scan(&integral, &flow, &state); err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	if state == 1 {
		return 0, xecode.AppWelfareClubRegistered
	}
	res, err := d.db.Exec(ctx, addSQL, mid, phone, usermob, integral, flow, 1, time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) UnbindUser(ctx context.Context, mid int64, phone string) (int64, error) {
	const updateStateSQL = "UPDATE unicom_user_bind SET state=?,ctime=? WHERE mid=? AND phone=? AND state!=?"
	res, err := d.db.Exec(ctx, updateStateSQL, 0, time.Now(), mid, phone, 0)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UserBind unicom select user bind
func (d *Dao) UserBind(ctx context.Context, mid int64) (res *unicom.UserBind, err error) {
	row := d.userBindSQL.QueryRow(ctx, mid)
	if row == nil {
		log.Error("userBindSQL is null")
		return
	}
	res = &unicom.UserBind{}
	if err = row.Scan(&res.Usermob, &res.Phone, &res.Mid, &res.State, &res.Integral, &res.Flow, &res.Monthly); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("userBindSQL row.Scan error(%v)", err)
		}
		res = nil
		return
	}
	return
}

// UserPacks user pack list
func (d *Dao) UserPacks(ctx context.Context, states []int) (res []*unicom.UserPack, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, state := range states {
		sqls = append(sqls, "?")
		args = append(args, state)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_userPacksSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("user pack sql error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		u := &unicom.UserPack{}
		if err = rows.Scan(&u.ID, &u.Type, &u.Desc, &u.Amount, &u.Capped, &u.Integral, &u.Param, &u.Original, &u.Kind, &u.Cover); err != nil {
			log.Error("user pack sql error(%v)", err)
			return
		}
		res = append(res, u)
	}
	err = rows.Err()
	return
}

// UserPackByID user pack by id
func (d *Dao) UserPackByID(ctx context.Context, id int64) (*unicom.UserPack, error) {
	row := d.db.QueryRow(ctx, _userPacksByIDSQL, id)
	u := &unicom.UserPack{}
	if err := row.Scan(&u.ID, &u.Type, &u.Desc, &u.Amount, &u.Capped, &u.Integral, &u.Param, &u.State, &u.Original, &u.Kind, &u.Cover, &u.NewParam); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("userPacksByIDSQL row.Scan error(%v)", err)
		return nil, err
	}
	return u, nil
}

// UserBindPhoneMid mid by phone
func (d *Dao) UserBindPhoneMid(ctx context.Context, phone string) (mid int64, err error) {
	row := d.userBindPhoneMidSQL.QueryRow(ctx, phone)
	if row == nil {
		log.Error("user pack sql is null")
		return
	}
	if err = row.Scan(&mid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("userPacksByIDSQL row.Scan error(%v)", err)
		}
		return
	}
	return
}

// InUserPackLog insert unicom user pack log
func (d *Dao) InUserPackLog(ctx context.Context, u *unicom.UserPackLog) (row int64, err error) {
	res, err := d.inUserPackLogSQL.Exec(ctx, u.Phone, u.Usermob, u.Mid, u.RequestNo, u.Type, u.Integral, u.Desc)
	if err != nil {
		log.Error("insert user pack log integral sql error(%v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

// UserPacksLog user pack logs
func (d *Dao) UserPacksLog(ctx context.Context, start, end time.Time) ([]*unicom.UserPackLog, error) {
	rows, err := d.logDB.Query(ctx, _userPacksLogSQL, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*unicom.UserPackLog
	for rows.Next() {
		u := &unicom.UserPackLog{}
		if err = rows.Scan(&u.Phone, &u.Integral, &u.Type, &u.UserDesc); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// UserBindByMids unicom select user bind
func (d *Dao) UserBindByMids(ctx context.Context, mids []int64) (res map[int64]*unicom.UserBind, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_userBindByMidsSQL, xstr.JoinInts(mids)))
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	res = map[int64]*unicom.UserBind{}
	for rows.Next() {
		u := &unicom.UserBind{}
		if err = rows.Scan(&u.Usermob, &u.Phone, &u.Mid, &u.State, &u.Integral, &u.Flow, &u.Monthly); err != nil {
			log.Error("userBindSQL row.Scan error(%v)", err)
			return
		}
		res[u.Mid] = u
	}
	err = rows.Err()
	return
}

// BeginTran begin a transacition
func (d *Dao) BeginTran(ctx context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(ctx)
}

func (d *Dao) UserBindInfoByPhone(ctx context.Context, phone string) (res *unicom.UserBindV2, err error) {
	row := d.db.QueryRow(ctx, _userBindByPhoneSQL, phone)
	if row == nil {
		log.Error("UserBindInfoByPhone is null")
		err = ecode.NothingFound
		return
	}
	var (
		ctime, mtime time.Time
	)
	res = &unicom.UserBindV2{}
	if err = row.Scan(&res.Mid, &res.Usermob, &res.Integral, &res.Flow, &ctime, &mtime); err != nil {
		if err == sql.ErrNoRows {
			res = nil
			err = xecode.AppWelfareClubNoBinding
		} else {
			log.Error("userPacksByIDSQL row.Scan error(%v)", err)
		}
		return
	}
	res.UserBindDateChange(ctime, mtime)
	return
}

func monthStatistic(now time.Time) (period time.Time, interval string) {
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local), "month"
}

func (d *Dao) ConsumeUserBindIntegral(ctx context.Context, mid int64, phone string, integral int, usermob string, now time.Time) (rows int64, err error) {
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
	if _, err := tx.Exec(_consumePackStatisticSQL, phone, periodTime, period, usermob, integral, integral); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_consumeUserIntegralSQL, integral, mid, phone, integral)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) ConsumeUserBindFlow(ctx context.Context, mid int64, phone string, flow int, usermob string, now time.Time) (rows int64, err error) {
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
	if _, err := tx.Exec(_consumeFlowStatisticSQL, phone, periodTime, period, usermob, flow, flow); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_consumeUserFlowSQL, flow, mid, phone, flow)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) AddUserBindIntegral(ctx context.Context, mid int64, phone string, integral int, usermob string, now time.Time) (rows int64, err error) {
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
	if _, err := tx.Exec(_addPackStatisticSQL, phone, periodTime, period, usermob, integral, integral); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_addUserIntegralSQL, integral, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) AddUserBindFlow(ctx context.Context, mid int64, phone string, flow int, usermob string, now time.Time) (rows int64, err error) {
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
	if _, err := tx.Exec(_addFlowStatisticSQL, phone, periodTime, period, usermob, flow, flow); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_addUserFlowSQL, flow, mid, phone)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) SetUserPackFlow(ctx context.Context, id int64, capped int8) (rows int64, err error) {
	res, err := d.db.Exec(ctx, _upUserPacksCappedSQL, capped, id)
	if err != nil {
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ReducePackIntegral(ctx context.Context, phone string, now time.Time) (int64, error) {
	periodTime, period := monthStatistic(now)
	row := d.db.QueryRow(ctx, _reducePackIntegralSQL, phone, periodTime, period)
	var reducePackIntegral int64
	if err := row.Scan(&reducePackIntegral); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return reducePackIntegral, nil
}

func (d *Dao) cardType(spid string) bool {
	for _, product := range d.c.Unicom.CardProduct {
		if spid == product.Spid {
			return true
		}
	}
	for _, product := range d.c.Unicom.FlowProduct {
		if spid == product.Spid {
			return true
		}
	}
	return false
}

func (d *Dao) SelectUserMobInfo(ctx context.Context, fakeID string, period int64) (*unicom.UserMobInfo, error) {
	row := d.db.QueryRow(ctx, _selectUsermobInfoSQL, fakeID, period)
	info := &unicom.UserMobInfo{}
	if err := row.Scan(&info.Usermob, &info.FakeID, &info.Period, &info.Month); err != nil {
		if err == sql.ErrNoRows {
			log.Warn("[dao.SelectUserMobInfo] usermob is empty, fake_id:%v, period:%v, error:%v", fakeID, period, err)
			return &unicom.UserMobInfo{}, nil
		}
		log.Error("[dao.SelectUserMobInfo] error, fake_id:%v, period:%v, error:%v", fakeID, period, err)
		return nil, err
	}
	return info, nil
}

func (d *Dao) InsertOrUpdateUserMobInfo(ctx context.Context, info *unicom.UserMobInfo) error {
	_, err := d.db.Exec(ctx, _syncUsermobInfoSQL, info.Usermob, info.Period, info.FakeID, info.Month)
	if err != nil {
		log.Error("[dao.InsertOrUpdateUserMobInfo] sync usermob info error, info:%+v, error:%v", info, err)
		return err
	}
	return nil
}
