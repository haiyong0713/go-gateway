package dao

import (
	"context"
	"database/sql"

	pb "go-gateway/app/web-svr/space/interface/api/v1"

	"go-common/library/log"
)

const (
	_officialSQL = `SELECT uid,name,icon,scheme,rcmd,ios_url,android_url,button FROM space_official WHERE uid = ? AND deleted = 0`
)

// RawOfficial .
func (d *Dao) RawOfficial(c context.Context, req *pb.OfficialRequest) (res *pb.OfficialReply, err error) {
	res = &pb.OfficialReply{}
	row := d.db.QueryRow(c, _officialSQL, req.Mid)
	if err = row.Scan(&res.Uid, &res.Name, &res.Icon, &res.Scheme, &res.Rcmd, &res.IosUrl, &res.AndroidUrl, &res.Button); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("RawOfficial.Scan mid(%d) error(%v)", req.Mid, err)
		}
	}
	return
}
