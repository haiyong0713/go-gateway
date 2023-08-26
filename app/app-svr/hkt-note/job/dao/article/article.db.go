package article

import (
	"context"
	"fmt"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/pkg/errors"
)

const (
	_artContentTb              = "article_content"
	_selectArtStatus           = "SELECT cvid,pub_status FROM article_detail WHERE cvid IN (%s) AND deleted=0 ORDER BY cvid,pub_version DESC"
	_selectCvidInArc           = "SELECT DISTINCT cvid FROM article_detail WHERE oid=? AND oid_type=? AND pub_status=? AND deleted=0"
	_selectLatestArtCont       = "SELECT cvid,mtime,deleted,pub_version FROM %s WHERE cvid=? ORDER BY pub_version DESC LIMIT 1"
	_selectArtContByVer        = "SELECT cvid,note_id,mid,content,tag,deleted,mtime,pub_version FROM %s WHERE cvid=? AND pub_version=? AND deleted=0 "
	_selectArtDetail           = "SELECT cvid,note_id,mid,oid,oid_type,title,summary,pub_status,pub_reason,pubtime,mtime,deleted,pub_version FROM article_detail WHERE %s=?"
	_selectLatestArtByNoteId   = "SELECT cvid,pub_version,pub_status FROM article_detail WHERE note_id=? AND mid=? AND oid=? AND oid_type=? AND pub_from=? AND deleted=0 ORDER BY pub_version DESC LIMIT 1"
	_filterDel                 = " AND deleted=0"
	_pubStatusEq               = " AND pub_status=?"
	_pubStatusNotEq            = " AND pub_status<>?"
	_versionOrder              = " ORDER BY pub_version DESC LIMIT 1"
	_insertArtDetail           = "INSERT IGNORE INTO article_detail(cvid,note_id,mid,oid,oid_type,pub_status,inject_time,title,summary,pub_version,auto_comment,pub_from) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	_insertArtContent          = "INSERT IGNORE INTO %s(cvid,note_id,content,tag,mid,pub_version,cont_len,img_cnt) VALUES (?,?,?,?,?,?,?,?)"
	_updateCommentInfo         = "UPDATE article_detail SET comment_info=? WHERE cvid=? AND pub_version=? AND pub_from=?"
	_updateArtPubStatus        = "UPDATE article_detail SET pub_status=?,pub_reason=?,pubtime=? WHERE cvid=? AND pub_version=?"
	_delArtContent             = "UPDATE %s SET deleted=1 WHERE cvid=? AND mid=?"
	_delArtDetail              = "UPDATE article_detail SET deleted=1 WHERE cvid=? AND mid=?"
	_selectPubSuccessArtDetail = "SELECT cvid,pub_status,pub_version FROM article_detail WHERE cvid = ? AND pub_status = ? AND deleted=0"
	_delArtContentByCvid       = "UPDATE %s SET deleted=1 WHERE cvid=?"
	_delArtDetailByCvid        = "UPDATE article_detail SET deleted=1 WHERE cvid=?"
)

func (d *Dao) UpdateCommentInfo(c context.Context, cvid, pubVer int64, content string) error {
	res, err := d.db.Exec(c, _updateCommentInfo, content, cvid, pubVer, article.PubFromReply)
	if err != nil {
		return errors.Wrapf(err, "UpdateCommentInfo cvid(%d) pubVer(%d) content(%s)", cvid, pubVer, content)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		return errors.Wrapf(ecode.NothingFound, "UpdateCommentInfo cvid(%d) pubVer(%d) content(%s)", cvid, pubVer, content)
	}
	return nil
}

func (d *Dao) LatestArtByNoteId(c context.Context, data *note.ReplyMsg) (cvid, pubVer, pubStatus int64, err error) {
	rows, e := d.db.Query(c, _selectLatestArtByNoteId, data.NoteId, data.Mid, data.Oid, note.OidTypeUGC, article.PubFromReply)
	if e != nil {
		if e == sql.ErrNoRows {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, errors.Wrapf(e, "LatestPass data(%+v)", data)
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&cvid, &pubVer, &pubStatus); err != nil {
			return 0, 0, 0, errors.Wrapf(e, "LatestPass data(%+v)", data)

		}
	}
	if err = rows.Err(); err != nil {
		return 0, 0, 0, errors.Wrapf(e, "LatestPass data(%+v)", data)
	}
	return cvid, pubVer, pubStatus, nil
}

