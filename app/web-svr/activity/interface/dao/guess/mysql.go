package guess

import (
	"context"
	xsql "database/sql"
	"fmt"
	"math"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/guess"

	"github.com/pkg/errors"
)

const (
	_userSub     = 100
	_userViewMax = 1000
	_twoDecimal  = 100.00
	_mainAddSQL  = `
INSERT INTO act_guess_main (business, oid, max_stake, stake_type, title
	, stime, etime, template_type)
VALUES (?, ?, ?, ?, ?
	, ?, ?, ?)`

	_mainUpSQL   = "UPDATE act_guess_main SET stime = ?,etime = ? WHERE is_deleted = 0 AND business = ? AND oid = ?"
	_mdResultSQL = `
SELECT m.id, m.business, m.oid, m.title, m.stake_type
	, m.max_stake, m.result_id, m.guess_count, m.stime, m.etime, m.template_type
	, d.id AS detail_id, d.option, d.odds, d.total_stake, m.is_deleted
FROM act_guess_main m
	INNER JOIN act_guess_detail d ON m.id = d.main_id
WHERE m.business = ?
	AND m.id = ?
ORDER BY d.id
`
	_mdsResultSQL = `
SELECT m.id, m.business, m.oid, m.title, m.stake_type
	, m.max_stake, m.result_id, m.guess_count, m.stime, m.etime, m.template_type
	, d.id AS detail_id, d.option, d.odds, d.total_stake, m.is_deleted
FROM act_guess_main m
	INNER JOIN act_guess_detail d ON m.id = d.main_id
WHERE m.business = ?
	AND m.id IN (%s)
ORDER BY m.id, d.id
`
	_oMIDsSQL          = "SELECT id,oid,is_deleted FROM act_guess_main WHERE business = ? AND oid = ? ORDER BY id"
	_osMIDsSQL         = "SELECT id,oid,is_deleted FROM act_guess_main WHERE business = ? AND oid in (%s) ORDER BY oid,id"
	_mainGuessSQL      = "SELECT id,business,oid,title,stake_type,max_stake,result_id,guess_count,is_deleted,ctime,mtime,stime,etime FROM act_guess_main WHERE is_deleted = 0 AND id = ?"
	_haveGuessSQL      = "SELECT id,option,odds,total_stake,ctime,mtime FROM act_guess_detail WHERE main_id = ? AND id = ?"
	_mainDelSQL        = "UPDATE act_guess_main SET is_deleted = 1 WHERE result_id = 0 AND is_deleted = 0 AND id = ?"
	_mainResultSQL     = "UPDATE act_guess_main SET result_id = ? WHERE result_id = 0 AND is_deleted = 0 AND id = ?"
	_detailBatchAddSQL = "INSERT INTO act_guess_detail (main_id,option,total_stake) VALUES %s"
	_guessAddSQL       = "INSERT INTO act_guess_user_%s (mid,main_id,detail_id,stake_type,stake,business) VALUES (?,?,?,?,?,?)"
	_userStatSQL       = "SELECT id,business,mid,total_guess,total_success,success_rate,stake_type,total_stake,total_income,ranking,ctime,mtime FROM act_guess_user_log WHERE business = ? AND mid = ? AND stake_type = ?"
	_userStatUpSQL     = "INSERT INTO act_guess_user_log (business,mid,stake_type,total_stake,total_guess) VALUES (?,?,?,?,1) ON DUPLICATE KEY UPDATE total_stake = total_stake + ? ,total_guess = total_guess + 1"
	_mainCountUpSQL    = "UPDATE act_guess_main SET guess_count = guess_count + 1 WHERE id = ?"
	_detailTotalUpSQL  = "UPDATE act_guess_detail SET total_stake = total_stake + ? WHERE id = ?"
	_userGuessesSQL    = "SELECT id,mid,main_id,detail_id,stake_type,stake,income,status,ctime FROM act_guess_user_%s WHERE business = ? AND mid = ? ORDER BY id DESC LIMIT ?"
	_userGuessSQL      = "SELECT id,mid,main_id,detail_id,stake_type,stake,income,status,ctime FROM act_guess_user_%s WHERE mid = ? AND main_id in (%s) ORDER BY id DESC"
)

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%_userSub)
}

