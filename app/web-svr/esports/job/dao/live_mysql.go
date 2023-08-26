package dao

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	mdlesp "go-gateway/app/web-svr/esports/job/model"

	"github.com/pkg/errors"
)

const (
	_addOffLineImageSQL = "REPLACE INTO `es_live_offline_image` (item_type,item_id,item_name,nick_name,score_image,bfs_image) VALUES  %s"
	_addBattleListSQL   = "REPLACE INTO `es_live_battle_list` (match_id,battle_list) VALUES (?,?)"
	_addBattleInfoSQL   = "REPLACE INTO `es_live_battle_info` (live_type,battle_string,battle_info) VALUES (?,?,?)"
	_selOffLineImageSQL = "SELECT distinct score_image FROM es_live_offline_image"
)

// AddOffLineImage .
func (d *Dao) AddOffLineImage(ctx context.Context, data []*mdlesp.OffLineImage) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\")",
			v.ItemType, v.ItemId, v.ItemName, v.NickName, v.ScoreImage, v.BfsImage))
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_addOffLineImageSQL, strings.Join(rowStrings, ","))); err != nil {
		log.Errorc(ctx, "AddOffLineImage db.Exec error(%+v)", err)
	}
	return
}

func (d *Dao) AddBattleList(ctx context.Context, matchID, data string) (err error) {
	if _, err = d.db.Exec(ctx, _addBattleListSQL, matchID, data); err != nil {
		log.Errorc(ctx, "AddBattleList db.Exec error(%+v)", err)
	}
	return
}

func (d *Dao) AddBattleInfo(ctx context.Context, tp int, battleString, data string) (err error) {
	if _, err = d.db.Exec(ctx, _addBattleInfoSQL, tp, battleString, data); err != nil {
		log.Errorc(ctx, "AddBattleInfo db.Exec error(%+v)", err)
	}
	return
}

func (d *Dao) OffLineImage(c context.Context) (rs map[string]bool, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _selOffLineImageSQL)
	if err != nil {
		err = errors.Wrap(err, "OffLineImage:d.db.Query")
		return
	}
	defer rows.Close()
	rs = make(map[string]bool)
	for rows.Next() {
		var img string
		if err = rows.Scan(&img); err != nil {
			err = errors.Wrap(err, "OffLineImage:rows.Scan() error")
			return
		}
		rs[img] = true
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RuleWhite:rows.Err")
	}
	return
}
