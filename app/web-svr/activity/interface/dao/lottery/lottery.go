package lottery

import (
	"context"
	xsql "database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/lottery"
	l "go-gateway/app/web-svr/activity/interface/model/lottery"
	lotterymdl "go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/pkg/errors"
)

const (
	_lotterySQL                    = "SELECT id,lottery_id,name,stime,etime,ctime,mtime,type,state FROM act_lottery WHERE lottery_id = ? and state = 0 and type = 0"
	_lotteryInfoSQL                = "SELECT id,sid,level,regtime_stime,regtime_etime,vip_check,account_check,coin,fs_ip, gift_rate,high_type,high_rate,state,sender_id,activity_link FROM act_lottery_info WHERE sid = ? and state = 0"
	_lotteryTimesConfSQL           = "SELECT id,sid,type,add_type,times,info,most,state FROM act_lottery_times WHERE sid = ? and state = 0"
	_lotteryGiftSQL                = "SELECT id,sid,ctime,mtime,name,num,type,source,img_url,time_limit,is_show,least_mark,msg_title,msg_content,efficient,send_num,state FROM act_lottery_gift WHERE sid = ? and state = 0"
	_updateLotteryGiftNumSQL       = "UPDATE act_lottery_gift SET send_num = send_num+1 WHERE id = ? AND send_num < num"
	_lotteryGiftNumTimingSQL       = "SELECT id,send_num FROM act_lottery_gift WHERE sid IN (%s)"
	_lotterySidSQL                 = "SELECT lottery_id FROM act_lottery WHERE etime > ?"
	_lotteryAddrSQL                = "SELECT address_id FROM act_lottery_gift_address_%d WHERE mid = ? and state = 0"
	_lotteryWinListSQL             = "SELECT mid,gift_id,ctime FROM act_lottery_win_%d WHERE mid > 0 and gift_id in (%s) ORDER BY mtime DESC LIMIT ?"
	_insertLotteryAddrSQL          = "INSERT INTO act_lottery_gift_address_%d(mid,address_id) VALUES(?,?) ON DUPLICATE KEY UPDATE address_id = ?"
	_insertLotteryAddTimesSQL      = "INSERT INTO act_lottery_addtimes_%d(mid,type,num,cid,ip,order_no) VALUES(?,?,?,?,?,?)"
	_insertLotteryRecordSQL        = "INSERT INTO act_lottery_action_%d(mid,num,gift_id,type,cid,ip) VALUES %s"
	_lotteryAddTimesSQL            = "SELECT id,mid,type,num,cid,ctime FROM act_lottery_addtimes_%d WHERE mid = ? AND state = 0"
	_lotteryUsedTimesSQL           = "SELECT id,mid,num,gift_id,type,ctime,cid FROM act_lottery_action_%d WHERE mid = ? AND state = 0 order by ctime desc"
	_updateLotteryWin              = "UPDATE act_lottery_win_%d SET mid = ?, ip = ? WHERE gift_id = ? AND mid = 0 ORDER BY id LIMIT 1"
	_insertLotteryWin              = "INSERT INTO act_lottery_win_%d(mid,gift_id,ip) VALUES(?,?,?)"
	_lotteryWinOneSQL              = "SELECT cdkey FROM act_lottery_win_%d WHERE mid = ? AND gift_id = ? ORDER BY mtime DESC,id DESC"
	_lotteryOrderNoCheckSQL        = "SELECT id FROM act_lottery_addtimes_%d WHERE order_no = ? and mid = ?"
	_giftNumUpdateSQL              = "UPDATE act_lottery_gift SET num = num+? WHERE sid=? AND state=0"
	_getAddressURL                 = "/api/basecenter/addr/view"
	_memberCouponURI               = "/x/internal/coupon/allowance/receive"
	_memberVipURI                  = "/x/internal/vip/resources/grant"
	_daily                         = 1
	_winType                       = 2
	_insertLotteryRecordOrderNoSQL = "INSERT INTO act_lottery_action_%d(mid,num,gift_id,type,cid,ip,order_no) VALUES %s"
)

