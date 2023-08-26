package like

import (
	"context"
	xsql "database/sql"
	"fmt"
	"time"

	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const _newstartSelSQL = "SELECT id,mid,inviter_mid,ctime,is_identity FROM act_newstar WHERE v_status=1 AND finish_task=0 AND id>? ORDER BY id ASC limit 100"

// BigVUsers 获取大V用户
func (d *Dao) BigVUsers(ctx context.Context, id int64) (rs []*like.BigVUser, err error) {
	rs = []*like.BigVUser{}
	rows, err := d.db.Query(ctx, _newstartSelSQL, id)
	if err != nil {
		err = errors.Wrap(err, "BigVUsers:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &like.BigVUser{}
		err = rows.Scan(&r.ID, &r.Mid, &r.InviterMid, &r.Ctime, &r.IsIdentity)
		if err != nil {
			err = errors.Wrap(err, "BigVUsers:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "BigVUsers:rows.Err")
	}
	return
}

const _updateArcSQL = "UPDATE act_newstar SET up_archives = CASE %s END WHERE id IN (%s)"

// UpdateArcStatus
func (d *Dao) UpdateArcStatus(ctx context.Context, arcStatusMap map[int64]int64) (affected int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	if len(arcStatusMap) == 0 {
		return
	}
	for id, status := range arcStatusMap {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %d", caseStr, id, status)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(ctx, fmt.Sprintf(_updateArcSQL, caseStr, xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpdateArcStatus:db.Exec error")
		return
	}
	return res.RowsAffected()
}

const _finishTaskSQL = "UPDATE act_newstar SET is_name=?,is_mobile=?,is_identity=?,fans_count=?,up_archives=?,finish_time=?,finish_task=1 WHERE id=?"

// FinishBigV
func (d *Dao) FinishBigV(ctx context.Context, id, isName, isMobile, isIdentity, fansCount, upArchives int64) (int64, error) {
	row, err := d.db.Exec(ctx, _finishTaskSQL, isName, isMobile, isIdentity, fansCount, upArchives, time.Now().Unix(), id)
	if err != nil {
		return 0, errors.Wrap(err, "FinishBigV Exec")
	}
	return row.RowsAffected()
}

const _upIdentitySQL = "UPDATE act_newstar SET is_identity = CASE %s END WHERE finish_task = 0 AND id IN (%s)"

// UpIdentity
func (d *Dao) UpIdentity(ctx context.Context, idMap map[int64]int64) (rs int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	if len(idMap) == 0 {
		return
	}
	for id, status := range idMap {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %d", caseStr, id, status)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(ctx, fmt.Sprintf(_upIdentitySQL, caseStr, xstr.JoinInts(ids))); err != nil {
		err = errors.Wrap(err, "UpIdentity dao.db.Exec")
		return
	}
	if rs, err = res.RowsAffected(); err != nil {
		err = errors.Wrap(err, "UpIdentity res.RowsAffected")
	}
	return
}