// AddGuess user add guess.
func (d *Dao) AddGuess(c context.Context, tx *sql.Tx, mainID int64, p *api.GuessUserAddReq, business int64) (lastID int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(fmt.Sprintf(_guessAddSQL, userHit(p.Mid)), p.Mid, mainID, p.DetailID, p.StakeType, p.Stake, business); err != nil {
		err = errors.Wrap(err, "AddGuess: db.Exec")
		return
	}
	return res.LastInsertId()
}

// AddMainGuess add  guess main.
func (d *Dao) AddMainGuess(c context.Context, tx *sql.Tx, business, oid, maxStake, stakeType int64, title string, stime, etime, templateType int64) (lastID int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(_mainAddSQL, business, oid, maxStake, stakeType, title, stime, etime, templateType); err != nil {
		err = errors.Wrap(err, "AddMainGuess: tx.Exec")
		return
	}
	return res.LastInsertId()
}

// UpGuess update guess .
func (d *Dao) UpGuess(c context.Context, p *api.GuessEditReq) (rs int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, _mainUpSQL, p.Stime, p.Etime, p.Business, p.Oid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpGuess res.RowsAffected")
	}
	return
}

// BatchAddDetail add guess details.
func (d *Dao) BatchAddDetail(c context.Context, tx *sql.Tx, mainID int64, detailAdd []*api.GuessDetailAdd) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, detail := range detailAdd {
		rowsTmp = append(rowsTmp, "(?,?,?)")
		rows = append(rows, mainID, detail.Option, detail.TotalStake)
	}
	sql := fmt.Sprintf(_detailBatchAddSQL, strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddDetail: tx.Exec")
	}
	return
}

// DelGroup set main .
func (d *Dao) DelGroup(c context.Context, mainID int64) (rs int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, _mainDelSQL, mainID); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "DelGroup res.RowsAffected")
	}
	return
}

// UpGuessResult update main result detail id .
func (d *Dao) UpGuessResult(c context.Context, mainID, DetailID int64) (rs int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, _mainResultSQL, DetailID, mainID); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpGuessResult res.RowsAffected")
	}
	return
}

