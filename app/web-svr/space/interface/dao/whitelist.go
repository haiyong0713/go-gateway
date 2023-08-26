package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"
	xtime "go-common/library/time"

	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_whitelistSQL           = `SELECT mid FROM whitelist WHERE mid = ? AND deleted = 0 AND state=1`
	_whitelistAddSQL        = `INSERT INTO whitelist(mid,mid_name, state,stime,etime,username,deleted) VALUES (?,?,?,?,?,?,?); `
	_whitelistQueryValidSQL = `SELECT * FROM whitelist WHERE mid=? and state != 3 and deleted=0`
	_whitelistUpSQL         = `UPDATE whitelist SET etime=? where mid = ? and state = 1`
	_whitelistQueryTimeSQL  = `SELECT mid, stime, etime FROM whitelist WHERE mid = ? AND deleted = 0 AND state=1`
)

// RawUserTab
func (d *Dao) RawWhitelist(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistReply, err error) {
	res = &pb.WhitelistReply{}
	tmp := struct {
		Mid int64 `json:"mid"`
	}{}
	row := d.db.QueryRow(c, _whitelistSQL, req.Mid)
	if err = row.Scan(&tmp.Mid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res.IsWhite = false
			return
		}
		log.Error("RawWhitelist.Scan mid(%d) error(%v)", req.Mid, err)
		return
	}
	res.IsWhite = true
	return
}

func (d *Dao) ValidWhitelistMid(c context.Context, mid int64) (ok bool, err error) {
	tmp := &model.WhitelistAdd{}
	row := d.db.QueryRow(c, _whitelistQueryValidSQL, mid)
	if err = row.Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			ok = true
			return
		}
		log.Error("ValidWhitelist.Scan mid(%d) error(%v)", mid, err)
		return
	}
	return false, nil
}

func (d *Dao) AddWhiteList(c context.Context, arg *model.WhitelistAdd) (id int64, err error) {
	res, err := d.db.Exec(c, _whitelistAddSQL, arg.Mid, arg.MidName, arg.State, arg.Stime, arg.Etime, arg.Username, 0)
	if err != nil {
		log.Error("space.whitelist.dao.AddWhitelist add(%+v) error(%+v)", arg, err)
		return
	}
	return res.LastInsertId()
}

func (d *Dao) UpWhitelist(c context.Context, arg *pb.WhitelistAddReq) (rows int64, err error) {
	res, err := d.db.Exec(c, _whitelistUpSQL, arg.Etime, arg.Mid)
	if err != nil {
		log.Error("space.whitelist.dao.UpWhitelist Up(%v) error(%v)", arg, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) RawQueryWhitelistValid(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistValidTimeReply, err error) {
	res = &pb.WhitelistValidTimeReply{}
	tmp := struct {
		Mid   int64      `json:"mid"`
		Stime xtime.Time `json:"stime"`
		Etime xtime.Time `json:"etime"`
	}{}
	row := d.db.QueryRow(c, _whitelistQueryTimeSQL, req.Mid)
	if err = row.Scan(&tmp.Mid, &tmp.Stime, &tmp.Etime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res.IsWhite = false
			return
		}
		log.Error("RawWhitelist.Scan mid(%d) error(%v)", req.Mid, err)
		return
	}
	res.IsWhite = true
	res.Stime = tmp.Stime
	res.Etime = tmp.Etime
	return
}
