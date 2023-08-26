package cards

import (
	"context"
	"fmt"
	"strings"

	xsql "database/sql"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	cards "go-gateway/app/web-svr/activity/interface/model/cards"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

const (
	_winCardSQL             = "SELECT mid,gift_id FROM act_lottery_win_%d WHERE mid = ? AND state = 0 ORDER BY id asc"
	_relationInviterSQL     = "SELECT mid FROM act_cards_relation WHERE invitee = ? and activity=? "
	_getTokenToMidSQL       = "SELECT mid FROM act_invite_tokens WHERE token = ? and activity=?"
	_getMidToTokenSQL       = "SELECT token FROM act_invite_tokens WHERE mid = ? and activity=?"
	_getMidArchiveNumsSQL   = "SELECT nums FROM act_spring_archive WHERE mid = ? "
	_NumsSQL                = "SELECT mid,card_1,card_1_used,card_2,card_2_used,card_3,card_3_used,card_4,card_4_used,card_5,card_5_used,card_6,card_6_used,card_7,card_7_used,card_8,card_8_used,card_9,card_9_used,compose FROM act_youth_cards_nums WHERE mid = ?"
	_updateSpringNumsSQL    = "UPDATE act_youth_cards_nums set card_1 = ?,card_1_used = ?,card_2 =?,card_2_used =?,card_3 =?,card_3_used =?,card_4=?,card_4_used=?,card_5=?,card_5_used=?,card_6=?,card_6_used=?,card_7=?,card_7_used=?,card_8=?,card_8_used=?,card_9=?,card_9_used=?,compose=?  WHERE mid = ?"
	_insertNumsSQL          = "INSERT INTO act_youth_cards_nums(mid) VALUES(?)"
	_incrSpringNumsSQL      = "UPDATE act_youth_cards_nums set %s where mid = ? %s"
	_insertComposeLogSQL    = "INSERT INTO act_compose_card_log(mid,activity) VALUES(?,?)"
	_insertSendCardLogSQL   = "INSERT INTO act_send_card_log(mid,activity,card_id,receiver_mid) VALUES(?,?,?,?)"
	_insertMidTokenSQL      = "INSERT INTO act_invite_tokens(mid,token,activity) VALUES(?,?,?)"
	_relationBindSQL        = "INSERT INTO act_cards_relation(mid,invitee,token,activity) VALUES(?,?,?,?)"
	_selectNumsForUpdateSQL = "SELECT mid,card_1,card_1_used,card_2,card_2_used,card_3,card_3_used,card_4,card_4_used,card_5,card_5_used,card_6,card_6_used,card_7,card_7_used,card_8,card_8_used,card_9,card_9_used,compose from act_youth_cards_nums where mid = ?  FOR UPDATE"
	_selectComposeUseSQL    = "SELECT mid,compose_used from act_youth_compose_used where mid=? and state=1"
	_getAllSpringTaskSQL    = "SELECT id,activity_id,task_name,link_name,order_id,activity,counter,task_desc,link,finish_times,state FROM act_task WHERE activity_id = ? and state = 1 order by order_id asc "
	_getAllTaskSQL          = "SELECT id,activity_id,task_name,link_name,order_id,activity,counter,task_desc,link,finish_times,state FROM act_task WHERE state = 1 order by order_id asc "
	_getCardsConfig         = "SELECT id,name,lottery_id,reserve_id,cards_num,cards,sid FROM act_cards WHERE name=?"

	_insertAllCardsNumsSQL = "INSERT INTO act_cards_nums_%d(activity_id,mid,card_id) VALUES %s"
	_NumsNewSQL            = "SELECT mid,activity_id,card_id,nums,used FROM act_cards_nums_%d WHERE mid = ?"

	_incrCardsNumNewSQL        = "UPDATE act_cards_nums_%d set %s where mid = ? "
	_selectNumsForUpdateSQLNew = "SELECT mid,activity_id,card_id,nums,used FROM act_cards_nums_%d where mid = ?  FOR UPDATE"
)

func (d *Dao) InitAddMidCards(ctx context.Context, activity_id, mid int64, cardsNums int64) error {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for i := 0; i <= int(cardsNums); i++ {
		rowsTmp = append(rowsTmp, "(?,?,?)")
		rows = append(rows, activity_id, mid, i)
	}
	_, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_insertAllCardsNumsSQL, activity_id, strings.Join(rowsTmp, ",")), rows...)
	if err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
		return errors.Wrap(err, "AddUserAward Exec")
	}
	return nil
}

