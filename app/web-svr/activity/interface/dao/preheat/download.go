package preheat

import (
	"context"
	"database/sql"

	"go-gateway/app/web-svr/activity/interface/model/preheat"

	"github.com/pkg/errors"
)

const (
	_downloadSQL = "SELECT id, title, img_url, down_url, schema_url FROM download_middle WHERE id = ?"
)

func (d *Dao) GetByID(c context.Context, ID int64) (res *preheat.DownInfo, err error) {
	res, err = d.CacheGetByID(c, ID)
	if err != nil {
		err = nil
	}
	if res != nil {
		return
	}
	res, err = d.RawGetByID(c, ID)
	if err != nil {
		return
	}
	if res != nil {
		d.AddCacheGetByID(c, ID, res)
	}
	return
}

func (d *Dao) RawGetByID(c context.Context, ID int64) (res *preheat.DownInfo, err error) {
	res = new(preheat.DownInfo)
	row := d.db.QueryRow(c, _downloadSQL, ID)
	if err = row.Scan(&res.ID, &res.Title, &res.ImgURL, &res.DownURL, &res.SchemaURL); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLotteryInfo:QueryRow")
		}
	}
	return
}