// lotteryMcNumKey
func lotteryMcNumKey(sid int64, high, mc int) string {
	return fmt.Sprintf("lottery_mc_%d_%d_%d", sid, mc, high)
}

func rsKey(actionType int, cid int64) string {
	return strconv.Itoa(actionType) + strconv.FormatInt(cid, 10)
}

// RawLottery get lottery by sid
func (dao *Dao) RawLottery(c context.Context, sid string) (res *lottery.Lottery, err error) {
	res = new(lotterymdl.Lottery)
	row := dao.db.QueryRow(c, _lotterySQL, sid)
	if err = row.Scan(&res.ID, &res.LotteryID, &res.Name, &res.Stime, &res.Etime, &res.Ctime, &res.Mtime, &res.Type, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLottery:QueryRow")
		}
	}
	return
}

// RawLotteryInfo get lotteryInfo by sid
func (dao *Dao) RawLotteryInfo(c context.Context, sid string) (res *lotterymdl.LotteryInfo, err error) {
	res = new(lotterymdl.LotteryInfo)
	row := dao.db.QueryRow(c, _lotteryInfoSQL, sid)
	if err = row.Scan(&res.ID, &res.Sid, &res.Level, &res.RegTimeStime, &res.RegTimeEtime, &res.VipCheck, &res.AccountCheck, &res.Coin, &res.FsIP, &res.GiftRate, &res.HighType, &res.HighRate, &res.State, &res.SenderID, &res.ActivityLink); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryInfo:QueryRow")
		}
	}
	return
}

