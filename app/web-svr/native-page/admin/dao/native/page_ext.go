package native

import (
	"context"

	"go-common/library/log"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"

	"github.com/jinzhu/gorm"
)

var (
	_pageExtend    = "native_page_ext"
	_upPageExtSQL  = "UPDATE native_page_ext SET `white_value`=? WHERE `pid`=?"
	_addPageExtSQL = "INSERT INTO native_page_ext(`white_value`,`pid`) VALUES (?,?)"
)

func (d *Dao) FindExtByPid(c context.Context, pid int64) (*natmdl.PageExt, error) {
	rly := &natmdl.PageExt{}
	err := d.DB.Table(_pageExtend).Where("pid=?", pid).First(rly).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		return &natmdl.PageExt{Pid: pid}, nil
	}
	return rly, nil
}

func (d *Dao) UpPageExt(c context.Context, natPage *natmdl.PageExt) error {
	//先查询是否有pid对应的数据
	rly := &natmdl.PageExt{}
	err := d.DB.Table(_pageExtend).Where("pid=?", natPage.Pid).First(rly).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		//插入 pid是唯一索引
		if err = d.DB.Exec(_addPageExtSQL, natPage.WhiteValue, natPage.Pid).Error; err != nil {
			log.Error("UpPageExt add %d error(%v)", natPage.ID, err)
			return err
		}
	} else {
		//更新
		if err = d.DB.Exec(_upPageExtSQL, natPage.WhiteValue, natPage.Pid).Error; err != nil {
			log.Error("UpPageExt update %d error(%v)", natPage.ID, err)
			return err
		}
	}
	return nil
}
