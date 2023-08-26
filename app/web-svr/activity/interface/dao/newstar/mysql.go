package newstar

import (
	"context"
	"database/sql"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/newstar"
)

const (
	_newstarAddSQL      = "INSERT IGNORE INTO act_newstar(activity_uid,v_status,mid,inviter_mid) VALUES(?,?,?,?)"
	_newstarCreationSQL = "SELECT id,activity_uid,v_status,mid,inviter_mid,is_name,is_mobile,is_identity,fans_count,up_archives,finish_task,finish_time,ctime FROM act_newstar WHERE mid=? AND activity_uid=?"
	_newstartInviteSQL  = "SELECT id,activity_uid,v_status,mid,inviter_mid,is_name,is_mobile,is_identity,fans_count,up_archives,finish_task,finish_time,ctime FROM act_newstar WHERE inviter_mid=? AND v_status=1 AND activity_uid=? ORDER BY id DESC LIMIT 50"
	_inviteCountSQL     = "SELECT count(*) FROM act_newstar WHERE inviter_mid=? AND v_status=1 AND activity_uid=?"
	_newstarAwardSQL    = "SELECT id,activity_uid,award_type,`condition`,finish_money,invite_money FROM act_newstar_award WHERE is_deleted=0 ORDER BY award_type ASC,`condition` ASC"
)

// JoinNewstar.
func (d *Dao) JoinNewstar(ctx context.Context, ActivityUID string, vStatus, mid, inviterMid int64) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _newstarAddSQL, ActivityUID, vStatus, mid, inviterMid); err != nil {
		log.Error("JoinNewstar error d.db.Exec(%s,%d,%d,%s) error(%v)", ActivityUID, vStatus, mid, inviterMid, err)
		return
	}
	return res.LastInsertId()
}

// RawCreation.
func (d *Dao) RawCreation(c context.Context, ActivityUID string, mid int64) (data *newstar.Newstar, err error) {
	data = new(newstar.Newstar)
	row := d.db.QueryRow(c, _newstarCreationSQL, mid, ActivityUID)
	if err = row.Scan(&data.ID, &data.ActivityUID, &data.VStatus, &data.Mid, &data.InviterMid, &data.IsName, &data.IsMobile, &data.IsIdentity, &data.FansCount, &data.UpArchives, &data.FinishTask, &data.FinishTime, &data.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawCreation QueryRow mid(%d) activityUID(%s) error(%v)", mid, ActivityUID, err)
		}
	}
	return
}

// InviteCount .
func (d *Dao) InviteCount(c context.Context, ActivityUID string, inviterMid int64) (count int64, err error) {
	row := d.db.QueryRow(c, _inviteCountSQL, inviterMid, ActivityUID)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("InviteCount row.Scan mid(%d) activityUID(%s) error(%v)", inviterMid, ActivityUID, err)
		}
	}
	return
}

// RawInvites.
func (d *Dao) RawInvites(c context.Context, ActivityUID string, inviterMid int64) (list []*newstar.Newstar, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _newstartInviteSQL, inviterMid, ActivityUID)
	if err != nil {
		log.Error("RawInvites:d.db.Query(%d) error(%v)", inviterMid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := new(newstar.Newstar)
		if err = rows.Scan(&data.ID, &data.ActivityUID, &data.VStatus, &data.Mid, &data.InviterMid, &data.IsName, &data.IsMobile, &data.IsIdentity, &data.FansCount, &data.UpArchives, &data.FinishTask, &data.FinishTime, &data.Ctime); err != nil {
			log.Error("RawInvites:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, data)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawInvites:rows.Err() error(%v)", err)
	}
	return
}

// RawAwards .
func (d *Dao) RawAwards(c context.Context) (res map[string][]*newstar.NewstarAward, err error) {
	var (
		rows *xsql.Rows
		list []*newstar.NewstarAward
	)
	rows, err = d.db.Query(c, _newstarAwardSQL)
	if err != nil {
		log.Error("RawAwards:d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := new(newstar.NewstarAward)
		if err = rows.Scan(&row.ID, &row.ActivityUID, &row.AwardType, &row.Condition, &row.FinishMoney, &row.InviteMoney); err != nil {
			log.Error("RawAwards:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawAwards:rows.Err() error(%v)", err)
		return
	}
	res = make(map[string][]*newstar.NewstarAward, len(list))
	for _, award := range list {
		item := &newstar.NewstarAward{
			ID:          award.ID,
			ActivityUID: award.ActivityUID,
			AwardType:   award.AwardType,
			Condition:   award.Condition,
			FinishMoney: award.FinishMoney,
			InviteMoney: award.InviteMoney,
		}
		res[award.ActivityUID] = append(res[award.ActivityUID], item)
	}
	return
}
