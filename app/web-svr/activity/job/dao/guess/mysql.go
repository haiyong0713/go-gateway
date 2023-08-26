package guess

import (
	"context"
	xsql "database/sql"
	"fmt"

	"go-gateway/app/web-svr/activity/job/dao"

	"go-common/library/database/sql"
	"go-common/library/xstr"

	guemdl "go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/job/model/guess"

	"github.com/pkg/errors"
)

const (
	_userSub              = 100
	_int                  = 100
	_upUserSQL            = "UPDATE  act_guess_user_%s SET status = 1 ,income = CASE %s END WHERE status = 0 AND id IN (%s)"
	_upUserStatusSQL      = "UPDATE  act_guess_user_%s SET status = 1 WHERE status = 0 AND id IN (%s)"
	_detailGuessSQL       = "SELECT id,option,odds,total_stake,ctime,mtime FROM act_guess_detail  WHERE  main_id = ?"
	_userFinishSQL        = "SELECT id,mid,main_id,detail_id,stake_type,stake,income,status,ctime,mtime  FROM  act_guess_user_%s  WHERE  status = 0 AND main_id = ?"
	_userFinishSQLByLimit = `
SELECT a.id, a.mid, a.main_id, a.detail_id, a.stake_type
	, a.stake, a.income, a.status, a.ctime, a.mtime
FROM act_guess_user_%s a
	INNER JOIN (
		SELECT id
		FROM act_guess_user_%s
		WHERE (status = 0
			AND main_id = ?
			AND id > ?)
		ORDER BY id ASC
		LIMIT 1000
	) b
	ON a.id = b.id
`
	_upDetailOddsSQL = "UPDATE act_guess_detail SET odds = CASE %s END WHERE id IN (%s)"
	_upUserLogSQL    = "UPDATE act_guess_user_log SET total_success = CASE %s END ,success_rate = (CASE WHEN total_guess = 0 THEN 0 ELSE (convert(total_success/total_guess,decimal(10,2)))*100 end),total_income = CASE %s END WHERE mid IN (%s) AND business = ?"
	_userLogRankSQL  = "SELECT id,business,mid,stake_type,total_guess,total_success,success_rate FROM act_guess_user_log FORCE INDEX (ix_success_rate) WHERE business = ? ORDER BY success_rate DESC LIMIT ?,?"
	_userCountSQL    = "SELECT count(1) as c FROM act_guess_user_log WHERE business = ?"
	_upUserRankSQL   = "UPDATE act_guess_user_log SET ranking = CASE %s END WHERE id IN (%s)"
	_mainsGuessSQL   = "SELECT id,business,oid,title,stake_type,max_stake,result_id,guess_count,is_deleted,ctime,mtime,stime,etime FROM act_guess_main WHERE is_deleted = 0 AND oid IN (%s)"

	sql4SetGuessAsInProcess = `
UPDATE act_guess_main SET settlement_status = 1
WHERE id = ?
        AND oid = ?
`
	sql4SetGuessAsCleared = `
UPDATE act_guess_main SET settlement_status = 2
WHERE id = ?
        AND oid = ?
`
	sql4UnClearedGuess = `
SELECT id, mid, main_id, detail_id, stake_type
    , stake, income, status, ctime, mtime
FROM act_guess_user_%s
WHERE (status = 0
	AND mid IN (%s)
	AND main_id = ?)
`
	sql4UnClearedUserGuessList = `
SELECT id, mid, main_id, detail_id, stake_type
	, stake, income, status, ctime, mtime
FROM act_guess_user_%v
WHERE status = 0
	AND mid = ?
`
	sql4ClearGuessMain = `
SELECT id, business, oid, title, stake_type
	, max_stake, result_id, guess_count, is_deleted, ctime
	, mtime, stime, etime
FROM act_guess_main
WHERE is_deleted = 0
	AND settlement_status = 2
`
	sql4AllMidFormUserGuess = `
SELECT DISTINCT mid
FROM act_guess_user_%v
`
	sql4UnClearedMIDList = `
SELECT mid
FROM act_guess_user_%v
WHERE main_id = ?
	AND status = 0
`

	sql4AllGuessList = `
SELECT id, mid, main_id, detail_id, stake_type
    , stake, income, status
FROM act_guess_user_%s
WHERE mid = ?
`
	sql4UpdateUserLog = `
UPDATE act_guess_user_log
SET success_rate = ?, total_guess = ?, total_income = ?, total_success = ?
WHERE mid = ?
	AND business = ?
`
	sql4UpdateUserGuess = `
UPDATE act_guess_user_%v
SET status = 1, income = ?
WHERE status = 0
	AND id = ?
`
	sql4GuessMainList = `
SELECT id, business, oid, title, stake_type
	, max_stake, result_id, guess_count, is_deleted, ctime
	, mtime, stime, etime
FROM act_guess_main
WHERE id IN (%v)
`
	sql4UserGuessList = `
SELECT id, mid, main_id, detail_id, stake_type
	, stake, income, status, ctime, mtime
FROM act_guess_user_%v
WHERE mid = ? and main_id = ?
`
	sql4UpdateUserLogCoins = `
UPDATE act_guess_user_log
SET total_income = total_income + ?
WHERE mid = ?
	AND business = ?
`
	sql4UpdateUserGuessLogCoins = `
UPDATE act_guess_user_%v
SET income = ?, status = 1
WHERE id = ?
`
	_mdsOptionSQL = `
SELECT m.id,d.id AS detail_id, d.option,m.oid FROM act_guess_main m
	INNER JOIN act_guess_detail d ON m.id = d.main_id
WHERE m.is_deleted=0 AND m.oid IN (%s)
ORDER BY m.id, d.id`
)

