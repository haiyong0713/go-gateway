package dao

import (
	"context"
	"fmt"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/space/interface/model"

	"github.com/pkg/errors"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

const _topPhotoArcSQL = "SELECT aid,image_url,mid,ctime,mtime FROM topphoto_arc_%d WHERE mid=?"

func (d *dao) RawTopPhotoArc(ctx context.Context, mid int64) (*model.TopPhotoArc, error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_topPhotoArcSQL, mid%10), mid)
	data := &model.TopPhotoArc{}
	if err := row.Scan(&data.Aid, &data.ImageUrl, &data.Mid, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawTopPhotoArc")
	}
	return data, nil
}

const _topPhotoArcCancelSQL = `UPDATE topphoto_arc_%d SET aid=0,image_url="" WHERE mid=?`

func (d *dao) TopPhotoArcCancel(ctx context.Context, mid int64) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_topPhotoArcCancelSQL, mid%10), mid)
	if err != nil {
		return 0, errors.Wrap(err, "TopPhotoArcCancel")
	}
	return res.RowsAffected()
}
