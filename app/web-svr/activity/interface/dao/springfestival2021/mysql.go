package springfestival2021

import (
	"context"
	"fmt"
	"strings"

	xsql "database/sql"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	springfestival2021 "go-gateway/app/web-svr/activity/interface/model/springfestival2021"

	"github.com/pkg/errors"
)

const (
	_winCardSQL             = "SELECT mid,gift_id FROM act_lottery_win_%d WHERE mid = ? AND state = 0 ORDER BY id asc"
	_relationInviterSQL     = "SELECT mid FROM act_spring_relation WHERE invitee = ? "
	_getTokenToMidSQL       = "SELECT mid FROM act_spring_tokens WHERE token = ? "
	_getMidToTokenSQL       = "SELECT token FROM act_spring_tokens WHERE mid = ? "
	_getMidArchiveNumsSQL   = "SELECT nums FROM act_spring_archive WHERE mid = ? "
	_NumsSQL                = "SELECT mid,card_1,card_1_used,card_2,card_2_used,card_3,card_3_used,card_4,card_4_used,card_5,card_5_used,compose FROM act_spring_cards_nums WHERE mid = ?"
	_updateSpringNumsSQL    = "UPDATE act_spring_cards_nums set card_1 = ?,card_1_used = ?,card_2 =?,card_2_used =?,card_3 =?,card_3_used =?,card_4=?,card_4_used=?,card_5=?,card_5_used=?,compose=?  WHERE mid = ?"
	_insertNumsSQL          = "INSERT INTO act_spring_cards_nums(mid) VALUES(?)"
	_incrSpringNumsSQL      = "UPDATE act_spring_cards_nums set %s where mid = ? %s"
	_insertComposeLogSQL    = "INSERT INTO act_spring_compose_card_log(mid) VALUES(?)"
	_insertSendCardLogSQL   = "INSERT INTO act_spring_send_card_log(mid,card_id,receiver_mid) VALUES(?,?,?)"
	_insertMidTokenSQL      = "INSERT INTO act_spring_tokens(mid,token) VALUES(?,?)"
	_relationBindSQL        = "INSERT INTO act_spring_relation(mid,invitee,token) VALUES(?,?,?)"
	_selectNumsForUpdateSQL = "SELECT mid,card_1,card_1_used,card_2,card_2_used,card_3,card_3_used,card_4,card_4_used,card_5,card_5_used,compose from act_spring_cards_nums where mid = ? FOR UPDATE"
	_getAllSpringTaskSQL    = "SELECT id,activity_id,task_name,link_name,order_id,activity,counter,task_desc,link,finish_times,state FROM act_task WHERE activity_id = ? and state = 1 order by order_id asc "
)

// RawLotteryWinList ...
func (d *Dao) RawLotteryWinList(c context.Context, lotteryID int64, mid int64) (res []*lottery.GiftMid, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, fmt.Sprintf(_winCardSQL, lotteryID), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryWinList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*lottery.GiftMid, 0)
	for rows.Next() {
		l := &lottery.GiftMid{}
		if err = rows.Scan(&l.Mid, &l.GiftID); err != nil {
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

// MidNums 数量获取
func (d *Dao) MidNums(c context.Context, mid int64) (res *springfestival2021.MidNums, err error) {
	res = new(springfestival2021.MidNums)
	row := component.GlobalDB.QueryRow(c, _NumsSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Compose); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "MidNums:QueryRow")
		}
	}
	return
}

// MidNumsEmptyErr 数量获取
func (d *Dao) MidNumsEmptyErr(c context.Context, mid int64) (res *springfestival2021.MidNums, isEmpty bool, err error) {
	res = new(springfestival2021.MidNums)
	row := component.GlobalDB.QueryRow(c, _NumsSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Compose); err != nil {
		if err == sql.ErrNoRows {
			isEmpty = true
			err = nil
		} else {
			err = errors.Wrap(err, "MidNums:QueryRow")
		}
	}
	return
}

// MidNumsForUpdateTx 合成套数
func (d *Dao) MidNumsForUpdateTx(c context.Context, tx *sql.Tx, mid int64) (res *springfestival2021.MidNums, err error) {
	res = new(springfestival2021.MidNums)
	row := tx.QueryRow(_selectNumsForUpdateSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Compose); err != nil {
		log.Errorc(c, "MidNumsForUpdateTx error(%v)", err)
		err = errors.Wrap(err, "MidNumsForUpdateTx:QueryRow")
	}
	return
}

// UpdateCardNums ....
func (d *Dao) UpdateCardNums(c context.Context, tx *sql.Tx, mid, card1, card1Used, card2, card2Used, card3, card3Used, card4, card4Used, card5, card5Used, compose int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_updateSpringNumsSQL), card1, card1Used, card2, card2Used, card3, card3Used, card4, card4Used, card5, card5Used, compose, mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNums:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// InsertSpringNums ...
func (d *Dao) InsertSpringNums(c context.Context, mid int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_insertNumsSQL), mid); err != nil {
		err = errors.Wrap(err, "InsertSpringNums:dao.db.Exec")
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = nil
			return
		}
	}
	return res.LastInsertId()
}