func (dao *Dao) RawLotteryTimesConfig(c context.Context, sid string) (res []*lotterymdl.LotteryTimesConfig, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _lotteryTimesConfSQL, sid); err != nil {
		err = errors.Wrap(err, "RawLotteryTimesConfig:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotterymdl.LotteryTimesConfig, 0)
	for rows.Next() {
		l := &lotterymdl.LotteryTimesConfig{}
		if err = rows.Scan(&l.ID, &l.Sid, &l.Type, &l.AddType, &l.Times, &l.Info, &l.Most, &l.State); err != nil {
			err = errors.Wrap(err, "RawLotteryTimesConfig:rows.Scan")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryTimesConfig: rows.Err()")
	}
	return
}

// RawLotteryGift get lotteryGift by sid
func (dao *Dao) RawLotteryGift(c context.Context, sid string) (res []*lotterymdl.LotteryGift, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _lotteryGiftSQL, sid); err != nil {
		err = errors.Wrap(err, "RawLotteryGift:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotterymdl.LotteryGift, 0)
	for rows.Next() {
		l := &lotterymdl.LotteryGift{}
		if err = rows.Scan(&l.ID, &l.Sid, &l.Ctime, &l.Mtime, &l.Name, &l.Num, &l.Type, &l.Source, &l.ImgUrl, &l.TimeLimit, &l.IsShow, &l.LeastMark, &l.MessageTitle, &l.MessageContent, &l.Efficient, &l.SendNum, &l.State); err != nil {
			err = errors.Wrap(err, "RawLotteryGift:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryGift:rows.Err()")
	}
	return
}

// RawLotteryAddrCheck
func (dao *Dao) RawLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error) {
	row := dao.db.QueryRow(c, fmt.Sprintf(_lotteryAddrSQL, id), mid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryAddrCheck:QueryRow")
		}
	}
	return
}

// InsertLotteryAddr
func (dao *Dao) InsertLotteryAddr(c context.Context, id, mid, addressId int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, fmt.Sprintf(_insertLotteryAddrSQL, id), mid, addressId, addressId); err != nil {
		err = errors.Wrap(err, "InsertLotteryAddr:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertLotteryAddTimes
func (dao *Dao) InsertLotteryAddTimes(c context.Context, id int64, mid int64, addType, num int, cid int64, ip, orderNo string) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, fmt.Sprintf(_insertLotteryAddTimesSQL, id), mid, addType, num, cid, ip, orderNo); err != nil {
		err = errors.Wrap(err, "InsertLotteryAddTimes:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertLotteryRecard
func (dao *Dao) InsertLotteryRecard(c context.Context, id int64, record []*lotterymdl.InsertRecord, gid []int64, ip string) (count int64, err error) {
	if len(record) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(record))
		args = make([]interface{}, 0, len(record)*2)
	)
	for k, v := range record {
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, v.Mid, v.Num, gid[k], v.Type, v.CID, ip)
	}
	rows, err := dao.db.Exec(c, fmt.Sprintf(_insertLotteryRecordSQL, id, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("InsertLotteryRecard: dao.db.Exec() id(%d) error(%v)", id, err)
		return
	}
	return rows.LastInsertId()
}

// GetLotteryAddTimes
func (dao *Dao) RawLotteryAddTimes(c context.Context, id int64, mid int64) (res []*lotterymdl.LotteryAddTimes, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_lotteryAddTimesSQL, id), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryAddTimes:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotterymdl.LotteryAddTimes, 0)
	for rows.Next() {
		a := &lotterymdl.LotteryAddTimes{}
		if err = rows.Scan(&a.ID, &a.Mid, &a.Type, &a.Num, &a.CID, &a.Ctime); err != nil {
			err = errors.Wrap(err, "RawLotteryAddTimes:rows.Scan()")
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryAddTimes:rows.Err()")
	}
	return
}

// InsertLotteryRecardOrderNo ...
func (dao *Dao) InsertLotteryRecardOrderNo(c context.Context, id int64, record []*lotterymdl.InsertRecord, gid []int64, ip string) (count int64, err error) {
	if len(record) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(record))
		args = make([]interface{}, 0, len(record)*2)
	)
	for k, v := range record {
		sqls = append(sqls, "(?,?,?,?,?,?,?)")
		args = append(args, v.Mid, v.Num, gid[k], v.Type, v.CID, ip, v.OrderNo)
	}
	rows, err := dao.db.Exec(c, fmt.Sprintf(_insertLotteryRecordOrderNoSQL, id, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("InsertLotteryRecardOrderNo: dao.db.Exec() id(%d) error(%v)", id, err)
		return
	}
	return rows.LastInsertId()
}

// RawLotteryUsedTimes
func (dao *Dao) RawLotteryUsedTimes(c context.Context, id int64, mid int64) (res []*lotterymdl.LotteryRecordDetail, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_lotteryUsedTimesSQL, id), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryUsedTimes:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotterymdl.LotteryRecordDetail, 0)
	for rows.Next() {
		l := &lotterymdl.LotteryRecordDetail{}
		if err = rows.Scan(&l.ID, &l.Mid, &l.Num, &l.GiftID, &l.Type, &l.Ctime, &l.CID); err != nil {
			err = errors.Wrap(err, "RawLotteryUsedTimes:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryUsedTimes:rows.Err()")
	}
	return
}

func (dao *Dao) GetMemberAddress(c context.Context, id, mid int64) (val *lotterymdl.AddressInfo, err error) {
	var res struct {
		Errno int                     `json:"errno"`
		Msg   string                  `json:"msg"`
		Data  *lotterymdl.AddressInfo `json:"data"`
	}
	params := url.Values{}
	params.Set("app_id", dao.c.Lottery.AppKey)
	params.Set("app_token", dao.c.Lottery.AppToken)
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("uid", strconv.FormatInt(mid, 10))
	if err = dao.client.Get(c, dao.getAddressURL, "", params, &res); err != nil {
		log.Error("GetMemberAddress:dao.client.Get id(%d) mid(%d) error(%v)", id, mid, err)
		return
	}
	if res.Errno != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Errno), dao.getAddressURL+"?"+params.Encode())
	}
	val = res.Data
	return
}