// UpdateCardNumsIncrNew ....
func (d *Dao) UpdateCardNumsIncrNew(c context.Context, mid int64, activityID int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var sqlStr string
	sqlStr = " nums =( case  "
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("when card_id= %s then nums+%d ", card, nums))
	}
	endSql := " else nums end) "
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_incrCardsNumNewSQL, activityID, sqlStr+strings.Join(params, " ")+endSql), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsIncrNew:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// UpdateCardNumsIncrNew ....
func (d *Dao) UpdateCardNumsIncrNewTx(c context.Context, tx *sql.Tx, mid int64, activityID int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var sqlStr string
	sqlStr = " nums =( case  "
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("when card_id= %s then nums+%d ", card, nums))
	}
	endSql := " else nums end) "
	if res, err = tx.Exec(fmt.Sprintf(_incrCardsNumNewSQL, activityID, sqlStr+strings.Join(params, " ")+endSql), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsIncrNew:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// UpdateCardNumsIncrNew ....
func (d *Dao) UpdateCardNumsDescNewTx(c context.Context, tx *sql.Tx, mid int64, activityID int64, cards map[string]int64) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var sqlStr string
	sqlStr = " used =( case  "
	for card, nums := range cards {
		params = append(params, fmt.Sprintf("when card_id= %s then used+%d ", card, nums))
	}
	endSql := " else used end) "
	if res, err = tx.Exec(fmt.Sprintf(_incrCardsNumNewSQL, activityID, sqlStr+strings.Join(params, " ")+endSql), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsDescNewTx:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// RawCardsConfig
func (d *Dao) RawCardsConfig(c context.Context, name string) (res *cards.Cards, err error) {
	res = new(cards.Cards)
	row := component.GlobalDB.QueryRow(c, _getCardsConfig, name)
	if err = row.Scan(&res.ID, &res.Name, &res.LotteryID, &res.ReserveID, &res.CardsNum, &res.Cards, &res.SID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CardsConfig:QueryRow")
		}
	}
	return
}

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
func (d *Dao) MidNums(c context.Context, mid int64) (res *cards.MidNums, err error) {
	res = new(cards.MidNums)
	row := component.GlobalDB.QueryRow(c, _NumsSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Card6, &res.Card6Used, &res.Card7, &res.Card7Used, &res.Card8, &res.Card8Used, &res.Card9, &res.Card9Used, &res.Compose); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "MidNums:QueryRow")
		}
	}
	return
}

