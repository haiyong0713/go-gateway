package like

import (
	"context"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _sqlListOnlineDataByVid = "SELECT id, vid, state, name, data, stime, etime FROM act_web_data WHERE vid = ? state = 1 order by ctime desc"

func (d *Dao) RawGetOnlineWebViewDataByVid(c context.Context, vid int64) (list []*like.WebDataItem, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _sqlListOnlineDataByVid, vid)
	if err != nil {
		log.Errorc(c, "RawGetOnlineWebViewDataByVid:d.db.Query(%v) error(%v)", vid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.WebDataItem)
		var data string
		if err = rows.Scan(&n.ID, &n.VID, &n.State, &n.Name, &data, &n.STime, &n.ETime); err != nil {
			log.Errorc(c, "RawGetOnlineWebViewDataByVid:rows.Scan() vid(%d) error(%v)", vid, err)
			return
		}
		n.Raw = []byte(data)
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "RawGetOnlineWebViewDataByVid:rows.Err() vid(%d) error(%v)", vid, err)
		return
	}
	return
}

const _sqlListDataByVid = "SELECT id, vid, state, name, data, stime, etime FROM act_web_data WHERE vid = ? order by ctime desc"

func (d *Dao) RawGetWebViewDataByVid(c context.Context, vid int64) (list []*like.WebDataItem, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _sqlListDataByVid, vid)
	if err != nil {
		log.Errorc(c, "GetWebViewDataByVid:d.db.Query(%v) error(%v)", vid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.WebDataItem)
		var data string
		if err = rows.Scan(&n.ID, &n.VID, &n.State, &n.Name, &data, &n.STime, &n.ETime); err != nil {
			log.Errorc(c, "GetWebViewDataByVid:rows.Scan() vid(%d) error(%v)", vid, err)
			return
		}
		n.Raw = []byte(data)
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "GetWebViewDataByVid:rows.Err() vid(%d) error(%v)", vid, err)
		return
	}
	return
}
