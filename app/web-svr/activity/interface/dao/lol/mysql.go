package lol

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/log"
	"go-common/library/xstr"
	guemdl "go-gateway/app/web-svr/activity/interface/model/guess"
	lolmdl "go-gateway/app/web-svr/activity/interface/model/lol"
)

const (
	_mainsGuessSQL = "SELECT id,business,oid,title,stake_type,max_stake,result_id,guess_count,is_deleted,ctime,mtime,stime,etime FROM act_guess_main WHERE is_deleted = 0 AND oid IN (%s)"
	_userGuessSQL  = "SELECT id,main_id,detail_id,stake,income,status FROM act_guess_user_%s WHERE mid = ? AND main_id in (%s) ORDER BY id DESC"

	sql4AllUnSettlementContestIDList = `
SELECT oid, settlement_status
FROM act_guess_main
WHERE settlement_status IN(0, 1)
`
)

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%100)
}

func UnSettlementContestIDList(ctx context.Context) (m map[int64]int64, err error) {
	rows, err := component.S10GlobalDB.Query(ctx, sql4AllUnSettlementContestIDList)
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	m = make(map[int64]int64, 0)
	for rows.Next() {
		var contestID, status int64
		err = rows.Scan(&contestID, &status)
		if err != nil {
			return
		}

		m[contestID] = status
	}
	err = rows.Err()

	return
}

// MainList guess main list.
func (d *Dao) MainList(ctx context.Context, oids []int64) (rs []*guemdl.MainGuess, err error) {
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_mainsGuessSQL, xstr.JoinInts(oids)))
	if err != nil {
		log.Error("d.MainList.Query oids:%+v error(%+v)", oids, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		mod := &guemdl.MainGuess{}
		err = rows.Scan(&mod.ID, &mod.Business, &mod.Oid, &mod.Title, &mod.StakeType, &mod.MaxStake, &mod.ResultID, &mod.GuessCount, &mod.IsDeleted, &mod.Ctime, &mod.Mtime, &mod.Stime, &mod.Etime)
		if err != nil {
			log.Error("d.MainList.Scan oids:%+v error(%+v)", oids, err)
			return
		}
		rs = append(rs, mod)
	}
	err = rows.Err()
	return
}

// RawUserGuessOid get user guess oid.
func RawUserGuessOid(ctx context.Context, mid int64, mainIDs []int64) (res []*lolmdl.UserGuessOid, err error) {
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_userGuessSQL, userHit(mid), xstr.JoinInts(mainIDs)), mid)
	if err != nil {
		log.Error("d.RawUserGuess.Query mainIDs:%+v error(%+v)", mainIDs, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		mod := &lolmdl.UserGuessOid{}
		err = rows.Scan(&mod.ID, &mod.MainID, &mod.DetailID, &mod.Stake, &mod.Income, &mod.Status)
		if err != nil {
			log.Error("d.RawUserGuess.Scan mainIDs:%+v error(%+v)", mainIDs, err)
			return
		}
		res = append(res, mod)
	}
	err = rows.Err()
	return
}
