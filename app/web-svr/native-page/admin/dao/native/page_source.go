package native

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_pageSource        = "native_page_source"
	_updatePgSourceSQL = "UPDATE `native_page_source` SET `sid`=? WHERE `id`=?"
)

func (d *Dao) PageSourcesByPid(c context.Context, pageID int64) (map[int64]*api.NativePageSource, error) {
	var pageSources []*api.NativePageSource
	db := d.DB.Table(_pageSource).Where("page_id=?", pageID)
	if err := db.Find(&pageSources).Error; err != nil {
		log.Errorc(c, "Fail to get native_page_source, page_id=%d error=%+v", pageID, err)
		return nil, err
	}
	rly := make(map[int64]*api.NativePageSource, len(pageSources))
	for _, source := range pageSources {
		rly[source.ActType] = source
	}
	return rly, nil
}

func (d *Dao) UpdatePageSource(c context.Context, id, sid int64) error {
	if err := d.DB.Exec(_updatePgSourceSQL, sid, id).Error; err != nil {
		log.Errorc(c, "Fail to update native_page_source, id=%d sid=%d error=%+v", id, sid, err)
		return err
	}
	return nil
}
