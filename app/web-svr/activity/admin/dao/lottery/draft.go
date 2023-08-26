package lottery

import (
	"context"
	"database/sql"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"

	"strings"
)

const (
	tableMemberGroupDraft = "act_lottery_member_group_draft"

	initLotDraftDetail   = "INSERT INTO act_lottery_info_draft(sid,fs_ip,info_level) VALUE(?,?,?)"
	initDraftTimes       = "INSERT INTO act_lottery_times_draft(sid,times_type,times,most,add_type) VALUE(?,?,?,?,?)"
	addLotteryDraft      = "INSERT INTO act_lottery_draft(lottery_id, lottery_name,stime,etime,lottery_type,state,author) VALUES(UUID(),?,?,?,?,?,?)"
	lotDraftDetailByID   = "SELECT id,lottery_id,lottery_name,lottery_type,state,stime,etime,ctime,mtime,author,reviewer,can_reviewer,reject_reason,last_audit_time FROM act_lottery_draft WHERE id=?"
	lotDraftDetailBySID  = "SELECT id,lottery_id,lottery_name,is_internal,lottery_type,state,stime,etime,ctime,mtime,author,reviewer,can_reviewer,reject_reason,last_audit_time FROM act_lottery_draft WHERE lottery_id=?"
	updateLotDraftInfo   = "UPDATE act_lottery_draft SET lottery_name=?,is_internal=?,state=?,stime=?,etime=?,author=?,can_reviewer=? WHERE id=?"
	updateLotDraftState  = "UPDATE act_lottery_draft SET state=?,reviewer=?,last_audit_time=? WHERE lottery_id=?"
	getLotRuleDraftBySID = "SELECT id,sid,info_level,regtime_stime,regtime_etime,vip_check,account_check,coin,fs_ip,gift_rate,sender_id," +
		"high_type,high_rate,state,activity_link,spy_score,figure_score FROM act_lottery_info_draft WHERE sid=?"
	ruleDraftUpdate                = "UPDATE act_lottery_info_draft SET info_level=?,regtime_stime=?,regtime_etime=?,vip_check=?,account_check=?,coin=?,fs_ip=?,high_type=?,high_rate=?,gift_rate=?,sender_id=?,activity_link=?,figure_score=?,spy_score=? WHERE id=?"
	timesDraftAddBatchPre          = "INSERT INTO act_lottery_times_draft(sid,times_type,info,times,add_type,most) VALUES%s"
	timesDraftAddBatchValues       = "(?,?,?,?,?,?)"
	timesDraftUpdate               = "UPDATE act_lottery_times_draft SET info=?,times=?,add_type=?,most=?,state=? WHERE id=?"
	giftDraftAdd                   = "INSERT INTO act_lottery_gift_draft(sid,gift_name,num,gift_type,gift_source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,params,member_group,day_num,probability,extra) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	updateOperatorDraftBySID       = "UPDATE act_lottery_draft SET author=?,state=? WHERE lottery_id=?"
	updateRejectReasonDraftBySID   = "UPDATE act_lottery_draft SET reviewer=? , reject_reason=?,state=? WHERE lottery_id=?"
	giftDraftDetailByID            = "SELECT id,sid,gift_name,num,gift_type,gift_source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,state,extra,ctime,mtime FROM act_lottery_gift_draft WHERE id=?"
	giftDraftEdit                  = "UPDATE act_lottery_gift_draft SET gift_name=?,num=?,gift_type=?,gift_source=?,is_show=?,least_mark=?,efficient=?,time_limit=?,msg_title=?,msg_content=?,img_url=?,params=?,member_group=?,day_num=?,probability=?,extra=? WHERE id=?"
	memberGroupInsertOrUpdateDraft = "INSERT INTO %s (`id`,`sid`,`group_name`,`member_group`,`state`) VALUES %s ON DUPLICATE KEY UPDATE id=VALUES(id), sid=VALUES(sid),group_name=VALUES(group_name),member_group=VALUES(member_group),state=VALUES(state)"
	giftDraftTotal                 = "SELECT count(id) FROM act_lottery_gift_draft WHERE sid=? %s"
	giftDraftList                  = "SELECT id,sid,gift_name,num,send_num,gift_type,gift_source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,extra,ctime,mtime FROM act_lottery_gift_draft WHERE sid=? %s ORDER BY ? DESC LIMIT ? OFFSET ?"
	allTimesConfDraft              = "SELECT id,sid,times_type,info,times,add_type,most,state,ctime,mtime FROM act_lottery_times_draft WHERE sid=? AND state=0"
	allTimesConfTxDraft            = "SELECT id,sid,times_type,info,times,add_type,most,state,ctime,mtime FROM act_lottery_times_draft WHERE sid=?"
	allGiftDraft                   = "SELECT id,sid,gift_name,num,send_num,gift_type,gift_source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,extra,upload,ctime,mtime FROM act_lottery_gift_draft WHERE sid=?"
	getMemberGroupDraft            = "SELECT id,sid,group_name,member_group,state,ctime,mtime FROM %s where state = 1 and sid = ?"
	listTotalDraft                 = "SELECT count(1) FROM act_lottery_draft %s"
	baseListDraft                  = "SELECT id,lottery_id,lottery_name,lottery_type,state,stime,etime,ctime,mtime,author,reviewer,can_reviewer,reject_reason,last_audit_time FROM act_lottery_draft %s LIMIT ? OFFSET ?"
	memberGroupDraftTotal          = "SELECT count(id) FROM %s WHERE sid=? %s"
	memberGroupDraftList           = "SELECT id,sid,group_name,member_group,state,ctime,mtime FROM %s WHERE sid=? %s ORDER BY ? DESC LIMIT ? OFFSET ?"
	deleteDraft                    = "UPDATE act_lottery_draft SET state=5,author=? WHERE id=?"
	_leastMarkCheckDraftList       = "SELECT id,sid,gift_name,num,gift_type,gift_source,img_url,time_limit,msg_title,msg_content,is_show,least_mark,efficient,upload,state,params,member_group,day_num,probability,extra,ctime,mtime FROM act_lottery_gift_draft WHERE sid=? AND least_mark=1"
	_updateGiftEffectTx            = "UPDATE act_lottery_gift_draft SET efficient=? WHERE id=?"
	_uploadStatusUpdateDraft       = "UPDATE act_lottery_gift_draft SET upload=? WHERE id=?"
)