func (d *Dao) ArtCountInArc(c context.Context, oid int64, oidType int) (int, error) {
	rows, err := d.db.Query(c, _selectCvidInArc, oid, oidType, article.PubStatusPassed)
	if err != nil {
		return 0, errors.Wrapf(err, "ArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	var cvids []int64
	defer rows.Close()
	for rows.Next() {
		var cvid int64
		if err = rows.Scan(&cvid); err != nil {
			return 0, errors.Wrapf(err, "ArtCountInArc oid(%d) oidType(%d)", oid, oidType)
		}
		cvids = append(cvids, cvid)
	}
	if err = rows.Err(); err != nil {
		return 0, errors.Wrapf(err, "ArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	if len(cvids) == 0 {
		return 0, nil
	}
	var filtered []int64
	if filtered, err = d.filterLockArts(c, cvids); err != nil {
		return 0, errors.Wrapf(err, "ArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	return len(filtered), nil
}

// 根据cvids获取artile_detail，过滤lock
func (d *Dao) filterLockArts(c context.Context, ids []int64) ([]int64, error) {
	selSql := fmt.Sprintf(_selectArtStatus, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, selSql)
	if err != nil {
		return nil, errors.Wrapf(err, "filterLockArts ids(%v)", ids)
	}
	var (
		filterCvids = make([]int64, 0, len(ids))
		exist       = make(map[int64]struct{})
	)
	defer rows.Close()
	for rows.Next() {
		var (
			cvid   int64
			status int64
		)
		if err = rows.Scan(&cvid, &status); err != nil {
			return nil, errors.Wrapf(err, "filterLockArts ids(%v)", ids)
		}
		if _, ok := exist[cvid]; ok {
			continue
		}
		exist[cvid] = struct{}{}
		if status == article.PubStatusLock {
			log.Warn("artInfo filterLockArts cvid(%d) locked,skip", cvid)
			continue
		}
		filterCvids = append(filterCvids, cvid)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "filterLockArts ids(%v)", ids)
	}
	return filterCvids, nil
}

// 是否在指定pub_version前成功发布过
func (d *Dao) GetPubSuccessCvidsBeforeAssignedVersion(c context.Context, cvid int64, pubVersion int64) (bool, error) {
	rows, err := d.db.Query(c, _selectPubSuccessArtDetail, cvid, article.PubStatusPassed)
	if err != nil {
		return false, errors.Wrapf(err, "GetPubSuccessCvidsBeforeAssignedVersion cvid(%v)", cvid)
	}

	defer rows.Close()
	for rows.Next() {
		var (
			cvid          int64
			status        int64
			curPubVersion int64
		)
		if err = rows.Scan(&cvid, &status, &curPubVersion); err != nil {
			return false, errors.Wrapf(err, "GetPubSuccessCvidsBeforeAssignedVersion cvid(%v)", cvid)
		}
		if curPubVersion < pubVersion {
			return true, nil
		}
	}
	if err = rows.Err(); err != nil {
		return false, errors.Wrapf(err, "GetPubSuccessCvidsBeforeAssignedVersion cvid(%v)", cvid)
	}
	return false, nil
}

func (d *Dao) UpPubStatus(c context.Context, val *article.ArtOriginalDB) error {
	newestArt, err := d.ArtDetail(c, val.Id, article.TpArtDetailCvid, 0, 0, true)
	log.Warnc(c, "UpPubStatus newestArt pubversion %v cvid %v ", newestArt.PubVersion, val.Id)
	if err != nil {
		return err
	}
	if newestArt.Cvid == -1 || newestArt.Deleted == 1 {
		return errors.Wrapf(ecode.NothingFound, "ArtDetail val(%+v)", val)
	}
	pubStatus := article.ToPubStatus(val.State)
	log.Warnc(c, "UpPubStatus val.State %v pubStatus %v cvid %v", val.State, pubStatus, val.Id)
	checkTime, _ := time.ParseInLocation("2006-01-02 15:04:05", val.CheckTime, time.Local)
	res, err := d.db.Exec(c, _updateArtPubStatus, pubStatus, val.Reason, xtime.Time(checkTime.Unix()), val.Id, newestArt.PubVersion)
	if err != nil {
		return errors.Wrapf(err, "UpPubStatus val(%+v)", val)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warn("artInfo UpPubStatus val(%+v) rowsAffected(%d) err(%+v)", val, rows, e)
	}
	return nil
}

func (d *Dao) DelArtDetail(c context.Context, cvid, mid int64) error {
	res, err := d.db.Exec(c, _delArtDetail, cvid, mid)
	if err != nil {
		return errors.Wrapf(err, "DelArtDetail cvid(%d) mid(%d)", cvid, mid)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warn("artInfo DelArtDetail cvid(%d) mid(%d) rowsAffected(%d) err(%+v)", cvid, mid, rows, e)
	}
	return nil
}

func (d *Dao) DelArtContent(c context.Context, cvid, mid int64) error {
	sql := fmt.Sprintf(_delArtContent, contTableName(cvid))
	res, err := d.db.Exec(c, sql, cvid, mid)
	if err != nil {
		return errors.Wrapf(err, "DelArtContent cvid(%d) mid(%d)", cvid, mid)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warn("artInfo DelArtContent cvid(%d) mid(%d) rowsAffected(%d) err(%+v)", cvid, mid, rows, e)
	}
	return nil
}

func (d *Dao) DelArtDetailByCvid(c context.Context, cvid int64) error {
	res, err := d.db.Exec(c, _delArtDetailByCvid, cvid)
	if err != nil {
		return errors.Wrapf(err, "DelArtDetailByCvid cvid(%d)", cvid)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warnc(c, "artInfo DelArtDetailByCvid cvid(%d) rowsAffected(%d) err(%+v)", cvid, rows, e)
	}
	return nil
}

func (d *Dao) DelArtContentByCvid(c context.Context, cvid int64) error {
	sql := fmt.Sprintf(_delArtContentByCvid, contTableName(cvid))
	res, err := d.db.Exec(c, sql, cvid)
	if err != nil {
		return errors.Wrapf(err, "DelArtContentByCvid cvid(%d)", cvid)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warnc(c, "artInfo DelArtContentByCvid cvid(%d) rowsAffected(%d) err(%+v)", cvid, rows, e)
	}
	return nil
}

func (d *Dao) InsertArtContent(c context.Context, val *article.ArtContCache) error {
	sql := fmt.Sprintf(_insertArtContent, contTableName(val.Cvid))
	res, err := d.db.Exec(c, sql, val.Cvid, val.NoteId, val.Content, val.Tag, val.Mid, val.PubVersion, val.ContLen, val.ImgCnt)
	if err != nil {
		return errors.Wrapf(err, "UpArticleContent val(%+v)", val)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		return errors.Wrapf(ecode.NothingFound, "InsertArtContent val(%+v) rowsAffect(%d) err(%+v)", val, rows, e)
	}
	return nil
}

func (d *Dao) InsertArtDetail(c context.Context, val *note.NtPubMsg) error {
	res, err := d.db.Exec(c, _insertArtDetail, val.Cvid, val.NoteId, val.Mid, val.Oid, val.OidType, article.PubStatusPending, xtime.Time(time.Now().Unix()), val.Title, val.Summary, val.PubVersion, val.AutoComment, val.PubFrom)
	if err != nil {
		return errors.Wrapf(err, "InsertArtDetail val(%+v)", val)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		return errors.Wrapf(ecode.NothingFound, "InsertArtDetail val(%+v) rowsAffect(%d) err(%+v)", val, rows, e)
	}
	return nil
}

func (d *Dao) ArtDetail(c context.Context, id int64, tp string, pubStatusEq, pubStatusNotEq int, filterDel bool) (*article.ArtDtlCache, error) {
	var (
		selSql = fmt.Sprintf(_selectArtDetail, tp)
		res    = &article.ArtDtlCache{}
		row    *sql.Row
		args   = []interface{}{id}
	)
	if pubStatusEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusEq)
		args = append(args, pubStatusEq)
	}
	if pubStatusNotEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusNotEq)
		args = append(args, pubStatusNotEq)
	}
	if filterDel {
		selSql = fmt.Sprintf("%s%s", selSql, _filterDel)
	}
	selSql = fmt.Sprintf("%s%s", selSql, _versionOrder)
	row = d.db.QueryRow(c, selSql, args...)
	if err := row.Scan(&res.Cvid, &res.NoteId, &res.Mid, &res.Oid, &res.OidType, &res.Title, &res.Summary, &res.PubStatus, &res.PubReason, &res.Pubtime, &res.Mtime, &res.Deleted, &res.PubVersion); err != nil {
		if err == sql.ErrNoRows {
			return &article.ArtDtlCache{Cvid: -1, Deleted: 1}, nil
		}
		return nil, errors.Wrapf(err, "ArtDetail id(%d) tp(%s)", id, tp)
	}
	return res, nil
}