// RawMDResult get main detail lists.
func (d *Dao) RawMDResult(c context.Context, id int64, business int64) (res *guess.MainRes, err error) {
	var (
		rows *sql.Rows
		rs   []*guess.MainDetail
	)
	rows, err = d.db.Query(c, _mdResultSQL, business, id)
	if err != nil {
		err = errors.Wrap(err, "RawMDResult:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.MainDetail)
		if err = rows.Scan(&r.ID, &r.Business, &r.Oid, &r.Title, &r.StakeType, &r.MaxStake, &r.ResultID, &r.GuessCount, &r.Stime, &r.Etime, &r.TemplateType, &r.DetailID, &r.Option, &r.Odds, &r.TotalStake, &r.IsDeleted); err != nil {
			err = errors.Wrap(err, "RawMDResult:rows.Scan() error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawMDResult:rows.Err")
		return
	}
	res = new(guess.MainRes)
	first := true
	for _, v := range rs {
		if first {
			item := &guess.MainRes{
				MainGuess: &guess.MainGuess{
					ID:           v.ID,
					Business:     v.Business,
					Oid:          v.Oid,
					Title:        v.Title,
					StakeType:    v.StakeType,
					MaxStake:     v.MaxStake,
					ResultID:     v.ResultID,
					GuessCount:   v.GuessCount,
					Stime:        v.Stime,
					Etime:        v.Etime,
					IsDeleted:    v.IsDeleted,
					TemplateType: v.TemplateType,
				},
				Details: []guess.DetailGuess{{
					ID:         v.DetailID,
					Option:     v.Option,
					Odds:       v.Odds,
					TotalStake: v.TotalStake,
				}},
			}
			first = false
			res = item
		} else {
			res.Details = append(res.Details, guess.DetailGuess{
				ID:         v.DetailID,
				Option:     v.Option,
				Odds:       v.Odds,
				TotalStake: v.TotalStake,
			})
		}
	}
	return
}

// RawMDsResult get main detail lists.
func (d *Dao) RawMDsResult(c context.Context, ids []int64, business int64) (res map[int64]*guess.MainRes, err error) {
	var (
		rows *sql.Rows
		rs   []*guess.MainDetail
	)
	if len(ids) == 0 {
		return
	}
	rows, err = d.db.Query(c, fmt.Sprintf(_mdsResultSQL, xstr.JoinInts(ids)), business)
	if err != nil {
		err = errors.Wrap(err, "RawMDsResult:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.MainDetail)
		if err = rows.Scan(&r.ID, &r.Business, &r.Oid, &r.Title, &r.StakeType, &r.MaxStake, &r.ResultID, &r.GuessCount, &r.Stime, &r.Etime, &r.TemplateType, &r.DetailID, &r.Option, &r.Odds, &r.TotalStake, &r.IsDeleted); err != nil {
			err = errors.Wrap(err, "RawMDsResult:rows.Scan() error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawMDsResult:rows.Err")
		return
	}
	res = make(map[int64]*guess.MainRes)
	for _, v := range rs {
		if _, ok := res[v.ID]; !ok {
			item := &guess.MainRes{
				MainGuess: &guess.MainGuess{
					ID:           v.ID,
					Business:     v.Business,
					Oid:          v.Oid,
					Title:        v.Title,
					StakeType:    v.StakeType,
					MaxStake:     v.MaxStake,
					ResultID:     v.ResultID,
					GuessCount:   v.GuessCount,
					Stime:        v.Stime,
					Etime:        v.Etime,
					IsDeleted:    v.IsDeleted,
					TemplateType: v.TemplateType,
				},
				Details: []guess.DetailGuess{{
					ID:         v.DetailID,
					Option:     v.Option,
					Odds:       v.Odds,
					TotalStake: v.TotalStake,
				}},
			}

			if v.ResultID == v.DetailID {
				item.RightOption = v.Option
			}

			res[v.ID] = item
		} else {
			if v.ResultID == v.DetailID {
				res[v.ID].RightOption = v.Option
			}

			res[v.ID].Details = append(res[v.ID].Details, guess.DetailGuess{
				ID:         v.DetailID,
				Option:     v.Option,
				Odds:       v.Odds,
				TotalStake: v.TotalStake,
			})
		}
	}
	return
}

// RawOidMIDs get oid main ids.
func (d *Dao) RawOidMIDs(c context.Context, oid, business int64) (rs []*guess.MainID, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(c, _oMIDsSQL, business, oid)
	if err != nil {
		err = errors.Wrap(err, "RawOidMids:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.MainID)
		if err = rows.Scan(&r.ID, &r.OID, &r.IsDeleted); err != nil {
			err = errors.Wrap(err, "RawOidMids:rows.Scan() error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawOidMids:rows.Err")
	}
	return
}

// RawOidsMIDs get oids main ids.
func (d *Dao) RawOidsMIDs(c context.Context, oids []int64, business int64) (rs map[int64][]*guess.MainID, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_osMIDsSQL, xstr.JoinInts(oids)), business)
	if err != nil {
		err = errors.Wrap(err, "RawOidsMids:d.db.Query")
		return
	}
	defer rows.Close()
	rs = make(map[int64][]*guess.MainID, len(oids))
	for rows.Next() {
		r := new(guess.MainID)
		if err = rows.Scan(&r.ID, &r.OID, &r.IsDeleted); err != nil {
			err = errors.Wrap(err, "RawOidsMids:rows.Scan() error")
			return
		}
		rs[r.OID] = append(rs[r.OID], r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawOidsMids:rows.Err")
	}
	return
}

// RawGuessMain get main guess.
func (d *Dao) RawGuessMain(c context.Context, mainID int64) (mod *guess.MainGuess, err error) {
	mod = &guess.MainGuess{}
	row := d.db.QueryRow(c, _mainGuessSQL, mainID)
	if err = row.Scan(&mod.ID, &mod.Business, &mod.Oid, &mod.Title, &mod.StakeType, &mod.MaxStake, &mod.ResultID, &mod.GuessCount, &mod.IsDeleted, &mod.Ctime, &mod.Mtime, &mod.Stime, &mod.Etime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "MainGuess:row.Scan error")
		}
	}
	return
}

// RawUserStat user guess stat.
func (d *Dao) RawUserStat(c context.Context, mid, stakeType, business int64) (mod *api.UserGuessDataReply, err error) {
	mod = &api.UserGuessDataReply{}
	row := d.db.QueryRow(c, _userStatSQL, business, mid, stakeType)
	if err = row.Scan(&mod.Id, &mod.Business, &mod.Mid, &mod.TotalGuess, &mod.TotalSuccess, &mod.SuccessRate, &mod.StakeType, &mod.TotalStake, &mod.TotalIncome, &mod.Ranking, &mod.Ctime, &mod.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "UserLog:row.Scan error")
		}
	}
	mod.TotalIncome = decimal(mod.TotalIncome/_twoDecimal, 1)
	return
}

// UserStatUp .
func (d *Dao) UserStatUp(c context.Context, tx *sql.Tx, business int64, p *api.GuessUserAddReq) (int64, error) {
	res, err := tx.Exec(_userStatUpSQL, business, p.Mid, p.StakeType, p.Stake, p.Stake)
	if err != nil {
		err = errors.Wrap(err, "UserStatUp: db.Exec")
		return 0, err
	}
	return res.RowsAffected()
}

// UpMainCount update main guess_count.
func (d *Dao) UpMainCount(c context.Context, tx *sql.Tx, id int64) (rs int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(_mainCountUpSQL, id); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpMainCount res.RowsAffected")
	}
	return
}

