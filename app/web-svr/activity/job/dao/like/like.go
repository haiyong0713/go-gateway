package like

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const (
	_selLikeSQL                  = "SELECT id,wid FROM likes WHERE state=1 AND sid=? ORDER BY type"
	_likeItemsSQL                = "SELECT id,wid,sid,type,mid,state FROM likes FORCE INDEX(ix_like_0)  WHERE sid = ? AND id > ? ORDER BY id LIMIT ?"
	_likeListSQL                 = "SELECT id,wid,stick_top FROM likes WHERE state= 1 AND sid = ? ORDER BY id LIMIT ?,?"
	_likeWidsByAllSidsSQL        = "SELECT id,wid,mid,stick_top,state FROM likes WHERE sid IN(%s) ORDER BY id LIMIT ?,?"
	_likeWidsByNormalSidsSQL     = "SELECT id,wid,mid,stick_top,state FROM likes WHERE state=1 and sid IN(%s) ORDER BY id LIMIT ?,?"
	_likeWidSQL                  = "SELECT id,wid,sid,type,mid,state FROM likes WHERE sid = ? AND type = ? AND wid = ?"
	_likeAidSQL                  = "SELECT `id` From likes WHERE sid = ? and wid = ? and `state` != -1"
	_likeListStateSQL            = "SELECT id,wid,state FROM likes WHERE sid = ? AND state = 1"
	_likesCntSQL                 = "SELECT COUNT(1) AS cnt FROM likes WHERE state = 1 AND sid = ?"
	_likesListStateSQL           = "SELECT id,wid,state FROM likes WHERE sid IN(%s) AND state = 1"
	_likeTypeCountSQL            = "SELECT type,COUNT(1) FROM likes WHERE sid=? AND state=1 GROUP BY type"
	_likeDistinctMidSQL          = "SELECT DISTINCT(mid) FROM likes WHERE sid=? AND state=1"
	_likeDistinctMidAllSQL       = "SELECT DISTINCT(mid) FROM likes WHERE sid=?"
	_likeDistinctMidAllBySidsSQL = "SELECT DISTINCT(mid) FROM likes WHERE sid IN(%s) LIMIT ?,?"
	_likeGetIdWidStateByPageSQL  = "SELECT id,wid,state FROM likes force index (`ix_like_0`) WHERE sid = ? and state = 1 order by id asc limit ? offset ?"
	_setObjStatURI               = "/x/internal/activity/object/stat/set"
	_setViewRankURI              = "/x/internal/activity/view/rank/set"
	_setLikeContentURI           = "/x/internal/activity/like/content/set"
	_likeOidsInfoURI             = "/x/internal/activity/oids/info"
	_likeUpURI                   = "/x/internal/activity/clear/like/up"
	_likeCtimeURI                = "/x/internal/activity/clear/like/ctime"
	_delLikeCtimeURI             = "/x/internal/activity/clear/like/del/ctime"
	_actLikeReloadURI            = "/x/internal/activity/clear/like/reload"
	_upListHisAddURI             = "/x/internal/activity/up/list/his/add"
)

