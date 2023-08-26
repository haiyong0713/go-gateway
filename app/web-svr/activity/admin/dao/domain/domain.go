package domain

import (
	"context"
	"database/sql"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	mdomain "go-gateway/app/web-svr/activity/admin/model/domain"
	xecode "go-gateway/app/web-svr/activity/ecode"
	"strings"
	xtime "time"
)

const (
	addDomainRecord    = "INSERT INTO act_domain_conf(act_name, page_link,first_domain,second_domain,stime,etime) VALUES(?,?,?,?,?,?)"
	updateDomainRecord = "UPDATE act_domain_conf SET act_name=?,page_link=?,first_domain=?,second_domain=?,state=?, stime=?,etime=? WHERE id=?"
	updateDomainState  = "UPDATE act_domain_conf SET state=? WHERE id=?"
	updateDomainTime   = "UPDATE act_domain_conf SET state=?, etime=? WHERE id=?"
	searchDomainList   = "SELECT id ,act_name, page_link,first_domain,second_domain,stime,etime,ctime,mtime  FROM act_domain_conf %s  ORDER BY id DESC  LIMIT ? OFFSET ?"
	totalCount         = "SELECT count(*)  FROM act_domain_conf %s  "
	syncFailList       = "SELECT id ,act_name, page_link,first_domain,second_domain,stime,etime,ctime,mtime  FROM act_domain_conf WHERE state = 0 ORDER BY id ASC LIMIT ?"
)

// AddRecord ...
func (d *Dao) AddRecord(ctx context.Context, param *mdomain.AddDomainParam) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(ctx, addDomainRecord, param.ActName, param.PageLink, param.FirstDomain,
		param.SecondDomain, param.Stime, param.Etime); err != nil {
		log.Errorc(ctx, "domain@Add d.db.Exec() INSERT failed. error(%v)", err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = xecode.ActivityDomainConflictError
		}
		return
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Errorc(ctx, "domain@Add result.LastInsertId() failed. error(%v)", err)
	}
	return
}

// UpdateRecord ...
func (d *Dao) UpdateRecord(ctx context.Context, param *mdomain.Record, state int) (rows int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(ctx, updateDomainRecord, param.ActName, param.PageLink, param.FirstDomain,
		param.SecondDomain, state, param.Stime, param.Etime, param.Id); err != nil {
		log.Errorc(ctx, "domain@UpdateRecord() Update act_domain_conf failed. error(%v)", err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = xecode.ActivityDomainConflictError
		}
		return
	}
	return result.RowsAffected()
}

// UpdateStatus ...
func (d *Dao) UpdateStatus(ctx context.Context, state int, id int64) (rows int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(ctx, updateDomainState, state, id); err != nil {
		log.Errorc(ctx, "domain@UpdateStatus() Update act_domain_conf failed. error(%v)", err)
		return
	}
	return result.RowsAffected()
}

func (d *Dao) UpdateEtime(ctx context.Context, state int, id int64) (rows int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(ctx, updateDomainTime, state, xtime.Now(), id); err != nil {
		log.Errorc(ctx, "domain@UpdateStatus() Update act_domain_conf failed. error(%v)", err)
		return
	}
	return result.RowsAffected()
}

func (d *Dao) Search(ctx context.Context, id int64, name string, limit int, offset int) (records []*mdomain.Record, total int, err error) {

	var (
		sqlAdd string
		args   []interface{}
	)
	if id != 0 || name != "" {
		sqlAdd = "WHERE "
		flag := false
		if id != 0 {
			args = append(args, id)
			sqlAdd += "id=? "
			flag = true
		}
		if name != "" {
			args = append(args, "%"+name+"%")
			if flag {
				sqlAdd += "AND "
			}
			sqlAdd += " act_name LIKE ?"
		}
	}

	var rows *xsql.Rows
	result := d.db.QueryRow(ctx, fmt.Sprintf(totalCount, sqlAdd), args...)
	if err = result.Scan(&total); err != nil || total <= 0 {
		if err != nil {
			log.Errorc(ctx, "domain@Search result.Scan() failed. error(%v) ", err)
		}
		return
	}

	args = append(args, limit, offset)
	if rows, err = d.db.Query(ctx, fmt.Sprintf(searchDomainList, sqlAdd), args...); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "domain@Search d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &mdomain.Record{}
		if err = rows.Scan(&tmp.Id, &tmp.ActName, &tmp.PageLink, &tmp.FirstDomain, &tmp.SecondDomain,
			&tmp.Stime, &tmp.Etime, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(ctx, "domain@Search rows.Scan failed. error(%v)", err)
			return
		}
		records = append(records, tmp)
	}
	err = rows.Err()
	return
}

func (d *Dao) SyncFailedList(ctx context.Context, limit int) (records []*mdomain.Record, err error) {

	var rows *xsql.Rows
	if rows, err = d.db.Query(ctx, syncFailList, limit); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "domain@SyncFailedList d.db.Query() SELECT failed. error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &mdomain.Record{}
		if err = rows.Scan(&tmp.Id, &tmp.ActName, &tmp.PageLink, &tmp.FirstDomain, &tmp.SecondDomain,
			&tmp.Stime, &tmp.Etime, &tmp.Ctime, &tmp.Mtime); err != nil {
			log.Errorc(ctx, "domain@Search rows.Scan failed. error(%v)", err)
			return
		}
		records = append(records, tmp)
	}
	err = rows.Err()
	return
}
