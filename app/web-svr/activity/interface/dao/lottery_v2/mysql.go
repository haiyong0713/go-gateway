package lottery

import (
	"context"
	xsql "database/sql"
	"fmt"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

const (
	_lotterySQL                    = "SELECT id,lottery_id,name,is_internal,stime,etime,ctime,mtime,type,author,state FROM act_lottery WHERE lottery_id = ? and state = 0 and type = 0"
	_lotteryInfoSQL                = "SELECT id,sid,level,regtime_stime,regtime_etime,vip_check,account_check,coin,fs_ip, gift_rate,high_type,high_rate,state,sender_id,activity_link,spy_score,figure_score FROM act_lottery_info WHERE sid = ? and state = 0"
	_lotteryTimesConfSQL           = "SELECT id,sid,type,add_type,times,info,most,state FROM act_lottery_times WHERE sid = ? and state = 0"
	_lotteryGiftSQL                = "SELECT id,sid,ctime,mtime,name,num,type,source,img_url,time_limit,is_show,least_mark,msg_title,msg_content,efficient,send_num,state,params,member_group,probability,day_num,extra FROM act_lottery_gift WHERE sid = ? and state = 0"
	_updateLotteryGiftNumSQL       = "UPDATE act_lottery_gift SET send_num = send_num+? WHERE id = ? AND send_num+? <= num"
	_lotteryGiftNumTimingSQL       = "SELECT id,send_num FROM act_lottery_gift WHERE sid IN (%s)"
	_lotterySidSQL                 = "SELECT lottery_id FROM act_lottery WHERE etime > ?"
	_lotteryAddrSQL                = "SELECT address_id FROM act_lottery_gift_address_%d WHERE mid = ? and state = 0"
	_lotteryWinListSQL             = "SELECT mid,gift_id,ctime FROM act_lottery_win_%d WHERE mid > 0 and gift_id in (%s) ORDER BY mtime DESC LIMIT ?"
	_insertLotteryAddrSQL          = "INSERT INTO act_lottery_gift_address_%d(mid,address_id) VALUES(?,?) ON DUPLICATE KEY UPDATE address_id = ?"
	_insertLotteryAddTimesSQL      = "INSERT INTO act_lottery_addtimes_%d(mid,type,num,cid,ip,order_no) VALUES(?,?,?,?,?,?)"
	_insertLotteryRecordSQL        = "INSERT INTO act_lottery_action_%d(mid,num,gift_id,type,cid,ip) VALUES %s"
	_getLotteryRecordByOrderNoSQL  = "SELECT id,mid,num,gift_id,type,cid,order_no from act_lottery_action_%d where order_no = ?"
	_lotteryAddTimesSQL            = "SELECT id,mid,type,num,cid,ctime FROM act_lottery_addtimes_%d WHERE mid = ? AND state = 0 and `type` != 1"
	_lotteryUsedTimesSQL           = "SELECT id,mid,num,gift_id,type,ctime,cid FROM act_lottery_action_%d WHERE mid = ? AND state = 0 order by ctime desc"
	_updateLotteryWin              = "UPDATE act_lottery_win_%d SET mid = ?, ip = ? WHERE gift_id = ? AND mid = 0 ORDER BY id LIMIT 1"
	_insertLotteryWin              = "INSERT INTO act_lottery_win_%d(mid,gift_id,ip) VALUES(?,?,?)"
	_lotteryWinOneSQL              = "SELECT cdkey FROM act_lottery_win_%d WHERE mid = ? AND gift_id = ? ORDER BY mtime DESC,id DESC"
	_lotteryWinMidSQL              = "SELECT id,mid,gift_id,cdkey,mtime FROM act_lottery_win_%d WHERE mid = ?  ORDER BY mtime DESC,id DESC limit ?,?"
	_lotteryOrderNoCheckSQL        = "SELECT id FROM act_lottery_addtimes_%d WHERE order_no = ? and mid = ?"
	_giftNumUpdateSQL              = "UPDATE act_lottery_gift SET num = num+? WHERE sid=? AND state=0"
	_getAddressURL                 = "/api/basecenter/addr/view"
	_memberCouponURI               = "/x/internal/coupon/allowance/receive"
	_memberVipURI                  = "/x/internal/vip/resources/grant"
	_lotteryMemberGroupSQL         = "SELECT id,sid,group_name,state,member_group,ctime,mtime FROM act_lottery_member_group WHERE sid = ? and state = 1"
	_insertLotteryRecordOrderNoSQL = "INSERT INTO act_lottery_action_%d(mid,num,gift_id,type,cid,ip,order_no) VALUES %s"
)

func (d *dao) RawLotteryActionByOrderNo(c context.Context, sid int64, orderNo string) (res *lottery.InsertRecord, err error) {
	res = new(lottery.InsertRecord)
	row := d.db.QueryRow(c, fmt.Sprintf(_getLotteryRecordByOrderNoSQL, sid), orderNo)
	if err = row.Scan(&res.ID, &res.Mid, &res.Num, &res.GiftID, &res.Type, &res.CID, &res.OrderNo); err != nil {
		log.Errorc(c, "RawLotteryActionByOrderNo err(%v)", err)
		if err == sql.ErrNoRows {
			err = ecode.ActivityOrderNoFindErr
		} else {
			err = errors.Wrap(err, "RawLotteryActionByOrderNo:QueryRow")

		}
	}
	return
}

// RawLottery get lottery by sid
func (d *dao) RawLottery(c context.Context, sid string) (res *lottery.Lottery, err error) {
	res = new(lottery.Lottery)
	row := d.db.QueryRow(c, _lotterySQL, sid)
	if err = row.Scan(&res.ID, &res.LotteryID, &res.Name, &res.IsInternal, &res.Stime, &res.Etime, &res.Ctime, &res.Mtime, &res.Type, &res.Author, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLottery:QueryRow")
		}
	}
	return
}

