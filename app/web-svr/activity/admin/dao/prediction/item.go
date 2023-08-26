package prediction

import (
	"context"
	"fmt"
	"strings"

	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"

	"github.com/pkg/errors"
)

const (
	_batchIstSQL = "INSERT INTO prediction_item(`pid`,`desc`,`image`,`state`,`sid`) VALUES %s"
	_upItemSQL   = "update prediction_item set `state` = ?,`desc` = ? ,`image`=? where `id` = ?"
)

// BatchItem .
func (d *Dao) BatchItem(c context.Context, list []*premdl.PredItem) (err error) {
	rowStr := make([]string, 0, len(list))
	args := make([]interface{}, 0)
	//ctime := time.Now().Format("2006-01-02 15:04:05")
	for _, v := range list {
		rowStr = append(rowStr, "(?,?,?,?,?)")
		args = append(args, v.Pid, v.Desc, v.Image, v.State, v.Sid)
	}
	tx := d.DB
	if err = tx.Model(&premdl.PredItem{}).Exec(fmt.Sprintf(_batchIstSQL, strings.Join(rowStr, ",")), args...).Error; err != nil {
		err = errors.Wrap(err, "premdl.PredItem")
	}
	return
}

// ItemUp pid,sid can not modify.
func (d *Dao) ItemUp(c context.Context, up *premdl.ItemUp) (err error) {
	tx := d.DB
	if err = tx.Model(&premdl.PredItem{}).Exec(_upItemSQL, up.State, up.Desc, up.Image, up.ID).Error; err != nil {
		err = errors.Wrapf(err, "%s", _upItemSQL)
	}
	return
}

// ItemSearch .
func (d *Dao) ItemSearch(c context.Context, arg *premdl.ItemSearch) (res *premdl.ItemSearchRes, err error) {
	var (
		count int64
		list  []*premdl.PredItem
	)
	tx := d.DB
	if arg.ID > 0 {
		tx = tx.Where("id = ?", arg.ID)
	}
	if arg.Pid > 0 {
		tx = tx.Where("pid = ?", arg.Pid)
	}
	if arg.State != -1 {
		tx = tx.Where("state = ?", arg.State)
	}
	pn := (arg.Pn - 1) * arg.Ps
	if err = tx.Model(&premdl.PredItem{}).Count(&count).Error; err != nil {
		err = errors.Wrap(err, "count err")
		return
	}
	if err = tx.Limit(arg.Ps).Offset(pn).Order("id desc").Find(&list).Error; err != nil {
		err = errors.Wrap(err, "find err")
		return
	}
	res = &premdl.ItemSearchRes{
		List: list,
		Pages: premdl.Pages{
			Total: count,
			Num:   arg.Pn,
			Size:  arg.Ps,
		},
	}
	return
}