// UploadStatusUpdateDraft update act_lottery_gift upload .
func (d *Dao) UploadStatusUpdateDraft(c context.Context, status int, id int64) (err error) {
	if _, err := d.db.Exec(c, _uploadStatusUpdateDraft, status, id); err != nil {
		log.Errorc(c, "lottery@UploadStatusUpdate d.db.Exec() failed. error(%v)", err)
	}
	return
}

// LeastMarkCheckDraftList .
func (d *Dao) LeastMarkCheckDraftList(c context.Context, sid string) (result []*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _leastMarkCheckDraftList, sid); err != nil {
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
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Extra, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllGift rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllMemberGroupDraftTx get all gift config by sid
func (d *Dao) AllMemberGroupDraftTx(c context.Context, tx *xsql.Tx, sid string) (result []*lotmdl.MemberGroupDB, err error) {
	var rows *xsql.Rows
	if rows, err = tx.Query(fmt.Sprintf(getMemberGroupDraft, tableMemberGroupDraft), sid); err != nil {
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

// InitLotDetailDraft ...
func (d *Dao) InitLotDetailDraft(c context.Context, tx *xsql.Tx, lotID string) (err error) {
	if _, err = tx.Exec(initLotDraftDetail, lotID, lotmdl.FsIPOn, lotmdl.InitLevel); err != nil {
		log.Errorc(c, "lottery@InitLotDetailDraft() INSERT act_lottery_info_draft failed. error(%v)", err)
		return
	}
	if _, err = tx.Exec(initDraftTimes, lotID, lotmdl.TimesTypeBase, 0, 0, lotmdl.TimesAddTypeAll); err != nil {
		log.Errorc(c, "lottery@InitLotDetailDraft() INSERT act_lottery_times_draft type=1 failed. error(%v)", err)
		return
	}
	if _, err = tx.Exec(initDraftTimes, lotID, lotmdl.TimesTypePrice, _initTimesNum, _initTimesNum, lotmdl.TimesAddTypeAll); err != nil {
		log.Errorc(c, "lottery@InitLotDetailDraft() INSERT act_lottery_times_draft type=2 failed. error(%v)", err)
	}
	return
}

// DeleteDraft update base lottery state=1
func (d *Dao) DeleteDraft(c context.Context, tx *xsql.Tx, id int64, operator string) (err error) {
	if _, err = tx.Exec(deleteDraft, operator, id); err != nil {
		log.Errorc(c, "lottery@Delete tx.Exec() failed. error(%v)", err)
	}
	return
}

// CreateDraft create lottery draft
func (d *Dao) CreateDraft(c context.Context, tx *xsql.Tx, name, operator string, stime, etime xtime.Time, state, lotType int) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = tx.Exec(addLotteryDraft, name, stime, etime, lotType, state, operator); err != nil {
		log.Errorc(c, "lottery@Add d.db.Exec() INSERT failed. error(%v)", err)
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Errorc(c, "lottery@Add result.LastInsertId() failed. error(%v)", err)
		return
	}

	return
}

// LotDraftDetailBySID ...
func (d *Dao) LotDraftDetailBySID(c context.Context, sid string) (detail *lotmdl.LotInfoDraft, err error) {
	detail = &lotmdl.LotInfoDraft{}
	row := d.db.QueryRow(c, lotDraftDetailBySID, sid)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.IsInternal, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author, &detail.Reviewer, &detail.CanReviewer, &detail.RejectReason, &detail.LastAuditPassTime); err != nil {
		log.Errorc(c, "lottery@LotDraftDetailBySID row.Scan() failed. error(%v)", err)
	}
	return
}

// LotDraftDetailByID ...
func (d *Dao) LotDraftDetailByID(c context.Context, id int64) (detail *lotmdl.LotInfoDraft, err error) {
	detail = &lotmdl.LotInfoDraft{}
	row := d.db.QueryRow(c, lotDraftDetailByID, id)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author, &detail.Reviewer, &detail.CanReviewer, &detail.RejectReason, &detail.LastAuditPassTime); err != nil {
		log.Errorc(c, "lottery@LotDetailByID row.Scan() failed. error(%v)", err)
	}
	return
}

