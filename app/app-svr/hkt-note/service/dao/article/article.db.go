package article

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/hkt-note/service/model/article"

	"github.com/pkg/errors"
)

const (
	_artContentTb = "article_content"

	_selectCvidInArc           = "SELECT DISTINCT cvid FROM article_detail WHERE oid=? AND oid_type=? AND pub_status=? AND deleted=0"
	_selectArtDetail           = "SELECT cvid,note_id,mid,oid,oid_type,title,summary,pub_status,pub_reason,pubtime,mtime,deleted,pub_version FROM article_detail WHERE %s=? AND deleted=0"
	_selectArtDetails          = "SELECT cvid,note_id,mid,oid,oid_type,title,summary,pub_status,pub_reason,pubtime,mtime,deleted,pub_version FROM article_detail WHERE %s IN (%s) AND deleted=0"
	_selectArtStatus           = "SELECT cvid,pub_status FROM article_detail WHERE cvid IN (%s) AND deleted=0 ORDER BY cvid,pub_version DESC"
	_pubStatusEq               = " AND pub_status=?"
	_pubStatusNotEq            = " AND pub_status<>?"
	_versionOrder              = " ORDER BY pub_version DESC LIMIT 1"
	_versionsOrder             = " ORDER BY %s,pub_version DESC"
	_selectArtContent          = "SELECT cvid,note_id,tag,content,deleted,pub_version,cont_len,img_cnt FROM %s WHERE cvid=? AND pub_version=? AND deleted=0"
	_selectArtListInArc        = "SELECT cvid,note_id,pubtime,mtime FROM article_detail WHERE oid=? AND oid_type=? AND pub_status=? AND deleted=0 ORDER BY pubtime DESC"
	_selectArtListInUser       = "SELECT cvid,note_id,pubtime,mtime FROM article_detail WHERE mid=? AND pub_status=? AND deleted=0 ORDER BY pubtime DESC"
	_selectPubSuccessArtDetail = "SELECT cvid,pub_status,pub_version FROM article_detail WHERE cvid = ? AND pub_status = ? AND deleted=0"
)

func (d *Dao) rawArtCountInArc(c context.Context, oid, oidType int64) (int64, error) {
	//oid下的cvid，只要有一条pubpass就算
	rows, err := d.dbr.Query(c, _selectCvidInArc, oid, oidType, article.PubStatusPassed)
	if err != nil {
		return 0, errors.Wrapf(err, "rawArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	var cvids []int64
	defer rows.Close()
	for rows.Next() {
		var cvid int64
		if err = rows.Scan(&cvid); err != nil {
			return 0, errors.Wrapf(err, "rawArtCountInArc oid(%d) oidType(%d)", oid, oidType)
		}
		cvids = append(cvids, cvid)
	}
	if err = rows.Err(); err != nil {
		return 0, errors.Wrapf(err, "rawArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	if len(cvids) == 0 {
		return 0, nil
	}
	// 过滤被锁定但还未被删除的专栏
	var filtered []int64
	if filtered, err = d.filterLockArts(c, cvids); err != nil {
		return 0, errors.Wrapf(err, "rawArtCountInArc oid(%d) oidType(%d)", oid, oidType)
	}
	return int64(len(filtered)), nil
}

func (d *Dao) filterLockArts(c context.Context, ids []int64) ([]int64, error) {
	selSql := fmt.Sprintf(_selectArtStatus, xstr.JoinInts(ids))
	rows, err := d.dbr.Query(c, selSql)
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

func (d *Dao) rawArtDetails(c context.Context, ids []int64, tp string, pubStatusEq, pubStatusNotEq int) (map[int64]*article.ArtDtlCache, error) {
	var err error
	// 过滤锁定且未删除的
	if tp == article.TpArtDetailCvid {
		if ids, err = d.filterLockArts(c, ids); err != nil {
			return nil, err
		}
	}
	if len(ids) == 0 {
		return make(map[int64]*article.ArtDtlCache), nil
	}
	var (
		selSql = fmt.Sprintf(_selectArtDetails, tp, xstr.JoinInts(ids))
		verSql = fmt.Sprintf(_versionsOrder, tp)
		rows   *sql.Rows
		args   = make([]interface{}, 0)
	)
	if pubStatusEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusEq)
		args = append(args, pubStatusEq)
	}
	if pubStatusNotEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusNotEq)
		args = append(args, pubStatusNotEq)
	}
	selSql = fmt.Sprintf("%s%s", selSql, verSql)
	rows, err = d.dbr.Query(c, selSql, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "rawArtDetails ids(%v) tp(%s)", ids, tp)
	}
	defer rows.Close()
	res := make(map[int64]*article.ArtDtlCache)
	for rows.Next() {
		tmp := &article.ArtDtlCache{}
		if err = rows.Scan(&tmp.Cvid, &tmp.NoteId, &tmp.Mid, &tmp.Oid, &tmp.OidType, &tmp.Title, &tmp.Summary, &tmp.PubStatus, &tmp.PubReason, &tmp.Pubtime, &tmp.Mtime, &tmp.Deleted, &tmp.PubVersion); err != nil {
			return nil, errors.Wrapf(err, "rawArtDetails ids(%v) tp(%s)", ids, tp)
		}
		switch tp {
		case article.TpArtDetailCvid:
			if _, ok := res[tmp.Cvid]; !ok {
				res[tmp.Cvid] = tmp
			}
		case article.TpArtDetailNoteId:
			if _, ok := res[tmp.NoteId]; !ok {
				res[tmp.NoteId] = tmp
			}
		default:
		}
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rawArtDetails ids(%v) tp(%s)", ids, tp)
	}
	return res, nil
}