// Like get like by sid
func (d *Dao) Like(c context.Context, sid int64) (ns []*like.Like, err error) {
	rows, err := d.db.Query(c, _selLikeSQL, sid)
	if err != nil {
		log.Error("notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &like.Like{}
		if err = rows.Scan(&n.ID, &n.Wid); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	return
}

// GetIdWidStateBySQL
func (d *Dao) GetIdWidStateBySQL(c context.Context, sid, limit, offset int64) (ns []*like.Like, err error) {
	rows, err := d.db.Query(c, _likeGetIdWidStateByPageSQL, sid, limit, offset)
	if err != nil {
		log.Errorc(c, "notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &like.Like{}
		if err = rows.Scan(&n.ID, &n.Wid, &n.State); err != nil {
			log.Errorc(c, "row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	return
}

// LikeListState get like list by sid.
func (d *Dao) LikeListState(c context.Context, sid int64) (list []*like.Like, err error) {
	rows, err := d.db.Query(c, _likeListStateSQL, sid)
	if err != nil {
		err = errors.Wrapf(err, "LikeListState:d.db.Query(%d)", sid)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Like)
		if err = rows.Scan(&n.ID, &n.Wid, &n.State); err != nil {
			err = errors.Wrapf(err, "LikeListState:row.Scan row (%d)", sid)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikeListState:rowsErr(%d)", sid)
	}
	return
}

// LikesAllList get likes list by sids.
func (d *Dao) LikesAllList(c context.Context, sids []int64, offset, limit int) (list []*like.Like, err error) {
	var (
		args []string
		sqls []interface{}
	)
	for _, sid := range sids {
		args = append(args, "?")
		sqls = append(sqls, sid)
	}
	sqls = append(sqls, offset, limit)
	rows, err := d.db.Query(c, fmt.Sprintf(_likeWidsByAllSidsSQL, strings.Join(args, ",")), sqls...)
	if err != nil {
		err = errors.Wrapf(err, "LikesAllList:d.db.Query(%v)", sids)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Like)
		if err = rows.Scan(&n.ID, &n.Wid, &n.Mid, &n.StickTop, &n.State); err != nil {
			err = errors.Wrapf(err, "LikesAllList:row.Scan row (%v)", sids)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikesAllList:rowsErr(%v)", sids)
	}
	return
}

// LikesNormalList get likes list by sids.
func (d *Dao) LikesNormalList(c context.Context, sids []int64, offset, limit int) (list []*like.Like, err error) {
	var (
		args []string
		sqls []interface{}
	)
	for _, sid := range sids {
		args = append(args, "?")
		sqls = append(sqls, sid)
	}
	sqls = append(sqls, offset, limit)
	rows, err := d.db.Query(c, fmt.Sprintf(_likeWidsByNormalSidsSQL, strings.Join(args, ",")), sqls...)
	if err != nil {
		err = errors.Wrapf(err, "LikesNormalList:d.db.Query(%v)", sids)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Like)
		if err = rows.Scan(&n.ID, &n.Wid, &n.Mid, &n.StickTop, &n.State); err != nil {
			err = errors.Wrapf(err, "LikesNormalList:row.Scan row (%v)", sids)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikesNormalList:rowsErr(%v)", sids)
	}
	return
}

// LikesListState get likes list by sids.
func (d *Dao) LikesListState(c context.Context, sids []int64) (list []*like.Like, err error) {
	var (
		args []string
		sqls []interface{}
	)
	for _, sid := range sids {
		args = append(args, "?")
		sqls = append(sqls, sid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_likesListStateSQL, strings.Join(args, ",")), sqls...)
	if err != nil {
		err = errors.Wrapf(err, "LikesListState:d.db.Query(%v)", sids)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Like)
		if err = rows.Scan(&n.ID, &n.Wid, &n.State); err != nil {
			err = errors.Wrapf(err, "LikeListState:row.Scan row (%v)", sids)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikeListState:rowsErr(%v)", sids)
	}
	return
}

// LikeList get like list by sid.
func (d *Dao) LikeList(c context.Context, sid int64, offset, limit int) (list []*like.Like, err error) {
	rows, err := d.db.Query(c, _likeListSQL, sid, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "LikeList:d.db.Query(%d,%d,%d)", sid, offset, limit)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Like)
		if err = rows.Scan(&n.ID, &n.Wid, &n.StickTop); err != nil {
			err = errors.Wrapf(err, "LikeList:row.Scan row (%d,%d,%d)", sid, offset, limit)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikeList:rowsErr(%d,%d,%d)", sid, offset, limit)
	}
	return
}

// LikeItems  by sid.
func (d *Dao) LikeItems(c context.Context, sid int64, likeID, limit int) (list []*like.ObjItem, err error) {
	rows, err := d.db.Query(c, _likeItemsSQL, sid, likeID, limit)
	if err != nil {
		err = errors.Wrapf(err, "LikeItems:d.db.Query(%d,%d,%d)", sid, likeID, limit)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.ObjItem)
		if err = rows.Scan(&n.ID, &n.Wid, &n.Sid, &n.Type, &n.Mid, &n.State); err != nil {
			err = errors.Wrapf(err, "LikeItems:row.Scan row (%d,%d,%d)", sid, likeID, limit)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikeItems:rowsErr(%d,%d,%d)", sid, likeID, limit)
	}
	return
}

// LikeWidItem .
func (d *Dao) LikeWidItem(c context.Context, sid, wid int64, typ int) (data *like.ObjItem, err error) {
	row := d.db.QueryRow(c, _likeWidSQL, sid, typ, wid)
	data = new(like.ObjItem)
	if err = row.Scan(&data.ID, &data.Wid, &data.Sid, &data.Type, &data.Mid, &data.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "LikeWidItem:QueryRow(%d,%d)", sid, wid)
		}
	}
	return
}

// LikeAidItem .
func (d *Dao) LikeAidItem(c context.Context, sid, aid int64) (data *like.ObjItem, err error) {
	row := d.db.QueryRow(c, _likeAidSQL, sid, aid)
	data = new(like.ObjItem)
	if err = row.Scan(&data.ID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			data = nil
		} else {
			err = errors.Wrapf(err, "LikeAidItem:QueryRow(%d,%d)", sid, aid)
		}
	}
	return
}

// LikeCnt get like list total count by sid.
func (d *Dao) LikeCnt(c context.Context, sid int64) (count int, err error) {
	row := d.db.QueryRow(c, _likesCntSQL, sid)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "LikeCnt:QueryRow(%d)", sid)
		}
	}
	return
}