func (d *Dao) AllUnClearedGuessListByMid(ctx context.Context, mid int64) (res []*guess.GuessUser, err error) {
	res = make([]*guess.GuessUser, 0)

	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(sql4UnClearedUserGuessList, userHit(mid)), mid)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		r := new(guess.GuessUser)
		err = rows.Scan(
			&r.ID,
			&r.Mid,
			&r.MainID,
			&r.DetailID,
			&r.StakeType,
			&r.Stake,
			&r.Income,
			&r.Status,
			&r.Ctime,
			&r.Mtime)
		if err != nil {
			return
		}

		res = append(res, r)
	}

	err = rows.Err()

	return
}

func (d *Dao) UpdateUserLogCoins(ctx context.Context, mid int64, coins int64) error {
	_, err := d.db.Exec(ctx, sql4UpdateUserLogCoins, coins, mid, 1)

	return err
}

func (d *Dao) UpdateUserGuessLogCoins(ctx context.Context, id, mid int64, coins float64) error {
	_, err := d.db.Exec(ctx, fmt.Sprintf(sql4UpdateUserGuessLogCoins, userHit(mid)), coins*_int, id)

	return err
}

// UserLogCount business user count.
func (d *Dao) UserGuessInfo(ctx context.Context, mid, mainID int64) (userGuess *guess.GuessUser, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4UserGuessList, userHit(mid)), mid, mainID)
	userGuess = new(guess.GuessUser)
	err = row.Scan(
		&userGuess.ID,
		&userGuess.Mid,
		&userGuess.MainID,
		&userGuess.DetailID,
		&userGuess.StakeType,
		&userGuess.Stake,
		&userGuess.Income,
		&userGuess.Status,
		&userGuess.Ctime,
		&userGuess.Mtime)

	return
}

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%_userSub)
}

func (d *Dao) SetGuessAsInProcess(ctx context.Context, mainID, contestID int64) error {
	_, err := d.db.Exec(ctx, sql4SetGuessAsInProcess, mainID, contestID)

	return err
}

func (d *Dao) SetGuessAsCleared(ctx context.Context, mainID, contestID int64) error {
	_, err := d.db.Exec(ctx, sql4SetGuessAsCleared, mainID, contestID)

	return err
}

// UpUser set user finish .
func (d *Dao) UpUser(c context.Context, mid int64, userIncome map[int64]float64) (rs int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	if len(userIncome) == 0 {
		return
	}
	for id, income := range userIncome {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %f", caseStr, id, income*_int)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upUserSQL, userHit(mid), caseStr, xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpUser dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpUser res.RowsAffected")
	}
	return
}

