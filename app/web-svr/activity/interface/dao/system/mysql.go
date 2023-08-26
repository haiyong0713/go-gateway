package system

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/system"
)

const (
	_userInfoByWorkCodeSQL          = "SELECT /*master*/ `id`, `uid`, `token` FROM system_user WHERE uid = ? limit 1"
	_updateUserInfoSQL              = "UPDATE system_user SET `token` = ? WHERE uid = ?"
	_createUserInfoSQL              = "INSERT INTO system_user (`uid`, `token`) VALUES (?, ?)"
	_userInfoByTokenSQL             = "SELECT /*master*/ `id`, `uid`, `token` FROM system_user WHERE token = ? limit 1"
	_activityInfoByIDSQL            = "SELECT /*master*/ `id`, `name`, `type`, `stime`, `etime`, `config` FROM system_activity WHERE id = ? and state = 0"
	_activitySignedSQL              = "SELECT /*master*/ `id`, `aid`, `uid`, `location` FROM system_activity_sign WHERE aid = ? and uid = ?"
	_createSignSQL                  = "INSERT INTO system_activity_sign (`aid`, `uid`, `location`) VALUES (?, ?, ?)"
	_activitySeatInfoSQL            = "SELECT /*master*/ `id`, `aid`, `uid`, `content` FROM system_activity_seat WHERE aid = ? and uid = ?"
	_activityExtraUsersInfoSQL      = "SELECT /*master*/ `id`, `avatar`, `last_name`, `login_id`, `nick_name`, `work_code` , `department_name`, `use_kind` FROM system_extra_user"
	_activityExtraGetUserInfoSQL    = "SELECT /*master*/ `id`, `avatar`, `last_name`, `login_id`, `nick_name`, `work_code`, `department_name`, `use_kind` FROM system_extra_user WHERE `work_code` = ?"
	_activityExtraSetUserInfoSQL    = "INSERT INTO `system_extra_user` (`avatar`, `last_name`, `login_id`, `nick_name`, `work_code`, `department_name`, `use_kind`) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_activityExtraUpdateUserInfoSQL = "UPDATE `system_extra_user` SET `avatar` = ?, `last_name` = ?, `login_id` = ?, `nick_name` = ?, `work_code` = ?, `department_name` = ?, `use_kind` = ? WHERE `work_code` = ?"
	_activityVotedSQL               = "SELECT /*master*/ `id`, `aid`, `uid` FROM system_activity_vote WHERE aid = ? and uid = ?"
	_activityQuestionListSQL        = "SELECT /*master*/ `id`, `aid`, `qid`, `question`, `uid`, `ctime` FROM system_activity_question where aid = ? and state = 0 order by id desc"
)

func (d *Dao) GetDBUserInfoByUID(ctx context.Context, uid string) (res *model.DBUserInfo, err error) {
	res = new(model.DBUserInfo)
	row := d.db.QueryRow(ctx, _userInfoByWorkCodeSQL, uid)
	if err = row.Scan(&res.ID, &res.UID, &res.Token); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "GetDBUserInfoByUID:rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) UpdateUserInfoByUID(ctx context.Context, uid string, sessionToken string) (err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _updateUserInfoSQL, sessionToken, uid); err != nil {
		log.Errorc(ctx, "d.UpdateUserInfoByUID(%v,%v) Exec error(%+v)", sessionToken, uid, err)
		return
	}
	if ef, _ := res.RowsAffected(); ef == 0 {
		log.Errorc(ctx, "d.UpdateUserInfoByUID(%v,%v) RowsAffected Error", sessionToken, uid)
		return
	}
	return
}

func (d *Dao) CreateUserInfo(ctx context.Context, uid string, sessionToken string) (err error) {
	if _, err = d.db.Exec(ctx, _createUserInfoSQL, uid, sessionToken); err != nil {
		log.Errorc(ctx, "d.CreateUserInfo(%v,%v) error(%+v)", uid, sessionToken, err)
		return
	}
	return
}

func (d *Dao) GetDBUserInfoByToken(ctx context.Context, sessionToken string) (res *model.DBUserInfo, err error) {
	res = new(model.DBUserInfo)
	row := d.db.QueryRow(ctx, _userInfoByTokenSQL, sessionToken)
	if err = row.Scan(&res.ID, &res.UID, &res.Token); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "GetDBUserInfoByToken:rows.Err() error(%v)", err)
	}
	return
}

// 获取活动信息
func (d *Dao) GetActivityInfoFromDB(ctx context.Context, aid int64) (res *model.Activity, err error) {
	res = new(model.Activity)
	row := d.db.QueryRow(ctx, _activityInfoByIDSQL, aid)
	if err = row.Scan(&res.ID, &res.Name, &res.Type, &res.Stime, &res.Etime, &res.Config); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "GetActivityInfoFromDB:rows.Err() error(%v)", err)
	}
	return
}

// 查询是否签到过
func (d *Dao) ActivitySigned(ctx context.Context, aid int64, uid string) (res *model.ActivitySign, err error) {
	res = new(model.ActivitySign)
	row := d.db.QueryRow(ctx, _activitySignedSQL, aid, uid)
	if err = row.Scan(&res.ID, &res.AID, &res.UID, &res.Location); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "ActivitySigned:rows.Err() error(%v)", err)
	}
	return
}

// 签到
func (d *Dao) DoActivitySign(ctx context.Context, aid int64, uid string, location string) (err error) {
	if _, err = d.db.Exec(ctx, _createSignSQL, aid, uid, location); err != nil {
		log.Errorc(ctx, "d.DoActivitySign(%v,%v) error(%+v)", aid, uid, err)
		return
	}
	return
}

