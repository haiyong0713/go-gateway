package fawkes

import (
	"context"
	"fmt"
	"time"

	"go-gateway/app/app-svr/fawkes/service/model/gitlab"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

const (
	_gitMergeInfoGet             = `SELECT id,merge_id,app_key,path_with_namespace,state,git_action,mr_title,mr_start_time FROM bilibili_fawkes.git_merge WHERE merge_id=?`
	_gitMergeInfoAdd             = `INSERT INTO bilibili_fawkes.git_merge (merge_id,app_key,path_with_namespace,state,git_action,request_user,mr_title,mr_start_time,merged_time) VALUES (?,?,?,?,?,?,?,?,?)`
	_gitMergedTimeUpdate         = `UPDATE bilibili_fawkes.git_merge SET state=?,git_action=?,merged_time=?,mr_start_time=IF((mr_start_time>'1900-01-01 00:00:00'),mr_start_time,?) WHERE merge_id=?`
	_gitMergedStateUpdate        = `UPDATE bilibili_fawkes.git_merge SET state=?,git_action=? WHERE merge_id=?`
	_gitMergeStartTimeUpdate     = `UPDATE bilibili_fawkes.git_merge SET mr_start_time=?,state='merging',git_action='note' WHERE merge_id=? AND state!='merged'`
	_gitMergingLongProcessSelect = `SELECT id,merge_id,app_key,path_with_namespace,state,git_action,request_user,mr_title,mr_start_time,merged_time FROM bilibili_fawkes.git_merge WHERE state='merging' AND mr_start_time<? AND mtime>?`
	_gitMergedLongProcessSelect  = `SELECT id,merge_id,app_key,path_with_namespace,state,git_action,request_user,mr_title,mr_start_time,merged_time FROM bilibili_fawkes.git_merge WHERE state='merged' AND TIMESTAMPDIFF(second, mr_start_time, merged_time)>? AND mtime>?`
)

func (d *Dao) MergeInfoSelect(c context.Context, mergeId int64) (res *gitlab.GitMerge, err error) {
	rows, err := d.db.Query(c, _gitMergeInfoGet, mergeId)
	if err != nil {
		log.Error("d.db.Query(%s, %d) error(%+v)", _gitMergeInfoGet, mergeId, err)
		return
	}
	defer rows.Close()
	var r = &gitlab.GitMerge{}
	var count int
	for rows.Next() {
		count++
		if err = rows.Scan(&r.Id, &r.MergeId, &r.AppKey, &r.PathWithNamespace, &r.State, &r.GitAction, &r.MrTitle, &r.MrStartTime); err != nil {
			log.Error("rows.Scan mergeInfo(%v) error(%+v)", r, err)
			return
		}
		res = r
	}
	if count > 1 {
		err = errors.New(fmt.Sprintf("mergeId: %d get %d merge infos, should be one.", mergeId, count))
	}
	err = rows.Err()
	return
}

func (d *Dao) MergeInfoInsert(c context.Context, mergeId int64, appKey, path, state, action, requestUser, mrTitle string, mrStartTime, mergedTime time.Time) (id int64, err error) {
	rows, err := d.db.Exec(c, _gitMergeInfoAdd, mergeId, appKey, path, state, action, requestUser, mrTitle, mrStartTime, mergedTime)
	if err != nil {
		log.Error("MergeInfoInsert error: %#v", err)
		return
	}
	return rows.LastInsertId()
}

func (d *Dao) MergedStateUpdate(c context.Context, mergeId int64, state, action string) (affected int64, err error) {
	rows, err := d.db.Exec(c, _gitMergedStateUpdate, state, action, mergeId)
	if err != nil {
		log.Error("MergedTimeUpdate error: %#v", err)
		return
	}
	return rows.RowsAffected()
}

func (d *Dao) MergedTimeUpdate(c context.Context, mergeId int64, state, action string, mergedTime time.Time) (affected int64, err error) {
	rows, err := d.db.Exec(c, _gitMergedTimeUpdate, state, action, mergedTime, mergedTime, mergeId)
	if err != nil {
		log.Error("MergedTimeUpdate error: %#v", err)
		return
	}
	return rows.RowsAffected()
}

func (d *Dao) MergeStartTimeUpdate(c context.Context, mergeId int64, mergeStartTime time.Time) (affected int64, err error) {
	rows, err := d.db.Exec(c, _gitMergeStartTimeUpdate, mergeStartTime, mergeId)
	if err != nil {
		log.Error("MergeStartTimeUpdate error: %#v", err)
		return
	}
	return rows.RowsAffected()
}

func (d *Dao) LongMergingProcessSelect(c context.Context, mergeStartTime, mtime time.Time) (data []*gitlab.GitMerge, err error) {
	rows, err := d.db.Query(c, _gitMergingLongProcessSelect, mergeStartTime, mtime)
	if err != nil {
		log.Error("d.db.Query(%s, %v) error(%+v)", _gitMergingLongProcessSelect, mergeStartTime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r = &gitlab.GitMerge{}
		if err = rows.Scan(&r.Id, &r.MergeId, &r.AppKey, &r.PathWithNamespace, &r.State, &r.GitAction, &r.RequestUser, &r.MrTitle, &r.MrStartTime, &r.MergedTime); err != nil {
			log.Error("rows.Scan mergeInfo(%v) error(%+v)", r, err)
			return
		}
		data = append(data, r)
	}
	err = rows.Err()
	return
}

func (d *Dao) LongMergedProcessSelect(c context.Context, duration time.Duration, mtime time.Time) (data []*gitlab.GitMerge, err error) {
	rows, err := d.db.Query(c, _gitMergedLongProcessSelect, duration/time.Second, mtime)
	if err != nil {
		log.Error("d.db.Query(%s, %v) error(%+v)", _gitMergedLongProcessSelect, duration/time.Second, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r = &gitlab.GitMerge{}
		if err = rows.Scan(&r.Id, &r.MergeId, &r.AppKey, &r.PathWithNamespace, &r.State, &r.GitAction, &r.RequestUser, &r.MrTitle, &r.MrStartTime, &r.MergedTime); err != nil {
			log.Error("rows.Scan mergeInfo(%v) error(%+v)", r, err)
			return
		}
		data = append(data, r)
	}
	err = rows.Err()
	return
}
