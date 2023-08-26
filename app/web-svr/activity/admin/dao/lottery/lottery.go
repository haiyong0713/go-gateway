package lottery

import (
	"context"
	"database/sql"
	"fmt"
	"go-gateway/app/web-svr/activity/admin/component"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"

	"github.com/pkg/errors"
)

const (
	_initTimesNum     = 3
	_tableAddTimes    = "act_lottery_addtimes_%d"
	_tableAddress     = "act_lottery_gift_address_%d"
	_tableWin         = "act_lottery_win_%d"
	_tableMemberGroup = "act_lottery_member_group"
	_tableAddtimeslog = "act_lottery_addtimes_batch"
	_add              = "INSERT INTO act_lottery(lottery_id, name,stime,etime,type,author) VALUES(UUID(),?,?,?,?,?)"
	_addNew           = "INSERT INTO act_lottery(id,lottery_id, name,stime,etime,type,author) VALUES(?,?,?,?,?,?,?)"
	_lotDetailByID    = "SELECT id,lottery_id,name,type,state,stime,etime,ctime,mtime,author FROM act_lottery WHERE id=?"
	_lotDetailBySID   = "SELECT id,lottery_id,name,is_internal,type,state,stime,etime,ctime,mtime,author FROM act_lottery WHERE lottery_id=?"
	_initLotDetail    = "INSERT INTO act_lottery_info(sid,fs_ip,level) VALUE(?,?,?)"
	_initTimes        = "INSERT INTO act_lottery_times(sid,type,times,most,add_type) VALUE(?,?,?,?,?)"
	_listTotal        = "SELECT count(1) FROM act_lottery %s"
	_baseList         = "SELECT id,lottery_id,name,type,state,stime,etime,ctime,mtime,author FROM act_lottery %s LIMIT ? OFFSET ?"
	_delete           = "UPDATE act_lottery SET state=1,author=? WHERE id=?"
	_getLotRuleBySID  = "SELECT id,sid,level,regtime_stime,regtime_etime,vip_check,account_check,coin,fs_ip,gift_rate,sender_id," +
		"high_type,high_rate,state,activity_link,figure_score,spy_score FROM act_lottery_info WHERE sid=?"
	_allTimesConf                = "SELECT id,sid,type,info,times,add_type,most,state,ctime,mtime FROM act_lottery_times WHERE sid=? AND state=0"
	_allGift                     = "SELECT id,sid,name,num,send_num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,extra,ctime,mtime FROM act_lottery_gift WHERE sid=?"
	_ruleUpdate                  = "UPDATE act_lottery_info SET level=?,regtime_stime=?,regtime_etime=?,vip_check=?,account_check=?,coin=?,fs_ip=?,high_type=?,high_rate=?,gift_rate=?,sender_id=?,activity_link=? WHERE id=?"
	_timesAddBatchPre            = "INSERT INTO act_lottery_times(sid,type,info,times,add_type,most) VALUES%s"
	_timesAddBatchValues         = "(?,?,?,?,?,?)"
	_timesUpdate                 = "UPDATE act_lottery_times SET info=?,times=?,add_type=?,most=?,state=? WHERE id=?"
	_giftAdd                     = "INSERT INTO act_lottery_gift(sid,name,num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,params,member_group,day_num,probability,extra) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_giftEdit                    = "UPDATE act_lottery_gift SET name=?,num=?,type=?,source=?,is_show=?,least_mark=?,efficient=?,time_limit=?,msg_title=?,msg_content=?,img_url=?,params=?,member_group=?,day_num=?,probability=?,extra=? WHERE id=?"
	_giftTotal                   = "SELECT count(id) FROM act_lottery_gift WHERE sid=? %s"
	_giftList                    = "SELECT id,sid,name,num,send_num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,ctime,mtime FROM act_lottery_gift WHERE sid=? %s ORDER BY ? DESC LIMIT ? OFFSET ?"
	_giftWinTotal                = "SELECT count(id) FROM %s WHERE gift_id=? AND mid!=0"
	_giftWinList                 = "SELECT a.id,a.mid,a.gift_id,a.cdkey,a.ctime,a.mtime,IFNULL(b.address_id,0) FROM %s AS a LEFT JOIN %s AS b ON a.mid=b.mid WHERE a.gift_id=? AND a.mid!=0 LIMIT ? OFFSET ?"
	_giftWinListAll              = "SELECT a.id,a.mid,a.gift_id,a.cdkey,a.ctime,a.mtime,IFNULL(b.address_id,0) FROM %s AS a LEFT JOIN %s AS b ON a.mid=b.mid WHERE a.gift_id=? AND a.mid!=0"
	_giftUpload                  = "INSERT INTO %s(gift_id,cdkey) VALUES%s"
	_giftDetailByID              = "SELECT id,sid,name,num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,state,ctime,mtime FROM act_lottery_gift WHERE id=?"
	_updateGiftEffect            = "UPDATE act_lottery_gift SET efficient=? WHERE id=?"
	_giftTaskCheck               = "SELECT id,sid,time_limit,type FROM act_lottery_gift WHERE time_limit>mtime AND efficient=0"
	_uploadStatusUpdate          = "UPDATE act_lottery_gift SET upload=? WHERE id=?"
	_updateLotInfo               = "UPDATE act_lottery SET name=?,is_internal=?,stime=?,etime=?,author=? WHERE id=?"
	_checkAction                 = "SELECT id,sid,type,info,times,add_type,most,state,ctime,mtime FROM act_lottery_times WHERE type=? AND info=? AND state=0"
	_countUpload                 = "SELECT count(1) FROM %s WHERE gift_id=?"
	_leastMarkCheckList          = "SELECT id,sid,name,num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,ctime,mtime FROM act_lottery_gift WHERE sid=? AND least_mark=1"
	_updateOperatorBySID         = "UPDATE act_lottery SET author=? WHERE lottery_id=?"
	_createAction                = "CREATE TABLE IF NOT EXISTS act_lottery_action_%d LIKE act_lottery_action"
	_createAddTimes              = "CREATE TABLE IF NOT EXISTS act_lottery_addtimes_%d LIKE act_lottery_addtimes"
	_createAddress               = "CREATE TABLE IF NOT EXISTS act_lottery_gift_address_%d LIKE act_lottery_gift_address"
	_createWin                   = "CREATE TABLE IF NOT EXISTS act_lottery_win_%d LIKE act_lottery_win"
	_timesByID                   = "SELECT `most` FROM act_lottery_times WHERE id=?"
	_timesBatchAdd               = "INSERT INTO %s(`mid`,`type`,`num`,`cid`,`order_no`) VALUES %s"
	_giftWinListWithoutAid       = "SELECT a.id,a.mid,a.gift_id,a.cdkey,a.ctime,a.mtime,IFNULL(b.address_id,0) FROM %s AS a LEFT JOIN %s AS b ON a.mid=b.mid WHERE a.mid!=0"
	_giftDetailBySid             = "SELECT id,sid,name,type FROM act_lottery_gift WHERE sid=?"
	_memberGroupInsertOrUpdate   = "INSERT INTO %s (`id`,`sid`,`group_name`,`member_group`,`state`) VALUES %s ON DUPLICATE KEY UPDATE id=VALUES(id), sid=VALUES(sid),group_name=VALUES(group_name),member_group=VALUES(member_group),state=VALUES(state)"
	_timesInsertOrUpdate         = "INSERT INTO act_lottery_times (id,sid,type,info,times,add_type,most,state) VALUES %s ON DUPLICATE KEY UPDATE sid=VALUES(sid),type=VALUES(type),info=VALUES(info),times=VALUES(times),add_type=VALUES(add_type),most=VALUES(most),state=VALUES(state)"
	_ruleInsertOrUpdate          = "INSERT INTO act_lottery_info (`id`,`sid`,`level`,`regtime_stime`,`regtime_etime`,`vip_check`,`account_check`,`coin`,`fs_ip`,`high_type`,`high_rate`,`gift_rate`,`sender_id`,`activity_link`,`spy_score`,`figure_score`) VALUES %s ON DUPLICATE KEY UPDATE sid=VALUES(sid),activity_link=VALUES(activity_link),figure_score=VALUES(figure_score),spy_score=VALUES(spy_score),level=VALUES(level),regtime_stime=VALUES(regtime_stime),regtime_etime=VALUES(regtime_etime),vip_check=VALUES(vip_check),account_check=VALUES(account_check),coin=VALUES(coin),fs_ip=VALUES(fs_ip),high_type=VALUES(high_type),high_rate=VALUES(high_rate),gift_rate=VALUES(gift_rate),sender_id=VALUES(sender_id)"
	_giftInsertOrUpdate          = "INSERT INTO act_lottery_gift (`id`,sid,efficient,name,num,type,source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,params,member_group,day_num,probability,extra,upload) VALUES %s ON DUPLICATE KEY UPDATE sid=VALUES(sid),efficient=VALUES(efficient),name=VALUES(name),num=VALUES(num),type=VALUES(type),source=VALUES(source),img_url=VALUES(img_url),time_limit=VALUES(time_limit),msg_title=VALUES(msg_title),msg_content=VALUES(msg_content),is_show=VALUES(is_show),least_mark=VALUES(least_mark),params=VALUES(params),member_group=VALUES(member_group),day_num=VALUES(day_num),probability=VALUES(probability),extra=VALUES(extra)"
	_getMemberGroup              = "SELECT id,sid,group_name,member_group,state,ctime,mtime FROM %s where state = 1 and sid = ?"
	_memberGroupTotal            = "SELECT count(id) FROM %s WHERE sid=? %s"
	_memberGroupList             = "SELECT id,sid,group_name,member_group,state,ctime,mtime FROM %s WHERE sid=? %s ORDER BY ? DESC LIMIT ? OFFSET ?"
	_lotteryUsedTimesSQL         = "SELECT id,mid,num,gift_id,type,ctime,cid FROM act_lottery_action_%d WHERE mid = ? AND state = 0"
	_addTimesBatchLogSQL         = "INSERT INTO act_lottery_addtimes_batch(author,sid,cid) VALUES(?,?,?)"
	_updateTimesBatchLogSQL      = "UPDATE act_lottery_addtimes_batch SET state=?,filename=? WHERE id=?"
	_updateTimesBatchLogStateSQL = "UPDATE act_lottery_addtimes_batch SET state=? WHERE id=?"
	_lotteryTimesSQL             = "SELECT id,sid,type,info,times,add_type,most,state FROM act_lottery_times WHERE id = ? and sid = ? and state = 0"
	_addtimesBatchList           = "SELECT id,sid,cid,author,state,filename,ctime,mtime FROM act_lottery_addtimes_batch %s order by id desc LIMIT ? OFFSET ?"
	_timesBatchSQL               = "SELECT id,sid,cid,author,state,ctime,mtime FROM act_lottery_addtimes_batch where id = ?"
	_addtimesBatchListTotal      = "SELECT count(id) FROM act_lottery_addtimes_batch %s"
	_lotteryAddTimesSQL          = "SELECT id,mid,type,num,cid,ctime FROM act_lottery_addtimes_%d WHERE mid = ? AND state = 0 %s order by id desc LIMIT ? OFFSET ?"
	_lotteryAddTimesTotal        = "SELECT count(id) FROM act_lottery_addtimes_%d WHERE mid=? and state=0 %s"
)