func (d *Dao) rawArtListInUser(c context.Context, min, max, mid int64) ([]*article.ArtList, []string, error) {
	// 过滤锁定且未删除的客态笔记
	lockRows, err := d.dbr.Query(c, _selectArtListInUser, mid, article.PubStatusLock)
	if err != nil {
		return nil, nil, err
	}
	locked := make(map[int64]struct{})
	defer lockRows.Close()
	for lockRows.Next() {
		r := &article.ArtList{}
		if err = lockRows.Scan(&r.Cvid, &r.NoteId, &r.Pubtime, &r.Mtime); err != nil {
			return nil, nil, errors.Wrapf(err, "rawArtListInUser mid(%d)", mid)
		}
		locked[r.Cvid] = struct{}{}
	}
	if err = lockRows.Err(); err != nil {
		return nil, nil, errors.Wrapf(err, "rawArtListInUser mid(%d)", mid)
	}
	rows, err := d.dbr.Query(c, _selectArtListInUser, mid, article.PubStatusPassed)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var (
		res    []*article.ArtList
		idsArr []string
		exist  = make(map[int64]struct{})
	)
	for rows.Next() {
		r := &article.ArtList{}
		if err = rows.Scan(&r.Cvid, &r.NoteId, &r.Pubtime, &r.Mtime); err != nil {
			return nil, nil, errors.Wrapf(err, "rawArtListInUser mid(%d)", mid)
		}
		if _, ok := exist[r.NoteId]; ok {
			continue
		}
		if _, ok := locked[r.Cvid]; ok {
			continue
		}
		exist[r.NoteId] = struct{}{}
		res = append(res, r)
		idsArr = append(idsArr, article.ToArtListVal(r.Cvid, r.NoteId))
	}
	if err = rows.Err(); err != nil {
		return nil, nil, errors.Wrapf(err, "rawArtListInUser mid(%d)", mid)
	}
	// 根据min,max截断cvids
	cutidsArr := func() []string {
		if min == 0 && max == -1 {
			return idsArr
		}
		if int(min) >= len(idsArr) {
			return nil
		}
		if int(max) >= len(idsArr) {
			return idsArr[min:]
		}
		return idsArr[min : max+1]
	}()
	return res, cutidsArr, nil
}

