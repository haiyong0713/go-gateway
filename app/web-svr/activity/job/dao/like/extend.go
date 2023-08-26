package like

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/like"
	"strconv"
	"strings"
)

var (
	_addExtSQL     = "INSERT INTO `like_extend` (`lid`,`like`) VALUES (?,?) ON DUPLICATE KEY UPDATE `like`=VALUES(`like`) + ?"
	_extendIDSQL   = "SELECT `id`,`lid` FROM `like_extend` WHERE `lid` = ?"
	_modifyLikeSQL = "UPDATE `like_extend` SET `like` = `like` + ? WHERE `lid` = ? "
	_extendIDsSQL  = "SELECT `lid`, `like` FROM `like_extend` WHERE `lid` in (%s)"
)

// RawLikeExtend .
func (d *Dao) RawLikeExtend(c context.Context, lid int64) (list *like.Extend, err error) {
	row := d.db.QueryRow(c, _extendIDSQL, lid)
	list = &like.Extend{}
	if err = row.Scan(&list.ID, &list.Lid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// UpExtend .
func (d *Dao) UpExtend(c context.Context, lid int64, score int64) (err error) {
	if _, err = d.db.Exec(c, _modifyLikeSQL, score, lid); err != nil {
		log.Error("UpExtend:d.db.Exec(%s,%d,%d) error(%v)", _modifyLikeSQL, score, lid, err)
	}
	return
}

// AddExtend .
func (d *Dao) AddExtend(c context.Context, lid, score int64) (err error) {
	if _, err = d.db.Exec(c, _addExtSQL, lid, score, score); err != nil {
		log.Error("AddExtend:d.db.Exec(%s,%d,%d) error(%v)", _addExtSQL, lid, score, err)
	}
	return
}

func (d *Dao) RawLikeExtendByLids(c context.Context, lids []int64) (map[int64]*like.Extend, error) {
	var lIDs []string
	for _, lid := range lids {
		lIDs = append(lIDs, strconv.FormatInt(lid, 10))
	}
	res := make(map[int64]*like.Extend)
	if len(lIDs) == 0 {
		return res, nil
	}

	query := fmt.Sprintf(_extendIDsSQL, strings.Join(lIDs, ","))
	rows, err := d.db.Query(c, query)
	if err != nil {
		log.Errorc(c, "RawLikeExtendByLids query Err sql:%v lids:%v err:%v", _extendIDsSQL, lids, err)
		return res, err
	}

	defer rows.Close()
	for rows.Next() {
		exItem := &like.Extend{}
		if err := rows.Scan(&exItem.Lid, &exItem.Like); err != nil && err != sql.ErrNoRows {
			log.Errorc(c, "RawLikeExtendByLids Each Scan Err sql:%v err:%v", _extendIDsSQL, err)
			return res, err
		}
		res[exItem.Lid] = exItem
	}
	if err := rows.Err(); err != nil {
		log.Errorc(c, "RawLikeExtendByLids Each Scan Err sql:%v err:%v", _extendIDsSQL, err)
		return res, err
	}
	return res, err
}

func (d *Dao) FlushVoteCacheBySid(c context.Context, sid int64, v string) (err error) {
	var (
		key = buildKey("vote", sid)
	)
	if _, err = redis.Bool(component.GlobalRedis.Do(c, "SET", key, v)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(c, fmt.Sprintf("FlushVoteCacheBySid Err key:%v err:%v", key, err))
			return err
		}
	}

	return nil
}