// RawLotteryAddTimes ...
func (d *Dao) RawLotteryAddTimes(c context.Context, id int64, mid, cid int64, pn, ps int) (res []*lotmdl.LotteryAddTimes, err error) {
	var rows *xsql.Rows
	var sqlAdd string
	var arg []interface{}
	arg = append(arg, mid)
	if cid > 0 {
		arg = append(arg, cid)
		sqlAdd = " and cid = ?"
	}
	arg = append(arg, ps, (pn-1)*ps)

	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryAddTimesSQL, id, sqlAdd), arg...); err != nil {
		err = errors.Wrap(err, "RawLotteryAddTimes:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotmdl.LotteryAddTimes, 0)
	for rows.Next() {
		a := &lotmdl.LotteryAddTimes{}
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

// RawLotteryAddTimesTotal get memberGroup total
func (d *Dao) RawLotteryAddTimesTotal(c context.Context, id int64, mid, cid int64) (total int, err error) {
	var (
		arg    []interface{}
		sqlAdd string
	)
	arg = append(arg, mid)
	if cid > 0 {
		arg = append(arg, cid)
		sqlAdd = " and cid = ?"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_lotteryAddTimesTotal, id, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@RawLotteryAddTimesTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// ListTotal get list information total
func (d *Dao) ListTotal(c context.Context, state int, keyword string) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if state != 0 || keyword != "" {
		sqlAdd = "WHERE "
		flag := false
		if state != 0 {
			args = append(args, state-1)
			sqlAdd += "state=? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(name LIKE ? OR lottery_id LIKE ?)"
		}
	}
	result := d.db.QueryRow(c, fmt.Sprintf(_listTotal, sqlAdd), args...)
	if err = result.Scan(&total); err != nil {
		log.Error("lottery@ListTotal result.Scan() failed. error(%v)", err)
	}
	return
}

// BaseList get lottery base information list
func (d *Dao) BaseList(c context.Context, pn, ps, state int, keyword, rank string) (list []*lotmdl.LotInfo, err error) {
	var (
		sqlAdd string
		args   []interface{}
		rows   *xsql.Rows
	)
	if state != 0 || keyword != "" {
		sqlAdd = "WHERE "
		flag := false
		if state != 0 {
			args = append(args, state-1)
			sqlAdd += "state=? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(name LIKE ? OR lottery_id LIKE ?)"
		}
	}
	args = append(args, ps)
	args = append(args, (pn-1)*ps)
	if rank != "" {
		sqlAdd += " ORDER BY " + rank + " DESC"
	} else {
		sqlAdd += " ORDER BY ctime DESC"
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_baseList, sqlAdd), args...); err != nil {
		log.Error("lottery@BaseList d.db.Query() failed. error(%v)", err)
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.LotInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.LotteryID, &tmp.Name, &tmp.Type, &tmp.State, &tmp.STime, &tmp.ETime, &tmp.CTime, &tmp.MTime, &tmp.Author); err != nil {
			log.Errorc(c, "lottery@BaseList rows.Scan() failed. error(%v)", err)
			return
		}
		list = append(list, tmp)
	}
	err = rows.Err()
	return
}

// LotDetailByID
func (d *Dao) LotDetailByID(c context.Context, id int64) (detail *lotmdl.LotInfo, err error) {
	detail = &lotmdl.LotInfo{}
	row := d.db.QueryRow(c, _lotDetailByID, id)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author); err != nil {
		log.Errorc(c, "lottery@LotDetailByID row.Scan() failed. error(%v)", err)
	}
	return
}

// UpdateLotInfo update lottery base information
func (d *Dao) UpdateLotInfo(tx *xsql.Tx, c context.Context, id int64, is_internal int, name, operator string, stime, etime xtime.Time) (err error) {
	if _, err = tx.Exec(_updateLotInfo, name, is_internal, stime, etime, operator, id); err != nil {
		log.Errorc(c, "lottery@UpdateLotInfo() INSERT act_lottery failed. error(%v)", err)
	}
	return
}

// LotDetailBySID
func (d *Dao) LotDetailBySID(c context.Context, sid string) (detail *lotmdl.LotInfo, err error) {
	detail = &lotmdl.LotInfo{}
	row := d.db.QueryRow(c, _lotDetailBySID, sid)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.IsInternal, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			detail = nil
		} else {
			log.Errorc(c, "lottery@LotDetailBySID row.Scan() failed. error(%v)", err)
		}
	}

	return
}