func (d *Dao) LatestArtCont(c context.Context, cvid int64) (*article.ArtContCache, error) {
	var (
		selSql = fmt.Sprintf(_selectLatestArtCont, contTableName(cvid))
		res    = &article.ArtContCache{}
		row    = d.db.QueryRow(c, selSql, cvid)
	)
	if err := row.Scan(&res.Cvid, &res.Mtime, &res.Deleted, &res.PubVersion); err != nil {
		if err == sql.ErrNoRows {
			return &article.ArtContCache{Cvid: -1}, nil
		}
		return nil, errors.Wrapf(err, "LatestArtCont cvid(%d)", cvid)
	}
	return res, nil
}

func (d *Dao) ArtContByVer(c context.Context, cvid, ver int64) (*article.ArtContCache, error) {
	var (
		selSql = fmt.Sprintf(_selectArtContByVer, contTableName(cvid))
		res    = &article.ArtContCache{}
		row    = d.db.QueryRow(c, selSql, cvid, ver)
	)
	if err := row.Scan(&res.Cvid, &res.NoteId, &res.Mid, &res.Content, &res.Tag, &res.Deleted, &res.Mtime, &res.PubVersion); err != nil {
		if err == sql.ErrNoRows {
			return &article.ArtContCache{Cvid: -1, Deleted: 1}, nil
		}
		return nil, errors.Wrapf(err, "ArtContByVer cvid(%d) ver(%d)", cvid, ver)
	}
	return res, nil
}

func contTableName(id int64) string {
	return fmt.Sprintf("%s_%02d", _artContentTb, id%10)
}
