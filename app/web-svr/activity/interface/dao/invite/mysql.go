package invite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"
	mdl "go-gateway/app/web-svr/activity/interface/model/invite"
)

const (
	_addTokenSQL             = "INSERT IGNORE INTO `act_tokens` (activity_uid,mid,token_type,expire_time,token,token_source) VALUES(?,?,?,?,?,?)"
	_selTokenSQL             = "SELECT id,mid,token_type,expire_time,token FROM act_tokens WHERE token=?"
	_addInviteRelationLogSQL = "INSERT INTO act_invite_relation_log(mid,activity_uid,tel,token,invited_time) VALUES (?,?,?,?,?)"
	_getInviteByTelHashSQL   = "SELECT mid,invited_mid,invited_time,expire_time,is_new,is_blocked,activity_uid FROM act_invite_relation WHERE tel_hash=? and activity_uid=? and mid != ? order by id desc LIMIT 1"
	_addInviteRelationSQL    = "INSERT INTO act_invite_relation(mid,activity_uid,tel,tel_hash,token,invited_time,expire_time,ip,invited_mid,is_new,invited_login_time,buvid) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"
	_setInviteRelationSQL    = "UPDATE act_invite_relation SET mid=?,token=?,invited_time=?,expire_time=?,ip=?,ctime=?,activity_uid=? WHERE tel=? AND invited_mid=0"
)

