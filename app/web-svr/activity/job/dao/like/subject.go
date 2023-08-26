package like

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
	commonsql "go-common/library/database/sql"
)

const (
	_selSubjectSQL         = "SELECT s.id,s.name,s.dic,s.cover,s.stime,f.interval,f.ltime,f.tlimit FROM act_subject s INNER JOIN act_time_config f ON s.id=f.sid WHERE s.id=?"
	_inOnlineLogSQL        = "INSERT INTO act_online_vote_end_log(sid,aid,stage,yes,no) VALUES(?,?,?,?,?)"
	_subjectsSQL           = "SELECT id,name,dic,cover,stime,etime,type FROM act_subject WHERE state = 1 AND type IN (%s) AND stime <= ? AND etime>= ?"
	_subjectStatSQL        = "SELECT num FROM subject_stat WHERE sid = ?"
	_awardSubjectListSQL   = "SELECT id,name,etime,sid,type,source_id,source_expire,other_sids,task_id FROM act_award_subject WHERE state = 1 AND sid_type = 1 AND etime > ?"
	_subjectRuleSQL        = "SELECT id,sid,state,task_id FROM act_subject_rule WHERE id=?"
	_subjectDetailByIds    = "SELECT id,name,dic,cover,stime,etime,type,shield_flag FROM act_subject WHERE state = 1 AND id IN (%s) "
	_subjectRuleBySidsSQL  = "SELECT id, sid, category, type_ids, tags, state, attribute, task_id, rule_name, sids, coefficient,aid_source,aid_source_type FROM act_subject_rule WHERE sid IN (%s)"
	_subjectChildSQL       = "SELECT id,child_sids FROM act_subject WHERE id=?"
	_subjectUpdateStateSQL = "UPDATE act_subject set state = ? where id = ?"
	_addLotteryTimesURI    = "/matsuri/api/add/times"
	_delLikeURI            = "/api/likes/uplist/%d/"
	_subjectUpURI          = "/x/internal/activity/clear/subject/up"
	_awardSubURI           = "/x/internal/activity/award/subject"
	_upArcEventURI         = "/x/internal/arcevent/rule/activity"
)

// Subject subject
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

const _actSubjectSQL = "SELECT id, name, flag, child_sids,type FROM act_subject WHERE id=?"

func (dao *Dao) ActSubject(ctx context.Context, sid int64) (*like.ActSubject, error) {
	row := dao.db.QueryRow(ctx, _actSubjectSQL, sid)
	r := &like.ActSubject{}
	if err := row.Scan(&r.ID, &r.Name, &r.Flag, &r.ChildSids, &r.Type); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("ActSubject row.Scan error(%v)", err)
		return nil, err
	}
	return r, nil
}

func (dao *Dao) SubjectRule(c context.Context, id int64) (*like.SubjectRule, error) {
	row := dao.db.QueryRow(c, _subjectRuleSQL, id)
	r := &like.SubjectRule{}
	if err := row.Scan(&r.ID, &r.Sid, &r.State, &r.TaskID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("SubjectRule row.Scan error(%v)", err)
		return nil, err
	}
	return r, nil
}

func (dao *Dao) SubjectRulesBySids(c context.Context, sids []int64) ([]*like.SubjectRule, error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_subjectRuleBySidsSQL, xstr.JoinInts(sids)))
	if err != nil {
		err = errors.Wrapf(err, "SubjectRulesBySids:d.db.Query(%v)", sids)
		return nil, err
	}
	defer rows.Close()
	res := make([]*like.SubjectRule, 0, len(sids)*5)
	for rows.Next() {
		n := new(like.SubjectRule)
		if err = rows.Scan(&n.ID, &n.Sid, &n.Category, &n.TypeIds, &n.Tags, &n.State, &n.Attribute, &n.TaskID, &n.RuleName, &n.Sids, &n.Coefficient, &n.AidSource, &n.AidSourceType); err != nil {
			err = errors.Wrapf(err, "SubjectRulesBySids:row.Scan row (%v)", sids)
			return nil, err
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "SubjectList:rowsErr(%v)", sids)
	}

	return res, nil
}