// UpUserStatus set user status .
func (d *Dao) UpUserStatus(c context.Context, mid int64, ids []int64) (rs int64, err error) {
	var res xsql.Result
	if len(ids) == 0 {
		return
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upUserStatusSQL, userHit(mid), xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpUserStatus dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpUserStatus res.RowsAffected")
	}
	return
}

// GuessDetail get detail by ids.
func (d *Dao) GuessDetail(c context.Context, mainID int64) (res []*guemdl.DetailGuess, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = d.db.Query(c, _detailGuessSQL, mainID)
	if err != nil {
		err = errors.Wrap(err, "GuessDetail:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guemdl.DetailGuess)
		if err = rows.Scan(&r.ID, &r.Option, &r.Odds, &r.TotalStake, &r.Ctime, &r.Mtime); err != nil {
			err = errors.Wrap(err, "GuessDetail:rows.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GuessDetail:rows.Err")
	}
	return
}

// GuessFinish get guess finish user.
func (d *Dao) GuessFinish(c context.Context, mid, mainID int64) (res []*guess.GuessUser, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = d.db.Query(c, fmt.Sprintf(_userFinishSQL, userHit(mid)), mainID)
	if err != nil {
		err = errors.Wrap(err, "GuessFinish:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.GuessUser)
		if err = rows.Scan(&r.ID, &r.Mid, &r.MainID, &r.DetailID, &r.StakeType, &r.Stake, &r.Income, &r.Status, &r.Ctime, &r.Mtime); err != nil {
			err = errors.Wrap(err, "GuessFinish:rows.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GuessFinish:rows.Err")
	}
	return
}

func (d *Dao) SingleTableGuessRecord(ctx context.Context, midList []int64, mainID int64) (res []*guess.GuessUser, err error) {
	res = make([]*guess.GuessUser, 0)
	if len(midList) == 0 {
		return
	}

	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(sql4UnClearedGuess, userHit(midList[0]), xstr.JoinInts(midList)), mainID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		r := new(guess.GuessUser)
		err = rows.Scan(
			&r.ID,
			&r.Mid,
			&r.MainID,
			&r.DetailID,
			&r.StakeType,
			&r.Stake,
			&r.Income,
			&r.Status,
			&r.Ctime,
			&r.Mtime)
		if err != nil {
			return
		}

		res = append(res, r)
	}

	err = rows.Err()

	return
}

// GuessFinish get guess finish user.
func (d *Dao) GuessFinishByLimit(c context.Context, mid, mainID, startID int64) (res []*guess.GuessUser, err error) {
	var (
		rows *sql.Rows
	)

	for i := 0; i < 3; i++ {
		if rows, err = d.db.Query(c, fmt.Sprintf(_userFinishSQLByLimit, userHit(mid), userHit(mid)), mainID, startID); err == nil {
			break
		}
	}
	if err != nil {
		err = errors.Wrap(err, "GuessFinish:d.db.Query")
		return
	}

	defer rows.Close()

	for rows.Next() {
		r := new(guess.GuessUser)
		if err = rows.Scan(&r.ID, &r.Mid, &r.MainID, &r.DetailID, &r.StakeType, &r.Stake, &r.Income, &r.Status, &r.Ctime, &r.Mtime); err != nil {
			err = errors.Wrap(err, "GuessFinish:rows.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GuessFinish:rows.Err")
	}

	return
}

// UpDetailOdds update detail guess odds.
func (d *Dao) UpDetailOdds(c context.Context, detailOdds map[int64]float64) (affected int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	for id, odds := range detailOdds {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %f", caseStr, id, odds*_int)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upDetailOddsSQL, caseStr, xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpDetailOdds:db.Exec error")
		return
	}
	return res.RowsAffected()
}

// UpUserLog update user log.
func (d *Dao) UpUserLog(c context.Context, userIncome map[int64]float64, business int64) (affected int64, err error) {
	var (
		incomeStr, successStr string
		mids                  []int64
		res                   xsql.Result
		right                 int64
	)
	if len(userIncome) == 0 {
		return
	}
	for mid, income := range userIncome {
		right = 0
		incomeStr = fmt.Sprintf("%s WHEN mid = %d THEN total_income + %f", incomeStr, mid, income*_int)
		if income > 0 {
			right = 1
		}
		successStr = fmt.Sprintf("%s WHEN mid = %d THEN total_success + %d", successStr, mid, right)
		mids = append(mids, mid)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upUserLogSQL, successStr, incomeStr, xstr.JoinInts(mids)), business); err != nil {
		err = errors.Wrap(err, "UpUserLog:db.Exec error")
		return
	}
	return res.RowsAffected()
}

// UserRank user rank by success_rate.
func (d *Dao) UserRank(ctx context.Context, business, offset, limit int64) (rs []*guess.UserLog, err error) {
	//use read-only slave for rows reading
	rows, err := dao.GlobalReadDB.Query(ctx, _userLogRankSQL, business, offset, limit)
	if err != nil {
		err = errors.Wrap(err, "UserRank:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &guess.UserLog{}
		err = rows.Scan(&r.ID, &r.Business, &r.Mid, &r.StakeType, &r.TotalGuess, &r.TotalSuccess, &r.SuccessRate)
		if err != nil {
			err = errors.Wrap(err, "UserRank:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "UserRank:rows.Err")
	}
	return
}

// UpUserRank update user rank.
func (d *Dao) UpUserRank(c context.Context, userRank map[int64]int64) (affected int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	for id, rank := range userRank {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %d", caseStr, id, rank)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upUserRankSQL, caseStr, xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpUserRank:db.Exec error")
		return
	}
	return res.RowsAffected()
}

// UserLogCount business user count.
func (d *Dao) UserLogCount(c context.Context, business int64) (total int, err error) {
	row := d.db.QueryRow(c, _userCountSQL, business)
	if err = row.Scan(&total); err != nil {
		err = errors.WithStack(err)
	}
	return
}

// MainList guess main list.
func (d *Dao) MainList(ctx context.Context, oids []int64) (rs []*guemdl.MainGuess, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_mainsGuessSQL, xstr.JoinInts(oids)))
	if err != nil {
		err = errors.Wrap(err, "MainList:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		mod := &guemdl.MainGuess{}
		err = rows.Scan(&mod.ID, &mod.Business, &mod.Oid, &mod.Title, &mod.StakeType, &mod.MaxStake, &mod.ResultID, &mod.GuessCount, &mod.IsDeleted, &mod.Ctime, &mod.Mtime, &mod.Stime, &mod.Etime)
		if err != nil {
			err = errors.Wrap(err, "UserRank:rows.Scan error")
			return
		}
		rs = append(rs, mod)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "MainList:rows.Err")
	}
	return
}

func (d *Dao) AllClearedGuessList(ctx context.Context) (rs []*guemdl.MainGuess, err error) {
	rows, err := d.db.Query(ctx, sql4ClearGuessMain)
	if err != nil {
		err = errors.Wrap(err, "MainList:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		mod := &guemdl.MainGuess{}
		err = rows.Scan(&mod.ID, &mod.Business, &mod.Oid, &mod.Title, &mod.StakeType, &mod.MaxStake, &mod.ResultID, &mod.GuessCount, &mod.IsDeleted, &mod.Ctime, &mod.Mtime, &mod.Stime, &mod.Etime)
		if err != nil {
			err = errors.Wrap(err, "UserRank:rows.Scan error")
			return
		}
		rs = append(rs, mod)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "MainList:rows.Err")
	}
	return
}

func (d *Dao) GuessMainList(ctx context.Context, ids []int64) (rs []*guemdl.MainGuess, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(sql4GuessMainList, xstr.JoinInts(ids)))
	if err != nil {
		err = errors.Wrap(err, "MainList:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		mod := &guemdl.MainGuess{}
		err = rows.Scan(&mod.ID, &mod.Business, &mod.Oid, &mod.Title, &mod.StakeType, &mod.MaxStake, &mod.ResultID, &mod.GuessCount, &mod.IsDeleted, &mod.Ctime, &mod.Mtime, &mod.Stime, &mod.Etime)
		if err != nil {
			return
		}
		rs = append(rs, mod)
	}
	err = rows.Err()

	return
}

func (d *Dao) UnClearedMIDListByTableSuffixAndMainID(ctx context.Context, suffix string, mainID int64) (list []int64, err error) {
	var rows *sql.Rows
	list = make([]int64, 0)

	rows, err = d.db.Query(ctx, fmt.Sprintf(sql4UnClearedMIDList, suffix), mainID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			err = errors.Wrap(err, "UnClearedMIDListByTableSuffixAndMainID:rows.Scan() error")
			return
		}

		list = append(list, mid)
	}

	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "UnClearedMIDListByTableSuffixAndMainID:rows.Err")
	}

	return
}