// SetObjectStat .
func (d *Dao) SetObjectStat(c context.Context, sid int64, stat *like.SubjectTotalStat, count int) (err error) {
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("like", strconv.FormatInt(stat.SumLike, 10))
	params.Set("view", strconv.FormatInt(stat.SumView, 10))
	params.Set("fav", strconv.FormatInt(stat.SumFav, 10))
	params.Set("coin", strconv.FormatInt(stat.SumCoin, 10))
	params.Set("count", strconv.Itoa(count))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpClient.Get(c, d.setObjStatURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "SetObjectStat(%d,%v)", sid, stat)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "SetObjectStat Code(%d,%v)", sid, stat)
	}
	return
}

// SetViewRank set view rank list.
func (d *Dao) SetViewRank(c context.Context, sid int64, aids []int64, typ string) (err error) {
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("aids", xstr.JoinInts(aids))
	if typ != "" {
		params.Set("type", typ)
	}
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpClient.Get(c, d.setViewRankURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "SetViewRank(%d,%v)", sid, aids)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "SetViewRank Code(%d,%v)", sid, aids)
	}
	return
}

// SetLikeContent .
func (d *Dao) SetLikeContent(c context.Context, id int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("lid", strconv.FormatInt(id, 10))
	if err = d.httpClient.Get(c, d.setLikeContentURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "SetLikeContent(%d)", id)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "SetLikeContent Code(%d)", id)
	}
	return
}

