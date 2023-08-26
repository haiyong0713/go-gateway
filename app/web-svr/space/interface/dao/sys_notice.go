package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_sysNoticeSQL    = `SELECT id,content,url,notice_type FROM system_notice WHERE status = 1 AND scopes LIKE '%1%'`
	_sysNoticeUIDSQL = `SELECT system_notice_id,uid FROM system_notice_uid WHERE is_deleted = 2`
)

// SysNoticelist get system notice list from db.
func (d *Dao) SysNoticelist(c context.Context) (sysNotice map[int64]*model.SysNotice, err error) {
	var (
		rows *xsql.Rows
	)
	sysNotice = make(map[int64]*model.SysNotice)
	if rows, err = d.db.Query(c, _sysNoticeSQL); err != nil {
		log.Error("dao.SysNoticelist error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		sn := &model.SysNotice{}
		if err = rows.Scan(&sn.ID, &sn.Content, &sn.Url, &sn.NoticeType); err != nil {
			log.Error("dao.SysNoticelist:row.Scan() error(%v)", err)
			return
		}
		sysNotice[sn.ID] = sn
	}
	if err = rows.Err(); err != nil {
		log.Error("dao SysNoticelist error(%v)", err)
	}
	return
}

// SysNoticeUIDlist get system notice uid list from db.
func (d *Dao) SysNoticeUIDlist(c context.Context) (sysNoticeUID []*model.SysNoticeUid, err error) {
	var (
		rows *xsql.Rows
	)
	sysNoticeUID = make([]*model.SysNoticeUid, 0)
	if rows, err = d.db.Query(c, _sysNoticeUIDSQL); err != nil {
		log.Error("dao.SysNoticeUIDlist error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		uid := &model.SysNoticeUid{}
		if err = rows.Scan(&uid.SystemNoticeId, &uid.Uid); err != nil {
			log.Error("dao.SysNoticeUIDlist:row.Scan() error(%v)", err)
			return
		}
		sysNoticeUID = append(sysNoticeUID, uid)
	}
	if err = rows.Err(); err != nil {
		log.Error("dao.SysNoticeUIDlist error(%v)", err)
	}
	return
}
