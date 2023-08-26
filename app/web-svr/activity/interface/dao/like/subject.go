package like

import (
	"context"
	"fmt"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/queue/databus/report"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_selSubjectSQL                     = "SELECT s.id,s.name,s.dic,s.cover,s.stime,f.interval,f.ltime,f.tlimit FROM act_subject s INNER JOIN act_time_config f ON s.id=f.sid WHERE s.id = ?"
	_votLogSQL                         = "INSERT INTO act_online_vote_log(sid,aid,mid,stage,vote) VALUES(?,?,?,?,?)"
	_subjectNewestSQL                  = "SELECT id,ctime FROM act_subject WHERE state = 1 AND type IN (%s) AND stime <= ? AND etime >= ? ORDER BY ctime DESC LIMIT 1"
	_actSubjectSQL                     = "SELECT id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id,calendar,audit_platform FROM act_subject WHERE id = ? and state = 1"
	_actSubjectsSQL                    = "SELECT id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id,calendar,audit_platform FROM act_subject WHERE id IN (%s) and state = 1"
	_actSubjectWithStateSQL            = "SELECT id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id FROM act_subject WHERE id = ?"
	_actSubjectsWithStateSQL           = "SELECT id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id FROM act_subject WHERE id IN (%s)"
	_actSubjectsWithStateFromMasterSQL = "SELECT /*master*/ id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id FROM act_subject WHERE id IN (%s)"
	_subjectInitSQL                    = "SELECT id,name,dic,cover,stime,etime,flag,type,lstime,letime,act_url,uetime,ustime,level,h5_cover,rank,author,oid,state,ctime,mtime,like_limit,android_url,ios_url,daily_like_limit,daily_single_like_limit,up_level,up_uetime,up_ustime,up_score,fan_limit_max,fan_limit_min,month_score,year_score,child_sids,up_figure_score,relation_id FROM act_subject WHERE id > ? order by id asc limit 1000"
	_subjectMaxIDSQL                   = "SELECT id FROM act_subject order by id desc limit 1"
	_subjectsOnGoingSQL                = "SELECT id FROM act_subject WHERE state = 1 AND type IN (%s) AND stime <= ? AND etime >= ? ORDER BY ctime"
	_subjectsBeforeOrOnGoingSQL        = "SELECT id, relation_id, stime, etime FROM act_subject WHERE state = 1 AND type IN (%s) AND etime >= ?"
	_upActReserveWhiteListSQL          = "SELECT mid, type FROM up_act_reserve_white_list WHERE id > ? AND id <= ?"
	_upActReserveWhiteListCountSQL     = "SELECT `id` FROM up_act_reserve_white_list order by id desc limit 1"
	//SubjectValidState act_subject valid state
	SubjectValidState = 1
)

// Subject Dao sql
func (dao *Dao) Subject(c context.Context, sid int64) (n *like.Subject, err error) {
	rows := dao.subjectStmt.QueryRow(c, sid)
	n = &like.Subject{}
	if err = rows.Scan(&n.ID, &n.Name, &n.Dic, &n.Cover, &n.Stime, &n.Interval, &n.Ltime, &n.Tlimit); err != nil {
		if err == sql.ErrNoRows {
			n = nil
			err = nil
		} else {
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	return
}

// VoteLog Dao sql
func (dao *Dao) VoteLog(c context.Context, sid int64, aid int64, mid int64, stage int64, vote int64) (rows int64, err error) {
	rs, err := dao.voteLogStmt.Exec(c, sid, aid, mid, stage, vote)
	if err != nil {
		log.Error("d.VoteLog.Exec(%d, %d,%d, %d, %d) error(%v)", sid, aid, mid, stage, vote, err)
		return
	}
	rows, err = rs.RowsAffected()
	return
}

// NewestSubject get newest subject list.
func (dao *Dao) NewestSubject(c context.Context, typeIDs []int64) (res *like.SubItem, err error) {
	res = new(like.SubItem)
	now := time.Now()
	row := dao.db.QueryRow(c, fmt.Sprintf(_subjectNewestSQL, xstr.JoinInts(typeIDs)), now, now)
	if err = row.Scan(&res.ID, &res.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "NewestPage:QueryRow")
		}
	}
	return
}