// LotDraftDetailTxByID ...
func (d *Dao) LotDraftDetailTxByID(c context.Context, tx *xsql.Tx, id int64) (detail *lotmdl.LotInfoDraft, err error) {
	detail = &lotmdl.LotInfoDraft{}
	row := tx.QueryRow(lotDraftDetailByID, id)
	if err = row.Scan(&detail.ID, &detail.LotteryID, &detail.Name, &detail.Type, &detail.State, &detail.STime, &detail.ETime, &detail.CTime,
		&detail.MTime, &detail.Author, &detail.Reviewer, &detail.CanReviewer, &detail.RejectReason, &detail.LastAuditPassTime); err != nil {
		log.Errorc(c, "lottery@LotDetailByID row.Scan() failed. error(%v)", err)
	}
	return
}

// UpdateLotDraftInfo update lottery base information
func (d *Dao) UpdateLotDraftInfo(c context.Context, tx *xsql.Tx, id int64, isInternal, state int, name, operator, canReviewer string, stime, etime xtime.Time) (err error) {
	if _, err = tx.Exec(updateLotDraftInfo, name, isInternal, state, stime, etime, operator, canReviewer, id); err != nil {
		log.Errorc(c, "lottery@updateLotDraftInfo() Update act_lottery_draft failed. error(%v)", err)
	}
	return
}

// UpdateLotDraftStatePass update lottery base information
func (d *Dao) UpdateLotDraftStatePass(c context.Context, tx *xsql.Tx, sid string, state int, reviewer string, lastAuditPassTime int64) (err error) {
	if _, err = tx.Exec(updateLotDraftState, state, reviewer, lastAuditPassTime, sid); err != nil {
		log.Errorc(c, "lottery@UpdateLotDraftStatePass() Update act_lottery_draft failed. error(%v)", err)
	}
	return
}

// GetLotRuleDraftBySID ...
func (d *Dao) GetLotRuleDraftBySID(c context.Context, sid string) (result *lotmdl.RuleInfo, err error) {
	row := d.db.QueryRow(c, getLotRuleDraftBySID, sid)
	result = &lotmdl.RuleInfo{}
	if err = row.Scan(&result.ID, &result.Sid, &result.Level, &result.RegtimeStime, &result.RegtimeEtime, &result.VipCheck, &result.AccountCheck,
		&result.Coin, &result.FsIP, &result.GiftRate, &result.SenderMid, &result.HighType, &result.HighRate, &result.State, &result.ActivityLink, &result.SpyScore, &result.FigureScore); err != nil {
		log.Errorc(c, "lottery@GetLotRuleDraftBySID row.Scan() failed. error(%v)", err)
	}
	return
}