func (d *Dao) rawArtListInArc(c context.Context, min, max, oid, oidType int64) ([]*article.ArtList, []string, error) {
	// 过滤锁定且未删除的客态笔记
	lockRows, err := d.dbr.Query(c, _selectArtListInArc, oid, oidType, article.PubStatusLock)
	if err != nil {
		return nil, nil, err
	}
	locked := make(map[int64]struct{})
	defer lockRows.Close()
	for lockRows.Next() {
		r := &article.ArtList{}
		if err = lockRows.Scan(&r.Cvid, &r.NoteId, &r.Pubtime, &r.Mtime); err != nil {
			return nil, nil, errors.Wrapf(err, "rawArtListInArc oid(%d) oidType(%d)", oid, oidType)
		}
		locked[r.Cvid] = struct{}{}
	}
	if err = lockRows.Err(); err != nil {
		return nil, nil, errors.Wrapf(err, "rawArtListInArc oid(%d) oidType(%d)", oid, oidType)
	}
	//稿件oid下状态为公开浏览且未删除的笔记（同一个cvid可能有多条）
	rows, err := d.dbr.Query(c, _selectArtListInArc, oid, oidType, article.PubStatusPassed)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var (
		res    = make([]*article.ArtList, 0)
		idsArr = make([]string, 0)
		exist  = make(map[int64]struct{})
	)
	for rows.Next() {
		r := &article.ArtList{}
		if err = rows.Scan(&r.Cvid, &r.NoteId, &r.Pubtime, &r.Mtime); err != nil {
			return nil, nil, errors.Wrapf(err, "rawArtListInArc oid(%d) oidType(%d)", oid, oidType)
		}
		if _, ok := exist[r.NoteId]; ok {
			continue
		}
		exist[r.NoteId] = struct{}{}
		res = append(res, r)
		idsArr = append(idsArr, article.ToArtListVal(r.Cvid, r.NoteId))
	}
	if err = rows.Err(); err != nil {
		return nil, nil, errors.Wrapf(err, "rawArtListInArc oid(%d) oidType(%d)", oid, oidType)
	}
	// 根据min,max截断cvids
	cutIdsArr := func() []string {
		if min == 0 && max == -1 {
			return idsArr
		}
		if int(min) >= len(idsArr) {
			return nil
		}
		if int(max) >= len(idsArr) {
			return idsArr[min:]
		}
		return idsArr[min : max+1]
	}()
	return res, cutIdsArr, nil
}

func (d *Dao) RawArtDetail(c context.Context, id int64, tp string, pubStatusEq, pubStatusNotEq int) (*article.ArtDtlCache, error) {
	selSql := fmt.Sprintf(_selectArtDetail, tp)
	// 客态场景，先判断cvid最后一个version是否为锁定，锁定直接返回空
	if tp == article.TpArtDetailCvid {
		var (
			latestSql = fmt.Sprintf("%s%s", selSql, _versionOrder)
			latestRow = d.dbr.QueryRow(c, latestSql, id)
			latest    = &article.ArtDtlCache{}
		)
		if err := latestRow.Scan(&latest.Cvid, &latest.NoteId, &latest.Mid, &latest.Oid, &latest.OidType, &latest.Title, &latest.Summary, &latest.PubStatus, &latest.PubReason, &latest.Pubtime, &latest.Mtime, &latest.Deleted, &latest.PubVersion); err != nil {
			return nil, err
		}
		if latest.PubStatus == article.PubStatusLock {
			return &article.ArtDtlCache{Cvid: -1, PubStatus: article.PubStatusLock}, nil
		}
	}
	var (
		res  = &article.ArtDtlCache{}
		row  *sql.Row
		args = []interface{}{id}
	)
	if pubStatusEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusEq)
		args = append(args, pubStatusEq)
	}
	if pubStatusNotEq > 0 {
		selSql = fmt.Sprintf("%s%s", selSql, _pubStatusNotEq)
		args = append(args, pubStatusNotEq)
	}
	selSql = fmt.Sprintf("%s%s", selSql, _versionOrder)
	row = d.dbr.QueryRow(c, selSql, args...)
	if err := row.Scan(&res.Cvid, &res.NoteId, &res.Mid, &res.Oid, &res.OidType, &res.Title, &res.Summary, &res.PubStatus, &res.PubReason, &res.Pubtime, &res.Mtime, &res.Deleted, &res.PubVersion); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawArtContent(c context.Context, cvid, ver int64) (*article.ArtContCache, error) {
	res := &article.ArtContCache{}
	selSql := fmt.Sprintf(_selectArtContent, artTableName(_artContentTb, cvid))
	row := d.dbr.QueryRow(c, selSql, cvid, ver)
	if err := row.Scan(&res.Cvid, &res.NoteId, &res.Tag, &res.Content, &res.Deleted, &res.PubVersion, &res.ContLen, &res.ImgCnt); err != nil {
		return nil, err
	}
	return res, nil
}

//是否在指定pub_version前成功发布过
func (d *Dao) GetPubSuccessCvidsBeforeAssignedVersion(ctx context.Context, cvid int64, pubVersion int64) (bool, error) {
	rows, err := d.dbr.Query(ctx, _selectPubSuccessArtDetail, cvid, article.PubStatusPassed)
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

func artTableName(table string, id int64) string {
	return fmt.Sprintf("%s_%02d", table, id%10)
}