// RawActSubject get act_subject by id .
func (dao *Dao) RawActSubject(c context.Context, id int64) (res *like.SubjectItem, err error) {
	res = new(like.SubjectItem)
	row := dao.db.QueryRow(c, _actSubjectSQL, id)
	if err = row.Scan(&res.ID, &res.Name, &res.Dic, &res.Cover, &res.Stime, &res.Etime, &res.Flag, &res.Type, &res.Lstime, &res.Letime, &res.ActURL, &res.Uetime, &res.Ustime, &res.Level, &res.H5Cover, &res.Rank, &res.Author, &res.Oid, &res.State, &res.Ctime, &res.Mtime, &res.LikeLimit, &res.AndroidURL, &res.IosURL, &res.DailyLikeLimit, &res.DailySingleLikeLimit, &res.UpLevel, &res.UpUetime, &res.UpUstime, &res.UpScore, &res.FanLimitMax, &res.FanLimitMin, &res.MonthScore, &res.YearScore, &res.ChildSids, &res.UpFigureScore, &res.RelationID, &res.Calendar, &res.AuditPlatform); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "ActSubject:QueryRow")
		}
	}
	return
}

// RawActSubjects batch get subject.
func (dao *Dao) RawActSubjects(c context.Context, ids []int64) (res map[int64]*like.SubjectItem, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_actSubjectsSQL, xstr.JoinInts(ids))); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawActSubjects:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]*like.SubjectItem)
	for rows.Next() {
		r := new(like.SubjectItem)
		if err = rows.Scan(&r.ID, &r.Name, &r.Dic, &r.Cover, &r.Stime, &r.Etime, &r.Flag, &r.Type, &r.Lstime, &r.Letime, &r.ActURL, &r.Uetime, &r.Ustime, &r.Level, &r.H5Cover, &r.Rank, &r.Author, &r.Oid, &r.State, &r.Ctime, &r.Mtime, &r.LikeLimit, &r.AndroidURL, &r.IosURL, &r.DailyLikeLimit, &r.DailySingleLikeLimit, &r.UpLevel, &r.UpUetime, &r.UpUstime, &r.UpScore, &r.FanLimitMax, &r.FanLimitMin, &r.MonthScore, &r.YearScore, &r.ChildSids, &r.UpFigureScore, &r.RelationID, &r.Calendar, &r.AuditPlatform); err != nil {
			err = errors.Wrap(err, "RawActSubjects:QueryRow")
			return
		}
		res[r.ID] = r
	}
	err = rows.Err()
	return
}

// SubjectListMoreSid get subject more sid .
func (dao *Dao) SubjectListMoreSid(c context.Context, minSid int64) (res []*like.SubjectItem, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _subjectInitSQL, minSid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "SubjectInitialize:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make([]*like.SubjectItem, 0, 1000)
	for rows.Next() {
		a := &like.SubjectItem{}
		if err = rows.Scan(&a.ID, &a.Name, &a.Dic, &a.Cover, &a.Stime, &a.Etime, &a.Flag, &a.Type, &a.Lstime, &a.Letime, &a.ActURL, &a.Uetime, &a.Ustime, &a.Level, &a.H5Cover, &a.Rank, &a.Author, &a.Oid, &a.State, &a.Ctime, &a.Mtime, &a.LikeLimit, &a.AndroidURL, &a.IosURL, &a.DailyLikeLimit, &a.DailySingleLikeLimit, &a.UpLevel, &a.UpUetime, &a.UpUstime, &a.UpScore, &a.FanLimitMax, &a.FanLimitMin, &a.MonthScore, &a.YearScore, &a.ChildSids, &a.UpFigureScore, &a.RelationID); err != nil {
			err = errors.Wrap(err, "SubjectInitialize:rows.Scan()")
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "SubjectInitialize:rows.Err()")
	}
	return
}