// SubjectDetailByIds 详情
func (dao *Dao) SubjectDetailByIds(c context.Context, sids []int64) ([]*like.ActSubject, error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_subjectDetailByIds, xstr.JoinInts(sids)))
	if err != nil {
		err = errors.Wrapf(err, "SubjectDetailByIds:d.db.Query(%v)", sids)
		return nil, err
	}
	defer rows.Close()
	res := make([]*like.ActSubject, 0, len(sids)*5)
	for rows.Next() {
		n := new(like.ActSubject)
		if err = rows.Scan(&n.ID, &n.Name, &n.Dic, &n.Cover, &n.Stime, &n.Etime, &n.Type, &n.ShieldFlag); err != nil {
			err = errors.Wrapf(err, "SubjectDetailByIds:row.Scan row (%v)", sids)
			return nil, err
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "SubjectDetailByIds:rowsErr(%v)", sids)
	}
	return res, nil
}

func (dao *Dao) RawSubjectStat(c context.Context, sid int64) (total int64, err error) {
	row := dao.db.QueryRow(c, _subjectStatSQL, sid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		err = errors.Wrap(err, "QueryRow")
	}
	return
}

// InOnlinelog InOnlinelog
func (dao *Dao) InOnlinelog(c context.Context, sid, aid, stage, yes, no int64) (rows int64, err error) {
	rs, err := dao.inOnlineLog.Exec(c, sid, aid, stage, yes, no)
	if err != nil {
		log.Error("d.InOnlinelog.Exec(%d, %d, %d, %d, %d) error(%v)", sid, aid, stage, yes, no, err)
		return
	}
	return rs.RowsAffected()
}

// SubjectList get online subject list by type.
func (dao *Dao) SubjectList(c context.Context, types []int64, ts time.Time) (res []*like.ActSubject, err error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_subjectsSQL, xstr.JoinInts(types)), ts, ts)
	if err != nil {
		err = errors.Wrapf(err, "SubjectList:d.db.Query(%v,%d)", types, ts.Unix())
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.ActSubject)
		if err = rows.Scan(&n.ID, &n.Name, &n.Dic, &n.Cover, &n.Stime, &n.Etime, &n.Type); err != nil {
			err = errors.Wrapf(err, "SubjectList:row.Scan row (%v,%d)", types, ts.Unix())
			return
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "SubjectList:rowsErr(%v,%d)", types, ts.Unix())
	}
	return
}

// SubjectChild get online subject child by id.
func (dao *Dao) SubjectChild(c context.Context, ID int64) (r *like.SubjectChild, err error) {
	r = &like.SubjectChild{}
	row := dao.db.QueryRow(c, fmt.Sprintf(_subjectChildSQL), ID)
	if err != nil {
		err = errors.Wrapf(err, "SubjectChild:d.db.Query(%d)", ID)
		return
	}
	if err = row.Scan(&r.ID, &r.ChildIds); err != nil {
		err = errors.Wrapf(err, "SubjectChild:d.db.Query(%d)", ID)
		return
	}
	if r != nil {
		strintList := strings.Split(r.ChildIds, ",")
		for _, v := range strintList {
			childID, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				r.ChildIdsList = append(r.ChildIdsList, childID)
			}
		}
	}
	return
}

// ListFromEs .
func (dao *Dao) ListFromEs(c context.Context, arg *like.EsParams) (list []*like.EsItem, err error) {
	req := dao.es.NewRequest(_activity).Index(_activity)
	if arg.Sid > 0 {
		req.WhereEq("sid", arg.Sid)
	}
	if arg.State != -1 {
		req.WhereEq("state", arg.State)
	}
	if arg.Ps != 0 && arg.Pn != 0 {
		req.Ps(arg.Ps).Pn(arg.Pn)
	}
	if arg.Order != "" && arg.Sort != "" {
		req.Order(arg.Order, arg.Sort)
	}
	actResult := new(struct {
		Result []*like.EsItem `json:"result"`
	})
	req.Fields("id", "wid")
	if err = req.Scan(c, &actResult); err != nil || actResult == nil {
		log.Error("ListFromEs req.Scan error(%v)", err)
		return
	}
	list = actResult.Result
	return
}