func (dao *Dao) RawLotteryWinList(c context.Context, id int64, giftIDs []int64, num int64) (res []*lotterymdl.GiftList, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_lotteryWinListSQL, id, xstr.JoinInts(giftIDs)), num); err != nil {
		err = errors.Wrap(err, "RawLotteryWinList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotterymdl.GiftList, 0)
	for rows.Next() {
		l := &lotterymdl.GiftList{}
		if err = rows.Scan(&l.Mid, &l.GiftID, &l.Ctime); err != nil {
			err = errors.Wrap(err, "RawLotteryWinList:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryWinList:rows.Err()")
	}
	return
}

func (dao *Dao) InsertLotteryWin(c context.Context, id, giftID, mid int64, ip string) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, fmt.Sprintf(_insertLotteryWin, id), mid, giftID, ip); err != nil {
		err = errors.Wrap(err, "InsertLotteryWin:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

func (dao *Dao) UpdateLotteryWin(c context.Context, id int64, mid int64, giftID int64, ip string) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, fmt.Sprintf(_updateLotteryWin, id), mid, ip, giftID); err != nil {
		err = errors.Wrap(err, "UpdateLotteryWin:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

func (dao *Dao) RawLotteryWinOne(c context.Context, id, mid, giftID int64) (res string, err error) {
	row := dao.db.QueryRow(c, fmt.Sprintf(_lotteryWinOneSQL, id), mid, giftID)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryWinOne:QueryRow")
		}
	}
	return
}

// Lottery get data from cache if miss will call source method, then add to cache.
func (dao *Dao) Lottery(c context.Context, sid string) (res *l.Lottery, err error) {
	addCache := true
	res, err = dao.CacheLottery(c, sid)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("Lottery")
		return
	}
	cache.MetricMisses.Inc("Lottery")
	res, err = dao.RawLottery(c, sid)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLottery(c, sid, miss)
	})
	return
}

// LotteryInfo get data from cache if miss will call source method, then add to cache.
func (dao *Dao) LotteryInfo(c context.Context, sid string) (res *l.LotteryInfo, err error) {
	addCache := true
	res, err = dao.CacheLotteryInfo(c, sid)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("LotteryInfo")
		return
	}
	cache.MetricMisses.Inc("LotteryInfo")
	res, err = dao.RawLotteryInfo(c, sid)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryInfo(c, sid, miss)
	})
	return
}

// LotteryTimesConf get data from cache if miss will call source method, then add to cache.
func (dao *Dao) LotteryTimesConfig(c context.Context, sid string) (res []*l.LotteryTimesConfig, err error) {
	addCache := true
	res, err = dao.CacheLotteryTimesConfig(c, sid)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("LotteryTimesConfig")
		return
	}
	cache.MetricMisses.Inc("LotteryTimesConfig")
	res, err = dao.RawLotteryTimesConfig(c, sid)
	if err != nil || len(res) == 0 {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryTimesConfig(c, sid, miss)
	})
	return
}

func (dao *Dao) LotteryGift(c context.Context, sid string) (res []*l.LotteryGift, err error) {
	addCache := true
	res, err = dao.CacheLotteryGift(c, sid)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryGift")
		return
	}
	cache.MetricMisses.Inc("LotteryGift")
	res, err = dao.RawLotteryGift(c, sid)
	if err != nil || len(res) == 0 {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryGift(c, sid, miss)
	})
	return
}

func (dao *Dao) LotteryAddr(c context.Context, id int64, mid int64) (res int64, err error) {
	addCache := true
	res, err = dao.CacheLotteryAddrCheck(c, id, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("LotteryAddr")
		return
	}
	cache.MetricMisses.Inc("LotteryAddr")
	res, err = dao.RawLotteryAddrCheck(c, id, mid)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryAddrCheck(c, id, mid, miss)
	})
	return
}