// AddToken add token.
func (d *dao) AddToken(ctx context.Context, mid, tp, expire int64, activityUID, token string, source int64) (int64, error) {
	res, err := d.db.Exec(ctx, _addTokenSQL, activityUID, mid, tp, expire, token, source)
	if err != nil {
		log.Error("AddToken:d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.LastInsertId()
}

const _addUserShareLogSQL = `
	insert into act_user_share_log (mid,activity_uid,first_share_time,last_share_time,first_enter_time)
    values (?,?,?,?,?)
`

func (d *dao) AddUserShareLog(ctx context.Context, mid int64, activityUID string, now time.Time) error {
	query := fmt.Sprintf(_addUserShareLogSQL)
	_, err := d.db.Exec(ctx, query, mid, activityUID, now.Unix(), now.Unix(), now.Unix())
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "Duplicate entry") {
		log.Error("AddUserShareLog d.db.Exec error:%+v", err)
		return err
	}
	// 尝试更新first share time
	rows, err := d.updateFirstShareTime(ctx, mid, activityUID, now)
	if err != nil {
		log.Error("AddUserShareLog d.db.Exec error:%+v", err)
		return err
	}
	if rows > 0 {
		return nil
	}
	// 1. isShareExpire + LastShareTime
	// 2. isShareExpire
	// first share time 已存在，尝试更新是否分享超过时间
	rows, err = d.updateFirstShareExpire(ctx, mid, activityUID, now)
	if err != nil {
		log.Error("AddUserShareLog d.db.Exec error:%+v", err)
		return err
	}
	if rows > 0 {
		return nil
	}
	// 更新最新一次分享时间
	if _, err = d.updateLastShareTime(ctx, mid, activityUID, now); err != nil {
		log.Error("AddUserShareLog d.db.Exec error:%+v", err)
		return err
	}
	return nil
}

const _setFirstShareTimeSQL = `
	update act_user_share_log 
	set first_share_time=?,last_share_time=?
	where mid=? and activity_uid=? and first_share_time=0
`

func (d *dao) updateFirstShareTime(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error) {
	query := fmt.Sprintf(_setFirstShareTimeSQL)
	res, err := d.db.Exec(ctx, query, now.Unix(), now.Unix(), mid, activityUID)
	if err != nil {
		log.Error("updateFirstShareTime d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

const _setLastShareTimeSQL = `
	update act_user_share_log 
	set last_share_time=? 
	where mid=? and activity_uid=?
`

func (d *dao) updateLastShareTime(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error) {
	query := fmt.Sprintf(_setLastShareTimeSQL)
	res, err := d.db.Exec(ctx, query, now.Unix(), mid, activityUID)
	if err != nil {
		log.Error("updateLastShareTime d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

// const last_share_time=0 不更新数据
// NOTE:last_share_time > 0 验证数据库数据last_share_time 有值的一部分
// NOTE:first_share_time>0 and last_share_time =0 验证和修改 数据库数据last_share_time=0 有过分享的
const _setFirstShareExpireSQL = `
	update act_user_share_log 
	set last_share_time=?, is_share_expire=1 
	where mid=? and activity_uid=? and ((last_share_time>0 and (?-last_share_time)>?) or (first_share_time>0 and last_share_time =0 and (?-first_share_time)>? ))
`

// (usl.FirstShareTime > 0 && usl.LastShareTime == 0 && (now-usl.FirstShareTime) > shareExpire)

func (d *dao) updateFirstShareExpire(ctx context.Context, mid int64, activityUID string, now time.Time) (int64, error) {
	if d.firstShareExpire == 0 {
		d.firstShareExpire = 5 * 24 * 3600
	}
	query := fmt.Sprintf(_setFirstShareExpireSQL)
	res, err := d.db.Exec(ctx, query, now.Unix(), mid, activityUID, now.Unix(), d.firstShareExpire, now.Unix(), d.firstShareExpire)
	if err != nil {
		log.Error("updateFirstShareExpire d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

const _addAllInviteSQL = "INSERT IGNORE INTO invite_relation_all_log(mid,activity_uid,tel,token,invited_time,relation_source,invite_status) VALUES (?,?,?,?,?,?,?)"

// AddAllInviteLog.
func (d *dao) AddAllInviteLog(ctx context.Context, param *mdl.AllInviteLog) error {
	_, err := d.db.Exec(ctx, _addAllInviteSQL, param.Mid, param.ActivityUID, param.Tel, param.Token, param.InvitedTime, param.Source, param.InviteStatus)
	if err != nil {
		log.Error("AddInviteRelation d.db.Exec error(%+v)", err)
		return err
	}
	return nil
}

const _updateIsShareExpireSQL = `
	update act_user_share_log set is_share_expire=0 where mid=? and activity_uid=? ;
`

// UpdateIsShareExpire is_share_expire 发完奖励改成0
func (d *dao) UpdateIsShareExpire(ctx context.Context, mid int64, activityUID string) (int64, error) {
	query := fmt.Sprintf(_updateIsShareExpireSQL)
	res, err := d.db.Exec(ctx, query, mid, activityUID)
	if err != nil {
		log.Error("AddUserAwardExt d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

const _getUserShareLogSQL = `
	SELECT id,first_share_time,last_share_time,is_share_expire,first_enter_time,last_enter_time
	FROM act_user_share_log
	WHERE mid=? AND activity_uid=?
`

// UserShareLog .
func (d *dao) UserShareLog(ctx context.Context, mid int64, activityUID string) (*mdl.UserShareLog, error) {
	res := &mdl.UserShareLog{
		Mid:         mid,
		ActivityUID: activityUID,
	}
	query := fmt.Sprintf(_getUserShareLogSQL)
	row := d.db.QueryRow(ctx, query, mid, activityUID)
	if err := row.Scan(&res.ID, &res.FirstShareTime, &res.LastShareTime, &res.IsShareExpire, &res.FirstEnterTime, &res.LastEnterTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("UserShareLog row.Scan error(%+v)", err)
		return nil, err
	}
	return res, nil
}

const (
	_getInviteMidByTelSQL = "SELECT invited_mid, is_blocked FROM act_invite_relation WHERE tel=? LIMIT 1"
)

// GetInviteMidByTel get invite mid by tel.
func (d *dao) GetInviteMidByTel(ctx context.Context, tel string) (res *mdl.InviteRelation, err error) {
	res = &mdl.InviteRelation{}
	row := d.db.QueryRow(ctx, _getInviteMidByTelSQL, tel)
	if err = row.Scan(&res.InvitedMid, &res.IsBlocked); err != nil {
		if err == sql.ErrNoRows {
			res = nil
			err = nil
			return
		}
		log.Error("GetInviteMidByTel row.Scan error(%+v)", err)
	}
	return
}

// SelToken select token.
func (d *dao) SelToken(ctx context.Context, token string) (res *mdl.FiToken, err error) {
	res = &mdl.FiToken{}
	row := d.db.QueryRow(ctx, _selTokenSQL, token)
	if err = row.Scan(&res.ID, &res.Mid, &res.Tp, &res.ExpireTime, &res.Token); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("SelToken:token:%s, error:%+v", token, err)
	}
	return
}

// AddInviteRelationLog add relation log
func (d *dao) AddInviteRelationLog(ctx context.Context, inviteRel *mdl.InviteRelation) error {
	_, err := d.db.Exec(ctx, _addInviteRelationLogSQL, inviteRel.Mid, inviteRel.ActivityUID, inviteRel.Tel, inviteRel.Token, inviteRel.InvitedTime)
	if err != nil {
		log.Error("AddInviteRelationLog d.db.Exec error(%+v)", err)
		return err
	}
	return nil
}

// AddInviteRelation insert into invite_relation
func (d *dao) AddInviteRelation(ctx context.Context, ir *mdl.InviteRelation) error {
	_, err := d.db.Exec(ctx, _addInviteRelationSQL, ir.Mid, ir.ActivityUID, ir.Tel, ir.TelHash, ir.Token, ir.InvitedTime, ir.ExpireTime, ir.IP, ir.InvitedMid, ir.IsNew, ir.InvitedLoginTime, ir.Buvid)
	if err != nil {
		log.Error("AddInviteRelation d.db.Exec error(%+v)", err)
		return err
	}
	return nil
}

// SetInviteRelation update invite_relation
func (d *dao) SetInviteRelation(ctx context.Context, ir *mdl.InviteRelation) (int64, error) {
	res, err := d.db.Exec(ctx, _setInviteRelationSQL, ir.Mid, ir.Token, ir.InvitedTime, ir.ExpireTime, ir.IP, time.Now(), ir.ActivityUID, ir.Tel)
	if err != nil {
		log.Error("SetInviteRelation d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()
}

// GetInviteByTelHash get invite by tel hash.
func (d *dao) GetInviteByTelHash(ctx context.Context, mid int64, activityUID string, telHash string) (res *mdl.InviteRelation, err error) {
	res = &mdl.InviteRelation{TelHash: telHash}
	row := d.db.QueryRow(ctx, _getInviteByTelHashSQL, telHash, activityUID, mid)
	if err = row.Scan(&res.Mid, &res.InvitedMid, &res.InvitedTime, &res.ExpireTime, &res.IsNew, &res.IsBlocked, &res.ActivityUID); err != nil {
		log.Errorc(ctx, "GetInviteByTelHash row.Scan error(%+v)", err)
		return
	}
	return
}
