package dao

import (
	"context"
	"go-common/library/database/sql"
	"go-common/library/log"
)

const (
	_bannedMIdQuery = `SELECT mid FROM up_rcmd_black_list WHERE is_deleted = 0`
)

// RawOfficial .
func (d *Dao) GetBannedRcmdMids(c context.Context) (bannedMids []int64, err error) {
	bannedMids = make([]int64, 0)
	rows, err := d.db.Query(c, _bannedMIdQuery)
	if err != nil {
		if err == sql.ErrNoRows {
			return bannedMids, nil
		}
		log.Error("[banner-rcmd-mid]GetBannedRcmdMids query err: %s", err.Error())
		return bannedMids, err
	}
	defer rows.Close()

	for rows.Next() {
		mid := int64(0)
		if err = rows.Scan(&mid); err != nil {
			log.Error("[banner-rcmd-mid]GetBannedRcmdMids scan err: %s", err.Error())
			return bannedMids, err
		}
		bannedMids = append(bannedMids, mid)
	}
	if err = rows.Err(); err != nil {
		log.Error("[banner-rcmd-mid]GetBannedRcmdMids rows err: %s", err.Error())
	}
	return bannedMids, err
}