func (dao *Dao) LotteryUsedTimes(c context.Context, lotteryTimesConfig []*l.LotteryTimesConfig, id int64, mid int64) (res map[string]int, err error) {
	addCache := true
	var (
		total int
		flag  bool
	)
	res, err = dao.CacheLotteryTimes(c, id, mid, lotteryTimesConfig, _usedTimes)
	if err != nil {
		addCache = false
		err = nil
	}
	for _, v := range res {
		if v == 1 {
			flag = true
		}
		total = total + v
	}
	if total != 0 || flag {
		cache.MetricHits.Inc("LotteryUsedTimes")
		return
	}
	cache.MetricMisses.Inc("LotteryUsedTimes")
	var val []*l.LotteryRecordDetail
	val, err = dao.RawLotteryUsedTimes(c, id, mid)
	if err != nil {
		return
	}
	res = make(map[string]int)
	ltcMap := make(map[string]int) //k 类型+id 值 整个活动期/每日
	for _, ltc := range lotteryTimesConfig {
		ltcMap[rsKey(ltc.Type, ltc.ID)] = ltc.AddType
		if ltc.Type == _winType {
			res[rsKey(ltc.Type, ltc.ID)] = -1
		}
	}
	var winTimes int
	if len(val) != 0 {
		// 遍历所有抽奖记录，如果配置的是整个活动，则计入所有该类型的num和，如果配置的是每日，则只计入该类型下当日的num和
		for k, v := range ltcMap {
			// 每日
			if v == _daily {
				nowT := time.Now().Format("2006-01-02")
				timeTemplate := "2006-01-02 15:04:05"
				start, _ := time.ParseInLocation(timeTemplate, nowT+" 00:00:00", time.Local)
				s := start.Unix()
				end, _ := time.ParseInLocation(timeTemplate, nowT+" 23:59:59", time.Local)
				e := end.Unix()
				for _, value := range val {
					if k != rsKey(value.Type, value.CID) {
						continue
					}
					if value.Ctime >= xtime.Time(s) && value.Ctime <= xtime.Time(e) {
						res[k] = res[k] + value.Num
					}
				}
				continue
			}
			// 整个活动期间
			for _, value := range val {
				if value.GiftID > 0 {
					winTimes++
				}
				if k != rsKey(value.Type, value.CID) {
					continue
				}
				res[k] = res[k] + value.Num
			}
		}
	}
	if winTimes > 0 {
		for _, ltc := range lotteryTimesConfig {
			if ltc.Type == _winType {
				res[rsKey(ltc.Type, ltc.ID)] += winTimes
				break
			}
		}
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryTimes(c, id, mid, lotteryTimesConfig, _usedTimes, miss)
	})
	return
}

func (dao *Dao) LotteryAddTimes(c context.Context, lotteryTimesConfig []*l.LotteryTimesConfig, id, mid int64) (res map[string]int, err error) {
	addCache := true
	var (
		total int
		flag  bool
	)
	res, err = dao.CacheLotteryTimes(c, id, mid, lotteryTimesConfig, _addTimes)
	if err != nil {
		addCache = false
		err = nil
	}
	for _, v := range res {
		if v == 1 {
			flag = true
		}
		total = total + v
	}
	if total != 0 || flag {
		cache.MetricHits.Inc("LotteryAddTimes")
		return
	}
	cache.MetricMisses.Inc("LotteryAddTimes")
	val, err := dao.RawLotteryAddTimes(c, id, mid)
	if err != nil {
		return
	}
	res = make(map[string]int)
	ltcMap := make(map[string]int) //k 类型+id 值 整个活动期/每日
	for _, ltc := range lotteryTimesConfig {
		ltcMap[rsKey(ltc.Type, ltc.ID)] = ltc.AddType
		if ltc.Type == _winType {
			res[rsKey(ltc.Type, ltc.ID)] = -1
		}
	}
	// 遍历所有抽奖记录，如果配置的是整个活动，则计入所有该类型的num和，如果配置的是每日，则只计入该类型下当日的num和
	if len(val) != 0 {
		for k, v := range ltcMap {
			// 每日
			if v == _daily {
				nowT := time.Now().Format("2006-01-02")
				timeTemplate := "2006-01-02 15:04:05"
				start, _ := time.ParseInLocation(timeTemplate, nowT+" 00:00:00", time.Local)
				s := start.Unix()
				end, _ := time.ParseInLocation(timeTemplate, nowT+" 23:59:59", time.Local)
				e := end.Unix()
				for _, value := range val {
					if k != rsKey(value.Type, value.CID) {
						continue
					}
					if value.Ctime >= xtime.Time(s) && value.Ctime <= xtime.Time(e) {
						res[k] = res[k] + value.Num
					}
				}
				continue
			}
			// 整个活动期间
			for _, value := range val {
				if k != rsKey(value.Type, value.CID) {
					continue
				}
				res[k] = res[k] + value.Num
			}
		}
	}
	miss := res
	if !addCache {
		return
	}
	dao.AddCacheLotteryTimes(c, id, mid, lotteryTimesConfig, _addTimes, miss)
	return
}