// RawLotteryInfo get lotteryInfo by sid
func (d *dao) RawLotteryInfo(c context.Context, sid string) (res *lottery.Info, err error) {
	res = new(lottery.Info)
	row := d.db.QueryRow(c, _lotteryInfoSQL, sid)
	if err = row.Scan(&res.ID, &res.Sid, &res.Level, &res.RegTimeStime, &res.RegTimeEtime, &res.VipCheck, &res.AccountCheck, &res.Coin, &res.FsIP, &res.GiftRate, &res.HighType, &res.HighRate, &res.State, &res.SenderID, &res.ActivityLink, &res.SpyScore, &res.FigureScore); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryInfo:QueryRow")
		}
	}
	return
}

// RawLotteryTimesConfig lottery times config
func (d *dao) RawLotteryTimesConfig(c context.Context, sid string) (res []*lottery.TimesConfig, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _lotteryTimesConfSQL, sid); err != nil {
		err = errors.Wrap(err, "RawLotteryTimesConfig:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.TimesConfig, 0)
	for rows.Next() {
		l := &lottery.TimesConfig{}
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

// UpdateLotteryWin ....
func (d *dao) UpdateLotteryWin(c context.Context, id int64, mid int64, giftID int64, ip string) (ef int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_updateLotteryWin, id), mid, ip, giftID); err != nil {
		err = errors.Wrap(err, "UpdateLotteryWin:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// RawLotteryWinOne ...
func (d *dao) RawLotteryWinOne(c context.Context, id, mid, giftID int64) (res string, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_lotteryWinOneSQL, id), mid, giftID)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryWinOne:QueryRow")
		}
	}
	return
}