// RuleDraftUpdate edit lottery rule information
func (d *Dao) RuleDraftUpdate(c context.Context, tx *xsql.Tx, rule *lotmdl.RuleInfo) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(ruleDraftUpdate, rule.Level, rule.RegtimeStime, rule.RegtimeEtime, rule.VipCheck, rule.AccountCheck, rule.Coin,
		rule.FsIP, rule.HighType, rule.HighRate, rule.GiftRate, rule.SenderMid, rule.ActivityLink, rule.FigureScore, rule.SpyScore, rule.ID); err != nil {
		log.Errorc(c, "lottery@RuleUpdate tx.Exec UPDATE act_lottery_info_draft failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// UpdateGiftEffectDraftTx ...
func (d *Dao) UpdateGiftEffectDraftTx(c context.Context, tx *xsql.Tx, id int64, effect int) (err error) {
	if _, err := tx.Exec(_updateGiftEffectTx, effect, id); err != nil {
		log.Errorc(c, "lottery@UpdateGiftEffectDraftTx d.db.Exec() failed. error(%v)", err)
	}
	return
}

// TimesDraftAddBatch ...
func (d *Dao) TimesDraftAddBatch(c context.Context, tx *xsql.Tx, arr []*lotmdl.TimesConf) (r int64, err error) {
	var (
		res   sql.Result
		value string
		arg   []interface{}
	)
	for i, item := range arr {
		if i == 0 {
			value += timesDraftAddBatchValues
		} else {
			value += "," + timesDraftAddBatchValues
		}
		arg = append(arg, item.Sid)
		arg = append(arg, item.Type)
		arg = append(arg, item.Info)
		arg = append(arg, item.Times)
		arg = append(arg, item.AddType)
		arg = append(arg, item.Most)
	}
	if res, err = tx.Exec(fmt.Sprintf(timesDraftAddBatchPre, value), arg...); err != nil {
		log.Errorc(c, "lottery@TimesAddBatch INSERT batch failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TimesDraftUpdateBatch update act_lottery_times batch
func (d *Dao) TimesDraftUpdateBatch(c context.Context, tx *xsql.Tx, arr []*lotmdl.TimesConf) (r int64, err error) {
	var (
		res    sql.Result
		effect int64
	)
	for _, item := range arr {
		if res, err = tx.Exec(timesDraftUpdate, item.Info, item.Times, item.AddType, item.Most, item.State, item.ID); err != nil {
			log.Errorc(c, "lottery@TimesDraftUpdateBatch tx.Exec() failed. error(%v)", err)
			return
		}
		effect, err = res.RowsAffected()
		r += effect
	}
	return
}

// GiftDraftAdd INSERT INTO act_lottery_gift
func (d *Dao) GiftDraftAdd(c context.Context, tx *xsql.Tx, sid, name, source, msgTitle, msgContent, imgUrl, params, memberGroup, dayNum string, num, giftType, probability int, extra string, timeLimit xtime.Time) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(giftDraftAdd, sid, name, num, giftType, source, imgUrl, timeLimit, msgTitle, msgContent, lotmdl.GiftShow, lotmdl.GiftLeastMarkN, params, memberGroup, dayNum, probability, extra); err != nil {
		log.Errorc(c, "lottery@GiftDraftAdd tx.Exec() INSERT failed. error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// UpdateOperatorDraftBySIDAndState edit lottery operator
func (d *Dao) UpdateOperatorDraftBySIDAndState(c context.Context, sid, operator string, state int) (err error) {
	if _, err := d.db.Exec(c, updateOperatorDraftBySID, operator, state, sid); err != nil {
		log.Errorc(c, "lottery@UpdateOperatorDraftBySID d.db.Exec() failed. error(%v)", err)
	}
	return
}

// UpdateRejectReasonDraftBySID edit lottery operator
func (d *Dao) UpdateRejectReasonDraftBySID(c context.Context, sid, reviewer, rejectReason string, state int) (err error) {
	if _, err := d.db.Exec(c, updateRejectReasonDraftBySID, reviewer, rejectReason, state, sid); err != nil {
		log.Errorc(c, "lottery@UpdateRejectReasonDraftBySID d.db.Exec() failed. error(%v)", err)
	}
	return
}

// GiftDraftDetailByID ...
func (d *Dao) GiftDraftDetailByID(c context.Context, id int64) (giftInfo *lotmdl.GiftInfo, err error) {
	row := d.db.QueryRow(c, giftDraftDetailByID, id)
	giftInfo = &lotmdl.GiftInfo{}
	if err = row.Scan(&giftInfo.ID, &giftInfo.Sid, &giftInfo.Name, &giftInfo.Num, &giftInfo.Type, &giftInfo.Source, &giftInfo.ImgURL, &giftInfo.TimeLimit,
		&giftInfo.MessageTitle, &giftInfo.MessageContent, &giftInfo.IsShow, &giftInfo.LeastMark, &giftInfo.State, &giftInfo.Extra, &giftInfo.Ctime, &giftInfo.Mtime); err != nil {
		log.Errorc(c, "lottery@GiftDraftDetailByID row.Scan() failed. error(%v)", err)
	}
	return
}

// GiftDraftEdit UPDATE act_lottery_gift
func (d *Dao) GiftDraftEdit(c context.Context, tx *xsql.Tx, id int64, name, source, msgTitle, msgContent, imgURL, params, memberGroup, dayNum string, num, giftType, show, leastMark, effect, probability int, extra string, timeLimit xtime.Time) (r int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(giftDraftEdit, name, num, giftType, source, show, leastMark, effect, timeLimit, msgTitle, msgContent, imgURL, params, memberGroup, dayNum, probability, extra, id); err != nil {
		log.Errorc(c, "lottery@GiftEdit tx.Exec() UPDATE failed. error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// BacthInsertOrUpdateMemberGroupDraft batch insert or update  membergroup
func (d *Dao) BacthInsertOrUpdateMemberGroupDraft(c context.Context, tx *xsql.Tx, sid string, memberGroup []*lotmdl.MemberGroupDB) (err error) {
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
	_, err = tx.Exec(fmt.Sprintf(memberGroupInsertOrUpdateDraft, tableMemberGroupDraft, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "BacthInsertOrUpdateMemberGroupDraft:dao.db.Exec(%v) error(%v)", sqls, err)
	}
	return
}

// GiftDraftTotal get gift total
func (d *Dao) GiftDraftTotal(c context.Context, sid string, state, giftType int) (total int, err error) {
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
		sqlAdd += "AND gift_type=? "
		arg = append(arg, giftType)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(giftDraftTotal, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@GiftDraftTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// GiftDraftList get gift list
func (d *Dao) GiftDraftList(c context.Context, sid, rank string, state, giftType, pn, ps int) (result []*lotmdl.GiftInfo, err error) {
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
		sqlAdd += "AND gift_type=? "
		arg = append(arg, giftType)
	}
	arg = append(arg, rank)
	arg = append(arg, ps)
	arg = append(arg, (pn-1)*ps)
	if rows, err = d.db.Query(c, fmt.Sprintf(giftDraftList, sqlAdd), arg...); err != nil {
		log.Errorc(c, "lottery@GiftList d.db.Query() failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.SendNum, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Extra, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@GiftList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllTimesConfDraft get all times config by sid
func (d *Dao) AllTimesConfDraft(c context.Context, sid string) (result []*lotmdl.TimesConf, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, allTimesConfDraft, sid); err != nil {
		log.Errorc(c, "lottery@allTimesConfDraft d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.TimesConf{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Type, &tmp.Info, &tmp.Times, &tmp.AddType, &tmp.Most,
			&tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@allTimesConfDraft rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllTimesConfTxDraft get all times config by sid
func (d *Dao) AllTimesConfTxDraft(c context.Context, tx *xsql.Tx, sid string) (result []*lotmdl.TimesConf, err error) {
	var rows *xsql.Rows
	if rows, err = tx.Query(allTimesConfTxDraft, sid); err != nil {
		log.Errorc(c, "lottery@allTimesConfDraft tx.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.TimesConf{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Type, &tmp.Info, &tmp.Times, &tmp.AddType, &tmp.Most,
			&tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@allTimesConfDraft rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllGiftTxDraft get all gift config by sid
func (d *Dao) AllGiftTxDraft(c context.Context, tx *xsql.Tx, sid string) (result []*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = tx.Query(allGiftDraft, sid); err != nil {
		log.Errorc(c, "lottery@allGiftDraft txQuery() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.SendNum, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Extra, &tmp.Upload, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@allGiftDraft rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllGiftDraft get all gift config by sid
func (d *Dao) AllGiftDraft(c context.Context, sid string) (result []*lotmdl.GiftInfo, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, allGiftDraft, sid); err != nil {
		log.Errorc(c, "lottery@allGiftDraft d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Sid, &tmp.Name, &tmp.Num, &tmp.SendNum, &tmp.Type, &tmp.Source, &tmp.ImgURL, &tmp.TimeLimit, &tmp.MessageTitle,
			&tmp.MessageContent, &tmp.IsShow, &tmp.LeastMark, &tmp.Effect, &tmp.Upload, &tmp.State, &tmp.Params, &tmp.MemberGroup, &tmp.DayNum, &tmp.ProbabilityI, &tmp.Extra, &tmp.Upload, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@allGiftDraft rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// AllMemberGroupDraft get all gift config by sid
func (d *Dao) AllMemberGroupDraft(c context.Context, sid string) (result []*lotmdl.MemberGroupDB, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(getMemberGroupDraft, tableMemberGroupDraft), sid); err != nil {
		log.Errorc(c, "lottery@AllMemberGroupDraft d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.MemberGroupDB{}
		if err = rows.Scan(&tmp.ID, &tmp.SID, &tmp.Name, &tmp.Group, &tmp.State, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(c, "lottery@AllMemberGroupDraft rows.Scan failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}

// ListTotalDraft get list information total
func (d *Dao) ListTotalDraft(c context.Context, state int, keyword string) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if state != lotmdl.LotteryDraftListAll || keyword != "" {
		sqlAdd = "WHERE "
		flag := false
		if state != lotmdl.LotteryDraftListAll {
			args = append(args, state)
			sqlAdd += "state=? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(lottery_name LIKE ? OR lottery_id LIKE ?)"
		}
	}
	result := d.db.QueryRow(c, fmt.Sprintf(listTotalDraft, sqlAdd), args...)
	if err = result.Scan(&total); err != nil {
		log.Errorc(c, "lottery@listTotalDraft result.Scan() failed. error(%v)", err)
	}
	return
}

// BaseListDraft get lottery base information list
func (d *Dao) BaseListDraft(c context.Context, pn, ps, state int, keyword, rank string) (list []*lotmdl.LotInfoDraft, err error) {
	var (
		sqlAdd string
		args   []interface{}
		rows   *xsql.Rows
	)
	if state != lotmdl.LotteryDraftListAll || keyword != "" {
		sqlAdd = "WHERE "
		flag := false
		if state != lotmdl.LotteryDraftListAll {
			args = append(args, state)
			sqlAdd += "state=? "
			flag = true
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += "(lottery_name LIKE ? OR lottery_id LIKE ?)"
		}
	}
	args = append(args, ps)
	args = append(args, (pn-1)*ps)
	if rank != "" {
		sqlAdd += " ORDER BY " + rank + " DESC"
	} else {
		sqlAdd += " ORDER BY id DESC"
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(baseListDraft, sqlAdd), args...); err != nil {
		log.Error("lottery@BaseListDraft d.db.Query() failed. error(%v)", err)
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.LotInfoDraft{}
		if err = rows.Scan(&tmp.ID, &tmp.LotteryID, &tmp.Name, &tmp.Type, &tmp.State, &tmp.STime, &tmp.ETime, &tmp.CTime, &tmp.MTime, &tmp.Author, &tmp.Reviewer, &tmp.CanReviewer, &tmp.RejectReason, &tmp.LastAuditPassTime); err != nil {
			log.Errorc(c, "lottery@BaseListDraft rows.Scan() failed. error(%v)", err)
			return
		}
		list = append(list, tmp)
	}
	err = rows.Err()
	return
}

// MemberGroupDraftTotal get memberGroup total
func (d *Dao) MemberGroupDraftTotal(c context.Context, sid string, state int) (total int, err error) {
	var (
		sqlAdd string
		arg    []interface{}
	)
	arg = append(arg, sid)
	sqlAdd += "AND state=? "
	arg = append(arg, state)
	row := d.db.QueryRow(c, fmt.Sprintf(memberGroupDraftTotal, tableMemberGroupDraft, sqlAdd), arg...)
	if err = row.Scan(&total); err != nil {
		log.Errorc(c, "lottery@MemberGroupDraftTotal d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}

// MemberGroupDraftList get membergroup list
func (d *Dao) MemberGroupDraftList(c context.Context, sid, rank string, state, pn, ps int) (result []*lotmdl.MemberGroupDB, err error) {
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
	if rows, err = d.db.Query(c, fmt.Sprintf(memberGroupDraftList, tableMemberGroupDraft, sqlAdd), arg...); err != nil {
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