func (d *Dao) AllGuessedMidListByTableSuffix(ctx context.Context, suffix string) (list []int64, err error) {
	var rows *sql.Rows
	list = make([]int64, 0)

	rows, err = d.db.Query(ctx, fmt.Sprintf(sql4AllMidFormUserGuess, suffix))
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			err = errors.Wrap(err, "AllGuessedMidListByTableSuffix:rows.Scan() error")
			return
		}

		list = append(list, mid)
	}

	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "AllGuessedMidListByTableSuffix:rows.Err")
	}

	return
}

func (d *Dao) AllGuessListByMID(ctx context.Context, mid int64) (list []*guess.GuessUser, err error) {
	var rows *sql.Rows
	list = make([]*guess.GuessUser, 0)

	rows, err = d.db.Query(ctx, fmt.Sprintf(sql4AllGuessList, userHit(mid)), mid)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		r := new(guess.GuessUser)
		if err = rows.Scan(&r.ID, &r.Mid, &r.MainID, &r.DetailID, &r.StakeType, &r.Stake, &r.Income, &r.Status); err != nil {
			return
		}

		list = append(list, r)
	}

	err = rows.Err()

	return
}

func (d *Dao) ResetUserLog(ctx context.Context, mid, guessCount, income, succeedCount, business int64) (err error) {
	rate := succeedCount * 100 / guessCount
	_, err = d.db.Exec(ctx, sql4UpdateUserLog, rate, guessCount, income, succeedCount, mid, business)

	return
}