// SubjectTotalStat total stat.
func (dao *Dao) SubjectTotalStat(c context.Context, sid int64) (rs *like.SubjectTotalStat, err error) {
	req := dao.es.NewRequest(_activity).Index(_activity).WhereEq("state", 1).WhereEq("sid", sid).Sum("click").Sum("likes").Sum("fav").Sum("coin")
	res := new(struct {
		Result struct {
			SumCoin []struct {
				Value float64 `json:"value"`
			} `json:"sum_coin"`
			SumFav []struct {
				Value float64 `json:"value"`
			} `json:"sum_fav"`
			SumLikes []struct {
				Value float64 `json:"value"`
			} `json:"sum_likes"`
			SumClick []struct {
				Value float64 `json:"value"`
			} `json:"sum_click"`
		}
		Page struct {
			Total int `json:"total"`
		}
	})
	if err = req.Scan(c, &res); err != nil || res == nil {
		log.Error("SearchArc req.Scan error(%v)", err)
		return
	}
	rs = &like.SubjectTotalStat{
		SumCoin: int64(res.Result.SumCoin[0].Value),
		SumFav:  int64(res.Result.SumFav[0].Value),
		SumLike: int64(res.Result.SumLikes[0].Value),
		SumView: int64(res.Result.SumClick[0].Value),
		Count:   res.Page.Total,
	}
	return
}

// AddLotteryTimes .
func (dao *Dao) AddLotteryTimes(c context.Context, sid, mid int64) (err error) {
	params := url.Values{}
	params.Set("act_id", strconv.FormatInt(sid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = dao.httpClient.Get(c, dao.addLotteryTimesURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", dao.addLotteryTimesURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// SubjectUp .
func (dao *Dao) SubjectUp(c context.Context, sid int64) (err error) {
	//_subjectUpURI
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = dao.httpClient.Get(c, dao.subjectUpURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "dao.httpClient.Get(%s)", dao.subjectUpURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// DelLikeState .
func (dao *Dao) DelLikeState(c context.Context, sid int64, lids []int64, state int, reply string) (err error) {
	params := url.Values{}
	params.Set("ids", xstr.JoinInts(lids))
	params.Set("state", strconv.Itoa(state))
	params.Set("reply", reply)
	var res struct {
		Code int `json:"code"`
	}
	url := fmt.Sprintf(dao.delLikeURL, sid)
	if err = dao.httpClient.Post(c, url, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "dao.client.Get(%s)", url+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

func (dao *Dao) AwardSubject(c context.Context, sid, mid int64) (err error) {
	params := url.Values{}
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = dao.httpClient.Post(c, dao.awardSubjectURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("AwardSubject dao.httpClient.Post sid:%d mid:%d error(%v)", sid, mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("AwardSubject res code(%d) sid:%d mid:%d", res.Code, sid, mid)
	}
	return
}

func (dao *Dao) AwardSubjectList(c context.Context, etime time.Time) ([]*like.AwardSubject, error) {
	rows, err := dao.db.Query(c, _awardSubjectListSQL, etime)
	if err != nil {
		log.Error("AwardSubjectList etime(%v) error(%v)", etime, err)
		return nil, err
	}
	defer rows.Close()
	var list []*like.AwardSubject
	for rows.Next() {
		r := new(like.AwardSubject)
		if err = rows.Scan(&r.ID, &r.Name, &r.Etime, &r.Sid, &r.Type, &r.SourceID, &r.SourceExpire, &r.OtherSids, &r.TaskID); err != nil {
			log.Error("AwardSubjectList etime(%v) error(%v)", etime, err)
			return nil, err
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("AwardSubjectList rows.Err etime(%v) error(%v)", etime, err)
		return nil, err
	}
	return list, nil
}

func (dao *Dao) UpArcEventRule(c context.Context, sid int64) (err error) {
	params := url.Values{}
	params.Set("id", strconv.FormatInt(sid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = dao.httpClient.Post(c, dao.upArcEventURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("UpArcEventRule dao.httpClient.Post sid:%d error(%v)", sid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("UpArcEventRule res code(%d) sid:%d", res.Code, sid)
		err = ecode.Int(res.Code)
	}
	return
}

func (d *Dao) TXUpdateActSubjectState(tx *commonsql.Tx, state int64, sid int64) (err error) {
	if _, err = tx.Exec(_subjectUpdateStateSQL, state, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

const _actSubjectFromMasterSQL = "SELECT /*master*/ id, name FROM act_subject WHERE id = ?"

func (dao *Dao) ActSubjectFromMaster(ctx context.Context, sid int64) (*like.ActSubject, error) {
	row := dao.db.QueryRow(ctx, _actSubjectFromMasterSQL, sid)
	r := &like.ActSubject{}
	if err := row.Scan(&r.ID, &r.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("ActSubjectFromMaster row.Scan error(%v)", err)
		return nil, err
	}
	return r, nil
}