// SubjectMaxID get act_subject max id .
func (dao *Dao) SubjectMaxID(c context.Context) (res *like.SubjectItem, err error) {
	res = new(like.SubjectItem)
	row := dao.db.QueryRow(c, _subjectMaxIDSQL)
	if err = row.Scan(&res.ID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "SubjectMaxID:QueryRow")
		}
	}
	return
}

// SubjectsOnGoing get subject ids .
func (dao *Dao) RawSubjectsOnGoing(c context.Context, typeIDs []int64) (res []int64, err error) {
	var (
		rows *sql.Rows
		nowT = time.Now()
	)
	if rows, err = dao.db.Query(c, fmt.Sprintf(_subjectsOnGoingSQL, xstr.JoinInts(typeIDs)), nowT, nowT); err != nil {
		err = errors.Wrap(err, "SubjectsOnGoing:dao.db.Query()")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var i int64
		if err = rows.Scan(&i); err != nil {
			err = errors.Wrap(err, "SubjectsOnGoing:rows.Scan()")
			return
		}
		res = append(res, i)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "SubjectInitialize:rows.Err()")
	}
	return
}

// ActSubjectsOnGoing get data from cache if miss will call source method, then add to cache.
func (dao *Dao) ActSubjectsOnGoing(c context.Context, typeIds []int64) (res []int64, err error) {
	addCache := true
	res, err = dao.CacheActSubjectsOnGoing(c, typeIds)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		return
	}
	res, err = dao.RawSubjectsOnGoing(c, typeIds)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.AddCacheActSubjectsOnGoing(c, typeIds, miss)
	return
}

func (d *Dao) AddSubAwardLog(_ context.Context, business int, action string, oid, mid int64) error {
	ui := &report.UserInfo{
		Mid:      mid,
		Platform: "",
		Build:    0,
		Buvid:    "",
		Business: business,
		Type:     0,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		IP:       "",
	}
	return report.User(ui)
}

func (dao *Dao) RawSubjectsBeforeOrOnGoing(c context.Context, typeIDs []int64, ts int64) (res []*like.SubjectItem, err error) {
	res = make([]*like.SubjectItem, 0)
	var rows *sql.Rows

	if rows, err = dao.db.Query(c, fmt.Sprintf(_subjectsBeforeOrOnGoingSQL, xstr.JoinInts(typeIDs)), ts); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		} else {
			err = errors.Wrap(err, "RawSubjectsBeforeOrOnGoing:dao.db.Query()")
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		item := new(like.SubjectItem)
		if err = rows.Scan(&item.ID, &item.RelationID, &item.Stime, &item.Etime); err != nil {
			err = errors.Wrap(err, "SubjectsOnGoing:rows.Scan()")
			return
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawSubjectsBeforeOrOnGoing:rows.Err()")
	}
	return
}

// RawActSubjectWithSate get act_subject by id .
func (dao *Dao) RawActSubjectWithState(c context.Context, id int64) (res *like.SubjectItem, err error) {
	res = new(like.SubjectItem)
	row := dao.db.QueryRow(c, _actSubjectWithStateSQL, id)
	if err = row.Scan(&res.ID, &res.Name, &res.Dic, &res.Cover, &res.Stime, &res.Etime, &res.Flag, &res.Type, &res.Lstime, &res.Letime, &res.ActURL, &res.Uetime, &res.Ustime, &res.Level, &res.H5Cover, &res.Rank, &res.Author, &res.Oid, &res.State, &res.Ctime, &res.Mtime, &res.LikeLimit, &res.AndroidURL, &res.IosURL, &res.DailyLikeLimit, &res.DailySingleLikeLimit, &res.UpLevel, &res.UpUetime, &res.UpUstime, &res.UpScore, &res.FanLimitMax, &res.FanLimitMin, &res.MonthScore, &res.YearScore, &res.ChildSids, &res.UpFigureScore, &res.RelationID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "ActSubjectWithState:QueryRow")
		}
	}
	return
}