func (d *Dao) UpdateUserGuess(ctx context.Context, primaryKey, mid int64, income float64) (err error) {
	newIncome := int64(income * 100)
	_, err = d.db.Exec(ctx, fmt.Sprintf(sql4UpdateUserGuess, userHit(mid)), newIncome, primaryKey)

	return
}

// RawMDsGuess get main detail options.
func (d *Dao) RawMDsGuess(c context.Context, ids []int64) (res map[int64][]*guess.DetailOption, err error) {
	var (
		rows *sql.Rows
		rs   []*guess.DetailOption
	)
	if len(ids) == 0 {
		return
	}
	rows, err = d.db.Query(c, fmt.Sprintf(_mdsOptionSQL, xstr.JoinInts(ids)))
	if err != nil {
		err = errors.Wrap(err, "RawMDsGuess:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.DetailOption)
		if err = rows.Scan(&r.MainID, &r.DetailID, &r.Option, &r.Oid); err != nil {
			err = errors.Wrap(err, "RawMDsGuess:rows.Scan() error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawMDsGuess:rows.Err")
		return
	}
	res = make(map[int64][]*guess.DetailOption)
	for _, v := range rs {
		if _, ok := res[v.MainID]; !ok {
			item := &guess.DetailOption{
				MainID:   v.MainID,
				DetailID: v.DetailID,
				Option:   v.Option,
				Oid:      v.Oid,
			}
			res[v.MainID] = []*guess.DetailOption{item}
		} else {
			res[v.MainID] = append(res[v.MainID], &guess.DetailOption{
				MainID:   v.MainID,
				DetailID: v.DetailID,
				Option:   v.Option,
				Oid:      v.Oid,
			})
		}
	}
	return
}