// MidNums 数量获取
func (d *Dao) MidNumsNew(c context.Context, mid int64, activityID int64) (res []*cards.CardMid, err error) {

	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, fmt.Sprintf(_NumsNewSQL, activityID), mid); err != nil {
		err = errors.Wrap(err, "RawLotteryWinList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*cards.CardMid, 0)
	for rows.Next() {
		l := &cards.CardMid{}
		if err = rows.Scan(&l.MID, &l.ActivityID, &l.CardID, &l.Nums, &l.Used); err != nil {
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

// MidNumsEmptyErr 数量获取
func (d *Dao) MidNumsEmptyErr(c context.Context, mid int64) (res *cards.MidNums, isEmpty bool, err error) {
	res = new(cards.MidNums)
	row := component.GlobalDB.QueryRow(c, _NumsSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Card6, &res.Card6Used, &res.Card7, &res.Card7Used, &res.Card8, &res.Card8Used, &res.Card9, &res.Card9Used, &res.Compose); err != nil {
		if err == sql.ErrNoRows {
			isEmpty = true
			err = nil
		} else {
			err = errors.Wrap(err, "MidNums:QueryRow")
		}
	}
	return
}

// MidComposeUsed 合成套数
func (d *Dao) MidComposeUsed(c context.Context, mid int64) (res *cards.MidComposeUsed, err error) {
	res = new(cards.MidComposeUsed)
	row := component.GlobalDB.QueryRow(c, _selectComposeUseSQL, mid)
	if err = row.Scan(&res.MID, &res.ComposeUsed); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "MidComposeUsed error(%v)", err)
		}
	}
	return
}

// MidNumsForUpdateTx 合成套数
func (d *Dao) MidNumsForUpdateTx(c context.Context, tx *sql.Tx, mid int64) (res *cards.MidNums, err error) {
	res = new(cards.MidNums)
	row := tx.QueryRow(_selectNumsForUpdateSQL, mid)
	if err = row.Scan(&res.MID, &res.Card1, &res.Card1Used, &res.Card2, &res.Card2Used, &res.Card3, &res.Card3Used, &res.Card4, &res.Card4Used, &res.Card5, &res.Card5Used, &res.Card6, &res.Card6Used, &res.Card7, &res.Card7Used, &res.Card8, &res.Card8Used, &res.Card9, &res.Card9Used, &res.Compose); err != nil {
		log.Errorc(c, "MidNumsForUpdateTx error(%v)", err)
		err = errors.Wrap(err, "MidNumsForUpdateTx:QueryRow")
	}
	return
}

// MidNumsForUpdateTx 合成套数
func (d *Dao) MidNumsForUpdateTxNew(c context.Context, tx *sql.Tx, mid int64, activityID int64) (res []*cards.CardMid, err error) {
	var rows *sql.Rows
	if rows, err = tx.Query(fmt.Sprintf(_selectNumsForUpdateSQLNew, activityID), mid); err != nil {
		err = errors.Wrap(err, "MidNumsForUpdateTxNew:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*cards.CardMid, 0)
	for rows.Next() {
		l := &cards.CardMid{}
		if err = rows.Scan(&l.MID, &l.ActivityID, &l.CardID, &l.Nums, &l.Used); err != nil {
			err = errors.Wrap(err, "MidNumsForUpdateTxNew:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "MidNumsForUpdateTxNew:rows.Err()")
	}
	return
}

// UpdateCardNums ....
func (d *Dao) UpdateCardNums(c context.Context, tx *sql.Tx, mid, card1, card1Used, card2, card2Used, card3, card3Used, card4, card4Used, card5, card5Used, card6, card6Used, card7, card7Used, card8, card8Used, card9, card9Used, compose int64) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_updateSpringNumsSQL), card1, card1Used, card2, card2Used, card3, card3Used, card4, card4Used, card5, card5Used, card6, card6Used, card7, card7Used, card8, card8Used, card9, card9Used, compose, mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNums:dao.db.Exec")
		return
	}
	return res.RowsAffected()
}

// UpdateCardNums ....
func (d *Dao) UpdateCardNumsNew(c context.Context, tx *sql.Tx, mid, activityID int64, cards []*cards.CardMid) (ef int64, err error) {
	var res xsql.Result
	var params = make([]string, 0)
	var paramsNew = make([]string, 0)
	var sqlStr string
	sqlStr = " used =( case  "
	for _, v := range cards {
		paramsNew = append(paramsNew, fmt.Sprintf("when card_id= %d then %d ", v.CardID, v.Nums))
		params = append(params, fmt.Sprintf("when card_id= %d then %d ", v.CardID, v.Used))
	}
	endSql := " else used end) "
	sqlNewStr := " , nums =( case  "
	endNewSql := " else nums end) "

	if res, err = tx.Exec(fmt.Sprintf(_incrCardsNumNewSQL, activityID, sqlStr+strings.Join(params, " ")+endSql+sqlNewStr+strings.Join(paramsNew, " ")+endNewSql), mid); err != nil {
		err = errors.Wrap(err, "UpdateCardNumsIncrNew:dao.db.Exec")
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
func (d *Dao) InsertSpringMidInviteToken(c context.Context, mid int64, token, activity string) (ef int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_insertMidTokenSQL), mid, token, activity); err != nil {
		err = errors.Wrap(err, "InsertSpringMidInviteToken:dao.db.Exec")
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = nil
			return
		}
	}
	return res.LastInsertId()
}

// InsertComposeLogTx ...
func (d *Dao) InsertComposeLogTx(c context.Context, tx *sql.Tx, mid int64, activity string) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_insertComposeLogSQL), mid, activity); err != nil {
		err = errors.Wrap(err, "InsertComposeLogTx:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// InsertSendCardLogTx ...
func (d *Dao) InsertSendCardLogTx(c context.Context, tx *sql.Tx, mid, receiverMid, cardID int64, activity string) (ef int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_insertSendCardLogSQL), mid, activity, cardID, receiverMid); err != nil {
		err = errors.Wrap(err, "InsertSendCardLogTx:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// MidInviterDB 获取用户的邀请者
func (d *Dao) MidInviterDB(c context.Context, mid int64, activity string) (res int64, err error) {
	row := component.GlobalDB.QueryRow(c, _relationInviterSQL, mid, activity)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "MidInviter error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InviteTokenToMidDB 根据token获取用户信息
func (d *Dao) InviteTokenToMidDB(c context.Context, token string, activity string) (res int64, err error) {
	row := component.GlobalDB.QueryRow(c, _getTokenToMidSQL, token, activity)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "InviteTokenToMidDB error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InviteMidToTokenDB 根据token获取用户信息
func (d *Dao) InviteMidToTokenDB(c context.Context, mid int64, activity string) (res string, err error) {
	row := component.GlobalDB.QueryRow(c, _getMidToTokenSQL, mid, activity)
	if err = row.Scan(&res); err != nil {
		log.Errorc(c, "InviteMidToTokenDB error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// InsertRelationBind ...
func (d *Dao) InsertRelationBind(c context.Context, inviter, invitee int64, token, activity string) (ef int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_relationBindSQL), inviter, invitee, token, activity); err != nil {
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
func (d *Dao) RawTaskList(c context.Context, activityID int64) (res []*cards.Task, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, _getAllSpringTaskSQL, activityID); err != nil {
		err = errors.Wrap(err, "RawTaskList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*cards.Task, 0)
	for rows.Next() {
		l := &cards.Task{}
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

func (d *Dao) AllTaskList(c context.Context) (res []*cards.Task, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, _getAllTaskSQL); err != nil {
		err = errors.Wrap(err, "RawTaskList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*cards.Task, 0)
	for rows.Next() {
		l := &cards.Task{}
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
