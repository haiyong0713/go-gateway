package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const (
	_webDataCntSQL  = "SELECT COUNT(1) AS cnt FROM act_web_data WHERE state = 1 AND vid = ?"
	_webDataListSQL = "SELECT id,vid,data FROM act_web_data WHERE state= 1 AND vid = ? ORDER BY id LIMIT ?,?"
	_webDataViewSQL = "SELECT id,vid,data,`name`,stime,etime,ctime,mtime,state FROM act_web_data WHERE vid = ? ORDER BY id LIMIT ?,?"

	sql4WebViewData = `
SELECT data
FROM act_web_data
WHERE vid = ?
`
)

func FetchWebViewData(ctx context.Context, vid int64) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	var dataStr string

	err = component.GlobalDB.QueryRow(ctx, sql4WebViewData, vid).Scan(&dataStr)
	if err == nil {
		err = json.Unmarshal([]byte(dataStr), &m)
	}

	return
}

// WebDataCnt get web data count.
func (d *Dao) WebDataCnt(c context.Context, vid int64) (count int, err error) {
	row := d.db.QueryRow(c, _webDataCntSQL, vid)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "WebDataCnt:QueryRow(%d)", vid)
		}
	}
	return
}

// WebDataList get web data list by vid.
func (d *Dao) WebDataList(c context.Context, vid int64, offset, limit int) (list []*like.WebData, err error) {
	rows, err := d.db.Query(c, _webDataListSQL, vid, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "WebDataList:d.db.Query(%d,%d,%d)", vid, offset, limit)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.WebData)
		if err = rows.Scan(&n.ID, &n.Vid, &n.Data); err != nil {
			err = errors.Wrapf(err, "WebDataList:row.Scan row (%d,%d,%d)", vid, offset, limit)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LikeList:rowsErr(%d,%d,%d)", vid, offset, limit)
	}
	return
}

// SourceItem get source data json raw message.
func (d *Dao) SourceItem(c context.Context, sid int64) (source json.RawMessage, err error) {
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.httpClient.RESTfulGet(c, d.sourceItemURL, metadata.String(c, metadata.RemoteIP), url.Values{}, &res, sid); err != nil {
		err = errors.Wrapf(err, "SourceItem d.httpClient.Get sid(%d)", sid)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "d.httpClient.Get sid(%d)", sid)
		return
	}
	source = res.Data
	return
}

// WebDataView get web data list by vid.
func (d *Dao) WebDataView(c context.Context, vid int64, offset, limit int) (list []*like.WebData, err error) {
	rows, err := d.db.Query(c, _webDataViewSQL, vid, offset, limit)
	if err != nil {
		err = errors.Wrapf(err, "WebDataView:d.db.Query(%d,%d,%d)", vid, offset, limit)
		return
	}
	isInfo := d.isInfoSid(vid) // 情报不上线也要显示
	defer rows.Close()
	for rows.Next() {
		n := new(like.WebData)
		if err = rows.Scan(&n.ID, &n.Vid, &n.Data, &n.Name, &n.Stime, &n.Etime, &n.Ctime, &n.Mtime, &n.State); err != nil {
			err = errors.Wrapf(err, "WebDataView:row.Scan row (%d,%d,%d)", vid, offset, limit)
			return
		}
		if isInfo || n.State == 1 {
			list = append(list, n)
		}
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "WebDataView:rowsErr(%d,%d,%d)", vid, offset, limit)
	}
	return
}

func (d *Dao) isInfoSid(sid int64) bool {
	for _, v := range d.c.OperationSource.InfoSids {
		if sid == v {
			return true
		}
	}
	return false
}