// LikeOidsInfo get like oids info.
func (d *Dao) LikeOidsInfo(c context.Context, typ int64, oids []int64) (data map[int64]*like.ObjItem, err error) {
	var res struct {
		Code int                     `json:"code"`
		Data map[int64]*like.ObjItem `json:"data"`
	}
	params := url.Values{}
	params.Set("type", strconv.FormatInt(typ, 10))
	params.Set("oids", xstr.JoinInts(oids))
	if err = d.httpClient.Get(c, d.likeOidsInfoURL, "", params, &res); err != nil {
		log.Error("LikeOidsInfo:d.httpClient.Get typ(%d) oids(%v) error(%v)", typ, oids, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.likeOidsInfoURL+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}

// LikeUp .
func (d *Dao) LikeUp(c context.Context, lid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("lid", strconv.FormatInt(lid, 10))
	if err = d.httpClient.Get(c, d.likeUpURL, "", params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", d.likeUpURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// AddLikeCtimeCache .
func (d *Dao) AddLikeCtimeCache(c context.Context, lid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("lid", strconv.FormatInt(lid, 10))
	if err = d.httpClient.Get(c, d.likeCtimeURL, "", params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", d.likeCtimeURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// ActSetReload .
func (d *Dao) ActSetReload(c context.Context, lid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("lid", strconv.FormatInt(lid, 10))
	if err = d.httpClient.Get(c, d.actLikeReloadURL, "", params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", d.actLikeReloadURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// DelLikeCtimeCache .
func (d *Dao) DelLikeCtimeCache(c context.Context, lid, sid int64, likeType int) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("lid", strconv.FormatInt(lid, 10))
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("like_type", strconv.Itoa(likeType))
	if err = d.httpClient.Get(c, d.delLikeCtimeURL, "", params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", d.delLikeCtimeURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// UpListHisAdd
func (d *Dao) UpListHisAdd(c context.Context, sid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	if err = d.httpClient.Post(c, d.likeHisAddURL, "", params, &res); err != nil {
		log.Error("UpListHisAdd:d.httpClient.Post sid(%d) error(%v)", sid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.likeHisAddURL+"?"+params.Encode())
	}
	return
}

func (d *Dao) RawLikeTypeCount(ctx context.Context, sid int64) (map[int64]int64, error) {
	rows, err := d.db.Query(ctx, _likeTypeCountSQL, sid)
	if err != nil {
		err = errors.Wrapf(err, "RawLikeTypeCount:Query(%s,%d)", _likeTypeCountSQL, sid)
		return nil, err
	}
	defer rows.Close()
	counts := make(map[int64]int64)
	for rows.Next() {
		var typeID, count int64
		if err = rows.Scan(&typeID, &count); err != nil {
			err = errors.Wrapf(err, "RawLikeTypeCount:rows.Scan(%d)", sid)
			return nil, err
		}
		counts[typeID] = count
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RawLikeTypeCount:rows.Err(%d)", sid)
		return nil, err
	}
	return counts, nil
}

// AllMid get like all mid by sid
func (d *Dao) AllMid(c context.Context, sid int64) (ns []*like.Item, err error) {
	rows, err := d.db.Query(c, _likeDistinctMidSQL, sid)
	if err != nil {
		log.Error("notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &like.Item{}
		if err = rows.Scan(&n.Mid); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	return
}

// AllDistinctMid get like all mid by sid
func (d *Dao) AllDistinctMid(c context.Context, sid int64) (ns []*like.Item, err error) {
	rows, err := d.db.Query(c, _likeDistinctMidAllSQL, sid)
	if err != nil {
		log.Errorc(c, "notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &like.Item{}
		if err = rows.Scan(&n.Mid); err != nil {
			log.Errorc(c, "row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	return
}

// AllDistinctMidBySids get like all mid by sids
func (d *Dao) AllDistinctMidBySids(c context.Context, sids []int64, offset, limit int) (ns []*like.Item, err error) {
	if len(sids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_likeDistinctMidAllBySidsSQL, xstr.JoinInts(sids)), offset, limit)
	if err != nil {
		log.Error("notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &like.Item{}
		if err = rows.Scan(&n.Mid); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	err = rows.Err()
	return
}

const _upLikeStickTopSQL = "UPDATE likes SET stick_top=? WHERE id IN(%s)"

func (d *Dao) UpLikeStickTop(ctx context.Context, ids []int64, stickTop int) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_upLikeStickTopSQL, xstr.JoinInts(ids)), stickTop)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const _upLikeStickTopByWidSQL = "UPDATE likes SET stick_top=? WHERE sid=? AND wid=? LIMIT 1"

func (d *Dao) UpLikeStickTopByWid(ctx context.Context, sid, wid int64, stickTop int) (int64, error) {
	res, err := d.db.Exec(ctx, _upLikeStickTopByWidSQL, stickTop, sid, wid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const _likeScanSQL = "SELECT id, wid, sid, mid, state FROM likes WHERE sid IN (%s) AND id > ? ORDER BY id ASC LIMIT ?"

func (d *Dao) ScanLikesBySID(ctx context.Context, sids []int64, batch int, action func(context.Context, []*like.Item)) error {
	var id int64
	for {
		log.Infoc(ctx, "ScanLikesBySID d.db.Query(ctx, %v, %d, %d) error(%v)", sids, id, batch)
		rows, err := d.db.Query(ctx, fmt.Sprintf(_likeScanSQL, xstr.JoinInts(sids)), id, batch)
		if err != nil {
			log.Errorc(ctx, "ScanLikesBySID d.db.Query(ctx, %v, %d, %d) error(%v)", sids, id, batch, err)
			return err
		}
		defer rows.Close()
		ns := make([]*like.Item, 0, batch)
		for rows.Next() {
			n := &like.Item{}
			if err = rows.Scan(&n.ID, &n.Wid, &n.Sid, &n.Mid, &n.State); err != nil {
				log.Errorc(ctx, "ScanLikesBySID row.Scan error(%v)", err)
				return err
			}
			ns = append(ns, n)
		}
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "ScanLikesBySID row.Err error(%v)", err)
			return err
		}
		if len(ns) == 0 {
			break
		}
		id = ns[len(ns)-1].ID
		action(ctx, ns)
	}
	return nil
}
