package archive

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/archive/job/model/archive"
)

const (
	_ffSQL  = "SELECT cid,first_frame FROM video_extra WHERE cid=?"
	_ffsSQL = "SELECT cid,first_frame FROM video_extra WHERE cid in (%s)"
)

// RawVideoFistFrame get video first frame
func (d *Dao) RawVideoFistFrame(c context.Context, cid int64) (*archive.VideoFF, error) {
	row := d.tidb.QueryRow(c, _ffSQL, cid)
	ff := &archive.VideoFF{}
	if err := row.Scan(&ff.Cid, &ff.FirstFrame); err != nil {
		return nil, err
	}
	return ff, nil
}

// RawVideoFistFrame get video first frame
func (d *Dao) RawVideoFistFrames(c context.Context, cids []int64) (map[int64]*archive.VideoFF, error) {
	rows, err := d.tidb.Query(c, fmt.Sprintf(_ffsSQL, xstr.JoinInts(cids)))
	if err != nil {
		log.Error("d.db.Query(%s, %+v) error(%+v)", _ffsSQL, cids, err)
		return nil, err
	}
	defer rows.Close()
	ff := make(map[int64]*archive.VideoFF)
	for rows.Next() {
		tmpff := &archive.VideoFF{}
		if err = rows.Scan(&tmpff.Cid, &tmpff.FirstFrame); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return nil, err
		}
		ff[tmpff.Cid] = tmpff
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err(%+v)", err)
		return nil, err
	}
	return ff, nil
}