// 获取座位表信息
func (d *Dao) GetSeatInfo(ctx context.Context, aid int64, uid string) (res *model.SystemActivitySeat, err error) {
	res = new(model.SystemActivitySeat)
	row := d.db.QueryRow(ctx, _activitySeatInfoSQL, aid, uid)
	if err = row.Scan(&res.ID, &res.AID, &res.UID, &res.Content); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "GetSeatInfo:rows.Err() error(%v)", err)
	}
	return
}

// 获取额外员工信息
func (d *Dao) GetExtraUsersInfo(ctx context.Context) (res []*model.User, err error) {
	rows, err := d.db.Query(ctx, _activityExtraUsersInfoSQL)
	if err != nil {
		log.Errorc(ctx, "GetExtraUsersInfo error(%v)", err)
		return
	}
	defer rows.Close()
	res = make([]*model.User, 0)
	for rows.Next() {
		tmp := new(model.User)
		if err = rows.Scan(&tmp.ID, &tmp.Avatar, &tmp.LastName, &tmp.LoginID, &tmp.NickName, &tmp.WorkCode, &tmp.DepartmentName, &tmp.UseKind); err != nil {
			log.Errorc(ctx, "GetExtraUsersInfo scan error(%v)", err)
			return
		}
		res = append(res, tmp)
	}
	err = rows.Err()
	return
}

// 写入额外员工信息
func (d *Dao) SetExtraUsersInfo(ctx context.Context, data *model.User) (err error) {
	// 先查询 再判断是新增还是更新
	tmp := new(model.User)
	row := d.db.QueryRow(ctx, _activityExtraGetUserInfoSQL, data.WorkCode)
	if err = row.Scan(&tmp.ID, &tmp.Avatar, &tmp.LastName, &tmp.LoginID, &tmp.NickName, &tmp.WorkCode, &tmp.DepartmentName, &tmp.UseKind); err != nil {
		if err == sql.ErrNoRows {
			_, err = d.db.Exec(ctx, _activityExtraSetUserInfoSQL, data.Avatar, data.LastName, data.LoginID, data.NickName, data.WorkCode, data.DepartmentName, data.UseKind)
			if err != nil {
				log.Errorc(ctx, "SetExtraUsersInfo d.db.Exec(ctx, _activityExtraSetUsersInfoSQL) Err error(%v)", err)
				return
			}
		} else {
			return
		}
	}
	if tmp.ID > 0 {
		_, err = d.db.Exec(ctx, _activityExtraUpdateUserInfoSQL, data.Avatar, data.LastName, data.LoginID, data.NickName, data.WorkCode, data.DepartmentName, data.UseKind, data.WorkCode)
		if err != nil {
			log.Errorc(ctx, "SetExtraUsersInfo d.db.Exec(ctx, _activityExtraUpdateUserInfoSQL) Err error(%v)", err)
			return
		}
	}
	return
}

// 查询是否投票过
func (d *Dao) ActivityVoted(ctx context.Context, aid int64, uid string) (res *model.ActivityVote, err error) {
	res = new(model.ActivityVote)
	row := d.db.QueryRow(ctx, _activityVotedSQL, aid, uid)
	if err = row.Scan(&res.ID, &res.AID, &res.UID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "ActivityVoted:rows.Err() error(%v)", err)
	}
	return
}

// 签到
func (d *Dao) InsertVote(ctx context.Context, data []*model.ActivityVote) (err error) {
	if len(data) == 0 {
		return
	}
	SQL := "insert into system_activity_vote (aid, uid, item_id, option_id, score) values "
	length := len(data)
	values := ""
	var args []interface{}
	for k, v := range data {
		if k == length-1 {
			values += "(?, ?, ?, ?, ?)"
		} else {
			values += "(?, ?, ?, ?, ?),"
		}
		args = append(args, v.AID, v.UID, v.ItemID, v.OptionID, v.Score)
	}
	SQL += values

	if _, err = d.db.Exec(ctx, SQL, args...); err != nil {
		log.Errorc(ctx, "d.InsertVote(%v,%+v) error(%+v)", SQL, args, err)
		return
	}
	return
}

// 提问
func (d *Dao) InsertQuestion(ctx context.Context, aid int64, uid string, data []model.QuestionEachItem) (err error) {
	if len(data) == 0 {
		return
	}
	SQL := "insert into system_activity_question (aid, qid, question, uid) values "
	length := len(data)
	values := ""
	var args []interface{}
	for k, v := range data {
		if k == length-1 {
			values += "(?, ?, ?, ?)"
		} else {
			values += "(?, ?, ?, ?),"
		}
		args = append(args, aid, v.QID, v.Question, uid)
	}
	SQL += values
	if _, err = d.db.Exec(ctx, SQL, args...); err != nil {
		err = errors.Wrapf(err, "d.InsertQuestion(%v,%+v) error(%+v)", SQL, args, err)
		return
	}
	return
}

// 查询问卷信息
func (d *Dao) GetQuestionList(ctx context.Context, aid int64) (res []*model.ActivitySystemQuestion, err error) {
	rows, err := d.db.Query(ctx, _activityQuestionListSQL, aid)
	if err != nil {
		err = errors.Wrap(err, "d.db.Query err")
		return
	}
	defer rows.Close()
	res = make([]*model.ActivitySystemQuestion, 0)
	for rows.Next() {
		tmp := new(model.ActivitySystemQuestion)
		if err = rows.Scan(&tmp.ID, &tmp.AID, &tmp.QID, &tmp.Question, &tmp.UID, &tmp.Ctime); err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, "rows.Scan err")
			return
		}
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		res = append(res, tmp)
	}
	err = rows.Err()
	return
}