// LotDetailBySIDTx ...
func (d *Dao) LotDetailBySIDTx(c context.Context, tx *xsql.Tx, sid string) (detail *lotmdl.LotInfo, err error) {
	detail = &lotmdl.LotInfo{}
	row := tx.QueryRow(_lotDetailBySID, sid)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.IsInternal, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author); err != nil {
		log.Errorc(c, "lottery@LotDetailBySID row.Scan() failed. error(%v)", err)
	}

	return
}

// InitLotDetail
func (d *Dao) InitLotDetail(tx *xsql.Tx, c context.Context, lotID string) (err error) {
	if _, err = tx.Exec(_initLotDetail, lotID, lotmdl.FsIPOn, lotmdl.InitLevel); err != nil {
		log.Errorc(c, "lottery@InitLotDetail() INSERT act_lottery_info failed. error(%v)", err)
		return
	}
	if _, err = tx.Exec(_initTimes, lotID, lotmdl.TimesTypeBase, 0, 0, lotmdl.TimesAddTypeAll); err != nil {
		log.Errorc(c, "lottery@InitLotDetail() INSERT act_lottery_times type=1 failed. error(%v)", err)
		return
	}
	if _, err = tx.Exec(_initTimes, lotID, lotmdl.TimesTypePrice, _initTimesNum, _initTimesNum, lotmdl.TimesAddTypeAll); err != nil {
		log.Errorc(c, "lottery@InitLotDetail() INSERT act_lottery_times type=2 failed. error(%v)", err)
	}
	return
}

