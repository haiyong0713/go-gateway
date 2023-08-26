package image

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/xstr"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

const (
	_userImageTb = "user_image"

	_addNoteImage = "INSERT INTO %s(mid,location) VALUES(?,?)"
	_selectImage  = "SELECT id,location FROM %s WHERE id=? AND mid=? AND deleted=0"
	_selectImages = "SELECT id,location FROM %s WHERE mid=? AND id IN (%s)"
)

func (d *Dao) rawImages(c context.Context, imageIds []int64, mid int64) (map[int64]*api.PublishImgInfo, error) {
	selSql := fmt.Sprintf(_selectImages, tableName(_userImageTb, mid), xstr.JoinInts(imageIds))
	rows, err := d.dbr.Query(c, selSql, mid)
	if err != nil {
		err = errors.Wrapf(err, "rawImages imageIds(%v) mid(%d)", imageIds, mid)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*api.PublishImgInfo)
	for rows.Next() {
		tmp := &api.PublishImgInfo{}
		if err = rows.Scan(&tmp.ImageId, &tmp.Location); err != nil {
			err = errors.Wrapf(err, "rawImages imageIds(%v) mid(%d)", imageIds, mid)
			return nil, err
		}
		res[tmp.ImageId] = tmp
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rawImages imageIds(%v) mid(%d)", imageIds, mid)
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawImage(c context.Context, mid int64, id int64) (*note.ImgInfo, error) {
	selSql := fmt.Sprintf(_selectImage, tableName(_userImageTb, mid))
	row := d.dbr.QueryRow(c, selSql, id, mid)
	res := &note.ImgInfo{}
	if err := row.Scan(&res.ImageId, &res.Location); err != nil {
		if err == sql.ErrNoRows {
			return nil, ecode.NothingFound
		}
		return nil, err
	}
	if res.Location == "" {
		return nil, xecode.ImageURLInvalid
	}
	return res, nil
}

func (d *Dao) AddImage(c context.Context, mid int64, location string) (int64, error) {
	sql := fmt.Sprintf(_addNoteImage, tableName(_userImageTb, mid))
	rows, err := d.dbw.Exec(c, sql, mid, location)
	if err != nil {
		return 0, errors.Wrapf(err, "AddImage mid(%d) location(%s)", mid, location)
	}
	id, err := rows.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "AddImage mid(%d) location(%s)", mid, location)
	}
	if id == 0 {
		return 0, xecode.ImageURLInvalid
	}
	return id, nil
}

func tableName(table string, id int64) string {
	return fmt.Sprintf("%s_%02d", table, id%50)
}
