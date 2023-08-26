package guess

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
)

// AddMatchGuess .
func (d *Dao) AddMatchGuess(c context.Context, p *api.GuessAddReq) (err error) {
	var (
		tx     *sql.Tx
		mainID int64
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("d.db.Begin() error(%v)", err)
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	for _, group := range p.Groups {
		if mainID, err = d.AddMainGuess(c, tx, p.Business, p.Oid, p.MaxStake, p.StakeType, group.Title, p.Stime, p.Etime, group.TemplateType); err != nil {
			log.Error("d.AddMainGues params(%d,%d,%d,%s) error(%+v)", p.Business, p.Oid, p.MaxStake, group.Title, err)
			return
		}
		if err = d.BatchAddDetail(c, tx, mainID, group.DetailAdd); err != nil {
			log.Error("d.BatchAddDetail mainID(%d) error(%+v)", mainID, err)
			return
		}
	}
	return
}

// UserAddGuess user add guess.
func (d *Dao) UserAddGuess(c context.Context, business, mainID int64, p *api.GuessUserAddReq) (userLogID int64, err error) {
	var (
		tx    *sql.Tx
		count int64
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("d.db.Begin() error(%v)", err)
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// add user record
	if userLogID, err = d.AddGuess(c, tx, mainID, p, business); err != nil {
		log.Error("UserAddGuess d.AddGuess mainID(%d) mid(%d) error(%+v)", mainID, p.Mid, err)
		err = ecode.ActGuessFail
		return
	}
	// update main guess_count
	if count, err = d.UpMainCount(c, tx, mainID); err != nil {
		log.Error("UserAddGuess d.UpUserLog mainID(%d) mid(%d)  count(%d) error(%+v)", mainID, p.Mid, count, err)
		err = ecode.ActGuessDataFail
		return
	}
	// update detail total_stake
	if count, err = d.UpDetailTotal(c, tx, p.Stake, p.DetailID); err != nil {
		log.Error("UserAddGuess d.UpUserLog mainID(%d) mid(%d)  count(%d) error(%+v)", mainID, p.Mid, count, err)
		err = ecode.ActGuessDataFail
		return
	}
	// user log
	if _, err = d.UserStatUp(c, tx, business, p); err != nil {
		log.Error("UserAddGuess d.UserStatUp mainID(%d) mid(%d) error(%+v)", mainID, p.Mid, err)
		err = ecode.ActGuessDataFail
	}
	return
}
