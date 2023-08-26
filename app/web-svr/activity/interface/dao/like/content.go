package like

import (
	"context"
	"fmt"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	l "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_contentSQL      = "select id,message,ip,plat,device,ctime,mtime,image,reply,link,ex_name from like_content where id in (%s)"
	_addContSQL      = "insert into %s (`id`,`message`,`ipv6`,`plat`,`device`,`ctime`,`mtime`,`image`,`reply`,`link`,`ex_name`) values(?,?,?,?,?,?,?,?,?,?,?)"
	_oldContTable    = "like_content"
	_oldContNewTable = "like_content_new"
)

// RawLikeContent .
func (dao *Dao) RawLikeContent(c context.Context, ids []int64) (res map[int64]*l.LikeContent, err error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_contentSQL, xstr.JoinInts(ids)))
	if err != nil {
		err = errors.Wrap(err, "dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make(map[int64]*l.LikeContent, len(ids))
	for rows.Next() {
		t := &l.LikeContent{}
		if err = rows.Scan(&t.ID, &t.Message, &t.IP, &t.Plat, &t.Device, &t.Ctime, &t.Mtime, &t.Image, &t.Reply, &t.Link, &t.ExName); err != nil {
			err = errors.Wrapf(err, "rows.Scan()")
			return
		}
		res[t.ID] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, " rows.Err()")
	}
	return
}

// TxAddContent .
func (dao *Dao) TxAddContent(c context.Context, tx *xsql.Tx, cont *l.LikeContent) (err error) {
	var (
		now = time.Now().Format("2006-01-02 15:04:05")
	)
	if _, err = tx.Exec(fmt.Sprintf(_addContSQL, _oldContTable), cont.ID, cont.Message, cont.IPv6, cont.Plat, cont.Device, now, now, cont.Image, cont.Reply, cont.Link, cont.ExName); err != nil {
		log.Error("TxAddContent:tx.Exec(%s) error(%v)", _oldTableName, err)
	}
	return
}

// TxAddContentNew .
func (dao *Dao) TxAddContentNew(c context.Context, tx *xsql.Tx, cont *l.LikeContent) (err error) {
	var (
		now = time.Now().Format("2006-01-02 15:04:05")
	)
	if _, err = tx.Exec(fmt.Sprintf(_addContSQL, _oldContNewTable), cont.ID, cont.Message, cont.IPv6, cont.Plat, cont.Device, now, now, cont.Image, cont.Reply, cont.Link, cont.ExName); err != nil {
		log.Errorc(c, "ItemAndContentNew TxAddContentNew:tx.Exec(%s) error(%v)", _oldContNewTable, err)
	}
	return
}