// UpDetailTotal update detail total_stake.
func (d *Dao) UpDetailTotal(c context.Context, tx *sql.Tx, stake, id int64) (rs int64, err error) {
	var res xsql.Result
	if res, err = tx.Exec(_detailTotalUpSQL, stake, id); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpDetailTotal res.RowsAffected")
	}
	return
}

// RawUserGuessList get user guess list.
func (d *Dao) RawUserGuessList(c context.Context, mid, business int64) (res []*guess.UserGuessLog, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = d.db.Query(c, fmt.Sprintf(_userGuessesSQL, userHit(mid)), business, mid, _userViewMax)
	if err != nil {
		err = errors.Wrap(err, "RawUserGuessList:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(guess.UserGuessLog)
		if err = rows.Scan(&r.ID, &r.Mid, &r.MainID, &r.DetailID, &r.StakeType, &r.Stake, &r.Income, &r.Status, &r.Ctime); err != nil {
			err = errors.Wrap(err, "RawUserGuessList:rows.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawUserGuessList:rows.Err")
	}
	return
}

// RawUserGuess get user guess.
func (d *Dao) RawUserGuess(c context.Context, mainIDs []int64, mid int64) (res map[int64]*guess.UserGuessLog, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = d.db.Query(c, fmt.Sprintf(_userGuessSQL, userHit(mid), xstr.JoinInts(mainIDs)), mid)
	if err != nil {
		err = errors.Wrap(err, "UserGuess:d.db.Query")
		return
	}
	defer rows.Close()
	res = make(map[int64]*guess.UserGuessLog, len(mainIDs))
	for rows.Next() {
		r := new(guess.UserGuessLog)
		if err = rows.Scan(&r.ID, &r.Mid, &r.MainID, &r.DetailID, &r.StakeType, &r.Stake, &r.Income, &r.Status, &r.Ctime); err != nil {
			err = errors.Wrap(err, "UserGuess:rows.Scan() error")
			return
		}
		res[r.MainID] = r
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "UserGuess:rows.Err")
	}
	return
}

// HaveGuess have guess.
func (d *Dao) HaveGuess(c context.Context, mainID, detailID int64) (r *guess.DetailGuess, err error) {
	r = &guess.DetailGuess{}
	row := d.db.QueryRow(c, _haveGuessSQL, mainID, detailID)
	if err = row.Scan(&r.ID, &r.Option, &r.Odds, &r.TotalStake, &r.Ctime, &r.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "HaveGuess:row.Scan error")
		}
	}
	return
}

func decimal(f float32, n int) float32 {
	n10 := math.Pow10(n)
	rs64 := math.Trunc((float64(f)+0.5/n10)*n10) / n10
	return float32(rs64)
}
