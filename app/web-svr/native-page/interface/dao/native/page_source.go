package native

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_addPgSourceSQL    = "INSERT INTO `native_page_source` (`page_id`,`sid`,`partitions`,`act_type`) VALUES (?,?,?,?)"
	_pgSourcesByPidSQL = "SELECT `id`,`page_id`,`sid`,`partitions`,`act_type` FROM `native_page_source` WHERE `page_id`=?"
	_updatePgSourceSQL = "UPDATE `native_page_source` SET `partitions`=? WHERE `id`=?"
)

func (d *Dao) AddPageSource(c context.Context, source *api.NativePageSource) (int64, error) {
	rly, err := d.db.Exec(c, _addPgSourceSQL, source.PageId, source.Sid, source.Partitions, source.ActType)
	if err != nil {
		log.Errorc(c, "Fail to create native_page_source, data=%+v error=%+v", source, err)
		return 0, err
	}
	return rly.LastInsertId()
}

func (d *Dao) UpdatePageSource(c context.Context, id int64, partitions string) error {
	if _, err := d.db.Exec(c, _updatePgSourceSQL, partitions, id); err != nil {
		log.Errorc(c, "Fail to update native_page_source, id=%d partitions=%s error=%+v", id, partitions, err)
		return err
	}
	return nil
}

func (d *Dao) PageSourcesByPid(c context.Context, pageID int64) (map[int64]*api.NativePageSource, error) {
	rows, err := d.db.Query(c, _pgSourcesByPidSQL, pageID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return map[int64]*api.NativePageSource{}, nil
		}
		log.Errorc(c, "Fail to query _pgSourcesByPidSQL, pageID=%d error=%+v", pageID, err)
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64]*api.NativePageSource)
	for rows.Next() {
		m := &api.NativePageSource{}
		if err = rows.Scan(&m.Id, &m.PageId, &m.Sid, &m.Partitions, &m.ActType); err != nil {
			log.Errorc(c, "Fail to scan NativePageSource row, pageID=%d error=%+v", pageID, err)
			continue
		}
		list[m.ActType] = m
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get NativePageSource rows, pageID=%d error=%+v", pageID, err)
		return nil, err
	}
	return list, nil
}
