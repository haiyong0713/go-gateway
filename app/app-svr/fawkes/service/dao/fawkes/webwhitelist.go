package fawkes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	"go-gateway/app/app-svr/fawkes/service/api/app/webcontainer"
	webmdl "go-gateway/app/app-svr/fawkes/service/model/webcontainer"
)

//go:generate ../sqlgenerate/gensql -filter _getAllWhiteList

const (
	_addWhiteList               = "INSERT INTO web_white_list (app_key,title,domain,reason,is_third_party,feature,effective,expires,comet_id,is_domain_active) VALUES (?,?,?,?,?,?,?,?,?,?)"
	_delWhiteList               = "DELETE FROM web_white_list WHERE id=?"
	_updateWhiteList            = "UPDATE web_white_list SET %s WHERE id=?"
	_getWhiteList               = "SELECT id,app_key,title,domain,reason,is_third_party,feature,effective,expires,comet_id,is_domain_active,mtime,ctime FROM web_white_list WHERE %s"
	_getWhiteListByActiveStatus = "SELECT id,app_key,title,domain,reason,is_third_party,feature,effective,expires,comet_id,is_domain_active,mtime,ctime FROM web_white_list WHERE FIND_IN_SET(?,app_key) AND is_domain_active=?"
	_getAllWhiteList            = "SELECT id,app_key,title,domain,reason,is_third_party,feature,effective,expires,comet_id,is_domain_active,mtime,ctime FROM web_white_list"                      // []*webmdl.WebWhiteList
	_whitelistByDomain          = "SELECT id,app_key,title,domain,reason,is_third_party,feature,effective,expires,comet_id,is_domain_active,mtime,ctime FROM web_white_list WHERE domain IN (%s)" //webmdl.WebWhiteList
)

func (d *Dao) AddWhiteList(ctx context.Context, data *webcontainer.AddWhiteListReq) (id int64, err error) {
	s, err := json.Marshal(data.Feature)
	if err != nil {
		return
	}
	row, err := d.db.Exec(ctx, _addWhiteList, data.AppKey, data.Title, data.Domain, data.Reason, data.IsThirdParty.Value, strings.Trim(string(s), "[]"), time.Unix(data.Effective, 0), time.Unix(data.Expires, 0), data.CometId, true)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) BatchAddWhiteList(ctx context.Context, data []*webcontainer.AddWhiteListReq) (err error) {
	for _, v := range data {
		if _, err = d.AddWhiteList(ctx, v); err != nil {
			return
		}
	}
	return
}

func (d *Dao) DelWhiteList(ctx context.Context, id int64) (rowsAffected int64, err error) {
	result, err := d.db.Exec(ctx, _delWhiteList, id)
	if err != nil {
		return
	}
	return result.RowsAffected()
}

func (d *Dao) UpdateWhiteList(ctx context.Context, update *webcontainer.UpdateWhiteListReq) (rowsAffected int64, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if update.Title != "" {
		args = append(args, update.Title)
		sqls = append(sqls, "title=?")
	}
	if update.Reason != "" {
		args = append(args, update.Reason)
		sqls = append(sqls, "reason=?")
	}
	if update.IsThirdParty != nil {
		args = append(args, update.IsThirdParty.Value)
		sqls = append(sqls, "is_third_party=?")
	}
	if update.IsDomainActive != nil {
		args = append(args, update.IsDomainActive.Value)
		sqls = append(sqls, "is_domain_active=?")
	}
	if update.CometId != "" {
		args = append(args, update.CometId)
		sqls = append(sqls, "comet_id=?")
	}
	if update.Expires != nil {
		args = append(args, time.Unix(update.Expires.Value, 0))
		sqls = append(sqls, "expires=?")
	}
	if update.Effective != nil {
		args = append(args, time.Unix(update.Effective.Value, 0))
		sqls = append(sqls, "effective=?")
	}
	if len(update.Feature) != 0 {
		var b []byte
		b, err = json.Marshal(update.Feature)
		if err != nil {
			return
		}
		args = append(args, strings.Trim(string(b), "[]"))
		sqls = append(sqls, "feature=?")
	}
	args = append(args, update.Id)
	rows, err := d.db.Exec(ctx, fmt.Sprintf(_updateWhiteList, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	return rows.RowsAffected()
}

func (d *Dao) BatchUpdateWhiteList(ctx context.Context, update []*webcontainer.UpdateWhiteListReq) (err error) {
	for _, v := range update {
		_, err = d.UpdateWhiteList(ctx, v)
		if err != nil {
			return
		}
	}
	return
}

func (d *Dao) SelectWhiteList(ctx context.Context, query *webcontainer.GetWhiteListReq) (items []*webmdl.WebWhiteList, err error) {
	var (
		args []interface{}
		sqls []string
		rows *sql.Rows
	)
	if query.AppKey != "" {
		args = append(args, query.AppKey)
		sqls = append(sqls, "FIND_IN_SET(?,app_key)")
	}
	if query.Domain != "" {
		args = append(args, query.Domain)
		sqls = append(sqls, "domain=?")
	}
	if query.IsThirdParty != nil {
		args = append(args, query.IsThirdParty.Value)
		sqls = append(sqls, "is_third_party=?")
	}
	if query.IsDomainActive != nil {
		args = append(args, query.IsDomainActive.Value)
		sqls = append(sqls, "is_domain_active=?")
	}
	if query.Effective != nil {
		args = append(args, query.Effective.Value)
		sqls = append(sqls, "effective=?")
	}
	if query.Expires != nil {
		args = append(args, query.Expires.Value)
		sqls = append(sqls, "expires=?")
	}
	if query.CometId != "" {
		args = append(args, query.CometId)
		sqls = append(sqls, "comet_id=?")
	}
	if len(query.Feature) != 0 {
		var feature []string
		for _, f := range query.Feature {
			feature = append(feature, fmt.Sprintf("FIND_IN_SET('%s',feature)", f))
		}
		sqls = append(sqls, strings.Join(feature, " AND"))
	}
	if len(sqls) == 0 {
		args = append(args, 1)
		sqls = append(sqls, "1=?")
	}
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_getWhiteList, strings.Join(sqls, " AND ")), args...); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	return
}

func (d *Dao) SelectWhitelistByDomain(ctx context.Context, domains []string) (item []*webmdl.WebWhiteList, err error) {
	var (
		args []interface{}
		sqls []string
		rows *sql.Rows
	)
	if len(domains) == 0 {
		return
	}
	for _, item := range domains {
		sqls = append(sqls, "?")
		args = append(args, item)
	}
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_whitelistByDomain, strings.Join(sqls, ",")), args...); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &item); err != nil {
		return
	}
	return
}

func (d *Dao) SelectWhitelistMapByDomain(ctx context.Context, domains []string) (m map[string]*webmdl.WebWhiteList, err error) {
	list, err := d.SelectWhitelistByDomain(ctx, domains)
	if err != nil {
		return
	}
	wm := make(map[string]*webmdl.WebWhiteList)
	for _, v := range list {
		wm[v.Domain] = v
	}
	return wm, nil
}

func (d *Dao) SelectWhiteListByActiveStatus(ctx context.Context, appKey string, idDomainActive bool) (item []*webmdl.WebWhiteList, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _getWhiteListByActiveStatus, appKey, idDomainActive); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &item); err != nil {
		return
	}
	return
}

func (d *Dao) SelectAllWhiteList(ctx context.Context) (items []*webmdl.WebWhiteList, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _getAllWhiteList); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	return
}