// RawLotteryGift get lotteryGift by sid
func (d *dao) RawLotteryMidWinList(c context.Context, sid, mid, offset, limit int64) (res []*lottery.MidWinList, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryWinMidSQL, sid), mid, offset, limit); err != nil {
		err = errors.Wrap(err, "RawLotteryMidWinList:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.MidWinList, 0)
	for rows.Next() {
		l := &lottery.MidWinList{}
		if err = rows.Scan(&l.ID, &l.Mid, &l.GiftID, &l.Cdkey, &l.Mtime); err != nil {
			err = errors.Wrap(err, "RawLotteryMidWinList:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryMidWinList:rows.Err()")
	}
	return
}

// RawLotteryGift get lotteryGift by sid
func (d *dao) RawLotteryGift(c context.Context, sid string) (res []*lottery.GiftDB, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _lotteryGiftSQL, sid); err != nil {
		err = errors.Wrap(err, "RawLotteryGift:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.GiftDB, 0)
	for rows.Next() {
		l := &lottery.GiftDB{}
		if err = rows.Scan(&l.ID, &l.Sid, &l.Ctime, &l.Mtime, &l.Name, &l.Num, &l.Type, &l.Source, &l.ImgURL, &l.TimeLimit, &l.IsShow, &l.LeastMark, &l.MessageTitle, &l.MessageContent, &l.Efficient, &l.SendNum, &l.State, &l.Params, &l.MemberGroup, &l.Probability, &l.DayNum, &l.Extra); err != nil {
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

// RawLotteryAddrCheck ...
func (d *dao) RawLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_lotteryAddrSQL, id), mid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryAddrCheck:QueryRow")
		}
	}
	return
}

// InsertLotteryAddr ...
func (d *dao) InsertLotteryAddr(c context.Context, id, mid, addressID int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_insertLotteryAddrSQL, id), mid, addressID, addressID); err != nil {
		err = errors.Wrap(err, "InsertLotteryAddr:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertLotteryAddTimes ...
func (d *dao) InsertLotteryAddTimes(c context.Context, id int64, mid int64, addType, num int, cid int64, ip, orderNo string) (ef int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_insertLotteryAddTimesSQL, id), mid, addType, num, cid, ip, orderNo); err != nil {
		err = errors.Wrap(err, "InsertLotteryAddTimes:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertLotteryRecard ...
func (d *dao) InsertLotteryRecard(c context.Context, id int64, record []*lottery.InsertRecord, gid []int64, ip string) (count int64, err error) {
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
	rows, err := d.db.Exec(c, fmt.Sprintf(_insertLotteryRecordSQL, id, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("InsertLotteryRecard: dao.db.Exec() id(%d) error(%v)", id, err)
		return
	}
	return rows.LastInsertId()
}

// InsertLotteryRecardOrderNo ...
func (d *dao) InsertLotteryRecardOrderNo(c context.Context, id int64, record []*lottery.InsertRecord, gid []int64, ip string) (count int64, err error) {
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
	rows, err := d.db.Exec(c, fmt.Sprintf(_insertLotteryRecordOrderNoSQL, id, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("InsertLotteryRecardOrderNo: dao.db.Exec() id(%d) error(%v)", id, err)
		return
	}
	return rows.LastInsertId()
}

// RawLotteryUsedTimes ...
func (d *dao) RawLotteryUsedTimes(c context.Context, id int64, mid int64) (res []*lottery.RecordDetail, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryUsedTimesSQL, id), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryUsedTimes:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.RecordDetail, 0)
	for rows.Next() {
		l := &lottery.RecordDetail{}
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

// RawLotteryAddTimes ...
func (d *dao) RawLotteryAddTimes(c context.Context, id int64, mid int64) (res []*lottery.AddTimes, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryAddTimesSQL, id), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryAddTimes:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.AddTimes, 0)
	for rows.Next() {
		a := &lottery.AddTimes{}
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

// UpdatelotteryGiftNumSQL ...
func (d *dao) UpdatelotteryGiftNumSQL(c context.Context, id int64, num int) (ef int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, _updateLotteryGiftNumSQL, num, id, num); err != nil {
		err = errors.Wrap(err, "UpdatelotteryGiftNumSQL:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// RawLotteryMemberGroup get lotteryMemberGroup by sid
func (d *dao) RawMemberGroup(c context.Context, sid string) (res []*lottery.MemberGroupDB, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _lotteryMemberGroupSQL, sid); err != nil {
		err = errors.Wrap(err, "RawLotteryMemberGroup:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.MemberGroupDB, 0)
	for rows.Next() {
		l := &lottery.MemberGroupDB{}
		if err = rows.Scan(&l.ID, &l.SID, &l.Name, &l.State, &l.Group, &l.Ctime, &l.Mtime); err != nil {
			err = errors.Wrap(err, "RawLotteryMemberGroup:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLotteryGift:rows.Err()")
	}
	return
}

// InsertLotteryWin ..
func (d *dao) InsertLotteryWin(c context.Context, id, giftID, mid int64, ip string) (ef int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_insertLotteryWin, id), mid, giftID, ip); err != nil {
		err = errors.Wrap(err, "InsertLotteryWin:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// RawLotteryWinList ...
func (d *dao) RawLotteryWinList(c context.Context, id int64, giftIDs []int64, num int64) (res []*lottery.GiftMid, err error) {
	if len(giftIDs) == 0 {
		return
	}
	var rows *sql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryWinListSQL, id, xstr.JoinInts(giftIDs)), num); err != nil {
		err = errors.Wrap(err, "RawLotteryWinList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.GiftMid, 0)
	for rows.Next() {
		l := &lottery.GiftMid{}
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