// Create create lottery information
func (d *Dao) Create(tx *xsql.Tx, name, operator string, stime, etime xtime.Time, lotType int) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = tx.Exec(_add, name, stime, etime, lotType, operator); err != nil {
		log.Error("lottery@Add d.db.Exec() INSERT failed. error(%v)", err)
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Error("lottery@Add result.LastInsertId() failed. error(%v)", err)
		return
	}
	if err = d.createAction(tx, id); err != nil {
		log.Error("lottery@Add d.createAction(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createAddTimes(tx, id); err != nil {
		log.Error("lottery@Add d.createAddTimes(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createAddress(tx, id); err != nil {
		log.Error("lottery@Add d.createAddress(%d) failed. error(%v)", id, err)
		return
	}

	return
}

// CreateNew create lottery information
func (d *Dao) CreateNew(tx *xsql.Tx, id int64, lotteryID, name, operator string, stime, etime xtime.Time, lotType int) (err error) {

	if err = d.createAction(tx, id); err != nil {
		log.Error("lottery@Add d.createAction(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createAddTimes(tx, id); err != nil {
		log.Error("lottery@Add d.createAddTimes(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createAddress(tx, id); err != nil {
		log.Error("lottery@Add d.createAddress(%d) failed. error(%v)", id, err)
		return
	}
	if err = d.createWin(tx, id); err != nil {
		log.Error("lottery@Add d.createWin(%d) failed. error(%v)", id, err)
		return
	}

	if _, err = tx.Exec(_addNew, id, lotteryID, name, stime, etime, lotType, operator); err != nil {
		log.Error("lottery@Add d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	return
}

func (d *Dao) createAction(tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(_createAction, id)); err != nil {
		log.Error("lottery@createAction CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *Dao) createAddTimes(tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(_createAddTimes, id)); err != nil {
		log.Error("lottery@createAddTimes CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *Dao) createAddress(tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(_createAddress, id)); err != nil {
		log.Error("lottery@createAddress CREATE TABLE failed. error(%v)", err)
	}
	return
}

func (d *Dao) CreateWin(tx *xsql.Tx, id int64) (err error) {
	return d.createWin(tx, id)
}

func (d *Dao) createWin(tx *xsql.Tx, id int64) (err error) {
	if _, err = tx.Exec(fmt.Sprintf(_createWin, id)); err != nil {
		log.Error("lottery@createWin CREATE TABLE failed. error(%v)", err)
	}
	return
}

// Delete update base lottery state=1
func (d *Dao) Delete(tx *xsql.Tx, c context.Context, id int64, operator string) (err error) {
	if _, err = tx.Exec(_delete, operator, id); err != nil {
		log.Errorc(c, "lottery@Delete tx.Exec() failed. error(%v)", err)
	}
	return
}

// RawLotteryUsedTimes ...
func (d *Dao) RawLotteryUsedTimes(c context.Context, id int64, mid int64) (res []*lotmdl.RecordDetail, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_lotteryUsedTimesSQL, id), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryUsedTimes:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lotmdl.RecordDetail, 0)
	for rows.Next() {
		l := &lotmdl.RecordDetail{}
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

// GetLotRuleBySID  ...
func (d *Dao) GetLotRuleBySID(c context.Context, sid string) (result *lotmdl.RuleInfo, err error) {
	row := d.db.QueryRow(c, _getLotRuleBySID, sid)
	result = &lotmdl.RuleInfo{}
	if err = row.Scan(&result.ID, &result.Sid, &result.Level, &result.RegtimeStime, &result.RegtimeEtime, &result.VipCheck, &result.AccountCheck,
		&result.Coin, &result.FsIP, &result.GiftRate, &result.SenderMid, &result.HighType, &result.HighRate, &result.State, &result.ActivityLink, &result.FigureScore, &result.SpyScore); err != nil {
		log.Errorc(c, "lottery@GetLotRuleBySID row.Scan() failed. error(%v)", err)
	}
	return
}

// AllTimesConf get all times config by sid
func (d *Dao) AllTimesConf(c context.Context, sid string) (result []*lotmdl.TimesConf, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _allTimesConf, sid); err != nil {
		log.Errorc(c, "lottery@AllTimesConf d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.TimesConf{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Type, &tmp.Info, &tmp.Times, &tmp.AddType, &tmp.Most,
			&tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllTimesConf rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllGift get all gift config by sid
func (d *Dao) AllGift(c context.Context, sid string) (result []*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _allGift, sid); err != nil {
		log.Errorc(c, "lottery@AllGift d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.SendNum, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Extra, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllGift rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// RuleUpdate edit lottery rule information
func (d *Dao) RuleUpdate(tx *xsql.Tx, c context.Context, rule *lotmdl.RuleInfo) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(_ruleUpdate, rule.Level, rule.RegtimeStime, rule.RegtimeEtime, rule.VipCheck, rule.AccountCheck, rule.Coin,
		rule.FsIP, rule.HighType, rule.HighRate, rule.GiftRate, rule.SenderMid, rule.ActivityLink, rule.ID); err != nil {
		log.Errorc(c, "lottery@RuleUpdate tx.Exec UPDATE act_lottery_info failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TimesAddBatch INSERT INTO act_lottery_times
func (d *Dao) TimesAddBatch(tx *xsql.Tx, c context.Context, arr []*lotmdl.TimesConf) (r int64, err error) {
	var (
		res   sql.Result
		value string
		arg   []interface{}
	)
	for i, item := range arr {
		if i == 0 {
			value += _timesAddBatchValues
		} else {
			value += "," + _timesAddBatchValues
		}
		arg = append(arg, item.Sid)
		arg = append(arg, item.Type)
		arg = append(arg, item.Info)
		arg = append(arg, item.Times)
		arg = append(arg, item.AddType)
		arg = append(arg, item.Most)
	}
	if res, err = tx.Exec(fmt.Sprintf(_timesAddBatchPre, value), arg...); err != nil {
		log.Errorc(c, "lottery@TimesAddBatch INSERT batch failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TimesUpdateBatch update act_lottery_times batch
func (d *Dao) TimesUpdateBatch(tx *xsql.Tx, c context.Context, arr []*lotmdl.TimesConf) (r int64, err error) {
	var (
		res    sql.Result
		effect int64
	)
	for _, item := range arr {
		if res, err = tx.Exec(_timesUpdate, item.Info, item.Times, item.AddType, item.Most, item.State, item.ID); err != nil {
			log.Errorc(c, "lottery@TimesUpdateBatch tx.Exec() failed. error(%v)", err)
			return
		}
		effect, err = res.RowsAffected()
		r += effect
	}
	return
}

// GiftAdd INSERT INTO act_lottery_gift
func (d *Dao) GiftAdd(tx *xsql.Tx, c context.Context, sid, name, source, msgTitle, msgContent, imgUrl, params, memberGroup, dayNum string, num, giftType, probability int, extra string, timeLimit xtime.Time) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(_giftAdd, sid, name, num, giftType, source, imgUrl, timeLimit, msgTitle, msgContent, lotmdl.GiftShow, lotmdl.GiftLeastMarkN, params, memberGroup, dayNum, probability, extra); err != nil {
		log.Errorc(c, "lottery@GiftAdd tx.Exec() INSERT failed. error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// GiftEdit UPDATE act_lottery_gift
func (d *Dao) GiftEdit(tx *xsql.Tx, c context.Context, id int64, name, source, msgTitle, msgContent, imgURL, params, memberGroup, dayNum string, num, giftType, show, leastMark, effect, probability int, extra string, timeLimit xtime.Time) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(_giftEdit, name, num, giftType, source, show, leastMark, effect, timeLimit, msgTitle, msgContent, imgURL, params, memberGroup, dayNum, probability, extra, id); err != nil {
		log.Errorc(c, "lottery@GiftEdit tx.Exec() UPDATE failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// GiftTotal get gift total
func (d *Dao) GiftTotal(c context.Context, sid string, state, giftType int) (total int, err error) {
	var (
		sqlAdd string
		arg    []interface{}
	)
	arg = append(arg, sid)
	if state != 0 {
		sqlAdd += "AND state=? "
		arg = append(arg, state-1)
	}
	if giftType != 0 {
		sqlAdd += "AND type=? "
		arg = append(arg, giftType)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_giftTotal, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@GiftTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// MemberGroupTotal get memberGroup total
func (d *Dao) MemberGroupTotal(c context.Context, sid string, state int) (total int, err error) {
	var (
		sqlAdd string
		arg    []interface{}
	)
	arg = append(arg, sid)
	sqlAdd += "AND state=? "
	arg = append(arg, state)
	row := d.db.QueryRow(c, fmt.Sprintf(_memberGroupTotal, _tableMemberGroup, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@MemberGroupTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// GiftList get gift list
func (d *Dao) GiftList(c context.Context, sid, rank string, state, giftType, pn, ps int) (result []*lotmdl.GiftInfo, err error) {
	var (
		sqlAdd string
		arg    []interface{}
		rows   *xsql.Rows
	)
	arg = append(arg, sid)
	if state != 0 {
		sqlAdd += "AND state=? "
		arg = append(arg, state-1)
	}
	if giftType != 0 {
		sqlAdd += "AND type=? "
		arg = append(arg, giftType)
	}
	arg = append(arg, rank)
	arg = append(arg, ps)
	arg = append(arg, (pn-1)*ps)
	if rows, err = d.db.Query(c, fmt.Sprintf(_giftList, sqlAdd), arg...); err != nil {
		log.Errorc(c, "lottery@GiftList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.SendNum, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@GiftList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// MemberGroupList get membergroup list
func (d *Dao) MemberGroupList(c context.Context, sid, rank string, state, pn, ps int) (result []*lotmdl.MemberGroupDB, err error) {
	var (
		sqlAdd string
		arg    []interface{}
		rows   *xsql.Rows
	)
	arg = append(arg, sid)
	sqlAdd += "AND state=? "
	arg = append(arg, state)
	arg = append(arg, rank)
	arg = append(arg, ps)
	arg = append(arg, (pn-1)*ps)
	if rows, err = d.db.Query(c, fmt.Sprintf(_memberGroupList, _tableMemberGroup, sqlAdd), arg...); err != nil {
		log.Errorc(c, "lottery@MemberGroupList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.MemberGroupDB{}
		if err = rows.Scan(&tmp.ID, &tmp.SID, &tmp.Name, &tmp.Group, &tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@GiftList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// GiftWinTotal get gift win total
func (d *Dao) GiftWinTotal(c context.Context, id, giftID int64) (total int, err error) {
	var (
		row   *xsql.Row
		table = fmt.Sprintf(_tableWin, id)
	)
	row = d.db.QueryRow(c, fmt.Sprintf(_giftWinTotal, table), giftID)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@GiftWinTotal d.db.Query() failed. error(%v)", err)
	}
	return
}

// GiftWinList get gift win list
func (d *Dao) GiftWinList(c context.Context, id, giftID int64, pn, ps int) (result []*lotmdl.GiftWinInfo, err error) {
	var (
		rows      *xsql.Rows
		tableWin  = fmt.Sprintf(_tableWin, id)
		tableAddr = fmt.Sprintf(_tableAddress, id)
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_giftWinList, tableWin, tableAddr), giftID, ps, (pn-1)*ps); err != nil {
		log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftWinInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Mid, &tmp.GiftId, &tmp.CDKey, &tmp.CTime, &tmp.MTime, &tmp.GiftAddrID); err != nil {
			log.Errorc(c, "lottery@GiftWinList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// GiftUpload .
func (d *Dao) GiftUpload(tx *xsql.Tx, c context.Context, lotID, aid int64, keys []string) (err error) {
	var (
		table  = fmt.Sprintf(_tableWin, lotID)
		sqlAdd string
		arg    []interface{}
	)
	for i, item := range keys {
		if i == 0 {
			sqlAdd += "(?,?)"
		} else {
			sqlAdd += ",(?,?)"
		}
		arg = append(arg, aid)
		arg = append(arg, item)
	}
	if _, err = tx.Exec(fmt.Sprintf(_giftUpload, table, sqlAdd), arg...); err != nil {
		log.Errorc(c, "lottery@GiftUpload tx.Exec() batch INSERT failed. error(%v)", err)
	}
	return
}

// GiftWinListAll get gift win list all
func (d *Dao) GiftWinListAll(c context.Context, id, giftID int64) (result []*lotmdl.GiftWinInfo, err error) {
	var (
		rows      *xsql.Rows
		tableWin  = fmt.Sprintf(_tableWin, id)
		tableAddr = fmt.Sprintf(_tableAddress, id)
	)
	if rows, err = component.ExportDB.Query(c, fmt.Sprintf(_giftWinListAll, tableWin, tableAddr), giftID); err != nil {
		log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftWinInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Mid, &tmp.GiftId, &tmp.CDKey, &tmp.CTime, &tmp.MTime, &tmp.GiftAddrID); err != nil {
			log.Errorc(c, "lottery@GiftWinList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// GiftDetailByID
func (d *Dao) GiftDetailByID(c context.Context, id int64) (giftInfo *lotmdl.GiftInfo, err error) {
	row := d.db.QueryRow(c, _giftDetailByID, id)
	giftInfo = &lotmdl.GiftInfo{}
	if err = row.Scan(&giftInfo.ID, &giftInfo.Sid, &giftInfo.Name, &giftInfo.Num, &giftInfo.Type, &giftInfo.Source, &giftInfo.ImgURL, &giftInfo.TimeLimit,
		&giftInfo.MessageTitle, &giftInfo.MessageContent, &giftInfo.IsShow, &giftInfo.LeastMark, &giftInfo.State, &giftInfo.Ctime, &giftInfo.Mtime); err != nil {
		log.Errorc(c, "lottery@GiftWinList row.Scan() failed. error(%v)", err)
	}
	return
}

// GiftDetailByIDTx ...
func (d *Dao) GiftDetailByIDTx(c context.Context, tx *xsql.Tx, id int64) (giftInfo *lotmdl.GiftInfo, err error) {
	row := tx.QueryRow(_giftDetailByID, id)
	giftInfo = &lotmdl.GiftInfo{}
	if err = row.Scan(&giftInfo.ID, &giftInfo.Sid, &giftInfo.Name, &giftInfo.Num, &giftInfo.Type, &giftInfo.Source, &giftInfo.ImgURL, &giftInfo.TimeLimit,
		&giftInfo.MessageTitle, &giftInfo.MessageContent, &giftInfo.IsShow, &giftInfo.LeastMark, &giftInfo.State, &giftInfo.Ctime, &giftInfo.Mtime); err != nil {
		log.Errorc(c, "lottery@GiftDetailByIDTx row.Scan() failed. error(%v)", err)
	}
	return
}

// UpdateGiftEffect
func (d *Dao) UpdateGiftEffect(c context.Context, id int64, effect int) (err error) {
	if _, err := d.db.Exec(c, _updateGiftEffect, effect, id); err != nil {
		log.Errorc(c, "lottery@UpdateGiftEffect d.db.Exec() failed. error(%v)", err)
	}
	return
}

// UpdateGiftEffect
func (d *Dao) UpdateGiftEffectTx(c context.Context, tx *xsql.Tx, id int64, effect int) (err error) {
	if _, err := tx.Exec(_updateGiftEffect, effect, id); err != nil {
		log.Errorc(c, "lottery@UpdateGiftEffect d.db.Exec() failed. error(%v)", err)
	}
	return
}

// GiftTaskCheck
func (d *Dao) GiftTaskCheck(c context.Context) (task []*lotmdl.GiftTask, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _giftTaskCheck); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(c, "lottery@GiftTaskCheck d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftTask{}
		if err = rows.Scan(&tmp.ID, &tmp.SID, &tmp.TimeLimit, &tmp.Type); err != nil {
			log.Errorc(c, "lottery@GiftTaskCheck rows.Scan() failed. error(%v)", err)
			return
		}
		task = append(task, tmp)
	}
	err = rows.Err()
	return
}

// UploadStatusUpdate update act_lottery_gift upload .
func (d *Dao) UploadStatusUpdate(c context.Context, status int, id int64) (err error) {
	if _, err := d.db.Exec(c, _uploadStatusUpdate, status, id); err != nil {
		log.Errorc(c, "lottery@UploadStatusUpdate d.db.Exec() failed. error(%v)", err)
	}
	return
}

// CheckAction get action data by type and info
func (d *Dao) CheckAction(c context.Context, actionType int, info string) (result []*lotmdl.TimesConf, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _checkAction, actionType, info); err != nil {
		log.Errorc(c, "lottery@CheckAction d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.TimesConf{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Type, &tmp.Info, &tmp.Times, &tmp.AddType, &tmp.Most,
			&tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@CheckAction rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// CountUploadTx  .
func (d *Dao) CountUploadTx(c context.Context, tx *xsql.Tx, lotID int64, giftID int64) (count int, err error) {
	var (
		table = fmt.Sprintf(_tableWin, lotID)
	)
	result := tx.QueryRow(fmt.Sprintf(_countUpload, table), giftID)
	if err = result.Scan(&count); err != nil {
		log.Error("lottery@CountUploadTx result.Scan() failed. error(%v)", err)
	}
	return
}

// CountUpload .
func (d *Dao) CountUpload(c context.Context, lotID int64, giftID int64) (count int, err error) {
	var (
		table = fmt.Sprintf(_tableWin, lotID)
	)
	result := d.db.QueryRow(c, fmt.Sprintf(_countUpload, table), giftID)
	if err = result.Scan(&count); err != nil {
		log.Error("lottery@ListTotal result.Scan() failed. error(%v)", err)
	}
	return
}

// LeastMarkCheckList .
func (d *Dao) LeastMarkCheckList(c context.Context, sid string) (result []*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _leastMarkCheckList, sid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(c, "lottery@CheckAction d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllGift rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// UpdateOperatorBySID edit lottery operator
func (d *Dao) UpdateOperatorBySID(c context.Context, sid, operator string) (err error) {
	if _, err := d.db.Exec(c, _updateOperatorBySID, operator, sid); err != nil {
		log.Errorc(c, "lottery@UpdateOperatorBySID d.db.Exec() failed. error(%v)", err)
	}
	return
}

// RawTimesByID get times by id
func (d *Dao) RawTimesByID(c context.Context, id int64) (res int64, err error) {
	result := d.db.QueryRow(c, _timesByID, id)
	if err = result.Scan(&res); err != nil {
		log.Error("lottery@RawTimesByID result.Scan() failed. error(%v)", err)
	}
	return
}

// UpdateLotInfo update lottery base information
func (d *Dao) BatchAddLotTimes(c context.Context, id, times, cid int64, mids []int64, orderNo string) (err error) {
	var (
		sqls = make([]string, 0, len(mids))
		args = make([]interface{}, 0, len(mids)*2)
	)
	if len(mids) == 0 {
		return
	}
	for _, v := range mids {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, v, 7, times, cid, orderNo)
	}
	_, err = d.db.Exec(c, fmt.Sprintf(_timesBatchAdd, fmt.Sprintf(_tableAddTimes, id), strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("BatchAddExtend:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

func lotteryTimesKey(sid, mid int64, remark string) string {
	return fmt.Sprintf("lottery_times_%s_%d_%d", remark, sid, mid)
}

func (d *Dao) DeleteLotteryTimesCache(c context.Context, sid, mid int64) (err error) {
	var (
		key  = lotteryTimesKey(sid, mid, "add")
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		err = errors.Wrap(err, "conn.Do()")
	}
	return
}

// GiftWinListWithoutAid get gift win list all without aid
func (d *Dao) GiftWinListWithoutAid(c context.Context, id int64) (result []*lotmdl.GiftWinInfo, err error) {
	var (
		rows      *xsql.Rows
		tableWin  = fmt.Sprintf(_tableWin, id)
		tableAddr = fmt.Sprintf(_tableAddress, id)
	)
	if rows, err = component.ExportDB.Query(c, fmt.Sprintf(_giftWinListWithoutAid, tableWin, tableAddr)); err != nil {
		log.Errorc(c, "lottery@GiftWinListWithoutAid d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftWinInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Mid, &tmp.GiftId, &tmp.CDKey, &tmp.CTime, &tmp.MTime, &tmp.GiftAddrID); err != nil {
			log.Errorc(c, "lottery@GiftWinList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// GiftDetailBySid ...
func (d *Dao) GiftDetailBySid(c context.Context, sid string) (giftInfo map[int64]*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _giftDetailBySid, sid); err != nil {
		err = errors.Wrap(err, "d.db.Query()")
		return
	}
	defer rows.Close()
	giftInfo = make(map[int64]*lotmdl.GiftInfo)
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Type); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		giftInfo[tmp.ID] = tmp
	}
	err = rows.Err()
	return
}

// BatchInsertOrUpdateMemberGroup batch insert or update  membergroup
func (d *Dao) BatchInsertOrUpdateMemberGroup(c context.Context, tx *xsql.Tx, sid string, memberGroup []*lotmdl.MemberGroupDB) (err error) {
	var (
		sqls = make([]string, 0, len(memberGroup))
		args = make([]interface{}, 0)
	)
	if len(memberGroup) == 0 {
		return
	}
	for _, v := range memberGroup {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, v.ID, sid, v.Name, v.Group, v.State)
	}
	_, err = tx.Exec(fmt.Sprintf(_memberGroupInsertOrUpdate, _tableMemberGroup, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BatchInsertOrUpdateMemberGroup:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BatchInsertOrUpdateRules batch insert or update  membergroup
func (d *Dao) BatchInsertOrUpdateRules(c context.Context, tx *xsql.Tx, rules *lotmdl.RuleInfo) (err error) {
	var (
		sqls = make([]string, 0)
		args = make([]interface{}, 0)
	)
	sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	args = append(args, rules.ID, rules.Sid, rules.Level, rules.RegtimeStime, rules.RegtimeEtime, rules.VipCheck, rules.AccountCheck, rules.Coin, rules.FsIP, rules.HighType, rules.HighRate, rules.GiftRate, rules.SenderMid, rules.ActivityLink, rules.SpyScore, rules.FigureScore)
	_, err = tx.Exec(fmt.Sprintf(_ruleInsertOrUpdate, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BatchInsertOrUpdateRules:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BatchInsertOrUpdateTimes batch insert or update  membergroup
func (d *Dao) BatchInsertOrUpdateTimes(c context.Context, tx *xsql.Tx, times []*lotmdl.TimesConf) (err error) {
	var (
		sqls = make([]string, 0, len(times))
		args = make([]interface{}, 0)
	)
	if len(times) == 0 {
		return
	}
	for _, v := range times {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?)")
		args = append(args, v.ID, v.Sid, v.Type, v.Info, v.Times, v.AddType, v.Most, v.State)
	}
	_, err = tx.Exec(fmt.Sprintf(_timesInsertOrUpdate, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateTimes:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// BatchInsertOrGift ...
func (d *Dao) BatchInsertOrGift(c context.Context, tx *xsql.Tx, gifts []*lotmdl.GiftInfo) (err error) {
	var (
		sqls = make([]string, 0, len(gifts))
		args = make([]interface{}, 0)
	)
	if len(gifts) == 0 {
		return
	}
	for _, v := range gifts {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, v.ID, v.Sid, v.Effect, v.Name, v.Num, v.Type, v.Source, v.ImgURL, v.TimeLimit, v.MessageTitle, v.MessageContent, v.IsShow, v.LeastMark, v.Params, v.MemberGroup, v.DayNum, v.ProbabilityI, v.Extra, v.Upload)
	}
	_, err = tx.Exec(fmt.Sprintf(_giftInsertOrUpdate, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BatchInsertOrGift:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return

}

// AllMemberGroup get all gift config by sid
func (d *Dao) AllMemberGroup(c context.Context, sid string) (result []*lotmdl.MemberGroupDB, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_getMemberGroup, _tableMemberGroup), sid); err != nil {
		log.Errorc(c, "lottery@AllMemberGroup d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.MemberGroupDB{}
		if err = rows.Scan(&tmp.ID, &tmp.SID, &tmp.Name, &tmp.Group, &tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllMemberGroup rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// BatchAddTimesLog 批量增加抽奖次数记录
func (d *Dao) BatchAddTimesLog(c context.Context, sid, operator string, cid int64) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(c, _addTimesBatchLogSQL, operator, sid, cid); err != nil {
		log.Errorc(c, "lottery@BatchAddTimesLog d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Errorc(c, "lottery@Add result.LastInsertId() failed. error(%v)", err)
		return
	}
	return
}

// UpdateBatchAddTimesLog ...
func (d *Dao) UpdateBatchAddTimesLog(c context.Context, id int64, state int, url string) (err error) {
	if _, err := d.db.Exec(c, _updateTimesBatchLogSQL, state, url, id); err != nil {
		log.Errorc(c, "lottery@_updateTimesBatchLogSQL d.db.Exec() failed. error(%v)", err)
	}
	return
}

// UpdateBatchAddTimesLog ...
func (d *Dao) UpdateBatchAddTimesLogState(c context.Context, id int64, state int) (err error) {
	if _, err := d.db.Exec(c, _updateTimesBatchLogStateSQL, state, id); err != nil {
		log.Errorc(c, "lottery@UpdateBatchAddTimesLogState d.db.Exec() failed. error(%v)", err)
	}
	return
}

// AddTimesLogByID ...
func (d *Dao) AddTimesLogByID(c context.Context, id int64) (tmp *lotmdl.AddTimesLog, err error) {
	tmp = &lotmdl.AddTimesLog{}
	row := d.db.QueryRow(c, _timesBatchSQL, id)
	if err = row.Scan(&tmp.ID, &tmp.Sid, &tmp.Cid, &tmp.Author, &tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
		log.Errorc(c, "lottery@TimesConfigByID row.Scan() failed. error(%v)", err)
	}
	return
}

// TimesConfigByID ...
func (d *Dao) TimesConfigByID(c context.Context, id int64, sid string) (detail *lotmdl.TimesConf, err error) {
	detail = &lotmdl.TimesConf{}
	row := d.db.QueryRow(c, _lotteryTimesSQL, id, sid)
	if err = row.Scan(&detail.ID, &detail.Sid, &detail.Type, &detail.Info, &detail.Times, &detail.AddType, &detail.Most, &detail.State); err != nil {
		log.Errorc(c, "lottery@TimesConfigByID row.Scan() failed. error(%v)", err)
	}
	return
}

// BatchAddTimesLogList ...
func (d *Dao) BatchAddTimesLogList(c context.Context, sid string, pn, ps int) (result []*lotmdl.AddTimesLog, err error) {
	var (
		sqlAdd string
		arg    []interface{}
		rows   *xsql.Rows
	)
	if sid != "" {
		sqlAdd += "where sid = ? "
		arg = append(arg, sid)
	}
	arg = append(arg, ps)
	arg = append(arg, (pn-1)*ps)
	if rows, err = d.db.Query(c, fmt.Sprintf(_addtimesBatchList, sqlAdd), arg...); err != nil {
		log.Errorc(c, "lottery@BatchAddTimesLogList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.AddTimesLog{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Cid, &tmp.Author, &tmp.State, &tmp.FileName, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@BatchAddTimesLogList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// BatchAddTimesLogTotal get gift total
func (d *Dao) BatchAddTimesLogTotal(c context.Context, sid string) (total int, err error) {
	var (
		sqlAdd string
		arg    []interface{}
	)
	if sid != "" {
		sqlAdd += "where sid=? "
		arg = append(arg, sid)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_addtimesBatchListTotal, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@BatchAddTimesLogTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}
