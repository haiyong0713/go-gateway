package native

import (
	"context"

	"go-common/library/log"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
)

const (
	_tsPage = "native_ts_page"
)

// ModifyPage .
func (d *Dao) ModifyTsPage(c context.Context, id int64, arg map[string]interface{}) error {
	if err := d.DB.Table(_tsPage).Where("id=?", id).Update(arg).Error; err != nil {
		log.Error("ModifyTsPage d.DB.Table(%d,%v) error(%v)", id, arg, err)
		return err
	}
	return nil
}

// FindTsPage .
func (d *Dao) FindTsPage(c context.Context, id int64) (*natmdl.NatTsPage, error) {
	res := &natmdl.NatTsPage{}
	if err := d.DB.Table(_tsPage).Where("id=?", id).Find(&res).Error; err != nil {
		log.Error("[FindPage] error(%v)", err)
		return nil, err
	}
	return res, nil
}