// RawActSubjectsWithState batch get subject.
func (dao *Dao) RawActSubjectsWithState(c context.Context, ids []int64) (res map[int64]*like.SubjectItem, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_actSubjectsWithStateSQL, xstr.JoinInts(ids))); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawActSubjectsWithState:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]*like.SubjectItem)
	for rows.Next() {
		r := new(like.SubjectItem)
		if err = rows.Scan(&r.ID, &r.Name, &r.Dic, &r.Cover, &r.Stime, &r.Etime, &r.Flag, &r.Type, &r.Lstime, &r.Letime, &r.ActURL, &r.Uetime, &r.Ustime, &r.Level, &r.H5Cover, &r.Rank, &r.Author, &r.Oid, &r.State, &r.Ctime, &r.Mtime, &r.LikeLimit, &r.AndroidURL, &r.IosURL, &r.DailyLikeLimit, &r.DailySingleLikeLimit, &r.UpLevel, &r.UpUetime, &r.UpUstime, &r.UpScore, &r.FanLimitMax, &r.FanLimitMin, &r.MonthScore, &r.YearScore, &r.ChildSids, &r.UpFigureScore, &r.RelationID); err != nil {
			err = errors.Wrap(err, "RawActSubjects:QueryRow")
			return
		}
		res[r.ID] = r
	}
	err = rows.Err()
	return
}

func (dao *Dao) RawActSubjectsWithStateFromMaster(c context.Context, ids []int64) (res map[int64]*like.SubjectItem, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_actSubjectsWithStateFromMasterSQL, xstr.JoinInts(ids))); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawActSubjectsWithState:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]*like.SubjectItem)
	for rows.Next() {
		r := new(like.SubjectItem)
		if err = rows.Scan(&r.ID, &r.Name, &r.Dic, &r.Cover, &r.Stime, &r.Etime, &r.Flag, &r.Type, &r.Lstime, &r.Letime, &r.ActURL, &r.Uetime, &r.Ustime, &r.Level, &r.H5Cover, &r.Rank, &r.Author, &r.Oid, &r.State, &r.Ctime, &r.Mtime, &r.LikeLimit, &r.AndroidURL, &r.IosURL, &r.DailyLikeLimit, &r.DailySingleLikeLimit, &r.UpLevel, &r.UpUetime, &r.UpUstime, &r.UpScore, &r.FanLimitMax, &r.FanLimitMin, &r.MonthScore, &r.YearScore, &r.ChildSids, &r.UpFigureScore, &r.RelationID); err != nil {
			err = errors.Wrap(err, "RawActSubjects:QueryRow")
			return
		}
		res[r.ID] = r
	}
	err = rows.Err()
	return
}

func (dao *Dao) GetUpActReserveWhiteList(c context.Context, start int64, end int64) (res map[int64][]int64, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _upActReserveWhiteListSQL, start, end); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "GetUpActReserveWhiteList:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64][]int64, 0)
	for rows.Next() {
		r := new(like.UpActReserveWhiteList)
		if err = rows.Scan(&r.Mid, &r.Type); err != nil {
			err = errors.Wrap(err, "UpActReserveWhiteList:QueryRow")
			return
		}
		if v, ok := res[r.Type]; ok {
			res[r.Type] = append(v, r.Mid)
		} else {
			res[r.Type] = append(make([]int64, 0), r.Mid)
		}
	}
	err = rows.Err()
	return
}

func (dao *Dao) GetUpActReserveWhiteListCount(c context.Context) (res int, err error) {
	row := dao.db.QueryRow(c, _upActReserveWhiteListCountSQL)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "GetUpActReserveWhiteListCount:QueryRow")
		}
	}
	return
}