// InsertSpringMidInviteToken ...
func (d *Dao) InsertSpringMidInviteToken(c context.Context, mid int64, token string) (ef int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_insertMidTokenSQL), mid, token); err != nil {
		err = errors.Wrap(err, "InsertSpringMidInviteToken:dao.db.Exec")
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = nil
			return
		}
	}
	return res.LastInsertId()
}

// InsertComposeLogTx ...
func (d *Dao) InsertComposeLogTx(c context.Context, tx *sql.Tx, mid int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_insertComposeLogSQL), mid); err != nil {
		err = errors.Wrap(err, "InsertComposeLogTx:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertSendCardLogTx ...
func (d *Dao) InsertSendCardLogTx(c context.Context, tx *sql.Tx, mid, receiverMid, cardID int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_insertSendCardLogSQL), mid, cardID, receiverMid); err != nil {
		err = errors.Wrap(err, "InsertSendCardLogTx:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// MidInviterDB 获取用户的邀请者
func (d *Dao) MidInviterDB(c context.Context, mid int64) (res int64, err error) {
	row := component.GlobalDB.QueryRow(c, _relationInviterSQL, mid)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "MidInviter error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InviteTokenToMidDB 根据token获取用户信息
func (d *Dao) InviteTokenToMidDB(c context.Context, token string) (res int64, err error) {
	row := component.GlobalDB.QueryRow(c, _getTokenToMidSQL, token)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "InviteTokenToMidDB error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InviteMidToTokenDB 根据token获取用户信息
func (d *Dao) InviteMidToTokenDB(c context.Context, mid int64) (res string, err error) {
	row := component.GlobalDB.QueryRow(c, _getMidToTokenSQL, mid)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "InviteMidToTokenDB error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InsertRelationBind ...
func (d *Dao) InsertRelationBind(c context.Context, inviter, invitee int64, token string) (ef int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_relationBindSQL), inviter, invitee, token); err != nil {
		err = errors.Wrap(err, "InsertRelationBind:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// UpdateCardNumsIncr ....
func (d *Dao) UpdateCardNumsIncr(c context.Context, mid int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var paramsOther = make([]string, 0)
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("%s=%s+%d", card, card, nums))
		paramsOther = append(paramsOther, fmt.Sprintf(" and %s-%s_used<999", card, card))
	}
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_incrSpringNumsSQL, strings.Join(params, ","), strings.Join(paramsOther, " ")), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsIncr:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// UpdateCardNumsIncrTx ....
func (d *Dao) UpdateCardNumsIncrTx(c context.Context, tx *sql.Tx, mid int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var paramsOther = make([]string, 0)
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("%s=%s+%d", card, card, nums))
		paramsOther = append(paramsOther, fmt.Sprintf(" and %s-%s_used<999", card, card))
	}
	if res, err = tx.Exec(fmt.Sprintf(_incrSpringNumsSQL, strings.Join(params, ","), strings.Join(paramsOther, " ")), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsIncrTx:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// UpdateCardNumsUsedIncrTx ....
func (d *Dao) UpdateCardNumsUsedIncrTx(c context.Context, tx *sql.Tx, mid int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var paramsOther = make([]string, 0)
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("%s_used=%s_used+%d", card, card, nums))
		paramsOther = append(paramsOther, fmt.Sprintf(" and %s-%s_used>1", card, card))
	}
	if res, err = tx.Exec(fmt.Sprintf(_incrSpringNumsSQL, strings.Join(params, ","), strings.Join(paramsOther, " ")), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsUsedIncrTx:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// RawTaskList ...
func (d *Dao) RawTaskList(c context.Context, activityID int64) (res []*springfestival2021.Task, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, _getAllSpringTaskSQL, activityID); err != nil {
		err = errors.Wrap(err, "RawTaskList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*springfestival2021.Task, 0)
	for rows.Next() {
		l := &springfestival2021.Task{}
		if err = rows.Scan(&l.ID, &l.ActivityID, &l.TaskName, &l.LinkName, &l.OrderID, &l.Activity, &l.Counter, &l.Desc, &l.Link, &l.FinishTimes, &l.State); err != nil {
			err = errors.Wrap(err, "RawTaskList:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawTaskList:rows.Err()")
	}
	return
}

// ArchiveNumsDB 获取用户投稿数
func (d *Dao) ArchiveNumsDB(c context.Context, mid int64) (res int64, err error) {
	row := component.GlobalDB.QueryRow(c, _getMidArchiveNumsSQL, mid)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "InviteTokenToMidDB error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}
