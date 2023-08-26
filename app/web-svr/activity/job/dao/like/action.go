package like

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/admin/model"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

const (
	_likeActStateURI        = "/x/internal/activity/likeact/state"
	_likeActSumSQL          = "SELECT SUM(`action`) AS `like`,lid FROM like_action WHERE lid IN(%s) GROUP BY lid"
	_likeActSumRangeTimeSQL = "SELECT SUM(`action`) AS `like`,lid FROM like_action WHERE lid IN(%s) and ctime >= '%s' and ctime <= '%s' GROUP BY lid"
	_likeActListSQL         = "SELECT id,mid FROM like_action WHERE lid = ? AND id > ? ORDER BY id LIMIT ?"
)

// BatchLikeActSum .
func (d *Dao) BatchLikeActSum(c context.Context, lids []int64) (res map[int64]int64, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_likeActSumSQL, xstr.JoinInts(lids)))
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "d.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		like := sql.NullInt64{}
		lid := sql.NullInt64{}
		if err = rows.Scan(&like, &lid); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res[lid.Int64] = like.Int64
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

// LikeActList .
func (d *Dao) LikeActList(c context.Context, lid, minID, limit int64) (res []*model.LikeAction, err error) {
	rows, err := d.db.Query(c, _likeActListSQL, lid, minID, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "d.db.Query()")
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		action := new(model.LikeAction)
		if err = rows.Scan(&action.ID, &action.Mid); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res = append(res, action)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

// LikeActState .
func (d *Dao) LikeActState(c context.Context, sid, mid int64, lids []int64) (data map[int64]int, err error) {
	var res struct {
		Code int           `json:"code"`
		Data map[int64]int `json:"data"`
	}
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("lids", xstr.JoinInts(lids))
	if err = d.httpClient.Get(c, d.likeActStateURL, "", params, &res); err != nil {
		log.Error("LikeActState:d.httpClient.Get sid(%d) mid(%d) lids(%v) error(%v)", sid, mid, lids, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.likeActStateURL+"?"+params.Encode())
	}
	data = res.Data
	return
}

// BatchLikeActSumRangeTime .
func (d *Dao) BatchLikeActSumRangeTime(c context.Context, lids []int64, sTime, eTime string) (res map[int64]int64, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_likeActSumRangeTimeSQL, xstr.JoinInts(lids), sTime, eTime))
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "d.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		like := sql.NullInt64{}
		lid := sql.NullInt64{}
		if err = rows.Scan(&like, &lid); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res[lid.Int64] = like.Int64
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}