func (dao *Dao) LotteryWinList(c context.Context, sid int64, giftIDs []int64, num int64, needCache bool) (res []*l.GiftList, err error) {
	addCache := true
	if needCache {
		res, err = dao.CacheLotteryWinList(c, sid)
		if err != nil {
			addCache = false
			err = nil
		}
		if len(res) != 0 {
			cache.MetricHits.Inc("LotteryWinList")
			return
		}
	}

	cache.MetricMisses.Inc("LotteryWinList")
	res, err = dao.RawLotteryWinList(c, sid, giftIDs, num)
	if err != nil {
		return
	}
	miss := res
	if len(res) == 0 {
		miss = []*l.GiftList{{GiftID: -1}}
	}
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLotteryWinList(c, sid, miss)
	})
	return
}

// CacheLotteryMcNum get data from mc
func (dao *Dao) CacheLotteryMcNum(c context.Context, sid int64, high, mc int) (res int64, err error) {
	key := lotteryMcNumKey(sid, high, mc)
	var v string
	err = dao.mc.Get(c, key).Scan(&v)
	if err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		prom.BusinessErrCount.Incr("mc:CacheLotteryMcNum")
		log.Errorv(c, log.KV("CacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		prom.BusinessErrCount.Incr("mc:CacheLotteryMcNum")
		log.Errorv(c, log.KV("CacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	res = r
	return
}

// AddCacheLotteryMcNum Set data to mc
func (dao *Dao) AddCacheLotteryMcNum(c context.Context, sid int64, high, mc int, val int64) (err error) {
	key := lotteryMcNumKey(sid, high, mc)
	bs := []byte(strconv.FormatInt(val, 10))
	item := &memcache.Item{Key: key, Value: bs, Expiration: dao.mcLotteryExpire, Flags: memcache.FlagRAW}
	if err = dao.mc.Set(c, item); err != nil {
		prom.BusinessErrCount.Incr("mc:AddCacheLotteryMcNum")
		log.Errorv(c, log.KV("AddCacheLotteryMcNum", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// SendSysMsg send sys msg.
func (dao *Dao) SendSysMsg(c context.Context, uids []int64, mc, title string, context string, ip string) (err error) {
	params := url.Values{}
	params.Set("mc", mc)
	params.Set("title", title)
	params.Set("data_type", "4")
	params.Set("context", context)
	params.Set("mid_list", xstr.JoinInts(uids))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			Status int8   `json:"status"`
			Remark string `json:"remark"`
		} `json:"data"`
	}
	if err = dao.client.Post(c, dao.msgURL, ip, params, &res); err != nil {
		log.Errorc(c, "SendSysMsg d.client.Post(%s) error(%+v)", dao.msgURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "SendSysMsg dao.client.Post(%s,%d)", dao.msgURL+"?"+params.Encode(), res.Code)
		return
	}
	log.Infoc(c, "send msg ok, resdata=%+v", res.Data)
	return
}

func (dao *Dao) MemberCoupon(c context.Context, mid int64, batchToken string) (data string, err error) {
	midStr := strconv.FormatInt(mid, 10)
	orderNo := strconv.FormatInt(time.Now().UnixNano(), 10) + midStr + strconv.FormatInt(rand.Int63n(1000), 10)
	params := url.Values{}
	params.Set("mid", midStr)
	params.Set("batch_token", batchToken)
	params.Set("order_no", orderNo)
	var res struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	if err = dao.client.Post(c, dao.couponURL, "", params, &res); err != nil {
		log.Error("MemberCoupon dao.client.Post(%s) error(%+v)", dao.couponURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "MemberCoupon dao.client.Post(%s,%d)", dao.msgURL+"?"+params.Encode(), res.Code)
		return
	}
	data = res.Data
	return
}

func (dao *Dao) MemberVip(c context.Context, mid int64, batchToken, remark string) (err error) {
	midStr := strconv.FormatInt(mid, 10)
	orderNo := strconv.FormatInt(time.Now().UnixNano(), 10) + midStr + strconv.FormatInt(rand.Int63n(1000), 10)
	params := url.Values{}
	params.Set("mid", midStr)
	params.Set("batch_token", batchToken)
	params.Set("order_no", orderNo)
	params.Set("remark", remark)
	var res struct {
		Code int `json:"code"`
	}
	if err = dao.client.Post(c, dao.vipURL, "", params, &res); err != nil {
		log.Error("MemberVip dao.client.Post(%s) error(%+v)", dao.couponURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "MemberVip dao.client.Post(%s,%d)", dao.vipURL+"?"+params.Encode(), res.Code)
		return
	}
	return
}

func (dao *Dao) UpdatelotteryGiftNumSQL(c context.Context, id int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, _updateLotteryGiftNumSQL, id); err != nil {
		err = errors.Wrap(err, "UpdatelotteryGiftNumSQL:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

func (dao *Dao) LotteryGiftNum(c context.Context, sids []string) (res map[int64]int64, err error) {
	var (
		rows    *sql.Rows
		sidsLen = len(sids)
		args    = make([]interface{}, 0)
		str     []string
	)
	if sidsLen == 0 {
		return
	}
	for _, v := range sids {
		str = append(str, "?")
		args = append(args, v)
	}
	if rows, err = dao.db.Query(c, fmt.Sprintf(_lotteryGiftNumTimingSQL, strings.Join(str, ",")), args...); err != nil {
		err = errors.Wrap(err, "LotteryGiftNum:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		l := &lotterymdl.LotteryGift{}
		if err = rows.Scan(&l.ID, &l.SendNum); err != nil {
			err = errors.Wrap(err, "LotteryGiftNum:rows.Scan")
			return
		}
		res[l.ID] = l.SendNum
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LotteryGiftNum: rows.Err()")
		return
	}
	return
}

func (dao *Dao) LotterySid(c context.Context, nowT int64) (res []string, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _lotterySidSQL, nowT); err != nil {
		err = errors.Wrap(err, "LotterySid:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]string, 0)
	for rows.Next() {
		var str string
		if err = rows.Scan(&str); err != nil {
			err = errors.Wrap(err, "LotteryGiftNum:rows.Scan")
			return
		}
		res = append(res, str)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LotteryGiftNum: rows.Err()")
	}
	return
}

// RawLotteryOrderNo check orderNo
func (dao *Dao) RawLotteryOrderNo(c context.Context, id, mid int64, orderNo string) (res int64, err error) {
	row := dao.db.QueryRow(c, fmt.Sprintf(_lotteryOrderNoCheckSQL, id), mid, orderNo)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryOrderNo:QueryRow")
		}
	}
	return
}

func (dao *Dao) UpdateGiftNum(c context.Context, sid string, incrNum int64) (af int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, _giftNumUpdateSQL, incrNum, sid); err != nil {
		err = errors.Wrap(err, "UpdateLotteryWin:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

func (dao *Dao) LotteryMyWinList(ctx context.Context, id, mid int64) (res []*l.LotteryRecordDetail, err error) {
	addCache := true
	res, err = dao.CacheLotteryActionLog(ctx, id, mid, 0, -1)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryMyWinList")
		return
	}
	cache.MetricMisses.Inc("LotteryMyWinList")
	res, err = dao.RawLotteryUsedTimes(ctx, id, mid)
	if err != nil {
		return
	}
	miss := res
	if len(res) == 0 {
		return
	}
	if !addCache {
		return
	}
	dao.cache.Do(ctx, func(ctx context.Context) {
		dao.AddCacheLotteryActionLog(ctx, id, mid, miss)
	})
	return
}
